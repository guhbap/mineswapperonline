<template>
  <div class="main-page">
    <NicknameModal
      v-if="!nickname"
      :show="!nickname"
      @submit="handleNicknameSubmit"
    />
    <div v-else class="game-wrapper">
      <MinesweeperGame
        :ws-client="wsClient"
        :nickname="nickname"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onUnmounted } from 'vue'
import NicknameModal from '@/components/NicknameModal.vue'
import MinesweeperGame from '@/components/MinesweeperGame.vue'
import { WebSocketClient, type WebSocketMessage, type IWebSocketClient } from '@/api/websocket'

const nickname = ref('')
const wsClient = ref<IWebSocketClient | null>(null)

const handleNicknameSubmit = (submittedNickname: string) => {
  nickname.value = submittedNickname

  // Создаем WebSocket соединение
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const host = import.meta.env.DEV
    ? 'localhost:8080'
    : window.location.host
  const wsUrl = `${protocol}//${host}/ws`

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
        wsClient.value.sendNickname(submittedNickname)
        console.log('Никнейм отправлен:', submittedNickname)
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
