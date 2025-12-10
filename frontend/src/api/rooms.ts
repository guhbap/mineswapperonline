import axios from 'axios'

const API_BASE = import.meta.env.DEV ? 'http://localhost:8080/api' : '/api'

export interface Room {
  id: string
  name: string
  hasPassword: boolean
  rows: number
  cols: number
  mines: number
  players: number
  createdAt: string
}

export interface CreateRoomRequest {
  name: string
  password?: string
  rows: number
  cols: number
  mines: number
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

