package game

import (
	"log"
	mathrand "math/rand"
	"time"
)

// NewGameState создает новое состояние игры
// seed: если 0, генерируется новый seed; иначе используется переданный
func NewGameState(rows, cols, mines int, gameMode string, seed int64) *GameState {
	log.Printf("NewGameState: начало создания, rows=%d, cols=%d, mines=%d, gameMode=%s, seed=%d", rows, cols, mines, gameMode, seed)
	// По умолчанию classic
	if gameMode == "" {
		gameMode = "classic"
	}
	
	// Генерируем seed для воспроизводимости, если не передан
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	gs := &GameState{
		Rows:          rows,
		Cols:          cols,
		Mines:         mines,
		Seed:          seed,
		GameOver:      false,
		GameWon:       false,
		Revealed:      0,
		HintsUsed:     0,
		LoserPlayerID: "",
		LoserNickname: "",
		Board:         make([][]Cell, rows),
		FlagSetInfo:   make(map[int]FlagInfo),
	}
	log.Printf("NewGameState: структура создана, seed=%d, инициализируем поле", seed)

	// Инициализация поля
	for i := range gs.Board {
		gs.Board[i] = make([]Cell, cols)
	}
	log.Printf("NewGameState: поле инициализировано")

	// В режимах training и fair мины НЕ размещаются заранее - они определяются динамически при клике
	// В классическом режиме размещаем мины случайно
	if gameMode == "classic" {
		log.Printf("NewGameState: размещаем мины в классическом режиме с seed=%d", seed)
		// Используем seed для генерации
		rng := mathrand.New(mathrand.NewSource(seed))
		minesPlaced := 0
		for minesPlaced < mines {
			row := rng.Intn(rows)
			col := rng.Intn(cols)
			if !gs.Board[row][col].IsMine {
				gs.Board[row][col].IsMine = true
				minesPlaced++
			}
		}
		log.Printf("NewGameState: мины размещены, подсчитываем соседние мины")

		// Подсчет соседних мин для обычного режима
		for i := 0; i < rows; i++ {
			for j := 0; j < cols; j++ {
				if !gs.Board[i][j].IsMine {
					count := 0
					for di := -1; di <= 1; di++ {
						for dj := -1; dj <= 1; dj++ {
							ni, nj := i+di, j+dj
							if ni >= 0 && ni < rows && nj >= 0 && nj < cols {
								if gs.Board[ni][nj].IsMine {
									count++
								}
							}
						}
					}
					gs.Board[i][j].NeighborMines = count
				}
			}
		}
		log.Printf("NewGameState: подсчет соседних мин завершен")
	}
	// В режимах training и fair подсчет соседних мин будет происходить динамически при размещении мин

	log.Printf("NewGameState: завершено создание GameState")
	return gs
}

// Copy создает копию состояния игры
func (gs *GameState) Copy() *GameState {
	gs.Mu.RLock()
	defer gs.Mu.RUnlock()

	gsCopy := &GameState{
		Rows:          gs.Rows,
		Cols:          gs.Cols,
		Mines:         gs.Mines,
		Seed:          gs.Seed,
		GameOver:      gs.GameOver,
		GameWon:       gs.GameWon,
		Revealed:      gs.Revealed,
		HintsUsed:     gs.HintsUsed,
		SafeCells:     make([]SafeCell, len(gs.SafeCells)),
		CellHints:     make([]CellHint, len(gs.CellHints)),
		LoserPlayerID: gs.LoserPlayerID,
		LoserNickname: gs.LoserNickname,
		Board:         make([][]Cell, len(gs.Board)),
		FlagSetInfo:   make(map[int]FlagInfo),
	}

	copy(gsCopy.SafeCells, gs.SafeCells)
	copy(gsCopy.CellHints, gs.CellHints)
	for k, v := range gs.FlagSetInfo {
		gsCopy.FlagSetInfo[k] = v
	}

	for i := range gs.Board {
		gsCopy.Board[i] = make([]Cell, len(gs.Board[i]))
		copy(gsCopy.Board[i], gs.Board[i])
	}

	return gsCopy
}

// calculateNeighborMines подсчитывает соседние мины для всех ячеек
func (gs *GameState) calculateNeighborMines() {
	for i := 0; i < gs.Rows; i++ {
		for j := 0; j < gs.Cols; j++ {
			if !gs.Board[i][j].IsMine {
				gs.Board[i][j].NeighborMines = gs.countNeighborMines(i, j)
			}
		}
	}
}

// countNeighborMines подсчитывает мины вокруг ячейки
func (gs *GameState) countNeighborMines(row, col int) int {
	count := 0
	for di := -1; di <= 1; di++ {
		for dj := -1; dj <= 1; dj++ {
			if di == 0 && dj == 0 {
				continue
			}
			ni, nj := row+di, col+dj
			if gs.isValidCell(ni, nj) && gs.Board[ni][nj].IsMine {
				count++
			}
		}
	}
	return count
}

// isValidCell проверяет валидность координат
func (gs *GameState) isValidCell(row, col int) bool {
	return row >= 0 && row < gs.Rows && col >= 0 && col < gs.Cols
}

// isInRadius проверяет, находится ли ячейка в радиусе от заданной точки
func (gs *GameState) isInRadius(row, col, centerRow, centerCol, radius int) bool {
	for di := -radius; di <= radius; di++ {
		for dj := -radius; dj <= radius; dj++ {
			if row == centerRow+di && col == centerCol+dj {
				return true
			}
		}
	}
	return false
}

// EnsureFirstClickSafe перемещает мины из радиуса первой ячейки
func (gs *GameState) EnsureFirstClickSafe(firstRow, firstCol int) {
	gs.Mu.Lock()
	defer gs.Mu.Unlock()

	// Собираем мины в радиусе 1
	minesToMove := []struct{ row, col int }{}
	for di := -1; di <= 1; di++ {
		for dj := -1; dj <= 1; dj++ {
			ni, nj := firstRow+di, firstCol+dj
			if gs.isValidCell(ni, nj) && gs.Board[ni][nj].IsMine {
				minesToMove = append(minesToMove, struct{ row, col int }{ni, nj})
				gs.Board[ni][nj].IsMine = false
			}
		}
	}

	// Перемещаем мины в случайные свободные места
	rng := mathrand.New(mathrand.NewSource(gs.Seed))
	for range minesToMove {
		attempts := 0
		for attempts < 100 {
			newRow := rng.Intn(gs.Rows)
			newCol := rng.Intn(gs.Cols)

			if !gs.isInRadius(newRow, newCol, firstRow, firstCol, 1) &&
				!gs.Board[newRow][newCol].IsMine {
				gs.Board[newRow][newCol].IsMine = true
				break
			}
			attempts++
		}
	}

	// Пересчитываем соседние мины
	gs.calculateNeighborMines()
}

// RevealNeighbors открывает соседние пустые ячейки и возвращает измененные ячейки
func (gs *GameState) RevealNeighbors(row, col int, changedCells map[[2]int]bool) {
	for di := -1; di <= 1; di++ {
		for dj := -1; dj <= 1; dj++ {
			if di == 0 && dj == 0 {
				continue
			}
			ni, nj := row+di, col+dj
			if gs.isValidCell(ni, nj) {
				cell := &gs.Board[ni][nj]
				if !cell.IsRevealed && !cell.IsFlagged && !cell.IsMine {
					cell.IsRevealed = true
					gs.Revealed++
					changedCells[[2]int{ni, nj}] = true
					if cell.NeighborMines == 0 {
						gs.RevealNeighbors(ni, nj, changedCells)
					}
				}
			}
		}
	}
}

// CheckWin проверяет условие победы
func (gs *GameState) CheckWin() bool {
	gs.Mu.RLock()
	defer gs.Mu.RUnlock()
	totalCells := gs.Rows * gs.Cols
	return gs.Revealed == totalCells-gs.Mines
}
