<template>
  <div class="main-page">
    <!-- Выбор комнаты -->
    <div v-if="!selectedRoom" class="rooms-wrapper">
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
        @cancel="selectedRoomForJoin = null"
      />
    </div>

    <!-- Ввод никнейма (только для гостей) -->
    <NicknameModal
      v-if="shouldShowNicknameModal"
      :show="shouldShowNicknameModal"
      @submit="handleNicknameSubmit"
    />

    <!-- Игра -->
    <div v-if="getNickname && selectedRoom" class="game-wrapper">
      <MinesweeperGame
        :ws-client="wsClient"
        :nickname="getNickname"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onUnmounted, computed } from 'vue'
import { useAuthStore } from '@/stores/auth'
import NicknameModal from '@/components/NicknameModal.vue'
import MinesweeperGame from '@/components/MinesweeperGame.vue'
import RoomsList from '@/components/RoomsList.vue'
import CreateRoomModal from '@/components/CreateRoomModal.vue'
import JoinRoomModal from '@/components/JoinRoomModal.vue'
import { WebSocketClient, type WebSocketMessage, type IWebSocketClient } from '@/api/websocket'
import { createRoom, type Room } from '@/api/rooms'

const authStore = useAuthStore()
const nickname = ref('')
const wsClient = ref<IWebSocketClient | null>(null)
const selectedRoom = ref<Room | null>(null)
const selectedRoomForJoin = ref<Room | null>(null)
const showCreateModal = ref(false)

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

const handleJoinRoom = (room: Room) => {
  selectedRoom.value = room
  selectedRoomForJoin.value = null
  
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
  const wsUrl = `${protocol}//${host}/api/ws?room=${selectedRoom.value.id}`

  wsClient.value = new WebSocketClient(
    wsUrl,
    (msg: WebSocketMessage) => {
      // Обработка сообщений будет в компоненте игры через событие
      console.log('MainPage: получено сообщение, отправка события:', msg.type)
      const event = new CustomEvent('ws-message', { detail: msg })
      window.dispatchEvent(event)
      console.log('MainPage: событие отправлено')
    },
    () => {
      console.log('WebSocket подключен')
      // Отправляем никнейм после подключения
      if (wsClient.value) {
        wsClient.value.sendNickname(playerNickname)
        console.log('Никнейм отправлен:', playerNickname)
      }
    },
    () => {
      console.log('WebSocket отключен')
    },
    (error) => {
      console.error('Ошибка WebSocket:', error)
    }
  )

  wsClient.value.connect()
}

onUnmounted(() => {
  wsClient.value?.disconnect()
})
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
