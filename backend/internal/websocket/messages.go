package websocket

// Message представляет сообщение WebSocket
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

// CursorPosition представляет позицию курсора
type CursorPosition struct {
	PlayerID string  `json:"pid"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
}

// ChatMessage представляет сообщение чата
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

// Hint представляет подсказку
type Hint struct {
	Row int `json:"row"`
	Col int `json:"col"`
}

// GameState представляет состояние игры (используется только для JSON fallback)
type GameState struct {
	Board     [][]Cell   `json:"b"`
	Rows      int        `json:"r"`
	Cols      int        `json:"c"`
	Mines     int        `json:"m"`
	Seed      string     `json:"seed,omitempty"`
	GameOver  bool       `json:"go"`
	GameWon   bool       `json:"gw"`
	Revealed  int        `json:"rv"`
	HintsUsed int        `json:"hu"`
	SafeCells []SafeCell `json:"sc,omitempty"`
	CellHints []CellHint `json:"hints,omitempty"`
}

// Cell представляет ячейку (используется только для JSON fallback)
type Cell struct {
	IsMine        bool   `json:"m"`
	IsRevealed    bool   `json:"r"`
	IsFlagged     bool   `json:"f"`
	NeighborMines int    `json:"n"`
	FlagColor     string `json:"fc,omitempty"`
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

