package game

import (
	"fmt"
	"log"
	"time"
)

// SetServer устанавливает ссылку на сервер для доступа к DeleteRoom
func (rm *RoomManager) SetServer(server interface{}) {
	rm.server = server
}

// UpdateRoom обновляет параметры комнаты
func (rm *RoomManager) UpdateRoom(roomID string, name, password string, rows, cols, mines int, gameMode string, quickStart bool, chording bool) error {
	rm.mu.RLock()
	room, exists := rm.rooms[roomID]
	rm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("room not found")
	}

	log.Printf("[MUTEX] UpdateRoom: блокируем room.Mu.Lock() для комнаты %s", roomID)
	room.Mu.Lock()
	log.Printf("[MUTEX] UpdateRoom: room.Mu.Lock() заблокирован для комнаты %s", roomID)
	defer func() {
		log.Printf("[MUTEX] UpdateRoom: разблокируем room.Mu.Unlock() для комнаты %s", roomID)
		room.Mu.Unlock()
		log.Printf("[MUTEX] UpdateRoom: room.Mu.Unlock() разблокирован для комнаты %s", roomID)
	}()

	// Обновляем параметры комнаты
	room.Name = name
	if password == "__KEEP__" {
		// Не меняем пароль
	} else {
		// Устанавливаем новый пароль (может быть пустой строкой для удаления)
		room.Password = password
	}
	room.Rows = rows
	room.Cols = cols
	room.Mines = mines
	room.GameMode = gameMode
	room.QuickStart = quickStart
	room.Chording = chording

	// Пересоздаем игровое поле с новыми параметрами
	room.GameState = NewGameState(rows, cols, mines, gameMode)
	room.StartTime = nil // Сбрасываем время начала игры

	log.Printf("Комната обновлена: %s (ID: %s, GameMode: %s, QuickStart: %v, Chording: %v)", name, roomID, gameMode, quickStart, chording)
	
	// Сохраняем обновленную комнату в БД
	// Используем saveRoomUnsafe, так как room.Mu уже заблокирован
	if err := rm.saveRoomUnsafe(room); err != nil {
		log.Printf("Предупреждение: не удалось сохранить обновленную комнату %s в БД: %v", roomID, err)
	}
	
	return nil
}

// ScheduleRoomDeletion планирует удаление комнаты через указанное время
func (rm *RoomManager) ScheduleRoomDeletion(roomID string, delay time.Duration) {
	rm.mu.RLock()
	room, exists := rm.rooms[roomID]
	rm.mu.RUnlock()

	if !exists {
		return
	}

	room.deleteTimerMu.Lock()
	defer room.deleteTimerMu.Unlock()

	// Отменяем предыдущий таймер, если он существует
	if room.deleteTimer != nil {
		room.deleteTimer.Stop()
	}

	// Создаем новый таймер
	room.deleteTimer = time.AfterFunc(delay, func() {
		// Проверяем, что комната все еще пустая перед удалением
		log.Printf("[MUTEX] ScheduleRoomDeletion timer: блокируем room.Mu.RLock() для комнаты %s", roomID)
		room.Mu.RLock()
		log.Printf("[MUTEX] ScheduleRoomDeletion timer: room.Mu.RLock() заблокирован для комнаты %s", roomID)
		playersCount := len(room.Players)
		log.Printf("[MUTEX] ScheduleRoomDeletion timer: разблокируем room.Mu.RUnlock() для комнаты %s", roomID)
		room.Mu.RUnlock()
		log.Printf("[MUTEX] ScheduleRoomDeletion timer: room.Mu.RUnlock() разблокирован для комнаты %s", roomID)

		if playersCount == 0 {
			log.Printf("Комната %s пуста более %v, удаляем", roomID, delay)
			rm.DeleteRoom(roomID)
		} else {
			log.Printf("Комната %s больше не пуста (%d игроков), отмена удаления", roomID, playersCount)
		}
	})

	log.Printf("Запланировано удаление комнаты %s через %v", roomID, delay)
}

// CancelDeletion отменяет запланированное удаление комнаты
func (r *Room) CancelDeletion() {
	r.deleteTimerMu.Lock()
	defer r.deleteTimerMu.Unlock()

	if r.deleteTimer != nil {
		r.deleteTimer.Stop()
		r.deleteTimer = nil
		log.Printf("Отмена удаления комнаты %s", r.ID)
	}
}
