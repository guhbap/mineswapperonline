import { ref, watch, onMounted } from 'vue'

type Theme = 'light' | 'dark'

const theme = ref<Theme>('light')
let initialized = false

const applyTheme = (newTheme: Theme) => {
  document.documentElement.setAttribute('data-theme', newTheme)
  localStorage.setItem('theme', newTheme)
}

const initTheme = () => {
  if (initialized) return
  
  const savedTheme = localStorage.getItem('theme') as Theme | null
  if (savedTheme && (savedTheme === 'light' || savedTheme === 'dark')) {
    theme.value = savedTheme
  } else {
    // Проверяем системную тему
    const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
    theme.value = prefersDark ? 'dark' : 'light'
  }
  applyTheme(theme.value)
  initialized = true
}

export function useTheme() {
  // Инициализируем тему при первом использовании
  if (typeof window !== 'undefined' && !initialized) {
    initTheme()
  }

  const toggleTheme = () => {
    theme.value = theme.value === 'light' ? 'dark' : 'light'
    applyTheme(theme.value)
  }

  const setTheme = (newTheme: Theme) => {
    theme.value = newTheme
    applyTheme(newTheme)
  }

  onMounted(() => {
    initTheme()
    
    // Слушаем изменения системной темы
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
    const handleChange = (e: MediaQueryListEvent) => {
      if (!localStorage.getItem('theme')) {
        setTheme(e.matches ? 'dark' : 'light')
      }
    }
    mediaQuery.addEventListener('change', handleChange)
  })

  watch(theme, (newTheme) => {
    applyTheme(newTheme)
  })

  return {
    theme,
    toggleTheme,
    setTheme,
    initTheme
  }
}

