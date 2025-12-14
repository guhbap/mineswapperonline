package main

import (
	"encoding/json"
	"fmt"
	"log"
	mathrand "math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	"minesweeperonline/internal/config"
	"minesweeperonline/internal/database"
	"minesweeperonline/internal/game"
	"minesweeperonline/internal/handlers"
	"minesweeperonline/internal/middleware"
	"minesweeperonline/internal/utils"
	ws "minesweeperonline/internal/websocket"

	"github.com/gorilla/mux"
	gorillaWS "github.com/gorilla/websocket"
)

var upgrader = gorillaWS.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Разрешаем все источники для разработки
	},
}

type Player struct {
	ID                 string `json:"id"`
	UserID             int    `json:"userId,omitempty"` // ID пользователя из БД, если авторизован
	Nickname           string `json:"nickname"`
	Color              string `json:"color"`
	Conn               *gorillaWS.Conn
	mu                 sync.Mutex
	LastCursorX        float64   // Последняя отправленная позиция курсора X
	LastCursorY        float64   // Последняя отправленная позиция курсора Y
	LastCursorSendTime time.Time // Время последней отправки курсора
}

type CursorPosition struct {
	PlayerID string  `json:"pid"` // playerId сокращено до pid
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
}

type FlagInfo struct {
	SetTime  time.Time
	PlayerID string
}

type SafeCell struct {
	Row int `json:"r"`
	Col int `json:"c"`
}

type CellHint struct {
	Row  int    `json:"r"`
	Col  int    `json:"c"`
	Type string `json:"t"` // "MINE", "SAFE", "UNKNOWN"
}

type GameState struct {
	Board         [][]Cell         `json:"b"`
	Rows          int              `json:"r"`
	Cols          int              `json:"c"`
	Mines         int              `json:"m"`
	Seed          string           `json:"seed,omitempty"` // Seed для генерации поля (UUID)
	GameOver      bool             `json:"go"`
	GameWon       bool             `json:"gw"`
	Revealed      int              `json:"rv"`
	HintsUsed     int              `json:"hu"`              // Количество использованных подсказок (глобально для комнаты)
	SafeCells     []SafeCell       `json:"sc,omitempty"`    // Безопасные ячейки для режима без угадываний
	CellHints     []CellHint       `json:"hints,omitempty"` // Подсказки для ячеек (показываются в training всегда, в fair при проигрыше)
	LoserPlayerID string           `json:"lpid,omitempty"`
	LoserNickname string           `json:"ln,omitempty"`
	flagSetInfo   map[int]FlagInfo // Информация об установке флага для каждой ячейки (ключ: row*cols + col)
	mu            sync.RWMutex
}

type Cell struct {
	IsMine        bool   `json:"m"`
	IsRevealed    bool   `json:"r"`
	IsFlagged     bool   `json:"f"`
	NeighborMines int    `json:"n"`
	FlagColor     string `json:"fc,omitempty"` // Цвет игрока, который поставил флаг
}

type Message struct {
	Type      string          `json:"type"`
	PlayerID  string          `json:"playerId,omitempty"`
	Nickname  string          `json:"nickname,omitempty"`
	Color     string          `json:"color,omitempty"`
	Cursor    *CursorPosition `json:"cursor,omitempty"`
	CellClick *CellClick      `json:"cellClick,omitempty"`
	Hint      *Hint           `json:"hint,omitempty"`
	GameState *GameState      `json:"gameState,omitempty"`
	Chat      *ChatMessage    `json:"chat,omitempty"`
}

type ChatMessage struct {
	Text     string `json:"text"`
	IsSystem bool   `json:"isSystem,omitempty"`
	Action   string `json:"action,omitempty"` // "flag", "reveal", "explode"
	Row      int    `json:"row,omitempty"`
	Col      int    `json:"col,omitempty"`
}

type CellClick struct {
	Row  int  `json:"row"`
	Col  int  `json:"col"`
	Flag bool `json:"flag"`
}

type Hint struct {
	Row int `json:"row"`
	Col int `json:"col"`
}

// Server управляет WebSocket соединениями и игровой логикой
type Server struct {
	roomManager    *game.RoomManager
	db             *database.DB
	profileHandler *handlers.ProfileHandler
	// Хранилище WebSocket соединений игроков (playerID -> *Player)
	wsPlayers   map[string]*Player
	wsPlayersMu sync.RWMutex
}

var colors = []string{
	"#FF6B6B", "#4ECDC4", "#45B7D1", "#FFA07A", "#98D8C8",
	"#F7DC6F", "#BB8FCE", "#85C1E2", "#F8B739", "#52BE80",
}

func NewServer(roomManager *game.RoomManager, db *database.DB) *Server {
	server := &Server{
		roomManager:    roomManager,
		db:             db,
		profileHandler: handlers.NewProfileHandler(db),
		wsPlayers:      make(map[string]*Player),
	}
	// Устанавливаем ссылку на сервер в RoomManager для доступа к DeleteRoom
	roomManager.SetServer(server)
	return server
}

func NewGameState(rows, cols, mines int, gameMode string) *GameState {
	// Эта функция используется только в main.go для совместимости
	// Внутри используется game.NewGameState с пустым seed (будет сгенерирован UUID)
	gs := game.NewGameState(rows, cols, mines, gameMode, "")
	return convertGameStateToMain(gs)
}

// generateRandomBoard создает случайное поле (используется как fallback)
//
//lint:ignore U1000 Используется для отладки и тестирования
func generateRandomBoard(rows, cols, mines int) *GameState {
	gs := &GameState{
		Rows:          rows,
		Cols:          cols,
		Mines:         mines,
		GameOver:      false,
		GameWon:       false,
		Revealed:      0,
		HintsUsed:     0,
		LoserPlayerID: "",
		LoserNickname: "",
		Board:         make([][]Cell, rows),
		flagSetInfo:   make(map[int]FlagInfo),
	}

	for i := range gs.Board {
		gs.Board[i] = make([]Cell, cols)
	}

	minesPlaced := 0
	for minesPlaced < mines {
		row := mathrand.Intn(rows)
		col := mathrand.Intn(cols)
		if !gs.Board[row][col].IsMine {
			gs.Board[row][col].IsMine = true
			minesPlaced++
		}
	}

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			if !gs.Board[i][j].IsMine {
				count := 0
				for di := -1; di <= 1; di++ {
					for dj := -1; dj <= 1; dj++ {
						ni, nj := i+di, j+dj
						if ni >= 0 && ni < rows && nj >= 0 && nj < cols {
							if gs.Board[ni][nj].IsMine {
								count++
							}
						}
					}
				}
				gs.Board[i][j].NeighborMines = count
			}
		}
	}

	return gs
}

// Вспомогательная функция: получить соседей
//
//lint:ignore U1000 Используется для отладки и тестирования
func neighbors(rows, cols, i, j int) [][2]int {
	out := [][2]int{}
	for di := -1; di <= 1; di++ {
		for dj := -1; dj <= 1; dj++ {
			if di == 0 && dj == 0 {
				continue
			}
			ni, nj := i+di, j+dj
			if ni < 0 || ni >= rows || nj < 0 || nj >= cols {
				continue
			}
			out = append(out, [2]int{ni, nj})
		}
	}
	return out
}

// Методы RoomManager перемещены в internal/game/room_manager.go

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
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

	room := s.roomManager.GetRoom(roomID)
	if room == nil {
		errorMsg, _ := encodeErrorProtobuf("Room not found")
		conn.WriteMessage(gorillaWS.BinaryMessage, errorMsg)
		conn.Close()
		return
	}

	// Отменяем удаление комнаты, если кто-то подключается
	room.CancelDeletion()

	playerID := utils.GenerateID()
	color := colors[mathrand.Intn(len(colors))]

	// Пытаемся получить userID из query параметра (если пользователь авторизован)
	userIDStr := r.URL.Query().Get("userId")
	var userID int
	var initialNickname string
	if userIDStr != "" {
		// Парсим userID, игнорируем ошибку если не число
		if id, err := strconv.Atoi(userIDStr); err == nil {
			userID = id
			// Обновляем last_seen для пользователя
			if s.profileHandler != nil {
				s.profileHandler.UpdateLastSeen(userID)
				// Получаем сохраненный цвет пользователя, если есть
				if userColor, err := s.profileHandler.FindUserColor(userID); err == nil && userColor != "" {
					color = userColor
				}
				// Получаем username из базы данных для авторизованного пользователя
				if user, err := s.profileHandler.FindUserByID(userID); err == nil {
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

	// Сохраняем WebSocket Player в Server
	s.wsPlayersMu.Lock()
	s.wsPlayers[playerID] = player
	s.wsPlayersMu.Unlock()

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
			player.mu.Lock()
			if err := conn.WriteMessage(gorillaWS.PingMessage, nil); err != nil {
				log.Printf("Ошибка отправки ping игроку %s: %v", playerID, err)
				player.mu.Unlock()
				return
			}
			player.mu.Unlock()
		}
	}()

	// Отправка начального состояния игры
	log.Printf("Отправка начального состояния игры игроку %s", playerID)
	s.sendGameStateToPlayer(room, player)
	log.Printf("Начальное состояние игры отправлено игроку %s", playerID)

	// Отправка списка игроков новому игроку
	s.sendPlayerListToPlayer(room, player)

	// Обработка сообщений
	for {
		messageType, data, err := conn.ReadMessage()
		if err != nil {
			if gorillaWS.IsUnexpectedCloseError(err, gorillaWS.CloseGoingAway, gorillaWS.CloseAbnormalClosure) {
				log.Printf("Ошибка чтения сообщения: %v", err)
			}
			break
		}

		// Обновляем deadline при получении сообщения
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))

		var msg *Message
		var parseErr error

		// Обрабатываем бинарные сообщения (protobuf)
		if messageType == gorillaWS.BinaryMessage {
			msg, parseErr = decodeClientMessageProtobuf(data)
			if parseErr != nil {
				log.Printf("Ошибка декодирования protobuf сообщения: %v", parseErr)
				continue
			}
		} else if messageType == gorillaWS.TextMessage {
			// Fallback: парсим JSON сообщение (для обратной совместимости)
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
			log.Printf("Получено сообщение от игрока %s: тип=%s, полное сообщение: %+v", playerID, msg.Type, *msg)
		}
		switch msg.Type {
		case "ping":
			// Отвечаем pong на ping сообщение
			player.mu.Lock()
			if player.Conn != nil {
				pongMsg, _ := encodePongProtobuf()
				if err := player.Conn.WriteMessage(gorillaWS.BinaryMessage, pongMsg); err != nil {
					log.Printf("Ошибка отправки pong игроку %s: %v", playerID, err)
				}
			}
			player.mu.Unlock()
			continue

		case "chat":
			if msg.Chat != nil {
				player.mu.Lock()
				msg.PlayerID = playerID
				msg.Nickname = player.Nickname
				msg.Color = player.Color
				player.mu.Unlock()
				// Отправляем сообщение всем игрокам в комнате
				s.broadcastToAll(room, *msg)
			}
			continue

		case "nickname":
			player.mu.Lock()
			player.Nickname = msg.Nickname
			player.mu.Unlock()
			// Обновляем никнейм также в room.Players
			log.Printf("[MUTEX] nickname: блокируем room.Mu.Lock() для комнаты %s, игрок %s", roomID, playerID)
			room.Mu.Lock()
			log.Printf("[MUTEX] nickname: room.Mu.Lock() заблокирован для комнаты %s, игрок %s", roomID, playerID)
			if roomPlayer := room.Players[playerID]; roomPlayer != nil {
				roomPlayer.Nickname = msg.Nickname
			}
			log.Printf("[MUTEX] nickname: разблокируем room.Mu.Unlock() для комнаты %s, игрок %s", roomID, playerID)
			room.Mu.Unlock()
			log.Printf("[MUTEX] nickname: room.Mu.Unlock() разблокирован для комнаты %s, игрок %s", roomID, playerID)
			log.Printf("Никнейм игрока %s установлен: %s", playerID, msg.Nickname)
			s.broadcastPlayerList(room)

		case "cursor":
			if msg.Cursor != nil {
				player.mu.Lock()
				now := time.Now()
				// Throttling: отправляем не чаще чем раз в 100ms и только если позиция изменилась минимум на 5px
				timeSinceLastSend := now.Sub(player.LastCursorSendTime)
				dx := msg.Cursor.X - player.LastCursorX
				dy := msg.Cursor.Y - player.LastCursorY
				distance := dx*dx + dy*dy // квадрат расстояния для оптимизации

				// Отправляем только если прошло достаточно времени И позиция изменилась значительно
				if timeSinceLastSend < 100*time.Millisecond && distance < 25 { // 5px * 5px = 25
					player.mu.Unlock()
					continue // Пропускаем это сообщение
				}

				// Обрезаем playerID до 5 символов
				truncatedPlayerID := truncatePlayerID(playerID)
				msg.PlayerID = truncatedPlayerID
				msg.Cursor.PlayerID = truncatedPlayerID
				msg.Nickname = player.Nickname
				msg.Color = player.Color

				// Обновляем последнюю позицию и время
				player.LastCursorX = msg.Cursor.X
				player.LastCursorY = msg.Cursor.Y
				player.LastCursorSendTime = now
				player.mu.Unlock()

				s.broadcastToOthers(room, playerID, *msg)
			}

		case "cellClick":
			log.Printf("Обработка cellClick: msg.CellClick=%+v", msg.CellClick)
			if msg.CellClick != nil {
				log.Printf("Обработка клика: row=%d, col=%d, flag=%v", msg.CellClick.Row, msg.CellClick.Col, msg.CellClick.Flag)
				defer func() {
					if r := recover(); r != nil {
						log.Printf("ПАНИКА в handleCellClick: %v", r)
					}
				}()
				s.handleCellClick(room, playerID, msg.CellClick)
				log.Printf("Клик обработан, состояние игры обновлено")
			} else {
				log.Printf("ОШИБКА: msg.CellClick == nil для сообщения типа cellClick")
			}

		case "hint":
			if msg.Hint != nil {
				log.Printf("Обработка подсказки: row=%d, col=%d", msg.Hint.Row, msg.Hint.Col)
				s.handleHint(room, playerID, msg.Hint)
			}

		case "newGame":
			// Сбрасываем игру асинхронно, чтобы не блокировать обработку сообщений
			log.Printf("Обработка newGame от игрока %s", playerID)
			go func() {
				log.Printf("Сброс игры для комнаты %s (асинхронно)", roomID)
				room.ResetGame()
				log.Printf("Новая игра начата для комнаты %s", roomID)
				// Сохраняем комнату в БД после сброса игры
				if err := s.roomManager.SaveRoom(room); err != nil {
					log.Printf("Предупреждение: не удалось сохранить комнату %s после сброса игры: %v", roomID, err)
				}
				// Отправляем состояние игры после сброса
				log.Printf("Отправка состояния новой игры для комнаты %s", roomID)
				s.broadcastGameState(room)
				log.Printf("Состояние новой игры отправлено для комнаты %s", roomID)
			}()
		}
	}

	// Отключение игрока
	// Удаляем из WebSocket хранилища
	s.removeWSPlayer(playerID)

	// Удаляем из комнаты
	room.RemovePlayer(playerID)

	s.broadcastPlayerList(room)
	conn.Close()

	// Получаем количество игроков для логирования
	playersLeft := room.GetPlayerCount()

	log.Printf("Игрок отключен: %s, игроков в комнате: %d", playerID, playersLeft)

	// Планируем удаление комнаты через 5 минут, если она пустая
	if playersLeft == 0 {
		s.roomManager.ScheduleRoomDeletion(roomID, 5*time.Minute)
	}
}

func (s *Server) handleCellClick(room *game.Room, playerID string, click *CellClick) {
	log.Printf("handleCellClick: начало, row=%d, col=%d, flag=%v", click.Row, click.Col, click.Flag)
	log.Printf("handleCellClick: пытаемся заблокировать GameState.mu")
	room.GameState.Mu.Lock()
	log.Printf("handleCellClick: мьютекс GameState заблокирован успешно")

	if room.GameState.GameOver || room.GameState.GameWon {
		log.Printf("Игра уже окончена, клик игнорируется")
		room.GameState.Mu.Unlock()
		return
	}

	row, col := click.Row, click.Col
	if row < 0 || row >= room.GameState.Rows || col < 0 || col >= room.GameState.Cols {
		log.Printf("Некорректные координаты: row=%d, col=%d", row, col)
		room.GameState.Mu.Unlock()
		return
	}

	log.Printf("handleCellClick: координаты валидны, получаем ячейку")
	cell := &room.GameState.Board[row][col]
	log.Printf("handleCellClick: ячейка получена, isRevealed=%v, isFlagged=%v, isMine=%v", cell.IsRevealed, cell.IsFlagged, cell.IsMine)

	// Получаем информацию об игроке для сервисных сообщений
	log.Printf("[MUTEX] handleCellClick: блокируем room.Mu.RLock() для комнаты, игрок %s", playerID)
	room.Mu.RLock()
	log.Printf("[MUTEX] handleCellClick: room.Mu.RLock() заблокирован для комнаты, игрок %s", playerID)
	player := room.Players[playerID]
	var nickname string
	var playerColor string
	if player != nil {
		nickname = player.Nickname
		playerColor = player.Color
	}
	log.Printf("[MUTEX] handleCellClick: разблокируем room.Mu.RUnlock() для комнаты, игрок %s", playerID)
	room.Mu.RUnlock()
	log.Printf("[MUTEX] handleCellClick: room.Mu.RUnlock() разблокирован для комнаты, игрок %s", playerID)

	if click.Flag {
		// Переключение флага - нельзя ставить на открытые ячейки
		if cell.IsRevealed {
			log.Printf("Нельзя поставить флаг на открытую ячейку: row=%d, col=%d", row, col)
			room.GameState.Mu.Unlock()
			return
		}

		wasFlagged := cell.IsFlagged
		cellKey := row*room.GameState.Cols + col
		now := time.Now()

		// Если пытаемся снять флаг, проверяем защиту от одновременных кликов
		if wasFlagged {
			if flagInfo, exists := room.GameState.FlagSetInfo[cellKey]; exists {
				// Если это тот же игрок, который поставил флаг - разрешаем снять сразу
				if flagInfo.PlayerID != playerID {
					// Если это другой игрок - применяем защиту в 1 секунду
					timeSinceFlagSet := now.Sub(flagInfo.SetTime)
					if timeSinceFlagSet < 1*time.Second {
						log.Printf("Нельзя снять флаг сразу после установки другим игроком (защита от одновременных кликов): row=%d, col=%d, прошло %v", row, col, timeSinceFlagSet)
						room.GameState.Mu.Unlock()
						return
					}
				}
			}
			// Удаляем информацию об установке при снятии флага
			delete(room.GameState.FlagSetInfo, cellKey)
			cell.FlagColor = "" // Очищаем цвет при снятии флага
		} else {
			// Сохраняем время установки и playerID того, кто поставил флаг
			room.GameState.FlagSetInfo[cellKey] = game.FlagInfo{
				SetTime:  now,
				PlayerID: playerID,
			}
			// Сохраняем цвет игрока, который поставил флаг
			cell.FlagColor = playerColor
		}

		cell.IsFlagged = !cell.IsFlagged
		log.Printf("Флаг переключен: row=%d, col=%d, flagged=%v", row, col, cell.IsFlagged)

		// В режиме training пересчитываем подсказки асинхронно после установки/снятия флага (в fair подсказки только при проигрыше)
		// Получаем gameMode из room (нужен доступ через game пакет)
		gameMode := room.GameMode

		room.GameState.Mu.Unlock()
		// Отправляем состояние игры асинхронно, чтобы не блокировать обработку
		go func() {
			s.broadcastGameState(room)
		}()

		// Выполняем асинхронно, чтобы не блокировать ответ
		if gameMode == "training" {
			go func() {
				// calculateCellHints сама блокирует мьютекс, не нужно блокировать здесь
				s.calculateCellHints(room)
				s.broadcastGameState(room)
			}()
		}

		// Отправляем сервисное сообщение в чат
		if nickname != "" {
			action := "поставил флаг"
			if wasFlagged {
				action = "убрал флаг"
			}
			chatMsg := Message{
				Type:     "chat",
				PlayerID: playerID,
				Nickname: nickname,
				Color:    playerColor,
				Chat: &ChatMessage{
					Text:     fmt.Sprintf("%s %s на (%d, %d)", nickname, action, row+1, col+1),
					IsSystem: true,
					Action:   "flag",
					Row:      row,
					Col:      col,
				},
			}
			s.broadcastToAll(room, chatMsg)
		}
		return
	}

	// Открытие ячейки (проверка уже выполнена выше)
	log.Printf("handleCellClick: начинаем открытие ячейки")

	// Проверка: нельзя открыть ячейку с флагом
	if cell.IsFlagged {
		log.Printf("Нельзя открыть ячейку с флагом: row=%d, col=%d", row, col)
		room.GameState.Mu.Unlock()
		return
	}

	// Проверяем режим игры
	gameMode := room.GameMode

	// Chording: если клик на открытую клетку с цифрой и вокруг стоит нужное количество флагов
	if room.Chording && cell.IsRevealed && cell.NeighborMines > 0 {
		log.Printf("handleCellClick: проверяем chording для row=%d, col=%d, neighborMines=%d", row, col, cell.NeighborMines)
		// Подсчитываем количество флагов вокруг
		flagCount := 0
		for di := -1; di <= 1; di++ {
			for dj := -1; dj <= 1; dj++ {
				if di == 0 && dj == 0 {
					continue
				}
				ni, nj := row+di, col+dj
				if ni >= 0 && ni < room.GameState.Rows && nj >= 0 && nj < room.GameState.Cols {
					if room.GameState.Board[ni][nj].IsFlagged {
						flagCount++
					}
				}
			}
		}
		log.Printf("handleCellClick: chording - флагов вокруг: %d, нужно: %d", flagCount, cell.NeighborMines)

		// Если количество флагов равно цифре на клетке, открываем соседние закрытые клетки
		if flagCount == cell.NeighborMines {
			log.Printf("handleCellClick: chording активирован, открываем соседние клетки")
			changedCells := make(map[[2]int]bool)
			for di := -1; di <= 1; di++ {
				for dj := -1; dj <= 1; dj++ {
					if di == 0 && dj == 0 {
						continue
					}
					ni, nj := row+di, col+dj
					if ni >= 0 && ni < room.GameState.Rows && nj >= 0 && nj < room.GameState.Cols {
						neighborCell := &room.GameState.Board[ni][nj]
						// Открываем только закрытые клетки без флагов
						if !neighborCell.IsRevealed && !neighborCell.IsFlagged {
							neighborCell.IsRevealed = true
							room.GameState.Revealed++
							changedCells[[2]int{ni, nj}] = true
							log.Printf("handleCellClick: chording - открыта клетка (%d, %d), isMine=%v", ni, nj, neighborCell.IsMine)

							// Если это мина - игра окончена
							if neighborCell.IsMine {
								room.GameState.GameOver = true
								wsPlayer := s.getWSPlayer(playerID)
								var userID int
								var nickname string
								if wsPlayer != nil {
									wsPlayer.mu.Lock()
									userID = wsPlayer.UserID
									nickname = wsPlayer.Nickname
									wsPlayer.mu.Unlock()
								} else {
									roomPlayer := room.GetPlayer(playerID)
									if roomPlayer != nil {
										userID = roomPlayer.UserID
										nickname = roomPlayer.Nickname
									}
								}
								if nickname != "" {
									room.GameState.LoserPlayerID = playerID
									room.GameState.LoserNickname = nickname
								}

								log.Printf("[MUTEX] handleCellClick (взрыв chording): блокируем room.Mu.RLock() для получения gameTime")
								room.Mu.RLock()
								log.Printf("[MUTEX] handleCellClick (взрыв chording): room.Mu.RLock() заблокирован для получения gameTime")
								var gameTime float64
								if room.StartTime != nil {
									gameTime = time.Since(*room.StartTime).Seconds()
								}
								log.Printf("[MUTEX] handleCellClick (взрыв chording): разблокируем room.Mu.RUnlock() после получения gameTime")
								room.Mu.RUnlock()
								log.Printf("[MUTEX] handleCellClick (взрыв chording): room.Mu.RUnlock() разблокирован после получения gameTime")

								if userID > 0 {
									// Собираем список участников для записи результата
									log.Printf("[MUTEX] handleCellClick (взрыв chording): блокируем room.Mu.RLock() для сбора участников")
									room.Mu.RLock()
									log.Printf("[MUTEX] handleCellClick (взрыв chording): room.Mu.RLock() заблокирован для сбора участников")
									participants := make([]game.GameParticipant, 0)
									for _, p := range room.Players {
										if p.UserID > 0 {
											participants = append(participants, game.GameParticipant{
												UserID:   p.UserID,
												Nickname: p.Nickname,
												Color:    p.Color,
											})
										}
									}
									log.Printf("[MUTEX] handleCellClick (взрыв chording): разблокируем room.Mu.RUnlock() после сбора участников")
									chording := room.Chording
									quickStart := room.QuickStart
									roomID := room.ID
									creatorID := room.CreatorID
									hasCustomSeed := room.HasCustomSeed
									seed := ""
									if room.GameState != nil {
										seed = room.GameState.Seed
										log.Printf("RecordGameResult (chording взрыв): seed=%s (len=%d)", seed, len(seed))
									}
									room.Mu.RUnlock()
									log.Printf("[MUTEX] handleCellClick (взрыв chording): room.Mu.RUnlock() разблокирован после сбора участников")

									go func() {
										log.Printf("RecordGameResult (chording взрыв): передаем seed=%s (len=%d)", seed, len(seed))
										if err := s.profileHandler.RecordGameResult(userID, room.Cols, room.Rows, room.Mines, gameTime, false, chording, quickStart, roomID, seed, hasCustomSeed, creatorID, participants); err != nil {
											log.Printf("Ошибка записи результата игры: %v", err)
										}
										// Сохраняем комнату в БД после проигрыша
										if err := s.roomManager.SaveRoom(room); err != nil {
											log.Printf("Предупреждение: не удалось сохранить комнату %s после проигрыша (chording): %v", room.ID, err)
										}
									}()
								}

								room.GameState.Mu.Unlock()
								// Отправляем состояние игры асинхронно, чтобы не блокировать обработку
								go func() {
									s.broadcastGameState(room)
								}()
								return
							}

							// Автоматическое открытие соседних пустых ячеек
							if neighborCell.NeighborMines == 0 {
								room.GameState.RevealNeighbors(ni, nj, changedCells)
							}
						}
					}
				}
			}

			// Проверка победы
			totalCells := room.GameState.Rows * room.GameState.Cols
			if room.GameState.Revealed == totalCells-room.GameState.Mines {
				room.GameState.GameWon = true
				wsPlayer := s.getWSPlayer(playerID)
				var userID int
				if wsPlayer != nil {
					wsPlayer.mu.Lock()
					userID = wsPlayer.UserID
					wsPlayer.mu.Unlock()
				} else {
					roomPlayer := room.GetPlayer(playerID)
					if roomPlayer != nil {
						userID = roomPlayer.UserID
					}
				}

				log.Printf("[MUTEX] handleCellClick (chording победа): блокируем room.Mu.RLock() для получения gameTime")
				room.Mu.RLock()
				log.Printf("[MUTEX] handleCellClick (chording победа): room.Mu.RLock() заблокирован для получения gameTime")
				var gameTime float64
				if room.StartTime != nil {
					gameTime = time.Since(*room.StartTime).Seconds()
				}
				log.Printf("[MUTEX] handleCellClick (chording победа): разблокируем room.Mu.RUnlock() после получения gameTime")
				room.Mu.RUnlock()
				log.Printf("[MUTEX] handleCellClick (chording победа): room.Mu.RUnlock() разблокирован после получения gameTime")

				if userID > 0 {
					// Собираем список участников для записи результата
					log.Printf("[MUTEX] handleCellClick (chording победа): блокируем room.Mu.RLock() для сбора участников")
					room.Mu.RLock()
					log.Printf("[MUTEX] handleCellClick (chording победа): room.Mu.RLock() заблокирован для сбора участников")
					participants := make([]game.GameParticipant, 0)
					for _, p := range room.Players {
						if p.UserID > 0 {
							participants = append(participants, game.GameParticipant{
								UserID:   p.UserID,
								Nickname: p.Nickname,
								Color:    p.Color,
							})
						}
					}
					log.Printf("[MUTEX] handleCellClick (chording победа): разблокируем room.Mu.RUnlock() после сбора участников")
					chording := room.Chording
					quickStart := room.QuickStart
					roomID := room.ID
					creatorID := room.CreatorID
					hasCustomSeed := room.HasCustomSeed
					seed := ""
					if room.GameState != nil {
						seed = room.GameState.Seed
					}
					room.Mu.RUnlock()
					log.Printf("[MUTEX] handleCellClick (chording победа): room.Mu.RUnlock() разблокирован после сбора участников")

					go func() {
						if err := s.profileHandler.RecordGameResult(userID, room.Cols, room.Rows, room.Mines, gameTime, true, chording, quickStart, roomID, seed, hasCustomSeed, creatorID, participants); err != nil {
							log.Printf("Ошибка записи результата игры: %v", err)
						}
					}()
				}
			}

			room.GameState.Mu.Unlock()
			// Отправляем состояние игры асинхронно, чтобы не блокировать обработку
			go func() {
				s.broadcastGameState(room)
			}()
			return
		} else {
			// Chording не активирован, игнорируем клик на открытую клетку
			log.Printf("handleCellClick: chording не активирован (флагов: %d, нужно: %d), игнорируем клик", flagCount, cell.NeighborMines)
			room.GameState.Mu.Unlock()
			return
		}
	}

	// Если клик на уже открытую клетку без chording - игнорируем
	if cell.IsRevealed {
		log.Printf("handleCellClick: клик на открытую клетку без chording, игнорируем")
		room.GameState.Mu.Unlock()
		return
	}

	// Если это первое открытие, устанавливаем время начала игры
	// Примечание: StartTime нужно устанавливать через метод или работать напрямую
	isFirstClick := room.GameState.Revealed == 0
	if isFirstClick && room.StartTime == nil {
		now := time.Now()
		room.StartTime = &now
		log.Printf("StartTime установлен при первом клике: %v, Revealed=%d", now, room.GameState.Revealed)
	}

	// Для classic режима с QuickStart: делаем первую клетку нулевой
	// НО только если не используется seed (seed == 0 означает что seed не был установлен или это старая игра)
	if gameMode == "classic" && isFirstClick && room.QuickStart && room.GameState.Seed == "" {
		log.Printf("handleCellClick: QuickStart включен, делаем первую клетку нулевой (без seed)")
		room.GameState.Mu.Unlock()
		room.GameState.EnsureFirstClickSafe(row, col)
		room.GameState.Mu.Lock()
		// Обновляем ссылку на ячейку после перемещения мин
		cell = &room.GameState.Board[row][col]
	} else if gameMode == "classic" && isFirstClick && room.QuickStart && room.GameState.Seed != "" {
		log.Printf("handleCellClick: QuickStart включен, но используется seed=%s, не перемещаем мины", room.GameState.Seed)
		// При использовании seed мины уже размещены так, чтобы первая клетка была безопасной (если QuickStart был включен при создании)
		// Если первая клетка оказалась миной - это означает что QuickStart не был учтен при генерации, но мы не можем переместить мины без нарушения seed
	}

	// В режимах training и fair мины размещаются динамически при клике
	if gameMode == "training" || gameMode == "fair" {
		log.Printf("handleCellClick: режим %s, начинаем динамическое размещение мин", gameMode)
		// Разблокируем для вычисления безопасных ячеек
		room.GameState.Mu.Unlock()
		log.Printf("handleCellClick: мьютекс GameState разблокирован для determineMinePlacement")
		startTime := time.Now()
		mineGrid := s.determineMinePlacement(room, row, col)
		elapsed := time.Since(startTime)
		log.Printf("handleCellClick: determineMinePlacement завершен за %v, получена mineGrid размером %dx%d", elapsed, len(mineGrid), len(mineGrid[0]))
		room.GameState.Mu.Lock()
		log.Printf("handleCellClick: мьютекс GameState заблокирован после determineMinePlacement")

		// Применяем размещение мин и собираем измененные ячейки
		changedCells := make(map[[2]int]bool)
		for i := 0; i < room.GameState.Rows; i++ {
			for j := 0; j < room.GameState.Cols; j++ {
				if !room.GameState.Board[i][j].IsRevealed {
					oldMine := room.GameState.Board[i][j].IsMine
					room.GameState.Board[i][j].IsMine = mineGrid[i][j]
					// Если статус мины изменился, помечаем эту ячейку и всех её соседей для пересчета
					if oldMine != mineGrid[i][j] {
						changedCells[[2]int{i, j}] = true
						// Помечаем соседей для пересчета
						for di := -1; di <= 1; di++ {
							for dj := -1; dj <= 1; dj++ {
								if di == 0 && dj == 0 {
									continue
								}
								ni, nj := i+di, j+dj
								if ni >= 0 && ni < room.GameState.Rows && nj >= 0 && nj < room.GameState.Cols {
									changedCells[[2]int{ni, nj}] = true
								}
							}
						}
					}
				}
			}
		}

		// Пересчитываем соседние мины для всех измененных ячеек (включая открытые)
		for pos := range changedCells {
			i, j := pos[0], pos[1]
			if !room.GameState.Board[i][j].IsMine {
				count := 0
				for di := -1; di <= 1; di++ {
					for dj := -1; dj <= 1; dj++ {
						if di == 0 && dj == 0 {
							continue
						}
						ni, nj := i+di, j+dj
						if ni >= 0 && ni < room.GameState.Rows && nj >= 0 && nj < room.GameState.Cols {
							if room.GameState.Board[ni][nj].IsMine {
								count++
							}
						}
					}
				}
				room.GameState.Board[i][j].NeighborMines = count
			}
		}

		// Обновляем ссылку на ячейку
		cell = &room.GameState.Board[row][col]
	}

	log.Printf("handleCellClick: открываем ячейку row=%d, col=%d", row, col)

	// Собираем измененные клетки
	changedCells := make(map[[2]int]bool)
	if gameMode == "training" || gameMode == "fair" {
		// changedCells уже заполнен при размещении мин
	} else {
		changedCells[[2]int{row, col}] = true
	}

	cell.IsRevealed = true
	room.GameState.Revealed++
	changedCells[[2]int{row, col}] = true
	log.Printf("Ячейка открыта: row=%d, col=%d, isMine=%v, neighborMines=%d, revealed=%d",
		row, col, cell.IsMine, cell.NeighborMines, room.GameState.Revealed)

	if cell.IsMine {
		room.GameState.GameOver = true
		// Сохраняем информацию об игроке, который проиграл
		// Получаем информацию об игроке
		wsPlayer := s.getWSPlayer(playerID)
		var userID int
		var nickname string
		if wsPlayer != nil {
			wsPlayer.mu.Lock()
			userID = wsPlayer.UserID
			nickname = wsPlayer.Nickname
			wsPlayer.mu.Unlock()
		} else {
			roomPlayer := room.GetPlayer(playerID)
			if roomPlayer != nil {
				userID = roomPlayer.UserID
				nickname = roomPlayer.Nickname
			}
		}
		if nickname != "" {
			room.GameState.LoserPlayerID = playerID
			room.GameState.LoserNickname = nickname
		}

		// Вычисляем время игры
		log.Printf("[MUTEX] handleMineExplosion: блокируем room.Mu.RLock() для получения gameTime")
		room.Mu.RLock()
		log.Printf("[MUTEX] handleMineExplosion: room.Mu.RLock() заблокирован для получения gameTime")
		var gameTime float64
		if room.StartTime != nil {
			gameTime = time.Since(*room.StartTime).Seconds()
			log.Printf("Время игры (поражение): %.2f секунд, StartTime был: %v", gameTime, *room.StartTime)
		} else {
			// Если StartTime не установлен (не должно происходить), используем 0
			gameTime = 0.0
			log.Printf("ВНИМАНИЕ: StartTime == nil при вычислении времени игры (поражение)!")
		}
		log.Printf("[MUTEX] handleMineExplosion: разблокируем room.Mu.RUnlock() после получения gameTime")
		room.Mu.RUnlock()
		log.Printf("[MUTEX] handleMineExplosion: room.Mu.RUnlock() разблокирован после получения gameTime")

		// Записываем поражение в БД (поражения не влияют на рейтинг)
		if userID > 0 && s.profileHandler != nil {
			// Собираем список участников игры
			participants := make([]game.GameParticipant, 0)
			log.Printf("[MUTEX] handleMineExplosion: блокируем room.Mu.RLock() для сбора участников")
			room.Mu.RLock()
			log.Printf("[MUTEX] handleMineExplosion: room.Mu.RLock() заблокирован для сбора участников")
			for _, p := range room.Players {
				if p.UserID > 0 {
					participants = append(participants, game.GameParticipant{
						UserID:   p.UserID,
						Nickname: p.Nickname,
						Color:    p.Color,
					})
				}
			}
			log.Printf("[MUTEX] handleMineExplosion: разблокируем room.Mu.RUnlock() после сбора участников")
			chording := room.Chording
			quickStart := room.QuickStart
			roomID := room.ID
			creatorID := room.CreatorID
			hasCustomSeed := room.HasCustomSeed
			seed := ""
			if room.GameState != nil {
				seed = room.GameState.Seed
				log.Printf("RecordGameResult (проигрыш): seed=%s (len=%d)", seed, len(seed))
			}
			room.Mu.RUnlock()
			log.Printf("[MUTEX] handleMineExplosion: room.Mu.RUnlock() разблокирован после сбора участников")

			log.Printf("RecordGameResult (проигрыш): передаем seed=%s (len=%d)", seed, len(seed))
			if err := s.profileHandler.RecordGameResult(userID, room.Cols, room.Rows, room.Mines, gameTime, false, chording, quickStart, roomID, seed, hasCustomSeed, creatorID, participants); err != nil {
				log.Printf("Ошибка записи результата игры: %v", err)
			}
		}
		log.Printf("Игра окончена - подорвалась мина! Игрок: %s (%s)", room.GameState.LoserNickname, playerID)

		// В режиме fair вычисляем подсказки при проигрыше
		log.Printf("[MUTEX] handleMineExplosion: блокируем room.Mu.RLock() для получения gameMode")
		room.Mu.RLock()
		log.Printf("[MUTEX] handleMineExplosion: room.Mu.RLock() заблокирован для получения gameMode")
		gameMode := room.GameMode
		log.Printf("[MUTEX] handleMineExplosion: разблокируем room.Mu.RUnlock() после получения gameMode")
		room.Mu.RUnlock()
		log.Printf("[MUTEX] handleMineExplosion: room.Mu.RUnlock() разблокирован после получения gameMode")
		if gameMode == "fair" {
			room.GameState.Mu.Unlock()
			s.calculateCellHints(room)
			room.GameState.Mu.Lock()
		}

		// Отправляем сервисное сообщение о взрыве
		if nickname != "" {
			chatMsg := Message{
				Type:     "chat",
				PlayerID: playerID,
				Nickname: nickname,
				Color:    playerColor,
				Chat: &ChatMessage{
					Text:     fmt.Sprintf("%s подорвался на мине на (%d, %d) 💣", nickname, row+1, col+1),
					IsSystem: true,
					Action:   "explode",
					Row:      row,
					Col:      col,
				},
			}
			s.broadcastToAll(room, chatMsg)
		}
	} else {
		// Автоматическое открытие соседних пустых ячеек
		if cell.NeighborMines == 0 {
			log.Printf("Открытие соседних ячеек для row=%d, col=%d", row, col)
			s.revealNeighbors(room, row, col, changedCells)
		}

		// В режиме training пересчитываем подсказки асинхронно после каждого открытия (в fair подсказки только при проигрыше)
		log.Printf("[MUTEX] handleCellClick: блокируем room.Mu.RLock() для получения gameMode (training)")
		room.Mu.RLock()
		log.Printf("[MUTEX] handleCellClick: room.Mu.RLock() заблокирован для получения gameMode (training)")
		gameMode := room.GameMode
		log.Printf("[MUTEX] handleCellClick: разблокируем room.Mu.RUnlock() после получения gameMode (training)")
		room.Mu.RUnlock()
		log.Printf("[MUTEX] handleCellClick: room.Mu.RUnlock() разблокирован после получения gameMode (training)")
		if gameMode == "training" {
			// Выполняем асинхронно, чтобы не блокировать ответ
			go func() {
				// calculateCellHints сама блокирует мьютекс, не нужно блокировать здесь
				s.calculateCellHints(room)
				s.broadcastGameState(room)
			}()
		}

		// Отправляем сервисное сообщение об открытии поля
		if nickname != "" {
			chatMsg := Message{
				Type:     "chat",
				PlayerID: playerID,
				Nickname: nickname,
				Color:    playerColor,
				Chat: &ChatMessage{
					Text:     fmt.Sprintf("%s открыл поле на (%d, %d)", nickname, row+1, col+1),
					IsSystem: true,
					Action:   "reveal",
					Row:      row,
					Col:      col,
				},
			}
			s.broadcastToAll(room, chatMsg)
		}

		// Проверка победы
		totalCells := room.GameState.Rows * room.GameState.Cols
		if room.GameState.Revealed == totalCells-room.GameState.Mines {
			room.GameState.GameWon = true
			log.Printf("Победа! Все ячейки открыты!")

			// Вычисляем время игры
			log.Printf("[MUTEX] handleCellClick (победа): блокируем room.Mu.RLock() для получения gameTime")
			room.Mu.RLock()
			log.Printf("[MUTEX] handleCellClick (победа): room.Mu.RLock() заблокирован для получения gameTime")
			var gameTime float64
			if room.StartTime != nil {
				gameTime = time.Since(*room.StartTime).Seconds()
				log.Printf("Время игры (победа): %.2f секунд, StartTime был: %v", gameTime, *room.StartTime)
			} else {
				// Если StartTime не установлен (не должно происходить), используем 0
				gameTime = 0.0
				log.Printf("ВНИМАНИЕ: StartTime == nil при вычислении времени игры (победа)!")
			}
			log.Printf("[MUTEX] handleCellClick (победа): разблокируем room.Mu.RUnlock() после получения gameTime")
			room.Mu.RUnlock()
			log.Printf("[MUTEX] handleCellClick (победа): room.Mu.RUnlock() разблокирован после получения gameTime")
			loserID := room.GameState.LoserPlayerID

			// Собираем список участников игры и записываем победу
			// Делаем это в отдельной горутине, чтобы не блокировать обработку
			go func() {
				log.Printf("[MUTEX] handleCellClick (победа goroutine): блокируем room.Mu.RLock() для сбора участников")
				room.Mu.RLock()
				log.Printf("[MUTEX] handleCellClick (победа goroutine): room.Mu.RLock() заблокирован для сбора участников")
				participants := make([]game.GameParticipant, 0)
				for _, p := range room.Players {
					if p.UserID > 0 {
						participants = append(participants, game.GameParticipant{
							UserID:   p.UserID,
							Nickname: p.Nickname,
							Color:    p.Color,
						})
					}
				}

				// Записываем победу для всех игроков в комнате, которые не проиграли
				chording := room.Chording
				quickStart := room.QuickStart
				roomID := room.ID
				creatorID := room.CreatorID
				hasCustomSeed := room.HasCustomSeed
				seed := ""
				if room.GameState != nil {
					seed = room.GameState.Seed
				}
				for _, p := range room.Players {
					// Записываем победу только для игроков, которые не проиграли
					if p.ID != loserID && p.UserID > 0 && s.profileHandler != nil {
						if err := s.profileHandler.RecordGameResult(p.UserID, room.Cols, room.Rows, room.Mines, gameTime, true, chording, quickStart, roomID, seed, hasCustomSeed, creatorID, participants); err != nil {
							log.Printf("Ошибка записи результата игры: %v", err)
						}
					}
				}
				log.Printf("[MUTEX] handleCellClick (победа goroutine): разблокируем room.Mu.RUnlock() после записи результатов")
				room.Mu.RUnlock()
				log.Printf("[MUTEX] handleCellClick (победа goroutine): room.Mu.RUnlock() разблокирован после записи результатов")
				// Сохраняем комнату в БД после победы
				if err := s.roomManager.SaveRoom(room); err != nil {
					log.Printf("Предупреждение: не удалось сохранить комнату %s после победы: %v", room.ID, err)
				}
			}()
		}
	}

	log.Printf("Отправка обновленного состояния игры после клика")
	// Разблокируем мьютекс перед отправкой состояния игры
	room.GameState.Mu.Unlock()

	// Сохраняем комнату в БД после завершения игры (проигрыш)
	if room.GameState.GameOver {
		go func() {
			if err := s.roomManager.SaveRoom(room); err != nil {
				log.Printf("Предупреждение: не удалось сохранить комнату %s после проигрыша: %v", room.ID, err)
			}
		}()
	}

	// Отправляем только измененные клетки
	s.broadcastCellUpdates(room, changedCells, room.GameState.GameOver, room.GameState.GameWon, room.GameState.Revealed, room.GameState.HintsUsed, room.GameState.LoserPlayerID, room.GameState.LoserNickname)
}

// ensureFirstClickSafe обеспечивает безопасность первого клика
//
//lint:ignore U1000 Используется для отладки и тестирования
func (s *Server) ensureFirstClickSafe(room *game.Room, firstRow, firstCol int) {
	// Собираем все мины в радиусе 1 клетки от первой ячейки
	minesToMove := make([]struct{ row, col int }, 0)

	for di := -1; di <= 1; di++ {
		for dj := -1; dj <= 1; dj++ {
			ni, nj := firstRow+di, firstCol+dj
			if ni >= 0 && ni < room.GameState.Rows && nj >= 0 && nj < room.GameState.Cols {
				if room.GameState.Board[ni][nj].IsMine {
					minesToMove = append(minesToMove, struct{ row, col int }{ni, nj})
					room.GameState.Board[ni][nj].IsMine = false
				}
			}
		}
	}

	// Перемещаем мины в случайные свободные места
	for range minesToMove {
		// Ищем свободное место (не в радиусе 1 от первой ячейки и не занятое миной)
		attempts := 0
		for attempts < 100 {
			newRow := mathrand.Intn(room.GameState.Rows)
			newCol := mathrand.Intn(room.GameState.Cols)

			// Проверяем, что это не в радиусе 1 от первой ячейки
			tooClose := false
			for di := -1; di <= 1; di++ {
				for dj := -1; dj <= 1; dj++ {
					if newRow == firstRow+di && newCol == firstCol+dj {
						tooClose = true
						break
					}
				}
				if tooClose {
					break
				}
			}

			if !tooClose && !room.GameState.Board[newRow][newCol].IsMine {
				room.GameState.Board[newRow][newCol].IsMine = true
				break
			}
			attempts++
		}
	}

	// Пересчитываем соседние мины для всех ячеек
	for i := 0; i < room.GameState.Rows; i++ {
		for j := 0; j < room.GameState.Cols; j++ {
			if !room.GameState.Board[i][j].IsMine {
				count := 0
				for di := -1; di <= 1; di++ {
					for dj := -1; dj <= 1; dj++ {
						ni, nj := i+di, j+dj
						if ni >= 0 && ni < room.GameState.Rows && nj >= 0 && nj < room.GameState.Cols {
							if room.GameState.Board[ni][nj].IsMine {
								count++
							}
						}
					}
				}
				room.GameState.Board[i][j].NeighborMines = count
			}
		}
	}

	log.Printf("Мины перемещены, первая ячейка теперь безопасна")
}

func (s *Server) revealNeighbors(room *game.Room, row, col int, changedCells map[[2]int]bool) {
	for di := -1; di <= 1; di++ {
		for dj := -1; dj <= 1; dj++ {
			if di == 0 && dj == 0 {
				continue
			}
			ni, nj := row+di, col+dj
			if ni >= 0 && ni < room.GameState.Rows && nj >= 0 && nj < room.GameState.Cols {
				cell := &room.GameState.Board[ni][nj]
				if !cell.IsRevealed && !cell.IsFlagged && !cell.IsMine {
					cell.IsRevealed = true
					room.GameState.Revealed++
					changedCells[[2]int{ni, nj}] = true
					if cell.NeighborMines == 0 {
						s.revealNeighbors(room, ni, nj, changedCells)
					}
				}
			}
		}
	}
}

func (s *Server) handleHint(room *game.Room, playerID string, hint *Hint) {
	room.GameState.Mu.Lock()

	if room.GameState.GameOver || room.GameState.GameWon {
		log.Printf("Игра уже окончена, подсказка игнорируется")
		room.GameState.Mu.Unlock()
		return
	}

	// Проверяем лимит подсказок (3 подсказки глобально для комнаты)
	if room.GameState.HintsUsed >= 3 {
		log.Printf("Лимит подсказок исчерпан (использовано: %d)", room.GameState.HintsUsed)
		room.GameState.Mu.Unlock()
		return
	}

	row, col := hint.Row, hint.Col
	if row < 0 || row >= room.GameState.Rows || col < 0 || col >= room.GameState.Cols {
		log.Printf("Некорректные координаты подсказки: row=%d, col=%d", row, col)
		room.GameState.Mu.Unlock()
		return
	}

	cell := &room.GameState.Board[row][col]

	// Проверяем, что ячейка закрыта и не имеет флага
	if cell.IsRevealed || cell.IsFlagged {
		log.Printf("Ячейка уже открыта или помечена флагом: row=%d, col=%d", row, col)
		room.GameState.Mu.Unlock()
		return
	}

	// Получаем информацию об игроке для сервисных сообщений
	log.Printf("[MUTEX] handleHint: блокируем room.Mu.RLock() для комнаты, игрок %s", playerID)
	room.Mu.RLock()
	log.Printf("[MUTEX] handleHint: room.Mu.RLock() заблокирован для комнаты, игрок %s", playerID)
	player := room.Players[playerID]
	var nickname string
	var playerColor string
	if player != nil {
		nickname = player.Nickname
		playerColor = player.Color
	}
	log.Printf("[MUTEX] handleHint: разблокируем room.Mu.RUnlock() для комнаты, игрок %s", playerID)
	room.Mu.RUnlock()
	log.Printf("[MUTEX] handleHint: room.Mu.RUnlock() разблокирован для комнаты, игрок %s", playerID)

	// Если там мина - ставим флаг, иначе открываем
	if cell.IsMine {
		// Ставим флаг
		cell.IsFlagged = true
		cell.FlagColor = playerColor
		room.GameState.HintsUsed++
		log.Printf("Подсказка: поставлен флаг на мине row=%d, col=%d (использовано подсказок: %d)", row, col, room.GameState.HintsUsed)
		changedCellsHintFlag := make(map[[2]int]bool)
		changedCellsHintFlag[[2]int{row, col}] = true
		room.GameState.Mu.Unlock()
		s.broadcastCellUpdates(room, changedCellsHintFlag, room.GameState.GameOver, room.GameState.GameWon, room.GameState.Revealed, room.GameState.HintsUsed, room.GameState.LoserPlayerID, room.GameState.LoserNickname)

		// В режиме training пересчитываем подсказки асинхронно после использования подсказки
		log.Printf("[MUTEX] handleHint (flag): блокируем room.Mu.RLock() для получения gameMode")
		room.Mu.RLock()
		log.Printf("[MUTEX] handleHint (flag): room.Mu.RLock() заблокирован для получения gameMode")
		gameMode := room.GameMode
		log.Printf("[MUTEX] handleHint (flag): разблокируем room.Mu.RUnlock() после получения gameMode")
		room.Mu.RUnlock()
		log.Printf("[MUTEX] handleHint (flag): room.Mu.RUnlock() разблокирован после получения gameMode")
		if gameMode == "training" {
			go func() {
				// calculateCellHints сама блокирует мьютекс, не нужно блокировать здесь
				s.calculateCellHints(room)
				s.broadcastGameState(room)
			}()
		} else {
			// В других режимах отправляем полное состояние для обновления
			go func() {
				s.broadcastGameState(room)
			}()
		}

		// Отправляем сервисное сообщение в чат
		if nickname != "" {
			chatMsg := Message{
				Type:     "chat",
				PlayerID: playerID,
				Nickname: nickname,
				Color:    playerColor,
				Chat: &ChatMessage{
					Text:     fmt.Sprintf("%s использовал подсказку и поставил флаг на (%d, %d) 💡", nickname, row+1, col+1),
					IsSystem: true,
					Action:   "hint",
					Row:      row,
					Col:      col,
				},
			}
			s.broadcastToAll(room, chatMsg)
		}
	} else {
		// Открываем ячейку
		changedCellsHint := make(map[[2]int]bool)
		changedCellsHint[[2]int{row, col}] = true
		cell.IsRevealed = true
		room.GameState.Revealed++
		room.GameState.HintsUsed++
		log.Printf("Подсказка: открыта ячейка row=%d, col=%d, neighborMines=%d (использовано подсказок: %d)", row, col, cell.NeighborMines, room.GameState.HintsUsed)

		// Автоматическое открытие соседних пустых ячеек
		if cell.NeighborMines == 0 {
			log.Printf("Открытие соседних ячеек для row=%d, col=%d", row, col)
			s.revealNeighbors(room, row, col, changedCellsHint)
		}

		// Проверка победы
		totalCells := room.GameState.Rows * room.GameState.Cols
		if room.GameState.Revealed == totalCells-room.GameState.Mines {
			room.GameState.GameWon = true
			log.Printf("Победа! Все ячейки открыты!")

			// Вычисляем время игры
			log.Printf("[MUTEX] handleHint (победа): блокируем room.Mu.RLock() для получения gameTime")
			room.Mu.RLock()
			log.Printf("[MUTEX] handleHint (победа): room.Mu.RLock() заблокирован для получения gameTime")
			var gameTime float64
			if room.StartTime != nil {
				gameTime = time.Since(*room.StartTime).Seconds()
				log.Printf("Время игры (победа через hint): %.2f секунд, StartTime был: %v", gameTime, *room.StartTime)
			} else {
				gameTime = 0.0
				log.Printf("ВНИМАНИЕ: StartTime == nil при вычислении времени игры (победа через hint)!")
			}
			log.Printf("[MUTEX] handleHint (победа): разблокируем room.Mu.RUnlock() после получения gameTime")
			room.Mu.RUnlock()
			log.Printf("[MUTEX] handleHint (победа): room.Mu.RUnlock() разблокирован после получения gameTime")
			loserID := room.GameState.LoserPlayerID

			// Собираем список участников игры и записываем победу
			// Делаем это в отдельной горутине, чтобы не блокировать обработку
			go func() {
				log.Printf("[MUTEX] handleHint (победа goroutine): блокируем room.Mu.RLock() для сбора участников")
				room.Mu.RLock()
				log.Printf("[MUTEX] handleHint (победа goroutine): room.Mu.RLock() заблокирован для сбора участников")
				participants := make([]game.GameParticipant, 0)
				for _, p := range room.Players {
					if p.UserID > 0 {
						participants = append(participants, game.GameParticipant{
							UserID:   p.UserID,
							Nickname: p.Nickname,
							Color:    p.Color,
						})
					}
				}

				// Записываем победу для всех игроков в комнате, которые не проиграли
				chording := room.Chording
				quickStart := room.QuickStart
				roomID := room.ID
				creatorID := room.CreatorID
				hasCustomSeed := room.HasCustomSeed
				seed := ""
				if room.GameState != nil {
					seed = room.GameState.Seed
				}
				for _, p := range room.Players {
					if p.ID != loserID && p.UserID > 0 && s.profileHandler != nil {
						if err := s.profileHandler.RecordGameResult(p.UserID, room.Cols, room.Rows, room.Mines, gameTime, true, chording, quickStart, roomID, seed, hasCustomSeed, creatorID, participants); err != nil {
							log.Printf("Ошибка записи результата игры: %v", err)
						}
					}
				}
				log.Printf("[MUTEX] handleHint (победа goroutine): разблокируем room.Mu.RUnlock() после записи результатов")
				room.Mu.RUnlock()
				log.Printf("[MUTEX] handleHint (победа goroutine): room.Mu.RUnlock() разблокирован после записи результатов")
				// Сохраняем комнату в БД после победы
				if err := s.roomManager.SaveRoom(room); err != nil {
					log.Printf("Предупреждение: не удалось сохранить комнату %s после победы (hint): %v", room.ID, err)
				}
			}()
		}

		room.GameState.Mu.Unlock()
		s.broadcastCellUpdates(room, changedCellsHint, room.GameState.GameOver, room.GameState.GameWon, room.GameState.Revealed, room.GameState.HintsUsed, room.GameState.LoserPlayerID, room.GameState.LoserNickname)

		// В режиме training пересчитываем подсказки асинхронно после использования подсказки
		log.Printf("[MUTEX] handleHint (reveal): блокируем room.Mu.RLock() для получения gameMode")
		room.Mu.RLock()
		log.Printf("[MUTEX] handleHint (reveal): room.Mu.RLock() заблокирован для получения gameMode")
		gameMode := room.GameMode
		log.Printf("[MUTEX] handleHint (reveal): разблокируем room.Mu.RUnlock() после получения gameMode")
		room.Mu.RUnlock()
		log.Printf("[MUTEX] handleHint (reveal): room.Mu.RUnlock() разблокирован после получения gameMode")
		if gameMode == "training" {
			go func() {
				// calculateCellHints сама блокирует мьютекс, не нужно блокировать здесь
				s.calculateCellHints(room)
				s.broadcastGameState(room)
			}()
		} else {
			// В других режимах отправляем полное состояние для обновления
			go func() {
				s.broadcastGameState(room)
			}()
		}

		// Отправляем сервисное сообщение в чат
		if nickname != "" {
			chatMsg := Message{
				Type:     "chat",
				PlayerID: playerID,
				Nickname: nickname,
				Color:    playerColor,
				Chat: &ChatMessage{
					Text:     fmt.Sprintf("%s использовал подсказку и открыл поле на (%d, %d) 💡", nickname, row+1, col+1),
					IsSystem: true,
					Action:   "hint",
					Row:      row,
					Col:      col,
				},
			}
			s.broadcastToAll(room, chatMsg)
		}
	}
}

func (s *Server) sendGameStateToPlayer(room *game.Room, player *Player) {
	gameStateCopy := convertGameStateToMain(room.GameState)
	loserPlayerID := truncatePlayerID(gameStateCopy.LoserPlayerID)
	gameStateCopy.LoserPlayerID = loserPlayerID

	player.mu.Lock()
	defer player.mu.Unlock()

	// Кодируем gameState в protobuf формат
	binaryData, err := encodeGameStateProtobuf(gameStateCopy)
	if err != nil {
		log.Printf("Ошибка кодирования gameState: %v", err)
		return
	}

	log.Printf("Отправка gameState (protobuf): Rows=%d, Cols=%d, Mines=%d, Revealed=%d, Size=%d bytes",
		gameStateCopy.Rows, gameStateCopy.Cols, gameStateCopy.Mines, gameStateCopy.Revealed, len(binaryData))
	if err := player.Conn.WriteMessage(gorillaWS.BinaryMessage, binaryData); err != nil {
		log.Printf("Ошибка отправки состояния игры: %v", err)
	} else {
		log.Printf("Состояние игры успешно отправлено (binary)")
	}
}

func (s *Server) broadcastCellUpdates(room *game.Room, changedCells map[[2]int]bool, gameOver bool, gameWon bool, revealed int, hintsUsed int, loserPlayerID string, loserNickname string) {
	if len(changedCells) == 0 && !gameOver && !gameWon {
		// Нет изменений для отправки
		return
	}

	// Собираем обновления клеток
	updates := collectCellUpdates(room, changedCells)

	// Кодируем обновления в protobuf формат
	binaryData, err := encodeCellUpdateProtobuf(updates, gameOver, gameWon, revealed, hintsUsed, loserPlayerID, loserNickname)
	if err != nil {
		log.Printf("Ошибка кодирования обновлений клеток: %v", err)
		// Fallback: отправляем полное состояние
		s.broadcastGameState(room)
		return
	}

	log.Printf("Broadcast cell updates: %d клеток, GameOver=%v, GameWon=%v, Size=%d bytes",
		len(updates), gameOver, gameWon, len(binaryData))

	// Получаем список игроков из комнаты
	room.Mu.RLock()
	playerIDs := make([]string, 0, len(room.Players))
	for id := range room.Players {
		playerIDs = append(playerIDs, id)
	}
	room.Mu.RUnlock()

	for _, id := range playerIDs {
		wsPlayer := s.getWSPlayer(id)
		if wsPlayer != nil {
			wsPlayer.mu.Lock()
			if wsPlayer.Conn != nil {
				if err := wsPlayer.Conn.WriteMessage(gorillaWS.BinaryMessage, binaryData); err != nil {
					log.Printf("Ошибка отправки обновлений клеток игроку %s: %v", id, err)
				}
			}
			wsPlayer.mu.Unlock()
		}
	}
}

func (s *Server) broadcastGameState(room *game.Room) {
	gameStateCopy := convertGameStateToMain(room.GameState)
	loserPlayerID := truncatePlayerID(gameStateCopy.LoserPlayerID)
	gameStateCopy.LoserPlayerID = loserPlayerID

	// Кодируем gameState в protobuf формат
	binaryData, err := encodeGameStateProtobuf(gameStateCopy)
	if err != nil {
		log.Printf("Ошибка кодирования gameState: %v", err)
		return
	}

	log.Printf("Broadcast gameState (protobuf): Rows=%d, Cols=%d, Revealed=%d, GameOver=%v, GameWon=%v, Size=%d bytes",
		gameStateCopy.Rows, gameStateCopy.Cols, gameStateCopy.Revealed, gameStateCopy.GameOver, gameStateCopy.GameWon, len(binaryData))

	playersCount := room.GetPlayerCount()
	log.Printf("Отправка состояния игры %d игрокам", playersCount)

	// Получаем список игроков из комнаты и отправляем через WebSocket соединения
	log.Printf("[MUTEX] broadcastGameState: блокируем room.Mu.RLock()")
	room.Mu.RLock()
	log.Printf("[MUTEX] broadcastGameState: room.Mu.RLock() заблокирован")
	playerIDs := make([]string, 0, len(room.Players))
	for id := range room.Players {
		playerIDs = append(playerIDs, id)
	}
	log.Printf("[MUTEX] broadcastGameState: разблокируем room.Mu.RUnlock()")
	room.Mu.RUnlock()
	log.Printf("[MUTEX] broadcastGameState: room.Mu.RUnlock() разблокирован")

	for _, id := range playerIDs {
		wsPlayer := s.getWSPlayer(id)
		if wsPlayer != nil {
			wsPlayer.mu.Lock()
			if wsPlayer.Conn != nil {
				if err := wsPlayer.Conn.WriteMessage(gorillaWS.BinaryMessage, binaryData); err != nil {
					log.Printf("Ошибка отправки состояния игры игроку %s: %v", id, err)
				} else {
					log.Printf("Состояние игры отправлено игроку %s (protobuf)", id)
				}
			}
			wsPlayer.mu.Unlock()
		}
	}
}

func (s *Server) broadcastToOthers(room *game.Room, senderID string, msg Message) {
	playersCount := room.GetPlayerCount()
	if playersCount <= 1 {
		return
	}

	var binaryData []byte
	var err error
	if msg.Type == "cursor" && msg.Cursor != nil {
		binaryData, err = encodeCursorProtobuf(&msg)
		if err != nil {
			log.Printf("Ошибка кодирования курсора: %v", err)
			return
		}
	} else {
		return
	}

	// Получаем список игроков из комнаты
	log.Printf("[MUTEX] broadcastToOthers: блокируем room.Mu.RLock()")
	room.Mu.RLock()
	log.Printf("[MUTEX] broadcastToOthers: room.Mu.RLock() заблокирован")
	playerIDs := make([]string, 0, len(room.Players))
	for id := range room.Players {
		if id != senderID {
			playerIDs = append(playerIDs, id)
		}
	}
	log.Printf("[MUTEX] broadcastToOthers: разблокируем room.Mu.RUnlock()")
	room.Mu.RUnlock()
	log.Printf("[MUTEX] broadcastToOthers: room.Mu.RUnlock() разблокирован")

	sentCount := 0
	for _, id := range playerIDs {
		wsPlayer := s.getWSPlayer(id)
		if wsPlayer != nil {
			wsPlayer.mu.Lock()
			if wsPlayer.Conn != nil {
				if err := wsPlayer.Conn.WriteMessage(gorillaWS.BinaryMessage, binaryData); err != nil {
					log.Printf("Ошибка отправки сообщения игроку %s: %v", id, err)
				} else {
					sentCount++
				}
			}
			wsPlayer.mu.Unlock()
		}
	}
	log.Printf("Курсор отправлен %d игрокам (всего игроков: %d)", sentCount, playersCount)
}

func (s *Server) broadcastToAll(room *game.Room, msg Message) {
	var binaryData []byte
	var err error
	if msg.Type == "chat" && msg.Chat != nil {
		binaryData, err = encodeChatProtobuf(&msg)
		if err != nil {
			log.Printf("Ошибка кодирования чата: %v", err)
			return
		}
	} else {
		return
	}

	// Получаем список игроков из комнаты
	room.Mu.RLock()
	playerIDs := make([]string, 0, len(room.Players))
	for id := range room.Players {
		playerIDs = append(playerIDs, id)
	}
	room.Mu.RUnlock()

	for _, id := range playerIDs {
		wsPlayer := s.getWSPlayer(id)
		if wsPlayer != nil {
			wsPlayer.mu.Lock()
			if wsPlayer.Conn != nil {
				if err := wsPlayer.Conn.WriteMessage(gorillaWS.BinaryMessage, binaryData); err != nil {
					log.Printf("Ошибка отправки сообщения чата игроку %s: %v", id, err)
				}
			}
			wsPlayer.mu.Unlock()
		}
	}
}

func (s *Server) sendPlayerListToPlayer(room *game.Room, targetPlayer *Player) {
	log.Printf("[MUTEX] sendPlayerListToPlayer: блокируем room.Mu.RLock()")
	room.Mu.RLock()
	log.Printf("[MUTEX] sendPlayerListToPlayer: room.Mu.RLock() заблокирован")
	playersList := make([]map[string]string, 0, len(room.Players))
	for _, player := range room.Players {
		playersList = append(playersList, map[string]string{
			"id":       player.ID,
			"nickname": player.Nickname,
			"color":    player.Color,
		})
	}
	log.Printf("[MUTEX] sendPlayerListToPlayer: разблокируем room.Mu.RUnlock()")
	room.Mu.RUnlock()
	log.Printf("[MUTEX] sendPlayerListToPlayer: room.Mu.RUnlock() разблокирован")

	binaryData, err := encodePlayersProtobuf(playersList)
	if err != nil {
		log.Printf("Ошибка кодирования списка игроков: %v", err)
		return
	}

	targetPlayer.mu.Lock()
	defer targetPlayer.mu.Unlock()
	if targetPlayer.Conn != nil {
		if err := targetPlayer.Conn.WriteMessage(gorillaWS.BinaryMessage, binaryData); err != nil {
			log.Printf("Ошибка отправки списка игроков: %v", err)
		}
	}
}

// updateSafeCells обновляет список безопасных ячеек используя алгоритм kaboom
//
//lint:ignore U1000 Используется для отладки и тестирования
func (s *Server) updateSafeCells(room *game.Room) {
	room.GameState.Mu.Lock()
	defer room.GameState.Mu.Unlock()

	// Преобразуем Board в формат для CalculateSafeCells
	boardInfo := make([][]game.CellInfo, room.GameState.Rows)
	for i := 0; i < room.GameState.Rows; i++ {
		boardInfo[i] = make([]game.CellInfo, room.GameState.Cols)
		for j := 0; j < room.GameState.Cols; j++ {
			boardInfo[i][j] = game.CellInfo{
				IsRevealed:    room.GameState.Board[i][j].IsRevealed,
				NeighborMines: room.GameState.Board[i][j].NeighborMines,
			}
		}
	}

	// Вычисляем безопасные ячейки
	safeCellPositions := game.CalculateSafeCells(boardInfo, room.GameState.Rows, room.GameState.Cols, room.GameState.Mines)

	// Преобразуем в формат SafeCell
	room.GameState.SafeCells = make([]game.SafeCell, len(safeCellPositions))
	for i, pos := range safeCellPositions {
		room.GameState.SafeCells[i] = game.SafeCell{
			Row: pos.Row,
			Col: pos.Col,
		}
	}

	log.Printf("Обновлены безопасные ячейки: %d ячеек", len(room.GameState.SafeCells))
}

// calculateCellHints вычисляет подсказки только для ячеек на границе (в training всегда, в fair при проигрыше)
func (s *Server) calculateCellHints(room *game.Room) {
	room.GameState.Mu.Lock()
	defer room.GameState.Mu.Unlock()

	// Создаем LabelMap на основе открытых ячеек
	lm := game.NewLabelMap(room.GameState.Cols, room.GameState.Rows)

	for i := 0; i < room.GameState.Rows; i++ {
		for j := 0; j < room.GameState.Cols; j++ {
			if room.GameState.Board[i][j].IsRevealed {
				lm.SetLabel(i, j, room.GameState.Board[i][j].NeighborMines)
			}
		}
	}

	// Создаем решатель
	solver := game.MakeSolver(lm, room.GameState.Mines)

	// Вычисляем подсказки только для ячеек на границе
	hints := make([]game.CellHint, 0)
	boundary := lm.GetBoundary()

	for i, pos := range boundary {
		canBeDangerous := solver.CanBeDangerous(i)
		canBeSafe := solver.CanBeSafe(i)

		var hintType string
		if canBeDangerous && canBeSafe {
			hintType = "UNKNOWN"
		} else if canBeDangerous && !canBeSafe {
			hintType = "MINE"
		} else if !canBeDangerous && canBeSafe {
			hintType = "SAFE"
		} else {
			// Не должно происходить, но на всякий случай
			continue
		}

		hints = append(hints, game.CellHint{
			Row:  pos.Row,
			Col:  pos.Col,
			Type: hintType,
		})
	}

	room.GameState.CellHints = hints
	log.Printf("Вычислены подсказки для %d ячеек на границе", len(hints))
}

// determineMinePlacement определяет размещение мин при клике в режимах training и fair
func (s *Server) determineMinePlacement(room *game.Room, clickRow, clickCol int) [][]bool {
	log.Printf("determineMinePlacement: начало, clickRow=%d, clickCol=%d", clickRow, clickCol)

	// Проверяем QuickStart: если это первый клик и включен QuickStart, делаем клетку нулевой
	isFirstClick := room.GameState.Revealed == 0
	if isFirstClick && room.QuickStart {
		log.Printf("determineMinePlacement: QuickStart включен, делаем первую клетку нулевой")
		// Создаем сетку без мин вокруг кликнутой клетки
		mineGrid := make([][]bool, room.GameState.Rows)
		for i := 0; i < room.GameState.Rows; i++ {
			mineGrid[i] = make([]bool, room.GameState.Cols)
		}

		// Размещаем мины случайно, избегая кликнутой клетки и всех её соседей
		placed := 0
		attempts := 0
		maxAttempts := room.GameState.Rows * room.GameState.Cols * 2
		for placed < room.GameState.Mines && attempts < maxAttempts {
			row := mathrand.Intn(room.GameState.Rows)
			col := mathrand.Intn(room.GameState.Cols)
			attempts++

			// Пропускаем кликнутую клетку и все её соседи (радиус 1)
			isNearClick := false
			for di := -1; di <= 1; di++ {
				for dj := -1; dj <= 1; dj++ {
					if row == clickRow+di && col == clickCol+dj {
						isNearClick = true
						break
					}
				}
				if isNearClick {
					break
				}
			}

			if isNearClick || mineGrid[row][col] {
				continue
			}

			mineGrid[row][col] = true
			placed++
		}

		log.Printf("determineMinePlacement: QuickStart - размещено %d мин (попыток: %d)", placed, attempts)
		return mineGrid
	}

	// Создаем LabelMap на основе открытых ячеек
	lm := game.NewLabelMap(room.GameState.Cols, room.GameState.Rows)
	log.Printf("determineMinePlacement: LabelMap создан")

	revealedCount := 0
	for i := 0; i < room.GameState.Rows; i++ {
		for j := 0; j < room.GameState.Cols; j++ {
			if room.GameState.Board[i][j].IsRevealed {
				lm.SetLabel(i, j, room.GameState.Board[i][j].NeighborMines)
				revealedCount++
			}
		}
	}
	log.Printf("determineMinePlacement: установлено %d меток для открытых ячеек", revealedCount)

	// Создаем решатель
	// Важно: учитываем уже размещенные мины (только для неоткрытых ячеек)
	// Подсчитываем количество уже размещенных мин среди неоткрытых ячеек
	placedMines := 0
	for i := 0; i < room.GameState.Rows; i++ {
		for j := 0; j < room.GameState.Cols; j++ {
			if !room.GameState.Board[i][j].IsRevealed && room.GameState.Board[i][j].IsMine {
				placedMines++
			}
		}
	}

	// Общее количество мин минус уже размещенные = оставшиеся мины для размещения
	log.Printf("determineMinePlacement: room.GameState.Mines=%d, placedMines=%d", room.GameState.Mines, placedMines)
	remainingMines := room.GameState.Mines - placedMines
	if remainingMines < 0 {
		remainingMines = 0
	}
	log.Printf("determineMinePlacement: remainingMines=%d", remainingMines)

	solver := game.MakeSolver(lm, remainingMines)

	// Проверяем, на границе ли клик
	log.Printf("determineMinePlacement: проверяем, на границе ли клик")
	boundaryIdx := -1
	if clickRow >= 0 && clickRow < room.GameState.Rows && clickCol >= 0 && clickCol < room.GameState.Cols {
		boundaryIdx = lm.GetBoundaryIndex(clickRow, clickCol)
		log.Printf("determineMinePlacement: boundaryIdx=%d", boundaryIdx)
	}

	log.Printf("determineMinePlacement: проверяем HasSafeCells")
	hasSafeCells := solver.HasSafeCells()
	log.Printf("determineMinePlacement: hasSafeCells=%v", hasSafeCells)

	var shape *game.MineShape
	log.Printf("determineMinePlacement: определяем форму размещения мин")

	if boundaryIdx == -1 {
		// Клик вне границы
		outsideIsSafe := len(lm.GetBoundary()) == 0 || solver.OutsideIsSafe() || (!hasSafeCells && solver.OutsideCanBeSafe())

		if outsideIsSafe {
			// Размещаем пустую ячейку
			shape = solver.AnyShapeWithOneEmpty()
			if shape != nil {
				return shape.MineGridWithEmpty(clickRow, clickCol)
			}
		} else {
			// Размещаем мину (худший сценарий)
			shape = solver.AnyShapeWithRemaining()
			if shape != nil {
				return shape.MineGridWithMine(clickRow, clickCol)
			}
		}
	} else {
		// Клик на границе
		log.Printf("determineMinePlacement: клик на границе, boundaryIdx=%d", boundaryIdx)
		canBeSafe := solver.CanBeSafe(boundaryIdx)
		canBeDangerous := solver.CanBeDangerous(boundaryIdx)
		log.Printf("determineMinePlacement: canBeSafe=%v, canBeDangerous=%v, hasSafeCells=%v", canBeSafe, canBeDangerous, hasSafeCells)

		if canBeSafe && (!canBeDangerous || !hasSafeCells) {
			// Размещаем пустую ячейку
			log.Printf("determineMinePlacement: пытаемся получить AnySafeShape")
			shape = solver.AnySafeShape(boundaryIdx)
			if shape == nil {
				log.Printf("determineMinePlacement: AnySafeShape вернул nil")
			} else {
				log.Printf("determineMinePlacement: AnySafeShape получен")
			}
		} else {
			// Размещаем мину (худший сценарий)
			log.Printf("determineMinePlacement: пытаемся получить AnyDangerousShape")
			shape = solver.AnyDangerousShape(boundaryIdx)
			if shape == nil {
				log.Printf("determineMinePlacement: AnyDangerousShape вернул nil")
			} else {
				log.Printf("determineMinePlacement: AnyDangerousShape получен")
			}
		}
	}

	// Если не удалось получить форму, используем любую форму
	if shape == nil {
		log.Printf("determineMinePlacement: форма не получена, пытаемся получить AnyShape")
		shape = solver.AnyShape()
		if shape == nil {
			log.Printf("determineMinePlacement: AnyShape тоже вернул nil!")
		} else {
			log.Printf("determineMinePlacement: AnyShape получен")
		}
	}

	if shape != nil {
		log.Printf("determineMinePlacement: получена форма, создаем MineGrid")
		result := shape.MineGrid()
		log.Printf("determineMinePlacement: MineGrid создан, размер %dx%d", len(result), len(result[0]))

		// Подсчитываем мины в результате для отладки
		mineCount := 0
		for i := 0; i < len(result); i++ {
			for j := 0; j < len(result[i]); j++ {
				if result[i][j] {
					mineCount++
				}
			}
		}
		log.Printf("determineMinePlacement: в MineGrid размещено %d мин", mineCount)
		return result
	}

	// Fallback: создаем сетку с минами (не должно происходить, но лучше чем пустая)
	log.Printf("determineMinePlacement: WARNING - форма не получена, используем fallback с минами")
	log.Printf("determineMinePlacement: fallback - remainingMines=%d, room.GameState.Mines=%d", remainingMines, room.GameState.Mines)

	// Если remainingMines равен 0, но должны быть мины, используем общее количество мин
	minesToPlace := remainingMines
	if minesToPlace == 0 && room.GameState.Mines > 0 {
		log.Printf("determineMinePlacement: fallback - remainingMines=0, но Mines=%d, используем Mines", room.GameState.Mines)
		minesToPlace = room.GameState.Mines
	}

	mineGrid := make([][]bool, room.GameState.Rows)
	for i := 0; i < room.GameState.Rows; i++ {
		mineGrid[i] = make([]bool, room.GameState.Cols)
	}

	// Размещаем мины случайно (fallback), избегая кликнутой ячейки и уже открытых
	placed := 0
	attempts := 0
	maxAttempts := room.GameState.Rows * room.GameState.Cols * 2
	for placed < minesToPlace && attempts < maxAttempts {
		row := mathrand.Intn(room.GameState.Rows)
		col := mathrand.Intn(room.GameState.Cols)
		attempts++

		// Пропускаем кликнутую ячейку и уже открытые
		if (row == clickRow && col == clickCol) || room.GameState.Board[row][col].IsRevealed {
			continue
		}

		if !mineGrid[row][col] {
			mineGrid[row][col] = true
			placed++
		}
	}
	log.Printf("determineMinePlacement: fallback mineGrid создан с %d минами (попыток: %d, minesToPlace=%d)", placed, attempts, minesToPlace)
	return mineGrid
}

func (s *Server) broadcastPlayerList(room *game.Room) {
	log.Printf("[MUTEX] broadcastPlayerList: блокируем room.Mu.RLock()")
	room.Mu.RLock()
	log.Printf("[MUTEX] broadcastPlayerList: room.Mu.RLock() заблокирован")
	playersList := make([]map[string]string, 0, len(room.Players))
	for _, player := range room.Players {
		playersList = append(playersList, map[string]string{
			"id":       player.ID,
			"nickname": player.Nickname,
			"color":    player.Color,
		})
	}
	log.Printf("[MUTEX] broadcastPlayerList: разблокируем room.Mu.RUnlock()")
	room.Mu.RUnlock()
	log.Printf("[MUTEX] broadcastPlayerList: room.Mu.RUnlock() разблокирован")

	binaryData, err := encodePlayersProtobuf(playersList)
	if err != nil {
		log.Printf("Ошибка кодирования списка игроков: %v", err)
		return
	}

	// Получаем список игроков из комнаты
	room.Mu.RLock()
	playerIDs := make([]string, 0, len(room.Players))
	for id := range room.Players {
		playerIDs = append(playerIDs, id)
	}
	room.Mu.RUnlock()

	for _, id := range playerIDs {
		wsPlayer := s.getWSPlayer(id)
		if wsPlayer != nil {
			wsPlayer.mu.Lock()
			if wsPlayer.Conn != nil {
				if err := wsPlayer.Conn.WriteMessage(gorillaWS.BinaryMessage, binaryData); err != nil {
					log.Printf("Ошибка отправки списка игроков: %v", err)
				}
			}
			wsPlayer.mu.Unlock()
		}
	}
}

func main() {
	// Загрузка конфигурации
	cfg, err := config.ReadConfig()
	if err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}

	// Подключение к базе данных
	db, err := database.NewDB(cfg.DbHost, cfg.DbPort, cfg.DbUser, cfg.DbPassword, cfg.DbName)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Инициализация схемы БД
	if cfg.NeedMigrate {
		if err := db.InitSchema(); err != nil {
			log.Fatalf("Failed to initialize database schema: %v", err)
		}
	}

	roomManager := game.NewRoomManager()
	roomManager.SetDB(db)

	// Устанавливаем функции для кодирования/декодирования GameState
	roomManager.SetGameStateEncoder(EncodeGameStateForPersistence)
	roomManager.SetGameStateDecoder(DecodeGameStateFromPersistence)

	// Загружаем комнаты из БД при старте
	if err := roomManager.LoadRooms(); err != nil {
		log.Printf("Предупреждение: не удалось загрузить комнаты из БД: %v", err)
	}

	profileHandler := handlers.NewProfileHandler(db)
	authHandler := handlers.NewAuthHandler(db, profileHandler, cfg)
	roomHandler := handlers.NewRoomHandler(roomManager)

	// Создаем WebSocket Manager и Game Service
	// Сначала создаем временный wsManager для адаптера gameService
	tempWSManager := ws.NewManager(roomManager, profileHandler, nil)
	wsManagerAdapter := NewWSManagerAdapter(tempWSManager)
	gameService := game.NewService(roomManager, profileHandler, wsManagerAdapter)
	gameServiceAdapter := NewGameServiceAdapter(gameService)
	// Теперь создаем финальный wsManager с gameServiceAdapter
	wsManager := ws.NewManager(roomManager, profileHandler, gameServiceAdapter)
	// ВАЖНО: обновляем wsManagerAdapter, чтобы он указывал на финальный wsManager
	// Это нужно, чтобы gameService мог найти wsPlayers через правильный wsManager
	wsManagerAdapter.UpdateWSManager(wsManager)

	router := mux.NewRouter()

	r := router.PathPrefix("/api").Subrouter()
	// Публичные маршруты с опциональной авторизацией (для получения creatorID)
	r.Use(middleware.OptionalAuthMiddleware)
	r.HandleFunc("/auth/register", authHandler.Register).Methods("POST", "OPTIONS")
	r.HandleFunc("/auth/login", authHandler.Login).Methods("POST", "OPTIONS")
	r.HandleFunc("/auth/request-password-reset", authHandler.RequestPasswordReset).Methods("POST", "OPTIONS")
	r.HandleFunc("/auth/reset-password", authHandler.ResetPasswordByToken).Methods("POST", "OPTIONS")
	r.HandleFunc("/ws", wsManager.HandleWebSocket)
	r.HandleFunc("/rooms", roomHandler.GetRooms).Methods("GET", "OPTIONS")
	r.HandleFunc("/rooms", roomHandler.CreateRoom).Methods("POST", "OPTIONS")
	r.HandleFunc("/rooms/join", roomHandler.JoinRoom).Methods("POST", "OPTIONS")

	// Защищенные маршруты
	protected := router.PathPrefix("/api").Subrouter()
	protected.Use(middleware.AuthMiddleware)
	protected.HandleFunc("/auth/me", authHandler.GetMe).Methods("GET", "OPTIONS")
	protected.HandleFunc("/profile", profileHandler.GetProfile).Methods("GET", "OPTIONS")
	protected.HandleFunc("/profile/activity", profileHandler.UpdateActivity).Methods("POST", "OPTIONS")
	protected.HandleFunc("/profile/color", profileHandler.UpdateColor).Methods("POST", "OPTIONS")
	protected.HandleFunc("/profile/change-password", profileHandler.ChangePassword).Methods("POST", "OPTIONS")
	protected.HandleFunc("/auth/reset-password-admin", authHandler.ResetPasswordByAdmin).Methods("POST", "OPTIONS")
	protected.HandleFunc("/rooms/{id}", roomHandler.UpdateRoom).Methods("PUT", "OPTIONS")

	// Публичный маршрут для просмотра профиля по username
	r.HandleFunc("/profile", profileHandler.GetProfileByUsername).Methods("GET", "OPTIONS").Queries("username", "{username}")
	// Публичный маршрут для получения рейтинга
	r.HandleFunc("/leaderboard", profileHandler.GetLeaderboard).Methods("GET", "OPTIONS")
	// Публичный маршрут для получения топ-10 лучших игр по username (только с параметром username)
	r.HandleFunc("/profile/top-games", profileHandler.GetTopGames).Methods("GET", "OPTIONS").Queries("username", "{username}")
	// Защищенный маршрут для получения своих топ-10 игр (без параметра username)
	protected.HandleFunc("/profile/top-games", profileHandler.GetTopGames).Methods("GET", "OPTIONS")
	// Публичный маршрут для получения последних 10 игр по username (только с параметром username)
	r.HandleFunc("/profile/recent-games", profileHandler.GetRecentGames).Methods("GET", "OPTIONS").Queries("username", "{username}")
	// Защищенный маршрут для получения своих последних 10 игр (без параметра username)
	protected.HandleFunc("/profile/recent-games", profileHandler.GetRecentGames).Methods("GET", "OPTIONS")

	log.Printf("Сервер запущен на :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, middleware.CORSMiddleware(router)))
}

// HTTP handlers перемещены в internal/handlers/rooms.go
