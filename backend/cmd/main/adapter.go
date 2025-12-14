package main

import (
	"minesweeperonline/internal/game"
	"minesweeperonline/internal/websocket"
)

// GameServiceAdapter адаптирует game.Service для использования в websocket.Manager
type GameServiceAdapter struct {
	service *game.Service
}

// NewGameServiceAdapter создает новый адаптер
func NewGameServiceAdapter(service *game.Service) *GameServiceAdapter {
	return &GameServiceAdapter{service: service}
}

// HandleCellClick обрабатывает клик по ячейке
func (a *GameServiceAdapter) HandleCellClick(room interface{}, playerID string, click *game.CellClick) error {
	gameRoom, ok := room.(*game.Room)
	if !ok {
		return nil
	}
	return a.service.HandleCellClick(gameRoom, playerID, click)
}

// HandleHint обрабатывает подсказку
func (a *GameServiceAdapter) HandleHint(room interface{}, playerID string, hint *game.Hint) error {
	gameRoom, ok := room.(*game.Room)
	if !ok {
		return nil
	}
	return a.service.HandleHint(gameRoom, playerID, hint)
}

// CalculateCellHints вычисляет подсказки
func (a *GameServiceAdapter) CalculateCellHints(room interface{}) {
	gameRoom, ok := room.(*game.Room)
	if !ok {
		return
	}
	a.service.CalculateCellHints(gameRoom)
}

// DetermineMinePlacement определяет размещение мин
func (a *GameServiceAdapter) DetermineMinePlacement(room interface{}, clickRow, clickCol int) [][]bool {
	gameRoom, ok := room.(*game.Room)
	if !ok {
		return nil
	}
	return a.service.DetermineMinePlacement(gameRoom, clickRow, clickCol)
}

// BroadcastGameState отправляет состояние игры всем игрокам
func (a *GameServiceAdapter) BroadcastGameState(room interface{}) {
	gameRoom, ok := room.(*game.Room)
	if !ok {
		return
	}
	a.service.BroadcastGameState(gameRoom)
}

// BroadcastCellUpdates отправляет обновления клеток
func (a *GameServiceAdapter) BroadcastCellUpdates(room interface{}, changedCells map[[2]int]bool, gameOver, gameWon bool, revealed, hintsUsed int, loserPlayerID, loserNickname string) {
	gameRoom, ok := room.(*game.Room)
	if !ok {
		return
	}
	a.service.BroadcastCellUpdates(gameRoom, changedCells, gameOver, gameWon, revealed, hintsUsed, loserPlayerID, loserNickname)
}

// BroadcastToAll отправляет сообщение всем игрокам
func (a *GameServiceAdapter) BroadcastToAll(room interface{}, msg game.Message) {
	gameRoom, ok := room.(*game.Room)
	if !ok {
		return
	}
	a.service.BroadcastToAll(gameRoom, msg)
}

// BroadcastToOthers отправляет сообщение всем игрокам кроме отправителя
func (a *GameServiceAdapter) BroadcastToOthers(room interface{}, senderID string, msg game.Message) {
	gameRoom, ok := room.(*game.Room)
	if !ok {
		return
	}
	a.service.BroadcastToOthers(gameRoom, senderID, msg)
}

// BroadcastPlayerList отправляет список игроков
func (a *GameServiceAdapter) BroadcastPlayerList(room interface{}) {
	gameRoom, ok := room.(*game.Room)
	if !ok {
		return
	}
	a.service.BroadcastPlayerList(gameRoom)
}

// SendGameStateToPlayer отправляет состояние игры конкретному игроку
func (a *GameServiceAdapter) SendGameStateToPlayer(room interface{}, player *websocket.Player) {
	gameRoom, ok := room.(*game.Room)
	if !ok {
		return
	}
	// Создаем адаптер для websocket.Player
	playerAdapter := &WSPlayerAdapter{player: player}
	a.service.SendGameStateToPlayer(gameRoom, playerAdapter)
}

// SendPlayerListToPlayer отправляет список игроков конкретному игроку
func (a *GameServiceAdapter) SendPlayerListToPlayer(room interface{}, player *websocket.Player) {
	gameRoom, ok := room.(*game.Room)
	if !ok {
		return
	}
	playerAdapter := &WSPlayerAdapter{player: player}
	a.service.SendPlayerListToPlayer(gameRoom, playerAdapter)
}

// WSPlayerAdapter адаптирует websocket.Player для использования в game.Service
type WSPlayerAdapter struct {
	player *websocket.Player
}

func (a *WSPlayerAdapter) GetNickname() string {
	return a.player.GetNickname()
}

func (a *WSPlayerAdapter) GetColor() string {
	return a.player.GetColor()
}

func (a *WSPlayerAdapter) GetUserID() int {
	return a.player.GetUserID()
}

func (a *WSPlayerAdapter) GetMu() interface{} {
	return &a.player.Mu
}

func (a *WSPlayerAdapter) GetConn() interface{} {
	return a.player.Conn
}

func (a *WSPlayerAdapter) SetNickname(nickname string) {
	a.player.SetNickname(nickname)
}

func (a *WSPlayerAdapter) UpdateCursor(x, y float64) bool {
	return a.player.UpdateCursor(x, y)
}

// WSManagerAdapter адаптирует websocket.Manager для использования в game.Service
type WSManagerAdapter struct {
	manager *websocket.Manager
}

func NewWSManagerAdapter(manager *websocket.Manager) *WSManagerAdapter {
	return &WSManagerAdapter{manager: manager}
}

// UpdateWSManager обновляет ссылку на wsManager (используется после создания финального wsManager)
func (a *WSManagerAdapter) UpdateWSManager(manager *websocket.Manager) {
	a.manager = manager
}

func (a *WSManagerAdapter) GetWSPlayer(playerID string) game.WSPlayer {
	player := a.manager.GetWSPlayer(playerID)
	if player == nil {
		return nil
	}
	return &WSPlayerAdapter{player: player}
}


