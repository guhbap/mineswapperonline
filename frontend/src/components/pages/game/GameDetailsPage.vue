<template>
  <main class="game-details-page">
    <div v-if="loading" class="loading">
      <p>Загрузка информации об игре...</p>
    </div>
    <div v-else-if="error" class="error">
      <p>{{ error }}</p>
      <router-link to="/profile" class="back-link">← Вернуться к профилю</router-link>
    </div>
    <div v-else-if="gameDetails" class="game-details-content">
      <div class="game-details-header">
        <h1 class="game-details-title">Результаты игры</h1>
        <router-link to="/profile" class="back-link">← Вернуться к профилю</router-link>
      </div>

      <div class="game-details-grid">
        <!-- Основная информация -->
        <div class="detail-card detail-card--main">
          <h2 class="detail-card-title">Основная информация</h2>
          <div class="detail-item">
            <span class="detail-label">Результат:</span>
            <span class="detail-value" :class="{ 'detail-value--won': gameDetails.won, 'detail-value--lost': !gameDetails.won }">
              {{ gameDetails.won ? '🏆 Победа' : '💥 Поражение' }}
            </span>
          </div>
          <div class="detail-item">
            <span class="detail-label">Рейтинг:</span>
            <span class="detail-value detail-value--rating">
              {{ gameDetails.rating > 0 ? `+${Math.round(gameDetails.rating)}` : '—' }}
            </span>
          </div>
          <div class="detail-item">
            <span class="detail-label">Время начала:</span>
            <span class="detail-value">{{ formatDateTime(gameDetails.startTime) }}</span>
          </div>
          <div class="detail-item">
            <span class="detail-label">Длительность:</span>
            <span class="detail-value">{{ formatDuration(gameDetails.duration) }}</span>
          </div>
        </div>

        <!-- Параметры поля -->
        <div class="detail-card">
          <h2 class="detail-card-title">Параметры поля</h2>
          <div class="detail-item">
            <span class="detail-label">Размер:</span>
            <span class="detail-value">{{ gameDetails.width }} × {{ gameDetails.height }}</span>
          </div>
          <div class="detail-item">
            <span class="detail-label">Количество мин:</span>
            <span class="detail-value">{{ gameDetails.mines }}</span>
          </div>
          <div class="detail-item">
            <span class="detail-label">Плотность:</span>
            <span class="detail-value">{{ calculateDensity() }}%</span>
          </div>
          <div class="detail-item">
            <span class="detail-label">Seed:</span>
            <span class="detail-value detail-value--seed">{{ gameDetails.seed }}</span>
          </div>
          <div v-if="gameDetails.hasCustomSeed" class="detail-item">
            <span class="detail-label">Тип seed:</span>
            <span class="detail-value detail-value--custom">Пользовательский (нерейтинговая игра)</span>
          </div>
        </div>

        <!-- Параметры создания -->
        <div class="detail-card">
          <h2 class="detail-card-title">Параметры создания</h2>
          <div class="detail-item">
            <span class="detail-label">Создатель:</span>
            <span class="detail-value">{{ gameDetails.creatorName }}</span>
          </div>
          <div class="detail-item">
            <span class="detail-label">Room ID:</span>
            <span class="detail-value detail-value--room-id">{{ gameDetails.roomId }}</span>
          </div>
          <div class="detail-item">
            <span class="detail-label">Quick Start:</span>
            <span class="detail-value" :class="{ 'detail-value--enabled': gameDetails.quickStart }">
              {{ gameDetails.quickStart ? '✓ Включен' : '✗ Выключен' }}
            </span>
          </div>
          <div class="detail-item">
            <span class="detail-label">Chording:</span>
            <span class="detail-value" :class="{ 'detail-value--enabled': gameDetails.chording }">
              {{ gameDetails.chording ? '✓ Включен' : '✗ Выключен' }}
            </span>
          </div>
        </div>

        <!-- Участники -->
        <div class="detail-card detail-card--participants">
          <h2 class="detail-card-title">Участники игры</h2>
          <div v-if="gameDetails.participants.length === 0" class="no-participants">
            <p>Нет информации об участниках</p>
          </div>
          <div v-else class="participants-list">
            <div
              v-for="participant in gameDetails.participants"
              :key="participant.userId"
              class="participant-item"
            >
              <div
                class="participant-avatar"
                :style="participant.color ? { background: participant.color } : {}"
              >
                {{ participant.username[0].toUpperCase() }}
              </div>
              <div class="participant-info">
                <div class="participant-username">{{ participant.username }}</div>
                <div v-if="participant.nickname !== participant.username" class="participant-nickname">
                  {{ participant.nickname }}
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </main>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { getGameDetails, type GameDetails } from '@/api/profile'

const route = useRoute()
const loading = ref(true)
const error = ref<string | null>(null)
const gameDetails = ref<GameDetails | null>(null)

onMounted(async () => {
  const gameId = route.query.id
  if (!gameId) {
    error.value = 'ID игры не указан'
    loading.value = false
    return
  }

  try {
    const id = typeof gameId === 'string' ? parseInt(gameId, 10) : Number(gameId)
    if (isNaN(id)) {
      error.value = 'Неверный ID игры'
      loading.value = false
      return
    }

    gameDetails.value = await getGameDetails(id)
  } catch (err: any) {
    console.error('Error loading game details:', err)
    if (err.response?.status === 404) {
      error.value = 'Игра не найдена'
    } else {
      error.value = 'Ошибка при загрузке информации об игре'
    }
  } finally {
    loading.value = false
  }
})

function formatDateTime(dateString: string): string {
  const date = new Date(dateString)
  return date.toLocaleString('ru-RU', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

function formatDuration(seconds: number): string {
  const minutes = Math.floor(seconds / 60)
  const secs = Math.floor(seconds % 60)
  const ms = Math.floor((seconds % 1) * 100)
  
  if (minutes > 0) {
    return `${minutes}м ${secs}.${ms.toString().padStart(2, '0')}с`
  }
  return `${secs}.${ms.toString().padStart(2, '0')}с`
}

function calculateDensity(): number {
  if (!gameDetails.value) return 0
  const totalCells = gameDetails.value.width * gameDetails.value.height
  return Math.round((gameDetails.value.mines / totalCells) * 100 * 100) / 100
}
</script>

<style scoped>
.game-details-page {
  max-width: 1200px;
  margin: 0 auto;
  padding: 2rem;
}

.game-details-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
}

.game-details-title {
  font-size: 2rem;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0;
}

.back-link {
  color: var(--text-secondary);
  text-decoration: none;
  font-weight: 500;
  transition: color 0.2s ease;
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
}

.back-link:hover {
  color: #667eea;
}

.loading,
.error {
  text-align: center;
  padding: 3rem;
  color: var(--text-secondary);
}

.error {
  color: #dc2626;
}

.game-details-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 1.5rem;
}

.detail-card {
  background: var(--bg-primary);
  padding: 1.5rem;
  border-radius: 1rem;
  box-shadow: 0 2px 8px var(--shadow);
}

.detail-card--main {
  grid-column: 1 / -1;
}

.detail-card--participants {
  grid-column: 1 / -1;
}

.detail-card-title {
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0 0 1.5rem 0;
  padding-bottom: 0.75rem;
  border-bottom: 2px solid var(--border-color);
}

.detail-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem 0;
  border-bottom: 1px solid var(--border-color);
}

.detail-item:last-child {
  border-bottom: none;
}

.detail-label {
  font-weight: 600;
  color: var(--text-secondary);
  font-size: 0.875rem;
}

.detail-value {
  font-weight: 500;
  color: var(--text-primary);
  text-align: right;
  word-break: break-word;
}

.detail-value--won {
  color: #22c55e;
  font-weight: 700;
}

.detail-value--lost {
  color: #dc2626;
  font-weight: 700;
}

.detail-value--rating {
  color: #667eea;
  font-weight: 700;
  font-size: 1.125rem;
}

.detail-value--seed {
  font-family: 'Courier New', monospace;
  font-size: 0.875rem;
  color: var(--text-secondary);
  background: var(--bg-secondary);
  padding: 0.25rem 0.5rem;
  border-radius: 0.25rem;
}

.detail-value--custom {
  color: #f59e0b;
  font-weight: 600;
}

.detail-value--room-id {
  font-family: 'Courier New', monospace;
  font-size: 0.875rem;
  color: var(--text-secondary);
}

.detail-value--enabled {
  color: #22c55e;
  font-weight: 600;
}

.participants-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.participant-item {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1rem;
  background: var(--bg-secondary);
  border-radius: 0.75rem;
  transition: transform 0.2s, box-shadow 0.2s;
}

.participant-item:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px var(--shadow);
}

.participant-avatar {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 1.25rem;
  color: white;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  flex-shrink: 0;
}

.participant-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.participant-username {
  font-weight: 600;
  color: var(--text-primary);
  font-size: 1rem;
}

.participant-nickname {
  font-size: 0.875rem;
  color: var(--text-secondary);
  font-style: italic;
}

.no-participants {
  text-align: center;
  padding: 2rem;
  color: var(--text-secondary);
}

@media (max-width: 768px) {
  .game-details-page {
    padding: 1rem;
  }

  .game-details-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 1rem;
  }

  .game-details-grid {
    grid-template-columns: 1fr;
  }

  .detail-card--main,
  .detail-card--participants {
    grid-column: 1;
  }

  .detail-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 0.5rem;
  }

  .detail-value {
    text-align: left;
  }
}
</style>

