<template>
  <main class="main-page">
    <!-- Выбор комнаты -->
    <section v-if="!selectedRoom" class="rooms-wrapper" aria-label="Выбор игровой комнаты">
      <!-- Описание игры -->
      <div v-show="isDescriptionVisible" class="game-description">
        <h1 class="game-description__title">
          <svg class="game-description__title-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <rect x="2" y="6" width="20" height="12" rx="2" stroke="currentColor" stroke-width="2"/>
            <path d="M6 10H18M6 14H14" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
          </svg>
          Сапер Онлайн
        </h1>
        <p class="game-description__text">
          Классическая игра Сапер в многопользовательском режиме! Играйте вместе с друзьями в реальном времени,
          соревнуйтесь за лучший результат и получайте рейтинг за успешные игры.
        </p>
        <div class="game-description__features">
          <div class="feature-item">
            <svg class="feature-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M17 21V19C17 17.9391 16.5786 16.9217 15.8284 16.1716C15.0783 15.4214 14.0609 15 13 15H5C3.93913 15 2.92172 15.4214 2.17157 16.1716C1.42143 16.9217 1 17.9391 1 19V21" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
              <circle cx="9" cy="7" r="4" stroke="currentColor" stroke-width="2"/>
              <path d="M23 21V19C22.9993 18.1137 22.7044 17.2528 22.1614 16.5523C21.6184 15.8519 20.8581 15.3516 20 15.13" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
              <path d="M16 3.13C16.8604 3.35031 17.623 3.85071 18.1676 4.55232C18.7122 5.25392 19.0078 6.11683 19.0078 7.005C19.0078 7.89318 18.7122 8.75608 18.1676 9.45769C17.623 10.1593 16.8604 10.6597 16 10.88" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
            <span class="feature-text">Играйте с друзьями в реальном времени</span>
          </div>
          <div class="feature-item">
            <svg class="feature-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M6 9H4.5C3.67157 9 3 9.67157 3 10.5V19.5C3 20.3284 3.67157 21 4.5 21H19.5C20.3284 21 21 20.3284 21 19.5V10.5C21 9.67157 20.3284 9 19.5 9H18M6 9V6C6 4.34315 7.34315 3 9 3H15C16.6569 3 18 4.34315 18 6V9M6 9H18M12 9V21" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
              <path d="M9 12L12 9L15 12" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
            <span class="feature-text">Получайте рейтинг за успешные игры</span>
          </div>
          <div class="feature-item">
            <svg class="feature-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M21 15C21 15.5304 20.7893 16.0391 20.4142 16.4142C20.0391 16.7893 19.5304 17 19 17H7L3 21V5C3 4.46957 3.21071 3.96086 3.58579 3.58579C3.96086 3.21071 4.46957 3 5 3H19C19.5304 3 20.0391 3.21071 20.4142 3.58579C20.7893 3.96086 21 4.46957 21 5V15Z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
            <span class="feature-text">Общайтесь в чате во время игры</span>
          </div>
          <div class="feature-item">
            <svg class="feature-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M12 15C13.6569 15 15 13.6569 15 12C15 10.3431 13.6569 9 12 9C10.3431 9 9 10.3431 9 12C9 13.6569 10.3431 15 12 15Z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
              <path d="M19.4 15C19.2669 15.3016 19.2272 15.6362 19.286 15.9606C19.3448 16.285 19.4995 16.5843 19.73 16.82L19.79 16.88C19.976 17.0657 20.1235 17.2863 20.2241 17.5291C20.3248 17.7719 20.3766 18.0322 20.3766 18.295C20.3766 18.5578 20.3248 18.8181 20.2241 19.0609C20.1235 19.3037 19.976 19.5243 19.79 19.71C19.6043 19.896 19.3837 20.0435 19.1409 20.1441C18.8981 20.2448 18.6378 20.2966 18.375 20.2966C18.1122 20.2966 17.8519 20.2448 17.6091 20.1441C17.3663 20.0435 17.1457 19.896 16.96 19.71L16.9 19.65C16.6643 19.4195 16.365 19.2648 16.0406 19.206C15.7162 19.1472 15.3816 19.1869 15.08 19.32C14.7842 19.4468 14.532 19.6572 14.3543 19.9255C14.1766 20.1938 14.0813 20.5082 14.08 20.83V21C14.08 21.5304 13.8693 22.0391 13.4942 22.4142C13.1191 22.7893 12.6104 23 12.08 23C11.5496 23 11.0409 22.7893 10.6658 22.4142C10.2907 22.0391 10.08 21.5304 10.08 21V20.91C10.0723 20.579 9.96512 20.258 9.77251 19.9887C9.5799 19.7194 9.31074 19.5143 9 19.4C8.69838 19.2669 8.36381 19.2272 8.03941 19.286C7.71502 19.3448 7.41568 19.4995 7.18 19.73L7.12 19.79C6.93425 19.976 6.71368 20.1235 6.47088 20.2241C6.22808 20.3248 5.96783 20.3766 5.705 20.3766C5.44217 20.3766 5.18192 20.3248 4.93912 20.2241C4.69632 20.1235 4.47575 19.976 4.29 19.79C4.10405 19.6043 3.95653 19.3837 3.85588 19.1409C3.75523 18.8981 3.70343 18.6378 3.70343 18.375C3.70343 18.1122 3.75523 17.8519 3.85588 17.6091C3.95653 17.3663 4.10405 17.1457 4.29 16.96L4.35 16.9C4.58054 16.6643 4.73519 16.365 4.794 16.0406C4.85282 15.7162 4.81312 15.3816 4.68 15.08C4.55324 14.7842 4.34276 14.532 4.07447 14.3543C3.80618 14.1766 3.49179 14.0813 3.17 14.08H3C2.46957 14.08 1.96086 13.8693 1.58579 13.4942C1.21071 13.1191 1 12.6104 1 12.08C1 11.5496 1.21071 11.0409 1.58579 10.6658C1.96086 10.2907 2.46957 10.08 3 10.08H3.09C3.42099 10.0723 3.742 9.96512 4.01131 9.77251C4.28062 9.5799 4.48571 9.31074 4.6 9C4.73312 8.69838 4.77282 8.36381 4.714 8.03941C4.65519 7.71502 4.50054 7.41568 4.27 7.18L4.21 7.12C4.02405 6.93425 3.87653 6.71368 3.77588 6.47088C3.67523 6.22808 3.62343 5.96783 3.62343 5.705C3.62343 5.44217 3.67523 5.18192 3.77588 4.93912C3.87653 4.69632 4.02405 4.47575 4.21 4.29C4.39575 4.10405 4.61632 3.95653 4.85912 3.85588C5.10192 3.75523 5.36217 3.70343 5.625 3.70343C5.88783 3.70343 6.14808 3.75523 6.39088 3.85588C6.63368 3.95653 6.85425 4.10405 7.04 4.29L7.1 4.35C7.33568 4.58054 7.63502 4.73519 7.95941 4.794C8.28381 4.85282 8.61838 4.81312 8.92 4.68H9C9.29577 4.55324 9.54802 4.34276 9.72569 4.07447C9.90337 3.80618 9.99872 3.49179 10 3.17V3C10 2.46957 10.2107 1.96086 10.5858 1.58579C10.9609 1.21071 11.4696 1 12 1C12.5304 1 13.0391 1.21071 13.4142 1.58579C13.7893 1.96086 14 2.46957 14 3V3.09C14.0013 3.41179 14.0966 3.72618 14.2743 3.99447C14.452 4.26276 14.7042 4.47324 15 4.6C15.3016 4.73312 15.6362 4.77282 15.9606 4.714C16.285 4.65519 16.5843 4.50054 16.82 4.27L16.88 4.21C17.0657 4.02405 17.2863 3.87653 17.5291 3.77588C17.7719 3.67523 18.0322 3.62343 18.295 3.62343C18.5578 3.62343 18.8181 3.67523 19.0609 3.77588C19.3037 3.87653 19.5243 4.02405 19.71 4.21C19.896 4.39575 20.0435 4.61632 20.1441 4.85912C20.2448 5.10192 20.2966 5.36217 20.2966 5.625C20.2966 5.88783 20.2448 6.14808 20.1441 6.39088C20.0435 6.63368 19.896 6.85425 19.71 7.04L19.65 7.1C19.4195 7.33568 19.2648 7.63502 19.206 7.95941C19.1472 8.28381 19.1869 8.61838 19.32 8.92V9C19.4468 9.29577 19.6572 9.54802 19.9255 9.72569C20.1938 9.90337 20.5082 9.99872 20.83 10H21C21.5304 10 22.0391 10.2107 22.4142 10.5858C22.7893 10.9609 23 11.4696 23 12C23 12.5304 22.7893 13.0391 22.4142 13.4142C22.0391 13.7893 21.5304 14 21 14H20.91C20.5882 14.0013 20.2738 14.0966 20.0055 14.2743C19.7372 14.452 19.5268 14.7042 19.4 15Z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
            <span class="feature-text">Настраивайте размер поля и сложность</span>
          </div>
        </div>
        <div class="game-description__actions">
          <router-link to="/faq" class="game-description__link">
            <svg class="faq-link-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M4 19.5C4 18.1193 5.11929 17 6.5 17H20" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
              <path d="M6.5 2C5.11929 2 4 3.11929 4 4.5V19.5C4 20.8807 5.11929 22 6.5 22H19.5C20.8807 22 22 20.8807 22 19.5V4.5C22 3.11929 20.8807 2 19.5 2H6.5Z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
              <path d="M9 7H15M9 11H15M9 15H13" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
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
