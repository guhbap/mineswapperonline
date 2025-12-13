package websocket

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"minesweeperonline/internal/game"
	"minesweeperonline/internal/handlers"
	"minesweeperonline/internal/utils"

	"github.com/gorilla/websocket"
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

// Handler обрабатывает WebSocket соединения
type Handler struct {
	roomManager    *game.RoomManager
	profileHandler *handlers.ProfileHandler
	encoder        MessageEncoder
	decoder        MessageDecoder
}

// MessageEncoder интерфейс для кодирования сообщений
type MessageEncoder interface {
	EncodeGameState(gs *game.GameState) ([]byte, error)
	EncodeChat(msg *Message) ([]byte, error)
	EncodeCursor(msg *Message) ([]byte, error)
	EncodePlayers(players []map[string]string) ([]byte, error)
	EncodePong() ([]byte, error)
	EncodeError(errorMsg string) ([]byte, error)
	EncodeCellUpdate(updates []CellUpdate, gameOver, gameWon bool, revealed, hintsUsed int, loserPlayerID, loserNickname string) ([]byte, error)
}

// MessageDecoder интерфейс для декодирования сообщений
type MessageDecoder interface {
	DecodeClientMessage(data []byte) (*Message, error)
}

// NewHandler создает новый WebSocket handler
func NewHandler(roomManager *game.RoomManager, profileHandler *handlers.ProfileHandler, encoder MessageEncoder, decoder MessageDecoder) *Handler {
	return &Handler{
		roomManager:    roomManager,
		profileHandler: profileHandler,
		encoder:        encoder,
		decoder:        decoder,
	}
}

// HandleWebSocket обрабатывает WebSocket соединение
func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
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

	room := h.roomManager.GetRoom(roomID)
	if room == nil {
		errorMsg, _ := h.encoder.EncodeError("Room not found")
		conn.WriteMessage(websocket.BinaryMessage, errorMsg)
		conn.Close()
		return
	}

	// Отменяем удаление комнаты, если кто-то подключается
	room.CancelDeletion()

	playerID := utils.GenerateID()
	color := colors[rand.Intn(len(colors))]

	// Пытаемся получить userID из query параметра
	userIDStr := r.URL.Query().Get("userId")
	var userID int
	if userIDStr != "" {
		if id, err := strconv.Atoi(userIDStr); err == nil {
			userID = id
			if h.profileHandler != nil {
				h.profileHandler.UpdateLastSeen(userID)
				if userColor, err := h.profileHandler.FindUserColor(userID); err == nil && userColor != "" {
					color = userColor
				}
			}
		}
	}

	player := &Player{
		ID:     playerID,
		UserID: userID,
		Color:  color,
		Conn:   conn,
	}

	// Добавляем игрока в комнату
	// Примечание: нужно преобразовать Player в game.Player для комнаты
	roomPlayer := &game.Player{
		ID:       playerID,
		UserID:   userID,
		Nickname: "",
		Color:    color,
	}

	room.mu.Lock()
	// Сохраняем WebSocket Player отдельно для доступа к соединению
	// В реальной реализации можно использовать map[string]*Player в Room
	room.Players[playerID] = roomPlayer
	room.mu.Unlock()

	log.Printf("Игрок %s подключен к комнате %s", playerID, roomID)

	// Настройка ping-pong
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
			player.mu.Lock()
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("Ошибка отправки ping игроку %s: %v", playerID, err)
				player.mu.Unlock()
				return
			}
			player.mu.Unlock()
		}
	}()

	// Отправка начального состояния
	h.sendGameStateToPlayer(room, player)
	h.sendPlayerListToPlayer(room, player)

	// Обработка сообщений
	h.handleMessages(room, player, playerID)

	// Отключение игрока
	room.mu.Lock()
	delete(room.Players, playerID)
	playersLeft := len(room.Players)
	room.mu.Unlock()

	h.broadcastPlayerList(room)
	conn.Close()
	log.Printf("Игрок отключен: %s, игроков в комнате: %d", playerID, playersLeft)

	// Планируем удаление комнаты через 5 минут, если она пустая
	if playersLeft == 0 {
		h.roomManager.ScheduleRoomDeletion(roomID, 5*time.Minute)
	}
}

// handleMessages обрабатывает входящие сообщения
func (h *Handler) handleMessages(room *game.Room, player *Player, playerID string) {
	for {
		messageType, data, err := player.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Ошибка чтения сообщения: %v", err)
			}
			break
		}

		player.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))

		var msg *Message
		var parseErr error

		if messageType == websocket.BinaryMessage {
			msg, parseErr = h.decoder.DecodeClientMessage(data)
			if parseErr != nil {
				log.Printf("Ошибка декодирования protobuf сообщения: %v", parseErr)
				continue
			}
		} else if messageType == websocket.TextMessage {
			var jsonMsg Message
			if parseErr := json.Unmarshal(data, &jsonMsg); parseErr != nil {
				log.Printf("Ошибка парсинга JSON сообщения: %v", parseErr)
				continue
			}
			msg = &jsonMsg
		} else {
			continue
		}

		if msg == nil {
			continue
		}

		if msg.Type != "cursor" {
			log.Printf("Получено сообщение от игрока %s: тип=%s", playerID, msg.Type)
		}

		switch msg.Type {
		case "ping":
			h.handlePing(player)
		case "chat":
			h.handleChat(room, player, playerID, msg)
		case "nickname":
			h.handleNickname(room, player, playerID, msg)
		case "cursor":
			h.handleCursor(room, player, playerID, msg)
		case "cellClick":
			h.handleCellClick(room, player, playerID, msg)
		case "hint":
			h.handleHint(room, player, playerID, msg)
		case "newGame":
			h.handleNewGame(room)
		}
	}
}
