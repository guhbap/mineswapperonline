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
        <h1 class="profile-title">Профиль пользователя</h1>
        <div class="profile-user-info">
          <div class="user-avatar">
            <span class="avatar-text">{{ profile.user.username[0].toUpperCase() }}</span>
          </div>
          <div class="user-details">
            <h2 class="username">{{ profile.user.username }}</h2>
            <p class="email">{{ profile.user.email }}</p>
            <div class="status-badge" :class="{ 'status-online': profile.stats.isOnline, 'status-offline': !profile.stats.isOnline }">
              <span class="status-dot"></span>
              <span>{{ profile.stats.isOnline ? 'В сети' : 'Не в сети' }}</span>
            </div>
          </div>
        </div>
      </div>

      <div class="profile-stats">
        <h3 class="stats-title">Статистика игр</h3>
        <div class="stats-grid">
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
      </div>
    </div>
  </main>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { getProfile, type UserProfile } from '@/api/profile'

const profile = ref<UserProfile | null>(null)
const loading = ref(true)
const error = ref('')

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

const loadProfile = async () => {
  try {
    loading.value = true
    error.value = ''
    profile.value = await getProfile()
  } catch (err: any) {
    error.value = err.response?.data || 'Ошибка загрузки профиля'
    console.error('Ошибка загрузки профиля:', err)
  } finally {
    loading.value = false
  }
}

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
</style>

