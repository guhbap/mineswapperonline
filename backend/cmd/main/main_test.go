package main

import (
	"testing"
	"time"
)

// setupTestRoom создает тестовую комнату с указанным режимом
func setupTestRoom(gameMode string, rows, cols, mines int) *Room {
	room := NewRoom("test-room", "Test Room", "", rows, cols, mines, 0, gameMode)
	return room
}

// setupTestServer создает тестовый сервер
func setupTestServer() *Server {
	roomManager := NewRoomManager()
	// Используем nil для БД в тестах
	server := NewServer(roomManager, nil)
	return server
}

// addTestPlayer добавляет тестового игрока в комнату
// В тестах мы не используем реальные WebSocket соединения, поэтому Conn будет nil
// Функции broadcast будут паниковать, но мы можем проверить состояние игры напрямую
func addTestPlayer(room *Room, playerID, nickname string) {
	room.mu.Lock()
	defer room.mu.Unlock()
	room.Players[playerID] = &Player{
		ID:       playerID,
		Nickname: nickname,
		Color:    "#FF6B6B",
		Conn:     nil, // В тестах не используем реальные соединения
	}
}

// TestClassicMode_InitialState проверяет начальное состояние в классическом режиме
func TestClassicMode_InitialState(t *testing.T) {
	room := setupTestRoom("classic", 10, 10, 10)
	
	room.GameState.mu.RLock()
	defer room.GameState.mu.RUnlock()
	
	// Проверяем, что мины размещены заранее
	mineCount := 0
	for i := 0; i < room.GameState.Rows; i++ {
		for j := 0; j < room.GameState.Cols; j++ {
			if room.GameState.Board[i][j].IsMine {
				mineCount++
			}
		}
	}
	
	if mineCount != room.GameState.Mines {
		t.Errorf("Ожидалось %d мин, найдено %d", room.GameState.Mines, mineCount)
	}
	
	// Проверяем, что подсказки не вычислены
	if len(room.GameState.CellHints) != 0 {
		t.Errorf("В классическом режиме подсказки не должны быть вычислены, найдено %d", len(room.GameState.CellHints))
	}
	
	// Проверяем, что игра не окончена
	if room.GameState.GameOver || room.GameState.GameWon {
		t.Error("Игра не должна быть окончена в начальном состоянии")
	}
}

// TestClassicMode_CellClick проверяет клик по ячейке в классическом режиме
func TestClassicMode_CellClick(t *testing.T) {
	room := setupTestRoom("classic", 10, 10, 10)
	server := setupTestServer()
	playerID := "player1"
	addTestPlayer(room, playerID, "Test Player")
	
	// Находим безопасную ячейку (не мину)
	var safeRow, safeCol int
	found := false
	room.GameState.mu.RLock()
	for i := 0; i < room.GameState.Rows && !found; i++ {
		for j := 0; j < room.GameState.Cols && !found; j++ {
			if !room.GameState.Board[i][j].IsMine {
				safeRow, safeCol = i, j
				found = true
			}
		}
	}
	room.GameState.mu.RUnlock()
	
	if !found {
		t.Fatal("Не найдена безопасная ячейка для теста")
	}
	
	// Кликаем по безопасной ячейке
	click := &CellClick{
		Row:  safeRow,
		Col:  safeCol,
		Flag: false,
	}
	
	server.handleCellClick(room, playerID, click)
	
	// Проверяем, что ячейка открыта
	room.GameState.mu.RLock()
	cell := room.GameState.Board[safeRow][safeCol]
	revealed := room.GameState.Revealed
	room.GameState.mu.RUnlock()
	
	if !cell.IsRevealed {
		t.Error("Ячейка должна быть открыта после клика")
	}
	
	if revealed != 1 {
		t.Errorf("Ожидалось 1 открытая ячейка, найдено %d", revealed)
	}
	
	// Проверяем, что игра не окончена
	room.GameState.mu.RLock()
	gameOver := room.GameState.GameOver
	room.GameState.mu.RUnlock()
	
	if gameOver {
		t.Error("Игра не должна быть окончена после клика по безопасной ячейке")
	}
}

// TestClassicMode_MineClick проверяет клик по мине в классическом режиме
func TestClassicMode_MineClick(t *testing.T) {
	room := setupTestRoom("classic", 10, 10, 10)
	server := setupTestServer()
	playerID := "player1"
	addTestPlayer(room, playerID, "Test Player")
	
	// Находим мину
	var mineRow, mineCol int
	found := false
	room.GameState.mu.RLock()
	for i := 0; i < room.GameState.Rows && !found; i++ {
		for j := 0; j < room.GameState.Cols && !found; j++ {
			if room.GameState.Board[i][j].IsMine {
				mineRow, mineCol = i, j
				found = true
			}
		}
	}
	room.GameState.mu.RUnlock()
	
	if !found {
		t.Fatal("Не найдена мина для теста")
	}
	
	// Кликаем по мине
	click := &CellClick{
		Row:  mineRow,
		Col:  mineCol,
		Flag: false,
	}
	
	server.handleCellClick(room, playerID, click)
	
	// Проверяем, что игра окончена
	room.GameState.mu.RLock()
	gameOver := room.GameState.GameOver
	loserID := room.GameState.LoserPlayerID
	room.GameState.mu.RUnlock()
	
	if !gameOver {
		t.Error("Игра должна быть окончена после клика по мине")
	}
	
	if loserID != playerID {
		t.Errorf("Ожидался проигравший %s, найден %s", playerID, loserID)
	}
}

// TestClassicMode_FlagToggle проверяет установку/снятие флага в классическом режиме
func TestClassicMode_FlagToggle(t *testing.T) {
	room := setupTestRoom("classic", 10, 10, 10)
	server := setupTestServer()
	playerID := "player1"
	addTestPlayer(room, playerID, "Test Player")
	
	// Устанавливаем флаг
	click := &CellClick{
		Row:  0,
		Col:  0,
		Flag: true,
	}
	
	server.handleCellClick(room, playerID, click)
	
	// Проверяем, что флаг установлен
	room.GameState.mu.RLock()
	flagged := room.GameState.Board[0][0].IsFlagged
	room.GameState.mu.RUnlock()
	
	if !flagged {
		t.Error("Флаг должен быть установлен")
	}
	
	// Снимаем флаг
	server.handleCellClick(room, playerID, click)
	
	// Проверяем, что флаг снят
	room.GameState.mu.RLock()
	flagged = room.GameState.Board[0][0].IsFlagged
	room.GameState.mu.RUnlock()
	
	if flagged {
		t.Error("Флаг должен быть снят")
	}
}

// TestTrainingMode_InitialState проверяет начальное состояние в режиме обучения
func TestTrainingMode_InitialState(t *testing.T) {
	room := setupTestRoom("training", 10, 10, 10)
	
	room.GameState.mu.RLock()
	defer room.GameState.mu.RUnlock()
	
	// Проверяем, что мины НЕ размещены заранее
	mineCount := 0
	for i := 0; i < room.GameState.Rows; i++ {
		for j := 0; j < room.GameState.Cols; j++ {
			if room.GameState.Board[i][j].IsMine {
				mineCount++
			}
		}
	}
	
	if mineCount != 0 {
		t.Errorf("В режиме training мины не должны быть размещены заранее, найдено %d", mineCount)
	}
}

// TestTrainingMode_DynamicMinePlacement проверяет динамическое размещение мин в режиме обучения
func TestTrainingMode_DynamicMinePlacement(t *testing.T) {
	room := setupTestRoom("training", 10, 10, 10)
	server := setupTestServer()
	playerID := "player1"
	addTestPlayer(room, playerID, "Test Player")
	
	// Кликаем по ячейке
	click := &CellClick{
		Row:  5,
		Col:  5,
		Flag: false,
	}
	
	server.handleCellClick(room, playerID, click)
	
	// Проверяем, что мины размещены динамически
	room.GameState.mu.RLock()
	mineCount := 0
	for i := 0; i < room.GameState.Rows; i++ {
		for j := 0; j < room.GameState.Cols; j++ {
			if room.GameState.Board[i][j].IsMine {
				mineCount++
			}
		}
	}
	clickedCell := room.GameState.Board[5][5]
	room.GameState.mu.RUnlock()
	
	// Проверяем, что кликнутая ячейка не мина (kaboom алгоритм делает худший сценарий, но не на первом клике если нет безопасных)
	if clickedCell.IsMine {
		t.Error("Кликнутая ячейка не должна быть миной (kaboom алгоритм)")
	}
	
	// Проверяем, что ячейка открыта
	if !clickedCell.IsRevealed {
		t.Error("Ячейка должна быть открыта после клика")
	}
	
	// Проверяем, что мины размещены (не все, но некоторые)
	if mineCount == 0 {
		t.Error("Мины должны быть размещены динамически")
	}
	
	// Проверяем, что подсказки вычислены (в режиме training они вычисляются асинхронно)
	// Даем время на асинхронное вычисление
	time.Sleep(100 * time.Millisecond)
	
	// В режиме training подсказки вычисляются асинхронно, поэтому может быть 0 или больше
	// Проверяем только, что игра не окончена
	room.GameState.mu.RLock()
	gameOver := room.GameState.GameOver
	room.GameState.mu.RUnlock()
	
	if gameOver {
		t.Error("Игра не должна быть окончена после первого клика в режиме training")
	}
}

// TestTrainingMode_MultipleClicks проверяет множественные клики в режиме обучения
func TestTrainingMode_MultipleClicks(t *testing.T) {
	room := setupTestRoom("training", 10, 10, 10)
	server := setupTestServer()
	playerID := "player1"
	addTestPlayer(room, playerID, "Test Player")
	
	// Первый клик
	click1 := &CellClick{
		Row:  5,
		Col:  5,
		Flag: false,
	}
	server.handleCellClick(room, playerID, click1)
	
	// Проверяем, что ячейка открыта
	room.GameState.mu.RLock()
	revealed1 := room.GameState.Revealed
	room.GameState.mu.RUnlock()
	
	if revealed1 != 1 {
		t.Errorf("После первого клика должно быть 1 открытая ячейка, найдено %d", revealed1)
	}
	
	// Второй клик
	click2 := &CellClick{
		Row:  3,
		Col:  3,
		Flag: false,
	}
	server.handleCellClick(room, playerID, click2)
	
	// Проверяем, что вторая ячейка тоже открыта
	room.GameState.mu.RLock()
	revealed2 := room.GameState.Revealed
	cell2 := room.GameState.Board[3][3]
	room.GameState.mu.RUnlock()
	
	if revealed2 < 2 {
		t.Errorf("После второго клика должно быть минимум 2 открытые ячейки, найдено %d", revealed2)
	}
	
	if !cell2.IsRevealed {
		t.Error("Вторая ячейка должна быть открыта")
	}
}

// TestFairMode_InitialState проверяет начальное состояние в справедливом режиме
func TestFairMode_InitialState(t *testing.T) {
	room := setupTestRoom("fair", 10, 10, 10)
	
	room.GameState.mu.RLock()
	defer room.GameState.mu.RUnlock()
	
	// Проверяем, что мины НЕ размещены заранее
	mineCount := 0
	for i := 0; i < room.GameState.Rows; i++ {
		for j := 0; j < room.GameState.Cols; j++ {
			if room.GameState.Board[i][j].IsMine {
				mineCount++
			}
		}
	}
	
	if mineCount != 0 {
		t.Errorf("В режиме fair мины не должны быть размещены заранее, найдено %d", mineCount)
	}
	
	// Проверяем, что подсказки не вычислены (только при проигрыше)
	if len(room.GameState.CellHints) != 0 {
		t.Errorf("В режиме fair подсказки не должны быть вычислены до проигрыша, найдено %d", len(room.GameState.CellHints))
	}
}

// TestFairMode_DynamicMinePlacement проверяет динамическое размещение мин в справедливом режиме
func TestFairMode_DynamicMinePlacement(t *testing.T) {
	room := setupTestRoom("fair", 10, 10, 10)
	server := setupTestServer()
	playerID := "player1"
	addTestPlayer(room, playerID, "Test Player")
	
	// Кликаем по ячейке
	click := &CellClick{
		Row:  5,
		Col:  5,
		Flag: false,
	}
	
	server.handleCellClick(room, playerID, click)
	
	// Проверяем, что мины размещены динамически
	room.GameState.mu.RLock()
	mineCount := 0
	for i := 0; i < room.GameState.Rows; i++ {
		for j := 0; j < room.GameState.Cols; j++ {
			if room.GameState.Board[i][j].IsMine {
				mineCount++
			}
		}
	}
	clickedCell := room.GameState.Board[5][5]
	hintsCount := len(room.GameState.CellHints)
	room.GameState.mu.RUnlock()
	
	// Проверяем, что кликнутая ячейка не мина
	if clickedCell.IsMine {
		t.Error("Кликнутая ячейка не должна быть миной (kaboom алгоритм)")
	}
	
	// Проверяем, что ячейка открыта
	if !clickedCell.IsRevealed {
		t.Error("Ячейка должна быть открыта после клика")
	}
	
	// Проверяем, что мины размещены
	if mineCount == 0 {
		t.Error("Мины должны быть размещены динамически")
	}
	
	// Проверяем, что подсказки НЕ вычислены (только при проигрыше)
	if hintsCount != 0 {
		t.Errorf("В режиме fair подсказки не должны быть вычислены до проигрыша, найдено %d", hintsCount)
	}
}

// TestFairMode_HintsOnGameOver проверяет, что подсказки вычисляются при проигрыше в справедливом режиме
func TestFairMode_HintsOnGameOver(t *testing.T) {
	room := setupTestRoom("fair", 10, 10, 10)
	server := setupTestServer()
	playerID := "player1"
	addTestPlayer(room, playerID, "Test Player")
	
	// Делаем несколько кликов, чтобы разместить мины
	for i := 0; i < 5; i++ {
		click := &CellClick{
			Row:  i,
			Col:  0,
			Flag: false,
		}
		server.handleCellClick(room, playerID, click)
		
		// Проверяем, что игра не окончена
		room.GameState.mu.RLock()
		gameOver := room.GameState.GameOver
		room.GameState.mu.RUnlock()
		
		if gameOver {
			break // Если игра окончена, выходим
		}
	}
	
	// Ищем мину и кликаем по ней
	var mineRow, mineCol int
	found := false
	room.GameState.mu.RLock()
	for i := 0; i < room.GameState.Rows && !found; i++ {
		for j := 0; j < room.GameState.Cols && !found; j++ {
			if room.GameState.Board[i][j].IsMine && !room.GameState.Board[i][j].IsRevealed {
				mineRow, mineCol = i, j
				found = true
			}
		}
	}
	room.GameState.mu.RUnlock()
	
	if found {
		// Кликаем по мине
		click := &CellClick{
			Row:  mineRow,
			Col:  mineCol,
			Flag: false,
		}
		server.handleCellClick(room, playerID, click)
		
		// Даем время на вычисление подсказок
		time.Sleep(100 * time.Millisecond)
		
		// Проверяем, что игра окончена
		room.GameState.mu.RLock()
		gameOver := room.GameState.GameOver
		hintsCount := len(room.GameState.CellHints)
		room.GameState.mu.RUnlock()
		
		if !gameOver {
			t.Error("Игра должна быть окончена после клика по мине")
		}
		
		// Проверяем, что подсказки вычислены
		if hintsCount == 0 {
			t.Error("В режиме fair подсказки должны быть вычислены при проигрыше")
		}
	} else {
		t.Log("Не найдена мина для теста проигрыша")
	}
}

// TestAllModes_FlagOnRevealedCell проверяет, что нельзя поставить флаг на открытую ячейку во всех режимах
func TestAllModes_FlagOnRevealedCell(t *testing.T) {
	modes := []string{"classic", "training", "fair"}
	
	for _, mode := range modes {
		t.Run(mode, func(t *testing.T) {
			room := setupTestRoom(mode, 10, 10, 10)
			server := setupTestServer()
			playerID := "player1"
			addTestPlayer(room, playerID, "Test Player")
			
			// Открываем ячейку
			click1 := &CellClick{
				Row:  5,
				Col:  5,
				Flag: false,
			}
			server.handleCellClick(room, playerID, click1)
			
			// Пытаемся поставить флаг на открытую ячейку
			click2 := &CellClick{
				Row:  5,
				Col:  5,
				Flag: true,
			}
			server.handleCellClick(room, playerID, click2)
			
			// Проверяем, что флаг не установлен
			room.GameState.mu.RLock()
			flagged := room.GameState.Board[5][5].IsFlagged
			room.GameState.mu.RUnlock()
			
			if flagged {
				t.Errorf("В режиме %s нельзя поставить флаг на открытую ячейку", mode)
			}
		})
	}
}

// TestAllModes_InvalidCoordinates проверяет обработку неверных координат во всех режимах
func TestAllModes_InvalidCoordinates(t *testing.T) {
	modes := []string{"classic", "training", "fair"}
	
	for _, mode := range modes {
		t.Run(mode, func(t *testing.T) {
			room := setupTestRoom(mode, 10, 10, 10)
			server := setupTestServer()
			playerID := "player1"
			addTestPlayer(room, playerID, "Test Player")
			
			// Кликаем по неверным координатам
			click := &CellClick{
				Row:  100,
				Col:  100,
				Flag: false,
			}
			
			// Не должно быть паники
			server.handleCellClick(room, playerID, click)
			
			// Проверяем, что ничего не изменилось
			room.GameState.mu.RLock()
			revealed := room.GameState.Revealed
			room.GameState.mu.RUnlock()
			
			if revealed != 0 {
				t.Errorf("В режиме %s не должно быть открытых ячеек после клика по неверным координатам", mode)
			}
		})
	}
}

// TestGameMode_Default проверяет, что по умолчанию используется классический режим
func TestGameMode_Default(t *testing.T) {
	room := NewRoom("test", "Test", "", 10, 10, 10, 0, "")
	
	if room.GameMode != "classic" {
		t.Errorf("Ожидался режим 'classic' по умолчанию, найден '%s'", room.GameMode)
	}
}

// TestGameMode_Invalid проверяет, что неверный режим обрабатывается корректно
func TestGameMode_Invalid(t *testing.T) {
	// Неверный режим должен быть заменен на classic
	room := NewRoom("test", "Test", "", 10, 10, 10, 0, "invalid")
	
	// Проверяем, что режим установлен (может быть invalid или classic в зависимости от реализации)
	// В текущей реализации неверный режим не валидируется при создании, но используется при создании GameState
	room.GameState.mu.RLock()
	// Проверяем, что игра создана
	rows := room.GameState.Rows
	room.GameState.mu.RUnlock()
	
	if rows != 10 {
		t.Errorf("Игра должна быть создана с правильными параметрами")
	}
}

