import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { User } from '@/api/auth'
import { login, register, getMe } from '@/api/auth'
import { getErrorMessage } from '@/utils/errorHandler'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const token = ref<string | null>(localStorage.getItem('token'))
  const loading = ref(false)
  const error = ref<string | null>(null)

  const isAuthenticated = computed(() => !!token.value && !!user.value)

  // Инициализация при загрузке приложения
  async function init() {
    const storedToken = localStorage.getItem('token')
    if (storedToken) {
      token.value = storedToken
      try {
        user.value = await getMe()
      } catch (err) {
        // Токен невалидный, очищаем
        logout()
      }
    }
  }

  async function loginUser(username: string, password: string) {
    loading.value = true
    error.value = null
    try {
      const response = await login({ username, password })
      token.value = response.token
      user.value = response.user
      localStorage.setItem('token', response.token)
      return response
    } catch (err: any) {
      error.value = getErrorMessage(err, 'Ошибка входа')
      throw err
    } finally {
      loading.value = false
    }
  }

  async function registerUser(username: string, email: string, password: string) {
    loading.value = true
    error.value = null
    try {
      const response = await register({ username, email, password })
      token.value = response.token
      user.value = response.user
      localStorage.setItem('token', response.token)
      return response
    } catch (err: any) {
      error.value = getErrorMessage(err, 'Ошибка регистрации')
      throw err
    } finally {
      loading.value = false
    }
  }

  function logout() {
    user.value = null
    token.value = null
    localStorage.removeItem('token')
  }

  return {
    user,
    token,
    loading,
    error,
    isAuthenticated,
    init,
    loginUser,
    registerUser,
    logout
  }
})

