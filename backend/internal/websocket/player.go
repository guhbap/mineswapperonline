package websocket

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Player представляет игрока с WebSocket соединением
type Player struct {
	ID                 string
	UserID             int
	Nickname           string
	Color              string
	Conn               *websocket.Conn
	mu                 sync.Mutex
	LastCursorX        float64
	LastCursorY        float64
	LastCursorSendTime time.Time
}

// GetNickname возвращает никнейм игрока
func (p *Player) GetNickname() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.Nickname
}

// GetColor возвращает цвет игрока
func (p *Player) GetColor() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.Color
}

// GetUserID возвращает ID пользователя
func (p *Player) GetUserID() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.UserID
}

// SetNickname устанавливает никнейм игрока
func (p *Player) SetNickname(nickname string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Nickname = nickname
}

// UpdateCursor обновляет позицию курсора с throttling
func (p *Player) UpdateCursor(x, y float64) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	now := time.Now()
	timeSinceLastSend := now.Sub(p.LastCursorSendTime)
	dx := x - p.LastCursorX
	dy := y - p.LastCursorY
	distance := dx*dx + dy*dy

	// Отправляем только если прошло достаточно времени И позиция изменилась значительно
	if timeSinceLastSend < 100*time.Millisecond && distance < 25 {
		return false // Пропускаем это сообщение
	}

	p.LastCursorX = x
	p.LastCursorY = y
	p.LastCursorSendTime = now
	return true
}

