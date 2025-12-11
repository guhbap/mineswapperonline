package game

import (
	"sync"
	"time"

	"minesweeperonline/internal/utils"
)

type Player struct {
	ID       string `json:"id"`
	Nickname string `json:"nickname"`
	Color    string `json:"color"`
	mu       sync.Mutex
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

func (rm *RoomManager) CreateRoom(name, password string, rows, cols, mines int) *Room {
	roomID := utils.GenerateID()
	room := NewRoom(roomID, name, password, rows, cols, mines)
	rm.mu.Lock()
	rm.rooms[roomID] = room
	rm.mu.Unlock()
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
}

func (r *Room) ToResponse() map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return map[string]interface{}{
		"id":          r.ID,
		"name":        r.Name,
		"hasPassword":  r.Password != "",
		"rows":        r.Rows,
		"cols":        r.Cols,
		"mines":       r.Mines,
		"createdAt":   r.CreatedAt,
	}
}

func (r *Room) ValidatePassword(password string) bool {
	return r.Password == "" || r.Password == password
}

