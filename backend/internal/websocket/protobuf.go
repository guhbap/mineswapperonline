package websocket

import (
	"fmt"
	"log"
	"minesweeperonline/internal/game"
	pb "minesweeperonline/proto"

	"google.golang.org/protobuf/proto"
)

// truncatePlayerID обрезает playerID до 5 символов
func truncatePlayerID(playerID string) string {
	if len(playerID) > 5 {
		return playerID[:5]
	}
	return playerID
}

// Типы клеток для бинарного формата
const (
	CellTypeClosed  = byte(255) // Закрыта
	CellTypeMine    = byte(9)   // Мина
	CellTypeSafe    = byte(10)  // Зеленая (SAFE)
	CellTypeUnknown = byte(11)  // Желтая (UNKNOWN)
	CellTypeDanger  = byte(12)  // Красная (MINE)
)

// byteToCellType преобразует byte в CellType enum
func byteToCellType(b byte) pb.CellType {
	if b <= 8 {
		return pb.CellType(b)
	}
	if b == CellTypeMine {
		return pb.CellType_CELL_TYPE_MINE
	}
	if b == CellTypeSafe {
		return pb.CellType_CELL_TYPE_SAFE
	}
	if b == CellTypeUnknown {
		return pb.CellType_CELL_TYPE_UNKNOWN
	}
	if b == CellTypeDanger {
		return pb.CellType_CELL_TYPE_DANGER
	}
	if b == CellTypeClosed {
		return pb.CellType_CELL_TYPE_CLOSED
	}
	return pb.CellType_CELL_TYPE_CLOSED
}


// EncodeChatProtobuf кодирует сообщение чата в protobuf формат
func EncodeChatProtobuf(msg *game.Message) ([]byte, error) {
	chatMsg := &pb.ChatMessage{
		PlayerId: truncatePlayerID(msg.PlayerID),
		Nickname: msg.Nickname,
		Color:    msg.Color,
		Text:     msg.Chat.Text,
		IsSystem: msg.Chat.IsSystem,
		Action:   msg.Chat.Action,
		Row:      int32(msg.Chat.Row),
		Col:      int32(msg.Chat.Col),
	}

	wsMsg := &pb.WebSocketMessage{
		Message: &pb.WebSocketMessage_Chat{
			Chat: chatMsg,
		},
	}

	return proto.Marshal(wsMsg)
}

// EncodeCursorProtobuf кодирует позицию курсора в protobuf формат
func EncodeCursorProtobuf(msg *game.Message) ([]byte, error) {
	cursorMsg := &pb.CursorMessage{
		PlayerId: truncatePlayerID(msg.Cursor.PlayerID),
		Nickname: msg.Nickname,
		Color:    msg.Color,
		X:        msg.Cursor.X,
		Y:        msg.Cursor.Y,
	}

	wsMsg := &pb.WebSocketMessage{
		Message: &pb.WebSocketMessage_Cursor{
			Cursor: cursorMsg,
		},
	}

	return proto.Marshal(wsMsg)
}

// EncodePlayersProtobuf кодирует список игроков в protobuf формат
func EncodePlayersProtobuf(players []map[string]string) ([]byte, error) {
	playerList := make([]*pb.Player, len(players))
	for i, p := range players {
		playerList[i] = &pb.Player{
			Id:       truncatePlayerID(p["id"]),
			Nickname: p["nickname"],
			Color:    p["color"],
		}
	}

	playersMsg := &pb.PlayersMessage{
		Players: playerList,
	}

	wsMsg := &pb.WebSocketMessage{
		Message: &pb.WebSocketMessage_Players{
			Players: playersMsg,
		},
	}

	return proto.Marshal(wsMsg)
}

// EncodePongProtobuf кодирует pong сообщение в protobuf формат
func EncodePongProtobuf() ([]byte, error) {
	pongMsg := &pb.PongMessage{}

	wsMsg := &pb.WebSocketMessage{
		Message: &pb.WebSocketMessage_Pong{
			Pong: pongMsg,
		},
	}

	return proto.Marshal(wsMsg)
}

// EncodeErrorProtobuf кодирует сообщение об ошибке в protobuf формат
func EncodeErrorProtobuf(errorMsg string) ([]byte, error) {
	errorMsgProto := &pb.ErrorMessage{
		Error: errorMsg,
	}

	wsMsg := &pb.WebSocketMessage{
		Message: &pb.WebSocketMessage_Error{
			Error: errorMsgProto,
		},
	}

	return proto.Marshal(wsMsg)
}

// CellUpdate представляет обновление одной клетки
type CellUpdate struct {
	Row  int
	Col  int
	Type byte
}


// decodeClientMessageProtobuf декодирует ClientMessage из protobuf формата
func decodeClientMessageProtobuf(data []byte) (*game.Message, error) {
	var clientMsg pb.ClientMessage
	if err := proto.Unmarshal(data, &clientMsg); err != nil {
		return nil, fmt.Errorf("ошибка unmarshal protobuf: %w", err)
	}

	msg := &game.Message{}

	// Детальное логирование для диагностики
	log.Printf("[DECODE] Декодирование ClientMessage: nickname=%v, cursor=%v, cellClick=%v, hint=%v, newGame=%v, chat=%v, ping=%v",
		clientMsg.GetNickname() != "",
		clientMsg.GetCursor() != nil,
		clientMsg.GetCellClick() != nil,
		clientMsg.GetHint() != nil,
		clientMsg.GetNewGame() != nil,
		clientMsg.GetChat() != nil,
		clientMsg.GetPing() != nil)

	switch {
	case clientMsg.GetNickname() != "":
		msg.Type = "nickname"
		msg.Nickname = clientMsg.GetNickname()
		log.Printf("[DECODE] Определен тип: nickname=%s", msg.Nickname)

	case clientMsg.GetCursor() != nil:
		cursorProto := clientMsg.GetCursor()
		msg.Type = "cursor"
		msg.Cursor = &game.CursorPosition{
			PlayerID: cursorProto.PlayerId,
			X:        cursorProto.X,
			Y:        cursorProto.Y,
		}
		log.Printf("[DECODE] Определен тип: cursor, x=%.2f, y=%.2f", msg.Cursor.X, msg.Cursor.Y)

	case clientMsg.GetCellClick() != nil:
		clickProto := clientMsg.GetCellClick()
		msg.Type = "cellClick"
		msg.CellClick = &game.CellClick{
			Row:  int(clickProto.Row),
			Col:  int(clickProto.Col),
			Flag: clickProto.Flag,
		}
		log.Printf("[DECODE] Определен тип: cellClick, row=%d, col=%d, flag=%v", msg.CellClick.Row, msg.CellClick.Col, msg.CellClick.Flag)

	case clientMsg.GetHint() != nil:
		hintProto := clientMsg.GetHint()
		msg.Type = "hint"
		msg.Hint = &game.Hint{
			Row: int(hintProto.Row),
			Col: int(hintProto.Col),
		}
		log.Printf("[DECODE] Определен тип: hint, row=%d, col=%d", msg.Hint.Row, msg.Hint.Col)

	case clientMsg.GetNewGame() != nil:
		msg.Type = "newGame"
		log.Printf("[DECODE] Определен тип: newGame")

	case clientMsg.GetChat() != nil:
		chatProto := clientMsg.GetChat()
		msg.Type = "chat"
		msg.Chat = &game.ChatMessage{
			Text:     chatProto.Text,
			IsSystem: chatProto.IsSystem,
			Action:   chatProto.Action,
			Row:      int(chatProto.Row),
			Col:      int(chatProto.Col),
		}
		log.Printf("[DECODE] Определен тип: chat, text=%s", msg.Chat.Text)

	case clientMsg.GetPing() != nil:
		msg.Type = "ping"
		log.Printf("[DECODE] Определен тип: ping")

	default:
		log.Printf("[DECODE] ОШИБКА: неизвестный тип сообщения в ClientMessage")
		return nil, fmt.Errorf("unknown message type in ClientMessage")
	}

	log.Printf("[DECODE] Сообщение успешно декодировано: type=%s", msg.Type)
	return msg, nil
}

