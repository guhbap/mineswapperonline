package game

import (
	"sync"
	"time"
)

// FlagInfo содержит информацию об установке флага
type FlagInfo struct {
	SetTime  time.Time
	PlayerID string
}

// SafeCell представляет безопасную ячейку
type SafeCell struct {
	Row int `json:"r"`
	Col int `json:"c"`
}

// CellHint представляет подсказку для ячейки
type CellHint struct {
	Row  int    `json:"r"`
	Col  int    `json:"c"`
	Type string `json:"t"` // "MINE", "SAFE", "UNKNOWN"
}

// Cell представляет ячейку игрового поля
type Cell struct {
	IsMine        bool   `json:"m"`
	IsRevealed    bool   `json:"r"`
	IsFlagged     bool   `json:"f"`
	NeighborMines int    `json:"n"`
	FlagColor     string `json:"fc,omitempty"` // Цвет игрока, который поставил флаг
}

// GameState представляет состояние игры
type GameState struct {
	Board         [][]Cell   `json:"b"`
	Rows          int         `json:"r"`
	Cols          int         `json:"c"`
	Mines         int         `json:"m"`
	GameOver      bool        `json:"go"`
	GameWon       bool        `json:"gw"`
	Revealed      int         `json:"rv"`
	HintsUsed     int         `json:"hu"`              // Количество использованных подсказок
	SafeCells     []SafeCell  `json:"sc,omitempty"`      // Безопасные ячейки для режима без угадываний
	CellHints     []CellHint  `json:"hints,omitempty"`   // Подсказки для ячеек
	LoserPlayerID string      `json:"lpid,omitempty"`
	LoserNickname string      `json:"ln,omitempty"`
	FlagSetInfo   map[int]FlagInfo // Информация об установке флага (ключ: row*cols + col)
	Mu            sync.RWMutex     // Экспортировано для доступа из main.go
}

// Room представляет игровую комнату
type Room struct {
	ID            string             `json:"id"`
	Name          string             `json:"name"`
	Password      string             `json:"-"`
	Rows          int                `json:"rows"`
	Cols          int                `json:"cols"`
	Mines         int                `json:"mines"`
	GameMode      string             `json:"gameMode"`  // "classic", "training", "fair"
	QuickStart    bool               `json:"quickStart"` // Быстрый старт - первая клетка всегда нулевая
	Chording      bool               `json:"chording"`  // Chording - открытие соседних клеток при клике на открытую клетку с цифрой
	CreatorID     int                `json:"creatorId"`
	Players       map[string]*Player `json:"-"`        // Используется только в WebSocket контексте
	GameState     *GameState         `json:"-"`
	CreatedAt     time.Time          `json:"createdAt"`
	StartTime     *time.Time         `json:"-"`        // Время начала игры
	deleteTimer   *time.Timer        // Таймер для отложенного удаления
	deleteTimerMu sync.Mutex         // Мьютекс для безопасной работы с таймером
	Mu            sync.RWMutex       // Экспортировано для доступа из main.go
}

// Player представляет игрока (используется в контексте комнаты)
type Player struct {
	ID       string `json:"id"`
	UserID   int    `json:"userId,omitempty"`
	Nickname string `json:"nickname"`
	Color    string `json:"color"`
}

// RoomManager управляет комнатами
type RoomManager struct {
	rooms  map[string]*Room
	mu     sync.RWMutex
	server interface{} // Ссылка на сервер для доступа к DeleteRoom (используется через интерфейс)
	db     interface{} // Ссылка на базу данных для персистентности
}

