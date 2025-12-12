package main

import (
	"fmt"
	"log"
	mathrand "math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	"minesweeperonline/internal/config"
	"minesweeperonline/internal/database"
	"minesweeperonline/internal/handlers"
	"minesweeperonline/internal/middleware"
	"minesweeperonline/internal/utils"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Разрешаем все источники для разработки
	},
}

type Player struct {
	ID                 string `json:"id"`
	UserID             int    `json:"userId,omitempty"` // ID пользователя из БД, если авторизован
	Nickname           string `json:"nickname"`
	Color              string `json:"color"`
	Conn               *websocket.Conn
	mu                 sync.Mutex
	LastCursorX        float64   // Последняя отправленная позиция курсора X
	LastCursorY        float64   // Последняя отправленная позиция курсора Y
	LastCursorSendTime time.Time // Время последней отправки курсора
}

type CursorPosition struct {
	PlayerID string  `json:"pid"` // playerId сокращено до pid
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
}

type FlagInfo struct {
	SetTime  time.Time
	PlayerID string
}

type SafeCell struct {
	Row int `json:"r"`
	Col int `json:"c"`
}

type GameState struct {
	Board         [][]Cell         `json:"b"`
	Rows          int              `json:"r"`
	Cols          int              `json:"c"`
	Mines         int              `json:"m"`
	GameOver      bool             `json:"go"`
	GameWon       bool             `json:"gw"`
	Revealed      int              `json:"rv"`
	HintsUsed     int              `json:"hu"`           // Количество использованных подсказок (глобально для комнаты)
	SafeCells     []SafeCell       `json:"sc,omitempty"` // Безопасные ячейки для режима без угадываний
	LoserPlayerID string           `json:"lpid,omitempty"`
	LoserNickname string           `json:"ln,omitempty"`
	flagSetInfo   map[int]FlagInfo // Информация об установке флага для каждой ячейки (ключ: row*cols + col)
	mu            sync.RWMutex
}

type Cell struct {
	IsMine        bool   `json:"m"`
	IsRevealed    bool   `json:"r"`
	IsFlagged     bool   `json:"f"`
	NeighborMines int    `json:"n"`
	FlagColor     string `json:"fc,omitempty"` // Цвет игрока, который поставил флаг
}

type Message struct {
	Type      string          `json:"type"`
	PlayerID  string          `json:"playerId,omitempty"`
	Nickname  string          `json:"nickname,omitempty"`
	Color     string          `json:"color,omitempty"`
	Cursor    *CursorPosition `json:"cursor,omitempty"`
	CellClick *CellClick      `json:"cellClick,omitempty"`
	Hint      *Hint           `json:"hint,omitempty"`
	GameState *GameState      `json:"gameState,omitempty"`
	Chat      *ChatMessage    `json:"chat,omitempty"`
}

type ChatMessage struct {
	Text     string `json:"text"`
	IsSystem bool   `json:"isSystem,omitempty"`
	Action   string `json:"action,omitempty"` // "flag", "reveal", "explode"
	Row      int    `json:"row,omitempty"`
	Col      int    `json:"col,omitempty"`
}

type CellClick struct {
	Row  int  `json:"row"`
	Col  int  `json:"col"`
	Flag bool `json:"flag"`
}

type Hint struct {
	Row int `json:"row"`
	Col int `json:"col"`
}

type Room struct {
	ID            string             `json:"id"`
	Name          string             `json:"name"`
	Password      string             `json:"-"`
	Rows          int                `json:"rows"`
	Cols          int                `json:"cols"`
	Mines         int                `json:"mines"`
	NoGuessing    bool               `json:"noGuessing"` // Режим без угадываний
	CreatorID     int                `json:"creatorId"`  // ID создателя комнаты (0 для гостей)
	Players       map[string]*Player `json:"-"`
	GameState     *GameState         `json:"-"`
	CreatedAt     time.Time          `json:"createdAt"`
	StartTime     *time.Time         `json:"-"` // Время начала игры (первый клик)
	deleteTimer   *time.Timer        // Таймер для отложенного удаления
	deleteTimerMu sync.Mutex         // Мьютекс для безопасной работы с таймером
	mu            sync.RWMutex
}

type RoomManager struct {
	rooms  map[string]*Room
	mu     sync.RWMutex
	server *Server // Ссылка на сервер для доступа к DeleteRoom
}

type Server struct {
	roomManager    *RoomManager
	db             *database.DB
	profileHandler *handlers.ProfileHandler
}

var colors = []string{
	"#FF6B6B", "#4ECDC4", "#45B7D1", "#FFA07A", "#98D8C8",
	"#F7DC6F", "#BB8FCE", "#85C1E2", "#F8B739", "#52BE80",
}

func NewRoomManager() *RoomManager {
	return &RoomManager{
		rooms: make(map[string]*Room),
	}
}

func (rm *RoomManager) SetServer(server *Server) {
	rm.server = server
}

func NewRoom(id, name, password string, rows, cols, mines int, creatorID int, noGuessing bool) *Room {
	return &Room{
		ID:         id,
		Name:       name,
		Password:   password,
		Rows:       rows,
		Cols:       cols,
		Mines:      mines,
		NoGuessing: noGuessing,
		CreatorID:  creatorID,
		Players:    make(map[string]*Player),
		GameState:  NewGameState(rows, cols, mines, noGuessing),
		CreatedAt:  time.Now(),
	}
}

func NewServer(roomManager *RoomManager, db *database.DB) *Server {
	server := &Server{
		roomManager:    roomManager,
		db:             db,
		profileHandler: handlers.NewProfileHandler(db),
	}
	// Устанавливаем ссылку на сервер в RoomManager для доступа к DeleteRoom
	roomManager.SetServer(server)
	return server
}

func NewGameState(rows, cols, mines int, noGuessing bool) *GameState {
	maxAttempts := 100 // Максимальное количество попыток генерации
	attempts := 0

	for attempts < maxAttempts {
		gs := &GameState{
			Rows:          rows,
			Cols:          cols,
			Mines:         mines,
			GameOver:      false,
			GameWon:       false,
			Revealed:      0,
			HintsUsed:     0,
			LoserPlayerID: "",
			LoserNickname: "",
			Board:         make([][]Cell, rows),
			flagSetInfo:   make(map[int]FlagInfo),
		}

		// Инициализация поля
		for i := range gs.Board {
			gs.Board[i] = make([]Cell, cols)
		}

		// Размещение мин
		mathrand.Seed(time.Now().UnixNano() + int64(attempts))
		minesPlaced := 0
		for minesPlaced < mines {
			row := mathrand.Intn(rows)
			col := mathrand.Intn(cols)
			if !gs.Board[row][col].IsMine {
				gs.Board[row][col].IsMine = true
				minesPlaced++
			}
		}

		// Подсчет соседних мин
		for i := 0; i < rows; i++ {
			for j := 0; j < cols; j++ {
				if !gs.Board[i][j].IsMine {
					count := 0
					for di := -1; di <= 1; di++ {
						for dj := -1; dj <= 1; dj++ {
							ni, nj := i+di, j+dj
							if ni >= 0 && ni < rows && nj >= 0 && nj < cols {
								if gs.Board[ni][nj].IsMine {
									count++
								}
							}
						}
					}
					gs.Board[i][j].NeighborMines = count
				}
			}
		}

		// Если режим без угадываний, проверяем решаемость поля
		if noGuessing {
			isSolvable, safeCells := isSolvableWithoutGuessing(gs)
			if isSolvable {
				gs.SafeCells = safeCells
				return gs
			}
			attempts++
			continue
		}

		// Если режим с угадываниями, возвращаем первое сгенерированное поле
		return gs
	}

	// Если не удалось сгенерировать поле без угадываний за maxAttempts попыток,
	// возвращаем последнее сгенерированное поле (или можно вернуть ошибку)
	// Для простоты возвращаем обычное поле
	return generateRandomBoard(rows, cols, mines)
}

// generateRandomBoard создает случайное поле (используется как fallback)
func generateRandomBoard(rows, cols, mines int) *GameState {
	gs := &GameState{
		Rows:          rows,
		Cols:          cols,
		Mines:         mines,
		GameOver:      false,
		GameWon:       false,
		Revealed:      0,
		HintsUsed:     0,
		LoserPlayerID: "",
		LoserNickname: "",
		Board:         make([][]Cell, rows),
		flagSetInfo:   make(map[int]FlagInfo),
	}

	for i := range gs.Board {
		gs.Board[i] = make([]Cell, cols)
	}

	mathrand.Seed(time.Now().UnixNano())
	minesPlaced := 0
	for minesPlaced < mines {
		row := mathrand.Intn(rows)
		col := mathrand.Intn(cols)
		if !gs.Board[row][col].IsMine {
			gs.Board[row][col].IsMine = true
			minesPlaced++
		}
	}

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			if !gs.Board[i][j].IsMine {
				count := 0
				for di := -1; di <= 1; di++ {
					for dj := -1; dj <= 1; dj++ {
						ni, nj := i+di, j+dj
						if ni >= 0 && ni < rows && nj >= 0 && nj < cols {
							if gs.Board[ni][nj].IsMine {
								count++
							}
						}
					}
				}
				gs.Board[i][j].NeighborMines = count
			}
		}
	}

	return gs
}

// Вспомогательная функция: получить соседей
func neighbors(rows, cols, i, j int) [][2]int {
	out := [][2]int{}
	for di := -1; di <= 1; di++ {
		for dj := -1; dj <= 1; dj++ {
			if di == 0 && dj == 0 {
				continue
			}
			ni, nj := i+di, j+dj
			if ni < 0 || ni >= rows || nj < 0 || nj >= cols {
				continue
			}
			out = append(out, [2]int{ni, nj})
		}
	}
	return out
}

// Главная функция
func isSolvableWithoutGuessing(gs *GameState) (bool, []SafeCell) {
	rows, cols := gs.Rows, gs.Cols

	// 1) Собираем входные видимые массивы
	revealed := make([][]bool, rows)
	flagged := make([][]bool, rows)
	totalRevealed := 0
	for i := 0; i < rows; i++ {
		revealed[i] = make([]bool, cols)
		flagged[i] = make([]bool, cols)
		for j := 0; j < cols; j++ {
			revealed[i][j] = gs.Board[i][j].IsRevealed
			flagged[i][j] = gs.Board[i][j].IsFlagged
			if revealed[i][j] {
				totalRevealed++
			}
		}
	}

	// Специальная обработка начального состояния: если все клетки закрыты,
	// возвращаем все клетки с нулевыми соседями как безопасные
	if totalRevealed == 0 {
		safeCells := []SafeCell{}
		for i := 0; i < rows; i++ {
			for j := 0; j < cols; j++ {
				// Клетка безопасна, если она не мина и не имеет соседних мин
				if !gs.Board[i][j].IsMine && gs.Board[i][j].NeighborMines == 0 {
					safeCells = append(safeCells, SafeCell{Row: i, Col: j})
				}
			}
		}
		// Если есть хотя бы одна безопасная клетка, возвращаем их
		if len(safeCells) > 0 {
			return true, safeCells
		}
		// Если нет безопасных клеток с нулевыми соседями, возвращаем все не-мины
		// (это гарантирует, что игрок сможет начать игру)
		allSafe := []SafeCell{}
		for i := 0; i < rows; i++ {
			for j := 0; j < cols; j++ {
				if !gs.Board[i][j].IsMine {
					allSafe = append(allSafe, SafeCell{Row: i, Col: j})
				}
			}
		}
		return len(allSafe) > 0, allSafe
	}

	// 2) Собираем фронтир: скрытые клетки, которые соседствуют с открытыми числами
	isHidden := make([][]bool, rows)
	for i := 0; i < rows; i++ {
		isHidden[i] = make([]bool, cols)
		for j := 0; j < cols; j++ {
			if !revealed[i][j] && !flagged[i][j] {
				isHidden[i][j] = true
			}
		}
	}

	// 3) для каждой открытой числовой клетки формируем ограничение:
	//    список соседних скрытых ячеек и требуемое число мин среди них = num - alreadyFlagged
	type Constraint struct {
		Cells [][2]int
		Need  int
	}
	constraints := []Constraint{}

	// Счётчик оставшихся мин глобально (для неконстрейнтных)
	// Рассчитаем сколько мин ещё не помечено флагом:
	totalFlagged := 0
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			if flagged[i][j] {
				totalFlagged++
			}
		}
	}
	minesRemaining := gs.Mines - totalFlagged
	if minesRemaining < 0 {
		minesRemaining = 0
	}

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			if !revealed[i][j] {
				continue
			}
			// только числовые клетки (>0) дают информацию; нули тоже (но их Need==0)
			num := gs.Board[i][j].NeighborMines
			adjHidden := [][2]int{}
			adjFlagged := 0
			for _, nb := range neighbors(rows, cols, i, j) {
				ni, nj := nb[0], nb[1]
				if flagged[ni][nj] {
					adjFlagged++
				} else if isHidden[ni][nj] {
					adjHidden = append(adjHidden, [2]int{ni, nj})
				}
			}
			need := num - adjFlagged
			if need < 0 {
				need = 0 // противоречие в состоянии — но пропустим
			}
			if len(adjHidden) > 0 {
				constraints = append(constraints, Constraint{Cells: adjHidden, Need: need})
			}
		}
	}

	// 4) Фронтир — уникальный набор скрытых клеток, которые входят в хоть одно ограничение
	frontierIndex := map[[2]int]int{} // map cell->index
	frontierCells := [][2]int{}
	for _, c := range constraints {
		for _, cell := range c.Cells {
			key := cell
			if _, ok := frontierIndex[key]; !ok {
				frontierIndex[key] = len(frontierCells)
				frontierCells = append(frontierCells, key)
			}
		}
	}

	// 5) Разбиваем фронтир на компоненты (связность через совместные ограничения)
	// Построим граф: ребро между двумя фронтирными клетками, если существуют constraint с ними обоими
	n := len(frontierCells)
	adj := make([][]int, n)
	for ci := range constraints {
		// для каждой пары клеток в ограничении — соединяем
		cells := constraints[ci].Cells
		for a := 0; a < len(cells); a++ {
			for b := a + 1; b < len(cells); b++ {
				ia := frontierIndex[cells[a]]
				ib := frontierIndex[cells[b]]
				adj[ia] = append(adj[ia], ib)
				adj[ib] = append(adj[ib], ia)
			}
		}
	}

	visited := make([]bool, n)
	components := [][]int{}
	for i := 0; i < n; i++ {
		if visited[i] {
			continue
		}
		// BFS/DFS
		stack := []int{i}
		visited[i] = true
		comp := []int{i}
		for len(stack) > 0 {
			v := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			for _, w := range adj[v] {
				if !visited[w] {
					visited[w] = true
					stack = append(stack, w)
					comp = append(comp, w)
				}
			}
		}
		components = append(components, comp)
	}

	// 6) Для каждой компоненты собираем локальные ограничения (только те, которые используют клетки из компоненты).
	// Затем перечисляем все возможные распределения мин по клеткам компоненты, которые удовлетворяют всем локальным ограничениям.
	// Собираем статистику: для каждой клетки — во всех решениях 0 или 1 или оба.
	safeMap := map[[2]int]bool{} // точно безопасна (всегда 0)
	mineMap := map[[2]int]bool{} // точно мина (всегда 1)
	foundAny := false

	// Ограничение на перебор: если компонент слишком большая, то перебор может быть
	// экспоненциальным. В реальной практике компоненты обычно небольшие.
	// Но мы всё равно ставим разумный предел (например 22 клеток) для избежания OOM.
	const maxComponentSize = 22

	for _, comp := range components {
		compSize := len(comp)
		if compSize == 0 {
			continue
		}
		if compSize > maxComponentSize {
			// слишком большая компонентa, не будем делать полный перебор — консервативно ничего не возвращаем из неё
			// (в "идеальной" реализации можно добавить SAT/BDD/ILP, но это сложнее)
			fmt.Printf("component size %d > %d: пропускаем полный перебор (консервативно)\n", compSize, maxComponentSize)
			continue
		}

		// Составим список ограничений, которые касаются этой компоненты.
		localConstraints := []Constraint{}
		for _, c := range constraints {
			// проверить, есть ли в c.Cells хоть одна клетка из comp
			contains := false
			for _, cell := range c.Cells {
				if _, ok := frontierIndex[cell]; ok {
					idx := frontierIndex[cell]
					// если idx в comp?
					inComp := false
					for _, v := range comp {
						if v == idx {
							inComp = true
							break
						}
					}
					if inComp {
						contains = true
						break
					}
				}
			}
			if contains {
				// отфильтруем клетки ограничения на те, что внутри компоненты, и учтём внешние как неизвестные (они НЕ должны быть, т.к. компонента построена по совместным ограничениям — но всё же)
				local := Constraint{Cells: make([][2]int, 0, len(c.Cells)), Need: c.Need}
				for _, cell := range c.Cells {
					if _, ok := frontierIndex[cell]; ok {
						idx := frontierIndex[cell]
						// если idx в comp — тогда добавляем, иначе оставляем (но в корректной разбивке таких не должно быть)
						inComp := false
						for _, v := range comp {
							if v == idx {
								inComp = true
								break
							}
						}
						if inComp {
							local.Cells = append(local.Cells, cell)
						} else {
							// клетка из другого компонента — это ошибка логики разбиения, но для безопасности — уменьшим Need на минимально возможное (ничего не делаем здесь).
							// На практике этого не случится.
						}
					}
				}
				localConstraints = append(localConstraints, local)
			}
		}

		// Индексация клеток компоненты: localIndex globalIndex -> cell
		localIndexToCell := make([][2]int, compSize)
		cellToLocalIndex := map[[2]int]int{}
		for li, gi := range comp {
			cell := frontierCells[gi]
			localIndexToCell[li] = cell
			cellToLocalIndex[cell] = li
		}

		// Преобразуем локальные ограничения: list of indices and need
		type LocC struct {
			Idxs []int
			Need int
		}
		locConstraints := []LocC{}
		for _, lc := range localConstraints {
			idxs := []int{}
			for _, cell := range lc.Cells {
				if li, ok := cellToLocalIndex[cell]; ok {
					idxs = append(idxs, li)
				}
			}
			// если idxs пуст (все клетки ограничения в других компонентах) — пропустим
			if len(idxs) == 0 {
				continue
			}
			locConstraints = append(locConstraints, LocC{Idxs: idxs, Need: lc.Need})
		}

		// Теперь полный перебор по 2^compSize с отсевами по локальным ограничениям.
		totalSolutions := 0
		alwaysZero := make([]bool, compSize)
		alwaysOne := make([]bool, compSize)
		for i := 0; i < compSize; i++ {
			alwaysZero[i] = true
			alwaysOne[i] = true
		}

		// рекурсивный backtrack с ранним отсевом:
		assign := make([]int, compSize) // 0 или 1
		var dfs func(pos int)
		dfs = func(pos int) {
			if pos == compSize {
				// проверим все локальные ограничения
				for _, lc := range locConstraints {
					sum := 0
					for _, idx := range lc.Idxs {
						sum += assign[idx]
					}
					if sum != lc.Need {
						return
					}
				}
				// решение валидно
				totalSolutions++
				for i := 0; i < compSize; i++ {
					if assign[i] == 0 {
						alwaysOne[i] = false
					} else {
						alwaysZero[i] = false
					}
				}
				return
			}

			// Пример раннего отсечения: для каждой ограничение, в которой уже участвует pos, можно проверить локальные bounds.
			// Но для простоты: пусть будет базовая версия — ставим и пробуем; компоненты ограничены maxComponentSize.

			// пробуем 0
			assign[pos] = 0
			// можно сделать локальные проверки ограничений, содержащих pos: если уже превышен верхний/нижний bound — отсекаем
			ok0 := true
			for _, lc := range locConstraints {
				need := lc.Need
				// считаем известную сумму и кол-во не назначенных в этой constraint
				sumKnown := 0
				unassigned := 0
				for _, idx := range lc.Idxs {
					if idx < pos {
						sumKnown += assign[idx]
					} else if idx == pos {
						sumKnown += 0
					} else {
						unassigned++
					}
				}
				// min possible = sumKnown
				// max possible = sumKnown + unassigned
				if need < sumKnown || need > sumKnown+unassigned {
					ok0 = false
					break
				}
			}
			if ok0 {
				dfs(pos + 1)
			}

			// пробуем 1
			assign[pos] = 1
			ok1 := true
			for _, lc := range locConstraints {
				need := lc.Need
				sumKnown := 0
				unassigned := 0
				for _, idx := range lc.Idxs {
					if idx < pos {
						sumKnown += assign[idx]
					} else if idx == pos {
						sumKnown += 1
					} else {
						unassigned++
					}
				}
				if need < sumKnown || need > sumKnown+unassigned {
					ok1 = false
					break
				}
			}
			if ok1 {
				dfs(pos + 1)
			}

			// очистим (необязательно)
			assign[pos] = 0
		}

		dfs(0)

		if totalSolutions == 0 {
			// Никаких решений нет — текущая конфигурация противоречива; пропускаем
			continue
		}

		// Проанализируем результаты
		for li := 0; li < compSize; li++ {
			cell := localIndexToCell[li]
			if alwaysZero[li] && !alwaysOne[li] {
				// всегда 0 => безопасна
				safeMap[cell] = true
				foundAny = true
			} else if alwaysOne[li] && !alwaysZero[li] {
				// всегда 1 => точно мина
				mineMap[cell] = true
				// пометка флага — это не "безопасный ход", но полезно
				foundAny = true
			}
		}
	}

	// 7) Обработка неконстрейнтных скрытых клеток (те, что не входят в frontier)
	// Если количество оставшихся нерасставленных мин ровно равно количеству неконстрейнтных скрытых => все они мины.
	// Если оставшихся мин = 0 => все они безопасны.
	nonFrontierHidden := [][2]int{}
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			if isHidden[i][j] {
				if _, ok := frontierIndex[[2]int{i, j}]; !ok {
					nonFrontierHidden = append(nonFrontierHidden, [2]int{i, j})
				}
			}
		}
	}
	// Сколько мин уже помечено локально как "точно мина" (mineMap) — мы не учитываем их в global check,
	// потому что они ещё не помечены в gs. Но для консистентности попробуем учесть:
	countKnownMines := 0
	for k := range mineMap {
		_ = k
		countKnownMines++
	}
	// реальное число мин, которые ещё могут лежать в неконстрейнтных клетках = minesRemaining - minesInFrontierPossibleMin
	// Но для точности лучше сделать консервативное: если minesRemaining == len(nonFrontierHidden) => все мины
	// Если minesRemaining == 0 => все безопасны
	if len(nonFrontierHidden) > 0 {
		if minesRemaining == len(nonFrontierHidden) {
			for _, c := range nonFrontierHidden {
				mineMap[c] = true
				foundAny = true
			}
		} else if minesRemaining == 0 {
			for _, c := range nonFrontierHidden {
				safeMap[c] = true
				foundAny = true
			}
		}
	}

	// 8) Собираем итоговый список безопасных клеток (без флагов)
	safeCells := []SafeCell{}
	for cell := range safeMap {
		// исключаем уже открытые или помеченные
		if revealed[cell[0]][cell[1]] || flagged[cell[0]][cell[1]] {
			continue
		}
		safeCells = append(safeCells, SafeCell{Row: cell[0], Col: cell[1]})
	}
	// Также можно вернуть детерминированные флаги как безопасные/важные подсказки — но вернём их отдельно через mineMap, если нужно.
	// Если найден хотя бы один детерминированный ход (безопасная клетка или точно мина) — считаем, что поле частично решаемо без угадываний
	if !foundAny || (len(safeCells) == 0 && len(mineMap) == 0) {
		return false, nil
	}

	// 9) Для удобства: расширяем все найденные нулевые ходы (BFS) — чтобы вернуть все реально раскрываемые клетки.
	// Симулируем раскрытие safeCells и всех нулей, чтобы вернуть пользователю кликабельные клетки.
	// Создадим копию revealedSim
	revealedSim := make([][]bool, rows)
	for i := 0; i < rows; i++ {
		revealedSim[i] = make([]bool, cols)
		for j := 0; j < cols; j++ {
			revealedSim[i][j] = revealed[i][j]
		}
	}
	queue := []struct{ r, c int }{}
	// добавим все safeCells в очередь (как клики)
	for _, sc := range safeCells {
		if !revealedSim[sc.Row][sc.Col] {
			revealedSim[sc.Row][sc.Col] = true
			queue = append(queue, struct{ r, c int }{sc.Row, sc.Col})
		}
	}
	// BFS: раскрываем нулевые области
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		i, j := cur.r, cur.c
		if gs.Board[i][j].NeighborMines == 0 {
			for _, nb := range neighbors(rows, cols, i, j) {
				ni, nj := nb[0], nb[1]
				if !revealedSim[ni][nj] && !flagged[ni][nj] {
					revealedSim[ni][nj] = true
					queue = append(queue, struct{ r, c int }{ni, nj})
				}
			}
		}
	}
	// Теперь соберём окончательный список безопасных ячеек, которые можно открыть сейчас (те, которые стали revealedSim==true, но ранее не были revealed)
	finalSafe := []SafeCell{}
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			if revealedSim[i][j] && !revealed[i][j] && !flagged[i][j] {
				finalSafe = append(finalSafe, SafeCell{Row: i, Col: j})
			}
		}
	}
	// Если нет новых раскрываемых ячеек, но есть джаст флаги (mineMap) — можем вернуть пустой список safe, но true (есть логические выводы).
	if len(finalSafe) == 0 && len(mineMap) > 0 {
		// Можно вернуть пустой список — но лучше вернуть nil? Вернём nil, но true — чтобы показать, что есть выводы (флаги).
		return true, nil
	}
	return true, finalSafe
}

func (rm *RoomManager) CreateRoom(name, password string, rows, cols, mines int, creatorID int, noGuessing bool) *Room {
	roomID := utils.GenerateID()
	room := NewRoom(roomID, name, password, rows, cols, mines, creatorID, noGuessing)
	rm.mu.Lock()
	rm.rooms[roomID] = room
	rm.mu.Unlock()
	log.Printf("Создана комната: %s (ID: %s, CreatorID: %d, NoGuessing: %v)", name, roomID, creatorID, noGuessing)
	return room
}

func (rm *RoomManager) UpdateRoom(roomID string, name, password string, rows, cols, mines int, noGuessing bool) error {
	rm.mu.RLock()
	room, exists := rm.rooms[roomID]
	rm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("room not found")
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	// Обновляем параметры комнаты
	room.Name = name
	if password == "__KEEP__" {
		// Не меняем пароль
	} else {
		// Устанавливаем новый пароль (может быть пустой строкой для удаления)
		room.Password = password
	}
	room.Rows = rows
	room.Cols = cols
	room.Mines = mines
	room.NoGuessing = noGuessing

	// Пересоздаем игровое поле с новыми параметрами
	room.GameState = NewGameState(rows, cols, mines, noGuessing)
	room.StartTime = nil // Сбрасываем время начала игры

	log.Printf("Комната обновлена: %s (ID: %s, NoGuessing: %v)", name, roomID, noGuessing)
	return nil
}

func (rm *RoomManager) GetRoom(roomID string) *Room {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.rooms[roomID]
}

func (rm *RoomManager) GetRoomsList() []map[string]interface{} {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	roomsList := make([]map[string]interface{}, 0, len(rm.rooms))
	for _, room := range rm.rooms {
		room.mu.RLock()
		playerCount := len(room.Players)
		room.mu.RUnlock()
		roomsList = append(roomsList, map[string]interface{}{
			"id":          room.ID,
			"name":        room.Name,
			"hasPassword": room.Password != "",
			"rows":        room.Rows,
			"cols":        room.Cols,
			"mines":       room.Mines,
			"noGuessing":  room.NoGuessing,
			"players":     playerCount,
			"createdAt":   room.CreatedAt,
			"creatorId":   room.CreatorID,
		})
	}
	return roomsList
}

func (rm *RoomManager) DeleteRoom(roomID string) {
	rm.mu.Lock()
	room, exists := rm.rooms[roomID]
	if exists {
		// Отменяем таймер удаления перед удалением комнаты
		room.CancelDeletion()
		delete(rm.rooms, roomID)
		log.Printf("Комната удалена: %s", roomID)
	}
	rm.mu.Unlock()
}

// ScheduleRoomDeletion планирует удаление комнаты через указанное время
func (rm *RoomManager) ScheduleRoomDeletion(roomID string, delay time.Duration) {
	rm.mu.RLock()
	room, exists := rm.rooms[roomID]
	rm.mu.RUnlock()

	if !exists {
		return
	}

	room.deleteTimerMu.Lock()
	defer room.deleteTimerMu.Unlock()

	// Отменяем предыдущий таймер, если он существует
	if room.deleteTimer != nil {
		room.deleteTimer.Stop()
	}

	// Создаем новый таймер
	room.deleteTimer = time.AfterFunc(delay, func() {
		// Проверяем, что комната все еще пустая перед удалением
		room.mu.RLock()
		playersCount := len(room.Players)
		room.mu.RUnlock()

		if playersCount == 0 {
			log.Printf("Комната %s пуста более %v, удаляем", roomID, delay)
			if rm.server != nil {
				rm.DeleteRoom(roomID)
			}
		} else {
			log.Printf("Комната %s больше не пуста (%d игроков), отмена удаления", roomID, playersCount)
		}
	})

	log.Printf("Запланировано удаление комнаты %s через %v", roomID, delay)
}

// CancelDeletion отменяет запланированное удаление комнаты
func (r *Room) CancelDeletion() {
	r.deleteTimerMu.Lock()
	defer r.deleteTimerMu.Unlock()

	if r.deleteTimer != nil {
		r.deleteTimer.Stop()
		r.deleteTimer = nil
		log.Printf("Отмена удаления комнаты %s", r.ID)
	}
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	roomID := r.URL.Query().Get("room")
	if roomID == "" {
		http.Error(w, "Room ID required", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Ошибка обновления соединения: %v", err)
		return
	}

	room := s.roomManager.GetRoom(roomID)
	if room == nil {
		conn.WriteJSON(map[string]string{"error": "Room not found"})
		conn.Close()
		return
	}

	// Отменяем удаление комнаты, если кто-то подключается
	room.CancelDeletion()

	playerID := utils.GenerateID()
	color := colors[mathrand.Intn(len(colors))]

	// Пытаемся получить userID из query параметра (если пользователь авторизован)
	userIDStr := r.URL.Query().Get("userId")
	var userID int
	if userIDStr != "" {
		// Парсим userID, игнорируем ошибку если не число
		if id, err := strconv.Atoi(userIDStr); err == nil {
			userID = id
			// Обновляем last_seen для пользователя
			if s.profileHandler != nil {
				s.profileHandler.UpdateLastSeen(userID)
				// Получаем сохраненный цвет пользователя, если есть
				if userColor, err := s.profileHandler.FindUserColor(userID); err == nil && userColor != "" {
					color = userColor
				}
			}
		}
	}

	player := &Player{
		ID:     playerID,
		UserID: userID,
		Color:  color,
		Conn:   conn,
	}

	room.mu.Lock()
	room.Players[playerID] = player
	room.mu.Unlock()

	log.Printf("Игрок %s подключен к комнате %s", playerID, roomID)

	// Настройка ping-pong для поддержания соединения
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// Запускаем горутину для отправки ping
	pingTicker := time.NewTicker(30 * time.Second)
	defer pingTicker.Stop()

	go func() {
		for range pingTicker.C {
			player.mu.Lock()
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("Ошибка отправки ping игроку %s: %v", playerID, err)
				player.mu.Unlock()
				return
			}
			player.mu.Unlock()
		}
	}()

	// Отправка начального состояния игры
	log.Printf("Отправка начального состояния игры игроку %s", playerID)
	s.sendGameStateToPlayer(room, player)
	log.Printf("Начальное состояние игры отправлено игроку %s", playerID)

	// Отправка списка игроков новому игроку
	s.sendPlayerListToPlayer(room, player)

	// Обработка сообщений
	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Ошибка чтения сообщения: %v", err)
			}
			break
		}

		// Обновляем deadline при получении сообщения
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))

		log.Printf("Получено сообщение от игрока %s: тип=%s", playerID, msg.Type)

		switch msg.Type {
		case "ping":
			// Отвечаем pong на ping сообщение
			pongMsg := Message{Type: "pong"}
			player.mu.Lock()
			if err := player.Conn.WriteJSON(pongMsg); err != nil {
				log.Printf("Ошибка отправки pong игроку %s: %v", playerID, err)
			}
			player.mu.Unlock()
			continue

		case "chat":
			if msg.Chat != nil {
				player.mu.Lock()
				msg.PlayerID = playerID
				msg.Nickname = player.Nickname
				msg.Color = player.Color
				player.mu.Unlock()
				// Отправляем сообщение всем игрокам в комнате
				s.broadcastToAll(room, msg)
			}
			continue

		case "nickname":
			player.mu.Lock()
			player.Nickname = msg.Nickname
			player.mu.Unlock()
			log.Printf("Никнейм игрока %s установлен: %s", playerID, msg.Nickname)
			s.broadcastPlayerList(room)

		case "cursor":
			if msg.Cursor != nil {
				player.mu.Lock()
				now := time.Now()
				// Throttling: отправляем не чаще чем раз в 100ms и только если позиция изменилась минимум на 5px
				timeSinceLastSend := now.Sub(player.LastCursorSendTime)
				dx := msg.Cursor.X - player.LastCursorX
				dy := msg.Cursor.Y - player.LastCursorY
				distance := dx*dx + dy*dy // квадрат расстояния для оптимизации

				// Отправляем только если прошло достаточно времени И позиция изменилась значительно
				if timeSinceLastSend < 100*time.Millisecond && distance < 25 { // 5px * 5px = 25
					player.mu.Unlock()
					continue // Пропускаем это сообщение
				}

				// Обрезаем playerID до 5 символов
				truncatedPlayerID := truncatePlayerID(playerID)
				msg.PlayerID = truncatedPlayerID
				msg.Cursor.PlayerID = truncatedPlayerID
				msg.Nickname = player.Nickname
				msg.Color = player.Color

				// Обновляем последнюю позицию и время
				player.LastCursorX = msg.Cursor.X
				player.LastCursorY = msg.Cursor.Y
				player.LastCursorSendTime = now
				player.mu.Unlock()

				s.broadcastToOthers(room, playerID, msg)
			}

		case "cellClick":
			if msg.CellClick != nil {
				log.Printf("Обработка клика: row=%d, col=%d, flag=%v", msg.CellClick.Row, msg.CellClick.Col, msg.CellClick.Flag)
				s.handleCellClick(room, playerID, msg.CellClick)
				log.Printf("Клик обработан, состояние игры обновлено")
			}

		case "hint":
			if msg.Hint != nil {
				log.Printf("Обработка подсказки: row=%d, col=%d", msg.Hint.Row, msg.Hint.Col)
				s.handleHint(room, playerID, msg.Hint)
			}

		case "newGame":
			room.mu.Lock()
			room.GameState = NewGameState(room.Rows, room.Cols, room.Mines, room.NoGuessing)
			room.StartTime = nil // Сбрасываем время начала игры
			room.mu.Unlock()
			log.Printf("Новая игра начата")
			s.broadcastGameState(room)
		}
	}

	// Отключение игрока
	room.mu.Lock()
	delete(room.Players, playerID)
	playersLeft := len(room.Players)
	room.mu.Unlock()

	s.broadcastPlayerList(room)
	conn.Close()
	log.Printf("Игрок отключен: %s, игроков в комнате: %d", playerID, playersLeft)

	// Планируем удаление комнаты через 5 минут, если она пустая
	if playersLeft == 0 {
		s.roomManager.ScheduleRoomDeletion(roomID, 5*time.Minute)
	}
}

func (s *Server) handleCellClick(room *Room, playerID string, click *CellClick) {
	room.GameState.mu.Lock()

	if room.GameState.GameOver || room.GameState.GameWon {
		log.Printf("Игра уже окончена, клик игнорируется")
		room.GameState.mu.Unlock()
		return
	}

	row, col := click.Row, click.Col
	if row < 0 || row >= room.GameState.Rows || col < 0 || col >= room.GameState.Cols {
		log.Printf("Некорректные координаты: row=%d, col=%d", row, col)
		room.GameState.mu.Unlock()
		return
	}

	cell := &room.GameState.Board[row][col]

	// Получаем информацию об игроке для сервисных сообщений
	room.mu.RLock()
	player := room.Players[playerID]
	var nickname string
	var playerColor string
	if player != nil {
		player.mu.Lock()
		nickname = player.Nickname
		playerColor = player.Color
		player.mu.Unlock()
	}
	room.mu.RUnlock()

	if click.Flag {
		// Переключение флага - нельзя ставить на открытые ячейки
		if cell.IsRevealed {
			log.Printf("Нельзя поставить флаг на открытую ячейку: row=%d, col=%d", row, col)
			room.GameState.mu.Unlock()
			return
		}

		wasFlagged := cell.IsFlagged
		cellKey := row*room.GameState.Cols + col
		now := time.Now()

		// Если пытаемся снять флаг, проверяем защиту от одновременных кликов
		if wasFlagged {
			if flagInfo, exists := room.GameState.flagSetInfo[cellKey]; exists {
				// Если это тот же игрок, который поставил флаг - разрешаем снять сразу
				if flagInfo.PlayerID != playerID {
					// Если это другой игрок - применяем защиту в 1 секунду
					timeSinceFlagSet := now.Sub(flagInfo.SetTime)
					if timeSinceFlagSet < 1*time.Second {
						log.Printf("Нельзя снять флаг сразу после установки другим игроком (защита от одновременных кликов): row=%d, col=%d, прошло %v", row, col, timeSinceFlagSet)
						room.GameState.mu.Unlock()
						return
					}
				}
			}
			// Удаляем информацию об установке при снятии флага
			delete(room.GameState.flagSetInfo, cellKey)
			cell.FlagColor = "" // Очищаем цвет при снятии флага
		} else {
			// Сохраняем время установки и playerID того, кто поставил флаг
			room.GameState.flagSetInfo[cellKey] = FlagInfo{
				SetTime:  now,
				PlayerID: playerID,
			}
			// Сохраняем цвет игрока, который поставил флаг
			cell.FlagColor = playerColor
		}

		cell.IsFlagged = !cell.IsFlagged
		log.Printf("Флаг переключен: row=%d, col=%d, flagged=%v", row, col, cell.IsFlagged)
		room.GameState.mu.Unlock()
		s.broadcastGameState(room)

		// Отправляем сервисное сообщение в чат
		if nickname != "" {
			action := "поставил флаг"
			if wasFlagged {
				action = "убрал флаг"
			}
			chatMsg := Message{
				Type:     "chat",
				PlayerID: playerID,
				Nickname: nickname,
				Color:    playerColor,
				Chat: &ChatMessage{
					Text:     fmt.Sprintf("%s %s на (%d, %d)", nickname, action, row+1, col+1),
					IsSystem: true,
					Action:   "flag",
					Row:      row,
					Col:      col,
				},
			}
			s.broadcastToAll(room, chatMsg)
		}
		return
	}

	// Открытие ячейки
	if cell.IsFlagged || cell.IsRevealed {
		log.Printf("Ячейка уже открыта или помечена флагом: row=%d, col=%d", row, col)
		room.GameState.mu.Unlock()
		return
	}

	// Проверка для режима без угадываний: можно открывать только безопасные ячейки
	room.mu.RLock()
	noGuessing := room.NoGuessing
	room.mu.RUnlock()

	if noGuessing && len(room.GameState.SafeCells) > 0 {
		// Проверяем, есть ли открытые ячейки (если нет - это первый клик)
		hasRevealedCells := room.GameState.Revealed > 0

		// Проверяем, является ли эта ячейка безопасной
		isSafe := false
		for _, safeCell := range room.GameState.SafeCells {
			if safeCell.Row == row && safeCell.Col == col {
				isSafe = true
				break
			}
		}

		// Если это не первый клик и ячейка не безопасна - блокируем
		if hasRevealedCells && !isSafe {
			log.Printf("В режиме без угадываний нельзя открыть непомеченную ячейку: row=%d, col=%d", row, col)
			room.GameState.mu.Unlock()
			return
		}
	}

	// Если это первое открытие, убеждаемся, что ячейка безопасна (0)
	isFirstClick := room.GameState.Revealed == 0
	if isFirstClick {
		// Если первая ячейка содержит мину или имеет соседние мины, перемещаем мины
		if cell.IsMine || cell.NeighborMines > 0 {
			log.Printf("Первое открытие небезопасно, перемещаем мины: row=%d, col=%d", row, col)
			s.ensureFirstClickSafe(room, row, col)
			// Обновляем ссылку на ячейку после перемещения мин
			cell = &room.GameState.Board[row][col]
		}
		// Устанавливаем время начала игры при первом клике
		room.mu.Lock()
		now := time.Now()
		room.StartTime = &now
		room.mu.Unlock()
	}

	cell.IsRevealed = true
	room.GameState.Revealed++
	log.Printf("Ячейка открыта: row=%d, col=%d, isMine=%v, neighborMines=%d, revealed=%d",
		row, col, cell.IsMine, cell.NeighborMines, room.GameState.Revealed)

	if cell.IsMine {
		room.GameState.GameOver = true
		// Сохраняем информацию об игроке, который проиграл
		if player != nil {
			player.mu.Lock()
			room.GameState.LoserPlayerID = playerID
			room.GameState.LoserNickname = player.Nickname
			userID := player.UserID
			player.mu.Unlock()

			// Вычисляем время игры
			room.mu.RLock()
			var gameTime float64
			if room.StartTime != nil {
				gameTime = time.Since(*room.StartTime).Seconds()
			} else {
				// Если StartTime не установлен (не должно происходить), используем 0
				gameTime = 0.0
			}
			room.mu.RUnlock()

			// Записываем поражение в БД (поражения не влияют на рейтинг)
			if userID > 0 && s.profileHandler != nil {
				// Собираем список участников игры
				participants := make([]handlers.GameParticipant, 0)
				room.mu.RLock()
				for _, p := range room.Players {
					p.mu.Lock()
					if p.UserID > 0 {
						participants = append(participants, handlers.GameParticipant{
							UserID:   p.UserID,
							Nickname: p.Nickname,
							Color:    p.Color,
						})
					}
					p.mu.Unlock()
				}
				room.mu.RUnlock()

				if err := s.profileHandler.RecordGameResult(userID, room.Cols, room.Rows, room.Mines, gameTime, false, participants); err != nil {
					log.Printf("Ошибка записи результата игры: %v", err)
				}
			}
		}
		log.Printf("Игра окончена - подорвалась мина! Игрок: %s (%s)", room.GameState.LoserNickname, playerID)

		// Отправляем сервисное сообщение о взрыве
		if nickname != "" {
			chatMsg := Message{
				Type:     "chat",
				PlayerID: playerID,
				Nickname: nickname,
				Color:    playerColor,
				Chat: &ChatMessage{
					Text:     fmt.Sprintf("%s подорвался на мине на (%d, %d) 💣", nickname, row+1, col+1),
					IsSystem: true,
					Action:   "explode",
					Row:      row,
					Col:      col,
				},
			}
			s.broadcastToAll(room, chatMsg)
		}
	} else {
		// Автоматическое открытие соседних пустых ячеек
		if cell.NeighborMines == 0 {
			log.Printf("Открытие соседних ячеек для row=%d, col=%d", row, col)
			s.revealNeighbors(room, row, col)
		}

		// Отправляем сервисное сообщение об открытии поля
		if nickname != "" {
			chatMsg := Message{
				Type:     "chat",
				PlayerID: playerID,
				Nickname: nickname,
				Color:    playerColor,
				Chat: &ChatMessage{
					Text:     fmt.Sprintf("%s открыл поле на (%d, %d)", nickname, row+1, col+1),
					IsSystem: true,
					Action:   "reveal",
					Row:      row,
					Col:      col,
				},
			}
			s.broadcastToAll(room, chatMsg)
		}

		// Проверка победы
		totalCells := room.GameState.Rows * room.GameState.Cols
		if room.GameState.Revealed == totalCells-room.GameState.Mines {
			room.GameState.GameWon = true
			log.Printf("Победа! Все ячейки открыты!")

			// Вычисляем время игры
			room.mu.RLock()
			var gameTime float64
			if room.StartTime != nil {
				gameTime = time.Since(*room.StartTime).Seconds()
			} else {
				// Если StartTime не установлен (не должно происходить), используем 0
				gameTime = 0.0
			}
			loserID := room.GameState.LoserPlayerID

			// Собираем список участников игры
			participants := make([]handlers.GameParticipant, 0)
			for _, p := range room.Players {
				p.mu.Lock()
				if p.UserID > 0 {
					participants = append(participants, handlers.GameParticipant{
						UserID:   p.UserID,
						Nickname: p.Nickname,
						Color:    p.Color,
					})
				}
				p.mu.Unlock()
			}

			// Записываем победу для всех игроков в комнате, которые не проиграли
			for _, p := range room.Players {
				// Записываем победу только для игроков, которые не проиграли
				if p.ID != loserID && p.UserID > 0 && s.profileHandler != nil {
					p.mu.Lock()
					if err := s.profileHandler.RecordGameResult(p.UserID, room.Cols, room.Rows, room.Mines, gameTime, true, participants); err != nil {
						log.Printf("Ошибка записи результата игры: %v", err)
					}
					p.mu.Unlock()
				}
			}
			room.mu.RUnlock()
		}
	}

	log.Printf("Отправка обновленного состояния игры после клика")
	room.GameState.mu.Unlock() // Разблокируем перед broadcastGameState
	s.broadcastGameState(room)
}

func (s *Server) ensureFirstClickSafe(room *Room, firstRow, firstCol int) {
	// Собираем все мины в радиусе 1 клетки от первой ячейки
	minesToMove := make([]struct{ row, col int }, 0)

	for di := -1; di <= 1; di++ {
		for dj := -1; dj <= 1; dj++ {
			ni, nj := firstRow+di, firstCol+dj
			if ni >= 0 && ni < room.GameState.Rows && nj >= 0 && nj < room.GameState.Cols {
				if room.GameState.Board[ni][nj].IsMine {
					minesToMove = append(minesToMove, struct{ row, col int }{ni, nj})
					room.GameState.Board[ni][nj].IsMine = false
				}
			}
		}
	}

	// Перемещаем мины в случайные свободные места
	mathrand.Seed(time.Now().UnixNano())
	for range minesToMove {
		// Ищем свободное место (не в радиусе 1 от первой ячейки и не занятое миной)
		attempts := 0
		for attempts < 100 {
			newRow := mathrand.Intn(room.GameState.Rows)
			newCol := mathrand.Intn(room.GameState.Cols)

			// Проверяем, что это не в радиусе 1 от первой ячейки
			tooClose := false
			for di := -1; di <= 1; di++ {
				for dj := -1; dj <= 1; dj++ {
					if newRow == firstRow+di && newCol == firstCol+dj {
						tooClose = true
						break
					}
				}
				if tooClose {
					break
				}
			}

			if !tooClose && !room.GameState.Board[newRow][newCol].IsMine {
				room.GameState.Board[newRow][newCol].IsMine = true
				break
			}
			attempts++
		}
	}

	// Пересчитываем соседние мины для всех ячеек
	for i := 0; i < room.GameState.Rows; i++ {
		for j := 0; j < room.GameState.Cols; j++ {
			if !room.GameState.Board[i][j].IsMine {
				count := 0
				for di := -1; di <= 1; di++ {
					for dj := -1; dj <= 1; dj++ {
						ni, nj := i+di, j+dj
						if ni >= 0 && ni < room.GameState.Rows && nj >= 0 && nj < room.GameState.Cols {
							if room.GameState.Board[ni][nj].IsMine {
								count++
							}
						}
					}
				}
				room.GameState.Board[i][j].NeighborMines = count
			}
		}
	}

	log.Printf("Мины перемещены, первая ячейка теперь безопасна")
}

func (s *Server) revealNeighbors(room *Room, row, col int) {
	for di := -1; di <= 1; di++ {
		for dj := -1; dj <= 1; dj++ {
			if di == 0 && dj == 0 {
				continue
			}
			ni, nj := row+di, col+dj
			if ni >= 0 && ni < room.GameState.Rows && nj >= 0 && nj < room.GameState.Cols {
				cell := &room.GameState.Board[ni][nj]
				if !cell.IsRevealed && !cell.IsFlagged && !cell.IsMine {
					cell.IsRevealed = true
					room.GameState.Revealed++
					if cell.NeighborMines == 0 {
						s.revealNeighbors(room, ni, nj)
					}
				}
			}
		}
	}
}

func (s *Server) handleHint(room *Room, playerID string, hint *Hint) {
	room.GameState.mu.Lock()

	if room.GameState.GameOver || room.GameState.GameWon {
		log.Printf("Игра уже окончена, подсказка игнорируется")
		room.GameState.mu.Unlock()
		return
	}

	// Проверяем лимит подсказок (3 подсказки глобально для комнаты)
	if room.GameState.HintsUsed >= 3 {
		log.Printf("Лимит подсказок исчерпан (использовано: %d)", room.GameState.HintsUsed)
		room.GameState.mu.Unlock()
		return
	}

	row, col := hint.Row, hint.Col
	if row < 0 || row >= room.GameState.Rows || col < 0 || col >= room.GameState.Cols {
		log.Printf("Некорректные координаты подсказки: row=%d, col=%d", row, col)
		room.GameState.mu.Unlock()
		return
	}

	cell := &room.GameState.Board[row][col]

	// Проверяем, что ячейка закрыта и не имеет флага
	if cell.IsRevealed || cell.IsFlagged {
		log.Printf("Ячейка уже открыта или помечена флагом: row=%d, col=%d", row, col)
		room.GameState.mu.Unlock()
		return
	}

	// Получаем информацию об игроке для сервисных сообщений
	room.mu.RLock()
	player := room.Players[playerID]
	var nickname string
	var playerColor string
	if player != nil {
		player.mu.Lock()
		nickname = player.Nickname
		playerColor = player.Color
		player.mu.Unlock()
	}
	room.mu.RUnlock()

	// Если там мина - ставим флаг, иначе открываем
	if cell.IsMine {
		// Ставим флаг
		cell.IsFlagged = true
		cell.FlagColor = playerColor
		room.GameState.HintsUsed++
		log.Printf("Подсказка: поставлен флаг на мине row=%d, col=%d (использовано подсказок: %d)", row, col, room.GameState.HintsUsed)
		room.GameState.mu.Unlock()
		s.broadcastGameState(room)

		// Отправляем сервисное сообщение в чат
		if nickname != "" {
			chatMsg := Message{
				Type:     "chat",
				PlayerID: playerID,
				Nickname: nickname,
				Color:    playerColor,
				Chat: &ChatMessage{
					Text:     fmt.Sprintf("%s использовал подсказку и поставил флаг на (%d, %d) 💡", nickname, row+1, col+1),
					IsSystem: true,
					Action:   "hint",
					Row:      row,
					Col:      col,
				},
			}
			s.broadcastToAll(room, chatMsg)
		}
	} else {
		// Открываем ячейку
		cell.IsRevealed = true
		room.GameState.Revealed++
		room.GameState.HintsUsed++
		log.Printf("Подсказка: открыта ячейка row=%d, col=%d, neighborMines=%d (использовано подсказок: %d)", row, col, cell.NeighborMines, room.GameState.HintsUsed)

		// Автоматическое открытие соседних пустых ячеек
		if cell.NeighborMines == 0 {
			log.Printf("Открытие соседних ячеек для row=%d, col=%d", row, col)
			s.revealNeighbors(room, row, col)
		}

		// Проверка победы
		totalCells := room.GameState.Rows * room.GameState.Cols
		if room.GameState.Revealed == totalCells-room.GameState.Mines {
			room.GameState.GameWon = true
			log.Printf("Победа! Все ячейки открыты!")

			// Вычисляем время игры
			room.mu.RLock()
			var gameTime float64
			if room.StartTime != nil {
				gameTime = time.Since(*room.StartTime).Seconds()
			} else {
				gameTime = 0.0
			}
			loserID := room.GameState.LoserPlayerID

			// Собираем список участников игры
			participants := make([]handlers.GameParticipant, 0)
			for _, p := range room.Players {
				p.mu.Lock()
				if p.UserID > 0 {
					participants = append(participants, handlers.GameParticipant{
						UserID:   p.UserID,
						Nickname: p.Nickname,
						Color:    p.Color,
					})
				}
				p.mu.Unlock()
			}

			// Записываем победу для всех игроков в комнате, которые не проиграли
			for _, p := range room.Players {
				if p.ID != loserID && p.UserID > 0 && s.profileHandler != nil {
					p.mu.Lock()
					if err := s.profileHandler.RecordGameResult(p.UserID, room.Cols, room.Rows, room.Mines, gameTime, true, participants); err != nil {
						log.Printf("Ошибка записи результата игры: %v", err)
					}
					p.mu.Unlock()
				}
			}
			room.mu.RUnlock()
		}

		room.GameState.mu.Unlock()
		s.broadcastGameState(room)

		// Отправляем сервисное сообщение в чат
		if nickname != "" {
			chatMsg := Message{
				Type:     "chat",
				PlayerID: playerID,
				Nickname: nickname,
				Color:    playerColor,
				Chat: &ChatMessage{
					Text:     fmt.Sprintf("%s использовал подсказку и открыл поле на (%d, %d) 💡", nickname, row+1, col+1),
					IsSystem: true,
					Action:   "hint",
					Row:      row,
					Col:      col,
				},
			}
			s.broadcastToAll(room, chatMsg)
		}
	}
}

func truncatePlayerID(playerID string) string {
	if len(playerID) > 5 {
		return playerID[:5]
	}
	return playerID
}

func (s *Server) sendGameStateToPlayer(room *Room, player *Player) {
	room.GameState.mu.RLock()
	loserPlayerID := truncatePlayerID(room.GameState.LoserPlayerID)
	gameStateCopy := GameState{
		Rows:          room.GameState.Rows,
		Cols:          room.GameState.Cols,
		Mines:         room.GameState.Mines,
		GameOver:      room.GameState.GameOver,
		GameWon:       room.GameState.GameWon,
		Revealed:      room.GameState.Revealed,
		HintsUsed:     room.GameState.HintsUsed,
		SafeCells:     room.GameState.SafeCells,
		LoserPlayerID: loserPlayerID,
		LoserNickname: room.GameState.LoserNickname,
	}
	boardCopy := make([][]Cell, len(room.GameState.Board))
	for i := range room.GameState.Board {
		boardCopy[i] = make([]Cell, len(room.GameState.Board[i]))
		copy(boardCopy[i], room.GameState.Board[i])
	}
	gameStateCopy.Board = boardCopy
	room.GameState.mu.RUnlock()

	player.mu.Lock()
	defer player.mu.Unlock()

	// Кодируем gameState в бинарный формат
	binaryData, err := encodeGameStateBinary(&gameStateCopy)
	if err != nil {
		log.Printf("Ошибка кодирования gameState: %v", err)
		return
	}

	// Отправляем бинарные данные с префиксом типа сообщения
	// Первый байт: тип сообщения (0 = gameState binary)
	message := append([]byte{0}, binaryData...)

	log.Printf("Отправка gameState (binary): Rows=%d, Cols=%d, Mines=%d, Revealed=%d, Size=%d bytes",
		gameStateCopy.Rows, gameStateCopy.Cols, gameStateCopy.Mines, gameStateCopy.Revealed, len(message))
	if err := player.Conn.WriteMessage(websocket.BinaryMessage, message); err != nil {
		log.Printf("Ошибка отправки состояния игры: %v", err)
	} else {
		log.Printf("Состояние игры успешно отправлено (binary)")
	}
}

func (s *Server) broadcastGameState(room *Room) {
	room.GameState.mu.RLock()
	loserPlayerID := truncatePlayerID(room.GameState.LoserPlayerID)
	gameStateCopy := GameState{
		Rows:          room.GameState.Rows,
		Cols:          room.GameState.Cols,
		Mines:         room.GameState.Mines,
		GameOver:      room.GameState.GameOver,
		GameWon:       room.GameState.GameWon,
		Revealed:      room.GameState.Revealed,
		HintsUsed:     room.GameState.HintsUsed,
		SafeCells:     room.GameState.SafeCells,
		LoserPlayerID: loserPlayerID,
		LoserNickname: room.GameState.LoserNickname,
	}
	boardCopy := make([][]Cell, len(room.GameState.Board))
	for i := range room.GameState.Board {
		boardCopy[i] = make([]Cell, len(room.GameState.Board[i]))
		copy(boardCopy[i], room.GameState.Board[i])
	}
	gameStateCopy.Board = boardCopy
	room.GameState.mu.RUnlock()

	// Кодируем gameState в бинарный формат
	binaryData, err := encodeGameStateBinary(&gameStateCopy)
	if err != nil {
		log.Printf("Ошибка кодирования gameState: %v", err)
		return
	}

	// Отправляем бинарные данные с префиксом типа сообщения
	message := append([]byte{0}, binaryData...)

	log.Printf("Broadcast gameState (binary): Rows=%d, Cols=%d, Revealed=%d, GameOver=%v, GameWon=%v, Size=%d bytes",
		gameStateCopy.Rows, gameStateCopy.Cols, gameStateCopy.Revealed, gameStateCopy.GameOver, gameStateCopy.GameWon, len(message))

	room.mu.RLock()
	playersCount := len(room.Players)
	room.mu.RUnlock()

	log.Printf("Отправка состояния игры %d игрокам", playersCount)

	room.mu.RLock()
	defer room.mu.RUnlock()
	for id, player := range room.Players {
		player.mu.Lock()
		if err := player.Conn.WriteMessage(websocket.BinaryMessage, message); err != nil {
			log.Printf("Ошибка отправки состояния игры игроку %s: %v", id, err)
		} else {
			log.Printf("Состояние игры отправлено игроку %s (binary)", id)
		}
		player.mu.Unlock()
	}
}

func (s *Server) broadcastToOthers(room *Room, senderID string, msg Message) {
	room.mu.RLock()
	playersCount := len(room.Players)
	room.mu.RUnlock()

	if playersCount <= 1 {
		log.Printf("Только один игрок, курсор не отправляется другим")
		return
	}

	room.mu.RLock()
	defer room.mu.RUnlock()
	sentCount := 0
	for id, player := range room.Players {
		if id != senderID {
			player.mu.Lock()
			if err := player.Conn.WriteJSON(msg); err != nil {
				log.Printf("Ошибка отправки сообщения игроку %s: %v", id, err)
			} else {
				sentCount++
			}
			player.mu.Unlock()
		}
	}
	log.Printf("Курсор отправлен %d игрокам (всего игроков: %d)", sentCount, playersCount)
}

func (s *Server) broadcastToAll(room *Room, msg Message) {
	room.mu.RLock()
	defer room.mu.RUnlock()
	for id, player := range room.Players {
		player.mu.Lock()
		if err := player.Conn.WriteJSON(msg); err != nil {
			log.Printf("Ошибка отправки сообщения чата игроку %s: %v", id, err)
		}
		player.mu.Unlock()
	}
}

func (s *Server) sendPlayerListToPlayer(room *Room, targetPlayer *Player) {
	room.mu.RLock()
	playersList := make([]map[string]string, 0, len(room.Players))
	for _, player := range room.Players {
		player.mu.Lock()
		playersList = append(playersList, map[string]string{
			"id":       player.ID,
			"nickname": player.Nickname,
			"color":    player.Color,
		})
		player.mu.Unlock()
	}
	room.mu.RUnlock()

	msgData := map[string]interface{}{
		"type":    "players",
		"players": playersList,
	}

	targetPlayer.mu.Lock()
	defer targetPlayer.mu.Unlock()
	if err := targetPlayer.Conn.WriteJSON(msgData); err != nil {
		log.Printf("Ошибка отправки списка игроков: %v", err)
	}
}

func (s *Server) broadcastPlayerList(room *Room) {
	room.mu.RLock()
	playersList := make([]map[string]string, 0, len(room.Players))
	for _, player := range room.Players {
		player.mu.Lock()
		playersList = append(playersList, map[string]string{
			"id":       player.ID,
			"nickname": player.Nickname,
			"color":    player.Color,
		})
		player.mu.Unlock()
	}
	room.mu.RUnlock()

	msgData := map[string]interface{}{
		"type":    "players",
		"players": playersList,
	}

	room.mu.RLock()
	defer room.mu.RUnlock()
	for _, player := range room.Players {
		player.mu.Lock()
		if err := player.Conn.WriteJSON(msgData); err != nil {
			log.Printf("Ошибка отправки списка игроков: %v", err)
		}
		player.mu.Unlock()
	}
}

func main() {
	// Загрузка конфигурации
	cfg, err := config.ReadConfig()
	if err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}

	// Подключение к базе данных
	db, err := database.NewDB(cfg.DbHost, cfg.DbPort, cfg.DbUser, cfg.DbPassword, cfg.DbName)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Инициализация схемы БД
	if cfg.NeedMigrate {
		if err := db.InitSchema(); err != nil {
			log.Fatalf("Failed to initialize database schema: %v", err)
		}
	}

	roomManager := NewRoomManager()
	server := NewServer(roomManager, db)
	authHandler := handlers.NewAuthHandler(db)
	profileHandler := handlers.NewProfileHandler(db)
	// roomHandler := handlers.NewRoomHandler(roomManager) // Используем старые обработчики для совместимости

	router := mux.NewRouter()

	r := router.PathPrefix("/api").Subrouter()
	// Публичные маршруты с опциональной авторизацией (для получения creatorID)
	r.Use(middleware.OptionalAuthMiddleware)
	r.HandleFunc("/auth/register", authHandler.Register).Methods("POST", "OPTIONS")
	r.HandleFunc("/auth/login", authHandler.Login).Methods("POST", "OPTIONS")
	r.HandleFunc("/ws", server.handleWebSocket)
	r.HandleFunc("/rooms", server.handleGetRooms).Methods("GET", "OPTIONS")
	r.HandleFunc("/rooms", server.handleCreateRoom).Methods("POST", "OPTIONS")
	r.HandleFunc("/rooms/join", server.handleJoinRoom).Methods("POST", "OPTIONS")

	// Защищенные маршруты
	protected := router.PathPrefix("/api").Subrouter()
	protected.Use(middleware.AuthMiddleware)
	protected.HandleFunc("/auth/me", authHandler.GetMe).Methods("GET", "OPTIONS")
	protected.HandleFunc("/profile", profileHandler.GetProfile).Methods("GET", "OPTIONS")
	protected.HandleFunc("/profile/activity", profileHandler.UpdateActivity).Methods("POST", "OPTIONS")
	protected.HandleFunc("/profile/color", profileHandler.UpdateColor).Methods("POST", "OPTIONS")
	protected.HandleFunc("/rooms/{id}", server.handleUpdateRoom).Methods("PUT", "OPTIONS")

	// Публичный маршрут для просмотра профиля по username
	r.HandleFunc("/profile", profileHandler.GetProfileByUsername).Methods("GET", "OPTIONS").Queries("username", "{username}")
	// Публичный маршрут для получения рейтинга
	r.HandleFunc("/leaderboard", profileHandler.GetLeaderboard).Methods("GET", "OPTIONS")
	// Публичный маршрут для получения топ-10 лучших игр по username (только с параметром username)
	r.HandleFunc("/profile/top-games", profileHandler.GetTopGames).Methods("GET", "OPTIONS").Queries("username", "{username}")
	// Защищенный маршрут для получения своих топ-10 игр (без параметра username)
	protected.HandleFunc("/profile/top-games", profileHandler.GetTopGames).Methods("GET", "OPTIONS")
	// Публичный маршрут для получения последних 10 игр по username (только с параметром username)
	r.HandleFunc("/profile/recent-games", profileHandler.GetRecentGames).Methods("GET", "OPTIONS").Queries("username", "{username}")
	// Защищенный маршрут для получения своих последних 10 игр (без параметра username)
	protected.HandleFunc("/profile/recent-games", profileHandler.GetRecentGames).Methods("GET", "OPTIONS")

	log.Printf("Сервер запущен на :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, middleware.CORSMiddleware(router)))
}

func (s *Server) handleCreateRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		utils.JSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		Name       string `json:"name"`
		Password   string `json:"password"`
		Rows       int    `json:"rows"`
		Cols       int    `json:"cols"`
		Mines      int    `json:"mines"`
		NoGuessing bool   `json:"noGuessing"`
	}

	if err := utils.DecodeJSON(r, &req); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := utils.ValidateRoomParams(req.Name, req.Rows, req.Cols, req.Mines); err != nil {
		utils.JSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Получаем creatorID из контекста (если пользователь авторизован)
	creatorID := 0
	if userID := r.Context().Value("userID"); userID != nil {
		if id, ok := userID.(int); ok {
			creatorID = id
		}
	}

	room := s.roomManager.CreateRoom(req.Name, req.Password, req.Rows, req.Cols, req.Mines, creatorID, req.NoGuessing)
	utils.JSONResponse(w, http.StatusOK, map[string]interface{}{
		"id":          room.ID,
		"name":        room.Name,
		"hasPassword": room.Password != "",
		"rows":        room.Rows,
		"cols":        room.Cols,
		"mines":       room.Mines,
		"noGuessing":  room.NoGuessing,
		"creatorId":   room.CreatorID,
	})
}

func (s *Server) handleGetRooms(w http.ResponseWriter, r *http.Request) {
	rooms := s.roomManager.GetRoomsList()
	utils.JSONResponse(w, http.StatusOK, rooms)
}

func (s *Server) handleJoinRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		utils.JSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		RoomID   string `json:"roomId"`
		Password string `json:"password"`
	}

	if err := utils.DecodeJSON(r, &req); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	room := s.roomManager.GetRoom(req.RoomID)
	if room == nil {
		utils.JSONError(w, http.StatusNotFound, "Room not found")
		return
	}

	if room.Password != "" && room.Password != req.Password {
		utils.JSONError(w, http.StatusUnauthorized, "Invalid password")
		return
	}

	room.mu.RLock()
	response := map[string]interface{}{
		"id":          room.ID,
		"name":        room.Name,
		"hasPassword": room.Password != "",
		"rows":        room.Rows,
		"cols":        room.Cols,
		"mines":       room.Mines,
		"noGuessing":  room.NoGuessing,
		"creatorId":   room.CreatorID,
	}
	room.mu.RUnlock()

	utils.JSONResponse(w, http.StatusOK, response)
}

func (s *Server) handleUpdateRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		utils.JSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Получаем userID из контекста (требуется авторизация)
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		utils.JSONError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Получаем roomID из URL
	vars := mux.Vars(r)
	roomID := vars["id"]
	if roomID == "" {
		utils.JSONError(w, http.StatusBadRequest, "Room ID required")
		return
	}

	// Используем map для проверки, было ли поле password передано
	var reqMap map[string]interface{}
	if err := utils.DecodeJSON(r, &reqMap); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Извлекаем значения из map
	name, _ := reqMap["name"].(string)
	rowsFloat, _ := reqMap["rows"].(float64)
	colsFloat, _ := reqMap["cols"].(float64)
	minesFloat, _ := reqMap["mines"].(float64)
	rows := int(rowsFloat)
	cols := int(colsFloat)
	mines := int(minesFloat)
	noGuessing := false
	if ng, exists := reqMap["noGuessing"]; exists {
		if ngBool, ok := ng.(bool); ok {
			noGuessing = ngBool
		}
	}

	// Проверяем, было ли передано поле password
	passwordProvided := false
	var password string
	if pwd, exists := reqMap["password"]; exists {
		passwordProvided = true
		if pwdStr, ok := pwd.(string); ok {
			password = pwdStr
		}
	}

	if err := utils.ValidateRoomParams(name, rows, cols, mines); err != nil {
		utils.JSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Проверяем, что комната существует и пользователь является создателем
	room := s.roomManager.GetRoom(roomID)
	if room == nil {
		utils.JSONError(w, http.StatusNotFound, "Room not found")
		return
	}

	room.mu.RLock()
	isCreator := room.CreatorID == userID
	room.mu.RUnlock()

	if !isCreator {
		utils.JSONError(w, http.StatusForbidden, "Only room creator can update room settings")
		return
	}

	// Обрабатываем пароль
	if !passwordProvided {
		// Если пароль не передан, сохраняем текущий пароль (используем специальное значение)
		room.mu.RLock()
		password = room.Password
		room.mu.RUnlock()
		password = "__KEEP__"
	}
	// Если passwordProvided == true, используем переданное значение (может быть пустой строкой для удаления)

	// Обновляем комнату
	if err := s.roomManager.UpdateRoom(roomID, name, password, rows, cols, mines, noGuessing); err != nil {
		utils.JSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Получаем обновленную комнату
	room = s.roomManager.GetRoom(roomID)
	room.mu.RLock()
	response := map[string]interface{}{
		"id":          room.ID,
		"name":        room.Name,
		"hasPassword": room.Password != "",
		"rows":        room.Rows,
		"cols":        room.Cols,
		"mines":       room.Mines,
		"noGuessing":  room.NoGuessing,
		"creatorId":   room.CreatorID,
	}
	room.mu.RUnlock()

	// Отправляем обновление всем игрокам в комнате через WebSocket
	room.mu.RLock()
	updateMsg := Message{
		Type: "roomUpdated",
		GameState: &GameState{
			Rows:  room.Rows,
			Cols:  room.Cols,
			Mines: room.Mines,
		},
	}
	room.mu.RUnlock()
	s.broadcastToAll(room, updateMsg)

	utils.JSONResponse(w, http.StatusOK, response)
}
