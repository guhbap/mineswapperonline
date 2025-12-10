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
      <p class="debug">gameState: {{ gameState ? 'есть' : 'null' }}</p>
      <p class="debug">wsClient: {{ wsClient ? 'есть' : 'null' }}</p>
      <p class="debug">Курсоров других игроков: {{ otherCursors.length }}</p>
    </div>
    <template v-else>
      <div class="debug-info">
        <p>Курсоров других игроков: {{ otherCursors.length }}</p>
        <div v-for="cursor in otherCursors" :key="cursor.playerId" class="debug-cursor">
          {{ cursor.nickname }}: ({{ Math.round(cursor.x) }}, {{ Math.round(cursor.y) }})
        </div>
      </div>
      <div
        class="game-board-wrapper"
      >
      <div
        class="game-board"
        :style="{ gridTemplateColumns: `repeat(${gameState.cols}, 1fr)` }"
        @mousemove="handleMouseMove"
        @mouseleave="handleMouseLeave"
      >
      <div
        v-for="(row, rowIndex) in gameState.board"
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
        v-for="cursor in otherCursors"
        :key="cursor.playerId"
        class="remote-cursor"
        :style="{
          left: cursor.x + 'px',
          top: cursor.y + 'px',
          '--cursor-color': cursor.color,
        }"
        :title="`${cursor.nickname} (${cursor.playerId})`"
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
    </template>

    <!-- Сообщения о состоянии игры -->
    <div v-if="gameState?.gameOver" class="game-message game-message--over">
      <h2>Игра окончена!</h2>
      <p v-if="gameState.loserNickname">
        <strong>{{ gameState.loserNickname }}</strong> подорвался на мине 💣
      </p>
      <p v-else>
        Вы подорвались на мине 💣
      </p>
    </div>
    <div v-else-if="gameState?.gameWon" class="game-message game-message--won">
      <h2>Победа! 🎉</h2>
      <p>Все мины найдены!</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import type { WebSocketMessage, Cell, IWebSocketClient } from '@/api/websocket'

const props = defineProps<{
  wsClient: IWebSocketClient | null
  nickname: string
}>()

const gameState = ref<WebSocketMessage['gameState'] | null>(null)
const otherCursors = ref<Array<{ playerId: string; x: number; y: number; nickname: string; color: string }>>([])
const cursorTimeout = ref<Map<string, number>>(new Map())

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
    console.warn('WebSocket не подключен')
    return
  }
  if (gameState.value?.gameOver || gameState.value?.gameWon) return

  // Проверка: нельзя ставить флаг на открытые ячейки
  if (isRightClick && gameState.value?.board?.[row]?.[col]?.isRevealed) {
    return
  }

  console.log('Отправка клика:', row, col, isRightClick)
  props.wsClient.sendCellClick(row, col, isRightClick)
}

const handleNewGame = () => {
  if (!props.wsClient?.isConnected()) return
  props.wsClient.sendNewGame()
}

const handleMessage = (msg: WebSocketMessage) => {
  console.log('MinesweeperGame handleMessage: получено сообщение:', msg.type, msg)
  if (msg.type === 'gameState' && msg.gameState) {
    console.log('MinesweeperGame: обновление состояния игры:', {
      rows: msg.gameState.rows,
      cols: msg.gameState.cols,
      mines: msg.gameState.mines,
      boardSize: msg.gameState.board?.length,
      revealed: msg.gameState.revealed
    })
    gameState.value = msg.gameState
    console.log('MinesweeperGame: gameState обновлен, текущее значение:', gameState.value)
  } else if (msg.type === 'cursor' && msg.cursor) {
    // playerId может быть на верхнем уровне или внутри cursor
    const playerId = msg.playerId || msg.cursor.playerId
    if (!playerId) {
      console.warn('MinesweeperGame: курсор без playerId', msg)
      return
    }
    console.log('MinesweeperGame: получен курсор от игрока:', playerId, msg.cursor, msg.nickname, msg.color)
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
      console.log('MinesweeperGame: курсор обновлен, всего курсоров:', otherCursors.value.length)
    } else {
      otherCursors.value.push(cursorData)
      console.log('MinesweeperGame: курсор добавлен, всего курсоров:', otherCursors.value.length)
    }

    // Удаление курсора через 2 секунды без обновлений
    const timeoutId = setTimeout(() => {
      const index = otherCursors.value.findIndex(c => c.playerId === playerId)
      if (index >= 0) {
        otherCursors.value.splice(index, 1)
      }
      cursorTimeout.value.delete(playerId)
    }, 2000)

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
    console.log('MinesweeperGame: получено событие ws-message:', customEvent.detail.type)
    handleMessage(customEvent.detail)
  } else {
    console.warn('MinesweeperGame: получено событие без detail:', event)
  }
}

onMounted(() => {
  // Слушаем события WebSocket сообщений
  window.addEventListener('ws-message', messageHandler)
  console.log('MinesweeperGame: слушатель событий установлен, wsClient:', props.wsClient?.isConnected())

  // Если состояние игры уже есть, логируем его
  if (gameState.value) {
    console.log('MinesweeperGame: начальное состояние игры:', gameState.value)
  } else {
    console.log('MinesweeperGame: состояние игры еще не получено')
  }
})

onUnmounted(() => {
  window.removeEventListener('ws-message', messageHandler)
  cursorTimeout.value.forEach(timeout => clearTimeout(timeout))
  cursorTimeout.value.clear()
})
</script>

<style scoped>
.minesweeper-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 2rem;
  position: relative;
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
  transform: translate(-12px, -12px);
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
  margin: 0;
  font-size: 1.125rem;
  color: var(--text-secondary);
  transition: color 0.3s ease;
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

.loading-message .debug {
  color: var(--text-secondary);
  font-size: 0.875rem;
  margin-top: 0.25rem;
  transition: color 0.3s ease;
}

.debug-info {
  position: fixed;
  top: 10px;
  right: 10px;
  background: var(--bg-primary);
  color: var(--text-primary);
  padding: 0.5rem;
  border-radius: 0.25rem;
  font-size: 0.75rem;
  z-index: 2000;
  max-width: 200px;
  box-shadow: 0 2px 8px var(--shadow);
  border: 1px solid var(--border-color);
  transition: background 0.3s ease, color 0.3s ease, border-color 0.3s ease;
}

.debug-cursor {
  margin-top: 0.25rem;
  font-size: 0.7rem;
}
</style>

