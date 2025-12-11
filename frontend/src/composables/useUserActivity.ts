import { onMounted, onUnmounted, ref } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { updateActivity } from '@/api/profile'

const ACTIVITY_THROTTLE_MS = 30000 // Обновляем статус не чаще чем раз в 30 секунд

export function useUserActivity() {
  const authStore = useAuthStore()
  const activityThrottleTimer = ref<ReturnType<typeof setTimeout> | null>(null)
  const lastActivityTime = ref(0)

  const handleActivity = () => {
    // Проверяем, что пользователь авторизован
    if (!authStore.isAuthenticated) {
      return
    }

    const now = Date.now()
    
    // Throttling: обновляем статус не чаще чем раз в 30 секунд
    if (now - lastActivityTime.value < ACTIVITY_THROTTLE_MS) {
      return
    }

    lastActivityTime.value = now

    // Очищаем предыдущий таймер, если есть
    if (activityThrottleTimer.value) {
      clearTimeout(activityThrottleTimer.value)
    }

    // Планируем обновление статуса
    activityThrottleTimer.value = setTimeout(async () => {
      try {
        await updateActivity()
      } catch (error) {
        // Игнорируем ошибки, чтобы не мешать пользователю
        // console.error('Failed to update activity:', error)
      }
    }, 100) // Небольшая задержка для батчинга событий
  }

  onMounted(() => {
    // Отслеживаем различные события активности
    const events = ['mousedown', 'mousemove', 'keydown', 'scroll', 'touchstart', 'click']
    
    events.forEach(event => {
      document.addEventListener(event, handleActivity, { passive: true })
    })
  })

  onUnmounted(() => {
    // Очищаем таймер при размонтировании
    if (activityThrottleTimer.value) {
      clearTimeout(activityThrottleTimer.value)
      activityThrottleTimer.value = null
    }

    // Удаляем обработчики событий
    const events = ['mousedown', 'mousemove', 'keydown', 'scroll', 'touchstart', 'click']
    events.forEach(event => {
      document.removeEventListener(event, handleActivity)
    })
  })
}

