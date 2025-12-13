package main

// Этот файл будет содержать функции для кодирования в protobuf
// после генерации кода из .proto файла командой:
// protoc --go_out=backend/proto --go_opt=paths=source_relative proto/messages.proto
//
// Временная заглушка - после генерации кода эти функции будут использовать
// сгенерированные структуры из backend/proto/messages.pb.go

// После генерации кода из .proto, раскомментируйте и обновите эти функции:

/*
import (
	"minesweeperonline/proto"
	"google.golang.org/protobuf/proto"
)

// encodeGameStateProtobuf кодирует GameState в protobuf формат
func encodeGameStateProtobuf(gs *GameState) ([]byte, error) {
	// Реализация будет после генерации кода
	return nil, nil
}

// encodeChatProtobuf кодирует сообщение чата в protobuf формат
func encodeChatProtobuf(msg *Message) ([]byte, error) {
	// Реализация будет после генерации кода
	return nil, nil
}

// encodeCursorProtobuf кодирует позицию курсора в protobuf формат
func encodeCursorProtobuf(msg *Message) ([]byte, error) {
	// Реализация будет после генерации кода
	return nil, nil
}

// encodePlayersProtobuf кодирует список игроков в protobuf формат
func encodePlayersProtobuf(players []map[string]string) ([]byte, error) {
	// Реализация будет после генерации кода
	return nil, nil
}

// encodePongProtobuf кодирует pong сообщение в protobuf формат
func encodePongProtobuf() ([]byte, error) {
	// Реализация будет после генерации кода
	return nil, nil
}

// encodeErrorProtobuf кодирует сообщение об ошибке в protobuf формат
func encodeErrorProtobuf(errorMsg string) ([]byte, error) {
	// Реализация будет после генерации кода
	return nil, nil
}

// encodeCellUpdateProtobuf кодирует обновления клеток в protobuf формат
func encodeCellUpdateProtobuf(updates []CellUpdate, gameOver bool, gameWon bool, revealed int, hintsUsed int, loserPlayerID string, loserNickname string) ([]byte, error) {
	// Реализация будет после генерации кода
	return nil, nil
}
*/

// Временные функции-заглушки, которые используют текущий бинарный формат
// После генерации protobuf кода эти функции будут заменены

func encodeGameStateProtobuf(gs *GameState) ([]byte, error) {
	// Пока используем бинарный формат
	return encodeGameStateBinary(gs)
}

func encodeChatProtobuf(msg *Message) ([]byte, error) {
	// Пока используем бинарный формат
	return encodeChatBinary(msg)
}

func encodeCursorProtobuf(msg *Message) ([]byte, error) {
	// Пока используем бинарный формат
	return encodeCursorBinary(msg)
}

func encodePlayersProtobuf(players []map[string]string) ([]byte, error) {
	// Пока используем бинарный формат
	return encodePlayersBinary(players)
}

func encodePongProtobuf() ([]byte, error) {
	// Пока используем бинарный формат
	return encodePongBinary()
}

func encodeErrorProtobuf(errorMsg string) ([]byte, error) {
	// Пока используем бинарный формат
	return encodeErrorBinary(errorMsg)
}

func encodeCellUpdateProtobuf(updates []CellUpdate, gameOver bool, gameWon bool, revealed int, hintsUsed int, loserPlayerID string, loserNickname string) ([]byte, error) {
	// Пока используем бинарный формат
	return encodeCellUpdateBinary(updates, gameOver, gameWon, revealed, hintsUsed, loserPlayerID, loserNickname)
}

