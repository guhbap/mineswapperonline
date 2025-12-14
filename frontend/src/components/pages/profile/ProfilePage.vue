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
              <span class="">Рейтинг:</span>
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

        <!-- Смена пароля (только для своего профиля) -->
        <div v-if="isOwnProfile" class="change-password-section">
          <h3 class="change-password-title">Смена пароля</h3>
          <form @submit.prevent="handleChangePassword" class="change-password-form">
            <TextInput
              v-model="currentPassword"
              label="Текущий пароль"
              placeholder="Введите текущий пароль"
              name="currentPassword"
              type="password"
              :disabled="changingPassword"
            />
            <TextInput
              v-model="newPassword"
              label="Новый пароль"
              placeholder="Введите новый пароль (минимум 6 символов)"
              name="newPassword"
              type="password"
              :disabled="changingPassword"
            />
            <TextInput
              v-model="confirmPassword"
              label="Подтвердите новый пароль"
              placeholder="Повторите новый пароль"
              name="confirmPassword"
              type="password"
              :disabled="changingPassword"
            />
            <div v-if="passwordError" class="password-error">{{ passwordError }}</div>
            <div v-if="passwordSuccess" class="password-success">{{ passwordSuccess }}</div>
            <button type="submit" class="change-password-button" :disabled="changingPassword">
              {{ changingPassword ? 'Сохранение...' : 'Изменить пароль' }}
            </button>
          </form>
        </div>

        <!-- Админ-панель (только для администратора) -->
        <div v-if="isOwnProfile && isAdmin" class="admin-section">
          <h3 class="admin-section-title">
            <IconSettings class="admin-section-icon" />
            Админ-панель
          </h3>
          <div class="admin-panel">
            <h4 class="admin-panel-title">Сброс пароля пользователя</h4>
            <form @submit.prevent="handleAdminResetPassword" class="admin-form">
              <div class="admin-form-group">
                <label class="admin-label">Найти пользователя по:</label>
                <div class="admin-radio-group">
                  <label class="admin-radio-label">
                    <input
                      v-model="adminSearchType"
                      type="radio"
                      value="username"
                      class="admin-radio"
                    />
                    <span>Username</span>
                  </label>
                  <label class="admin-radio-label">
                    <input
                      v-model="adminSearchType"
                      type="radio"
                      value="email"
                      class="admin-radio"
                    />
                    <span>Email</span>
                  </label>
                </div>
              </div>
              <TextInput
                v-model="adminSearchValue"
                :label="adminSearchType === 'username' ? 'Username пользователя' : 'Email пользователя'"
                :placeholder="adminSearchType === 'username' ? 'Введите username' : 'Введите email'"
                name="adminSearch"
                :disabled="adminResettingPassword"
              />
              <TextInput
                v-model="adminNewPassword"
                label="Новый пароль"
                placeholder="Введите новый пароль (минимум 6 символов)"
                name="adminNewPassword"
                type="password"
                :disabled="adminResettingPassword"
              />
              <div v-if="adminPasswordError" class="admin-error">{{ adminPasswordError }}</div>
              <div v-if="adminPasswordSuccess" class="admin-success">{{ adminPasswordSuccess }}</div>
              <button type="submit" class="admin-button" :disabled="adminResettingPassword">
                {{ adminResettingPassword ? 'Сброс...' : 'Сбросить пароль' }}
              </button>
            </form>
          </div>
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
        <h3 class="top-games-title">Топ-100 лучших игр</h3>
        <div v-if="topGamesLoading" class="top-games-loading">Загрузка...</div>
        <div v-else-if="topGamesError" class="top-games-error">{{ topGamesError }}</div>
        <div v-else-if="!topGames || topGames.length === 0" class="top-games-empty">
          Пока нет игр с начисленным рейтингом
        </div>
        <div v-else class="top-games-list">
          <router-link
            v-for="(game, index) in topGames"
            :key="game.id"
            :to="`/game/details?id=${game.id}`"
            class="top-game-item top-game-item--link"
          >
            <div class="game-rank">#{{ index + 1 }}</div>
            <div class="game-info">
              <div class="game-field">
                <span class="game-field-size">{{ game.height }}×{{ game.width }}</span>
                <span class="game-mines">
                  <IconMine class="game-mines-icon" />
                  {{ game.mines }}
                </span>
              </div>
              <div class="game-details">
                <div class="game-time">
                  <IconClock class="game-time-icon" />
                  {{ formatTime(game.gameTime) }}
                </div>
                <div class="game-date">{{ formatDate(game.createdAt) }}</div>
              </div>
            </div>
            <div class="game-rating">
              <div v-if="game.rating > 0" class="rating-details">
                <div class="rating-gain">
                  +{{ Math.round(game.ratingContributed) }}
                </div>
                <div class="rating-label">засчитано</div>
                <div class="rating-percent">
                  {{ Math.round(game.ratingPercent) }}%
                </div>
                <div class="rating-base" v-if="game.ratingContributed !== game.rating">
                  (из {{ Math.round(game.rating) }})
                </div>
              </div>
              <div v-else class="rating-label" style="color: var(--text-secondary);">—</div>
            </div>
          </router-link>
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
          <router-link
            v-for="game in recentGames"
            :key="game.id"
            :to="`/game/details?id=${game.id}`"
            class="recent-game-item recent-game-item--link"
            :class="{ 'recent-game-item--lost': !game.won }"
          >
            <div class="game-main-info">
              <div class="game-field-info">
                <div class="game-field">
                  <span class="game-field-size">{{ game.height }}×{{ game.width }}</span>
                  <span class="game-mines">
                  <IconMine class="game-mines-icon" />
                  {{ game.mines }}
                </span>
                </div>
                <div class="game-complexity">
                  <span class="complexity-label">Сложность:</span>
                  <span class="complexity-value">{{ calculateDifficulty(game.width, game.height, game.mines).toFixed(2) }}</span>
                </div>
              </div>
              <div class="game-time-info">
                <div class="game-time">
                  <IconClock class="game-time-icon" />
                  {{ formatTime(game.gameTime) }}
                </div>
                <div class="game-date">{{ formatDate(game.createdAt) }}</div>
              </div>
            </div>
            <div class="game-rating-info" v-if="game.won">
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
          </router-link>
        </div>
      </div>
    </div>
  </main>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { getProfile, getProfileByUsername, updateColor, changePassword, getTopGames, getRecentGames, type UserProfile, type TopGame, type RecentGame } from '@/api/profile'
import { resetPasswordByAdmin } from '@/api/auth'
import { getErrorMessage } from '@/utils/errorHandler'
import { calculateDifficulty } from '@/utils/ratingCalculator'
import TextInput from '@/components/inputs/TextInput.vue'
import IconClock from '@/components/icons/IconClock.vue'
import IconSettings from '@/components/icons/IconSettings.vue'
import IconMine from '@/components/icons/IconMine.vue'

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
const currentPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const changingPassword = ref(false)
const passwordError = ref('')
const passwordSuccess = ref('')
const isAdmin = computed(() => authStore.user?.isAdmin || false)
const adminSearchType = ref<'username' | 'email'>('username')
const adminSearchValue = ref('')
const adminNewPassword = ref('')
const adminResettingPassword = ref(false)
const adminPasswordError = ref('')
const adminPasswordSuccess = ref('')

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

    // Определяем, чей профиль загружать
    const username = route.params.username as string
    isOwnProfile.value = !username

    // Загружаем профиль
    profile.value = username
      ? await getProfileByUsername(username)
      : await getProfile()

    // Устанавливаем выбранный цвет только для своего профиля
    if (isOwnProfile.value) {
      selectedColor.value = profile.value?.user.color || ''
    }

    // Загружаем игры (одинаковая логика для обоих случаев)
    await Promise.all([
      loadTopGames(username),
      loadRecentGames(username)
    ])
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

const handleChangePassword = async () => {
  passwordError.value = ''
  passwordSuccess.value = ''

  // Валидация
  if (!currentPassword.value || !newPassword.value || !confirmPassword.value) {
    passwordError.value = 'Все поля обязательны для заполнения'
    return
  }

  if (newPassword.value.length < 6) {
    passwordError.value = 'Новый пароль должен содержать минимум 6 символов'
    return
  }

  if (newPassword.value !== confirmPassword.value) {
    passwordError.value = 'Новые пароли не совпадают'
    return
  }

  changingPassword.value = true

  try {
    await changePassword(currentPassword.value, newPassword.value)
    passwordSuccess.value = 'Пароль успешно изменен'
    // Очищаем поля формы
    currentPassword.value = ''
    newPassword.value = ''
    confirmPassword.value = ''
    // Очищаем сообщение об успехе через 3 секунды
    setTimeout(() => {
      passwordSuccess.value = ''
    }, 3000)
  } catch (err: any) {
    passwordError.value = getErrorMessage(err, 'Ошибка смены пароля')
  } finally {
    changingPassword.value = false
  }
}

const handleAdminResetPassword = async () => {
  adminPasswordError.value = ''
  adminPasswordSuccess.value = ''

  // Валидация
  if (!adminSearchValue.value || !adminNewPassword.value) {
    adminPasswordError.value = 'Все поля обязательны для заполнения'
    return
  }

  if (adminNewPassword.value.length < 6) {
    adminPasswordError.value = 'Новый пароль должен содержать минимум 6 символов'
    return
  }

  adminResettingPassword.value = true

  try {
    const requestData: any = {
      newPassword: adminNewPassword.value
    }
    if (adminSearchType.value === 'username') {
      requestData.username = adminSearchValue.value
    } else {
      requestData.email = adminSearchValue.value
    }

    await resetPasswordByAdmin(requestData)
    adminPasswordSuccess.value = `Пароль успешно сброшен для пользователя ${adminSearchValue.value}`
    // Очищаем поля формы
    adminSearchValue.value = ''
    adminNewPassword.value = ''
    // Очищаем сообщение об успехе через 5 секунд
    setTimeout(() => {
      adminPasswordSuccess.value = ''
    }, 5000)
  } catch (err: any) {
    adminPasswordError.value = getErrorMessage(err, 'Ошибка сброса пароля')
  } finally {
    adminResettingPassword.value = false
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

/* Обеспечиваем контрастность текста в светлой теме */
[data-theme="light"] .user-rating {
  color: #ffffff;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.5), 0 2px 6px rgba(0, 0, 0, 0.3);
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%) !important;
}

.rating-label {
  font-size: 0.875rem;
  color: #ffffff;
  opacity: 1;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.3);
  font-weight: 600;
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

.change-password-section {
  margin-top: 2rem;
  padding-top: 2rem;
  border-top: 2px solid var(--border-color);
}

.change-password-title {
  margin: 0 0 1rem 0;
  font-size: 1.25rem;
  color: var(--text-primary);
}

.change-password-form {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.password-error {
  margin-top: 0.5rem;
  margin-bottom: 0.5rem;
  font-size: 0.875rem;
  color: #dc2626;
  padding: 0.75rem;
  background: rgba(220, 38, 38, 0.1);
  border-radius: 0.5rem;
}

.password-success {
  margin-top: 0.5rem;
  margin-bottom: 0.5rem;
  font-size: 0.875rem;
  color: #22c55e;
  padding: 0.75rem;
  background: rgba(34, 197, 94, 0.1);
  border-radius: 0.5rem;
}

.change-password-button {
  width: 100%;
  max-width: 300px;
  padding: 0.875rem 1.5rem;
  font-size: 1rem;
  font-weight: 600;
  color: #ffffff;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
  border-radius: 0.5rem;
  cursor: pointer;
  transition: all 0.2s ease-in-out;
  margin-top: 1rem;
}

.change-password-button:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
}

.change-password-button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.admin-section {
  margin-top: 2rem;
  padding-top: 2rem;
  border-top: 2px solid var(--border-color);
}

.admin-section-title {
  margin: 0 0 1rem 0;
  font-size: 1.25rem;
  color: var(--text-primary);
  font-weight: 700;
}

.admin-panel {
  background: var(--bg-secondary);
  padding: 1.5rem;
  border-radius: 0.75rem;
  border: 2px solid #f59e0b;
}

.admin-panel-title {
  margin: 0 0 1rem 0;
  font-size: 1rem;
  color: var(--text-primary);
  font-weight: 600;
}

.admin-form {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.admin-form-group {
  margin-bottom: 1rem;
}

.admin-label {
  display: block;
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 0.5rem;
}

.admin-radio-group {
  display: flex;
  gap: 1.5rem;
  margin-top: 0.5rem;
}

.admin-radio-label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
  font-size: 0.875rem;
  color: var(--text-primary);
}

.admin-radio {
  cursor: pointer;
}

.admin-error {
  margin-top: 0.5rem;
  margin-bottom: 0.5rem;
  font-size: 0.875rem;
  color: #dc2626;
  padding: 0.75rem;
  background: rgba(220, 38, 38, 0.1);
  border-radius: 0.5rem;
}

.admin-success {
  margin-top: 0.5rem;
  margin-bottom: 0.5rem;
  font-size: 0.875rem;
  color: #22c55e;
  padding: 0.75rem;
  background: rgba(34, 197, 94, 0.1);
  border-radius: 0.5rem;
}

.admin-button {
  width: 100%;
  max-width: 300px;
  padding: 0.875rem 1.5rem;
  font-size: 1rem;
  font-weight: 600;
  color: #ffffff;
  background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%);
  border: none;
  border-radius: 0.5rem;
  cursor: pointer;
  transition: all 0.2s ease-in-out;
  margin-top: 1rem;
}

.admin-button:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(245, 158, 11, 0.4);
}

.admin-button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
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

.top-game-item--link {
  text-decoration: none;
  color: inherit;
  cursor: pointer;
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
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
}

.game-mines-icon {
  width: 0.875rem;
  height: 0.875rem;
  flex-shrink: 0;
}

[data-theme="dark"] .game-mines-icon circle[fill="currentColor"] {
  fill: #dc2626;
}

[data-theme="dark"] .game-mines-icon path[stroke="#333"],
[data-theme="dark"] .game-mines-icon line[stroke="#333"],
[data-theme="dark"] .game-mines-icon circle[stroke="#333"],
[data-theme="dark"] .game-mines-icon circle[fill="#333"] {
  stroke: #fff;
  fill: #fff;
}

.game-time {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
}

.game-time-icon {
  width: 0.875rem;
  height: 0.875rem;
  flex-shrink: 0;
}

.admin-section-icon {
  width: 1.25rem;
  height: 1.25rem;
  display: inline-block;
  vertical-align: middle;
  margin-right: 0.5rem;
}

.game-details {
  display: flex;
  gap: 1.5rem;
  font-size: 0.875rem;
  color: var(--text-secondary);
}

.game-rating {
  text-align: right;
  min-width: 140px;
}

.rating-details {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
}

.rating-gain {
  font-size: 1.25rem;
  font-weight: 700;
  color: #22c55e;
  line-height: 1.2;
}

.rating-label {
  font-size: 0.75rem;
  color: var(--text-secondary);
  font-weight: 500;
  line-height: 1.2;
}

.rating-percent {
  font-size: 0.875rem;
  color: #667eea;
  font-weight: 600;
  line-height: 1.2;
}

.rating-base {
  font-size: 0.75rem;
  color: var(--text-secondary);
  font-weight: 400;
  line-height: 1.2;
  opacity: 0.7;
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

.recent-game-item--link {
  text-decoration: none;
  color: inherit;
  cursor: pointer;
}

.recent-game-item--lost {
  background: rgba(239, 68, 68, 0.1);
  border-left-color: #ef4444;
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

