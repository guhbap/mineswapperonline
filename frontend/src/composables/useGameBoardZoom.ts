import { ref, computed, nextTick, type Ref } from 'vue'

export interface UseGameBoardZoomOptions {
  minZoom?: number
  maxZoom?: number
  zoomStep?: number
  initialZoom?: number
  wrapperSelector?: string
}

export interface UseGameBoardZoomReturn {
  zoomLevel: Ref<number>
  zoomPercentage: Ref<number>
  isZoomed: Ref<boolean>
  canZoomIn: Ref<boolean>
  canZoomOut: Ref<boolean>
  zoomIn: () => void
  zoomOut: () => void
  resetZoom: () => void
  setZoom: (level: number) => void
  handleTouchStart: (event: TouchEvent) => void
  handleTouchMove: (event: TouchEvent) => void
  handleTouchEnd: () => void
  containerStyle: Ref<{ transform: string; transformOrigin: string }>
}

const DEFAULT_OPTIONS: Required<Omit<UseGameBoardZoomOptions, 'wrapperSelector'>> = {
  minZoom: 0.5,
  maxZoom: 3,
  zoomStep: 0.1,
  initialZoom: 1,
}

/**
 * Composable для управления зумом игрового поля на мобильных устройствах
 * Поддерживает pinch-to-zoom и программное управление зумом
 */
export function useGameBoardZoom(
  options: UseGameBoardZoomOptions = {}
): UseGameBoardZoomReturn {
  const {
    minZoom = DEFAULT_OPTIONS.minZoom,
    maxZoom = DEFAULT_OPTIONS.maxZoom,
    zoomStep = DEFAULT_OPTIONS.zoomStep,
    initialZoom = DEFAULT_OPTIONS.initialZoom,
    wrapperSelector = '.game-board-wrapper',
  } = options

  // Состояние зума
  const zoomLevel = ref(initialZoom)

  // Вычисляемые свойства
  const zoomPercentage = computed(() => Math.round(zoomLevel.value * 100))
  const isZoomed = computed(() => zoomLevel.value !== 1)
  const canZoomIn = computed(() => zoomLevel.value < maxZoom)
  const canZoomOut = computed(() => zoomLevel.value > minZoom)

  // Стили для контейнера
  const containerStyle = computed(() => ({
    transform: `scale(${zoomLevel.value})`,
    transformOrigin: 'center center',
  }))

  /**
   * Увеличивает зум на один шаг
   */
  const zoomIn = () => {
    if (canZoomIn.value) {
      zoomLevel.value = Math.min(zoomLevel.value + zoomStep, maxZoom)
    }
  }

  /**
   * Уменьшает зум на один шаг
   */
  const zoomOut = () => {
    if (canZoomOut.value) {
      zoomLevel.value = Math.max(zoomLevel.value - zoomStep, minZoom)
    }
  }

  /**
   * Устанавливает конкретный уровень зума
   */
  const setZoom = (level: number) => {
    zoomLevel.value = Math.max(minZoom, Math.min(maxZoom, level))
  }

  /**
   * Сбрасывает зум до начального значения и центрирует поле
   */
  const resetZoom = () => {
    zoomLevel.value = initialZoom
    // Прокручиваем контейнер в центр
    nextTick(() => {
      const wrapper = document.querySelector(wrapperSelector) as HTMLElement
      if (wrapper) {
        wrapper.scrollTo({
          left: wrapper.scrollWidth / 2 - wrapper.clientWidth / 2,
          top: wrapper.scrollHeight / 2 - wrapper.clientHeight / 2,
          behavior: 'smooth',
        })
      }
    })
  }

  // Состояние для pinch-to-zoom
  const touchStartDistance = ref(0)
  const touchStartZoom = ref(1)
  const isPanning = ref(false)

  /**
   * Вычисляет расстояние между двумя точками касания
   */
  const getTouchDistance = (touches: TouchList): number => {
    if (touches.length < 2) return 0
    const dx = touches[0].clientX - touches[1].clientX
    const dy = touches[0].clientY - touches[1].clientY
    return Math.sqrt(dx * dx + dy * dy)
  }

  /**
   * Обработчик начала касания
   * Обрабатываем только pinch-to-zoom (2 пальца), панорамирование - через браузер
   */
  const handleTouchStart = (event: TouchEvent) => {
    // Обрабатываем только если это два пальца для pinch-to-zoom
    if (event.touches.length === 2) {
      touchStartDistance.value = getTouchDistance(event.touches)
      touchStartZoom.value = zoomLevel.value
      isPanning.value = false
    } else {
      // Для одного пальца сбрасываем состояние pinch-to-zoom
      touchStartDistance.value = 0
    }
    // Не предотвращаем событие - позволяем браузеру обрабатывать скролл
  }

  /**
   * Обработчик движения касания
   * Предотвращаем только для pinch-to-zoom, чтобы не мешать панорамированию
   */
  const handleTouchMove = (event: TouchEvent) => {
    // Обрабатываем только pinch-to-zoom (2 пальца)
    if (event.touches.length === 2 && touchStartDistance.value > 0) {
      // Предотвращаем стандартное поведение только для pinch-to-zoom
      event.preventDefault()
      const currentDistance = getTouchDistance(event.touches)
      if (currentDistance > 0 && touchStartDistance.value > 0) {
        const scale = currentDistance / touchStartDistance.value
        const newZoom = touchStartZoom.value * scale
        setZoom(newZoom)
      }
      isPanning.value = false
      return
    }
    // Для одного пальца ничего не делаем - позволяем браузеру обрабатывать скролл
    // Не вызываем preventDefault, чтобы скролл работал
  }

  /**
   * Обработчик окончания касания
   */
  const handleTouchEnd = (event: TouchEvent) => {
    // Если остался один палец после pinch-to-zoom, сбрасываем состояние
    if (event.touches.length === 0 || event.touches.length === 1) {
      touchStartDistance.value = 0
      isPanning.value = false
    }
    // Не предотвращаем событие - позволяем браузеру обрабатывать скролл
  }

  return {
    zoomLevel,
    zoomPercentage,
    isZoomed,
    canZoomIn,
    canZoomOut,
    zoomIn,
    zoomOut,
    resetZoom,
    setZoom,
    handleTouchStart,
    handleTouchMove,
    handleTouchEnd,
    containerStyle,
  }
}

