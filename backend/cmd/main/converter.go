package main

import (
	"fmt"
	"log"

	"minesweeperonline/internal/game"
	pb "minesweeperonline/proto"
	"google.golang.org/protobuf/proto"
)

// convertGameStateToMain конвертирует game.GameState в GameState для protobuf
func convertGameStateToMain(gs *game.GameState) *GameState {
	log.Printf("convertGameStateToMain: пытаемся заблокировать GameState.Mu (RLock)")
	gs.Mu.RLock()
	log.Printf("convertGameStateToMain: GameState.Mu заблокирован (RLock), начинаем конвертацию")
	defer func() {
		gs.Mu.RUnlock()
		log.Printf("convertGameStateToMain: GameState.Mu разблокирован")
	}()

	mainGS := &GameState{
		Rows:          gs.Rows,
		Cols:          gs.Cols,
		Mines:         gs.Mines,
		GameOver:      gs.GameOver,
		GameWon:       gs.GameWon,
		Revealed:      gs.Revealed,
		HintsUsed:     gs.HintsUsed,
		LoserPlayerID: gs.LoserPlayerID,
		LoserNickname: gs.LoserNickname,
		flagSetInfo:   make(map[int]FlagInfo),
	}

	// Конвертируем SafeCells
	mainGS.SafeCells = make([]SafeCell, len(gs.SafeCells))
	for i, sc := range gs.SafeCells {
		mainGS.SafeCells[i] = SafeCell{Row: sc.Row, Col: sc.Col}
	}

	// Конвертируем CellHints
	mainGS.CellHints = make([]CellHint, len(gs.CellHints))
	for i, ch := range gs.CellHints {
		mainGS.CellHints[i] = CellHint{Row: ch.Row, Col: ch.Col, Type: ch.Type}
	}

	// Конвертируем Board
	mainGS.Board = make([][]Cell, len(gs.Board))
	for i := range gs.Board {
		mainGS.Board[i] = make([]Cell, len(gs.Board[i]))
		for j := range gs.Board[i] {
			mainGS.Board[i][j] = Cell{
				IsMine:        gs.Board[i][j].IsMine,
				IsRevealed:    gs.Board[i][j].IsRevealed,
				IsFlagged:     gs.Board[i][j].IsFlagged,
				NeighborMines: gs.Board[i][j].NeighborMines,
				FlagColor:     gs.Board[i][j].FlagColor,
			}
		}
	}

	return mainGS
}

// getWSPlayer получает WebSocket Player по ID
func (s *Server) getWSPlayer(playerID string) *Player {
	s.wsPlayersMu.RLock()
	defer s.wsPlayersMu.RUnlock()
	return s.wsPlayers[playerID]
}

// removeWSPlayer удаляет WebSocket Player
func (s *Server) removeWSPlayer(playerID string) {
	s.wsPlayersMu.Lock()
	defer s.wsPlayersMu.Unlock()
	delete(s.wsPlayers, playerID)
}

// convertGameStateFromMain конвертирует main.GameState в game.GameState
func convertGameStateFromMain(mainGS *GameState) *game.GameState {
	gs := &game.GameState{
		Rows:          mainGS.Rows,
		Cols:          mainGS.Cols,
		Mines:         mainGS.Mines,
		GameOver:      mainGS.GameOver,
		GameWon:       mainGS.GameWon,
		Revealed:      mainGS.Revealed,
		HintsUsed:     mainGS.HintsUsed,
		LoserPlayerID: mainGS.LoserPlayerID,
		LoserNickname: mainGS.LoserNickname,
		FlagSetInfo:   make(map[int]game.FlagInfo),
	}

	// Конвертируем SafeCells
	gs.SafeCells = make([]game.SafeCell, len(mainGS.SafeCells))
	for i, sc := range mainGS.SafeCells {
		gs.SafeCells[i] = game.SafeCell{Row: sc.Row, Col: sc.Col}
	}

	// Конвертируем CellHints
	gs.CellHints = make([]game.CellHint, len(mainGS.CellHints))
	for i, ch := range mainGS.CellHints {
		gs.CellHints[i] = game.CellHint{Row: ch.Row, Col: ch.Col, Type: ch.Type}
	}

	// Конвертируем Board
	gs.Board = make([][]game.Cell, len(mainGS.Board))
	for i := range mainGS.Board {
		gs.Board[i] = make([]game.Cell, len(mainGS.Board[i]))
		for j := range mainGS.Board[i] {
			gs.Board[i][j] = game.Cell{
				IsMine:        mainGS.Board[i][j].IsMine,
				IsRevealed:    mainGS.Board[i][j].IsRevealed,
				IsFlagged:     mainGS.Board[i][j].IsFlagged,
				NeighborMines: mainGS.Board[i][j].NeighborMines,
				FlagColor:     mainGS.Board[i][j].FlagColor,
			}
		}
	}

	return gs
}

// EncodeGameStateForPersistence кодирует game.GameState в protobuf формат для сохранения
func EncodeGameStateForPersistence(gs *game.GameState) ([]byte, error) {
	mainGS := convertGameStateToMain(gs)
	return encodeGameStateProtobuf(mainGS)
}

// DecodeGameStateFromPersistence декодирует protobuf данные в game.GameState для загрузки
func DecodeGameStateFromPersistence(data []byte) (*game.GameState, error) {
	return decodeGameStateFromProtobuf(data)
}

// decodeGameStateFromProtobuf декодирует protobuf данные в game.GameState
func decodeGameStateFromProtobuf(data []byte) (*game.GameState, error) {
	var wsMsg pb.WebSocketMessage
	if err := proto.Unmarshal(data, &wsMsg); err != nil {
		return nil, err
	}

	gameStateProto := wsMsg.GetGameState()
	if gameStateProto == nil {
		return nil, fmt.Errorf("GameState message is nil")
	}

	// Конвертируем protobuf GameState в main.GameState
	mainGS := &GameState{
		Rows:          int(gameStateProto.Rows),
		Cols:          int(gameStateProto.Cols),
		Mines:         int(gameStateProto.Mines),
		GameOver:      gameStateProto.GameOver,
		GameWon:       gameStateProto.GameWon,
		Revealed:      int(gameStateProto.Revealed),
		HintsUsed:     int(gameStateProto.HintsUsed),
		LoserPlayerID: gameStateProto.LoserPlayerId,
		LoserNickname: gameStateProto.LoserNickname,
		flagSetInfo:   make(map[int]FlagInfo),
	}

	// Конвертируем Board
	if board := gameStateProto.Board; board != nil {
		mainGS.Board = make([][]Cell, len(board.Rows))
		for i, row := range board.Rows {
			mainGS.Board[i] = make([]Cell, len(row.Cells))
			for j, cell := range row.Cells {
				mainGS.Board[i][j] = Cell{
					IsMine:        cell.IsMine,
					IsRevealed:    cell.IsRevealed,
					IsFlagged:     cell.IsFlagged,
					NeighborMines: int(cell.NeighborMines),
					FlagColor:     cell.FlagColor,
				}
			}
		}
	}

	// Конвертируем SafeCells
	mainGS.SafeCells = make([]SafeCell, len(gameStateProto.SafeCells))
	for i, sc := range gameStateProto.SafeCells {
		mainGS.SafeCells[i] = SafeCell{Row: int(sc.Row), Col: int(sc.Col)}
	}

	// Конвертируем CellHints
	mainGS.CellHints = make([]CellHint, len(gameStateProto.CellHints))
	for i, hint := range gameStateProto.CellHints {
		mainGS.CellHints[i] = CellHint{Row: int(hint.Row), Col: int(hint.Col), Type: hint.Type}
	}

	// Конвертируем в game.GameState
	return convertGameStateFromMain(mainGS), nil
}

