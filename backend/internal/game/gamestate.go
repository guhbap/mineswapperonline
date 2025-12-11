package game

import (
	"math/rand"
	"sync"
	"time"
)

type Cell struct {
	IsMine        bool `json:"isMine"`
	IsRevealed    bool `json:"isRevealed"`
	IsFlagged     bool `json:"isFlagged"`
	NeighborMines int  `json:"neighborMines"`
}

type GameState struct {
	Board         [][]Cell `json:"board"`
	Rows          int      `json:"rows"`
	Cols          int      `json:"cols"`
	Mines         int      `json:"mines"`
	GameOver      bool     `json:"gameOver"`
	GameWon       bool     `json:"gameWon"`
	Revealed      int      `json:"revealed"`
	LoserPlayerID string   `json:"loserPlayerId,omitempty"`
	LoserNickname string   `json:"loserNickname,omitempty"`
	mu            sync.RWMutex
}

// NewGameState создает новое состояние игры
func NewGameState(rows, cols, mines int) *GameState {
	gs := &GameState{
		Rows:          rows,
		Cols:          cols,
		Mines:         mines,
		GameOver:      false,
		GameWon:       false,
		Revealed:      0,
		LoserPlayerID: "",
		LoserNickname: "",
		Board:         make([][]Cell, rows),
	}

	// Инициализация поля
	for i := range gs.Board {
		gs.Board[i] = make([]Cell, cols)
	}

	// Размещение мин
	rand.Seed(time.Now().UnixNano())
	minesPlaced := 0
	for minesPlaced < mines {
		row := rand.Intn(rows)
		col := rand.Intn(cols)
		if !gs.Board[row][col].IsMine {
			gs.Board[row][col].IsMine = true
			minesPlaced++
		}
	}

	// Подсчет соседних мин
	gs.calculateNeighborMines()

	return gs
}

// Copy создает копию состояния игры
func (gs *GameState) Copy() *GameState {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	copy := &GameState{
		Rows:          gs.Rows,
		Cols:          gs.Cols,
		Mines:         gs.Mines,
		GameOver:      gs.GameOver,
		GameWon:       gs.GameWon,
		Revealed:      gs.Revealed,
		LoserPlayerID: gs.LoserPlayerID,
		LoserNickname: gs.LoserNickname,
		Board:         make([][]Cell, len(gs.Board)),
	}

	for i := range gs.Board {
		copy.Board[i] = make([]Cell, len(gs.Board[i]))
		for j := range gs.Board[i] {
			copy.Board[i][j] = gs.Board[i][j]
		}
	}

	return copy
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
	gs.mu.Lock()
	defer gs.mu.Unlock()

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
	rand.Seed(time.Now().UnixNano())
	for range minesToMove {
		attempts := 0
		for attempts < 100 {
			newRow := rand.Intn(gs.Rows)
			newCol := rand.Intn(gs.Cols)

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

// RevealNeighbors открывает соседние пустые ячейки
func (gs *GameState) RevealNeighbors(row, col int) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

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
					if cell.NeighborMines == 0 {
						gs.mu.Unlock()
						gs.RevealNeighbors(ni, nj)
						gs.mu.Lock()
					}
				}
			}
		}
	}
}

// CheckWin проверяет условие победы
func (gs *GameState) CheckWin() bool {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	totalCells := gs.Rows * gs.Cols
	return gs.Revealed == totalCells-gs.Mines
}

