package websocket

import "minesweeperonline/internal/game"

// CellUpdate представляет обновление одной клетки
type CellUpdate struct {
	Row  int
	Col  int
	Type byte // Тип клетки: 0=закрыта, 0-8=количество мин, 9=мина, 10=зеленая, 11=желтая, 12=красная
}

// Типы клеток для бинарного формата
const (
	CellTypeClosed  = byte(255) // Закрыта
	CellTypeMine    = byte(9)   // Мина
	CellTypeSafe    = byte(10)  // Зеленая (SAFE)
	CellTypeUnknown = byte(11)  // Желтая (UNKNOWN)
	CellTypeDanger  = byte(12)  // Красная (MINE)
)

// getCellType возвращает тип клетки для бинарного формата
func getCellType(cell *game.Cell, row, col int, gameMode string, cellHints []game.CellHint) byte {
	// Если клетка открыта
	if cell.IsRevealed {
		if cell.IsMine {
			return CellTypeMine
		}
		// Возвращаем количество соседних мин (0-8)
		if cell.NeighborMines > 8 {
			return 8
		}
		return byte(cell.NeighborMines)
	}

	// Если клетка закрыта и не помечена флагом
	if !cell.IsFlagged {
		// Проверяем подсказки для режима обучения
		if gameMode == "training" || gameMode == "fair" {
			for i := range cellHints {
				if cellHints[i].Row == row && cellHints[i].Col == col {
					switch cellHints[i].Type {
					case "SAFE":
						return CellTypeSafe
					case "UNKNOWN":
						return CellTypeUnknown
					case "MINE":
						return CellTypeDanger
					}
					break
				}
			}
		}
		return CellTypeClosed
	}

	// Если клетка помечена флагом
	return CellTypeClosed // Флаг не меняет тип, клетка остается закрытой (255)
}

// collectCellUpdates собирает измененные клетки из gameState
func collectCellUpdates(room *game.Room, changedCells map[[2]int]bool) []CellUpdate {
	updates := make([]CellUpdate, 0)

	room.GameState.mu.RLock()
	cellHints := room.GameState.CellHints
	board := room.GameState.Board
	rows := room.GameState.Rows
	cols := room.GameState.Cols
	room.GameState.mu.RUnlock()

	room.mu.RLock()
	gameMode := room.GameMode
	room.mu.RUnlock()

	for pos := range changedCells {
		row, col := pos[0], pos[1]
		if row < 0 || row >= rows || col < 0 || col >= cols {
			continue
		}

		cell := &board[row][col]
		cellType := getCellType(cell, row, col, gameMode, cellHints)

		updates = append(updates, CellUpdate{
			Row:  row,
			Col:  col,
			Type: cellType,
		})
	}

	return updates
}

