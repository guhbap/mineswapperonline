import { ref, onMounted, onUnmounted } from 'vue'

interface CursorPosition {
  x: number
  y: number
  targetX: number
  targetY: number
}

/**
 * Создает плавную анимацию курсора с интерполяцией
 */
export function useCursorAnimation() {
  const animatedCursors = ref<Map<string, CursorPosition>>(new Map())
  let animationFrameId: number | null = null
  let isAnimating = false

  const updateCursor = (playerId: string, x: number, y: number) => {
    const current = animatedCursors.value.get(playerId)
    
    if (current) {
      // Обновляем целевую позицию
      current.targetX = x
      current.targetY = y
    } else {
      // Создаем новый курсор с начальной и целевой позицией
      animatedCursors.value.set(playerId, {
        x,
        y,
        targetX: x,
        targetY: y
      })
    }

    // Запускаем анимацию, если она еще не запущена
    if (!isAnimating) {
      startAnimation()
    }
  }

  const removeCursor = (playerId: string) => {
    animatedCursors.value.delete(playerId)
    // Останавливаем анимацию, если нет курсоров
    if (animatedCursors.value.size === 0 && animationFrameId) {
      cancelAnimationFrame(animationFrameId)
      animationFrameId = null
      isAnimating = false
    }
  }

  const animate = () => {
    const lerpFactor = 0.2 // Коэффициент интерполяции (0-1, больше = быстрее, но менее плавно)
    
    animatedCursors.value.forEach((cursor) => {
      // Интерполяция к целевой позиции
      const dx = cursor.targetX - cursor.x
      const dy = cursor.targetY - cursor.y
      
      // Если расстояние очень маленькое, сразу устанавливаем целевую позицию
      if (Math.abs(dx) < 0.05 && Math.abs(dy) < 0.05) {
        cursor.x = cursor.targetX
        cursor.y = cursor.targetY
      } else {
        // Используем экспоненциальное сглаживание для более плавного движения
        cursor.x += dx * lerpFactor
        cursor.y += dy * lerpFactor
      }
    })

    if (animatedCursors.value.size > 0) {
      animationFrameId = requestAnimationFrame(animate)
    } else {
      isAnimating = false
      animationFrameId = null
    }
  }

  const startAnimation = () => {
    if (!isAnimating) {
      isAnimating = true
      animate()
    }
  }

  onMounted(() => {
    // Запускаем анимацию при монтировании, если есть курсоры
    if (animatedCursors.value.size > 0) {
      startAnimation()
    }
  })

  onUnmounted(() => {
    if (animationFrameId) {
      cancelAnimationFrame(animationFrameId)
      animationFrameId = null
    }
    isAnimating = false
  })

  return {
    animatedCursors,
    updateCursor,
    removeCursor
  }
}

