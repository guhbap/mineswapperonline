package main

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	mathrand "math/rand"
	"net/http"
	"os"
	"sync"
	"time"

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

type Server struct {
	players   map[string]*Player
	gameState *GameState
	mu        sync.RWMutex
}

var colors = []string{
	"#FF6B6B", "#4ECDC4", "#45B7D1", "#FFA07A", "#98D8C8",
	"#F7DC6F", "#BB8FCE", "#85C1E2", "#F8B739", "#52BE80",
}

func NewServer() *Server {
	return &Server{
		players:   make(map[string]*Player),
		gameState: NewGameState(16, 16, 40),
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

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Ошибка обновления соединения: %v", err)
		return
	}

	playerID := generateID()
	color := colors[mathrand.Intn(len(colors))]

	player := &Player{
		ID:    playerID,
		Color: color,
		Conn:  conn,
	}

	s.mu.Lock()
	s.players[playerID] = player
	s.mu.Unlock()

	log.Printf("Игрок подключен: %s", playerID)

	// Отправка начального состояния игры
	log.Printf("Отправка начального состояния игры игроку %s", playerID)
	s.sendGameState(player)
	log.Printf("Начальное состояние игры отправлено игроку %s", playerID)

	// Обработка сообщений
	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("Ошибка чтения сообщения: %v", err)
			break
		}

		log.Printf("Получено сообщение от игрока %s: тип=%s", playerID, msg.Type)

		switch msg.Type {
		case "nickname":
			player.mu.Lock()
			player.Nickname = msg.Nickname
			player.mu.Unlock()
			log.Printf("Никнейм игрока %s установлен: %s", playerID, msg.Nickname)
			s.broadcastPlayerList()

		case "cursor":
			if msg.Cursor != nil {
				player.mu.Lock()
				msg.PlayerID = playerID
				msg.Cursor.PlayerID = playerID
				msg.Nickname = player.Nickname
				msg.Color = player.Color
				player.mu.Unlock()
				log.Printf("Курсор от игрока %s (%s): x=%.2f, y=%.2f", playerID, msg.Nickname, msg.Cursor.X, msg.Cursor.Y)
				s.broadcastToOthers(playerID, msg)
			}

		case "cellClick":
			if msg.CellClick != nil {
				log.Printf("Обработка клика: row=%d, col=%d, flag=%v", msg.CellClick.Row, msg.CellClick.Col, msg.CellClick.Flag)
				s.handleCellClick(playerID, msg.CellClick)
				log.Printf("Клик обработан, состояние игры обновлено")
			}

		case "newGame":
			s.mu.Lock()
			s.gameState = NewGameState(16, 16, 40)
			s.mu.Unlock()
			log.Printf("Новая игра начата")
			s.broadcastGameState()
		}
	}

	// Отключение игрока
	s.mu.Lock()
	delete(s.players, playerID)
	s.mu.Unlock()

	s.broadcastPlayerList()
	conn.Close()
	log.Printf("Игрок отключен: %s", playerID)
}

func (s *Server) handleCellClick(playerID string, click *CellClick) {
	s.gameState.mu.Lock()

	if s.gameState.GameOver || s.gameState.GameWon {
		log.Printf("Игра уже окончена, клик игнорируется")
		s.gameState.mu.Unlock()
		return
	}

	row, col := click.Row, click.Col
	if row < 0 || row >= s.gameState.Rows || col < 0 || col >= s.gameState.Cols {
		log.Printf("Некорректные координаты: row=%d, col=%d", row, col)
		s.gameState.mu.Unlock()
		return
	}

	cell := &s.gameState.Board[row][col]

	if click.Flag {
		// Переключение флага
		cell.IsFlagged = !cell.IsFlagged
		log.Printf("Флаг переключен: row=%d, col=%d, flagged=%v", row, col, cell.IsFlagged)
		s.gameState.mu.Unlock()
		s.broadcastGameState()
		return
	}

	// Открытие ячейки
	if cell.IsFlagged || cell.IsRevealed {
		log.Printf("Ячейка уже открыта или помечена флагом: row=%d, col=%d", row, col)
		s.gameState.mu.Unlock()
		return
	}

	cell.IsRevealed = true
	s.gameState.Revealed++
	log.Printf("Ячейка открыта: row=%d, col=%d, isMine=%v, neighborMines=%d, revealed=%d",
		row, col, cell.IsMine, cell.NeighborMines, s.gameState.Revealed)

	if cell.IsMine {
		s.gameState.GameOver = true
		// Сохраняем информацию об игроке, который проиграл
		s.mu.RLock()
		player := s.players[playerID]
		if player != nil {
			player.mu.Lock()
			s.gameState.LoserPlayerID = playerID
			s.gameState.LoserNickname = player.Nickname
			player.mu.Unlock()
		}
		s.mu.RUnlock()
		log.Printf("Игра окончена - подорвалась мина! Игрок: %s (%s)", s.gameState.LoserNickname, playerID)
	} else {
		// Автоматическое открытие соседних пустых ячеек
		if cell.NeighborMines == 0 {
			log.Printf("Открытие соседних ячеек для row=%d, col=%d", row, col)
			s.revealNeighbors(row, col)
		}

		// Проверка победы
		totalCells := s.gameState.Rows * s.gameState.Cols
		if s.gameState.Revealed == totalCells-s.gameState.Mines {
			s.gameState.GameWon = true
			log.Printf("Победа! Все ячейки открыты!")
		}
	}

	log.Printf("Отправка обновленного состояния игры после клика")
	s.gameState.mu.Unlock() // Разблокируем перед broadcastGameState
	s.broadcastGameState()
}

func (s *Server) revealNeighbors(row, col int) {
	for di := -1; di <= 1; di++ {
		for dj := -1; dj <= 1; dj++ {
			if di == 0 && dj == 0 {
				continue
			}
			ni, nj := row+di, col+dj
			if ni >= 0 && ni < s.gameState.Rows && nj >= 0 && nj < s.gameState.Cols {
				cell := &s.gameState.Board[ni][nj]
				if !cell.IsRevealed && !cell.IsFlagged && !cell.IsMine {
					cell.IsRevealed = true
					s.gameState.Revealed++
					if cell.NeighborMines == 0 {
						s.revealNeighbors(ni, nj)
					}
				}
			}
		}
	}
}

func (s *Server) sendGameState(player *Player) {
	s.gameState.mu.RLock()
	gameStateCopy := GameState{
		Rows:          s.gameState.Rows,
		Cols:          s.gameState.Cols,
		Mines:         s.gameState.Mines,
		GameOver:      s.gameState.GameOver,
		GameWon:       s.gameState.GameWon,
		Revealed:      s.gameState.Revealed,
		LoserPlayerID: s.gameState.LoserPlayerID,
		LoserNickname: s.gameState.LoserNickname,
	}
	boardCopy := make([][]Cell, len(s.gameState.Board))
	for i := range s.gameState.Board {
		boardCopy[i] = make([]Cell, len(s.gameState.Board[i]))
		copy(boardCopy[i], s.gameState.Board[i])
	}
	gameStateCopy.Board = boardCopy
	s.gameState.mu.RUnlock()

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

func (s *Server) broadcastGameState() {
	s.gameState.mu.RLock()
	gameStateCopy := GameState{
		Rows:          s.gameState.Rows,
		Cols:          s.gameState.Cols,
		Mines:         s.gameState.Mines,
		GameOver:      s.gameState.GameOver,
		GameWon:       s.gameState.GameWon,
		Revealed:      s.gameState.Revealed,
		LoserPlayerID: s.gameState.LoserPlayerID,
		LoserNickname: s.gameState.LoserNickname,
	}
	boardCopy := make([][]Cell, len(s.gameState.Board))
	for i := range s.gameState.Board {
		boardCopy[i] = make([]Cell, len(s.gameState.Board[i]))
		copy(boardCopy[i], s.gameState.Board[i])
	}
	gameStateCopy.Board = boardCopy
	s.gameState.mu.RUnlock()

	msg := Message{
		Type:      "gameState",
		GameState: &gameStateCopy,
	}

	log.Printf("Broadcast gameState: Rows=%d, Cols=%d, Revealed=%d, GameOver=%v, GameWon=%v",
		gameStateCopy.Rows, gameStateCopy.Cols, gameStateCopy.Revealed, gameStateCopy.GameOver, gameStateCopy.GameWon)

	s.mu.RLock()
	playersCount := len(s.players)
	s.mu.RUnlock()

	log.Printf("Отправка состояния игры %d игрокам", playersCount)

	s.mu.RLock()
	defer s.mu.RUnlock()
	for id, player := range s.players {
		player.mu.Lock()
		if err := player.Conn.WriteJSON(msg); err != nil {
			log.Printf("Ошибка отправки состояния игры игроку %s: %v", id, err)
		} else {
			log.Printf("Состояние игры отправлено игроку %s", id)
		}
		player.mu.Unlock()
	}
}

func (s *Server) broadcastToOthers(senderID string, msg Message) {
	s.mu.RLock()
	playersCount := len(s.players)
	s.mu.RUnlock()

	if playersCount <= 1 {
		log.Printf("Только один игрок, курсор не отправляется другим")
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	sentCount := 0
	for id, player := range s.players {
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

func (s *Server) broadcastPlayerList() {
	s.mu.RLock()
	playersList := make([]map[string]string, 0, len(s.players))
	for _, player := range s.players {
		player.mu.Lock()
		playersList = append(playersList, map[string]string{
			"id":       player.ID,
			"nickname": player.Nickname,
			"color":    player.Color,
		})
		player.mu.Unlock()
	}
	s.mu.RUnlock()

	msgData := map[string]interface{}{
		"type":    "players",
		"players": playersList,
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, player := range s.players {
		player.mu.Lock()
		if err := player.Conn.WriteJSON(msgData); err != nil {
			log.Printf("Ошибка отправки списка игроков: %v", err)
		}
		player.mu.Unlock()
	}
}

func generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func main() {
	server := NewServer()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r := mux.NewRouter()
	r.HandleFunc("/api/ws", server.handleWebSocket)
	// r.PathPrefix("/").Handler(http.FileServer(http.Dir("../frontend/dist/")))

	// CORS middleware
	corsMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	}

	log.Printf("Сервер запущен на :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, corsMiddleware(r)))
}
