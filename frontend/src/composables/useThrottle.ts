import { ref } from 'vue'

/**
 * Создает throttled функцию, которая вызывается не чаще указанного интервала
 */
export function useThrottle<T extends (...args: any[]) => any>(
  fn: T,
  delay: number
): T {
  let lastCall = 0
  let timeoutId: ReturnType<typeof setTimeout> | null = null
  let lastArgs: Parameters<T> | null = null

  return ((...args: Parameters<T>) => {
    const now = Date.now()
    const timeSinceLastCall = now - lastCall

    lastArgs = args

    if (timeSinceLastCall >= delay) {
      lastCall = now
      fn(...args)
      lastArgs = null
    } else {
      if (timeoutId) {
        clearTimeout(timeoutId)
      }
      timeoutId = setTimeout(() => {
        if (lastArgs) {
          lastCall = Date.now()
          fn(...lastArgs)
          lastArgs = null
        }
        timeoutId = null
      }, delay - timeSinceLastCall)
    }
  }) as T
}

