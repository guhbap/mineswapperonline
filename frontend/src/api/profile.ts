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
    color?: string
    rating: number
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

export async function updateColor(color: string): Promise<void> {
  await axios.post(`${API_BASE}/profile/color`, { color })
}

export async function changePassword(currentPassword: string, newPassword: string): Promise<void> {
  await axios.post(`${API_BASE}/profile/change-password`, {
    currentPassword,
    newPassword
  })
}

export interface LeaderboardEntry {
  id: number
  username: string
  color?: string
  rating: number
  gamesPlayed: number
  gamesWon: number
  gamesLost: number
}

export async function getLeaderboard(): Promise<LeaderboardEntry[]> {
  const response = await axios.get<LeaderboardEntry[]>(`${API_BASE}/leaderboard`)
  return response.data
}

export interface TopGame {
  id: number
  width: number
  height: number
  mines: number
  gameTime: number
  rating: number
  ratingPercent: number  // Процент засчитанного рейтинга (100%, 95%, 90.25% и т.д.)
  ratingContributed: number  // Конкретно полученный рейтинг (рейтинг * коэффициент)
  won: boolean
  createdAt: string
}

export async function getTopGames(username?: string): Promise<TopGame[]> {
  const url = username
    ? `${API_BASE}/profile/top-games?username=${encodeURIComponent(username)}`
    : `${API_BASE}/profile/top-games`
  const response = await axios.get<TopGame[]>(url)
  return response.data
}

export interface GameParticipant {
  userId: number
  nickname: string
  color?: string
}

export interface RecentGame {
  id: number
  width: number
  height: number
  mines: number
  gameTime: number
  rating: number
  won: boolean
  createdAt: string
  participants: GameParticipant[]
}

export async function getRecentGames(username?: string): Promise<RecentGame[]> {
  const url = username
    ? `${API_BASE}/profile/recent-games?username=${encodeURIComponent(username)}`
    : `${API_BASE}/profile/recent-games`
  const response = await axios.get<RecentGame[]>(url)
  return response.data
}

export interface GameParticipantInfo {
  userId: number
  username: string
  nickname: string
  color?: string
}

export interface GameDetails {
  id: number
  roomId: string
  width: number
  height: number
  mines: number
  seed: string
  hasCustomSeed: boolean
  creatorId: number
  creatorName: string
  won: boolean
  chording: boolean
  quickStart: boolean
  startTime: string
  duration: number
  rating: number
  participants: GameParticipantInfo[]
}

export async function getGameDetails(gameId: number): Promise<GameDetails> {
  const response = await axios.get<GameDetails>(`${API_BASE}/game/details?id=${gameId}`)
  return response.data
}

