package game

// CellClick представляет клик по ячейке
type CellClick struct {
	Row  int
	Col  int
	Flag bool
}

// Hint представляет подсказку
type Hint struct {
	Row int
	Col int
}

// ChatMessage представляет сообщение чата
type ChatMessage struct {
	Text     string
	IsSystem bool
	Action   string // "flag", "reveal", "explode"
	Row      int
	Col      int
}

// Message представляет сообщение WebSocket
type Message struct {
	Type      string
	PlayerID  string
	Nickname  string
	Color     string
	Cursor    *CursorPosition
	CellClick *CellClick
	Hint      *Hint
	GameState *GameState
	Chat      *ChatMessage
}

// CursorPosition представляет позицию курсора
type CursorPosition struct {
	PlayerID string
	X        float64
	Y        float64
}

