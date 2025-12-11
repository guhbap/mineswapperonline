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
// - 1 байт: Флаги (бит 0: GameOver, бит 1: GameWon)
// - 1 байт: Длина LoserPlayerID (0-5)
// - 5 байт: LoserPlayerID (ASCII)
// - 1 байт: Длина LoserNickname
// - N байт: LoserNickname (UTF-8)
// - Rows*Cols байт: Board (каждая Cell = 1 байт)
func encodeGameStateBinary(gs *GameState) ([]byte, error) {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	buf := new(bytes.Buffer)

	// Записываем размеры
	binary.Write(buf, binary.LittleEndian, uint16(gs.Rows))
	binary.Write(buf, binary.LittleEndian, uint16(gs.Cols))
	binary.Write(buf, binary.LittleEndian, uint16(gs.Mines))
	binary.Write(buf, binary.LittleEndian, uint16(gs.Revealed))

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

	return gs, nil
}

