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
      const requestUrl = error.config?.url || ''
      const requestMethod = error.config?.method?.toLowerCase() || ''

      // Не делаем редирект для ошибок пароля комнаты (POST /rooms/join)
      // Редирект только для ошибок авторизации пользователя (невалидный токен)
      const isRoomPasswordError =
        requestMethod === 'post' &&
        (requestUrl.endsWith('/rooms/join') || requestUrl.includes('/api/rooms/join'))

      if (!isRoomPasswordError) {
        // Токен невалидный, очищаем и перенаправляем на страницу входа
        localStorage.removeItem('token')
        if (window.location.pathname !== '/login' && window.location.pathname !== '/register') {
          window.location.href = '/login'
        }
      }
    }
    return Promise.reject(error)
  }
)

export interface Room {
  id: string
  name: string
  hasPassword: boolean
  rows: number
  cols: number
  mines: number
  gameMode?: string
  quickStart?: boolean
  chording?: boolean
  players: number
  createdAt: string
  creatorId?: number
}

export interface CreateRoomRequest {
  name: string
  password?: string
  rows: number
  cols: number
  mines: number
  gameMode: string
  quickStart: boolean
  chording: boolean
  seed?: number | null
}

export interface JoinRoomRequest {
  roomId: string
  password?: string
}

export async function getRooms(): Promise<Room[]> {
  const response = await axios.get<Room[]>(`${API_BASE}/rooms`)
  return response.data
}

export async function createRoom(data: CreateRoomRequest): Promise<Room> {
  const response = await axios.post<Room>(`${API_BASE}/rooms`, data)
  return response.data
}

export async function joinRoom(data: JoinRoomRequest): Promise<Room> {
  const response = await axios.post<Room>(`${API_BASE}/rooms/join`, data)
  return response.data
}

export interface UpdateRoomRequest {
  name: string
  password?: string
  rows: number
  cols: number
  mines: number
  gameMode?: string
  quickStart?: boolean
  chording?: boolean
}

export async function updateRoom(roomId: string, data: UpdateRoomRequest): Promise<Room> {
  const response = await axios.put<Room>(`${API_BASE}/rooms/${roomId}`, data)
  return response.data
}

