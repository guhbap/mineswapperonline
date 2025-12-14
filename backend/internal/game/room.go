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

func NewRoom(id, name, password string, rows, cols, mines int, creatorID int, gameMode string, quickStart bool, chording bool, seed string, hasCustomSeed bool) *Room {
	// По умолчанию classic, если не указан
	if gameMode == "" {
		gameMode = "classic"
	}
	return &Room{
		ID:            id,
		Name:          name,
		Password:      password,
		Rows:          rows,
		Cols:          cols,
		Mines:         mines,
		GameMode:      gameMode,
		QuickStart:    quickStart,
		Chording:      chording,
		CreatorID:     creatorID,
		HasCustomSeed: hasCustomSeed,
		Players:       make(map[string]*Player),
		GameState:     NewGameState(rows, cols, mines, gameMode, seed),
		CreatedAt:     time.Now(),
	}
}

func (rm *RoomManager) CreateRoom(name, password string, rows, cols, mines int, creatorID int, gameMode string, quickStart bool, chording bool, seed string) *Room {
	roomID := utils.GenerateID()
	// Определяем, был ли seed указан пользователем явно (непустая строка означает, что он был указан)
	hasCustomSeed := seed != ""
	log.Printf("RoomManager.CreateRoom: seed=%s, hasCustomSeed=%v", seed, hasCustomSeed)
	room := NewRoom(roomID, name, password, rows, cols, mines, creatorID, gameMode, quickStart, chording, seed, hasCustomSeed)
	log.Printf("RoomManager.CreateRoom: комната создана, GameState.Seed=%s", room.GameState.Seed)
	rm.mu.Lock()
	rm.rooms[roomID] = room
	rm.mu.Unlock()

	// Сохраняем комнату в БД
	if err := rm.SaveRoom(room); err != nil {
		log.Printf("Предупреждение: не удалось сохранить комнату %s в БД: %v", roomID, err)
	}

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
		log.Printf("[MUTEX] GetRoomsList: блокируем room.Mu.RLock() для комнаты %s", room.ID)
		room.Mu.RLock()
		log.Printf("[MUTEX] GetRoomsList: room.Mu.RLock() заблокирован для комнаты %s", room.ID)
		playerCount := len(room.Players)
		log.Printf("[MUTEX] GetRoomsList: разблокируем room.Mu.RUnlock() для комнаты %s", room.ID)
		room.Mu.RUnlock()
		log.Printf("[MUTEX] GetRoomsList: room.Mu.RUnlock() разблокирован для комнаты %s", room.ID)
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

	// Удаляем комнату из БД
	if err := rm.DeleteRoomFromDB(roomID); err != nil {
		log.Printf("Предупреждение: не удалось удалить комнату %s из БД: %v", roomID, err)
	}
}

func (r *Room) ToResponse() map[string]interface{} {
	log.Printf("[MUTEX] ToResponse: блокируем room.Mu.RLock() для комнаты %s", r.ID)
	r.Mu.RLock()
	log.Printf("[MUTEX] ToResponse: room.Mu.RLock() заблокирован для комнаты %s", r.ID)
	defer func() {
		log.Printf("[MUTEX] ToResponse: разблокируем room.Mu.RUnlock() для комнаты %s", r.ID)
		r.Mu.RUnlock()
		log.Printf("[MUTEX] ToResponse: room.Mu.RUnlock() разблокирован для комнаты %s", r.ID)
	}()
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
	log.Printf("[MUTEX] IsCreator: блокируем room.Mu.RLock() для комнаты %s", r.ID)
	r.Mu.RLock()
	log.Printf("[MUTEX] IsCreator: room.Mu.RLock() заблокирован для комнаты %s", r.ID)
	defer func() {
		log.Printf("[MUTEX] IsCreator: разблокируем room.Mu.RUnlock() для комнаты %s", r.ID)
		r.Mu.RUnlock()
		log.Printf("[MUTEX] IsCreator: room.Mu.RUnlock() разблокирован для комнаты %s", r.ID)
	}()
	return r.CreatorID == userID
}

// AddPlayer добавляет игрока в комнату
func (r *Room) AddPlayer(playerID string, player *Player) {
	log.Printf("[MUTEX] AddPlayer: блокируем room.Mu.Lock() для комнаты %s, игрок %s", r.ID, playerID)
	r.Mu.Lock()
	log.Printf("[MUTEX] AddPlayer: room.Mu.Lock() заблокирован для комнаты %s, игрок %s", r.ID, playerID)
	defer func() {
		log.Printf("[MUTEX] AddPlayer: разблокируем room.Mu.Unlock() для комнаты %s, игрок %s", r.ID, playerID)
		r.Mu.Unlock()
		log.Printf("[MUTEX] AddPlayer: room.Mu.Unlock() разблокирован для комнаты %s, игрок %s", r.ID, playerID)
	}()
	r.Players[playerID] = player
}

// RemovePlayer удаляет игрока из комнаты
func (r *Room) RemovePlayer(playerID string) {
	log.Printf("[MUTEX] RemovePlayer: блокируем room.Mu.Lock() для комнаты %s, игрок %s", r.ID, playerID)
	r.Mu.Lock()
	log.Printf("[MUTEX] RemovePlayer: room.Mu.Lock() заблокирован для комнаты %s, игрок %s", r.ID, playerID)
	defer func() {
		log.Printf("[MUTEX] RemovePlayer: разблокируем room.Mu.Unlock() для комнаты %s, игрок %s", r.ID, playerID)
		r.Mu.Unlock()
		log.Printf("[MUTEX] RemovePlayer: room.Mu.Unlock() разблокирован для комнаты %s, игрок %s", r.ID, playerID)
	}()
	delete(r.Players, playerID)
}

// GetPlayerCount возвращает количество игроков
func (r *Room) GetPlayerCount() int {
	log.Printf("[MUTEX] GetPlayerCount: блокируем room.Mu.RLock() для комнаты %s", r.ID)
	r.Mu.RLock()
	log.Printf("[MUTEX] GetPlayerCount: room.Mu.RLock() заблокирован для комнаты %s", r.ID)
	defer func() {
		log.Printf("[MUTEX] GetPlayerCount: разблокируем room.Mu.RUnlock() для комнаты %s", r.ID)
		r.Mu.RUnlock()
		log.Printf("[MUTEX] GetPlayerCount: room.Mu.RUnlock() разблокирован для комнаты %s", r.ID)
	}()
	return len(r.Players)
}

// GetPlayer возвращает игрока по ID
func (r *Room) GetPlayer(playerID string) *Player {
	log.Printf("[MUTEX] GetPlayer: блокируем room.Mu.RLock() для комнаты %s, игрок %s", r.ID, playerID)
	r.Mu.RLock()
	log.Printf("[MUTEX] GetPlayer: room.Mu.RLock() заблокирован для комнаты %s, игрок %s", r.ID, playerID)
	defer func() {
		log.Printf("[MUTEX] GetPlayer: разблокируем room.Mu.RUnlock() для комнаты %s, игрок %s", r.ID, playerID)
		r.Mu.RUnlock()
		log.Printf("[MUTEX] GetPlayer: room.Mu.RUnlock() разблокирован для комнаты %s, игрок %s", r.ID, playerID)
	}()
	return r.Players[playerID]
}

// ResetGame сбрасывает игру
func (r *Room) ResetGame() {
	log.Printf("ResetGame: начало для комнаты %s", r.ID)
	log.Printf("ResetGame: пытаемся заблокировать room.Mu (Lock)")

	// Пытаемся заблокировать с таймаутом для диагностики
	locked := make(chan bool, 1)
	go func() {
		r.Mu.Lock()
		locked <- true
	}()

	// Сохраняем seed из текущего GameState, если он был указан пользователем
	var savedSeed string = ""
	if r.GameState != nil && r.HasCustomSeed {
		savedSeed = r.GameState.Seed
		log.Printf("ResetGame: сохраняем пользовательский seed=%s", savedSeed)
	}

	select {
	case <-locked:
		log.Printf("ResetGame: room.Mu успешно заблокирован (Lock), создаем новый GameState")
		r.GameState = NewGameState(r.Rows, r.Cols, r.Mines, r.GameMode, savedSeed)
		// При сбросе игры HasCustomSeed сохраняется (не сбрасывается)
		log.Printf("ResetGame: новый GameState создан с seed=%d, сбрасываем StartTime", savedSeed)
		r.StartTime = nil
		log.Printf("ResetGame: разблокируем room.Mu")
		r.Mu.Unlock()
		log.Printf("ResetGame: room.Mu разблокирован, завершено для комнаты %s", r.ID)
	case <-time.After(5 * time.Second):
		log.Printf("ResetGame: ПРЕДУПРЕЖДЕНИЕ - не удалось заблокировать room.Mu за 5 секунд! Возможен deadlock.")
		// Все равно пытаемся продолжить, но это может быть проблемой
		r.Mu.Lock()
		log.Printf("ResetGame: room.Mu наконец заблокирован после ожидания")
		r.GameState = NewGameState(r.Rows, r.Cols, r.Mines, r.GameMode, savedSeed)
		// При сбросе игры HasCustomSeed сохраняется (не сбрасывается)
		r.StartTime = nil
		r.Mu.Unlock()
		log.Printf("ResetGame: завершено для комнаты %s (с задержкой)", r.ID)
	}
}
