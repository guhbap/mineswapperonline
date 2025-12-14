package game

import (
	"log"

	"minesweeperonline/internal/database"
	"minesweeperonline/internal/models"
)

// SetDB устанавливает ссылку на базу данных для персистентности
func (rm *RoomManager) SetDB(db *database.DB) {
	rm.db = db
}

// SetGameStateEncoder устанавливает функцию для кодирования GameState
func (rm *RoomManager) SetGameStateEncoder(encoder GameStateEncoder) {
	rm.gameStateEncoder = encoder
}

// SetGameStateDecoder устанавливает функцию для декодирования GameState
func (rm *RoomManager) SetGameStateDecoder(decoder GameStateDecoder) {
	rm.gameStateDecoder = decoder
}

// saveRoomUnsafe сохраняет комнату в базу данных без блокировки мьютекса
// Предполагается, что вызывающий код уже удерживает блокировку room.Mu
func (rm *RoomManager) saveRoomUnsafe(room *Room) error {
	if rm.db == nil {
		return nil // БД не установлена, пропускаем сохранение
	}

	db := rm.db.(*database.DB)

	dbRoom := &models.Room{
		ID:         room.ID,
		Name:       room.Name,
		Password:   room.Password,
		Rows:       room.Rows,
		Cols:       room.Cols,
		Mines:      room.Mines,
		GameMode:   room.GameMode,
		QuickStart: room.QuickStart,
		Chording:   room.Chording,
		CreatorID:  room.CreatorID,
		CreatedAt:  room.CreatedAt,
		StartTime:  room.StartTime,
	}

	// Сохраняем GameState, если есть функция кодирования
	if rm.gameStateEncoder != nil && room.GameState != nil {
		gameStateData, err := rm.gameStateEncoder(room.GameState)
		if err != nil {
			log.Printf("Ошибка кодирования GameState для комнаты %s: %v", room.ID, err)
			// Продолжаем сохранение без GameState
		} else {
			dbRoom.GameStateData = gameStateData
			log.Printf("GameState закодирован для комнаты %s, размер: %d байт", room.ID, len(gameStateData))
		}
	}

	// Используем Save для создания или обновления
	if err := db.Save(dbRoom).Error; err != nil {
		log.Printf("Ошибка сохранения комнаты %s в БД: %v", room.ID, err)
		return err
	}

	log.Printf("Комната %s сохранена в БД (GameState: %v, StartTime: %v)", room.ID, len(dbRoom.GameStateData) > 0, dbRoom.StartTime != nil)
	return nil
}

// SaveRoom сохраняет комнату в базу данных
// Блокирует room.Mu для чтения перед сохранением
func (rm *RoomManager) SaveRoom(room *Room) error {
	room.Mu.RLock()
	defer room.Mu.RUnlock()
	return rm.saveRoomUnsafe(room)
}

// DeleteRoomFromDB удаляет комнату из базы данных
func (rm *RoomManager) DeleteRoomFromDB(roomID string) error {
	if rm.db == nil {
		return nil // БД не установлена, пропускаем удаление
	}

	db := rm.db.(*database.DB)

	if err := db.Delete(&models.Room{}, "id = ?", roomID).Error; err != nil {
		log.Printf("Ошибка удаления комнаты %s из БД: %v", roomID, err)
		return err
	}

	log.Printf("Комната %s удалена из БД", roomID)
	return nil
}

// LoadRooms загружает все комнаты из базы данных в память
func (rm *RoomManager) LoadRooms() error {
	if rm.db == nil {
		return nil // БД не установлена, пропускаем загрузку
	}

	db := rm.db.(*database.DB)

	var dbRooms []models.Room
	if err := db.Find(&dbRooms).Error; err != nil {
		log.Printf("Ошибка загрузки комнат из БД: %v", err)
		return err
	}

	rm.mu.Lock()
	defer rm.mu.Unlock()

	loadedCount := 0
	for _, dbRoom := range dbRooms {
		// Устанавливаем значения по умолчанию для старых записей
		gameMode := dbRoom.GameMode
		if gameMode == "" {
			gameMode = "classic"
		}
		
		// Создаем комнату из данных БД
		room := NewRoom(
			dbRoom.ID,
			dbRoom.Name,
			dbRoom.Password,
			dbRoom.Rows,
			dbRoom.Cols,
			dbRoom.Mines,
			dbRoom.CreatorID,
			gameMode,
			dbRoom.QuickStart,
			dbRoom.Chording,
			0, // seed=0 при загрузке из БД (seed будет восстановлен из GameStateData)
			false, // hasCustomSeed=false при загрузке из БД (по умолчанию)
		)
		room.CreatedAt = dbRoom.CreatedAt
		room.StartTime = dbRoom.StartTime

		// Восстанавливаем GameState, если есть сохраненные данные и функция декодирования
		if len(dbRoom.GameStateData) > 0 && rm.gameStateDecoder != nil {
			gameState, err := rm.gameStateDecoder(dbRoom.GameStateData)
			if err != nil {
				log.Printf("Ошибка декодирования GameState для комнаты %s: %v, создаем новое состояние", room.ID, err)
				// Оставляем новое состояние, созданное в NewRoom
			} else {
				room.GameState = gameState
				log.Printf("GameState восстановлен для комнаты %s, размер: %d байт", room.ID, len(dbRoom.GameStateData))
			}
		} else if len(dbRoom.GameStateData) > 0 {
			log.Printf("GameState данные есть для комнаты %s, но функция декодирования не установлена", room.ID)
		}

		rm.rooms[room.ID] = room
		loadedCount++
		log.Printf("Загружена комната из БД: %s (%s), GameState: %v, StartTime: %v", 
			room.ID, room.Name, room.GameState != nil && len(dbRoom.GameStateData) > 0, room.StartTime != nil)
	}

	log.Printf("Загружено комнат из БД: %d", loadedCount)
	return nil
}

