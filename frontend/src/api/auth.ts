import axios from 'axios'

const API_BASE = import.meta.env.DEV ? 'http://localhost:8080/api' : '/api'

// Настройка axios для автоматического добавления токена
axios.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// Обработка ошибок авторизации
axios.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Токен невалидный, очищаем и перенаправляем на страницу входа
      localStorage.removeItem('token')
      if (window.location.pathname !== '/login' && window.location.pathname !== '/register') {
        window.location.href = '/login'
      }
    }
    return Promise.reject(error)
  }
)

export interface User {
  id: number
  username: string
  email: string
  color?: string
  rating: number
  createdAt: string
}

export interface RegisterRequest {
  username: string
  email: string
  password: string
}

export interface LoginRequest {
  username: string
  password: string
}

export interface AuthResponse {
  token: string
  user: User
}

export async function register(data: RegisterRequest): Promise<AuthResponse> {
  const response = await axios.post<AuthResponse>(`${API_BASE}/auth/register`, data)
  return response.data
}

export async function login(data: LoginRequest): Promise<AuthResponse> {
  const response = await axios.post<AuthResponse>(`${API_BASE}/auth/login`, data)
  return response.data
}

export async function getMe(): Promise<User> {
  const response = await axios.get<User>(`${API_BASE}/auth/me`)
  return response.data
}

export interface RequestPasswordResetRequest {
  email: string
}

export interface RequestPasswordResetResponse {
  status: string
  message: string
  resetToken?: string // Только для разработки!
  expiresIn?: string
  note?: string
}

export async function requestPasswordReset(data: RequestPasswordResetRequest): Promise<RequestPasswordResetResponse> {
  const response = await axios.post<RequestPasswordResetResponse>(`${API_BASE}/auth/request-password-reset`, data)
  return response.data
}

export interface ResetPasswordRequest {
  token: string
  newPassword: string
}

export interface ResetPasswordResponse {
  status: string
  message: string
}

export async function resetPassword(data: ResetPasswordRequest): Promise<ResetPasswordResponse> {
  const response = await axios.post<ResetPasswordResponse>(`${API_BASE}/auth/reset-password`, data)
  return response.data
}

export interface ResetPasswordByAdminRequest {
  username?: string
  email?: string
  newPassword: string
}

export async function resetPasswordByAdmin(data: ResetPasswordByAdminRequest): Promise<ResetPasswordResponse> {
  const response = await axios.post<ResetPasswordResponse>(`${API_BASE}/auth/reset-password-admin`, data)
  return response.data
}

