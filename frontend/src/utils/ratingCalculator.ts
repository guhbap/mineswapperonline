// Утилита для расчета рейтинга
// Формулы из rating.md:
// d = W * H * (M / (W * H)) ^ α, где α ≈ 1.5
// R = K * d / ln(t + 1), где K = 100, t - время в секундах
// Рейтинг пользователя - максимальное достигнутое значение за все его игры

const α = 1.5
const K = 100

// Вычисляет сложность поля по формуле: d = W * H * (M / (W * H)) ^ α
// где M - количество мин, W - ширина (cols), H - высота (rows)
export function calculateDifficulty(width: number, height: number, mines: number): number {
  const cells = width * height
  if (cells <= 0) return 0
  const density = mines / cells
  const difficulty = cells * Math.pow(density, α)
  return difficulty
}

// Вычисляет рейтинг за игру по формуле: R = K * d / ln(t + 1)
// где K = 100, d - сложность поля, t - время в секундах
export function calculateGameRating(
  width: number,
  height: number,
  mines: number,
  gameTime: number
): number {
  const d = calculateDifficulty(width, height, mines)
  if (d <= 0) return 0
  const timeFactor = Math.log(gameTime + 1)
  if (timeFactor <= 0) return 0
  const rating = K * d / timeFactor
  return rating
}

// Проверяет, может ли игра дать рейтинг
// Возвращает true, если:
// 1. Время игры >= 3 секунд
// 2. Плотность мин >= 10% (0.1)
export function isRatingEligible(
  width: number,
  height: number,
  mines: number,
  gameTime: number
): boolean {
  // Минимальное время - 3 секунды
  if (gameTime < 3.0) {
    return false
  }
  // Минимальная плотность мин - 10%
  const cells = width * height
  if (cells <= 0) {
    return false
  }
  const density = mines / cells
  if (density < 0.1) {
    return false
  }
  return true
}

// Вычисляет максимальный возможный рейтинг для данного поля (при минимальном времени 3 сек)
export function calculateMaxRating(
  width: number,
  height: number,
  mines: number,
): number {
  return calculateGameRating(width, height, mines, 3.0)
}

// Вычисляет изменение рейтинга на основе игры
// Возвращает новый рейтинг (максимум текущего и рейтинга за игру)
export function calculateRatingChange(
  width: number,
  height: number,
  mines: number,
  gameTime: number,
  currentRating: number = 0.0
): { gameRating: number; newRating: number } {
  const gameRating = calculateGameRating(width, height, mines, gameTime)
  // Рейтинг пользователя - максимальное достигнутое значение
  const newRating = Math.max(currentRating, gameRating)
  return { gameRating, newRating }
}

