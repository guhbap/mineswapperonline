<template>
  <article class="minesweeper-container" aria-label="Игра Сапер">
    <header class="game-header" role="banner">
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
    </header>

    <div v-if="!gameState" class="loading-message">
      <p>Ожидание состояния игры...</p>
      <p v-if="!wsClient?.isConnected()" class="error">WebSocket не подключен</p>
      <p v-else class="info">WebSocket подключен, ожидание данных...</p>
    </div>
    <!-- <template v-else> -->
      <div class="game-content-wrapper">
        <!-- Игровое поле -->
        <div
          class="game-board-wrapper"
          @contextmenu.prevent
          @touchstart="handleTouchStart"
          @touchmove="handleTouchMove"
          @touchend="handleTouchEnd"
        >
      <!-- Кнопки зума для мобильных -->
      <div v-if="isMobile" class="zoom-controls">
        <button
          @click="zoomOut"
          class="zoom-button zoom-button--minus"
          :disabled="!canZoomOut"
          aria-label="Уменьшить"
        >
          −
        </button>
        <span class="zoom-level">{{ zoomPercentage }}%</span>
        <button
          @click="zoomIn"
          class="zoom-button zoom-button--plus"
          :disabled="!canZoomIn"
          aria-label="Увеличить"
        >
          +
        </button>
        <button
          v-if="isZoomed"
          @click="resetZoom"
          class="zoom-button zoom-button--reset"
          aria-label="Сбросить зум"
          title="Сбросить зум и центрировать"
        >
          ⌂
        </button>
      </div>
      <div
        ref="boardContainer"
        class="game-board-container"
        :style="containerStyle"
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
              'cell--show-mine': (gameState?.gameOver || gameState?.gameWon) && cell.isMine && !cell.isRevealed,
            }
          ]"
          @click="handleCellClick(rowIndex, colIndex, false)"
          @contextmenu.prevent="handleCellClick(rowIndex, colIndex, true)"
          @touchstart.stop="handleCellTouchStart"
          @touchend.stop="handleCellTouchEnd(rowIndex, colIndex, $event, handleCellClick)"
        >
          <span v-if="cell.isRevealed && !cell.isMine && cell.neighborMines > 0" class="cell-number">
            {{ cell.neighborMines }}
          </span>
          <span v-else-if="cell.isRevealed && cell.isMine" class="cell-mine">💣</span>
          <span v-else-if="(gameState?.gameOver || gameState?.gameWon) && cell.isMine && !cell.isRevealed" class="cell-mine">💣</span>
          <span v-else-if="cell.isFlagged" class="cell-flag">🚩</span>
        </div>
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
        <router-link
          :to="`/profile/${cursor.nickname}`"
          class="cursor-label cursor-label--link"
          @click.stop
        >
          {{ cursor.nickname || 'Игрок' }}
        </router-link>
      </div>
      </div>

        <!-- Чат -->
        <div class="chat-wrapper">
          <Chat
            :ws-client="wsClient"
            :own-nickname="nickname"
          />
        </div>

      </div>
    <!-- </template> -->

    <!-- Сообщения о состоянии игры -->
    <div
      v-if="gameState?.gameOver"
      class="game-message game-message--over"
      :class="{ 'game-message--transparent': isModalTransparent }"
    >
      <button
        class="game-message__transparency-button"
        @mousedown="isModalTransparent = true"
        @mouseup="isModalTransparent = false"
        @mouseleave="isModalTransparent = false"
        @touchstart="isModalTransparent = true"
        @touchend="isModalTransparent = false"
        title="Удерживайте для прозрачности"
      >
        👁️
      </button>
      <h2>Игра окончена!</h2>
      <p v-if="gameState.loserNickname">
        <router-link
          :to="`/profile/${gameState.loserNickname}`"
          class="loser-link"
        >
          <strong>{{ gameState.loserNickname }}</strong>
        </router-link> подорвался на мине 💣
      </p>
      <p v-else>
        Вы подорвались на мине 💣
      </p>
      <button @click="handleNewGame" class="game-message__button">
        Новая игра
      </button>
    </div>
    <div
      v-else-if="gameState?.gameWon"
      class="game-message game-message--won"
      :class="{ 'game-message--transparent': isModalTransparent }"
    >
      <button
        class="game-message__transparency-button"
        @mousedown="isModalTransparent = true"
        @mouseup="isModalTransparent = false"
        @mouseleave="isModalTransparent = false"
        @touchstart="isModalTransparent = true"
        @touchend="isModalTransparent = false"
        title="Удерживайте для прозрачности"
      >
        👁️
      </button>
      <h2>Победа! 🎉</h2>
      <p>Все мины найдены!</p>
      <button @click="handleNewGame" class="game-message__button">
        Новая игра
      </button>
    </div>
  </article>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import type { WebSocketMessage, Cell, IWebSocketClient } from '@/api/websocket'
import { useCursorAnimation } from '@/composables/useCursorAnimation'
import { useGameBoardZoom } from '@/composables/useGameBoardZoom'
import { useCellTouch } from '@/composables/useCellTouch'
import Chat from '@/components/Chat.vue'

const props = defineProps<{
  wsClient: IWebSocketClient | null
  nickname: string
}>()

const gameState = ref<WebSocketMessage['gameState'] | null>(null)
const otherCursors = ref<Array<{ playerId: string; x: number; y: number; nickname: string; color: string }>>([])
const cursorTimeout = ref<Map<string, number>>(new Map())
const isModalTransparent = ref(false)
const boardContainer = ref<HTMLElement | null>(null)

// Определение мобильного устройства
const isMobile = computed(() => {
  return window.innerWidth <= 768
})

// Используем composable для зума игрового поля
const {
  zoomLevel,
  zoomPercentage,
  isZoomed,
  canZoomIn,
  canZoomOut,
  zoomIn,
  zoomOut,
  resetZoom,
  handleTouchStart,
  handleTouchMove,
  handleTouchEnd,
  containerStyle,
} = useGameBoardZoom({
  minZoom: 0.5,
  maxZoom: 3,
  zoomStep: 0.1,
  initialZoom: 1,
})

// Используем composable для обработки touch-событий на ячейках
const { handleTouchStart: handleCellTouchStart, handleTouchEnd: handleCellTouchEnd } = useCellTouch({
  clickThreshold: 10,
  clickDuration: 300,
})

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

const handleCellClick = (row: number, col: number, isRightClick: boolean = false) => {
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
})

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
  min-height: 100vh;
}

@media (max-width: 768px) {
  .minesweeper-container {
    padding: 1rem;
    min-height: auto;
  }
}

.game-content-wrapper {
  display: flex;
  justify-content: center;
  align-items: flex-start;
  gap: 2rem;
  width: 100%;
  max-width: 1600px;
}

@media (max-width: 768px) {
  .game-content-wrapper {
    flex-direction: column;
    gap: 1rem;
    width: 100%;
    max-width: 100%;
  }
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
  order: 4;
}

.game-board-wrapper {
  order: 2;
}

.chat-wrapper {
  order: 3;
  flex-shrink: 0;
  width: 300px;
  height: 500px;
  display: flex;
  flex-direction: column;
}

@media (max-width: 768px) {
  .chat-wrapper {
    width: 100%;
    height: 300px;
    order: 1;
  }
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

@media (max-width: 768px) {
  .game-header {
    flex-direction: column;
    gap: 1rem;
    padding: 0.75rem;
    margin-bottom: 1rem;
    width: 100%;
    max-width: 100%;
  }
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
  margin: 0 auto;
}

@media (max-width: 768px) {
  .game-board {
    gap: 1px;
    padding: 1px;
  }
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
  touch-action: manipulation;
}

@media (max-width: 768px) {
  .cell {
    width: 28px;
    height: 28px;
    font-size: 0.75rem;
    border-width: 1px;
  }
}

@media (max-width: 480px) {
  .cell {
    width: 24px;
    height: 24px;
    font-size: 0.7rem;
  }
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

.cell--show-mine {
  background: rgba(239, 68, 68, 0.3);
}

[data-theme="dark"] .cell--show-mine {
  background: rgba(127, 29, 29, 0.5);
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
  text-decoration: none;
  display: inline-block;
  transition: opacity 0.2s;
  pointer-events: auto;
}

.cursor-label--link:hover {
  opacity: 0.8;
}

.loser-link {
  color: inherit;
  text-decoration: none;
  transition: opacity 0.2s;
}

.loser-link:hover {
  opacity: 0.8;
  text-decoration: underline;
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
  z-index: 10000;
  animation: fadeIn 0.3s ease-out;
  transition: background 0.3s ease, opacity 0.2s ease;
  min-width: 300px;
  max-width: 90vw;
}

.game-message--transparent {
  opacity: 0.15;
  pointer-events: none;
}

.game-message--transparent .game-message__transparency-button {
  pointer-events: auto;
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

.game-message__transparency-button {
  position: absolute;
  top: 1rem;
  right: 1rem;
  width: 2.5rem;
  height: 2.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-secondary);
  border: 2px solid var(--border-color);
  border-radius: 50%;
  cursor: pointer;
  font-size: 1.25rem;
  transition: all 0.2s ease;
  user-select: none;
  z-index: 10001;
}

.game-message__transparency-button:hover {
  background: var(--bg-tertiary);
  transform: scale(1.1);
}

.game-message__transparency-button:active {
  transform: scale(0.95);
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

  .game-board-wrapper {
    width: 100%;
    overflow: auto;
    -webkit-overflow-scrolling: touch;
    /* Разрешаем панорамирование и pinch-to-zoom на уровне wrapper */
    touch-action: pan-x pan-y pinch-zoom;
    padding: 0.5rem;
    max-height: 60vh;
    position: relative;
    scroll-behavior: smooth;
    /* Убеждаемся, что скролл работает даже при увеличенном контенте */
    overscroll-behavior: contain;
    /* Улучшаем производительность скролла */
    will-change: scroll-position;
  }

  .game-board-container {
    width: fit-content;
    margin: 0 auto;
    /* Не блокируем touch-события для панорамирования, но обрабатываем pinch-to-zoom */
    touch-action: pan-x pan-y pinch-zoom;
    /* Улучшаем производительность трансформации */
    will-change: transform;
  }

  .game-board {
    min-width: fit-content;
  }

  .zoom-controls {
    position: sticky;
    top: 1rem;
    right: 1rem;
    z-index: 100;
    display: flex;
    align-items: center;
    gap: 0.5rem;
    background: var(--bg-primary);
    padding: 0.5rem;
    border-radius: 0.5rem;
    box-shadow: 0 2px 8px var(--shadow);
    margin-bottom: 0.5rem;
    justify-content: center;
    width: fit-content;
    margin-left: auto;
    margin-right: auto;
  }

  .zoom-button {
    width: 2.5rem;
    height: 2.5rem;
    border: 2px solid var(--border-color);
    background: var(--bg-secondary);
    color: var(--text-primary);
    border-radius: 50%;
    font-size: 1.5rem;
    font-weight: 700;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.2s ease;
    user-select: none;
    touch-action: manipulation;
  }

  .zoom-button:active:not(:disabled) {
    transform: scale(0.95);
    background: var(--bg-tertiary);
  }

  .zoom-button:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .zoom-level {
    min-width: 3rem;
    text-align: center;
    font-weight: 600;
    font-size: 0.875rem;
    color: var(--text-primary);
  }

  .zoom-button--reset {
    font-size: 1.25rem;
    line-height: 1;
  }

  .game-info {
    gap: 1rem;
    width: 100%;
    justify-content: space-between;
  }

  .info-item {
    flex: 1;
  }

  .info-value {
    font-size: 1.2rem;
  }

  .new-game-button {
    width: 100%;
    padding: 0.75rem 1rem;
    font-size: 0.9rem;
  }

  .game-message {
    padding: 1.5rem 1rem;
    max-width: 90vw;
    margin: 0 1rem;
  }

  .game-message h2 {
    font-size: 1.5rem;
  }

  .game-message p {
    font-size: 1rem;
  }

  .game-message__button {
    padding: 0.75rem 1.5rem;
    font-size: 1rem;
    width: 100%;
  }

  .game-message__transparency-button {
    width: 2rem;
    height: 2rem;
    font-size: 1rem;
    top: 0.5rem;
    right: 0.5rem;
  }
}

@media (max-width: 480px) {
  .minesweeper-container {
    padding: 0.5rem;
  }

  .game-header {
    padding: 0.5rem;
  }

  .info-value {
    font-size: 1rem;
  }

  .info-label {
    font-size: 0.75rem;
  }

  .game-board-wrapper {
    max-height: 50vh;
    padding: 0.25rem;
  }
}

</style>

