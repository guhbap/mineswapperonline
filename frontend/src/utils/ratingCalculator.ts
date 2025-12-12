// Утилита для расчета рейтинга
// Упрощенная система: каждая игра добавляет к рейтингу столько очков, сколько сложность у нее

// Вычисляет сложность поля по формуле: (M / (W * H)) * sqrt(W^2 + H^2)
// где M - количество мин, W - ширина (cols), H - высота (rows)
export function calculateDifficulty(width: number, height: number, mines: number): number {
  const totalCells = width * height
  if (totalCells <= 0) return 0

  const density = mines / totalCells
  const diagonal = Math.sqrt(width * width + height * height)

  return density * diagonal
}

// Вычисляет максимальный прирост рейтинга для данного поля
// Просто возвращает сложность поля
export function calculateMaxRatingGain(
  width: number,
  height: number,
  mines: number,
): number {
  const difficulty = calculateDifficulty(width, height, mines)
  return difficulty
}

// Вычисляет максимальный возможный рейтинг после победы
export function calculateMaxRating(
  width: number,
  height: number,
  mines: number,
): number {
  const difficulty = calculateDifficulty(width, height, mines)
  return difficulty
}

// Вычисляет изменение рейтинга на основе сложности поля
// Просто возвращает сложность как изменение рейтинга
export function calculateRatingChange(
  width: number,
  height: number,
  mines: number,
  currentRating: number = 1500.0
): { delta: number; newRating: number } {
  const difficulty = calculateDifficulty(width, height, mines)
  const newRating = currentRating + difficulty
  return { delta: difficulty, newRating }
}

