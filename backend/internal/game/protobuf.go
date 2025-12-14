package game

import (
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

// EncodeGameStateProtobuf кодирует game.GameState в protobuf формат
func EncodeGameStateProtobuf(gs *GameState) ([]byte, error) {
	gs.Mu.RLock()
	defer gs.Mu.RUnlock()

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

	safeCells := make([]*pb.SafeCell, len(gs.SafeCells))
	for i, sc := range gs.SafeCells {
		safeCells[i] = &pb.SafeCell{
			Row: int32(sc.Row),
			Col: int32(sc.Col),
		}
	}

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

// CellUpdate представляет обновление одной клетки
type CellUpdate struct {
	Row  int
	Col  int
	Type byte
}

// getCellType возвращает тип клетки для бинарного формата
func getCellType(cell *Cell, row, col int, gameMode string, cellHints []CellHint) byte {
	if cell.IsRevealed {
		if cell.IsMine {
			return CellTypeMine
		}
		if cell.NeighborMines > 8 {
			return 8
		}
		return byte(cell.NeighborMines)
	}

	if !cell.IsFlagged {
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

	return CellTypeClosed
}

// CollectCellUpdates собирает измененные клетки из gameState
func CollectCellUpdates(room *Room, changedCells map[[2]int]bool) []CellUpdate {
	updates := make([]CellUpdate, 0)

	room.GameState.Mu.RLock()
	cellHints := room.GameState.CellHints
	board := room.GameState.Board
	rows := room.GameState.Rows
	cols := room.GameState.Cols
	room.GameState.Mu.RUnlock()

	gameMode := room.GameMode

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

// EncodeCellUpdateProtobuf кодирует обновления клеток в protobuf формат
func EncodeCellUpdateProtobuf(updates []CellUpdate, gameOver bool, gameWon bool, revealed int, hintsUsed int, loserPlayerID string, loserNickname string) ([]byte, error) {
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

