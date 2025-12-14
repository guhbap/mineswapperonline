package game

import (
	pb "minesweeperonline/proto"

	"google.golang.org/protobuf/proto"
)

// EncodeChatProtobuf кодирует сообщение чата в protobuf формат
func EncodeChatProtobuf(msg *Message) ([]byte, error) {
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
func EncodeCursorProtobuf(msg *Message) ([]byte, error) {
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

