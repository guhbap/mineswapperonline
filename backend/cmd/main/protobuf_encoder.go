package main

import (
	"fmt"
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

// encodeGameStateProtobuf кодирует GameState в protobuf формат
func encodeGameStateProtobuf(gs *GameState) ([]byte, error) {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	// Создаем Board
	rows := make([]*pb.Row, gs.Rows)
	for i := 0; i < gs.Rows; i++ {
		cells := make([]*pb.Cell, gs.Cols)
		for j := 0; j < gs.Cols; j++ {
			cell := gs.Board[i][j]
			cells[j] = &pb.Cell{
				IsMine:        cell.IsMine,
				IsRevealed:    cell.IsRevealed,
				IsFlagged:     cell.IsFlagged,
				NeighborMines: int32(cell.NeighborMines),
				FlagColor:     cell.FlagColor,
			}
		}
		rows[i] = &pb.Row{Cells: cells}
	}

	// Создаем SafeCells
	safeCells := make([]*pb.SafeCell, len(gs.SafeCells))
	for i, sc := range gs.SafeCells {
		safeCells[i] = &pb.SafeCell{
			Row: int32(sc.Row),
			Col: int32(sc.Col),
		}
	}

	// Создаем CellHints
	cellHints := make([]*pb.CellHint, len(gs.CellHints))
	for i, hint := range gs.CellHints {
		cellHints[i] = &pb.CellHint{
			Row:  int32(hint.Row),
			Col:  int32(hint.Col),
			Type: hint.Type,
		}
	}

	gameStateMsg := &pb.GameStateMessage{
		Board:          &pb.Board{Rows: rows},
		Rows:           int32(gs.Rows),
		Cols:           int32(gs.Cols),
		Mines:          int32(gs.Mines),
		Seed:           gs.Seed,
		GameOver:       gs.GameOver,
		GameWon:        gs.GameWon,
		Revealed:       int32(gs.Revealed),
		HintsUsed:      int32(gs.HintsUsed),
		SafeCells:      safeCells,
		CellHints:      cellHints,
		LoserPlayerId:  truncatePlayerID(gs.LoserPlayerID),
		LoserNickname:  gs.LoserNickname,
	}

	wsMsg := &pb.WebSocketMessage{
		Message: &pb.WebSocketMessage_GameState{
			GameState: gameStateMsg,
		},
	}

	return proto.Marshal(wsMsg)
}

// encodeChatProtobuf кодирует сообщение чата в protobuf формат
func encodeChatProtobuf(msg *Message) ([]byte, error) {
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

// encodeCursorProtobuf кодирует позицию курсора в protobuf формат
func encodeCursorProtobuf(msg *Message) ([]byte, error) {
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

// encodePlayersProtobuf кодирует список игроков в protobuf формат
func encodePlayersProtobuf(players []map[string]string) ([]byte, error) {
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

// encodePongProtobuf кодирует pong сообщение в protobuf формат
func encodePongProtobuf() ([]byte, error) {
	pongMsg := &pb.PongMessage{}

	wsMsg := &pb.WebSocketMessage{
		Message: &pb.WebSocketMessage_Pong{
			Pong: pongMsg,
		},
	}

	return proto.Marshal(wsMsg)
}

// encodeErrorProtobuf кодирует сообщение об ошибке в protobuf формат
func encodeErrorProtobuf(errorMsg string) ([]byte, error) {
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

// byteToCellType преобразует byte в CellType enum
func byteToCellType(b byte) pb.CellType {
	// Значения 0-8 используются для открытых клеток с количеством соседних мин
	if b <= 8 {
		return pb.CellType(b)
	}
	// Значение 9 = мина
	if b == CellTypeMine {
		return pb.CellType_CELL_TYPE_MINE
	}
	// Значение 10 = зеленая (SAFE)
	if b == CellTypeSafe {
		return pb.CellType_CELL_TYPE_SAFE
	}
	// Значение 11 = желтая (UNKNOWN)
	if b == CellTypeUnknown {
		return pb.CellType_CELL_TYPE_UNKNOWN
	}
	// Значение 12 = красная (MINE/DANGER)
	if b == CellTypeDanger {
		return pb.CellType_CELL_TYPE_DANGER
	}
	// Значение 255 = закрыта
	if b == CellTypeClosed {
		return pb.CellType_CELL_TYPE_CLOSED
	}
	// По умолчанию возвращаем закрытую
	return pb.CellType_CELL_TYPE_CLOSED
}

// encodeCellUpdateProtobuf кодирует обновления клеток в protobuf формат
func encodeCellUpdateProtobuf(updates []CellUpdate, gameOver bool, gameWon bool, revealed int, hintsUsed int, loserPlayerID string, loserNickname string) ([]byte, error) {
	cellUpdates := make([]*pb.CellUpdate, len(updates))
	for i, update := range updates {
		cellUpdates[i] = &pb.CellUpdate{
			Row:  int32(update.Row),
			Col:  int32(update.Col),
			Type: byteToCellType(update.Type),
		}
	}

	cellUpdateMsg := &pb.CellUpdateMessage{
		GameOver:        gameOver,
		GameWon:         gameWon,
		Revealed:        int32(revealed),
		HintsUsed:       int32(hintsUsed),
		LoserPlayerId:   truncatePlayerID(loserPlayerID),
		LoserNickname:   loserNickname,
		Updates:         cellUpdates,
	}

	wsMsg := &pb.WebSocketMessage{
		Message: &pb.WebSocketMessage_CellUpdate{
			CellUpdate: cellUpdateMsg,
		},
	}

	return proto.Marshal(wsMsg)
}

// decodeClientMessageProtobuf декодирует ClientMessage из protobuf формата
func decodeClientMessageProtobuf(data []byte) (*Message, error) {
	var clientMsg pb.ClientMessage
	if err := proto.Unmarshal(data, &clientMsg); err != nil {
		return nil, err
	}

	msg := &Message{}

	// Обрабатываем различные типы сообщений
	switch {
	case clientMsg.GetNickname() != "":
		msg.Type = "nickname"
		msg.Nickname = clientMsg.GetNickname()

	case clientMsg.GetCursor() != nil:
		cursorProto := clientMsg.GetCursor()
		msg.Type = "cursor"
		msg.Cursor = &CursorPosition{
			PlayerID: cursorProto.PlayerId,
			X:        cursorProto.X,
			Y:        cursorProto.Y,
		}

	case clientMsg.GetCellClick() != nil:
		clickProto := clientMsg.GetCellClick()
		msg.Type = "cellClick"
		msg.CellClick = &CellClick{
			Row:  int(clickProto.Row),
			Col:  int(clickProto.Col),
			Flag: clickProto.Flag,
		}

	case clientMsg.GetHint() != nil:
		hintProto := clientMsg.GetHint()
		msg.Type = "hint"
		msg.Hint = &Hint{
			Row: int(hintProto.Row),
			Col: int(hintProto.Col),
		}

	case clientMsg.GetNewGame() != nil:
		msg.Type = "newGame"

	case clientMsg.GetChat() != nil:
		chatProto := clientMsg.GetChat()
		msg.Type = "chat"
		msg.Chat = &ChatMessage{
			Text:     chatProto.Text,
			IsSystem: chatProto.IsSystem,
			Action:   chatProto.Action,
			Row:      int(chatProto.Row),
			Col:      int(chatProto.Col),
		}

	case clientMsg.GetPing() != nil:
		msg.Type = "ping"

	default:
		return nil, fmt.Errorf("unknown message type in ClientMessage")
	}

	return msg, nil
}
