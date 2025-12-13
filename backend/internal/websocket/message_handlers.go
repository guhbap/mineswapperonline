package websocket

import (
	"log"

	"minesweeperonline/internal/game"
	"github.com/gorilla/websocket"
)

// handlePing обрабатывает ping сообщение
func (h *Handler) handlePing(player *Player) {
	player.mu.Lock()
	defer player.mu.Unlock()
	if player.Conn != nil {
		pongMsg, _ := h.encoder.EncodePong()
		if err := player.Conn.WriteMessage(websocket.BinaryMessage, pongMsg); err != nil {
			log.Printf("Ошибка отправки pong игроку %s: %v", player.ID, err)
		}
	}
}

// handleChat обрабатывает сообщение чата
func (h *Handler) handleChat(room *game.Room, player *Player, playerID string, msg *Message) {
	if msg.Chat != nil {
		player.mu.Lock()
		msg.PlayerID = playerID
		msg.Nickname = player.Nickname
		msg.Color = player.Color
		player.mu.Unlock()
		h.broadcastToAll(room, *msg)
	}
}

// handleNickname обрабатывает установку никнейма
func (h *Handler) handleNickname(room *game.Room, player *Player, playerID string, msg *Message) {
	player.SetNickname(msg.Nickname)
	
	// Обновляем никнейм в комнате через публичные методы
	// Примечание: нужно добавить метод UpdatePlayerNickname в Room
	log.Printf("Никнейм игрока %s установлен: %s", playerID, msg.Nickname)
	h.broadcastPlayerList(room)
}

// handleCursor обрабатывает позицию курсора
func (h *Handler) handleCursor(room *game.Room, player *Player, playerID string, msg *Message) {
	if msg.Cursor != nil {
		if !player.UpdateCursor(msg.Cursor.X, msg.Cursor.Y) {
			return // Пропускаем из-за throttling
		}

		truncatedPlayerID := truncatePlayerID(playerID)
		msg.PlayerID = truncatedPlayerID
		msg.Cursor.PlayerID = truncatedPlayerID
		player.mu.Lock()
		msg.Nickname = player.Nickname
		msg.Color = player.Color
		player.mu.Unlock()

		h.broadcastToOthers(room, playerID, *msg)
	}
}

// handleCellClick обрабатывает клик по ячейке
// Примечание: полная реализация будет в main.go через адаптер
func (h *Handler) handleCellClick(room *game.Room, player *Player, playerID string, msg *Message) {
	if msg.CellClick == nil {
		return
	}
	log.Printf("Обработка cellClick: row=%d, col=%d, flag=%v", msg.CellClick.Row, msg.CellClick.Col, msg.CellClick.Flag)
	// Полная обработка будет делегирована в main.go
}

// handleHint обрабатывает запрос подсказки
// Примечание: полная реализация будет в main.go через адаптер
func (h *Handler) handleHint(room *game.Room, player *Player, playerID string, msg *Message) {
	if msg.Hint != nil {
		log.Printf("Обработка подсказки: row=%d, col=%d", msg.Hint.Row, msg.Hint.Col)
		// Полная обработка будет делегирована в main.go
	}
}

// handleNewGame обрабатывает запрос новой игры
func (h *Handler) handleNewGame(room *game.Room) {
	// Примечание: нужно добавить метод ResetGame в Room
	log.Printf("Новая игра начата")
	h.broadcastGameState(room)
}

// truncatePlayerID обрезает playerID до 5 символов
func truncatePlayerID(playerID string) string {
	if len(playerID) > 5 {
		return playerID[:5]
	}
	return playerID
}

