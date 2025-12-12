import type { Ref } from 'vue'

export interface UseCellTouchOptions {
  clickThreshold?: number // Максимальное расстояние для клика (в пикселях)
  clickDuration?: number // Максимальная длительность для клика (в миллисекундах)
}

export interface UseCellTouchReturn {
  handleTouchStart: (event: TouchEvent) => void
  handleTouchEnd: (row: number, col: number, event: TouchEvent, onClick: (row: number, col: number) => void) => void
}

const DEFAULT_OPTIONS: Required<UseCellTouchOptions> = {
  clickThreshold: 10,
  clickDuration: 300,
}

/**
 * Composable для обработки touch-событий на ячейках игрового поля
 * Различает клики и панорамирование
 */
export function useCellTouch(options: UseCellTouchOptions = {}): UseCellTouchReturn {
  const {
    clickThreshold = DEFAULT_OPTIONS.clickThreshold,
    clickDuration = DEFAULT_OPTIONS.clickDuration,
  } = options

  // Состояние для отслеживания начала касания
  let touchStartTime = 0
  let touchStartPos = { x: 0, y: 0 }

  /**
   * Обработчик начала касания на ячейке
   */
  const handleTouchStart = (event: TouchEvent) => {
    if (event.touches.length === 1) {
      touchStartTime = Date.now()
      touchStartPos.x = event.touches[0].clientX
      touchStartPos.y = event.touches[0].clientY
    }
  }

  /**
   * Обработчик окончания касания на ячейке
   * Определяет, был ли это клик или панорамирование
   */
  const handleTouchEnd = (
    row: number,
    col: number,
    event: TouchEvent,
    onClick: (row: number, col: number) => void
  ) => {
    if (!event.changedTouches || event.changedTouches.length === 0) return

    const touchEnd = event.changedTouches[0]
    const touchDuration = Date.now() - touchStartTime
    const touchDistance = Math.sqrt(
      Math.pow(touchEnd.clientX - touchStartPos.x, 2) +
      Math.pow(touchEnd.clientY - touchStartPos.y, 2)
    )

    // Если это быстрый тап с небольшим перемещением, то это клик
    if (touchDuration < clickDuration && touchDistance < clickThreshold) {
      // Предотвращаем стандартное поведение, чтобы не конфликтовать с панорамированием
      event.preventDefault()
      // Небольшая задержка, чтобы не конфликтовать с панорамированием
      setTimeout(() => {
        onClick(row, col)
      }, 50)
    }
  }

  return {
    handleTouchStart,
    handleTouchEnd,
  }
}

