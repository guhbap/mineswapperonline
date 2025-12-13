package game

import (
	"log"
	"math/rand"
	"time"
)

// GameService предоставляет методы для работы с игровой логикой
type GameService struct {
	resultRecorder GameResultRecorder
}

// NewGameService создает новый экземпляр GameService
func NewGameService(resultRecorder GameResultRecorder) *GameService {
	return &GameService{
		resultRecorder: resultRecorder,
	}
}

// HandleCellClick обрабатывает клик по ячейке
func (s *GameService) HandleCellClick(room *Room, playerID string, row, col int, isFlag bool) {
	log.Printf("handleCellClick: начало, row=%d, col=%d, flag=%v", row, col, isFlag)
	
	room.GameState.mu.Lock()
	defer room.GameState.mu.Unlock()

	if room.GameState.GameOver || room.GameState.GameWon {
		log.Printf("Игра уже окончена, клик игнорируется")
		return
	}

	if row < 0 || row >= room.GameState.Rows || col < 0 || col >= room.GameState.Cols {
		log.Printf("Некорректные координаты: row=%d, col=%d", row, col)
		return
	}

	cell := &room.GameState.Board[row][col]

	// Получаем информацию об игроке
	// Примечание: в реальной реализации Player будет из websocket пакета
	// Здесь используем упрощенный подход - данные передаются отдельно
	room.mu.RLock()
	player := room.Players[playerID]
	var nickname string
	var playerColor string
	var userID int
	if player != nil {
		nickname = player.Nickname
		playerColor = player.Color
		userID = player.UserID
	}
	room.mu.RUnlock()

	if isFlag {
		s.handleFlagToggle(room, playerID, row, col, cell, playerColor, nickname)
		return
	}

	// Обработка открытия ячейки
	s.handleCellReveal(room, playerID, row, col, cell, userID, nickname, playerColor)
}

// handleFlagToggle обрабатывает переключение флага
func (s *GameService) handleFlagToggle(room *Room, playerID string, row, col int, cell *Cell, playerColor, nickname string) {
	if cell.IsRevealed {
		log.Printf("Нельзя поставить флаг на открытую ячейку: row=%d, col=%d", row, col)
		return
	}

	wasFlagged := cell.IsFlagged
	cellKey := row*room.GameState.Cols + col
	now := time.Now()

	if wasFlagged {
		if flagInfo, exists := room.GameState.flagSetInfo[cellKey]; exists {
			if flagInfo.PlayerID != playerID {
				timeSinceFlagSet := now.Sub(flagInfo.SetTime)
				if timeSinceFlagSet < 1*time.Second {
					log.Printf("Нельзя снять флаг сразу после установки другим игроком: row=%d, col=%d", row, col)
					return
				}
			}
		}
		delete(room.GameState.flagSetInfo, cellKey)
		cell.FlagColor = ""
	} else {
		room.GameState.flagSetInfo[cellKey] = FlagInfo{
			SetTime:  now,
			PlayerID: playerID,
		}
		cell.FlagColor = playerColor
	}

	cell.IsFlagged = !cell.IsFlagged
	log.Printf("Флаг переключен: row=%d, col=%d, flagged=%v", row, col, cell.IsFlagged)
}

// handleCellReveal обрабатывает открытие ячейки
func (s *GameService) handleCellReveal(room *Room, playerID string, row, col int, cell *Cell, userID int, nickname, playerColor string) {
	// Устанавливаем время начала игры при первом клике
	isFirstClick := room.GameState.Revealed == 0
	if isFirstClick && room.StartTime == nil {
		room.mu.Lock()
		now := time.Now()
		room.StartTime = &now
		room.mu.Unlock()
		log.Printf("StartTime установлен при первом клике: %v", now)
	}

	// В режимах training и fair мины размещаются динамически
	room.mu.RLock()
	gameMode := room.GameMode
	room.mu.RUnlock()

	if gameMode == "training" || gameMode == "fair" {
		room.GameState.mu.Unlock()
		mineGrid := s.determineMinePlacement(room, row, col)
		room.GameState.mu.Lock()

		// Применяем размещение мин
		changedCells := make(map[[2]int]bool)
		for i := 0; i < room.GameState.Rows; i++ {
			for j := 0; j < room.GameState.Cols; j++ {
				if !room.GameState.Board[i][j].IsRevealed {
					oldMine := room.GameState.Board[i][j].IsMine
					room.GameState.Board[i][j].IsMine = mineGrid[i][j]
					if oldMine != mineGrid[i][j] {
						changedCells[[2]int{i, j}] = true
						// Помечаем соседей для пересчета
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

		cell = &room.GameState.Board[row][col]
	}

	// Открываем ячейку
	cell.IsRevealed = true
	room.GameState.Revealed++
	changedCells := make(map[[2]int]bool)
	changedCells[[2]int{row, col}] = true

	if cell.IsMine {
		s.handleMineExplosion(room, playerID, userID, nickname, playerColor, row, col)
		return
	}

	// Автоматическое открытие соседних пустых ячеек
	if cell.NeighborMines == 0 {
		room.GameState.RevealNeighbors(row, col, changedCells)
	}

	// Проверка победы
	totalCells := room.GameState.Rows * room.GameState.Cols
	if room.GameState.Revealed == totalCells-room.GameState.Mines {
		s.handleGameWin(room, playerID, userID)
	}
}

// handleMineExplosion обрабатывает взрыв мины
func (s *GameService) handleMineExplosion(room *Room, playerID string, userID int, nickname, playerColor string, row, col int) {
	room.GameState.GameOver = true
	room.GameState.LoserPlayerID = playerID
	room.GameState.LoserNickname = nickname

	// Вычисляем время игры
	room.mu.RLock()
	var gameTime float64
	if room.StartTime != nil {
		gameTime = time.Since(*room.StartTime).Seconds()
	} else {
		gameTime = 0.0
	}
	room.mu.RUnlock()

	// Записываем поражение в БД
	if userID > 0 && s.resultRecorder != nil {
		participants := s.collectParticipants(room)
		if err := s.resultRecorder.RecordGameResult(userID, room.Cols, room.Rows, room.Mines, gameTime, false, participants); err != nil {
			log.Printf("Ошибка записи результата игры: %v", err)
		}
	}

	log.Printf("Игра окончена - подорвалась мина! Игрок: %s (%s)", nickname, playerID)
}

// handleGameWin обрабатывает победу
func (s *GameService) handleGameWin(room *Room, playerID string, userID int) {
	room.GameState.GameWon = true
	log.Printf("Победа! Все ячейки открыты!")

	// Вычисляем время игры
	room.mu.RLock()
	var gameTime float64
	if room.StartTime != nil {
		gameTime = time.Since(*room.StartTime).Seconds()
	} else {
		gameTime = 0.0
	}
	loserID := room.GameState.LoserPlayerID
	participants := s.collectParticipants(room)
	room.mu.RUnlock()

	// Записываем победу для всех игроков, которые не проиграли
	room.mu.RLock()
	for _, p := range room.Players {
		if p.ID != loserID && p.UserID > 0 && s.resultRecorder != nil {
			if err := s.resultRecorder.RecordGameResult(p.UserID, room.Cols, room.Rows, room.Mines, gameTime, true, participants); err != nil {
				log.Printf("Ошибка записи результата игры: %v", err)
			}
		}
	}
	room.mu.RUnlock()
}

// collectParticipants собирает список участников игры
func (s *GameService) collectParticipants(room *Room) []GameParticipant {
	participants := make([]GameParticipant, 0)
	room.mu.RLock()
	for _, p := range room.Players {
		if p.UserID > 0 {
			participants = append(participants, GameParticipant{
				UserID:   p.UserID,
				Nickname: p.Nickname,
				Color:    p.Color,
			})
		}
	}
	room.mu.RUnlock()
	return participants
}

// determineMinePlacement определяет размещение мин при клике в режимах training и fair
func (s *GameService) determineMinePlacement(room *Room, clickRow, clickCol int) [][]bool {
	log.Printf("determineMinePlacement: начало, clickRow=%d, clickCol=%d", clickRow, clickCol)
	
	// Создаем LabelMap на основе открытых ячеек
	lm := NewLabelMap(room.GameState.Cols, room.GameState.Rows)
	
	revealedCount := 0
	for i := 0; i < room.GameState.Rows; i++ {
		for j := 0; j < room.GameState.Cols; j++ {
			if room.GameState.Board[i][j].IsRevealed {
				lm.SetLabel(i, j, room.GameState.Board[i][j].NeighborMines)
				revealedCount++
			}
		}
	}
	log.Printf("determineMinePlacement: установлено %d меток", revealedCount)

	// Подсчитываем уже размещенные мины
	placedMines := 0
	for i := 0; i < room.GameState.Rows; i++ {
		for j := 0; j < room.GameState.Cols; j++ {
			if !room.GameState.Board[i][j].IsRevealed && room.GameState.Board[i][j].IsMine {
				placedMines++
			}
		}
	}

	remainingMines := room.GameState.Mines - placedMines
	if remainingMines < 0 {
		remainingMines = 0
	}

	solver := MakeSolver(lm, remainingMines)
	boundaryIdx := lm.GetBoundaryIndex(clickRow, clickCol)
	hasSafeCells := solver.HasSafeCells()

	var shape *MineShape
	if boundaryIdx == -1 {
		outsideIsSafe := len(lm.GetBoundary()) == 0 || solver.OutsideIsSafe() || (!hasSafeCells && solver.OutsideCanBeSafe())
		if outsideIsSafe {
			shape = solver.AnyShapeWithOneEmpty()
			if shape != nil {
				return shape.MineGridWithEmpty(clickRow, clickCol)
			}
		} else {
			shape = solver.AnyShapeWithRemaining()
			if shape != nil {
				return shape.MineGridWithMine(clickRow, clickCol)
			}
		}
	} else {
		canBeSafe := solver.CanBeSafe(boundaryIdx)
		canBeDangerous := solver.CanBeDangerous(boundaryIdx)
		if canBeSafe && (!canBeDangerous || !hasSafeCells) {
			shape = solver.AnySafeShape(boundaryIdx)
		} else {
			shape = solver.AnyDangerousShape(boundaryIdx)
		}
	}

	if shape != nil {
		return shape.MineGrid()
	}

	// Fallback: случайное размещение
	mineGrid := make([][]bool, room.GameState.Rows)
	for i := 0; i < room.GameState.Rows; i++ {
		mineGrid[i] = make([]bool, room.GameState.Cols)
	}

	minesToPlace := remainingMines
	if minesToPlace == 0 && room.GameState.Mines > 0 {
		minesToPlace = room.GameState.Mines
	}

	placed := 0
	attempts := 0
	maxAttempts := room.GameState.Rows * room.GameState.Cols * 2
	for placed < minesToPlace && attempts < maxAttempts {
		row := rand.Intn(room.GameState.Rows)
		col := rand.Intn(room.GameState.Cols)
		attempts++

		if (row == clickRow && col == clickCol) || room.GameState.Board[row][col].IsRevealed {
			continue
		}

		if !mineGrid[row][col] {
			mineGrid[row][col] = true
			placed++
		}
	}

	return mineGrid
}

