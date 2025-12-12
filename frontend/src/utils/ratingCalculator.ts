// Утилита для расчета максимального рейтинга
// Использует те же формулы, что и backend/internal/rating/rating.go

const ALPHA = 0.7 // exponent for density in D
const BETA = 0.5 // exponent in expected time
const C = 4.0 // base time multiplier
const K = 32.0 // K-factor for player rating updates
const RREF = 1500.0 // rating for reference field (16x16,40)
const MIN_S = 0.01
const MAX_S = 0.99
// MinComplexityRatio - минимальная сложность относительно референсного поля (25% от Dref)
// Поля с меньшей сложностью не дают рейтинг, чтобы предотвратить фарм на очень простых полях
const MIN_COMPLEXITY_RATIO = 0.25
// MinMines - минимальное количество мин для получения рейтинга
// Предотвращает фарм на очень больших полях с малым количеством мин
const MIN_MINES = 10
// MinDensity - минимальная плотность мин (mines/area) для получения рейтинга
// Референсное поле (16x16, 40) имеет плотность ~0.156 (15.6%)
// Минимум установлен в 5% (0.05) для предотвращения фарма на больших полях с низкой плотностью
const MIN_DENSITY = 0.05

// Вычисляет сложность D = A * (M/A)^alpha
function computeD(width: number, height: number, mines: number): number {
  const A = width * height
  if (A <= 0) return 0.0
  const density = mines / A
  return A * Math.pow(density, ALPHA)
}

// Вычисляет рейтинг головоломки на основе D, используя Dref для референсного поля
function computeRp(width: number, height: number, mines: number, Dref: number): number {
  const D = computeD(width, height, mines)
  if (D <= 0 || Dref <= 0) return RREF
  return RREF + 400.0 * Math.log10(D / Dref)
}

// Вычисляет ожидаемое время в секундах для поля
function expectedTime(width: number, height: number, mines: number): number {
  const A = width * height
  if (A <= 0) return 1.0
  const density = mines / A
  const den = Math.pow(density, BETA)
  if (den === 0) return C * Math.sqrt(A)
  return C * (Math.sqrt(A) / den)
}

// Вычисляет ожидаемый результат E по формуле Elo
function expectedResult(Rp: number, Rpl: number): number {
  return 1.0 / (1.0 + Math.pow(10.0, (Rp - Rpl) / 400.0))
}

// Вычисляет Dref (reference complexity для 16x16, 40 мин)
export function computeDref(): number {
  return computeD(16.0, 16.0, 40.0)
}

// Вычисляет максимальный прирост рейтинга для данного поля
// Максимальный прирост достигается при идеальном времени (S = MAX_S) и минимальном E
export function calculateMaxRatingGain(
  width: number,
  height: number,
  mines: number,
  currentRating: number = 1500.0
): number {
  const Dref = computeDref()
  const Rp = computeRp(width, height, mines, Dref)
  
  // Максимальный прирост при S = MAX_S и E минимально (когда Rp >> Rpl)
  // Но для более точного расчета используем реальный E
  const E = expectedResult(Rp, currentRating)
  
  // Максимальный прирост = K * (MAX_S - E)
  const maxDelta = K * (MAX_S - E)
  
  return maxDelta
}

// Вычисляет максимальный возможный рейтинг после победы
export function calculateMaxRating(
  width: number,
  height: number,
  mines: number,
  currentRating: number = 1500.0
): number {
  const maxGain = calculateMaxRatingGain(width, height, mines, currentRating)
  return currentRating + maxGain
}

// Проверяет, достаточно ли сложности поля для получения рейтинга
// Возвращает true, если:
// 1. Количество мин >= MinMines
// 2. Плотность мин (mines/area) >= MinDensity
// 3. Сложность поля >= MinComplexityRatio * Dref
export function isComplexitySufficient(
  width: number,
  height: number,
  mines: number
): boolean {
  const Dref = computeDref()
  if (Dref <= 0) return false
  // Проверка минимального количества мин
  if (mines < MIN_MINES) return false
  // Проверка минимальной плотности мин
  const A = width * height
  if (A <= 0) return false
  const density = mines / A
  if (density < MIN_DENSITY) return false
  const D = computeD(width, height, mines)
  const minComplexity = Dref * MIN_COMPLEXITY_RATIO
  return D >= minComplexity
}

// Вычисляет оценку производительности S в диапазоне (0.01..0.99) на основе реального времени T и ожидаемого времени Texp
function performanceScore(T: number, Texp: number): number {
  // linear mapping: S = 1 - (T - Texp) / (3*Texp)
  let s = 1.0 - (T - Texp) / (3.0 * Texp)
  if (s < MIN_S) {
    s = MIN_S
  }
  if (s > MAX_S) {
    s = MAX_S
  }
  return s
}

// Вычисляет изменение рейтинга на основе реального времени игры
// Возвращает изменение рейтинга (delta) и новый рейтинг
export function calculateRatingChange(
  width: number,
  height: number,
  mines: number,
  gameTime: number, // время игры в секундах
  currentRating: number = 1500.0
): { delta: number; newRating: number } {
  const Dref = computeDref()
  const Rp = computeRp(width, height, mines, Dref)
  const Texp = expectedTime(width, height, mines)
  const S = performanceScore(gameTime, Texp)
  const E = expectedResult(Rp, currentRating)
  const delta = K * (S - E)
  const newRating = currentRating + delta
  return { delta, newRating }
}

