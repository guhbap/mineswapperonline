<template>
  <main class="main-page">
    <!-- Выбор комнаты -->
    <section v-if="!selectedRoom" class="rooms-wrapper" aria-label="Выбор игровой комнаты">
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

const handleCreateRoom = async (data: { name: string; password?: string; rows: number; cols: number; mines: number }) => {
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
  }
}

const handleNicknameSubmit = (submittedNickname: string) => {
  nickname.value = submittedNickname
  connectToRoom(submittedNickname)
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
}
</style>
