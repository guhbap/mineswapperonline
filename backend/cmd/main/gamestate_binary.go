package main

import (
	"bytes"
	"encoding/binary"
)

// decodeGameStateBinary декодирует бинарный формат в GameState
//
//lint:ignore U1000 Используется для отладки и тестирования
func decodeGameStateBinary(data []byte) (*GameState, error) {
	buf := bytes.NewReader(data)

	gs := &GameState{}

	// Читаем размеры
	var rows, cols, mines, revealed uint16
	binary.Read(buf, binary.LittleEndian, &rows)
	binary.Read(buf, binary.LittleEndian, &cols)
	binary.Read(buf, binary.LittleEndian, &mines)
	binary.Read(buf, binary.LittleEndian, &revealed)

	gs.Rows = int(rows)
	gs.Cols = int(cols)
	gs.Mines = int(mines)
	gs.Revealed = int(revealed)

	// Читаем HintsUsed
	hintsUsed, _ := buf.ReadByte()
	gs.HintsUsed = int(hintsUsed)

	// Читаем флаги
	flags, _ := buf.ReadByte()
	gs.GameOver = (flags & (1 << 0)) != 0
	gs.GameWon = (flags & (1 << 1)) != 0

	// Читаем LoserPlayerID
	loserPIDLen, _ := buf.ReadByte()
	pidBytes := make([]byte, 5)
	buf.Read(pidBytes)
	if loserPIDLen > 0 && loserPIDLen <= 5 {
		gs.LoserPlayerID = string(pidBytes[:loserPIDLen])
	}

	// Читаем LoserNickname
	nicknameLen, _ := buf.ReadByte()
	if nicknameLen > 0 {
		nicknameBytes := make([]byte, nicknameLen)
		buf.Read(nicknameBytes)
		gs.LoserNickname = string(nicknameBytes)
	}

	// Инициализируем Board
	gs.Board = make([][]Cell, gs.Rows)
	for i := range gs.Board {
		gs.Board[i] = make([]Cell, gs.Cols)
	}

	// Читаем Board
	for i := 0; i < gs.Rows; i++ {
		for j := 0; j < gs.Cols; j++ {
			cellByte, _ := buf.ReadByte()
			cell := &gs.Board[i][j]
			cell.IsMine = (cellByte & (1 << 0)) != 0
			cell.IsRevealed = (cellByte & (1 << 1)) != 0
			cell.IsFlagged = (cellByte & (1 << 2)) != 0
			cell.NeighborMines = int((cellByte >> 3) & 0x0F) // Бит 3-6
		}
	}

	// Читаем цвета флагов (если есть данные)
	if buf.Len() > 0 {
		flagCount, err := buf.ReadByte()
		if err == nil && flagCount > 0 {
			// Читаем цвета флагов
			for i := 0; i < int(flagCount); i++ {
				// Читаем cellKey (2 байта)
				var cellKey uint16
				if err := binary.Read(buf, binary.LittleEndian, &cellKey); err != nil {
					break
				}
				// Читаем длину цвета
				colorLen, err := buf.ReadByte()
				if err != nil {
					break
				}
				// Читаем цвет
				if colorLen > 0 && colorLen <= 7 {
					colorBytes := make([]byte, colorLen)
					if n, err := buf.Read(colorBytes); err == nil && n == int(colorLen) {
						color := string(colorBytes)
						// Применяем цвет к соответствующей ячейке
						row := int(cellKey) / gs.Cols
						col := int(cellKey) % gs.Cols
						if row >= 0 && row < gs.Rows && col >= 0 && col < gs.Cols {
							gs.Board[row][col].FlagColor = color
						}
					}
				}
			}
		}
	}

	// Читаем SafeCells (если есть данные)
	if buf.Len() > 0 {
		var safeCellsCount uint16
		if err := binary.Read(buf, binary.LittleEndian, &safeCellsCount); err == nil && safeCellsCount > 0 {
			gs.SafeCells = make([]SafeCell, safeCellsCount)
			for i := 0; i < int(safeCellsCount); i++ {
				var row, col uint16
				if err := binary.Read(buf, binary.LittleEndian, &row); err != nil {
					break
				}
				if err := binary.Read(buf, binary.LittleEndian, &col); err != nil {
					break
				}
				gs.SafeCells[i] = SafeCell{Row: int(row), Col: int(col)}
			}
		}
	}

	// Читаем CellHints (если есть данные)
	if buf.Len() > 0 {
		var hintsCount uint16
		if err := binary.Read(buf, binary.LittleEndian, &hintsCount); err == nil && hintsCount > 0 {
			gs.CellHints = make([]CellHint, hintsCount)
			for i := 0; i < int(hintsCount); i++ {
				var row, col uint16
				if err := binary.Read(buf, binary.LittleEndian, &row); err != nil {
					break
				}
				if err := binary.Read(buf, binary.LittleEndian, &col); err != nil {
					break
				}
				hintTypeByte, err := buf.ReadByte()
				if err != nil {
					break
				}
				var hintType string
				switch hintTypeByte {
				case 0:
					hintType = "MINE"
				case 1:
					hintType = "SAFE"
				case 2:
					hintType = "UNKNOWN"
				default:
					hintType = "UNKNOWN"
				}
				gs.CellHints[i] = CellHint{Row: int(row), Col: int(col), Type: hintType}
			}
		}
	}

	return gs, nil
}
