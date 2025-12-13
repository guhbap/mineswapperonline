package main

import "minesweeperonline/internal/game"

// convertGameStateToMain конвертирует game.GameState в GameState для protobuf
func convertGameStateToMain(gs *game.GameState) *GameState {
	gs.Mu.RLock()
	defer gs.Mu.RUnlock()

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

