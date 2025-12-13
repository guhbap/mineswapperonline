package websocket

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Player представляет игрока в комнате
type Player struct {
	ID                 string
	UserID             int    // ID пользователя из БД, если авторизован
	Nickname           string
	Color              string
	Conn               *websocket.Conn
	mu                 sync.Mutex
	LastCursorX        float64
	LastCursorY        float64
	LastCursorSendTime time.Time
}

// CursorPosition представляет позицию курсора игрока
type CursorPosition struct {
	PlayerID string  `json:"pid"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
}

// Message представляет сообщение WebSocket
type Message struct {
	Type      string          `json:"type"`
	PlayerID  string          `json:"playerId,omitempty"`
	Nickname  string          `json:"nickname,omitempty"`
	Color     string          `json:"color,omitempty"`
	Cursor    *CursorPosition `json:"cursor,omitempty"`
	CellClick *CellClick      `json:"cellClick,omitempty"`
	Hint      *Hint           `json:"hint,omitempty"`
	Chat      *ChatMessage    `json:"chat,omitempty"`
}

// ChatMessage представляет сообщение в чате
type ChatMessage struct {
	Text     string `json:"text"`
	IsSystem bool   `json:"isSystem,omitempty"`
	Action   string `json:"action,omitempty"` // "flag", "reveal", "explode"
	Row      int    `json:"row,omitempty"`
	Col      int    `json:"col,omitempty"`
}

// CellClick представляет клик по ячейке
type CellClick struct {
	Row  int  `json:"row"`
	Col  int  `json:"col"`
	Flag bool `json:"flag"`
}

// Hint представляет запрос подсказки
type Hint struct {
	Row int `json:"row"`
	Col int `json:"col"`
}

