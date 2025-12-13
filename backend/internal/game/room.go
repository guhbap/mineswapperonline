package game

import (
	"log"
	"time"

	"minesweeperonline/internal/utils"
)

func NewRoomManager() *RoomManager {
	return &RoomManager{
		rooms: make(map[string]*Room),
	}
}

func NewRoom(id, name, password string, rows, cols, mines int, creatorID int, gameMode string, quickStart bool, chording bool) *Room {
	// По умолчанию classic, если не указан
	if gameMode == "" {
		gameMode = "classic"
	}
	return &Room{
		ID:         id,
		Name:       name,
		Password:   password,
		Rows:       rows,
		Cols:       cols,
		Mines:      mines,
		GameMode:   gameMode,
		QuickStart: quickStart,
		Chording:   chording,
		CreatorID:  creatorID,
		Players:    make(map[string]*Player),
		GameState:  NewGameState(rows, cols, mines, gameMode),
		CreatedAt:  time.Now(),
	}
}

func (rm *RoomManager) CreateRoom(name, password string, rows, cols, mines int, creatorID int, gameMode string, quickStart bool, chording bool) *Room {
	roomID := utils.GenerateID()
	room := NewRoom(roomID, name, password, rows, cols, mines, creatorID, gameMode, quickStart, chording)
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
		room.Mu.RLock()
		playerCount := len(room.Players)
		room.Mu.RUnlock()
		roomsList = append(roomsList, map[string]interface{}{
			"id":          room.ID,
			"name":        room.Name,
			"hasPassword": room.Password != "",
			"rows":        room.Rows,
			"cols":        room.Cols,
			"mines":       room.Mines,
			"gameMode":    room.GameMode,
			"quickStart":  room.QuickStart,
			"chording":    room.Chording,
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
	r.Mu.RLock()
	defer r.Mu.RUnlock()
	return map[string]interface{}{
		"id":          r.ID,
		"name":        r.Name,
		"hasPassword": r.Password != "",
		"rows":        r.Rows,
		"cols":        r.Cols,
		"mines":       r.Mines,
		"gameMode":    r.GameMode,
		"quickStart":  r.QuickStart,
		"chording":    r.Chording,
		"creatorId":   r.CreatorID,
		"createdAt":   r.CreatedAt,
	}
}

func (r *Room) ValidatePassword(password string) bool {
	return r.Password == "" || r.Password == password
}

// IsCreator проверяет, является ли пользователь создателем комнаты
func (r *Room) IsCreator(userID int) bool {
	r.Mu.RLock()
	defer r.Mu.RUnlock()
	return r.CreatorID == userID
}

// AddPlayer добавляет игрока в комнату
func (r *Room) AddPlayer(playerID string, player *Player) {
	r.Mu.Lock()
	defer r.Mu.Unlock()
	r.Players[playerID] = player
}

// RemovePlayer удаляет игрока из комнаты
func (r *Room) RemovePlayer(playerID string) {
	r.Mu.Lock()
	defer r.Mu.Unlock()
	delete(r.Players, playerID)
}

// GetPlayerCount возвращает количество игроков
func (r *Room) GetPlayerCount() int {
	r.Mu.RLock()
	defer r.Mu.RUnlock()
	return len(r.Players)
}

// GetPlayer возвращает игрока по ID
func (r *Room) GetPlayer(playerID string) *Player {
	r.Mu.RLock()
	defer r.Mu.RUnlock()
	return r.Players[playerID]
}

// ResetGame сбрасывает игру
func (r *Room) ResetGame() {
	log.Printf("ResetGame: блокируем room.Mu для комнаты %s", r.ID)
	r.Mu.Lock()
	log.Printf("ResetGame: room.Mu заблокирован, создаем новый GameState")
	r.GameState = NewGameState(r.Rows, r.Cols, r.Mines, r.GameMode)
	r.StartTime = nil
	log.Printf("ResetGame: новый GameState создан, разблокируем room.Mu")
	defer r.Mu.Unlock()
	log.Printf("ResetGame: завершено для комнаты %s", r.ID)
}

