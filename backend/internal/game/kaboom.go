package game

import (
	"math/rand"
)

// SAT solver для проверки решаемости поля
type Sat struct {
	numVars int
	clauses [][]int
	// Кэш для быстрого доступа к unit clauses
	unitClauses map[int]bool // var -> value (true/false)
}

func NewSat(numVars int) *Sat {
	return &Sat{
		numVars:     numVars,
		clauses:     make([][]int, 0),
		unitClauses: make(map[int]bool),
	}
}

// Assert добавляет дизъюнкцию (clause)
func (s *Sat) Assert(vars []int) {
	if len(vars) == 1 {
		// Unit clause - сохраняем для быстрого доступа
		varIdx := abs(vars[0])
		if varIdx <= s.numVars {
			s.unitClauses[varIdx] = vars[0] > 0
		}
	}
	s.clauses = append(s.clauses, vars)
}

// AssertAtLeast добавляет ограничение: хотя бы k переменных из vars должны быть true
func (s *Sat) AssertAtLeast(vars []int, k int) {
	if k <= 0 {
		return
	}
	if k > len(vars) {
		// Невозможно - добавляем невыполнимое условие
		s.Assert([]int{1, -1}) // Противоречие
		return
	}
	size := len(vars) - k + 1
	for _, comb := range combinations(vars, size) {
		s.Assert(comb)
	}
}

// AssertAtMost добавляет ограничение: не более k переменных из vars могут быть true
func (s *Sat) AssertAtMost(vars []int, k int) {
	if k < 0 {
		s.Assert([]int{1, -1}) // Противоречие
		return
	}
	if k >= len(vars) {
		return // Всегда выполнимо
	}
	size := k + 1
	for _, comb := range combinations(vars, size) {
		negated := make([]int, len(comb))
		for i, v := range comb {
			negated[i] = -v
		}
		s.Assert(negated)
	}
}

// SolveWith решает SAT с дополнительным условием
// Оптимизировано: не копируем clauses, а используем стек
func (s *Sat) SolveWith(additional func()) *[]bool {
	// Сохраняем текущее состояние
	oldClausesLen := len(s.clauses)
	oldUnitClauses := make(map[int]bool)
	for k, v := range s.unitClauses {
		oldUnitClauses[k] = v
	}
	
	// Добавляем дополнительное условие
	if additional != nil {
		additional()
	}
	
	// Решаем
	result := s.Solve()
	
	// Восстанавливаем состояние (удаляем добавленные clauses)
	if len(s.clauses) > oldClausesLen {
		s.clauses = s.clauses[:oldClausesLen]
	}
	
	// Восстанавливаем unit clauses (удаляем только те, что были добавлены)
	s.unitClauses = oldUnitClauses
	
	return result
}

// Solve решает SAT задачу (простая реализация DPLL)
func (s *Sat) Solve() *[]bool {
	if len(s.clauses) == 0 {
		// Всегда выполнимо
		result := make([]bool, s.numVars+1)
		return &result
	}
	
	// Простая реализация DPLL
	assignment := make([]bool, s.numVars+1)
	used := make([]bool, s.numVars+1)
	
	return s.dpll(assignment, used, 0)
}

func (s *Sat) dpll(assignment []bool, used []bool, depth int) *[]bool {
	// Unit propagation: применяем unit clauses из кэша
	for varIdx, val := range s.unitClauses {
		if varIdx <= s.numVars && !used[varIdx] {
			assignment[varIdx] = val
			used[varIdx] = true
		}
	}
	
	// Проверяем unit clauses из clauses (на случай если они не в кэше)
	for {
		unitFound := false
		for _, clause := range s.clauses {
			if len(clause) == 1 {
				lit := clause[0]
				varIdx := abs(lit)
				if varIdx > s.numVars {
					continue
				}
				if !used[varIdx] {
					assignment[varIdx] = lit > 0
					used[varIdx] = true
					unitFound = true
					break
				}
			}
		}
		if !unitFound {
			break
		}
	}
	
	// Объединенная проверка: противоречия и удовлетворенность
	allSatisfied := true
	for _, clause := range s.clauses {
		satisfied := false
		hasUnassigned := false
		
		for _, lit := range clause {
			varIdx := abs(lit)
			if varIdx > s.numVars {
				continue
			}
			if used[varIdx] {
				val := assignment[varIdx]
				if (lit > 0 && val) || (lit < 0 && !val) {
					satisfied = true
					break
				}
			} else {
				hasUnassigned = true
			}
		}
		
		if !satisfied {
			if !hasUnassigned {
				// Противоречие: все литералы ложны
				return nil
			}
			allSatisfied = false
		}
	}
	
	// Ранний выход: если все клаузы выполнены
	if allSatisfied {
		return &assignment
	}
	
	// Выбираем неиспользованную переменную (оптимизация: выбираем из наиболее часто встречающихся)
	var nextVar int
	for i := 1; i <= s.numVars; i++ {
		if !used[i] {
			nextVar = i
			break
		}
	}
	
	if nextVar == 0 {
		// Все переменные использованы, но решение не найдено
		return nil
	}
	
	// Пробуем true
	assignment[nextVar] = true
	used[nextVar] = true
	if result := s.dpll(assignment, used, depth+1); result != nil {
		return result
	}
	
	// Пробуем false
	assignment[nextVar] = false
	if result := s.dpll(assignment, used, depth+1); result != nil {
		return result
	}
	
	used[nextVar] = false
	return nil
}

// AddCounter добавляет счетчик для переменных (для ограничений на количество мин)
func (s *Sat) AddCounter(vars []int) []int {
	if len(vars) <= 1 {
		return vars
	}
	
	mid := len(vars) / 2
	left := s.AddCounter(vars[:mid])
	right := s.AddCounter(vars[mid:])
	
	counter := make([]int, len(vars))
	for i := range counter {
		s.numVars++
		counter[i] = s.numVars
	}
	
	// Добавляем ограничения для счетчика
	for a := 0; a <= len(left); a++ {
		for b := 0; b <= len(right); b++ {
			if a > 0 && b > 0 {
				s.Assert([]int{-left[a-1], -right[b-1], counter[a+b-1]})
			} else if a > 0 {
				s.Assert([]int{-left[a-1], counter[a-1]})
			} else if b > 0 {
				s.Assert([]int{-right[b-1], counter[b-1]})
			}
			
			if a < len(left) && b < len(right) {
				s.Assert([]int{left[a], right[b], -counter[a+b]})
			} else if a < len(left) {
				s.Assert([]int{left[a], -counter[a+b]})
			} else if b < len(right) {
				s.Assert([]int{right[b], -counter[a+b]})
			}
		}
	}
	
	return counter
}

// AssertCounterAtLeast: хотя бы k переменных из счетчика должны быть true
func (s *Sat) AssertCounterAtLeast(counter []int, k int) {
	for i := 0; i < k && i < len(counter); i++ {
		s.Assert([]int{counter[i]})
	}
}

// AssertCounterAtMost: не более k переменных из счетчика могут быть true
func (s *Sat) AssertCounterAtMost(counter []int, k int) {
	for i := k; i < len(counter); i++ {
		s.Assert([]int{-counter[i]})
	}
}

// Вспомогательные функции
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func combinations(list []int, n int) [][]int {
	if n == 0 {
		return [][]int{{}}
	}
	if len(list) < n {
		return [][]int{}
	}
	if n == len(list) {
		return [][]int{list}
	}
	
	result := make([][]int, 0)
	
	// Рекурсивно генерируем комбинации
	var generate func(start int, current []int, remaining int)
	generate = func(start int, current []int, remaining int) {
		if remaining == 0 {
			result = append(result, append([]int{}, current...))
			return
		}
		
		for i := start; i <= len(list)-remaining; i++ {
			generate(i+1, append(current, list[i]), remaining-1)
		}
	}
	
	generate(0, []int{}, n)
	return result
}

// LabelMap представляет карту открытых ячеек и границы
type LabelMap struct {
	width       int
	height      int
	labels      [][]int // -1 означает закрытую ячейку, иначе число соседних мин
	boundary    []CellPos // Граница (неоткрытые ячейки рядом с открытыми)
	boundaryGrid [][]int  // Индекс в boundary для каждой ячейки, -1 если не на границе
	cache       [][]*bool // Кэш для тривиально решаемых мин (true=мина, false=безопасна, nil=неизвестно)
	numOutside  int      // Количество закрытых ячеек вне границы
}

type CellPos struct {
	Row int
	Col int
}

func NewLabelMap(width, height int) *LabelMap {
	lm := &LabelMap{
		width:        width,
		height:       height,
		labels:       make([][]int, height),
		boundary:     make([]CellPos, 0),
		boundaryGrid: make([][]int, height),
		cache:        make([][]*bool, height),
	}
	
	for i := 0; i < height; i++ {
		lm.labels[i] = make([]int, width)
		lm.boundaryGrid[i] = make([]int, width)
		lm.cache[i] = make([]*bool, width)
		for j := 0; j < width; j++ {
			lm.labels[i][j] = -1
			lm.boundaryGrid[i][j] = -1
		}
	}
	
	return lm
}

// SetLabel устанавливает метку для ячейки (число соседних мин)
func (lm *LabelMap) SetLabel(row, col, label int) {
	if row < 0 || row >= lm.height || col < 0 || col >= lm.width {
		return
	}
	lm.labels[row][col] = label
	lm.recalc()
}

// GetLabel возвращает метку ячейки
func (lm *LabelMap) GetLabel(row, col int) int {
	if row < 0 || row >= lm.height || col < 0 || col >= lm.width {
		return -1
	}
	return lm.labels[row][col]
}

// SetCache устанавливает кэш для ячейки на границе
func (lm *LabelMap) SetCache(idx int, val bool) {
	if idx < 0 || idx >= len(lm.boundary) {
		return
	}
	pos := lm.boundary[idx]
	valPtr := new(bool)
	*valPtr = val
	lm.cache[pos.Row][pos.Col] = valPtr
}

// GetCache возвращает кэш для ячейки на границе
func (lm *LabelMap) GetCache(idx int) *bool {
	if idx < 0 || idx >= len(lm.boundary) {
		return nil
	}
	pos := lm.boundary[idx]
	return lm.cache[pos.Row][pos.Col]
}

// ResetCache очищает кэш
func (lm *LabelMap) ResetCache() {
	for i := 0; i < lm.height; i++ {
		for j := 0; j < lm.width; j++ {
			lm.cache[i][j] = nil
		}
	}
}

// GetBoundaryIndex возвращает индекс в boundary для ячейки, или -1 если не на границе
func (lm *LabelMap) GetBoundaryIndex(row, col int) int {
	if row < 0 || row >= lm.height || col < 0 || col >= lm.width {
		return -1
	}
	return lm.boundaryGrid[row][col]
}

// GetBoundary возвращает список ячеек на границе
func (lm *LabelMap) GetBoundary() []CellPos {
	return lm.boundary
}

// Recalc пересчитывает границу и кэш
func (lm *LabelMap) Recalc() {
	lm.recalc()
}

func (lm *LabelMap) recalc() {
	// Очищаем границу
	lm.boundary = lm.boundary[:0]
	for i := 0; i < lm.height; i++ {
		for j := 0; j < lm.width; j++ {
			lm.boundaryGrid[i][j] = -1
		}
	}
	
	revealedSquares := 0
	
	// Собираем границу
	for i := 0; i < lm.height; i++ {
		for j := 0; j < lm.width; j++ {
			if lm.labels[i][j] != -1 {
				revealedSquares++
				
				// Собираем соседей
				neighboringBoundary := make([]int, 0)
				hasUncached := false
				
				for di := -1; di <= 1; di++ {
					for dj := -1; dj <= 1; dj++ {
						if di == 0 && dj == 0 {
							continue
						}
						ni, nj := i+di, j+dj
						if ni >= 0 && ni < lm.height && nj >= 0 && nj < lm.width {
							if lm.labels[ni][nj] == -1 {
								boundaryId := lm.boundaryGrid[ni][nj]
								if boundaryId == -1 {
									boundaryId = len(lm.boundary)
									lm.boundaryGrid[ni][nj] = boundaryId
									lm.boundary = append(lm.boundary, CellPos{Row: ni, Col: nj})
									hasUncached = true
								}
								neighboringBoundary = append(neighboringBoundary, boundaryId)
								
								// Проверяем, есть ли некешированные
								if lm.cache[ni][nj] == nil {
									hasUncached = true
								}
							}
						}
					}
				}
				
				// Тривиальное решение: если количество соседей на границе равно метке, все они мины
				if len(neighboringBoundary) == lm.labels[i][j] && hasUncached {
					for _, trivialMineId := range neighboringBoundary {
						lm.SetCache(trivialMineId, true)
					}
				}
			}
		}
	}
	
	lm.numOutside = (lm.width * lm.height) - revealedSquares - len(lm.boundary)
}

// Solver решает задачу определения безопасных ячеек
type Solver struct {
	map_          *LabelMap
	numMines      int
	minMines      int
	maxMines      int
	labels        []int
	labelToMine   [][]int
	cache         []*bool
	sat           *Sat
	canBeSafe     []bool
	canBeDangerous []bool
	uncachedMines []int
	numCachedTrue int
	counter       []int
}

func NewSolver(lm *LabelMap, numMines, minMines, maxMines int) *Solver {
	s := &Solver{
		map_:          lm,
		numMines:      numMines,
		minMines:      minMines,
		maxMines:      maxMines,
		labels:        make([]int, 0),
		labelToMine:   make([][]int, 0),
		cache:         make([]*bool, numMines),
		sat:           NewSat(numMines),
		canBeSafe:     make([]bool, numMines),
		canBeDangerous: make([]bool, numMines),
		uncachedMines: make([]int, 0),
	}
	
	// Инициализируем кэш
	for i := 0; i < numMines; i++ {
		cached := lm.GetCache(i)
		s.cache[i] = cached
		if cached == nil {
			s.uncachedMines = append(s.uncachedMines, i)
		} else if *cached {
			s.numCachedTrue++
		}
	}
	
	return s
}

// AddLabel добавляет ограничение на основе открытой ячейки
func (s *Solver) AddLabel(label int, mineList []int) {
	uncachedMineList := make([]int, 0)
	adjustedLabel := label
	
	for _, m := range mineList {
		if s.cache[m] == nil {
			uncachedMineList = append(uncachedMineList, m)
		} else if *s.cache[m] {
			adjustedLabel--
		}
	}
	
	s.labels = append(s.labels, adjustedLabel)
	s.labelToMine = append(s.labelToMine, uncachedMineList)
}

// Run запускает решатель
func (s *Solver) Run() {
	// Добавляем ограничения из меток
	for i := 0; i < len(s.labels); i++ {
		label := s.labels[i]
		vars := make([]int, len(s.labelToMine[i]))
		for j, m := range s.labelToMine[i] {
			vars[j] = m + 1 // SAT переменные начинаются с 1
		}
		
		if len(vars) > 0 {
			s.sat.AssertAtLeast(vars, label)
			s.sat.AssertAtMost(vars, label)
		}
	}
	
	// Добавляем ограничения из кэша
	for i := 0; i < s.numMines; i++ {
		if s.cache[i] != nil {
			if *s.cache[i] {
				s.sat.Assert([]int{i + 1})
			} else {
				s.sat.Assert([]int{-(i + 1)})
			}
		}
	}
	
	// Добавляем счетчик для некешированных мин
	if len(s.uncachedMines) > 0 {
		vars := make([]int, len(s.uncachedMines))
		for i, m := range s.uncachedMines {
			vars[i] = m + 1
		}
		s.counter = s.sat.AddCounter(vars)
		
		minRemaining := s.minMines - s.numCachedTrue
		if minRemaining > 0 {
			s.sat.AssertCounterAtLeast(s.counter, minRemaining)
		}
		
		maxRemaining := s.maxMines - s.numCachedTrue
		if maxRemaining >= 0 && maxRemaining < len(s.counter) {
			s.sat.AssertCounterAtMost(s.counter, maxRemaining)
		}
	}
	
	// Проверяем каждую ячейку на границе
	for i := 0; i < s.numMines; i++ {
		if s.cache[i] != nil {
			if *s.cache[i] {
				s.canBeSafe[i] = false
				s.canBeDangerous[i] = true
			} else {
				s.canBeSafe[i] = true
				s.canBeDangerous[i] = false
			}
			continue
		}
		
		// Проверяем, может ли быть безопасной
		solution := s.sat.SolveWith(func() {
			s.sat.Assert([]int{-(i + 1)})
		})
		if solution != nil {
			s.canBeSafe[i] = true
			s.update(*solution)
		} else {
			s.canBeSafe[i] = false
		}
		
		// Проверяем, может ли быть опасной
		solution = s.sat.SolveWith(func() {
			s.sat.Assert([]int{i + 1})
		})
		if solution != nil {
			s.canBeDangerous[i] = true
			s.update(*solution)
		} else {
			s.canBeDangerous[i] = false
		}
		
		// Обновляем кэш, если можем
		if s.canBeDangerous[i] && !s.canBeSafe[i] {
			val := true
			s.cache[i] = &val
			s.map_.SetCache(i, true)
		} else if s.canBeSafe[i] && !s.canBeDangerous[i] {
			val := false
			s.cache[i] = &val
			s.map_.SetCache(i, false)
		}
	}
}

func (s *Solver) update(solution []bool) {
	for i := 0; i < s.numMines; i++ {
		if i+1 < len(solution) {
			if solution[i+1] {
				s.canBeDangerous[i] = true
			} else {
				s.canBeSafe[i] = true
			}
		}
	}
}

// CanBeSafe проверяет, может ли ячейка быть безопасной
func (s *Solver) CanBeSafe(idx int) bool {
	if idx < 0 || idx >= s.numMines {
		return false
	}
	return s.canBeSafe[idx]
}

// CanBeDangerous проверяет, может ли ячейка быть опасной
func (s *Solver) CanBeDangerous(idx int) bool {
	if idx < 0 || idx >= s.numMines {
		return false
	}
	return s.canBeDangerous[idx]
}

// HasSafeCells проверяет, есть ли безопасные ячейки
func (s *Solver) HasSafeCells() bool {
	for i := 0; i < s.numMines; i++ {
		if !s.canBeDangerous[i] {
			return true
		}
	}
	return false
}

// HasNonDeadlyCells проверяет, есть ли не смертельные ячейки (могут быть безопасными)
func (s *Solver) HasNonDeadlyCells() bool {
	for i := 0; i < s.numMines; i++ {
		if s.canBeSafe[i] {
			return true
		}
	}
	return false
}

// OutsideIsSafe проверяет, безопасна ли область вне границы
func (s *Solver) OutsideIsSafe() bool {
	return s.numMines >= s.maxMines &&
		s.sat.SolveWith(func() {
			s.sat.AssertCounterAtMost(s.counter, s.maxMines-s.numCachedTrue-1)
		}) == nil
}

// OutsideCanBeSafe проверяет, может ли область вне границы быть безопасной
func (s *Solver) OutsideCanBeSafe() bool {
	if s.minMines < 0 {
		return true
	}
	return s.sat.SolveWith(func() {
		s.sat.AssertCounterAtLeast(s.counter, s.minMines-s.numCachedTrue+1)
	}) != nil
}

// AnySafeShape возвращает форму с безопасной ячейкой по индексу
func (s *Solver) AnySafeShape(idx int) *MineShape {
	solution := s.sat.SolveWith(func() {
		s.sat.Assert([]int{-(idx + 1)})
	})
	return s.shape(solution)
}

// AnyDangerousShape возвращает форму с опасной ячейкой по индексу
func (s *Solver) AnyDangerousShape(idx int) *MineShape {
	solution := s.sat.SolveWith(func() {
		s.sat.Assert([]int{idx + 1})
	})
	return s.shape(solution)
}

// AnyShapeWithOneEmpty возвращает форму с одной пустой ячейкой вне границы
func (s *Solver) AnyShapeWithOneEmpty() *MineShape {
	solution := s.sat.SolveWith(func() {
		s.sat.AssertCounterAtLeast(s.counter, s.minMines-s.numCachedTrue+1)
	})
	return s.shape(solution)
}

// AnyShapeWithRemaining возвращает форму с оставшимися минами
func (s *Solver) AnyShapeWithRemaining() *MineShape {
	solution := s.sat.SolveWith(func() {
		s.sat.AssertCounterAtMost(s.counter, s.maxMines-s.numCachedTrue-1)
	})
	return s.shape(solution)
}

// AnyShape возвращает любую форму (решение SAT)
func (s *Solver) AnyShape() *MineShape {
	solution := s.sat.Solve()
	return s.shape(solution)
}

// shape создает MineShape из решения SAT
func (s *Solver) shape(solution *[]bool) *MineShape {
	if solution == nil {
		return nil
	}
	mines := make([]bool, s.numMines)
	sum := 0
	for i := 0; i < s.numMines; i++ {
		if i+1 < len(*solution) {
			mines[i] = (*solution)[i+1]
			if mines[i] {
				sum++
			}
		}
	}
	return NewMineShape(s.map_, mines, s.maxMines-sum)
}

// MineShape представляет возможное расположение мин
type MineShape struct {
	map_      *LabelMap
	mines     []bool // Мины на границе
	remaining int    // Количество мин вне границы
}

func NewMineShape(lm *LabelMap, mines []bool, remaining int) *MineShape {
	return &MineShape{
		map_:      lm,
		mines:     mines,
		remaining: remaining,
	}
}

// MineGrid возвращает полную сетку мин
func (ms *MineShape) MineGrid() [][]bool {
	return ms.mineGridWithCell(-1, -1, false)
}

// MineGridWithMine возвращает сетку мин с миной в указанной позиции
func (ms *MineShape) MineGridWithMine(row, col int) [][]bool {
	return ms.mineGridWithCell(row, col, true)
}

// MineGridWithEmpty возвращает сетку мин с пустой ячейкой в указанной позиции
func (ms *MineShape) MineGridWithEmpty(row, col int) [][]bool {
	return ms.mineGridWithCell(row, col, false)
}

func (ms *MineShape) mineGridWithCell(exceptRow, exceptCol int, exceptIsMine bool) [][]bool {
	mineGrid := make([][]bool, ms.map_.height)
	for i := 0; i < ms.map_.height; i++ {
		mineGrid[i] = make([]bool, ms.map_.width)
	}

	// Размещаем мины на границе
	for i, pos := range ms.map_.boundary {
		if i < len(ms.mines) && ms.mines[i] {
			mineGrid[pos.Row][pos.Col] = true
		}
	}

	// Размещаем оставшиеся мины вне границы
	if ms.remaining > 0 {
		toSelect := make([]CellPos, 0)
		for i := 0; i < ms.map_.height; i++ {
			for j := 0; j < ms.map_.width; j++ {
				if ms.map_.labels[i][j] == -1 && ms.map_.boundaryGrid[i][j] == -1 {
					if !(i == exceptRow && j == exceptCol) {
						toSelect = append(toSelect, CellPos{Row: i, Col: j})
					}
				}
			}
		}

		// Перемешиваем
		rand.Shuffle(len(toSelect), func(i, j int) {
			toSelect[i], toSelect[j] = toSelect[j], toSelect[i]
		})

		remaining := ms.remaining
		if exceptRow >= 0 && exceptCol >= 0 && exceptIsMine {
			// Если исключаемая ячейка должна быть миной, уменьшаем remaining
			remaining--
			mineGrid[exceptRow][exceptCol] = true
		}

		for i := 0; i < remaining && i < len(toSelect); i++ {
			pos := toSelect[i]
			mineGrid[pos.Row][pos.Col] = true
		}
	}

	return mineGrid
}

// MakeSolver создает решатель для текущего состояния карты
func MakeSolver(lm *LabelMap, maxMines int) *Solver {
	minMines := maxMines - lm.numOutside
	if minMines < 0 {
		minMines = 0
	}
	
	solver := NewSolver(lm, len(lm.boundary), minMines, maxMines)
	
	// Добавляем ограничения из всех открытых ячеек
	for i := 0; i < lm.height; i++ {
		for j := 0; j < lm.width; j++ {
			label := lm.labels[i][j]
			if label == -1 {
				continue
			}
			
			mineList := make([]int, 0)
			for di := -1; di <= 1; di++ {
				for dj := -1; dj <= 1; dj++ {
					if di == 0 && dj == 0 {
						continue
					}
					ni, nj := i+di, j+dj
					if ni >= 0 && ni < lm.height && nj >= 0 && nj < lm.width {
						mineIdx := lm.boundaryGrid[ni][nj]
						if mineIdx != -1 {
							mineList = append(mineList, mineIdx)
						}
					}
				}
			}
			
			if len(mineList) > 0 {
				solver.AddLabel(label, mineList)
			}
		}
	}
	
	solver.Run()
	return solver
}

// CheckSolvability проверяет, решаемо ли поле (для генерации)
func CheckSolvability(board [][]bool, rows, cols, mines int) bool {
	// Создаем LabelMap с полностью закрытым полем
	lm := NewLabelMap(cols, rows)
	
	// Симулируем открытие первой ячейки (обычно безопасной)
	// Находим безопасную ячейку
	firstRow, firstCol := -1, -1
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			if !board[i][j] {
				// Подсчитываем соседние мины
				count := 0
				for di := -1; di <= 1; di++ {
					for dj := -1; dj <= 1; dj++ {
						if di == 0 && dj == 0 {
							continue
						}
						ni, nj := i+di, j+dj
						if ni >= 0 && ni < rows && nj >= 0 && nj < cols {
							if board[ni][nj] {
								count++
							}
						}
					}
				}
				firstRow, firstCol = i, j
				lm.SetLabel(i, j, count)
				break
			}
		}
		if firstRow != -1 {
			break
		}
	}
	
	if firstRow == -1 {
		return false // Нет безопасных ячеек
	}
	
	// Открываем все соседние пустые ячейки (flood fill)
	revealed := make([][]bool, rows)
	for i := range revealed {
		revealed[i] = make([]bool, cols)
	}
	
	var floodFill func(r, c int)
	floodFill = func(r, c int) {
		if r < 0 || r >= rows || c < 0 || c >= cols || revealed[r][c] {
			return
		}
		if board[r][c] {
			return
		}
		
		revealed[r][c] = true
		count := 0
		for di := -1; di <= 1; di++ {
			for dj := -1; dj <= 1; dj++ {
				if di == 0 && dj == 0 {
					continue
				}
				ni, nj := r+di, c+dj
				if ni >= 0 && ni < rows && nj >= 0 && nj < cols {
					if board[ni][nj] {
						count++
					}
				}
			}
		}
		lm.SetLabel(r, c, count)
		
		if count == 0 {
			for di := -1; di <= 1; di++ {
				for dj := -1; dj <= 1; dj++ {
					if di != 0 || dj != 0 {
						floodFill(r+di, c+dj)
					}
				}
			}
		}
	}
	
	floodFill(firstRow, firstCol)
	
	// Проверяем решаемость
	solver := MakeSolver(lm, mines)
	return solver.HasSafeCells() || len(lm.boundary) == 0
}

// GenerateSolvableBoard генерирует решаемое поле
func GenerateSolvableBoard(rows, cols, mines int, maxAttempts int) ([][]bool, bool) {
	for attempt := 0; attempt < maxAttempts; attempt++ {
		// Генерируем случайное поле
		board := make([][]bool, rows)
		for i := range board {
			board[i] = make([]bool, cols)
		}
		
		// Размещаем мины случайно
		positions := make([]int, rows*cols)
		for i := range positions {
			positions[i] = i
		}
		
		// Перемешиваем
		rand.Shuffle(len(positions), func(i, j int) {
			positions[i], positions[j] = positions[j], positions[i]
		})
		
		// Размещаем мины
		for i := 0; i < mines && i < len(positions); i++ {
			pos := positions[i]
			row := pos / cols
			col := pos % cols
			board[row][col] = true
		}
		
		// Проверяем решаемость
		if CheckSolvability(board, rows, cols, mines) {
			return board, true
		}
	}
	
	return nil, false
}

// CellInfo представляет информацию о ячейке для расчета безопасных ячеек
type CellInfo struct {
	IsRevealed    bool
	NeighborMines int
}

// CalculateSafeCells вычисляет безопасные ячейки для текущего состояния игры
func CalculateSafeCells(board [][]CellInfo, rows, cols, mines int) []CellPos {
	// Создаем LabelMap на основе открытых ячеек
	lm := NewLabelMap(cols, rows)
	
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			if board[i][j].IsRevealed {
				lm.SetLabel(i, j, board[i][j].NeighborMines)
			}
		}
	}
	
	// Создаем решатель
	solver := MakeSolver(lm, mines)
	
	// Собираем безопасные ячейки
	safeCells := make([]CellPos, 0)
	for i, pos := range lm.boundary {
		if !solver.CanBeDangerous(i) {
			safeCells = append(safeCells, pos)
		}
	}
	
	return safeCells
}

