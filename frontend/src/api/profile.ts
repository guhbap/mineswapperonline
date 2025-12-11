import axios from 'axios'

const API_BASE = import.meta.env.DEV ? 'http://localhost:8080/api' : '/api'

export interface UserStats {
  userId: number
  gamesPlayed: number
  gamesWon: number
  gamesLost: number
  lastSeen: string
  isOnline: boolean
}

export interface UserProfile {
  user: {
    id: number
    username: string
    email: string
    createdAt: string
  }
  stats: UserStats
}

export async function getProfile(): Promise<UserProfile> {
  const response = await axios.get<UserProfile>(`${API_BASE}/profile`)
  return response.data
}

export async function getProfileByUsername(username: string): Promise<UserProfile> {
  const response = await axios.get<UserProfile>(`${API_BASE}/profile?username=${encodeURIComponent(username)}`)
  return response.data
}

export async function updateActivity(): Promise<void> {
  await axios.post(`${API_BASE}/profile/activity`)
}

