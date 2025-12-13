package websocket

import (
	"log"

	"minesweeperonline/internal/game"
)

// sendGameStateToPlayer отправляет состояние игры игроку
func (h *Handler) sendGameStateToPlayer(room *game.Room, player *Player) {
	room.GameState.mu.RLock()
	gameStateCopy := room.GameState.Copy()
	room.GameState.mu.RUnlock()

	binaryData, err := h.encoder.EncodeGameState(gameStateCopy)
	if err != nil {
		log.Printf("Ошибка кодирования gameState: %v", err)
		return
	}

	player.mu.Lock()
	defer player.mu.Unlock()

	log.Printf("Отправка gameState (protobuf): Rows=%d, Cols=%d, Mines=%d, Revealed=%d, Size=%d bytes",
		gameStateCopy.Rows, gameStateCopy.Cols, gameStateCopy.Mines, gameStateCopy.Revealed, len(binaryData))
	if err := player.Conn.WriteMessage(websocket.BinaryMessage, binaryData); err != nil {
		log.Printf("Ошибка отправки состояния игры: %v", err)
	}
}

// sendPlayerListToPlayer отправляет список игроков
func (h *Handler) sendPlayerListToPlayer(room *game.Room, targetPlayer *Player) {
	room.mu.RLock()
	playersList := make([]map[string]string, 0, len(room.Players))
	for _, player := range room.Players {
		playersList = append(playersList, map[string]string{
			"id":       player.ID,
			"nickname": player.Nickname,
			"color":    player.Color,
		})
	}
	room.mu.RUnlock()

	binaryData, err := h.encoder.EncodePlayers(playersList)
	if err != nil {
		log.Printf("Ошибка кодирования списка игроков: %v", err)
		return
	}

	targetPlayer.mu.Lock()
	defer targetPlayer.mu.Unlock()
	if targetPlayer.Conn != nil {
		if err := targetPlayer.Conn.WriteMessage(websocket.BinaryMessage, binaryData); err != nil {
			log.Printf("Ошибка отправки списка игроков: %v", err)
		}
	}
}

// broadcastPlayerList отправляет список игроков всем
func (h *Handler) broadcastPlayerList(room *game.Room) {
	room.mu.RLock()
	playersList := make([]map[string]string, 0, len(room.Players))
	for _, player := range room.Players {
		playersList = append(playersList, map[string]string{
			"id":       player.ID,
			"nickname": player.Nickname,
			"color":    player.Color,
		})
	}
	room.mu.RUnlock()

	binaryData, err := h.encoder.EncodePlayers(playersList)
	if err != nil {
		log.Printf("Ошибка кодирования списка игроков: %v", err)
		return
	}

	// Примечание: для отправки нужно получить WebSocket соединения
	// В текущей реализации это будет сделано через адаптер
	log.Printf("Broadcast player list: %d игроков", len(playersList))
}

// broadcastToOthers отправляет сообщение всем кроме отправителя
func (h *Handler) broadcastToOthers(room *game.Room, senderID string, msg Message) {
	room.mu.RLock()
	playersCount := len(room.Players)
	room.mu.RUnlock()

	if playersCount <= 1 {
		return
	}

	var binaryData []byte
	var err error
	if msg.Type == "cursor" && msg.Cursor != nil {
		binaryData, err = h.encoder.EncodeCursor(&msg)
		if err != nil {
			log.Printf("Ошибка кодирования курсора: %v", err)
			return
		}
	} else {
		return
	}

	// Примечание: для отправки нужно получить WebSocket соединения
	log.Printf("Broadcast cursor to others")
}

// broadcastToAll отправляет сообщение всем игрокам
func (h *Handler) broadcastToAll(room *game.Room, msg Message) {
	if msg.Type == "chat" && msg.Chat != nil {
		binaryData, err := h.encoder.EncodeChat(&msg)
		if err != nil {
			log.Printf("Ошибка кодирования чата: %v", err)
			return
		}

		// Примечание: для отправки нужно получить WebSocket соединения
		log.Printf("Broadcast chat message")
	}
}

// broadcastGameState отправляет состояние игры всем игрокам
func (h *Handler) broadcastGameState(room *game.Room) {
	room.GameState.mu.RLock()
	gameStateCopy := room.GameState.Copy()
	room.GameState.mu.RUnlock()

	binaryData, err := h.encoder.EncodeGameState(gameStateCopy)
	if err != nil {
		log.Printf("Ошибка кодирования gameState: %v", err)
		return
	}

	log.Printf("Broadcast gameState (protobuf): Rows=%d, Cols=%d, Revealed=%d, GameOver=%v, GameWon=%v, Size=%d bytes",
		gameStateCopy.Rows, gameStateCopy.Cols, gameStateCopy.Revealed, gameStateCopy.GameOver, gameStateCopy.GameWon, len(binaryData))

	// Примечание: для отправки нужно получить WebSocket соединения
	log.Printf("Broadcast gameState to all players")
}

