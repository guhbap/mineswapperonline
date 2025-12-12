package main

import (
	"bytes"
	"encoding/binary"
)

// encodeGameStateBinary кодирует GameState в бинарный формат
// Формат:
// - 2 байта: Rows (uint16)
// - 2 байта: Cols (uint16)
// - 2 байта: Mines (uint16)
// - 2 байта: Revealed (uint16)
// - 1 байт: HintsUsed (количество использованных подсказок, 0-3)
// - 1 байт: Флаги (бит 0: GameOver, бит 1: GameWon)
// - 1 байт: Длина LoserPlayerID (0-5)
// - 5 байт: LoserPlayerID (ASCII)
// - 1 байт: Длина LoserNickname
// - N байт: LoserNickname (UTF-8)
// - Rows*Cols байт: Board (каждая Cell = 1 байт)
// - 1 байт: Количество флагов с цветами
// - Для каждого флага: 2 байта cellKey (uint16), 1 байт длина цвета, N байт цвет
func encodeGameStateBinary(gs *GameState) ([]byte, error) {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	buf := new(bytes.Buffer)

	// Записываем размеры
	binary.Write(buf, binary.LittleEndian, uint16(gs.Rows))
	binary.Write(buf, binary.LittleEndian, uint16(gs.Cols))
	binary.Write(buf, binary.LittleEndian, uint16(gs.Mines))
	binary.Write(buf, binary.LittleEndian, uint16(gs.Revealed))

	// HintsUsed (1 байт, максимум 3)
	hintsUsed := byte(gs.HintsUsed)
	if hintsUsed > 3 {
		hintsUsed = 3
	}
	buf.WriteByte(hintsUsed)

	// Флаги (1 байт)
	flags := byte(0)
	if gs.GameOver {
		flags |= 1 << 0
	}
	if gs.GameWon {
		flags |= 1 << 1
	}
	buf.WriteByte(flags)

	// LoserPlayerID (максимум 5 символов)
	loserPID := truncatePlayerID(gs.LoserPlayerID)
	loserPIDLen := byte(len(loserPID))
	if loserPIDLen > 5 {
		loserPIDLen = 5
	}
	buf.WriteByte(loserPIDLen)
	pidBytes := make([]byte, 5)
	if loserPIDLen > 0 {
		copy(pidBytes, []byte(loserPID))
	}
	buf.Write(pidBytes) // Всегда записываем 5 байт

	// LoserNickname
	loserNicknameBytes := []byte(gs.LoserNickname)
	nicknameLen := byte(len(loserNicknameBytes))
	if nicknameLen > 255 {
		nicknameLen = 255
		loserNicknameBytes = loserNicknameBytes[:255]
	}
	buf.WriteByte(nicknameLen)
	if nicknameLen > 0 {
		buf.Write(loserNicknameBytes)
	}

	// Board: каждая Cell упакована в 1 байт
	// Бит 0: IsMine
	// Бит 1: IsRevealed
	// Бит 2: IsFlagged
	// Бит 3-6: NeighborMines (0-8, 4 бита)
	// Бит 7: зарезервирован
	for i := 0; i < gs.Rows; i++ {
		for j := 0; j < gs.Cols; j++ {
			cell := gs.Board[i][j]
			cellByte := byte(0)
			if cell.IsMine {
				cellByte |= 1 << 0
			}
			if cell.IsRevealed {
				cellByte |= 1 << 1
			}
			if cell.IsFlagged {
				cellByte |= 1 << 2
			}
			// NeighborMines (0-8) упаковываем в биты 3-6
			neighborMines := byte(cell.NeighborMines)
			if neighborMines > 15 {
				neighborMines = 15 // Ограничиваем максимумом
			}
			cellByte |= neighborMines << 3
			buf.WriteByte(cellByte)
		}
	}

	// Секция с цветами флагов
	// Собираем все флаги с цветами
	flagColors := make(map[uint16]string)
	for i := 0; i < gs.Rows; i++ {
		for j := 0; j < gs.Cols; j++ {
			cell := gs.Board[i][j]
			if cell.IsFlagged && cell.FlagColor != "" {
				cellKey := uint16(i*gs.Cols + j)
				flagColors[cellKey] = cell.FlagColor
			}
		}
	}

	// Записываем количество флагов с цветами
	flagCount := byte(len(flagColors))
	if flagCount > 255 {
		flagCount = 255
	}
	buf.WriteByte(flagCount)

	// Записываем цвета флагов
	count := 0
	for cellKey, color := range flagColors {
		if count >= 255 {
			break
		}
		// Записываем cellKey (2 байта)
		binary.Write(buf, binary.LittleEndian, cellKey)
		// Записываем длину цвета (максимум 7 для hex цвета #RRGGBB)
		colorBytes := []byte(color)
		colorLen := byte(len(colorBytes))
		if colorLen > 7 {
			colorLen = 7
			colorBytes = colorBytes[:7]
		}
		buf.WriteByte(colorLen)
		// Записываем цвет
		if colorLen > 0 {
			buf.Write(colorBytes)
		}
		count++
	}

	return buf.Bytes(), nil
}

// decodeGameStateBinary декодирует бинарный формат в GameState
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

	return gs, nil
}

