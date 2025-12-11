<template>
  <div class="minesweeper-container">
    <div class="game-header">
      <div class="game-info">
        <div class="info-item">
          <span class="info-label">Мин:</span>
          <span class="info-value">{{ gameState?.mines || 0 }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">Открыто:</span>
          <span class="info-value">{{ gameState?.revealed || 0 }}</span>
        </div>
      </div>
      <button @click="handleNewGame" class="new-game-button">
        Новая игра
      </button>
    </div>

    <div v-if="!gameState" class="loading-message">
      <p>Ожидание состояния игры...</p>
      <p v-if="!wsClient?.isConnected()" class="error">WebSocket не подключен</p>
      <p v-else class="info">WebSocket подключен, ожидание данных...</p>
    </div>
    <!-- <template v-else> -->
      <div class="game-content-wrapper">
        <!-- Левый рекламный блок -->
        <div class="ad-block ad-block--left">
          <div id="yandex_rtb_R-A-17973092-1"></div>
        </div>

        <!-- Игровое поле -->
        <div
          class="game-board-wrapper"
          @contextmenu.prevent
        >
      <div
        class="game-board"
        :style="{ gridTemplateColumns: `repeat(${gameState?.cols}, 1fr)` }"
        @mousemove="handleMouseMove"
        @mouseleave="handleMouseLeave"
      >
      <div
        v-for="(row, rowIndex) in gameState?.board"
        :key="rowIndex"
      >
        <div
          v-for="(cell, colIndex) in row"
          :key="colIndex"
          :class="[
            'cell',
            {
              'cell--revealed': cell.isRevealed,
              'cell--mine': cell.isRevealed && cell.isMine,
              'cell--flagged': cell.isFlagged,
            }
          ]"
          @click="handleCellClick(rowIndex, colIndex, false)"
          @contextmenu.prevent="handleCellClick(rowIndex, colIndex, true)"
        >
          <span v-if="cell.isRevealed && !cell.isMine && cell.neighborMines > 0" class="cell-number">
            {{ cell.neighborMines }}
          </span>
          <span v-else-if="cell.isRevealed && cell.isMine" class="cell-mine">💣</span>
          <span v-else-if="cell.isFlagged" class="cell-flag">🚩</span>
        </div>
      </div>
      </div>

      <!-- Курсоры других игроков -->
      <div
        v-for="cursor in displayCursors"
        :key="cursor.playerId"
        class="remote-cursor"
        :style="{
          transform: `translate(${cursor.x - 12}px, ${cursor.y - 12}px)`,
          '--cursor-color': cursor.color,
        }"
        :title="cursor.nickname"
      >
        <svg
          class="cursor-icon"
          width="24"
          height="24"
          viewBox="0 0 24 24"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
        >
          <path
            d="M3 3L10.07 19.97L12.58 12.58L19.97 10.07L3 3Z"
            :fill="cursor.color"
            stroke="white"
            stroke-width="1.5"
          />
        </svg>
        <div class="cursor-label">{{ cursor.nickname || 'Игрок' }}</div>
      </div>
      </div>

        <!-- Правый рекламный блок -->
        <div class="ad-block ad-block--right">
          <div id="yandex_rtb_R-A-17973092-2"></div>
        </div>
      </div>
    <!-- </template> -->

    <!-- Сообщения о состоянии игры -->
    <div v-if="gameState?.gameOver" class="game-message game-message--over">
      <h2>Игра окончена!</h2>
      <p v-if="gameState.loserNickname">
        <strong>{{ gameState.loserNickname }}</strong> подорвался на мине 💣
      </p>
      <p v-else>
        Вы подорвались на мине 💣
      </p>
      <button @click="handleNewGame" class="game-message__button">
        Новая игра
      </button>
    </div>
    <div v-else-if="gameState?.gameWon" class="game-message game-message--won">
      <h2>Победа! 🎉</h2>
      <p>Все мины найдены!</p>
      <button @click="handleNewGame" class="game-message__button">
        Новая игра
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import type { WebSocketMessage, Cell, IWebSocketClient } from '@/api/websocket'
import { useCursorAnimation } from '@/composables/useCursorAnimation'

const props = defineProps<{
  wsClient: IWebSocketClient | null
  nickname: string
}>()

const gameState = ref<WebSocketMessage['gameState'] | null>(null)
const otherCursors = ref<Array<{ playerId: string; x: number; y: number; nickname: string; color: string }>>([])
const cursorTimeout = ref<Map<string, number>>(new Map())

// Используем анимацию курсоров
const { animatedCursors, updateCursor, removeCursor } = useCursorAnimation()

// Вычисляемое свойство для отображения курсоров с плавной анимацией
// Фильтруем свой собственный курсор
const displayCursors = computed(() => {
  return Array.from(animatedCursors.value.entries())
    .map(([playerId, pos]) => {
      const cursorInfo = otherCursors.value.find(c => c.playerId === playerId)
      return {
        playerId,
        x: pos.x,
        y: pos.y,
        nickname: cursorInfo?.nickname || 'Игрок',
        color: cursorInfo?.color || '#667eea'
      }
    })
    .filter(cursor => cursor.nickname !== props.nickname) // Фильтруем свой курсор
})

const handleMouseMove = (event: MouseEvent) => {
  if (!props.wsClient?.isConnected()) return

  const rect = (event.currentTarget as HTMLElement).getBoundingClientRect()
  const x = event.clientX - rect.left
  const y = event.clientY - rect.top

  props.wsClient.sendCursor(x, y)
}

const handleMouseLeave = () => {
  // Можно отправить сообщение об уходе курсора, но для простоты просто очистим через таймаут
}

const handleCellClick = (row: number, col: number, isRightClick: boolean) => {
  if (!props.wsClient?.isConnected()) {
    return
  }
  if (gameState.value?.gameOver || gameState.value?.gameWon) return

  // Проверка: нельзя ставить флаг на открытые ячейки
  if (isRightClick && gameState.value?.board?.[row]?.[col]?.isRevealed) {
    return
  }

  props.wsClient.sendCellClick(row, col, isRightClick)
}

const handleNewGame = () => {
  if (!props.wsClient?.isConnected()) return
  props.wsClient.sendNewGame()
}

const handleMessage = (msg: WebSocketMessage) => {
  if (msg.type === 'gameState' && msg.gameState) {
    gameState.value = msg.gameState
  } else if (msg.type === 'cursor' && msg.cursor) {
    // playerId может быть на верхнем уровне или внутри cursor
    const playerId = msg.playerId || msg.cursor.playerId
    if (!playerId) {
      return
    }

    // Обновляем информацию о курсоре
    const existingIndex = otherCursors.value.findIndex(c => c.playerId === playerId)
    const cursorData = {
      playerId: playerId,
      x: msg.cursor.x,
      y: msg.cursor.y,
      nickname: msg.nickname || 'Игрок',
      color: msg.color || '#667eea',
    }

    if (existingIndex >= 0) {
      otherCursors.value[existingIndex] = cursorData
    } else {
      otherCursors.value.push(cursorData)
    }

    // Обновляем анимированную позицию курсора
    updateCursor(playerId, msg.cursor.x, msg.cursor.y)

    // Удаление курсора через 3 секунды без обновлений
    const timeoutId = setTimeout(() => {
      const index = otherCursors.value.findIndex(c => c.playerId === playerId)
      if (index >= 0) {
        otherCursors.value.splice(index, 1)
      }
      removeCursor(playerId)
      cursorTimeout.value.delete(playerId)
    }, 3000)

    const oldTimeout = cursorTimeout.value.get(playerId)
    if (oldTimeout) {
      clearTimeout(oldTimeout)
    }
    cursorTimeout.value.set(playerId, timeoutId as unknown as number)
  }
}

const messageHandler = (event: Event) => {
  const customEvent = event as CustomEvent<WebSocketMessage>
  if (customEvent && customEvent.detail) {
    handleMessage(customEvent.detail)
  }
}

// Очистка курсоров
const clearCursors = () => {
  cursorTimeout.value.forEach(timeout => clearTimeout(timeout))
  cursorTimeout.value.clear()
  otherCursors.value.forEach(cursor => {
    removeCursor(cursor.playerId)
  })
  otherCursors.value = []
}

// Слушаем событие для очистки игры
const handleResetGame = () => {
  clearCursors()
}

onMounted(() => {
  // Слушаем события WebSocket сообщений
  window.addEventListener('ws-message', messageHandler)
  // Слушаем событие для очистки игры
  window.addEventListener('reset-game', handleResetGame)

  // Инициализация рекламы Яндекса
  loadYandexAds()
})

const loadYandexAds = () => {
  const win = window as any

  // Инициализируем контекстную рекламу
  win.yaContextCb = win.yaContextCb || []

  // Функция для рендеринга рекламы
  const renderAds = () => {
    if (win.Ya && win.Ya.Context && win.Ya.Context.AdvManager) {
      // Левый блок
      win.Ya.Context.AdvManager.render({
        blockId: 'R-A-17973092-1',
        renderTo: 'yandex_rtb_R-A-17973092-1'
      })

      // Правый блок
      win.Ya.Context.AdvManager.render({
        blockId: 'R-A-17973092-1',
        renderTo: 'yandex_rtb_R-A-17973092-2'
      })
    }
  }

  // Если скрипт уже загружен, рендерим сразу
  if (win.Ya && win.Ya.Context) {
    renderAds()
    return
  }

  // Загружаем скрипт контекстной рекламы, если его еще нет
  if (!document.querySelector('script[src="https://yandex.ru/ads/system/context.js"]')) {
    const script = document.createElement('script')
    script.src = 'https://yandex.ru/ads/system/context.js'
    script.async = true
    script.onload = () => {
      // Ждем немного, чтобы Ya.Context был готов
      setTimeout(renderAds, 100)
    }
    document.head.appendChild(script)
  }

  // Добавляем в очередь на случай, если скрипт уже загружается
  win.yaContextCb.push(renderAds)
}

onUnmounted(() => {
  window.removeEventListener('ws-message', messageHandler)
  window.removeEventListener('reset-game', handleResetGame)
  clearCursors()
})
</script>

<style scoped>
.minesweeper-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 2rem;
  position: relative;
  width: 100%;
}

.game-content-wrapper {
  display: flex;
  justify-content: center;
  align-items: flex-start;
  gap: 2rem;
  width: 100%;
  max-width: 1400px;
}

.ad-block {
  flex-shrink: 0;
  width: 240px;
  min-height: 400px;
  display: flex;
  justify-content: center;
  align-items: flex-start;
}

.ad-block--left {
  order: 1;
}

.ad-block--right {
  order: 3;
}

.game-board-wrapper {
  order: 2;
}

.game-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
  max-width: 800px;
  margin-bottom: 1.5rem;
  padding: 1rem;
  background: var(--bg-primary);
  border-radius: 0.5rem;
  box-shadow: 0 2px 8px var(--shadow);
  transition: background 0.3s ease;
}

.game-info {
  display: flex;
  gap: 2rem;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.info-label {
  font-size: 0.875rem;
  color: var(--text-secondary);
  font-weight: 500;
  transition: color 0.3s ease;
}

.info-value {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--text-primary);
  transition: color 0.3s ease;
}

.new-game-button {
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

.new-game-button:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
}

.game-board-wrapper {
  position: relative;
  display: inline-block;
  overflow: visible;
}

.game-board {
  display: grid;
  gap: 2px;
  background: var(--border-color);
  padding: 2px;
  border-radius: 0.5rem;
  position: relative;
}

.cell {
  width: 32px;
  height: 32px;
  background: var(--bg-tertiary);
  border: 2px outset var(--border-color);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  font-weight: 700;
  font-size: 0.875rem;
  transition: background-color 0.1s, border-color 0.3s ease;
  user-select: none;
}

.cell:hover:not(.cell--revealed):not(.cell--flagged) {
  background: var(--border-color);
}

.cell--revealed {
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-style: inset;
}

.cell--mine {
  background: #fee2e2;
}

[data-theme="dark"] .cell--mine {
  background: #7f1d1d;
}

.cell--flagged {
  background: #fef3c7;
}

[data-theme="dark"] .cell--flagged {
  background: #78350f;
}

.cell-number {
  color: var(--text-primary);
  transition: color 0.3s ease;
}

.cell-mine {
  font-size: 1.25rem;
}

.cell-flag {
  font-size: 1rem;
}

.remote-cursor {
  position: absolute;
  pointer-events: none;
  z-index: 1000;
  left: 0;
  top: 0;
  will-change: transform;
}

.cursor-icon {
  filter: drop-shadow(0 2px 4px rgba(0, 0, 0, 0.3));
}

.cursor-label {
  position: absolute;
  top: 20px;
  left: 50%;
  transform: translateX(-50%);
  background: var(--cursor-color);
  color: white;
  padding: 0.25rem 0.5rem;
  border-radius: 0.25rem;
  font-size: 0.75rem;
  font-weight: 600;
  white-space: nowrap;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}

.game-message {
  position: fixed;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  background: var(--bg-primary);
  padding: 2rem 3rem;
  border-radius: 1rem;
  box-shadow: 0 20px 60px var(--shadow);
  text-align: center;
  z-index: 200;
  animation: fadeIn 0.3s ease-out;
  transition: background 0.3s ease;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translate(-50%, -50%) scale(0.9);
  }
  to {
    opacity: 1;
    transform: translate(-50%, -50%) scale(1);
  }
}

.game-message h2 {
  margin: 0 0 0.5rem 0;
  font-size: 2rem;
  color: var(--text-primary);
  transition: color 0.3s ease;
}

.game-message p {
  margin: 0 0 1.5rem 0;
  font-size: 1.125rem;
  color: var(--text-secondary);
  transition: color 0.3s ease;
}

.game-message__button {
  padding: 0.75rem 2rem;
  font-size: 1rem;
  font-weight: 600;
  color: white;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
  border-radius: 0.5rem;
  cursor: pointer;
  transition: transform 0.2s, box-shadow 0.2s;
  margin-top: 0.5rem;
}

.game-message__button:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
}

.game-message__button:active {
  transform: translateY(0);
}

.game-message--over h2 {
  color: #dc2626;
}

.game-message--won h2 {
  color: #16a34a;
}

.loading-message {
  padding: 2rem;
  text-align: center;
  color: var(--text-secondary);
  transition: color 0.3s ease;
}

.loading-message .error {
  color: #dc2626;
  margin-top: 1rem;
}

.loading-message .info {
  color: #16a34a;
  margin-top: 0.5rem;
}

@media (max-width: 1200px) {
  .ad-block {
    width: 200px;
  }
}

@media (max-width: 1024px) {
  .game-content-wrapper {
    flex-direction: column;
    align-items: center;
  }

  .ad-block {
    width: 100%;
    max-width: 728px;
    min-height: 90px;
    order: 3;
  }

  .ad-block--left {
    order: 1;
  }

  .ad-block--right {
    order: 2;
  }

  .game-board-wrapper {
    order: 0;
  }
}

@media (max-width: 768px) {
  .ad-block {
    display: none;
  }
}

</style>

