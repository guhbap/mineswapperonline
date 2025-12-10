<template>
  <div class="rooms-list">
    <div class="rooms-header">
      <h2 class="rooms-title">Комнаты</h2>
      <button @click="$emit('create')" class="btn-create">
        + Создать комнату
      </button>
    </div>

    <div v-if="loading" class="rooms-loading">Загрузка...</div>
    <div v-else-if="rooms.length === 0" class="rooms-empty">
      Нет доступных комнат. Создайте первую!
    </div>
    <div v-else class="rooms-grid">
      <div
        v-for="room in rooms"
        :key="room.id"
        class="room-card"
        @click="handleRoomClick(room)"
      >
        <div class="room-card__header">
          <h3 class="room-card__name">{{ room.name }}</h3>
          <span v-if="room.hasPassword" class="room-card__lock">🔒</span>
        </div>
        <div class="room-card__info">
          <div class="room-info-item">
            <span class="room-info-label">Поле:</span>
            <span class="room-info-value">{{ room.rows }}×{{ room.cols }}</span>
          </div>
          <div class="room-info-item">
            <span class="room-info-label">Мин:</span>
            <span class="room-info-value">{{ room.mines }}</span>
          </div>
          <div class="room-info-item">
            <span class="room-info-label">Игроков:</span>
            <span class="room-info-value">{{ room.players }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import type { Room } from '@/api/rooms'
import { getRooms } from '@/api/rooms'

const emit = defineEmits<{
  create: []
  join: [room: Room]
}>()

const rooms = ref<Room[]>([])
const loading = ref(true)
let refreshInterval: number | null = null

const loadRooms = async () => {
  try {
    rooms.value = await getRooms()
  } catch (error) {
    console.error('Ошибка загрузки комнат:', error)
  } finally {
    loading.value = false
  }
}

const handleRoomClick = (room: Room) => {
  emit('join', room)
}

onMounted(() => {
  loadRooms()
  // Обновляем список каждые 3 секунды
  refreshInterval = setInterval(loadRooms, 3000) as unknown as number
})

onUnmounted(() => {
  if (refreshInterval) {
    clearInterval(refreshInterval)
  }
})
</script>

<style scoped>
.rooms-list {
  padding: 2rem;
  max-width: 1200px;
  margin: 0 auto;
}

.rooms-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
}

.rooms-title {
  font-size: 2rem;
  color: var(--text-primary);
  margin: 0;
}

.btn-create {
  padding: 0.75rem 1.5rem;
  font-size: 1rem;
  font-weight: 600;
  color: white;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
  border-radius: 0.5rem;
  cursor: pointer;
  transition: transform 0.2s, box-shadow 0.2s;
}

.btn-create:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
}

.rooms-loading,
.rooms-empty {
  text-align: center;
  padding: 3rem;
  color: var(--text-secondary);
  font-size: 1.125rem;
}

.rooms-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 1.5rem;
}

.room-card {
  background: var(--bg-primary);
  border: 2px solid var(--border-color);
  border-radius: 0.75rem;
  padding: 1.5rem;
  cursor: pointer;
  transition: all 0.2s;
  box-shadow: 0 2px 8px var(--shadow);
}

.room-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 4px 16px var(--shadow-hover);
  border-color: #667eea;
}

.room-card__header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.room-card__name {
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0;
}

.room-card__lock {
  font-size: 1.25rem;
}

.room-card__info {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.room-info-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.room-info-label {
  color: var(--text-secondary);
  font-size: 0.875rem;
}

.room-info-value {
  color: var(--text-primary);
  font-weight: 600;
}
</style>

