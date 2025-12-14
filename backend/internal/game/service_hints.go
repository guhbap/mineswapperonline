package game

import (
	"fmt"
	"log"
	"math/rand"
)

// HandleHint обрабатывает подсказку
func (s *Service) HandleHint(room *Room, playerID string, hint *Hint) error {
	room.GameState.Mu.Lock()

	if room.GameState.GameOver || room.GameState.GameWon {
		log.Printf("Игра уже окончена, подсказка игнорируется")
		room.GameState.Mu.Unlock()
		return nil
	}

	if room.GameState.HintsUsed >= 3 {
		log.Printf("Лимит подсказок исчерпан (использовано: %d)", room.GameState.HintsUsed)
		room.GameState.Mu.Unlock()
		return nil
	}

	row, col := hint.Row, hint.Col
	if row < 0 || row >= room.GameState.Rows || col < 0 || col >= room.GameState.Cols {
		log.Printf("Некорректные координаты подсказки: row=%d, col=%d", row, col)
		room.GameState.Mu.Unlock()
		return fmt.Errorf("invalid coordinates")
	}

	cell := &room.GameState.Board[row][col]

	if cell.IsRevealed || cell.IsFlagged {
		log.Printf("Ячейка уже открыта или помечена флагом: row=%d, col=%d", row, col)
		room.GameState.Mu.Unlock()
		return nil
	}

	room.Mu.RLock()
	player := room.Players[playerID]
	var nickname string
	var playerColor string
	if player != nil {
		nickname = player.Nickname
		playerColor = player.Color
	}
	room.Mu.RUnlock()

	if cell.IsMine {
		cell.IsFlagged = true
		cell.FlagColor = playerColor
		room.GameState.HintsUsed++
		changedCells := make(map[[2]int]bool)
		changedCells[[2]int{row, col}] = true
		room.GameState.Mu.Unlock()

		s.BroadcastCellUpdates(room, changedCells, room.GameState.GameOver, room.GameState.GameWon, room.GameState.Revealed, room.GameState.HintsUsed, room.GameState.LoserPlayerID, room.GameState.LoserNickname)

		gameMode := room.GameMode
		if gameMode == "training" {
			go func() {
				s.CalculateCellHints(room)
				s.BroadcastGameState(room)
			}()
		} else {
			go func() {
				s.BroadcastGameState(room)
			}()
		}

		if nickname != "" {
			chatMsg := Message{
				Type:     "chat",
				PlayerID: playerID,
				Nickname: nickname,
				Color:    playerColor,
				Chat: &ChatMessage{
					Text:     fmt.Sprintf("%s использовал подсказку и поставил флаг на (%d, %d) 💡", nickname, row+1, col+1),
					IsSystem: true,
					Action:   "hint",
					Row:      row,
					Col:      col,
				},
			}
			s.BroadcastToAll(room, chatMsg)
		}
		return nil
	}

	// Открываем ячейку
	changedCells := make(map[[2]int]bool)
	changedCells[[2]int{row, col}] = true
	cell.IsRevealed = true
	room.GameState.Revealed++
	room.GameState.HintsUsed++

	if cell.NeighborMines == 0 {
		s.revealNeighbors(room, row, col, changedCells)
	}

	// Проверка победы
	totalCells := room.GameState.Rows * room.GameState.Cols
	if room.GameState.Revealed == totalCells-room.GameState.Mines {
		room.GameState.GameWon = true
		s.handleGameWin(room, playerID)
	}

	room.GameState.Mu.Unlock()
	s.BroadcastCellUpdates(room, changedCells, room.GameState.GameOver, room.GameState.GameWon, room.GameState.Revealed, room.GameState.HintsUsed, room.GameState.LoserPlayerID, room.GameState.LoserNickname)

	gameMode := room.GameMode
	if gameMode == "training" {
		go func() {
			s.CalculateCellHints(room)
			s.BroadcastGameState(room)
		}()
	} else {
		go func() {
			s.BroadcastGameState(room)
		}()
	}

	if nickname != "" {
		chatMsg := Message{
			Type:     "chat",
			PlayerID: playerID,
			Nickname: nickname,
			Color:    playerColor,
			Chat: &ChatMessage{
				Text:     fmt.Sprintf("%s использовал подсказку и открыл поле на (%d, %d) 💡", nickname, row+1, col+1),
				IsSystem: true,
				Action:   "hint",
				Row:      row,
				Col:      col,
			},
		}
		s.BroadcastToAll(room, chatMsg)
	}

	return nil
}

// CalculateCellHints вычисляет подсказки только для ячеек на границе
func (s *Service) CalculateCellHints(room *Room) {
	room.GameState.Mu.Lock()
	defer room.GameState.Mu.Unlock()

	lm := NewLabelMap(room.GameState.Cols, room.GameState.Rows)

	for i := 0; i < room.GameState.Rows; i++ {
		for j := 0; j < room.GameState.Cols; j++ {
			if room.GameState.Board[i][j].IsRevealed {
				lm.SetLabel(i, j, room.GameState.Board[i][j].NeighborMines)
			}
		}
	}

	solver := MakeSolver(lm, room.GameState.Mines)
	hints := make([]CellHint, 0)
	boundary := lm.GetBoundary()

	for i, pos := range boundary {
		canBeDangerous := solver.CanBeDangerous(i)
		canBeSafe := solver.CanBeSafe(i)

		var hintType string
		if canBeDangerous && canBeSafe {
			hintType = "UNKNOWN"
		} else if canBeDangerous && !canBeSafe {
			hintType = "MINE"
		} else if !canBeDangerous && canBeSafe {
			hintType = "SAFE"
		} else {
			continue
		}

		hints = append(hints, CellHint{
			Row:  pos.Row,
			Col:  pos.Col,
			Type: hintType,
		})
	}

	room.GameState.CellHints = hints
	log.Printf("Вычислены подсказки для %d ячеек на границе", len(hints))
}

// DetermineMinePlacement определяет размещение мин при клике в режимах training и fair
func (s *Service) DetermineMinePlacement(room *Room, clickRow, clickCol int) [][]bool {
	log.Printf("DetermineMinePlacement: начало, clickRow=%d, clickCol=%d", clickRow, clickCol)

	isFirstClick := room.GameState.Revealed == 0
	if isFirstClick && room.QuickStart {
		log.Printf("DetermineMinePlacement: QuickStart включен, делаем первую клетку нулевой")
		mineGrid := make([][]bool, room.GameState.Rows)
		for i := 0; i < room.GameState.Rows; i++ {
			mineGrid[i] = make([]bool, room.GameState.Cols)
		}

		placed := 0
		attempts := 0
		maxAttempts := room.GameState.Rows * room.GameState.Cols * 2
		for placed < room.GameState.Mines && attempts < maxAttempts {
			row := rand.Intn(room.GameState.Rows)
			col := rand.Intn(room.GameState.Cols)
			attempts++

			isNearClick := false
			for di := -1; di <= 1; di++ {
				for dj := -1; dj <= 1; dj++ {
					if row == clickRow+di && col == clickCol+dj {
						isNearClick = true
						break
					}
				}
				if isNearClick {
					break
				}
			}

			if isNearClick || mineGrid[row][col] {
				continue
			}

			mineGrid[row][col] = true
			placed++
		}

		log.Printf("DetermineMinePlacement: QuickStart - размещено %d мин", placed)
		return mineGrid
	}

	lm := NewLabelMap(room.GameState.Cols, room.GameState.Rows)

	for i := 0; i < room.GameState.Rows; i++ {
		for j := 0; j < room.GameState.Cols; j++ {
			if room.GameState.Board[i][j].IsRevealed {
				lm.SetLabel(i, j, room.GameState.Board[i][j].NeighborMines)
			}
		}
	}

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

	boundaryIdx := -1
	if clickRow >= 0 && clickRow < room.GameState.Rows && clickCol >= 0 && clickCol < room.GameState.Cols {
		boundaryIdx = lm.GetBoundaryIndex(clickRow, clickCol)
	}

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

	// Fallback
	minesToPlace := remainingMines
	if minesToPlace == 0 && room.GameState.Mines > 0 {
		minesToPlace = room.GameState.Mines
	}

	mineGrid := make([][]bool, room.GameState.Rows)
	for i := 0; i < room.GameState.Rows; i++ {
		mineGrid[i] = make([]bool, room.GameState.Cols)
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

	log.Printf("DetermineMinePlacement: fallback mineGrid создан с %d минами", placed)
	return mineGrid
}

