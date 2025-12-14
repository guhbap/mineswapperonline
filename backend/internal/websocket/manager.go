package websocket

import (
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"minesweeperonline/internal/game"
	"minesweeperonline/internal/handlers"
	"minesweeperonline/internal/utils"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Разрешаем все источники для разработки
	},
}

var colors = []string{
	"#FF6B6B", "#4ECDC4", "#45B7D1", "#FFA07A", "#98D8C8",
	"#F7DC6F", "#BB8FCE", "#85C1E2", "#F8B739", "#52BE80",
}

// Manager управляет WebSocket соединениями
type Manager struct {
	roomManager    *game.RoomManager
	profileHandler *handlers.ProfileHandler
	gameService    GameService
	wsPlayers      map[string]*Player
	wsPlayersMu    sync.RWMutex
}

// GameService интерфейс для игровой логики
type GameService interface {
	HandleCellClick(room interface{}, playerID string, click *game.CellClick) error
	HandleHint(room interface{}, playerID string, hint *game.Hint) error
	CalculateCellHints(room interface{})
	DetermineMinePlacement(room interface{}, clickRow, clickCol int) [][]bool
	BroadcastGameState(room interface{})
	BroadcastCellUpdates(room interface{}, changedCells map[[2]int]bool, gameOver, gameWon bool, revealed, hintsUsed int, loserPlayerID, loserNickname string)
	BroadcastToAll(room interface{}, msg game.Message)
	BroadcastToOthers(room interface{}, senderID string, msg game.Message)
	BroadcastPlayerList(room interface{})
	SendGameStateToPlayer(room interface{}, player *Player)
	SendPlayerListToPlayer(room interface{}, player *Player)
}

// NewManager создает новый менеджер WebSocket соединений
func NewManager(roomManager *game.RoomManager, profileHandler *handlers.ProfileHandler, gameService GameService) *Manager {
	return &Manager{
		roomManager:    roomManager,
		profileHandler: profileHandler,
		gameService:    gameService,
		wsPlayers:      make(map[string]*Player),
	}
}

// HandleWebSocket обрабатывает WebSocket соединение
func (m *Manager) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	roomID := r.URL.Query().Get("room")
	if roomID == "" {
		http.Error(w, "Room ID required", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Ошибка обновления соединения: %v", err)
		return
	}

	room := m.roomManager.GetRoom(roomID)
	if room == nil {
		errorMsg, _ := EncodeErrorProtobuf("Room not found")
		conn.WriteMessage(websocket.BinaryMessage, errorMsg)
		conn.Close()
		return
	}

	// Отменяем удаление комнаты, если кто-то подключается
	room.CancelDeletion()

	playerID := utils.GenerateID()
	color := colors[utils.RandInt(len(colors))]

	// Пытаемся получить userID из query параметра (если пользователь авторизован)
	userIDStr := r.URL.Query().Get("userId")
	var userID int
	var initialNickname string
	if userIDStr != "" {
		// Парсим userID, игнорируем ошибку если не число
		if id, err := strconv.Atoi(userIDStr); err == nil {
			userID = id
			// Обновляем last_seen для пользователя
			if m.profileHandler != nil {
				m.profileHandler.UpdateLastSeen(userID)
				// Получаем сохраненный цвет пользователя, если есть
				if userColor, err := m.profileHandler.FindUserColor(userID); err == nil && userColor != "" {
					color = userColor
				}
				// Получаем username из базы данных для авторизованного пользователя
				if user, err := m.profileHandler.FindUserByID(userID); err == nil {
					initialNickname = user.Username
				}
			}
		}
	}

	player := &Player{
		ID:       playerID,
		UserID:   userID,
		Nickname: initialNickname,
		Color:    color,
		Conn:     conn,
	}

	// Сохраняем WebSocket Player в Manager
	m.wsPlayersMu.Lock()
	m.wsPlayers[playerID] = player
	m.wsPlayersMu.Unlock()

	// Добавляем игрока в комнату (game.Player без WebSocket соединения)
	roomPlayer := &game.Player{
		ID:       playerID,
		UserID:   userID,
		Nickname: initialNickname,
		Color:    color,
	}

	// Добавляем игрока в комнату
	room.AddPlayer(playerID, roomPlayer)

	log.Printf("Игрок %s подключен к комнате %s", playerID, roomID)

	// Настройка ping-pong для поддержания соединения
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// Запускаем горутину для отправки ping
	pingTicker := time.NewTicker(30 * time.Second)
	defer pingTicker.Stop()

		go func() {
			for range pingTicker.C {
				player.Mu.Lock()
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					log.Printf("Ошибка отправки ping игроку %s: %v", playerID, err)
					player.Mu.Unlock()
					return
				}
				player.Mu.Unlock()
			}
		}()

	// Отправка начального состояния игры
	m.gameService.SendGameStateToPlayer(room, player)

	// Отправка списка игроков новому игроку
	m.gameService.SendPlayerListToPlayer(room, player)

	// Обработка сообщений
	m.handleMessages(conn, room, player, playerID, roomID)

	// Отключение игрока
	m.removeWSPlayer(playerID)

	// Удаляем из комнаты
	room.RemovePlayer(playerID)

	m.gameService.BroadcastPlayerList(room)
	conn.Close()

	// Получаем количество игроков для логирования
	playersLeft := room.GetPlayerCount()

	log.Printf("Игрок отключен: %s, игроков в комнате: %d", playerID, playersLeft)

	// Планируем удаление комнаты через 5 минут, если она пустая
	if playersLeft == 0 {
		m.roomManager.ScheduleRoomDeletion(roomID, 5*time.Minute)
	}
}

// handleMessages обрабатывает входящие сообщения
func (m *Manager) handleMessages(conn *websocket.Conn, room *game.Room, player *Player, playerID, roomID string) {
	for {
		messageType, data, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Ошибка чтения сообщения: %v", err)
			}
			break
		}

		// Обновляем deadline при получении сообщения
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))

		var msg *game.Message
		var parseErr error

		// Обрабатываем бинарные сообщения (protobuf)
		if messageType == websocket.BinaryMessage {
			log.Printf("[WS IN] Игрок %s: получено бинарное сообщение, размер=%d байт", playerID, len(data))
			msg, parseErr = decodeClientMessageProtobuf(data)
			if parseErr != nil {
				log.Printf("[WS IN] Ошибка декодирования protobuf сообщения от игрока %s: %v", playerID, parseErr)
				continue
			}
		} else if messageType == websocket.TextMessage {
			// Fallback: парсим JSON сообщение (для обратной совместимости)
			log.Printf("[WS IN] Игрок %s: получено текстовое сообщение, размер=%d байт", playerID, len(data))
			var jsonMsg game.Message
			if parseErr := utils.DecodeJSONFromBytes(data, &jsonMsg); parseErr != nil {
				log.Printf("[WS IN] Ошибка парсинга JSON сообщения от игрока %s: %v", playerID, parseErr)
				continue
			}
			msg = &jsonMsg
		} else {
			log.Printf("[WS IN] Игрок %s: неизвестный тип сообщения: %d", playerID, messageType)
			continue
		}

		if msg == nil {
			log.Printf("[WS IN] Игрок %s: сообщение декодировано как nil", playerID)
			continue
		}

		// Детальное логирование входящих сообщений
		if msg.Type == "cellClick" && msg.CellClick != nil {
			log.Printf("[WS IN] Игрок %s: cellClick - row=%d, col=%d, flag=%v", playerID, msg.CellClick.Row, msg.CellClick.Col, msg.CellClick.Flag)
		} else if msg.Type == "cursor" && msg.Cursor != nil {
			// Курсор логируем реже, чтобы не засорять логи
			// log.Printf("[WS IN] Игрок %s: cursor - x=%.2f, y=%.2f", playerID, msg.Cursor.X, msg.Cursor.Y)
		} else if msg.Type == "hint" && msg.Hint != nil {
			log.Printf("[WS IN] Игрок %s: hint - row=%d, col=%d", playerID, msg.Hint.Row, msg.Hint.Col)
		} else if msg.Type == "chat" && msg.Chat != nil {
			log.Printf("[WS IN] Игрок %s: chat - text=%s", playerID, msg.Chat.Text)
		} else {
			log.Printf("[WS IN] Игрок %s: тип=%s", playerID, msg.Type)
		}

		switch msg.Type {
		case "ping":
			m.handlePing(player, playerID)
		case "chat":
			m.handleChat(room, player, playerID, msg)
		case "nickname":
			m.handleNickname(room, player, playerID, msg, roomID)
		case "cursor":
			m.handleCursor(room, player, playerID, msg)
		case "cellClick":
			m.handleCellClick(room, playerID, msg)
		case "hint":
			m.handleHint(room, playerID, msg)
		case "newGame":
			m.handleNewGame(room, roomID)
		}
	}
}

// handlePing обрабатывает ping сообщение
func (m *Manager) handlePing(player *Player, playerID string) {
	player.Mu.Lock()
	defer player.Mu.Unlock()
	if player.Conn != nil {
		pongMsg, _ := EncodePongProtobuf()
		if err := player.Conn.WriteMessage(websocket.BinaryMessage, pongMsg); err != nil {
			log.Printf("[WS OUT] Ошибка отправки pong игроку %s: %v", playerID, err)
		} else {
			log.Printf("[WS OUT] Игрок %s: отправлен pong, размер=%d байт", playerID, len(pongMsg))
		}
	}
}

// handleChat обрабатывает сообщение чата
func (m *Manager) handleChat(room *game.Room, player *Player, playerID string, msg *game.Message) {
	if msg.Chat != nil {
		player.Mu.Lock()
		msg.PlayerID = playerID
		msg.Nickname = player.Nickname
		msg.Color = player.Color
		player.Mu.Unlock()
		m.gameService.BroadcastToAll(room, *msg)
	}
}

// handleNickname обрабатывает изменение никнейма
func (m *Manager) handleNickname(room *game.Room, player *Player, playerID string, msg *game.Message, roomID string) {
	player.Mu.Lock()
	player.Nickname = msg.Nickname
	player.Mu.Unlock()

	// Обновляем никнейм также в room.Players
	room.Mu.Lock()
	if roomPlayer := room.Players[playerID]; roomPlayer != nil {
		roomPlayer.Nickname = msg.Nickname
	}
	room.Mu.Unlock()

	log.Printf("Никнейм игрока %s установлен: %s", playerID, msg.Nickname)
	m.gameService.BroadcastPlayerList(room)
}

// handleCursor обрабатывает движение курсора
func (m *Manager) handleCursor(room *game.Room, player *Player, playerID string, msg *game.Message) {
	if msg.Cursor != nil {
		player.Mu.Lock()
		if !player.UpdateCursor(msg.Cursor.X, msg.Cursor.Y) {
			player.Mu.Unlock()
			return // Пропускаем это сообщение
		}

		truncatedPlayerID := truncatePlayerID(playerID)
		msg.PlayerID = truncatedPlayerID
		msg.Cursor.PlayerID = truncatedPlayerID
		msg.Nickname = player.Nickname
		msg.Color = player.Color
		player.Mu.Unlock()

		m.gameService.BroadcastToOthers(room, playerID, *msg)
	}
}

// handleCellClick обрабатывает клик по ячейке
func (m *Manager) handleCellClick(room *game.Room, playerID string, msg *game.Message) {
	if msg.CellClick != nil {
		log.Printf("Обработка cellClick: row=%d, col=%d, flag=%v", msg.CellClick.Row, msg.CellClick.Col, msg.CellClick.Flag)
		defer func() {
			if r := recover(); r != nil {
				log.Printf("ПАНИКА в handleCellClick: %v", r)
			}
		}()
		if err := m.gameService.HandleCellClick(room, playerID, msg.CellClick); err != nil {
			log.Printf("Ошибка обработки клика: %v", err)
		}
	}
}

// handleHint обрабатывает подсказку
func (m *Manager) handleHint(room *game.Room, playerID string, msg *game.Message) {
	if msg.Hint != nil {
		log.Printf("Обработка подсказки: row=%d, col=%d", msg.Hint.Row, msg.Hint.Col)
		if err := m.gameService.HandleHint(room, playerID, msg.Hint); err != nil {
			log.Printf("Ошибка обработки подсказки: %v", err)
		}
	}
}

// handleNewGame обрабатывает запрос новой игры
func (m *Manager) handleNewGame(room *game.Room, roomID string) {
	log.Printf("Обработка newGame для комнаты %s", roomID)
	go func() {
		log.Printf("Сброс игры для комнаты %s (асинхронно)", roomID)
		room.ResetGame()
		log.Printf("Новая игра начата для комнаты %s", roomID)
		// Сохраняем комнату в БД после сброса игры
		if err := m.roomManager.SaveRoom(room); err != nil {
			log.Printf("Предупреждение: не удалось сохранить комнату %s после сброса игры: %v", roomID, err)
		}
		// Отправляем состояние игры после сброса
		m.gameService.BroadcastGameState(room)
	}()
}

// GetWSPlayer получает WebSocket Player по ID
func (m *Manager) GetWSPlayer(playerID string) *Player {
	m.wsPlayersMu.RLock()
	defer m.wsPlayersMu.RUnlock()
	return m.wsPlayers[playerID]
}

// getWSPlayer получает WebSocket Player по ID (внутренний метод)
func (m *Manager) getWSPlayer(playerID string) *Player {
	return m.GetWSPlayer(playerID)
}

// removeWSPlayer удаляет WebSocket Player
func (m *Manager) removeWSPlayer(playerID string) {
	m.wsPlayersMu.Lock()
	defer m.wsPlayersMu.Unlock()
	delete(m.wsPlayers, playerID)
}

