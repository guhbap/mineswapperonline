package game

import (
	"time"

	"minesweeperonline/internal/utils"
)

func NewRoomManager() *RoomManager {
	return &RoomManager{
		rooms: make(map[string]*Room),
	}
}

func NewRoom(id, name, password string, rows, cols, mines int, creatorID int, gameMode string) *Room {
	// По умолчанию classic, если не указан
	if gameMode == "" {
		gameMode = "classic"
	}
	return &Room{
		ID:        id,
		Name:      name,
		Password:  password,
		Rows:      rows,
		Cols:      cols,
		Mines:     mines,
		GameMode:  gameMode,
		CreatorID: creatorID,
		Players:   make(map[string]*Player),
		GameState: NewGameState(rows, cols, mines, gameMode),
		CreatedAt: time.Now(),
	}
}

func (rm *RoomManager) CreateRoom(name, password string, rows, cols, mines int, creatorID int, gameMode string) *Room {
	roomID := utils.GenerateID()
	room := NewRoom(roomID, name, password, rows, cols, mines, creatorID, gameMode)
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
			"gameMode":    room.GameMode,
			"players":     playerCount,
			"createdAt":   room.CreatedAt,
			"creatorId":   room.CreatorID,
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
		"hasPassword": r.Password != "",
		"rows":        r.Rows,
		"cols":        r.Cols,
		"mines":       r.Mines,
		"gameMode":    r.GameMode,
		"creatorId":   r.CreatorID,
		"createdAt":   r.CreatedAt,
	}
}

func (r *Room) ValidatePassword(password string) bool {
	return r.Password == "" || r.Password == password
}

// IsCreator проверяет, является ли пользователь создателем комнаты
func (r *Room) IsCreator(userID int) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.CreatorID == userID
}

