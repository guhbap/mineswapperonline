<template>
  <main class="main-page">
    <!-- Выбор комнаты -->
    <section v-if="!selectedRoom" class="rooms-wrapper" aria-label="Выбор игровой комнаты">
      <!-- Описание игры -->
      <div v-show="isDescriptionVisible" class="game-description">
        <h1 class="game-description__title">
          <IconGamepad class="game-description__title-icon" />
          Сапер Онлайн
        </h1>
        <p class="game-description__text">
          Классическая игра Сапер в многопользовательском режиме! Играйте вместе с друзьями в реальном времени,
          соревнуйтесь за лучший результат и получайте рейтинг за успешные игры.
        </p>
        <div class="game-description__features">
          <div class="feature-item">
            <IconUsers class="feature-icon" />
            <span class="feature-text">Играйте с друзьями в реальном времени</span>
          </div>
          <div class="feature-item">
            <IconTrophy class="feature-icon" />
            <span class="feature-text">Получайте рейтинг за успешные игры</span>
          </div>
          <div class="feature-item">
            <IconChat class="feature-icon" />
            <span class="feature-text">Общайтесь в чате во время игры</span>
          </div>
          <div class="feature-item">
            <IconSettings class="feature-icon" />
            <span class="feature-text">Настраивайте размер поля и сложность</span>
          </div>
        </div>
        <div class="game-description__actions">
          <router-link to="/faq" class="game-description__link">
            <IconBook class="faq-link-icon" />
            Часто задаваемые вопросы
          </router-link>
          <a @click.prevent="toggleDescription" class="game-description__link game-description__link--toggle">
            {{ isDescriptionVisible ? 'Скрыть описание' : 'Показать описание' }}
          </a>
        </div>
      </div>
      <div v-if="!isDescriptionVisible" class="game-description-toggle-wrapper">
        <!-- <a @click.prevent="toggleDescription" class="game-description__link game-description__link--show">
          Показать описание игры
        </a> -->
      </div>
      <RoomsList
        @create="showCreateModal = true"
        @join="handleRoomSelect"
      />
      <CreateRoomModal
        v-if="showCreateModal"
        :show="showCreateModal"
        @submit="handleCreateRoom"
        @cancel="showCreateModal = false"
      />
      <JoinRoomModal
        v-if="selectedRoomForJoin"
        :show="!!selectedRoomForJoin"
        :room="selectedRoomForJoin"
        @submit="handleJoinRoom"
        @cancel="handleCancelJoinRoom"
      />
    </section>

    <!-- Ввод никнейма (только для гостей) -->
    <NicknameModal
      v-if="shouldShowNicknameModal"
      :show="shouldShowNicknameModal"
      @submit="handleNicknameSubmit"
    />

    <!-- Игра -->
    <section v-if="getNickname && selectedRoom" class="game-wrapper" aria-label="Игровое поле Сапера">
      <MinesweeperGame
        :ws-client="wsClient"
        :nickname="getNickname"
        :room="selectedRoom"
        @edit-room="showEditModal = true"
      />
      <EditRoomModal
        v-if="showEditModal && selectedRoom"
        :show="showEditModal"
        :room="selectedRoom"
        @submit="handleUpdateRoom"
        @cancel="showEditModal = false"
      />
    </section>
  </main>
</template>

<script setup lang="ts">
import { ref, onUnmounted, computed, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import NicknameModal from '@/components/NicknameModal.vue'
import MinesweeperGame from '@/components/MinesweeperGame.vue'
import RoomsList from '@/components/RoomsList.vue'
import CreateRoomModal from '@/components/CreateRoomModal.vue'
import JoinRoomModal from '@/components/JoinRoomModal.vue'
import EditRoomModal from '@/components/EditRoomModal.vue'
import { WebSocketClient, type WebSocketMessage, type IWebSocketClient } from '@/api/websocket'
import { createRoom, getRooms, type Room } from '@/api/rooms'
import IconGamepad from '@/components/icons/IconGamepad.vue'
import IconUsers from '@/components/icons/IconUsers.vue'
import IconTrophy from '@/components/icons/IconTrophy.vue'
import IconChat from '@/components/icons/IconChat.vue'
import IconSettings from '@/components/icons/IconSettings.vue'
import IconBook from '@/components/icons/IconBook.vue'

const router = useRouter()
const route = useRoute()

const authStore = useAuthStore()
const nickname = ref('')
const wsClient = ref<IWebSocketClient | null>(null)
const selectedRoom = ref<Room | null>(null)
const selectedRoomForJoin = ref<Room | null>(null)
const showCreateModal = ref(false)
const showEditModal = ref(false)

// Состояние видимости описания игры (загружаем из localStorage)
const savedDescriptionState = localStorage.getItem('gameDescriptionVisible')
const isDescriptionVisible = ref(savedDescriptionState !== null ? savedDescriptionState === 'true' : true)

// Функция для переключения видимости описания
const toggleDescription = () => {
  isDescriptionVisible.value = !isDescriptionVisible.value
  localStorage.setItem('gameDescriptionVisible', String(isDescriptionVisible.value))
}

// Функция для сброса состояния игры и возврата к выбору комнаты
const resetToRoomSelection = () => {
  // Отключаем WebSocket
  if (wsClient.value) {
    wsClient.value.disconnect()
    wsClient.value = null
  }

  // Сбрасываем выбранную комнату
  selectedRoom.value = null
  selectedRoomForJoin.value = null

  // Сбрасываем никнейм только для гостей
  if (!authStore.isAuthenticated) {
    nickname.value = ''
  }

  // Обновляем URL на корень
  if (route.path.startsWith('/room/')) {
    router.replace('/')
  }
}

// Слушаем событие для сброса игры
const handleResetGame = () => {
  resetToRoomSelection()
}

// Функция для загрузки комнаты по ID из URL
const loadRoomFromUrl = async () => {
  const roomId = route.params.id as string
  if (!roomId) return

  try {
    // Получаем список комнат и ищем нужную
    const rooms = await getRooms()
    const room = rooms.find(r => r.id === roomId)

    if (room) {
      // Открываем модалку подключения к комнате
      selectedRoomForJoin.value = room
    } else {
      // Комната не найдена, перенаправляем на главную
      router.replace('/')
    }
  } catch (error) {
    console.error('Ошибка загрузки комнаты:', error)
    router.replace('/')
  }
}

// Отслеживаем изменения роута
watch(() => route.params.id, (newId, oldId) => {
  if (newId && !selectedRoom.value && newId !== oldId) {
    // Если есть ID в URL, но мы не в комнате - загружаем комнату
    loadRoomFromUrl()
  } else if (!newId && selectedRoom.value) {
    // Если убрали ID из URL, но мы в комнате - выходим
    resetToRoomSelection()
  }
})

onMounted(() => {
  window.addEventListener('reset-game', handleResetGame)

  // Проверяем, есть ли ID комнаты в URL при загрузке
  if (route.params.id && !selectedRoom.value) {
    loadRoomFromUrl()
  }
})

onUnmounted(() => {
  window.removeEventListener('reset-game', handleResetGame)
  wsClient.value?.disconnect()
})

// Определяем, нужно ли показывать модалку никнейма
const shouldShowNicknameModal = computed(() => {
  return selectedRoom.value && !nickname.value && !authStore.isAuthenticated
})

// Автоматически используем username авторизованного пользователя
const getNickname = computed(() => {
  if (authStore.isAuthenticated && authStore.user) {
    return authStore.user.username
  }
  return nickname.value
})

const handleRoomSelect = (room: Room) => {
  selectedRoomForJoin.value = room
}

const handleCancelJoinRoom = () => {
  selectedRoomForJoin.value = null
  // Если мы пришли по ссылке с ID комнаты, возвращаемся на главную
  if (route.path.startsWith('/room/')) {
    router.replace('/')
  }
}

const handleJoinRoom = (room: Room) => {
  selectedRoom.value = room
  selectedRoomForJoin.value = null

  // Обновляем URL на /room/:id
  router.replace(`/room/${room.id}`)

  // Если пользователь авторизован, автоматически подключаемся
  if (authStore.isAuthenticated && authStore.user) {
    connectToRoom(authStore.user.username)
  }
  // Если гость - показываем модалку для ввода никнейма
}

const handleCreateRoom = async (data: { name: string; password?: string; rows: number; cols: number; mines: number; gameMode: string; quickStart: boolean; chording: boolean; seed?: string | null }) => {
  try {
    const room = await createRoom(data)
    selectedRoom.value = room
    showCreateModal.value = false

    // Обновляем URL на /room/:id
    router.replace(`/room/${room.id}`)

    // Если пользователь авторизован, автоматически подключаемся
    if (authStore.isAuthenticated && authStore.user) {
      connectToRoom(authStore.user.username)
    }
    // Если гость - показываем модалку для ввода никнейма
  } catch (error) {
    console.error('Ошибка создания комнаты:', error)
    // Ошибка обрабатывается в CreateRoomModal
  }
}

const handleNicknameSubmit = (submittedNickname: string) => {
  nickname.value = submittedNickname
  connectToRoom(submittedNickname)
}

const handleUpdateRoom = async (updatedRoom: Room) => {
  selectedRoom.value = updatedRoom
  showEditModal.value = false

  // Переподключаемся к WebSocket, чтобы получить обновленное состояние
  if (wsClient.value) {
    wsClient.value.disconnect()
    wsClient.value = null
  }

  // Если пользователь авторизован, автоматически подключаемся
  if (authStore.isAuthenticated && authStore.user) {
    connectToRoom(authStore.user.username)
  } else if (nickname.value) {
    connectToRoom(nickname.value)
  }
}

const connectToRoom = (playerNickname: string) => {
  if (!selectedRoom.value) return

  // Создаем WebSocket соединение с room ID
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const host = import.meta.env.DEV
    ? 'localhost:8080'
    : window.location.host

  // Добавляем userID в URL, если пользователь авторизован
  let wsUrl = `${protocol}//${host}/api/ws?room=${selectedRoom.value.id}`
  if (authStore.isAuthenticated && authStore.user?.id) {
    wsUrl += `&userId=${authStore.user.id}`
  }

  wsClient.value = new WebSocketClient(
    wsUrl,
      (msg: WebSocketMessage) => {
      // Обработка сообщений будет в компоненте игры через событие
      const event = new CustomEvent('ws-message', { detail: msg })
      window.dispatchEvent(event)
    },
    () => {
      // Отправляем никнейм после подключения
      if (wsClient.value) {
        wsClient.value.sendNickname(playerNickname)
      }
    },
    () => {
      // WebSocket отключен
    },
    (error) => {
      console.error('Ошибка WebSocket:', error)
    }
  )

  wsClient.value.connect()
}
</script>

<style scoped>
.main-page {
  min-height: 100vh;
  width: 100%;
}

.game-wrapper {
  width: 100%;
  min-height: 100vh;
  overflow-x: hidden;
}

.game-description {
  max-width: 1200px;
  margin: 0 auto;
  padding: 2rem;
  background: var(--bg-secondary);
  border-radius: 1rem;
  box-shadow: 0 4px 12px var(--shadow);
  margin-bottom: 2rem;
  animation: slideDown 0.3s ease-out;
}

.game-description-toggle-wrapper {
  max-width: 1200px;
  margin: 0 auto;
  margin-bottom: 2rem;
  text-align: center;
}

@keyframes slideDown {
  from {
    opacity: 0;
    transform: translateY(-10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.game-description__title {
  font-size: 2.5rem;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0 0 1rem 0;
  text-align: center;
}

.game-description__text {
  font-size: 1.125rem;
  color: var(--text-secondary);
  text-align: center;
  margin: 0 0 1.5rem 0;
  line-height: 1.6;
}

.game-description__features {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.feature-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem;
  background: var(--bg-primary);
  border-radius: 0.5rem;
  transition: transform 0.2s;
}

.feature-item:hover {
  transform: translateY(-2px);
}

.feature-icon {
  width: 1.5rem;
  height: 1.5rem;
  flex-shrink: 0;
}

.game-description__title-icon {
  width: 2.5rem;
  height: 2.5rem;
  display: inline-block;
  vertical-align: middle;
  margin-right: 0.5rem;
}

.faq-link-icon {
  width: 1rem;
  height: 1rem;
  display: inline-block;
  vertical-align: middle;
  margin-right: 0.5rem;
}

.feature-text {
  color: var(--text-primary);
  font-size: 0.875rem;
  font-weight: 500;
}

.game-description__actions {
  display: flex;
  justify-content: center;
  gap: 1rem;
  flex-wrap: wrap;
}

.game-description__link {
  display: inline-block;
  padding: 0.75rem 1.5rem;
  color: #667eea;
  text-decoration: none;
  font-weight: 600;
  border-radius: 0.5rem;
  transition: all 0.2s;
  border: 2px solid transparent;
  cursor: pointer;
}

.game-description__link:hover {
  color: #764ba2;
  border-color: #667eea;
  background: rgba(102, 126, 234, 0.1);
}

.game-description__link--toggle {
  color: var(--text-secondary);
  font-weight: 500;
  font-size: 0.875rem;
}

.game-description__link--toggle:hover {
  color: var(--text-primary);
  border-color: transparent;
  background: transparent;
  text-decoration: underline;
}

.game-description__link--show {
  color: #667eea;
  font-size: 1rem;
}

@media (max-width: 768px) {
  .game-wrapper {
    min-height: auto;
  }

  .game-description {
    padding: 1.5rem;
    margin-bottom: 1.5rem;
  }

  .game-description__title {
    font-size: 2rem;
  }

  .game-description__text {
    font-size: 1rem;
  }

  .game-description__features {
    grid-template-columns: 1fr;
    gap: 0.75rem;
  }

  .feature-item {
    padding: 0.625rem;
  }
}
</style>
