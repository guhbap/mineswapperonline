package game

// GameResultRecorder интерфейс для записи результатов игры
type GameResultRecorder interface {
	RecordGameResult(userID, cols, rows, mines int, gameTime float64, won bool, participants []GameParticipant) error
}

// GameParticipant представляет участника игры
type GameParticipant struct {
	UserID   int
	Nickname string
	Color    string
}

