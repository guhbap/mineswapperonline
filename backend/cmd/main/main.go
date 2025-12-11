package main

import (
	"log"
	mathrand "math/rand"
	"net/http"
	"sync"
	"time"

	"minesweeperonline/internal/config"
	"minesweeperonline/internal/database"
	"minesweeperonline/internal/handlers"
	"minesweeperonline/internal/middleware"
	"minesweeperonline/internal/utils"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Разрешаем все источники для разработки
	},
}

type Player struct {
	ID       string `json:"id"`
	Nickname string `json:"nickname"`
	Color    string `json:"color"`
	Conn     *websocket.Conn
	mu       sync.Mutex
}

type CursorPosition struct {
	PlayerID string  `json:"playerId"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
}

type GameState struct {
	Board         [][]Cell `json:"board"`
	Rows          int      `json:"rows"`
	Cols          int      `json:"cols"`
	Mines         int      `json:"mines"`
	GameOver      bool     `json:"gameOver"`
	GameWon       bool     `json:"gameWon"`
	Revealed      int      `json:"revealed"`
	LoserPlayerID string   `json:"loserPlayerId,omitempty"`
	LoserNickname string   `json:"loserNickname,omitempty"`
	mu            sync.RWMutex
}

type Cell struct {
	IsMine        bool `json:"isMine"`
	IsRevealed    bool `json:"isRevealed"`
	IsFlagged     bool `json:"isFlagged"`
	NeighborMines int  `json:"neighborMines"`
}

type Message struct {
	Type      string          `json:"type"`
	PlayerID  string          `json:"playerId,omitempty"`
	Nickname  string          `json:"nickname,omitempty"`
	Color     string          `json:"color,omitempty"`
	Cursor    *CursorPosition `json:"cursor,omitempty"`
	CellClick *CellClick      `json:"cellClick,omitempty"`
	GameState *GameState      `json:"gameState,omitempty"`
}

type CellClick struct {
	Row  int  `json:"row"`
	Col  int  `json:"col"`
	Flag bool `json:"flag"`
}

type Room struct {
	ID        string             `json:"id"`
	Name      string             `json:"name"`
	Password  string             `json:"-"`
	Rows      int                `json:"rows"`
	Cols      int                `json:"cols"`
	Mines     int                `json:"mines"`
	Players   map[string]*Player `json:"-"`
	GameState *GameState         `json:"-"`
	CreatedAt time.Time          `json:"createdAt"`
	mu        sync.RWMutex
}

type RoomManager struct {
	rooms map[string]*Room
	mu    sync.RWMutex
}

type Server struct {
	roomManager *RoomManager
}

var colors = []string{
	"#FF6B6B", "#4ECDC4", "#45B7D1", "#FFA07A", "#98D8C8",
	"#F7DC6F", "#BB8FCE", "#85C1E2", "#F8B739", "#52BE80",
}

func NewRoomManager() *RoomManager {
	return &RoomManager{
		rooms: make(map[string]*Room),
	}
}

func NewRoom(id, name, password string, rows, cols, mines int) *Room {
	return &Room{
		ID:        id,
		Name:      name,
		Password:  password,
		Rows:      rows,
		Cols:      cols,
		Mines:     mines,
		Players:   make(map[string]*Player),
		GameState: NewGameState(rows, cols, mines),
		CreatedAt: time.Now(),
	}
}

func NewServer(roomManager *RoomManager) *Server {
	return &Server{
		roomManager: roomManager,
	}
}

func NewGameState(rows, cols, mines int) *GameState {
	gs := &GameState{
		Rows:          rows,
		Cols:          cols,
		Mines:         mines,
		GameOver:      false,
		GameWon:       false,
		Revealed:      0,
		LoserPlayerID: "",
		LoserNickname: "",
		Board:         make([][]Cell, rows),
	}

	// Инициализация поля
	for i := range gs.Board {
		gs.Board[i] = make([]Cell, cols)
	}

	// Размещение мин
	mathrand.Seed(time.Now().UnixNano())
	minesPlaced := 0
	for minesPlaced < mines {
		row := mathrand.Intn(rows)
		col := mathrand.Intn(cols)
		if !gs.Board[row][col].IsMine {
			gs.Board[row][col].IsMine = true
			minesPlaced++
		}
	}

	// Подсчет соседних мин
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

func (rm *RoomManager) CreateRoom(name, password string, rows, cols, mines int) *Room {
	roomID := utils.GenerateID()
	room := NewRoom(roomID, name, password, rows, cols, mines)
	rm.mu.Lock()
	rm.rooms[roomID] = room
	rm.mu.Unlock()
	log.Printf("Создана комната: %s (ID: %s)", name, roomID)
	return room
}

func (rm *RoomManager) GetRoom(roomID string) *Room {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.rooms[roomID]
}

func (rm *RoomManager) GetRoomsList() []map[string]interface{} {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	roomsList := make([]map[string]interface{}, 0, len(rm.rooms))
	for _, room := range rm.rooms {
		room.mu.RLock()
		playerCount := len(room.Players)
		room.mu.RUnlock()
		roomsList = append(roomsList, map[string]interface{}{
			"id":          room.ID,
			"name":        room.Name,
			"hasPassword": room.Password != "",
			"rows":        room.Rows,
			"cols":        room.Cols,
			"mines":       room.Mines,
			"players":     playerCount,
			"createdAt":   room.CreatedAt,
		})
	}
	return roomsList
}

func (rm *RoomManager) DeleteRoom(roomID string) {
	rm.mu.Lock()
	delete(rm.rooms, roomID)
	rm.mu.Unlock()
	log.Printf("Комната удалена: %s", roomID)
}

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
		conn.WriteJSON(map[string]string{"error": "Room not found"})
		conn.Close()
		return
	}

	playerID := utils.GenerateID()
	color := colors[mathrand.Intn(len(colors))]

	player := &Player{
		ID:    playerID,
		Color: color,
		Conn:  conn,
	}

	room.mu.Lock()
	room.Players[playerID] = player
	room.mu.Unlock()

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
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
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

	// Обработка сообщений
	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Ошибка чтения сообщения: %v", err)
			}
			break
		}

		// Обновляем deadline при получении сообщения
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))

		log.Printf("Получено сообщение от игрока %s: тип=%s", playerID, msg.Type)

		switch msg.Type {
		case "ping":
			// Отвечаем pong на ping сообщение
			pongMsg := Message{Type: "pong"}
			player.mu.Lock()
			if err := player.Conn.WriteJSON(pongMsg); err != nil {
				log.Printf("Ошибка отправки pong игроку %s: %v", playerID, err)
			}
			player.mu.Unlock()
			continue

		case "nickname":
			player.mu.Lock()
			player.Nickname = msg.Nickname
			player.mu.Unlock()
			log.Printf("Никнейм игрока %s установлен: %s", playerID, msg.Nickname)
			s.broadcastPlayerList(room)

		case "cursor":
			if msg.Cursor != nil {
				player.mu.Lock()
				msg.PlayerID = playerID
				msg.Cursor.PlayerID = playerID
				msg.Nickname = player.Nickname
				msg.Color = player.Color
				player.mu.Unlock()
				log.Printf("Курсор от игрока %s (%s): x=%.2f, y=%.2f", playerID, msg.Nickname, msg.Cursor.X, msg.Cursor.Y)
				s.broadcastToOthers(room, playerID, msg)
			}

		case "cellClick":
			if msg.CellClick != nil {
				log.Printf("Обработка клика: row=%d, col=%d, flag=%v", msg.CellClick.Row, msg.CellClick.Col, msg.CellClick.Flag)
				s.handleCellClick(room, playerID, msg.CellClick)
				log.Printf("Клик обработан, состояние игры обновлено")
			}

		case "newGame":
			room.mu.Lock()
			room.GameState = NewGameState(room.Rows, room.Cols, room.Mines)
			room.mu.Unlock()
			log.Printf("Новая игра начата")
			s.broadcastGameState(room)
		}
	}

	// Отключение игрока
	room.mu.Lock()
	delete(room.Players, playerID)
	playersLeft := len(room.Players)
	room.mu.Unlock()

	s.broadcastPlayerList(room)
	conn.Close()
	log.Printf("Игрок отключен: %s, игроков в комнате: %d", playerID, playersLeft)

	// Удаляем комнату, если она пустая
	if playersLeft == 0 {
		s.roomManager.DeleteRoom(roomID)
	}
}

func (s *Server) handleCellClick(room *Room, playerID string, click *CellClick) {
	room.GameState.mu.Lock()

	if room.GameState.GameOver || room.GameState.GameWon {
		log.Printf("Игра уже окончена, клик игнорируется")
		room.GameState.mu.Unlock()
		return
	}

	row, col := click.Row, click.Col
	if row < 0 || row >= room.GameState.Rows || col < 0 || col >= room.GameState.Cols {
		log.Printf("Некорректные координаты: row=%d, col=%d", row, col)
		room.GameState.mu.Unlock()
		return
	}

	cell := &room.GameState.Board[row][col]

	if click.Flag {
		// Переключение флага - нельзя ставить на открытые ячейки
		if cell.IsRevealed {
			log.Printf("Нельзя поставить флаг на открытую ячейку: row=%d, col=%d", row, col)
			room.GameState.mu.Unlock()
			return
		}
		cell.IsFlagged = !cell.IsFlagged
		log.Printf("Флаг переключен: row=%d, col=%d, flagged=%v", row, col, cell.IsFlagged)
		room.GameState.mu.Unlock()
		s.broadcastGameState(room)
		return
	}

	// Открытие ячейки
	if cell.IsFlagged || cell.IsRevealed {
		log.Printf("Ячейка уже открыта или помечена флагом: row=%d, col=%d", row, col)
		room.GameState.mu.Unlock()
		return
	}

	// Если это первое открытие, убеждаемся, что ячейка безопасна (0)
	isFirstClick := room.GameState.Revealed == 0
	if isFirstClick {
		// Если первая ячейка содержит мину или имеет соседние мины, перемещаем мины
		if cell.IsMine || cell.NeighborMines > 0 {
			log.Printf("Первое открытие небезопасно, перемещаем мины: row=%d, col=%d", row, col)
			s.ensureFirstClickSafe(room, row, col)
			// Обновляем ссылку на ячейку после перемещения мин
			cell = &room.GameState.Board[row][col]
		}
	}

	cell.IsRevealed = true
	room.GameState.Revealed++
	log.Printf("Ячейка открыта: row=%d, col=%d, isMine=%v, neighborMines=%d, revealed=%d",
		row, col, cell.IsMine, cell.NeighborMines, room.GameState.Revealed)

	if cell.IsMine {
		room.GameState.GameOver = true
		// Сохраняем информацию об игроке, который проиграл
		room.mu.RLock()
		player := room.Players[playerID]
		if player != nil {
			player.mu.Lock()
			room.GameState.LoserPlayerID = playerID
			room.GameState.LoserNickname = player.Nickname
			player.mu.Unlock()
		}
		room.mu.RUnlock()
		log.Printf("Игра окончена - подорвалась мина! Игрок: %s (%s)", room.GameState.LoserNickname, playerID)
	} else {
		// Автоматическое открытие соседних пустых ячеек
		if cell.NeighborMines == 0 {
			log.Printf("Открытие соседних ячеек для row=%d, col=%d", row, col)
			s.revealNeighbors(room, row, col)
		}

		// Проверка победы
		totalCells := room.GameState.Rows * room.GameState.Cols
		if room.GameState.Revealed == totalCells-room.GameState.Mines {
			room.GameState.GameWon = true
			log.Printf("Победа! Все ячейки открыты!")
		}
	}

	log.Printf("Отправка обновленного состояния игры после клика")
	room.GameState.mu.Unlock() // Разблокируем перед broadcastGameState
	s.broadcastGameState(room)
}

func (s *Server) ensureFirstClickSafe(room *Room, firstRow, firstCol int) {
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
	mathrand.Seed(time.Now().UnixNano())
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

func (s *Server) revealNeighbors(room *Room, row, col int) {
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
					if cell.NeighborMines == 0 {
						s.revealNeighbors(room, ni, nj)
					}
				}
			}
		}
	}
}

func (s *Server) sendGameStateToPlayer(room *Room, player *Player) {
	room.GameState.mu.RLock()
	gameStateCopy := GameState{
		Rows:          room.GameState.Rows,
		Cols:          room.GameState.Cols,
		Mines:         room.GameState.Mines,
		GameOver:      room.GameState.GameOver,
		GameWon:       room.GameState.GameWon,
		Revealed:      room.GameState.Revealed,
		LoserPlayerID: room.GameState.LoserPlayerID,
		LoserNickname: room.GameState.LoserNickname,
	}
	boardCopy := make([][]Cell, len(room.GameState.Board))
	for i := range room.GameState.Board {
		boardCopy[i] = make([]Cell, len(room.GameState.Board[i]))
		copy(boardCopy[i], room.GameState.Board[i])
	}
	gameStateCopy.Board = boardCopy
	room.GameState.mu.RUnlock()

	msg := Message{
		Type:      "gameState",
		GameState: &gameStateCopy,
	}

	player.mu.Lock()
	defer player.mu.Unlock()
	log.Printf("Отправка gameState: Rows=%d, Cols=%d, Mines=%d, Revealed=%d, Board size=%d",
		gameStateCopy.Rows, gameStateCopy.Cols, gameStateCopy.Mines, gameStateCopy.Revealed, len(gameStateCopy.Board))
	if err := player.Conn.WriteJSON(msg); err != nil {
		log.Printf("Ошибка отправки состояния игры: %v", err)
	} else {
		log.Printf("Состояние игры успешно отправлено")
	}
}

func (s *Server) broadcastGameState(room *Room) {
	room.GameState.mu.RLock()
	gameStateCopy := GameState{
		Rows:          room.GameState.Rows,
		Cols:          room.GameState.Cols,
		Mines:         room.GameState.Mines,
		GameOver:      room.GameState.GameOver,
		GameWon:       room.GameState.GameWon,
		Revealed:      room.GameState.Revealed,
		LoserPlayerID: room.GameState.LoserPlayerID,
		LoserNickname: room.GameState.LoserNickname,
	}
	boardCopy := make([][]Cell, len(room.GameState.Board))
	for i := range room.GameState.Board {
		boardCopy[i] = make([]Cell, len(room.GameState.Board[i]))
		copy(boardCopy[i], room.GameState.Board[i])
	}
	gameStateCopy.Board = boardCopy
	room.GameState.mu.RUnlock()

	msg := Message{
		Type:      "gameState",
		GameState: &gameStateCopy,
	}

	log.Printf("Broadcast gameState: Rows=%d, Cols=%d, Revealed=%d, GameOver=%v, GameWon=%v",
		gameStateCopy.Rows, gameStateCopy.Cols, gameStateCopy.Revealed, gameStateCopy.GameOver, gameStateCopy.GameWon)

	room.mu.RLock()
	playersCount := len(room.Players)
	room.mu.RUnlock()

	log.Printf("Отправка состояния игры %d игрокам", playersCount)

	room.mu.RLock()
	defer room.mu.RUnlock()
	for id, player := range room.Players {
		player.mu.Lock()
		if err := player.Conn.WriteJSON(msg); err != nil {
			log.Printf("Ошибка отправки состояния игры игроку %s: %v", id, err)
		} else {
			log.Printf("Состояние игры отправлено игроку %s", id)
		}
		player.mu.Unlock()
	}
}

func (s *Server) broadcastToOthers(room *Room, senderID string, msg Message) {
	room.mu.RLock()
	playersCount := len(room.Players)
	room.mu.RUnlock()

	if playersCount <= 1 {
		log.Printf("Только один игрок, курсор не отправляется другим")
		return
	}

	room.mu.RLock()
	defer room.mu.RUnlock()
	sentCount := 0
	for id, player := range room.Players {
		if id != senderID {
			player.mu.Lock()
			if err := player.Conn.WriteJSON(msg); err != nil {
				log.Printf("Ошибка отправки сообщения игроку %s: %v", id, err)
			} else {
				sentCount++
			}
			player.mu.Unlock()
		}
	}
	log.Printf("Курсор отправлен %d игрокам (всего игроков: %d)", sentCount, playersCount)
}

func (s *Server) broadcastPlayerList(room *Room) {
	room.mu.RLock()
	playersList := make([]map[string]string, 0, len(room.Players))
	for _, player := range room.Players {
		player.mu.Lock()
		playersList = append(playersList, map[string]string{
			"id":       player.ID,
			"nickname": player.Nickname,
			"color":    player.Color,
		})
		player.mu.Unlock()
	}
	room.mu.RUnlock()

	msgData := map[string]interface{}{
		"type":    "players",
		"players": playersList,
	}

	room.mu.RLock()
	defer room.mu.RUnlock()
	for _, player := range room.Players {
		player.mu.Lock()
		if err := player.Conn.WriteJSON(msgData); err != nil {
			log.Printf("Ошибка отправки списка игроков: %v", err)
		}
		player.mu.Unlock()
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

	roomManager := NewRoomManager()
	server := NewServer(roomManager)
	authHandler := handlers.NewAuthHandler(db)
	// roomHandler := handlers.NewRoomHandler(roomManager) // Используем старые обработчики для совместимости

	router := mux.NewRouter()

	r := router.PathPrefix("/api").Subrouter()
	// Публичные маршруты
	r.HandleFunc("/auth/register", authHandler.Register).Methods("POST", "OPTIONS")
	r.HandleFunc("/auth/login", authHandler.Login).Methods("POST", "OPTIONS")
	r.HandleFunc("/ws", server.handleWebSocket)
	r.HandleFunc("/rooms", server.handleGetRooms).Methods("GET", "OPTIONS")
	r.HandleFunc("/rooms", server.handleCreateRoom).Methods("POST", "OPTIONS")
	r.HandleFunc("/rooms/join", server.handleJoinRoom).Methods("POST", "OPTIONS")

	// Защищенные маршруты
	protected := router.PathPrefix("/api").Subrouter()
	protected.Use(middleware.AuthMiddleware)
	protected.HandleFunc("/auth/me", authHandler.GetMe).Methods("GET", "OPTIONS")

	log.Printf("Сервер запущен на :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, middleware.CORSMiddleware(router)))
}

func (s *Server) handleCreateRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		utils.JSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		Name     string `json:"name"`
		Password string `json:"password"`
		Rows     int    `json:"rows"`
		Cols     int    `json:"cols"`
		Mines    int    `json:"mines"`
	}

	if err := utils.DecodeJSON(r, &req); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := utils.ValidateRoomParams(req.Name, req.Rows, req.Cols, req.Mines); err != nil {
		utils.JSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	room := s.roomManager.CreateRoom(req.Name, req.Password, req.Rows, req.Cols, req.Mines)
	utils.JSONResponse(w, http.StatusOK, map[string]interface{}{
		"id":          room.ID,
		"name":        room.Name,
		"hasPassword": room.Password != "",
		"rows":        room.Rows,
		"cols":        room.Cols,
		"mines":       room.Mines,
	})
}

func (s *Server) handleGetRooms(w http.ResponseWriter, r *http.Request) {
	rooms := s.roomManager.GetRoomsList()
	utils.JSONResponse(w, http.StatusOK, rooms)
}

func (s *Server) handleJoinRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		utils.JSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		RoomID   string `json:"roomId"`
		Password string `json:"password"`
	}

	if err := utils.DecodeJSON(r, &req); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	room := s.roomManager.GetRoom(req.RoomID)
	if room == nil {
		utils.JSONError(w, http.StatusNotFound, "Room not found")
		return
	}

	if room.Password != "" && room.Password != req.Password {
		utils.JSONError(w, http.StatusUnauthorized, "Invalid password")
		return
	}

	utils.JSONResponse(w, http.StatusOK, map[string]interface{}{
		"id":    room.ID,
		"name":  room.Name,
		"rows":  room.Rows,
		"cols":  room.Cols,
		"mines": room.Mines,
	})
}
