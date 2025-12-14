<template>
  <main class="rating-page">
    <div class="rating-header">
      <h1 class="rating-title">🏆 Рейтинг игроков</h1>
      <p class="rating-subtitle">Топ игроков по рейтинговым очкам</p>
    </div>

    <div v-if="loading" class="loading">
      <p>Загрузка рейтинга...</p>
    </div>

    <div v-else-if="error" class="error">
      <p>{{ error }}</p>
      <button @click="loadLeaderboard" class="retry-button">Попробовать снова</button>
    </div>

    <div v-else-if="leaderboard.length === 0" class="empty">
      <p>Пока нет игроков в рейтинге</p>
    </div>

    <div v-else class="leaderboard">
      <div class="leaderboard-header">
        <div class="header-rank">#</div>
        <div class="header-player">Игрок</div>
        <div class="header-rating">Рейтинг</div>
        <div class="header-games">Игр</div>
      </div>

      <div
        v-for="(player, index) in leaderboard"
        :key="player.id"
        class="leaderboard-entry"
        :class="{
          'leaderboard-entry--first': index === 0,
          'leaderboard-entry--second': index === 1,
          'leaderboard-entry--third': index === 2,
          'leaderboard-entry--top': index < 3
        }"
      >
        <div class="entry-rank">
          <span v-if="index === 0" class="rank-icon">🥇</span>
          <span v-else-if="index === 1" class="rank-icon">🥈</span>
          <span v-else-if="index === 2" class="rank-icon">🥉</span>
          <span v-else class="rank-number">{{ index + 1 }}</span>
        </div>
        <div class="entry-player">
          <router-link :to="`/profile/${player.username}`" class="player-link">
            <div
              class="player-avatar"
              :style="player.color ? { background: player.color } : {}"
            >
              <span class="avatar-text">{{ player.username[0].toUpperCase() }}</span>
            </div>
            <span class="player-name">{{ player.username }}</span>
          </router-link>
        </div>
        <div class="entry-rating">
          <span class="rating-value">{{ Math.round(player.rating) }}</span>
        </div>
        <div class="entry-games">
          <span class="games-count">{{ player.gamesPlayed }}</span>
        </div>
      </div>
    </div>
  </main>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getLeaderboard, type LeaderboardEntry } from '@/api/profile'
import { getErrorMessage } from '@/utils/errorHandler'

const leaderboard = ref<LeaderboardEntry[]>([])
const loading = ref(true)
const error = ref<string | null>(null)

const loadLeaderboard = async () => {
  loading.value = true
  error.value = null
  try {
    leaderboard.value = await getLeaderboard()
  } catch (err: any) {
    error.value = getErrorMessage(err, 'Ошибка загрузки рейтинга')
    console.error('Error loading leaderboard:', err)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadLeaderboard()
})
</script>

<style scoped>
.rating-page {
  max-width: 900px;
  margin: 0 auto;
  padding: 2rem;
}

.rating-header {
  text-align: center;
  margin-bottom: 2rem;
}

.rating-title {
  font-size: 2.5rem;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0 0 0.5rem 0;
}

.rating-subtitle {
  font-size: 1.125rem;
  color: var(--text-secondary);
  margin: 0;
}

.loading,
.error,
.empty {
  text-align: center;
  padding: 3rem;
  color: var(--text-secondary);
}

.error {
  color: #ef4444;
}

.retry-button {
  margin-top: 1rem;
  padding: 0.75rem 1.5rem;
  background: #667eea;
  color: white;
  border: none;
  border-radius: 0.5rem;
  cursor: pointer;
  font-weight: 600;
  transition: all 0.2s;
}

.retry-button:hover {
  background: #5568d3;
  transform: translateY(-2px);
}

.leaderboard {
  background: var(--bg-primary);
  border-radius: 1rem;
  box-shadow: 0 4px 12px var(--shadow);
  overflow: hidden;
}

.leaderboard-header {
  display: grid;
  grid-template-columns: 60px 1fr 120px 100px;
  gap: 1rem;
  padding: 1rem 1.5rem;
  background: var(--bg-secondary);
  border-bottom: 2px solid var(--border-color);
  font-weight: 600;
  color: var(--text-secondary);
  font-size: 0.875rem;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.header-rank {
  text-align: center;
}

.header-player {
  text-align: left;
}

.header-rating,
.header-games {
  text-align: center;
}

.leaderboard-entry {
  display: grid;
  grid-template-columns: 60px 1fr 120px 100px;
  gap: 1rem;
  padding: 1rem 1.5rem;
  border-bottom: 1px solid var(--border-color);
  transition: background 0.2s;
  align-items: center;
}

.leaderboard-entry:hover {
  background: var(--bg-secondary);
}

.leaderboard-entry--top {
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.05) 0%, rgba(118, 75, 162, 0.05) 100%);
}

.leaderboard-entry--top:hover {
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.1) 0%, rgba(118, 75, 162, 0.1) 100%);
}

/* Первое место - золото */
.leaderboard-entry--first {
  background: linear-gradient(135deg, rgba(255, 215, 0, 0.15) 0%, rgba(255, 193, 7, 0.15) 100%);
  border-left: 4px solid #ffd700;
  box-shadow: 0 4px 16px rgba(255, 215, 0, 0.3);
  position: relative;
  transform: scale(1.02);
  z-index: 3;
}

.leaderboard-entry--first::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 3px;
  background: linear-gradient(90deg, #ffd700 0%, #ffed4e 50%, #ffd700 100%);
}

.leaderboard-entry--first:hover {
  background: linear-gradient(135deg, rgba(255, 215, 0, 0.25) 0%, rgba(255, 193, 7, 0.25) 100%);
  box-shadow: 0 6px 20px rgba(255, 215, 0, 0.4);
  transform: scale(1.03);
}

.leaderboard-entry--first .entry-rank {
  color: #ffd700;
}

.leaderboard-entry--first .rating-value {
  color: #ffd700;
  text-shadow: 0 0 10px rgba(255, 215, 0, 0.5);
  font-size: 1.25rem;
}

.leaderboard-entry--first .player-avatar {
  box-shadow: 0 0 15px rgba(255, 215, 0, 0.6);
  border: 2px solid #ffd700;
}

/* Второе место - серебро */
.leaderboard-entry--second {
  background: linear-gradient(135deg, rgba(192, 192, 192, 0.15) 0%, rgba(169, 169, 169, 0.15) 100%);
  border-left: 4px solid #c0c0c0;
  box-shadow: 0 3px 12px rgba(192, 192, 192, 0.3);
  position: relative;
  transform: scale(1.01);
  z-index: 2;
}

.leaderboard-entry--second::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 3px;
  background: linear-gradient(90deg, #c0c0c0 0%, #e8e8e8 50%, #c0c0c0 100%);
}

.leaderboard-entry--second:hover {
  background: linear-gradient(135deg, rgba(192, 192, 192, 0.25) 0%, rgba(169, 169, 169, 0.25) 100%);
  box-shadow: 0 5px 16px rgba(192, 192, 192, 0.4);
  transform: scale(1.02);
}

.leaderboard-entry--second .entry-rank {
  color: #c0c0c0;
}

.leaderboard-entry--second .rating-value {
  color: #c0c0c0;
  text-shadow: 0 0 8px rgba(192, 192, 192, 0.4);
  font-size: 1.1875rem;
}

.leaderboard-entry--second .player-avatar {
  box-shadow: 0 0 12px rgba(192, 192, 192, 0.5);
  border: 2px solid #c0c0c0;
}

/* Третье место - бронза */
.leaderboard-entry--third {
  background: linear-gradient(135deg, rgba(205, 127, 50, 0.15) 0%, rgba(184, 115, 51, 0.15) 100%);
  border-left: 4px solid #cd7f32;
  box-shadow: 0 2px 10px rgba(205, 127, 50, 0.3);
  position: relative;
  z-index: 1;
}

.leaderboard-entry--third::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 3px;
  background: linear-gradient(90deg, #cd7f32 0%, #daa520 50%, #cd7f32 100%);
}

.leaderboard-entry--third:hover {
  background: linear-gradient(135deg, rgba(205, 127, 50, 0.25) 0%, rgba(184, 115, 51, 0.25) 100%);
  box-shadow: 0 4px 14px rgba(205, 127, 50, 0.4);
  transform: scale(1.01);
}

.leaderboard-entry--third .entry-rank {
  color: #cd7f32;
}

.leaderboard-entry--third .rating-value {
  color: #cd7f32;
  text-shadow: 0 0 8px rgba(205, 127, 50, 0.4);
  font-size: 1.125rem;
}

.leaderboard-entry--third .player-avatar {
  box-shadow: 0 0 10px rgba(205, 127, 50, 0.5);
  border: 2px solid #cd7f32;
}

/* Адаптация для темной темы */
[data-theme="dark"] .leaderboard-entry--first {
  background: linear-gradient(135deg, rgba(255, 215, 0, 0.2) 0%, rgba(255, 193, 7, 0.2) 100%);
}

[data-theme="dark"] .leaderboard-entry--second {
  background: linear-gradient(135deg, rgba(192, 192, 192, 0.2) 0%, rgba(169, 169, 169, 0.2) 100%);
}

[data-theme="dark"] .leaderboard-entry--third {
  background: linear-gradient(135deg, rgba(205, 127, 50, 0.2) 0%, rgba(184, 115, 51, 0.2) 100%);
}

.entry-rank {
  text-align: center;
  font-weight: 600;
  color: var(--text-secondary);
}

.rank-icon {
  font-size: 1.5rem;
}

.rank-number {
  font-size: 1.125rem;
}

.entry-player {
  display: flex;
  align-items: center;
}

.player-link {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  text-decoration: none;
  color: var(--text-primary);
  transition: color 0.2s;
}

.player-link:hover {
  color: #667eea;
}

.player-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #667eea;
  color: white;
  font-weight: 700;
  font-size: 1.125rem;
  flex-shrink: 0;
}

.avatar-text {
  user-select: none;
}

.player-name {
  font-weight: 600;
  font-size: 1rem;
}

.entry-rating,
.entry-games {
  text-align: center;
}

.rating-value {
  font-weight: 700;
  font-size: 1.125rem;
  color: var(--text-primary);
}

.games-count {
  color: var(--text-secondary);
  font-weight: 500;
}

@media (max-width: 768px) {
  .rating-page {
    padding: 1rem;
  }

  .rating-title {
    font-size: 2rem;
  }

  .leaderboard-header,
  .leaderboard-entry {
    grid-template-columns: 50px 1fr 80px 60px;
    gap: 0.5rem;
    padding: 0.75rem 1rem;
  }

  .player-avatar {
    width: 32px;
    height: 32px;
    font-size: 0.875rem;
  }

  .player-name {
    font-size: 0.875rem;
  }

  .rating-value {
    font-size: 1rem;
  }

  .header-rank,
  .header-rating,
  .header-games {
    font-size: 0.75rem;
  }

  /* Уменьшаем трансформации на мобильных */
  .leaderboard-entry--first {
    transform: scale(1.01);
  }

  .leaderboard-entry--first:hover {
    transform: scale(1.02);
  }

  .leaderboard-entry--second {
    transform: scale(1.005);
  }

  .leaderboard-entry--second:hover {
    transform: scale(1.01);
  }
}

@media (max-width: 480px) {
  .leaderboard-header,
  .leaderboard-entry {
    grid-template-columns: 40px 1fr 70px 50px;
    gap: 0.5rem;
    padding: 0.5rem;
  }

  .player-avatar {
    width: 28px;
    height: 28px;
    font-size: 0.75rem;
  }

  .player-name {
    font-size: 0.8125rem;
  }

  .rating-value {
    font-size: 0.875rem;
  }
}
</style>

