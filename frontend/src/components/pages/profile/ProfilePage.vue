<template>
  <main class="profile-page">
    <div v-if="loading" class="loading">
      <p>Загрузка профиля...</p>
    </div>
    <div v-else-if="error" class="error">
      <p>{{ error }}</p>
    </div>
    <div v-else-if="profile" class="profile-content">
      <div class="profile-header">
        <h1 class="profile-title">{{ isOwnProfile ? 'Мой профиль' : 'Профиль пользователя' }}</h1>
        <div class="profile-user-info">
          <div
            class="user-avatar"
            :style="profile.user.color ? { background: profile.user.color } : {}"
          >
            <span class="avatar-text">{{ profile.user.username[0].toUpperCase() }}</span>
          </div>
          <div class="user-details">
            <h2 class="username">{{ profile.user.username }}</h2>
            <p v-if="isOwnProfile" class="email">{{ profile.user.email }}</p>
            <div class="user-rating">
              <span class="rating-label">Рейтинг:</span>
              <span class="rating-value">{{ Math.round(profile.user.rating || 0) }}</span>
            </div>
            <div class="status-badge" :class="{ 'status-online': profile.stats.isOnline, 'status-offline': !profile.stats.isOnline }">
              <span class="status-dot"></span>
              <span>{{ profile.stats.isOnline ? 'В сети' : 'Не в сети' }}</span>
            </div>
          </div>
        </div>

        <!-- Выбор цвета (только для своего профиля) -->
        <div v-if="isOwnProfile" class="color-selector-section">
          <h3 class="color-selector-title">Цвет игрока</h3>
          <div class="color-selector">
            <div
              v-for="colorOption in colorOptions"
              :key="colorOption"
              class="color-option"
              :class="{ 'color-option--selected': selectedColor === colorOption }"
              :style="{ backgroundColor: colorOption }"
              @click="selectColor(colorOption)"
              :title="colorOption"
            >
              <span v-if="selectedColor === colorOption" class="color-check">✓</span>
            </div>
            <button
              v-if="selectedColor"
              @click="clearColor"
              class="color-clear-button"
              title="Сбросить цвет"
            >
              ✕
            </button>
          </div>
          <p v-if="savingColor" class="color-saving">Сохранение...</p>
          <p v-if="colorError" class="color-error">{{ colorError }}</p>
        </div>
      </div>

      <div class="profile-stats">
        <h3 class="stats-title">Статистика игр</h3>
        <div class="stats-grid">
          <div class="stat-card stat-card--rating">
            <div class="stat-value">{{ Math.round(profile.user.rating || 0) }}</div>
            <div class="stat-label">Рейтинг</div>
          </div>
          <div class="stat-card">
            <div class="stat-value">{{ profile.stats.gamesPlayed }}</div>
            <div class="stat-label">Игр сыграно</div>
          </div>
          <div class="stat-card stat-card--win">
            <div class="stat-value">{{ profile.stats.gamesWon }}</div>
            <div class="stat-label">Побед</div>
          </div>
          <div class="stat-card stat-card--loss">
            <div class="stat-value">{{ profile.stats.gamesLost }}</div>
            <div class="stat-label">Поражений</div>
          </div>
          <div class="stat-card stat-card--ratio">
            <div class="stat-value">{{ winRate }}%</div>
            <div class="stat-label">Процент побед</div>
          </div>
        </div>
      </div>

      <div class="profile-info">
        <div class="info-item">
          <span class="info-label">Дата регистрации:</span>
          <span class="info-value">{{ formatDate(profile.user.createdAt) }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">Последний раз в сети:</span>
          <span class="info-value">{{ formatDate(profile.stats.lastSeen) }}</span>
        </div>
        <div v-if="!isOwnProfile" class="info-item">
          <span class="info-label">Email:</span>
          <span class="info-value info-value--private">Скрыто</span>
        </div>
      </div>

      <div class="top-games-section">
        <h3 class="top-games-title">Топ-10 лучших игр</h3>
        <div v-if="topGamesLoading" class="top-games-loading">Загрузка...</div>
        <div v-else-if="topGamesError" class="top-games-error">{{ topGamesError }}</div>
        <div v-else-if="!topGames || topGames.length === 0" class="top-games-empty">
          Пока нет игр с начисленным рейтингом
        </div>
        <div v-else class="top-games-list">
          <div
            v-for="(game, index) in topGames"
            :key="game.id"
            class="top-game-item"
          >
            <div class="game-rank">#{{ index + 1 }}</div>
            <div class="game-info">
              <div class="game-field">
                <span class="game-field-size">{{ game.height }}×{{ game.width }}</span>
                <span class="game-mines">💣 {{ game.mines }}</span>
              </div>
              <div class="game-details">
                <div class="game-time">⏱️ {{ formatTime(game.gameTime) }}</div>
                <div class="game-date">{{ formatDate(game.createdAt) }}</div>
              </div>
            </div>
            <div class="game-rating">
              <div class="rating-gain" v-if="game.rating > 0">
                +{{ Math.round(game.rating) }}
              </div>
              <div class="rating-label" v-if="game.rating > 0">рейтинг</div>
              <div class="rating-label" v-else style="color: var(--text-secondary);">—</div>
            </div>
          </div>
        </div>
      </div>

      <div class="recent-games-section">
        <h3 class="recent-games-title">Последние 10 игр</h3>
        <div v-if="recentGamesLoading" class="recent-games-loading">Загрузка...</div>
        <div v-else-if="recentGamesError" class="recent-games-error">{{ recentGamesError }}</div>
        <div v-else-if="!recentGames || recentGames.length === 0" class="recent-games-empty">
          Пока нет сыгранных игр
        </div>
        <div v-else class="recent-games-list">
          <div
            v-for="game in recentGames"
            :key="game.id"
            class="recent-game-item"
          >
            <div class="game-main-info">
              <div class="game-field-info">
                <div class="game-field">
                  <span class="game-field-size">{{ game.height }}×{{ game.width }}</span>
                  <span class="game-mines">💣 {{ game.mines }}</span>
                </div>
                <div class="game-complexity">
                  <span class="complexity-label">Сложность:</span>
                  <span class="complexity-value">{{ calculateDifficulty(game.width, game.height, game.mines) }}</span>
                </div>
              </div>
              <div class="game-time-info">
                <div class="game-time">⏱️ {{ formatTime(game.gameTime) }}</div>
                <div class="game-date">{{ formatDate(game.createdAt) }}</div>
              </div>
            </div>
            <div class="game-rating-info">
              <div class="rating-label">Рейтинг за игру:</div>
              <div class="rating-value" v-if="game.rating > 0">
                +{{ Math.round(game.rating) }}
              </div>
              <div class="rating-value rating-value--none" v-else>
                —
              </div>
            </div>
            <div v-if="game.participants && game.participants.length > 0" class="game-participants">
              <div class="participants-label">Участники:</div>
              <div class="participants-list">
                <span
                  v-for="(participant, index) in game.participants"
                  :key="participant.userId"
                  class="participant-badge"
                  :style="participant.color ? { borderColor: participant.color } : {}"
                >
                  {{ participant.nickname }}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </main>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { getProfile, getProfileByUsername, updateColor, getTopGames, getRecentGames, type UserProfile, type TopGame, type RecentGame } from '@/api/profile'
import { getErrorMessage } from '@/utils/errorHandler'
import { calculateDifficulty } from '@/utils/ratingCalculator'

const route = useRoute()
const authStore = useAuthStore()
const profile = ref<UserProfile | null>(null)
const loading = ref(true)
const error = ref('')
const isOwnProfile = ref(true)
const selectedColor = ref<string>('')
const savingColor = ref(false)
const colorError = ref('')
const topGames = ref<TopGame[]>([])
const topGamesLoading = ref(false)
const topGamesError = ref('')
const recentGames = ref<RecentGame[]>([])
const recentGamesLoading = ref(false)
const recentGamesError = ref('')

const colorOptions = [
  '#FF6B6B', '#4ECDC4', '#45B7D1', '#FFA07A', '#98D8C8',
  '#F7DC6F', '#BB8FCE', '#85C1E2', '#F8B739', '#52BE80',
  '#E74C3C', '#3498DB', '#9B59B6', '#1ABC9C', '#F39C12',
  '#E67E22', '#95A5A6', '#34495E', '#16A085', '#27AE60'
]

const winRate = computed(() => {
  if (!profile.value || profile.value.stats.gamesPlayed === 0) {
    return 0
  }
  return Math.round((profile.value.stats.gamesWon / profile.value.stats.gamesPlayed) * 100)
})

const formatDate = (dateString: string) => {
  const date = new Date(dateString)
  return date.toLocaleDateString('ru-RU', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const formatTime = (seconds: number) => {
  if (seconds < 60) {
    return `${Math.round(seconds)}с`
  }
  const minutes = Math.floor(seconds / 60)
  const secs = Math.round(seconds % 60)
  return `${minutes}м ${secs}с`
}

const loadProfile = async () => {
  try {
    loading.value = true
    error.value = ''

    // Проверяем, есть ли username в параметрах роута
    const username = route.params.username as string
    if (username) {
      // Загружаем профиль другого пользователя
      isOwnProfile.value = false
      profile.value = await getProfileByUsername(username)
      // Загружаем топ-10 игр для этого пользователя (публичный запрос с username)
      await loadTopGames(username)
      await loadRecentGames(username)
    } else {
      // Загружаем свой профиль
      isOwnProfile.value = true
      profile.value = await getProfile()
      selectedColor.value = profile.value?.user.color || ''
      // Загружаем топ-10 игр для себя (защищенный запрос без username)
      await loadTopGames()
      await loadRecentGames()
    }
  } catch (err: any) {
    error.value = getErrorMessage(err, 'Ошибка загрузки профиля')
    console.error('Ошибка загрузки профиля:', err)
  } finally {
    loading.value = false
  }
}

const loadTopGames = async (username?: string) => {
  try {
    topGamesLoading.value = true
    topGamesError.value = ''
    const games = await getTopGames(username)
    topGames.value = games || []
  } catch (err: any) {
    topGamesError.value = getErrorMessage(err, 'Ошибка загрузки игр')
    topGames.value = [] // Убеждаемся, что это всегда массив
    console.error('Ошибка загрузки топ-10 игр:', err)
  } finally {
    topGamesLoading.value = false
  }
}

const loadRecentGames = async (username?: string) => {
  try {
    recentGamesLoading.value = true
    recentGamesError.value = ''
    const games = await getRecentGames(username)
    recentGames.value = games || []
  } catch (err: any) {
    recentGamesError.value = getErrorMessage(err, 'Ошибка загрузки последних игр')
    recentGames.value = [] // Убеждаемся, что это всегда массив
    console.error('Ошибка загрузки последних игр:', err)
  } finally {
    recentGamesLoading.value = false
  }
}

const selectColor = async (color: string) => {
  if (selectedColor.value === color) return

  selectedColor.value = color
  savingColor.value = true
  colorError.value = ''

  try {
    await updateColor(color)
    // Обновляем профиль после сохранения
    if (profile.value) {
      profile.value.user.color = color
    }
    // Обновляем цвет в auth store
    if (authStore.user) {
      authStore.user.color = color
    }
  } catch (err: any) {
    colorError.value = getErrorMessage(err, 'Ошибка сохранения цвета')
    // Откатываем выбор
    selectedColor.value = profile.value?.user.color || ''
  } finally {
    savingColor.value = false
  }
}

const clearColor = async () => {
  selectedColor.value = ''
  savingColor.value = true
  colorError.value = ''

  try {
    await updateColor('')
    // Обновляем профиль после сохранения
    if (profile.value) {
      profile.value.user.color = undefined
    }
    // Обновляем цвет в auth store
    if (authStore.user) {
      authStore.user.color = undefined
    }
  } catch (err: any) {
    colorError.value = getErrorMessage(err, 'Ошибка сохранения цвета')
    // Откатываем выбор
    selectedColor.value = profile.value?.user.color || ''
  } finally {
    savingColor.value = false
  }
}

// Отслеживаем изменения роута
watch(() => route.params.username, () => {
  loadProfile()
})

onMounted(() => {
  loadProfile()
})
</script>

<style scoped>
.profile-page {
  max-width: 1200px;
  margin: 0 auto;
  padding: 2rem;
  min-height: 100vh;
}

.loading,
.error {
  text-align: center;
  padding: 3rem;
  color: var(--text-primary);
}

.error {
  color: #dc2626;
}

.profile-content {
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.profile-header {
  background: var(--bg-primary);
  padding: 2rem;
  border-radius: 1rem;
  box-shadow: 0 2px 8px var(--shadow);
}

.profile-title {
  margin: 0 0 2rem 0;
  font-size: 2rem;
  color: var(--text-primary);
}

.profile-user-info {
  display: flex;
  align-items: center;
  gap: 2rem;
}

.user-avatar {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 2rem;
  font-weight: 700;
  color: white;
}

.avatar-text {
  text-transform: uppercase;
}

.user-details {
  flex: 1;
}

.username {
  margin: 0 0 0.5rem 0;
  font-size: 1.5rem;
  color: var(--text-primary);
}

.email {
  margin: 0 0 1rem 0;
  color: var(--text-secondary);
}

.user-rating {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  margin: 0 0 1rem 0;
  padding: 0.5rem 1rem;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 1rem;
  color: white;
  font-weight: 600;
}

.rating-label {
  font-size: 0.875rem;
  opacity: 0.9;
}

.rating-value {
  font-size: 1.25rem;
  font-weight: 700;
}

.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  border-radius: 1rem;
  font-size: 0.875rem;
  font-weight: 600;
}

.status-online {
  background: rgba(34, 197, 94, 0.1);
  color: #22c55e;
}

.status-offline {
  background: rgba(107, 114, 128, 0.1);
  color: var(--text-secondary);
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: currentColor;
}

.profile-stats {
  background: var(--bg-primary);
  padding: 2rem;
  border-radius: 1rem;
  box-shadow: 0 2px 8px var(--shadow);
}

.stats-title {
  margin: 0 0 1.5rem 0;
  font-size: 1.5rem;
  color: var(--text-primary);
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1.5rem;
}

.stat-card {
  background: var(--bg-secondary);
  padding: 1.5rem;
  border-radius: 0.75rem;
  text-align: center;
  transition: transform 0.2s, box-shadow 0.2s;
}

.stat-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px var(--shadow);
}

.stat-card--win {
  border-left: 4px solid #22c55e;
}

.stat-card--loss {
  border-left: 4px solid #dc2626;
}

.stat-card--ratio {
  border-left: 4px solid #667eea;
}

.stat-card--rating {
  border-left: 4px solid #f59e0b;
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.1) 0%, rgba(118, 75, 162, 0.1) 100%);
}

.stat-value {
  font-size: 2.5rem;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 0.5rem;
}

.stat-label {
  font-size: 0.875rem;
  color: var(--text-secondary);
  font-weight: 600;
}

.profile-info {
  background: var(--bg-primary);
  padding: 2rem;
  border-radius: 1rem;
  box-shadow: 0 2px 8px var(--shadow);
}

.info-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 0;
  border-bottom: 1px solid var(--border-color);
}

.info-item:last-child {
  border-bottom: none;
}

.info-label {
  color: var(--text-secondary);
  font-weight: 600;
}

.info-value {
  color: var(--text-primary);
}

.info-value--private {
  color: var(--text-secondary);
  font-style: italic;
}

.color-selector-section {
  margin-top: 2rem;
  padding-top: 2rem;
  border-top: 2px solid var(--border-color);
}

.color-selector-title {
  margin: 0 0 1rem 0;
  font-size: 1.25rem;
  color: var(--text-primary);
}

.color-selector {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  align-items: center;
}

.color-option {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  cursor: pointer;
  border: 3px solid transparent;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 2px 4px var(--shadow);
}

.color-option:hover {
  transform: scale(1.1);
  box-shadow: 0 4px 8px var(--shadow);
}

.color-option--selected {
  border-color: var(--text-primary);
  transform: scale(1.15);
  box-shadow: 0 4px 12px var(--shadow);
}

.color-check {
  color: white;
  font-size: 1.25rem;
  font-weight: 700;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.5);
}

.color-clear-button {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  border: 2px solid var(--border-color);
  background: var(--bg-secondary);
  color: var(--text-secondary);
  cursor: pointer;
  font-size: 1.25rem;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
}

.color-clear-button:hover {
  background: var(--bg-tertiary);
  border-color: var(--text-secondary);
  transform: scale(1.1);
}

.color-saving {
  margin-top: 0.5rem;
  font-size: 0.875rem;
  color: var(--text-secondary);
  font-style: italic;
}

.color-error {
  margin-top: 0.5rem;
  font-size: 0.875rem;
  color: #dc2626;
}

@media (max-width: 768px) {
  .profile-page {
    padding: 1rem;
  }

  .profile-user-info {
    flex-direction: column;
    text-align: center;
  }

  .stats-grid {
    grid-template-columns: 1fr;
  }

  .info-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 0.5rem;
  }
}

.top-games-section {
  background: var(--bg-primary);
  padding: 2rem;
  border-radius: 1rem;
  box-shadow: 0 2px 8px var(--shadow);
}

.top-games-title {
  margin: 0 0 1.5rem 0;
  font-size: 1.5rem;
  color: var(--text-primary);
}

.top-games-loading,
.top-games-error,
.top-games-empty {
  text-align: center;
  padding: 2rem;
  color: var(--text-secondary);
}

.top-games-error {
  color: #dc2626;
}

.top-games-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.top-game-item {
  display: flex;
  align-items: center;
  gap: 1.5rem;
  padding: 1.25rem;
  background: var(--bg-secondary);
  border-radius: 0.75rem;
  transition: transform 0.2s, box-shadow 0.2s;
  border-left: 4px solid #667eea;
}

.top-game-item:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px var(--shadow);
}

.game-rank {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--text-secondary);
  min-width: 3rem;
  text-align: center;
}

.game-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.game-field {
  display: flex;
  align-items: center;
  gap: 1rem;
  font-weight: 600;
  color: var(--text-primary);
}

.game-field-size {
  font-size: 1.125rem;
}

.game-mines {
  font-size: 0.875rem;
  color: var(--text-secondary);
}

.game-details {
  display: flex;
  gap: 1.5rem;
  font-size: 0.875rem;
  color: var(--text-secondary);
}

.game-rating {
  text-align: right;
  min-width: 120px;
}

.rating-gain {
  font-size: 1.25rem;
  font-weight: 700;
  color: #22c55e;
  margin-bottom: 0.25rem;
}

.rating-change {
  font-size: 0.875rem;
  color: var(--text-secondary);
}

@media (max-width: 768px) {
  .top-game-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 1rem;
  }

  .game-rating {
    text-align: left;
    width: 100%;
  }

  .game-details {
    flex-direction: column;
    gap: 0.5rem;
  }
}

.recent-games-section {
  background: var(--bg-primary);
  padding: 2rem;
  border-radius: 1rem;
  box-shadow: 0 2px 8px var(--shadow);
}

.recent-games-title {
  margin: 0 0 1.5rem 0;
  font-size: 1.5rem;
  color: var(--text-primary);
}

.recent-games-loading,
.recent-games-error,
.recent-games-empty {
  text-align: center;
  padding: 2rem;
  color: var(--text-secondary);
}

.recent-games-error {
  color: #dc2626;
}

.recent-games-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.recent-game-item {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  padding: 1.25rem;
  background: var(--bg-secondary);
  border-radius: 0.75rem;
  transition: transform 0.2s, box-shadow 0.2s;
  border-left: 4px solid #667eea;
}

.recent-game-item:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px var(--shadow);
}

.game-main-info {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 1rem;
  flex-wrap: wrap;
}

.game-field-info {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.game-complexity {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.875rem;
}

.complexity-label {
  color: var(--text-secondary);
  font-weight: 500;
}

.complexity-value {
  color: #667eea;
  font-weight: 700;
  font-size: 1rem;
}

.game-time-info {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  font-size: 0.875rem;
  color: var(--text-secondary);
  text-align: right;
}

.game-participants {
  padding-top: 0.75rem;
  border-top: 1px solid var(--border-color);
}

.participants-label {
  font-size: 0.875rem;
  color: var(--text-secondary);
  font-weight: 600;
  margin-bottom: 0.5rem;
}

.participants-list {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  align-items: center;
}

.participant-badge {
  display: inline-flex;
  align-items: center;
  padding: 0.375rem 0.75rem;
  background: var(--bg-tertiary);
  border: 2px solid var(--border-color);
  border-radius: 0.5rem;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--text-primary);
}

.participant-separator {
  margin-left: 0.25rem;
  color: var(--text-secondary);
}

.game-rating-info {
  padding-top: 0.75rem;
  border-top: 1px solid var(--border-color);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.game-rating-info .rating-label {
  color: var(--text-secondary);
  font-weight: 600;
  font-size: 0.875rem;
}

.game-rating-info .rating-value {
  font-size: 1.25rem;
  font-weight: 700;
  color: #22c55e;
}

.game-rating-info .rating-value--none {
  color: var(--text-secondary);
  font-weight: 500;
}

@media (max-width: 768px) {
  .recent-game-item {
    gap: 0.75rem;
  }

  .game-main-info {
    flex-direction: column;
    gap: 0.75rem;
  }

  .game-time-info {
    text-align: left;
  }

  .game-rating-info {
    text-align: left;
  }
}
</style>

