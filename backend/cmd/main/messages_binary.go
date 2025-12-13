package main

import (
	"bytes"
	"encoding/binary"
)

// Типы бинарных сообщений
const (
	MessageTypeGameState = byte(0)
	MessageTypeChat      = byte(1)
	MessageTypeCursor    = byte(2)
	MessageTypePlayers   = byte(3)
	MessageTypePong      = byte(4)
	MessageTypeError      = byte(5)
	MessageTypeCellUpdate = byte(6)
)

// Типы клеток для бинарного формата
const (
	CellTypeClosed  = byte(255) // Закрыта (используем 255 вместо 0, чтобы не конфликтовать с количеством мин)
	CellTypeMine    = byte(9)    // Мина
	CellTypeSafe    = byte(10)   // Зеленая (SAFE) - для режима обучения
	CellTypeUnknown = byte(11)   // Желтая (UNKNOWN) - для режима обучения
	CellTypeDanger  = byte(12)   // Красная (MINE) - для режима обучения
	// Значения 0-8 используются для открытых клеток с количеством соседних мин (0-8)
)

// encodeChatBinary кодирует сообщение чата в бинарный формат
// Формат:
// - 1 байт: тип сообщения (1)
// - 1 байт: длина PlayerID (0-5)
// - 5 байт: PlayerID (ASCII)
// - 1 байт: длина Nickname (0-255)
// - N байт: Nickname (UTF-8)
// - 1 байт: длина Color (0-7)
// - N байт: Color (UTF-8, hex цвет)
// - 1 байт: длина Text (0-255)
// - N байт: Text (UTF-8)
// - 1 байт: флаги (бит 0: IsSystem, бит 1: есть Action, бит 2: есть Row/Col)
// - Если есть Action: 1 байт длина Action, N байт Action
// - Если есть Row/Col: 2 байта Row (uint16), 2 байта Col (uint16)
func encodeChatBinary(msg *Message) ([]byte, error) {
	buf := new(bytes.Buffer)
	buf.WriteByte(MessageTypeChat)

	// PlayerID (максимум 5 символов)
	pid := truncatePlayerID(msg.PlayerID)
	pidLen := byte(len(pid))
	if pidLen > 5 {
		pidLen = 5
	}
	buf.WriteByte(pidLen)
	pidBytes := make([]byte, 5)
	if pidLen > 0 {
		copy(pidBytes, []byte(pid))
	}
	buf.Write(pidBytes)

	// Nickname
	nicknameBytes := []byte(msg.Nickname)
	nicknameLen := byte(len(nicknameBytes))
	if len(nicknameBytes) > 255 {
		nicknameLen = 255
		nicknameBytes = nicknameBytes[:255]
	}
	buf.WriteByte(nicknameLen)
	if nicknameLen > 0 {
		buf.Write(nicknameBytes)
	}

	// Color
	colorBytes := []byte(msg.Color)
	colorLen := byte(len(colorBytes))
	if colorLen > 7 {
		colorLen = 7
		colorBytes = colorBytes[:7]
	}
	buf.WriteByte(colorLen)
	if colorLen > 0 {
		buf.Write(colorBytes)
	}

	// Text
	textBytes := []byte(msg.Chat.Text)
	textLen := byte(len(textBytes))
	if len(textBytes) > 255 {
		textLen = 255
		textBytes = textBytes[:255]
	}
	buf.WriteByte(textLen)
	if textLen > 0 {
		buf.Write(textBytes)
	}

	// Флаги
	flags := byte(0)
	if msg.Chat.IsSystem {
		flags |= 1 << 0
	}
	if msg.Chat.Action != "" {
		flags |= 1 << 1
	}
	if msg.Chat.Row >= 0 && msg.Chat.Col >= 0 {
		flags |= 1 << 2
	}
	buf.WriteByte(flags)

	// Action (если есть)
	if msg.Chat.Action != "" {
		actionBytes := []byte(msg.Chat.Action)
		actionLen := byte(len(actionBytes))
		if actionLen > 10 {
			actionLen = 10
			actionBytes = actionBytes[:10]
		}
		buf.WriteByte(actionLen)
		if actionLen > 0 {
			buf.Write(actionBytes)
		}
	}

	// Row/Col (если есть)
	if msg.Chat.Row >= 0 && msg.Chat.Col >= 0 {
		binary.Write(buf, binary.LittleEndian, uint16(msg.Chat.Row))
		binary.Write(buf, binary.LittleEndian, uint16(msg.Chat.Col))
	}

	return buf.Bytes(), nil
}

// decodeChatBinary декодирует бинарное сообщение чата
func decodeChatBinary(data []byte) (*Message, error) {
	buf := bytes.NewReader(data)
	msg := &Message{Type: "chat", Chat: &ChatMessage{}}

	// Пропускаем тип сообщения (уже прочитан)
	// Читаем PlayerID
	pidLen, _ := buf.ReadByte()
	pidBytes := make([]byte, 5)
	buf.Read(pidBytes)
	if pidLen > 0 && pidLen <= 5 {
		msg.PlayerID = string(pidBytes[:pidLen])
	}

	// Читаем Nickname
	nicknameLen, _ := buf.ReadByte()
	if nicknameLen > 0 {
		nicknameBytes := make([]byte, nicknameLen)
		buf.Read(nicknameBytes)
		msg.Nickname = string(nicknameBytes)
	}

	// Читаем Color
	colorLen, _ := buf.ReadByte()
	if colorLen > 0 {
		colorBytes := make([]byte, colorLen)
		buf.Read(colorBytes)
		msg.Color = string(colorBytes)
	}

	// Читаем Text
	textLen, _ := buf.ReadByte()
	if textLen > 0 {
		textBytes := make([]byte, textLen)
		buf.Read(textBytes)
		msg.Chat.Text = string(textBytes)
	}

	// Читаем флаги
	flags, _ := buf.ReadByte()
	msg.Chat.IsSystem = (flags & (1 << 0)) != 0
	hasAction := (flags & (1 << 1)) != 0
	hasRowCol := (flags & (1 << 2)) != 0

	// Читаем Action (если есть)
	if hasAction {
		actionLen, _ := buf.ReadByte()
		if actionLen > 0 {
			actionBytes := make([]byte, actionLen)
			buf.Read(actionBytes)
			msg.Chat.Action = string(actionBytes)
		}
	}

	// Читаем Row/Col (если есть)
	if hasRowCol {
		var row, col uint16
		binary.Read(buf, binary.LittleEndian, &row)
		binary.Read(buf, binary.LittleEndian, &col)
		msg.Chat.Row = int(row)
		msg.Chat.Col = int(col)
	}

	return msg, nil
}

// encodeCursorBinary кодирует позицию курсора в бинарный формат
// Формат:
// - 1 байт: тип сообщения (2)
// - 1 байт: длина PlayerID (0-5)
// - 5 байт: PlayerID (ASCII)
// - 1 байт: длина Nickname (0-255)
// - N байт: Nickname (UTF-8)
// - 1 байт: длина Color (0-7)
// - N байт: Color (UTF-8)
// - 8 байт: X (float64)
// - 8 байт: Y (float64)
func encodeCursorBinary(msg *Message) ([]byte, error) {
	buf := new(bytes.Buffer)
	buf.WriteByte(MessageTypeCursor)

	// PlayerID
	pid := truncatePlayerID(msg.Cursor.PlayerID)
	pidLen := byte(len(pid))
	if pidLen > 5 {
		pidLen = 5
	}
	buf.WriteByte(pidLen)
	pidBytes := make([]byte, 5)
	if pidLen > 0 {
		copy(pidBytes, []byte(pid))
	}
	buf.Write(pidBytes)

	// Nickname
	nicknameBytes := []byte(msg.Nickname)
	nicknameLen := byte(len(nicknameBytes))
	if len(nicknameBytes) > 255 {
		nicknameLen = 255
		nicknameBytes = nicknameBytes[:255]
	}
	buf.WriteByte(nicknameLen)
	if nicknameLen > 0 {
		buf.Write(nicknameBytes)
	}

	// Color
	colorBytes := []byte(msg.Color)
	colorLen := byte(len(colorBytes))
	if colorLen > 7 {
		colorLen = 7
		colorBytes = colorBytes[:7]
	}
	buf.WriteByte(colorLen)
	if colorLen > 0 {
		buf.Write(colorBytes)
	}

	// X, Y (float64)
	binary.Write(buf, binary.LittleEndian, msg.Cursor.X)
	binary.Write(buf, binary.LittleEndian, msg.Cursor.Y)

	return buf.Bytes(), nil
}

// decodeCursorBinary декодирует бинарное сообщение курсора
func decodeCursorBinary(data []byte) (*Message, error) {
	buf := bytes.NewReader(data)
	msg := &Message{Type: "cursor", Cursor: &CursorPosition{}}

	// Пропускаем тип сообщения
	// Читаем PlayerID
	pidLen, _ := buf.ReadByte()
	pidBytes := make([]byte, 5)
	buf.Read(pidBytes)
	if pidLen > 0 && pidLen <= 5 {
		msg.Cursor.PlayerID = string(pidBytes[:pidLen])
	}

	// Читаем Nickname
	nicknameLen, _ := buf.ReadByte()
	if nicknameLen > 0 {
		nicknameBytes := make([]byte, nicknameLen)
		buf.Read(nicknameBytes)
		msg.Nickname = string(nicknameBytes)
	}

	// Читаем Color
	colorLen, _ := buf.ReadByte()
	if colorLen > 0 {
		colorBytes := make([]byte, colorLen)
		buf.Read(colorBytes)
		msg.Color = string(colorBytes)
	}

	// Читаем X, Y
	var x, y float64
	binary.Read(buf, binary.LittleEndian, &x)
	binary.Read(buf, binary.LittleEndian, &y)
	msg.Cursor.X = x
	msg.Cursor.Y = y

	return msg, nil
}

// encodePlayersBinary кодирует список игроков в бинарный формат
// Формат:
// - 1 байт: тип сообщения (3)
// - 1 байт: количество игроков (0-255)
// - Для каждого игрока:
//   - 1 байт: длина ID (0-5)
//   - 5 байт: ID (ASCII)
//   - 1 байт: длина Nickname (0-255)
//   - N байт: Nickname (UTF-8)
//   - 1 байт: длина Color (0-7)
//   - N байт: Color (UTF-8)
func encodePlayersBinary(players []map[string]string) ([]byte, error) {
	buf := new(bytes.Buffer)
	buf.WriteByte(MessageTypePlayers)

	playerCount := byte(len(players))
	if len(players) > 255 {
		playerCount = 255
	}
	buf.WriteByte(playerCount)

	for i := 0; i < int(playerCount); i++ {
		player := players[i]

		// ID
		id := truncatePlayerID(player["id"])
		idLen := byte(len(id))
		if idLen > 5 {
			idLen = 5
		}
		buf.WriteByte(idLen)
		idBytes := make([]byte, 5)
		if idLen > 0 {
			copy(idBytes, []byte(id))
		}
		buf.Write(idBytes)

		// Nickname
		nicknameBytes := []byte(player["nickname"])
		nicknameLen := byte(len(nicknameBytes))
		if len(nicknameBytes) > 255 {
			nicknameLen = 255
			nicknameBytes = nicknameBytes[:255]
		}
		buf.WriteByte(nicknameLen)
		if nicknameLen > 0 {
			buf.Write(nicknameBytes)
		}

		// Color
		colorBytes := []byte(player["color"])
		colorLen := byte(len(colorBytes))
		if colorLen > 7 {
			colorLen = 7
			colorBytes = colorBytes[:7]
		}
		buf.WriteByte(colorLen)
		if colorLen > 0 {
			buf.Write(colorBytes)
		}
	}

	return buf.Bytes(), nil
}

// decodePlayersBinary декодирует бинарное сообщение списка игроков
func decodePlayersBinary(data []byte) ([]map[string]string, error) {
	buf := bytes.NewReader(data)
	// Пропускаем тип сообщения

	playerCount, _ := buf.ReadByte()
	players := make([]map[string]string, 0, playerCount)

	for i := 0; i < int(playerCount); i++ {
		player := make(map[string]string)

		// ID
		idLen, _ := buf.ReadByte()
		idBytes := make([]byte, 5)
		buf.Read(idBytes)
		if idLen > 0 && idLen <= 5 {
			player["id"] = string(idBytes[:idLen])
		}

		// Nickname
		nicknameLen, _ := buf.ReadByte()
		if nicknameLen > 0 {
			nicknameBytes := make([]byte, nicknameLen)
			buf.Read(nicknameBytes)
			player["nickname"] = string(nicknameBytes)
		}

		// Color
		colorLen, _ := buf.ReadByte()
		if colorLen > 0 {
			colorBytes := make([]byte, colorLen)
			buf.Read(colorBytes)
			player["color"] = string(colorBytes)
		}

		players = append(players, player)
	}

	return players, nil
}

// encodePongBinary кодирует pong сообщение в бинарный формат
// Формат:
// - 1 байт: тип сообщения (4)
func encodePongBinary() ([]byte, error) {
	return []byte{MessageTypePong}, nil
}

// encodeErrorBinary кодирует сообщение об ошибке в бинарный формат
// Формат:
// - 1 байт: тип сообщения (5)
// - 1 байт: длина сообщения об ошибке (0-255)
// - N байт: сообщение об ошибке (UTF-8)
func encodeErrorBinary(errorMsg string) ([]byte, error) {
	buf := new(bytes.Buffer)
	buf.WriteByte(MessageTypeError)

	errorBytes := []byte(errorMsg)
	errorLen := byte(len(errorBytes))
	if len(errorBytes) > 255 {
		errorLen = 255
		errorBytes = errorBytes[:255]
	}
	buf.WriteByte(errorLen)
	if errorLen > 0 {
		buf.Write(errorBytes)
	}

	return buf.Bytes(), nil
}

// decodeErrorBinary декодирует бинарное сообщение об ошибке
func decodeErrorBinary(data []byte) (string, error) {
	buf := bytes.NewReader(data)
	// Пропускаем тип сообщения

	errorLen, _ := buf.ReadByte()
	if errorLen > 0 {
		errorBytes := make([]byte, errorLen)
		buf.Read(errorBytes)
		return string(errorBytes), nil
	}
	return "", nil
}

