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
	}

	// Используем Save для создания или обновления
	if err := db.Save(dbRoom).Error; err != nil {
		log.Printf("Ошибка сохранения комнаты %s в БД: %v", room.ID, err)
		return err
	}

	log.Printf("Комната %s сохранена в БД", room.ID)
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
		)
		room.CreatedAt = dbRoom.CreatedAt

		rm.rooms[room.ID] = room
		loadedCount++
		log.Printf("Загружена комната из БД: %s (%s)", room.ID, room.Name)
	}

	log.Printf("Загружено комнат из БД: %d", loadedCount)
	return nil
}

