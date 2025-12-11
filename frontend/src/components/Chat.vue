<template>
  <div class="chat-container">
    <div class="chat-header">
      <h3 class="chat-title">Чат</h3>
    </div>
    <div class="chat-messages" ref="messagesContainer">
      <div
        v-for="(message, index) in messages"
        :key="index"
        class="chat-message"
        :class="{
          'chat-message--system': message.isSystem,
          'chat-message--own': message.nickname === ownNickname
        }"
      >
        <span v-if="!message.isSystem" class="message-author" :style="{ color: message.color }">
          {{ message.nickname }}:
        </span>
        <span class="message-text">{{ message.text }}</span>
        <span class="message-time">{{ formatTime(message.timestamp) }}</span>
      </div>
    </div>
    <div class="chat-input-wrapper">
      <input
        v-model="inputMessage"
        @keyup.enter="sendMessage"
        type="text"
        class="chat-input"
        placeholder="Введите сообщение..."
        :disabled="!wsClient?.isConnected()"
      />
      <button
        @click="sendMessage"
        class="chat-send-button"
        :disabled="!wsClient?.isConnected() || !inputMessage.trim()"
      >
        Отправить
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick, watch } from 'vue'
import type { IWebSocketClient } from '@/api/websocket'

interface ChatMessage {
  text: string
  nickname: string
  color: string
  timestamp: number
  isSystem: boolean
  action?: string
  row?: number
  col?: number
}

const props = defineProps<{
  wsClient: IWebSocketClient | null
  ownNickname: string
}>()

const messages = ref<ChatMessage[]>([])
const inputMessage = ref('')
const messagesContainer = ref<HTMLElement | null>(null)

const formatTime = (timestamp: number) => {
  const date = new Date(timestamp)
  return date.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })
}

const scrollToBottom = () => {
  nextTick(() => {
    if (messagesContainer.value) {
      messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
    }
  })
}

const addMessage = (message: ChatMessage) => {
  messages.value.push(message)
  scrollToBottom()
  
  // Ограничиваем количество сообщений (храним только последние 100)
  if (messages.value.length > 100) {
    messages.value.shift()
  }
}

const sendMessage = () => {
  if (!props.wsClient?.isConnected() || !inputMessage.value.trim()) {
    return
  }

  const text = inputMessage.value.trim()
  inputMessage.value = ''

  if (props.wsClient) {
    props.wsClient.sendChatMessage(text)
  }
}

const handleChatMessage = (msg: any) => {
  if (msg.type === 'chat' && msg.chat) {
    addMessage({
      text: msg.chat.text,
      nickname: msg.nickname || 'Игрок',
      color: msg.color || '#667eea',
      timestamp: Date.now(),
      isSystem: msg.chat.isSystem || false,
      action: msg.chat.action,
      row: msg.chat.row,
      col: msg.chat.col
    })
  }
}

// Слушаем события WebSocket сообщений
const messageHandler = (event: Event) => {
  const customEvent = event as CustomEvent
  if (customEvent && customEvent.detail) {
    handleChatMessage(customEvent.detail)
  }
}

// Очищаем сообщения при сбросе игры
const handleResetGame = () => {
  messages.value = []
}

onMounted(() => {
  window.addEventListener('ws-message', messageHandler)
  window.addEventListener('reset-game', handleResetGame)
})

onUnmounted(() => {
  window.removeEventListener('ws-message', messageHandler)
  window.removeEventListener('reset-game', handleResetGame)
})
</script>

<style scoped>
.chat-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: var(--bg-primary);
  border-radius: 0.5rem;
  box-shadow: 0 2px 8px var(--shadow);
  overflow: hidden;
}

.chat-header {
  padding: 1rem;
  border-bottom: 2px solid var(--border-color);
  background: var(--bg-secondary);
}

.chat-title {
  margin: 0;
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--text-primary);
}

.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  min-height: 0;
}

.chat-message {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  padding: 0.5rem;
  border-radius: 0.5rem;
  background: var(--bg-secondary);
  font-size: 0.875rem;
  line-height: 1.4;
  word-break: break-word;
}

.chat-message--system {
  background: rgba(102, 126, 234, 0.1);
  color: var(--text-secondary);
  font-style: italic;
}

.chat-message--own {
  background: rgba(102, 126, 234, 0.15);
}

.message-author {
  font-weight: 600;
}

.message-text {
  flex: 1;
  color: var(--text-primary);
}

.message-time {
  font-size: 0.75rem;
  color: var(--text-secondary);
  margin-left: auto;
}

.chat-input-wrapper {
  display: flex;
  gap: 0.5rem;
  padding: 1rem;
  border-top: 2px solid var(--border-color);
  background: var(--bg-secondary);
  flex-shrink: 0;
  min-width: 0;
}

.chat-input {
  flex: 1;
  min-width: 0;
  padding: 0.75rem;
  font-size: 0.875rem;
  border: 2px solid var(--border-color);
  border-radius: 0.5rem;
  background: var(--bg-primary);
  color: var(--text-primary);
  transition: border-color 0.2s;
}

.chat-input:focus {
  outline: none;
  border-color: #667eea;
}

.chat-input:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.chat-send-button {
  padding: 0.75rem 1rem;
  font-size: 0.875rem;
  font-weight: 600;
  color: white;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
  border-radius: 0.5rem;
  cursor: pointer;
  transition: transform 0.2s, box-shadow 0.2s;
  flex-shrink: 0;
  white-space: nowrap;
}

.chat-send-button:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
}

.chat-send-button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Стилизация скроллбара */
.chat-messages::-webkit-scrollbar {
  width: 6px;
}

.chat-messages::-webkit-scrollbar-track {
  background: var(--bg-secondary);
}

.chat-messages::-webkit-scrollbar-thumb {
  background: var(--border-color);
  border-radius: 3px;
}

.chat-messages::-webkit-scrollbar-thumb:hover {
  background: var(--text-secondary);
}

@media (max-width: 768px) {
  .chat-container {
    height: 100%;
  }

  .chat-header {
    padding: 0.75rem;
  }

  .chat-title {
    font-size: 1rem;
  }

  .chat-messages {
    padding: 0.75rem;
    gap: 0.5rem;
  }

  .chat-message {
    font-size: 0.8rem;
    padding: 0.4rem;
  }

  .chat-input-wrapper {
    padding: 0.75rem;
    gap: 0.4rem;
  }

  .chat-input {
    padding: 0.6rem;
    font-size: 0.8rem;
  }

  .chat-send-button {
    padding: 0.6rem 0.75rem;
    font-size: 0.8rem;
  }
}
</style>

