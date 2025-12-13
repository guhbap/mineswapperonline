package main

import (
	"bytes"
	"encoding/binary"
)

// CellUpdate представляет обновление одной клетки
type CellUpdate struct {
	Row  int
	Col  int
	Type byte // Тип клетки: 0=закрыта, 0-8=количество мин, 9=мина, 10=зеленая, 11=желтая, 12=красная
}

// encodeCellUpdateBinary кодирует обновления клеток в бинарный формат
// Формат:
// - 1 байт: тип сообщения (6)
// - 1 байт: флаги (бит 0: GameOver, бит 1: GameWon, бит 2: есть Revealed, бит 3: есть HintsUsed)
// - Если GameOver: 1 байт длина LoserPlayerID, 5 байт LoserPlayerID, 1 байт длина LoserNickname, N байт LoserNickname
// - Если GameWon: (нет дополнительных данных)
// - Если есть Revealed: 2 байта Revealed (uint16)
// - Если есть HintsUsed: 1 байт HintsUsed
// - 2 байта: количество обновленных клеток (uint16)
// - Для каждой клетки: 2 байта Row (uint16), 2 байта Col (uint16), 1 байт Type
func encodeCellUpdateBinary(updates []CellUpdate, gameOver bool, gameWon bool, revealed int, hintsUsed int, loserPlayerID string, loserNickname string) ([]byte, error) {
	buf := new(bytes.Buffer)
	buf.WriteByte(MessageTypeCellUpdate)

	// Флаги
	flags := byte(0)
	if gameOver {
		flags |= 1 << 0
	}
	if gameWon {
		flags |= 1 << 1
	}
	if revealed >= 0 {
		flags |= 1 << 2
	}
	if hintsUsed >= 0 {
		flags |= 1 << 3
	}
	buf.WriteByte(flags)

	// GameOver данные
	if gameOver {
		loserPID := truncatePlayerID(loserPlayerID)
		loserPIDLen := byte(len(loserPID))
		if len(loserPID) > 5 {
			loserPIDLen = 5
		}
		buf.WriteByte(loserPIDLen)
		pidBytes := make([]byte, 5)
		if loserPIDLen > 0 {
			copy(pidBytes, []byte(loserPID))
		}
		buf.Write(pidBytes)

		loserNicknameBytes := []byte(loserNickname)
		nicknameLen := byte(len(loserNicknameBytes))
		if len(loserNicknameBytes) > 255 {
			nicknameLen = 255
			loserNicknameBytes = loserNicknameBytes[:255]
		}
		buf.WriteByte(nicknameLen)
		if nicknameLen > 0 {
			buf.Write(loserNicknameBytes)
		}
	}

	// Revealed
	if revealed >= 0 {
		binary.Write(buf, binary.LittleEndian, uint16(revealed))
	}

	// HintsUsed
	if hintsUsed >= 0 {
		buf.WriteByte(byte(hintsUsed))
	}

	// Количество обновленных клеток
	updateCount := uint16(len(updates))
	if len(updates) > 65535 {
		updateCount = 65535
	}
	binary.Write(buf, binary.LittleEndian, updateCount)

	// Обновленные клетки
	for i := 0; i < int(updateCount); i++ {
		binary.Write(buf, binary.LittleEndian, uint16(updates[i].Row))
		binary.Write(buf, binary.LittleEndian, uint16(updates[i].Col))
		buf.WriteByte(updates[i].Type)
	}

	return buf.Bytes(), nil
}

// getCellType возвращает тип клетки для бинарного формата
func getCellType(cell *Cell, row, col int, gameMode string, cellHints []CellHint) byte {
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
	return CellTypeClosed // Флаг не меняет тип, клетка остается закрытой
}

// collectCellUpdates собирает измененные клетки из gameState
func collectCellUpdates(room *Room, changedCells map[[2]int]bool) []CellUpdate {
	updates := make([]CellUpdate, 0)
	
	room.GameState.mu.RLock()
	gameMode := room.GameMode
	cellHints := room.GameState.CellHints
	board := room.GameState.Board
	rows := room.GameState.Rows
	cols := room.GameState.Cols
	room.GameState.mu.RUnlock()

	room.mu.RLock()
	gameMode = room.GameMode
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

