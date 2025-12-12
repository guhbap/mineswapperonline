// Утилита для расчета максимального рейтинга
// Использует те же формулы, что и backend/internal/rating/rating.go

const ALPHA = 0.7 // exponent for density in D
const BETA = 0.5 // exponent in expected time
const C = 4.0 // base time multiplier
const K = 32.0 // K-factor for player rating updates
const RREF = 1500.0 // rating for reference field (16x16,40)
const MIN_S = 0.01
const MAX_S = 0.99

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

