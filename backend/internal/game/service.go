package game

import (
	"fmt"
	"log"
	"time"
)

// ProfileHandler интерфейс для работы с профилями
type ProfileHandler interface {
	RecordGameResult(userID, cols, rows, mines int, gameTime float64, won bool, chording, quickStart bool, roomID, seed string, hasCustomSeed bool, creatorID int, participants []GameParticipant) error
}

// Service обрабатывает игровую логику
type Service struct {
	roomManager    *RoomManager
	profileHandler ProfileHandler
	wsManager      WSManager
}

// WSPlayer интерфейс для WebSocket игрока
type WSPlayer interface {
	GetNickname() string
	GetColor() string
	GetUserID() int
	GetMu() interface{}
	GetConn() interface{}
	SetNickname(nickname string)
	UpdateCursor(x, y float64) bool
}

// WSManager интерфейс для доступа к WebSocket менеджеру
type WSManager interface {
	GetWSPlayer(playerID string) WSPlayer
}

// NewService создает новый сервис игровой логики
func NewService(roomManager *RoomManager, profileHandler ProfileHandler, wsManager WSManager) *Service {
	return &Service{
		roomManager:    roomManager,
		profileHandler: profileHandler,
		wsManager:      wsManager,
	}
}

// HandleCellClick обрабатывает клик по ячейке
func (s *Service) HandleCellClick(room *Room, playerID string, click *CellClick) error {
	log.Printf("[GAME] HandleCellClick: начало, playerID=%s, row=%d, col=%d, flag=%v", playerID, click.Row, click.Col, click.Flag)
	
	room.GameState.Mu.Lock()
	log.Printf("[GAME] HandleCellClick: мьютекс заблокирован")

	if room.GameState.GameOver || room.GameState.GameWon {
		log.Printf("[GAME] HandleCellClick: игра уже окончена (GameOver=%v, GameWon=%v), клик игнорируется", room.GameState.GameOver, room.GameState.GameWon)
		room.GameState.Mu.Unlock()
		return nil
	}

	row, col := click.Row, click.Col
	if row < 0 || row >= room.GameState.Rows || col < 0 || col >= room.GameState.Cols {
		log.Printf("[GAME] HandleCellClick: некорректные координаты: row=%d, col=%d (размеры: rows=%d, cols=%d)", row, col, room.GameState.Rows, room.GameState.Cols)
		room.GameState.Mu.Unlock()
		return fmt.Errorf("invalid coordinates")
	}

	cell := &room.GameState.Board[row][col]
	log.Printf("[GAME] HandleCellClick: ячейка найдена, isRevealed=%v, isFlagged=%v, isMine=%v, neighborMines=%d", cell.IsRevealed, cell.IsFlagged, cell.IsMine, cell.NeighborMines)

	// Получаем информацию об игроке
	room.Mu.RLock()
	player := room.Players[playerID]
	var nickname string
	var playerColor string
	if player != nil {
		nickname = player.Nickname
		playerColor = player.Color
	}
	room.Mu.RUnlock()

	if click.Flag {
		log.Printf("[GAME] HandleCellClick: обработка флага")
		// handleFlagToggle разблокирует мьютекс сам перед возвратом
		return s.handleFlagToggle(room, playerID, row, col, cell, nickname, playerColor)
	}

	log.Printf("[GAME] HandleCellClick: обработка открытия ячейки")
	// handleCellReveal разблокирует мьютекс сам перед возвратом
	return s.handleCellReveal(room, playerID, row, col, cell, nickname, playerColor)
}

// handleFlagToggle обрабатывает переключение флага
// ВАЖНО: эта функция должна разблокировать room.GameState.Mu перед возвратом
func (s *Service) handleFlagToggle(room *Room, playerID string, row, col int, cell *Cell, nickname, playerColor string) error {
	log.Printf("[GAME] handleFlagToggle: начало, row=%d, col=%d", row, col)
	if cell.IsRevealed {
		log.Printf("[GAME] handleFlagToggle: нельзя поставить флаг на открытую ячейку: row=%d, col=%d", row, col)
		room.GameState.Mu.Unlock()
		return nil
	}

	wasFlagged := cell.IsFlagged
	cellKey := row*room.GameState.Cols + col
	now := time.Now()

	if wasFlagged {
		if flagInfo, exists := room.GameState.FlagSetInfo[cellKey]; exists {
			if flagInfo.PlayerID != playerID {
				timeSinceFlagSet := now.Sub(flagInfo.SetTime)
				if timeSinceFlagSet < 1*time.Second {
					log.Printf("[GAME] handleFlagToggle: нельзя снять флаг сразу после установки другим игроком: row=%d, col=%d", row, col)
					room.GameState.Mu.Unlock()
					return nil
				}
			}
		}
		delete(room.GameState.FlagSetInfo, cellKey)
		cell.FlagColor = ""
	} else {
		room.GameState.FlagSetInfo[cellKey] = FlagInfo{
			SetTime:  now,
			PlayerID: playerID,
		}
		cell.FlagColor = playerColor
	}

	cell.IsFlagged = !cell.IsFlagged
	log.Printf("[GAME] handleFlagToggle: флаг переключен: row=%d, col=%d, flagged=%v", row, col, cell.IsFlagged)

	gameMode := room.GameMode
	room.GameState.Mu.Unlock()
	log.Printf("[GAME] handleFlagToggle: мьютекс разблокирован, отправка BroadcastGameState")

	go func() {
		log.Printf("[GAME] handleFlagToggle: запуск BroadcastGameState в горутине")
		s.BroadcastGameState(room)
	}()

	if gameMode == "training" {
		go func() {
			s.CalculateCellHints(room)
			s.BroadcastGameState(room)
		}()
	}

	if nickname != "" {
		action := "поставил флаг"
		if wasFlagged {
			action = "убрал флаг"
		}
		chatMsg := Message{
			Type:     "chat",
			PlayerID: playerID,
			Nickname: nickname,
			Color:    playerColor,
			Chat: &ChatMessage{
				Text:     fmt.Sprintf("%s %s на (%d, %d)", nickname, action, row+1, col+1),
				IsSystem: true,
				Action:   "flag",
				Row:      row,
				Col:      col,
			},
		}
		s.BroadcastToAll(room, chatMsg)
	}

	return nil
}

// handleCellReveal обрабатывает открытие ячейки
func (s *Service) handleCellReveal(room *Room, playerID string, row, col int, cell *Cell, nickname, playerColor string) error {
	log.Printf("[GAME] handleCellReveal: начало, row=%d, col=%d", row, col)
	if cell.IsFlagged {
		log.Printf("[GAME] handleCellReveal: нельзя открыть ячейку с флагом: row=%d, col=%d", row, col)
		room.GameState.Mu.Unlock()
		return nil
	}

	gameMode := room.GameMode

	// Chording: если клик на открытую клетку с цифрой
	if room.Chording && cell.IsRevealed && cell.NeighborMines > 0 {
		return s.handleChording(room, playerID, row, col, cell, nickname, playerColor)
	}

	if cell.IsRevealed {
		log.Printf("[GAME] handleCellReveal: клик на открытую клетку без chording, игнорируем")
		// Разблокируем мьютекс перед возвратом
		room.GameState.Mu.Unlock()
		return nil
	}

	// Если это первое открытие, устанавливаем время начала игры
	isFirstClick := room.GameState.Revealed == 0
	if isFirstClick && room.StartTime == nil {
		now := time.Now()
		room.StartTime = &now
		log.Printf("StartTime установлен при первом клике: %v", now)
	}

	// Для classic режима с QuickStart: делаем первую клетку нулевой
	// Применяем QuickStart всегда, когда он включен, независимо от seed
	if gameMode == "classic" && isFirstClick && room.QuickStart {
		log.Printf("[GAME] handleCellReveal: QuickStart включен, делаем первую клетку нулевой (seed=%s)", room.GameState.Seed)
		room.GameState.Mu.Unlock()
		room.GameState.EnsureFirstClickSafe(row, col)
		room.GameState.Mu.Lock()
		cell = &room.GameState.Board[row][col]
		log.Printf("[GAME] handleCellReveal: QuickStart применен, cell.isMine=%v, neighborMines=%d", cell.IsMine, cell.NeighborMines)
	}

	// В режимах training и fair мины размещаются динамически при клике
	var changedCells map[[2]int]bool
	if gameMode == "training" || gameMode == "fair" {
		changedCells = s.handleDynamicMinePlacement(room, row, col)
		cell = &room.GameState.Board[row][col]
	} else {
		changedCells = make(map[[2]int]bool)
		changedCells[[2]int{row, col}] = true
	}

	// Открываем ячейку
	cell.IsRevealed = true
	room.GameState.Revealed++
	changedCells[[2]int{row, col}] = true
	log.Printf("Ячейка открыта: row=%d, col=%d, isMine=%v", row, col, cell.IsMine)

	if cell.IsMine {
		room.GameState.Mu.Unlock()
		err := s.handleMineExplosion(room, playerID, row, col, nickname, playerColor)
		if err != nil {
			return err
		}
		// Отправляем обновления после взрыва
		s.BroadcastCellUpdates(room, changedCells, room.GameState.GameOver, room.GameState.GameWon, room.GameState.Revealed, room.GameState.HintsUsed, room.GameState.LoserPlayerID, room.GameState.LoserNickname)
		return nil
	}

	// Автоматическое открытие соседних пустых ячеек
	if cell.NeighborMines == 0 {
		s.revealNeighbors(room, row, col, changedCells)
	}

	// В режиме training пересчитываем подсказки асинхронно
	if gameMode == "training" {
		go func() {
			s.CalculateCellHints(room)
			s.BroadcastGameState(room)
		}()
	}

	// Отправляем сервисное сообщение об открытии поля
	if nickname != "" {
		chatMsg := Message{
			Type:     "chat",
			PlayerID: playerID,
			Nickname: nickname,
			Color:    playerColor,
			Chat: &ChatMessage{
				Text:     fmt.Sprintf("%s открыл поле на (%d, %d)", nickname, row+1, col+1),
				IsSystem: true,
				Action:   "reveal",
				Row:      row,
				Col:      col,
			},
		}
		s.BroadcastToAll(room, chatMsg)
	}

	// Проверка победы
	totalCells := room.GameState.Rows * room.GameState.Cols
	if room.GameState.Revealed == totalCells-room.GameState.Mines {
		room.GameState.GameWon = true
		log.Printf("Победа! Все ячейки открыты!")
		s.handleGameWin(room, playerID)
	}

	// Сохраняем значения перед разблокировкой мьютекса
	gameOver := room.GameState.GameOver
	gameWon := room.GameState.GameWon
	revealed := room.GameState.Revealed
	hintsUsed := room.GameState.HintsUsed
	loserPlayerID := room.GameState.LoserPlayerID
	loserNickname := room.GameState.LoserNickname

	room.GameState.Mu.Unlock()

	// Сохраняем комнату в БД после завершения игры
	if gameOver {
		go func() {
			if err := s.roomManager.SaveRoom(room); err != nil {
				log.Printf("Предупреждение: не удалось сохранить комнату %s после проигрыша: %v", room.ID, err)
			}
		}()
	}

	// Отправляем только измененные клетки
	log.Printf("[GAME] handleCellReveal: отправка BroadcastCellUpdates, changedCells=%d, gameOver=%v, gameWon=%v", len(changedCells), gameOver, gameWon)
	s.BroadcastCellUpdates(room, changedCells, gameOver, gameWon, revealed, hintsUsed, loserPlayerID, loserNickname)
	log.Printf("[GAME] handleCellReveal: BroadcastCellUpdates завершен")

	return nil
}

// handleChording обрабатывает chording (клик на открытую клетку с цифрой)
// ВАЖНО: эта функция должна разблокировать room.GameState.Mu перед возвратом, так как мьютекс был заблокирован в HandleCellClick
func (s *Service) handleChording(room *Room, playerID string, row, col int, cell *Cell, nickname, playerColor string) error {
	log.Printf("[GAME] handleChording: начало, row=%d, col=%d", row, col)
	flagCount := 0
	for di := -1; di <= 1; di++ {
		for dj := -1; dj <= 1; dj++ {
			if di == 0 && dj == 0 {
				continue
			}
			ni, nj := row+di, col+dj
			if ni >= 0 && ni < room.GameState.Rows && nj >= 0 && nj < room.GameState.Cols {
				if room.GameState.Board[ni][nj].IsFlagged {
					flagCount++
				}
			}
		}
	}

	if flagCount != cell.NeighborMines {
		log.Printf("[GAME] handleChording: не активирован (флагов: %d, нужно: %d)", flagCount, cell.NeighborMines)
		// Разблокируем мьютекс перед возвратом
		room.GameState.Mu.Unlock()
		return nil
	}

	log.Printf("[GAME] handleChording: активирован, открываем соседние клетки")
	changedCells := make(map[[2]int]bool)
	for di := -1; di <= 1; di++ {
		for dj := -1; dj <= 1; dj++ {
			if di == 0 && dj == 0 {
				continue
			}
			ni, nj := row+di, col+dj
			if ni >= 0 && ni < room.GameState.Rows && nj >= 0 && nj < room.GameState.Cols {
				neighborCell := &room.GameState.Board[ni][nj]
				if !neighborCell.IsRevealed && !neighborCell.IsFlagged {
					neighborCell.IsRevealed = true
					room.GameState.Revealed++
					changedCells[[2]int{ni, nj}] = true

					if neighborCell.IsMine {
						room.GameState.GameOver = true
						s.setLoserInfo(room, playerID)
						s.recordGameResult(room, playerID, false)
						room.GameState.Mu.Unlock()
						go func() {
							s.BroadcastGameState(room)
						}()
						return nil
					}

					if neighborCell.NeighborMines == 0 {
						s.revealNeighbors(room, ni, nj, changedCells)
					}
				}
			}
		}
	}

	// Проверка победы
	totalCells := room.GameState.Rows * room.GameState.Cols
	if room.GameState.Revealed == totalCells-room.GameState.Mines {
		room.GameState.GameWon = true
		s.handleGameWin(room, playerID)
	}

	// Разблокируем мьютекс перед отправкой обновлений
	room.GameState.Mu.Unlock()
	go func() {
		s.BroadcastGameState(room)
	}()

	return nil
}

// handleDynamicMinePlacement обрабатывает динамическое размещение мин
func (s *Service) handleDynamicMinePlacement(room *Room, clickRow, clickCol int) map[[2]int]bool {
	log.Printf("Режим %s, начинаем динамическое размещение мин", room.GameMode)
	room.GameState.Mu.Unlock()
	mineGrid := s.DetermineMinePlacement(room, clickRow, clickCol)
	room.GameState.Mu.Lock()

	changedCells := make(map[[2]int]bool)
	for i := 0; i < room.GameState.Rows; i++ {
		for j := 0; j < room.GameState.Cols; j++ {
			if !room.GameState.Board[i][j].IsRevealed {
				oldMine := room.GameState.Board[i][j].IsMine
				room.GameState.Board[i][j].IsMine = mineGrid[i][j]
				if oldMine != mineGrid[i][j] {
					changedCells[[2]int{i, j}] = true
					for di := -1; di <= 1; di++ {
						for dj := -1; dj <= 1; dj++ {
							if di == 0 && dj == 0 {
								continue
							}
							ni, nj := i+di, j+dj
							if ni >= 0 && ni < room.GameState.Rows && nj >= 0 && nj < room.GameState.Cols {
								changedCells[[2]int{ni, nj}] = true
							}
						}
					}
				}
			}
		}
	}

	// Пересчитываем соседние мины
	for pos := range changedCells {
		i, j := pos[0], pos[1]
		if !room.GameState.Board[i][j].IsMine {
			count := 0
			for di := -1; di <= 1; di++ {
				for dj := -1; dj <= 1; dj++ {
					if di == 0 && dj == 0 {
						continue
					}
					ni, nj := i+di, j+dj
					if ni >= 0 && ni < room.GameState.Rows && nj >= 0 && nj < room.GameState.Cols {
						if room.GameState.Board[ni][nj].IsMine {
							count++
						}
					}
				}
			}
			room.GameState.Board[i][j].NeighborMines = count
		}
	}

	return changedCells
}

// handleMineExplosion обрабатывает взрыв мины
func (s *Service) handleMineExplosion(room *Room, playerID string, row, col int, nickname, playerColor string) error {
	room.GameState.GameOver = true
	s.setLoserInfo(room, playerID)

	if s.wsManager != nil {
		wsPlayer := s.wsManager.GetWSPlayer(playerID)
		var userID int
		if wsPlayer != nil {
			userID = wsPlayer.GetUserID()
		} else {
			roomPlayer := room.GetPlayer(playerID)
			if roomPlayer != nil {
				userID = roomPlayer.UserID
			}
		}

		if userID > 0 {
			s.recordGameResult(room, playerID, false)
		}
	}

	log.Printf("Игра окончена - подорвалась мина! Игрок: %s", nickname)

	// В режиме fair вычисляем подсказки при проигрыше
	gameMode := room.GameMode
	if gameMode == "fair" {
		room.GameState.Mu.Unlock()
		s.CalculateCellHints(room)
		room.GameState.Mu.Lock()
	}

	// Отправляем сервисное сообщение о взрыве
	if nickname != "" {
		chatMsg := Message{
			Type:     "chat",
			PlayerID: playerID,
			Nickname: nickname,
			Color:    playerColor,
			Chat: &ChatMessage{
				Text:     fmt.Sprintf("%s подорвался на мине на (%d, %d) 💣", nickname, row+1, col+1),
				IsSystem: true,
				Action:   "explode",
				Row:      row,
				Col:      col,
			},
		}
		s.BroadcastToAll(room, chatMsg)
	}

	// Отправляем полное состояние игры после взрыва, чтобы показать все мины
	s.BroadcastGameState(room)

	return nil
}

// handleGameWin обрабатывает победу
func (s *Service) handleGameWin(room *Room, playerID string) {
	var gameTime float64
	room.Mu.RLock()
	if room.StartTime != nil {
		gameTime = time.Since(*room.StartTime).Seconds()
	}
	loserID := room.GameState.LoserPlayerID
	room.Mu.RUnlock()

	go func() {
		room.Mu.RLock()
		participants := make([]GameParticipant, 0)
		for _, p := range room.Players {
			if p.UserID > 0 {
				participants = append(participants, GameParticipant{
					UserID:   p.UserID,
					Nickname: p.Nickname,
					Color:    p.Color,
				})
			}
		}
		chording := room.Chording
		quickStart := room.QuickStart
		roomID := room.ID
		creatorID := room.CreatorID
		hasCustomSeed := room.HasCustomSeed
		seed := ""
		if room.GameState != nil {
			seed = room.GameState.Seed
		}
		room.Mu.RUnlock()

		for _, p := range room.Players {
			if p.ID != loserID && p.UserID > 0 && s.profileHandler != nil {
				if err := s.profileHandler.RecordGameResult(p.UserID, room.Cols, room.Rows, room.Mines, gameTime, true, chording, quickStart, roomID, seed, hasCustomSeed, creatorID, participants); err != nil {
					log.Printf("Ошибка записи результата игры: %v", err)
				}
			}
		}

		if err := s.roomManager.SaveRoom(room); err != nil {
			log.Printf("Предупреждение: не удалось сохранить комнату %s после победы: %v", room.ID, err)
		}
	}()
}

// setLoserInfo устанавливает информацию о проигравшем
func (s *Service) setLoserInfo(room *Room, playerID string) {
	var nickname string
	if s.wsManager != nil {
		wsPlayer := s.wsManager.GetWSPlayer(playerID)
		if wsPlayer != nil {
			nickname = wsPlayer.GetNickname()
		} else {
			roomPlayer := room.GetPlayer(playerID)
			if roomPlayer != nil {
				nickname = roomPlayer.Nickname
			}
		}
	} else {
		roomPlayer := room.GetPlayer(playerID)
		if roomPlayer != nil {
			nickname = roomPlayer.Nickname
		}
	}

	if nickname != "" {
		room.GameState.LoserPlayerID = playerID
		room.GameState.LoserNickname = nickname
	}
}

// recordGameResult записывает результат игры
func (s *Service) recordGameResult(room *Room, playerID string, won bool) {
	var userID int
	if s.wsManager != nil {
		wsPlayer := s.wsManager.GetWSPlayer(playerID)
		if wsPlayer != nil {
			userID = wsPlayer.GetUserID()
		} else {
			roomPlayer := room.GetPlayer(playerID)
			if roomPlayer != nil {
				userID = roomPlayer.UserID
			}
		}
	} else {
		roomPlayer := room.GetPlayer(playerID)
		if roomPlayer != nil {
			userID = roomPlayer.UserID
		}
	}

	if userID == 0 || s.profileHandler == nil {
		return
	}

	var gameTime float64
	room.Mu.RLock()
	if room.StartTime != nil {
		gameTime = time.Since(*room.StartTime).Seconds()
	}
	participants := make([]GameParticipant, 0)
	for _, p := range room.Players {
		if p.UserID > 0 {
			participants = append(participants, GameParticipant{
				UserID:   p.UserID,
				Nickname: p.Nickname,
				Color:    p.Color,
			})
		}
	}
	chording := room.Chording
	quickStart := room.QuickStart
	roomID := room.ID
	creatorID := room.CreatorID
	hasCustomSeed := room.HasCustomSeed
	seed := ""
	if room.GameState != nil {
		seed = room.GameState.Seed
	}
	room.Mu.RUnlock()

	go func() {
		if err := s.profileHandler.RecordGameResult(userID, room.Cols, room.Rows, room.Mines, gameTime, won, chording, quickStart, roomID, seed, hasCustomSeed, creatorID, participants); err != nil {
			log.Printf("Ошибка записи результата игры: %v", err)
		}
		if err := s.roomManager.SaveRoom(room); err != nil {
			log.Printf("Предупреждение: не удалось сохранить комнату %s: %v", roomID, err)
		}
	}()
}

// revealNeighbors открывает соседние пустые ячейки
func (s *Service) revealNeighbors(room *Room, row, col int, changedCells map[[2]int]bool) {
	for di := -1; di <= 1; di++ {
		for dj := -1; dj <= 1; dj++ {
			if di == 0 && dj == 0 {
				continue
			}
			ni, nj := row+di, col+dj
			if ni >= 0 && ni < room.GameState.Rows && nj >= 0 && nj < room.GameState.Cols {
				cell := &room.GameState.Board[ni][nj]
				if !cell.IsRevealed && !cell.IsFlagged && !cell.IsMine {
					cell.IsRevealed = true
					room.GameState.Revealed++
					changedCells[[2]int{ni, nj}] = true
					if cell.NeighborMines == 0 {
						s.revealNeighbors(room, ni, nj, changedCells)
					}
				}
			}
		}
	}
}
