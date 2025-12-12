<template>
  <article class="minesweeper-container" aria-label="Игра Сапер">
    <header class="game-header" role="banner">
      <div class="game-info">
        <div class="info-item">
          <span class="info-label">Мин:</span>
          <span class="info-value">{{ gameState?.m || 0 }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">Открыто:</span>
          <span class="info-value">{{ gameState?.rv || 0 }}</span>
        </div>
      </div>
      <div class="game-actions">
        <button
          v-if="!room?.noGuessing"
          @click="handleHint"
          class="hint-button"
          :disabled="(gameState?.hu ?? 0) >= 3 || !gameState || gameState.go || gameState.gw || !hasClosedCells"
          :title="(gameState?.hu ?? 0) >= 3 ? 'Подсказки закончились' : `Подсказки: ${3 - (gameState?.hu ?? 0)}/3`"
        >
          💡 Подсказка ({{ 3 - (gameState?.hu ?? 0) }})
        </button>
        <button @click="handleNewGame" class="new-game-button">
          Новая игра
        </button>
        <button
          v-if="isRoomCreator"
          @click="handleEditRoom"
          class="edit-room-button"
          title="Редактировать параметры комнаты"
        >
          ⚙️ Настройки
        </button>
      </div>
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
          ref="boardWrapper"
          class="game-board-wrapper"
          @contextmenu.prevent
          @touchstart="handleTouchStart"
          @touchmove="handleTouchMove"
          @touchend="handleTouchEnd"
          @mousemove="handleBoardMouseMove"
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
        :style="{ gridTemplateColumns: `repeat(${gameState?.c}, 1fr)` }"
        @mousemove="handleMouseMove"
        @mouseleave="handleMouseLeave"
      >
      <div
          v-for="cellData in flatCells"
          :key="`${cellData.rowIndex}-${cellData.colIndex}`"
          :class="[
            'cell',
            {
              'cell--revealed': cellData.cell.r,
              'cell--mine': cellData.cell.r && cellData.cell.m,
              'cell--flagged': cellData.cell.f,
              'cell--show-mine': (gameState?.go || gameState?.gw) && cellData.cell.m && !cellData.cell.r,
              'cell--safe': isSafeCell(cellData.rowIndex, cellData.colIndex) && !cellData.cell.r && !cellData.cell.f,
              'cell--blocked': isCellBlocked(cellData.rowIndex, cellData.colIndex),
            }
          ]"
          @click="handleCellClick(cellData.rowIndex, cellData.colIndex, false)"
          @contextmenu.prevent="handleCellClick(cellData.rowIndex, cellData.colIndex, true)"
          @touchstart.stop="handleCellTouchStart"
          @touchend.stop="handleCellTouchEnd(cellData.rowIndex, cellData.colIndex, $event, handleCellClick)"
        >
          <span v-if="cellData.cell.r && !cellData.cell.m && cellData.cell.n > 0" class="cell-number">
            {{ cellData.cell.n }}
          </span>
          <span v-else-if="cellData.cell.r && cellData.cell.m" class="cell-mine">💣</span>
          <span v-else-if="(gameState?.go || gameState?.gw) && cellData.cell.m && !cellData.cell.r" class="cell-mine">💣</span>
          <svg
            v-else-if="cellData.cell.f"
            class="cell-flag"
            :style="cellData.cell.fc ? { '--flag-color': cellData.cell.fc } : {}"
            viewBox="0 0 24 24"
            width="18"
            height="18"
          >
            <!-- Древко флага -->
            <line
              x1="4"
              y1="2"
              x2="4"
              y2="22"
              stroke="#8b5a3c"
              stroke-width="2"
              stroke-linecap="round"
            />
            <!-- Полотнище флага -->
            <path
              d="M4 2 L18 8 L4 14 Z"
              class="flag-cloth"
              stroke="#000"
              stroke-width="0.5"
            />
          </svg>
          <!-- Зеленый крестик для безопасных ячеек -->
          <svg
            v-if="isSafeCell(cellData.rowIndex, cellData.colIndex) && !cellData.cell.r && !cellData.cell.f"
            class="cell-safe-marker"
            viewBox="0 0 24 24"
            width="20"
            height="20"
          >
            <line
              x1="6"
              y1="6"
              x2="18"
              y2="18"
              stroke="#22c55e"
              stroke-width="3"
              stroke-linecap="round"
            />
            <line
              x1="18"
              y1="6"
              x2="6"
              y2="18"
              stroke="#22c55e"
              stroke-width="3"
              stroke-linecap="round"
            />
          </svg>
      </div>
      </div>
      </div>

      <!-- Курсоры других игроков -->
      <div
        v-for="cursor in displayCursors"
        :key="cursor.playerId"
        class="remote-cursor"
        :class="{ 'remote-cursor--hovered': cursorHovered === cursor.playerId }"
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
        <span class="cursor-label">
          {{ cursor.nickname || 'Игрок' }}
        </span>
      </div>
      </div>

        <!-- Список игроков и Чат -->
        <div class="sidebar-wrapper">
          <!-- Список игроков -->
          <div class="players-list-wrapper">
            <div class="players-list-header">
              <h3 class="players-list-title">Игроки ({{ playersList.length }})</h3>
            </div>
            <div class="players-list">
              <div v-if="playersList.length === 0" class="players-list-empty">
                <p>Нет игроков в комнате</p>
              </div>
              <div
                v-for="player in playersList"
                :key="player.id"
                class="player-item"
                :class="{ 'player-item--own': player.nickname === nickname }"
              >
                <router-link
                  :to="`/profile/${player.nickname}`"
                  class="player-link"
                >
                  <div
                    class="player-avatar"
                    :style="player.color ? { background: player.color } : {}"
                  >
                    <span class="avatar-text">{{ player.nickname[0].toUpperCase() }}</span>
                  </div>
                  <span class="player-name">{{ player.nickname }}</span>
                  <span v-if="player.nickname === nickname" class="player-badge">Вы</span>
                </router-link>
              </div>
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

      </div>
    <!-- </template> -->

    <!-- Сообщения о состоянии игры -->
    <div
      v-if="gameState?.go"
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
      <p v-if="gameState.ln">
        <router-link
          :to="`/profile/${gameState.ln}`"
          class="loser-link"
        >
          <strong>{{ gameState.ln }}</strong>
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
      v-else-if="gameState?.gw"
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
      <div v-if="ratingChange !== null" class="rating-change">
        <div v-if="ratingChange > 0" class="rating-change__positive">
          +{{ Math.round(ratingChange) }} очков попытки
        </div>
        <div v-else class="rating-change__neutral">
          Очки попытки: {{ Math.round(ratingChange) }}
        </div>
        <div class="rating-change__note">
          Финальный прирост зависит от вашего лучшего результата на этом поле
        </div>
      </div>
      <div v-else-if="gameState && calculateDifficulty(gameState.c, gameState.r, gameState.m) === 0" class="rating-change">
        <div class="rating-change__hint">
          Поле слишком простое для получения рейтинга
        </div>
      </div>
      <button @click="handleNewGame" class="game-message__button">
        Новая игра
      </button>
    </div>
  </article>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import type { WebSocketMessage, IWebSocketClient } from '@/api/websocket'
import { useCursorAnimation } from '@/composables/useCursorAnimation'
import { useGameBoardZoom } from '@/composables/useGameBoardZoom'
import { useCellTouch } from '@/composables/useCellTouch'
import { useAuthStore } from '@/stores/auth'
import { calculateDifficulty, calculateRatingChange } from '@/utils/ratingCalculator'
import Chat from '@/components/Chat.vue'

const props = defineProps<{
  wsClient: IWebSocketClient | null
  nickname: string
  room?: { id: string; creatorId?: number; noGuessing?: boolean } | null
}>()

const emit = defineEmits<{
  'edit-room': []
}>()

const gameState = ref<WebSocketMessage['gameState'] | null>(null)
const otherCursors = ref<Array<{ playerId: string; x: number; y: number; nickname: string; color: string }>>([])
const cursorTimeout = ref<Map<string, number>>(new Map())
const cursorHovered = ref<string | null>(null)
const isModalTransparent = ref(false)
const boardWrapper = ref<HTMLElement | null>(null)
const authStore = useAuthStore()

// Проверяем, является ли пользователь создателем комнаты
const isRoomCreator = computed(() => {
  if (!props.room || !authStore.isAuthenticated || !authStore.user) {
    return false
  }
  // Убеждаемся, что creatorId существует и не равен 0 (0 означает гость)
  const creatorId = props.room.creatorId
  const userId = authStore.user.id

  // Если creatorId не определен или равен 0 (комната создана гостем), то не показываем кнопку
  if (creatorId === undefined || creatorId === null || creatorId === 0) {
    return false
  }

  // Строгое сравнение чисел
  return Number(creatorId) === Number(userId)
})

const handleEditRoom = () => {
  emit('edit-room')
}

// Отслеживание времени игры
const gameStartTime = ref<number | null>(null)
const ratingChange = ref<number | null>(null)

// Список игроков в комнате
const playersList = ref<Array<{ id: string; nickname: string; color: string }>>([])

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

// Вычисляемое свойство для преобразования двумерного массива в плоский для правильного отображения в CSS Grid
const flatCells = computed(() => {
  if (!gameState.value?.b) return []
  const cells: Array<{ cell: any; rowIndex: number; colIndex: number }> = []
  gameState.value.b.forEach((row, rowIndex) => {
    row.forEach((cell, colIndex) => {
      cells.push({ cell, rowIndex, colIndex })
    })
  })
  return cells
})

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

const handleBoardMouseMove = (event: MouseEvent) => {
  if (!boardWrapper.value) return

  const rect = boardWrapper.value.getBoundingClientRect()
  const x = event.clientX - rect.left
  const y = event.clientY - rect.top

  // Проверяем, находится ли мышь над каким-либо курсором
  let hoveredCursorId: string | null = null
  for (const cursor of displayCursors.value) {
    const cursorX = cursor.x - 12
    const cursorY = cursor.y - 12
    const cursorSize = 24 // размер курсора
    const labelHeight = 30 // примерная высота лейбла

    // Проверяем, находится ли мышь в области курсора (иконка + лейбл)
    if (
      x >= cursorX &&
      x <= cursorX + cursorSize &&
      y >= cursorY &&
      y <= cursorY + cursorSize + labelHeight
    ) {
      hoveredCursorId = cursor.playerId
      break
    }
  }

  cursorHovered.value = hoveredCursorId
}

const handleCellClick = (row: number, col: number, isRightClick: boolean = false) => {
  if (!props.wsClient?.isConnected()) {
    return
  }
  if (gameState.value?.go || gameState.value?.gw) return

  // Проверка: нельзя ставить флаг на открытые ячейки
  if (isRightClick && gameState.value?.b?.[row]?.[col]?.r) {
    return
  }

  // Блокируем клики на непомеченные ячейки в режиме без угадываний
  if (!isRightClick && isCellBlocked(row, col)) {
    return
  }

  // Запоминаем время начала игры при первом клике
  if (gameStartTime.value === null && !isRightClick) {
    gameStartTime.value = Date.now()
  }

  props.wsClient.sendCellClick(row, col, isRightClick)
}

// Проверяем, есть ли закрытые ячейки для подсказки
const hasClosedCells = computed(() => {
  if (!gameState.value?.b) return false
  for (const row of gameState.value.b) {
    for (const cell of row) {
      if (!cell.r && !cell.f) {
        return true
      }
    }
  }
  return false
})

// Проверяем, является ли ячейка безопасной
const isSafeCell = (row: number, col: number): boolean => {
  if (!gameState.value?.sc) return false
  return gameState.value.sc.some(cell => cell.r === row && cell.c === col)
}

// Проверяем, заблокирована ли ячейка для клика
const isCellBlocked = (row: number, col: number): boolean => {
  if (!props.room?.noGuessing) return false
  if (!gameState.value?.sc || gameState.value.sc.length === 0) return false
  
  // Если есть открытые ячейки (не первый клик) и ячейка не безопасна - блокируем
  const hasRevealedCells = (gameState.value?.rv ?? 0) > 0
  const isSafe = isSafeCell(row, col)
  const isRevealed = gameState.value?.b?.[row]?.[col]?.r ?? false
  const isFlagged = gameState.value?.b?.[row]?.[col]?.f ?? false
  
  return hasRevealedCells && !isSafe && !isRevealed && !isFlagged
}

const handleHint = () => {
  if (!props.wsClient?.isConnected()) return
  if ((gameState.value?.hu ?? 0) >= 3) return
  if (gameState.value?.go || gameState.value?.gw) return
  if (!hasClosedCells.value) return

  // Находим все закрытые ячейки (не открытые и не с флагом)
  const closedCells: Array<{ row: number; col: number }> = []
  if (gameState.value?.b) {
    for (let row = 0; row < gameState.value.b.length; row++) {
      for (let col = 0; col < gameState.value.b[row].length; col++) {
        const cell = gameState.value.b[row][col]
        if (!cell.r && !cell.f) {
          closedCells.push({ row, col })
        }
      }
    }
  }

  if (closedCells.length === 0) return

  // Выбираем случайную закрытую ячейку
  const randomIndex = Math.floor(Math.random() * closedCells.length)
  const selectedCell = closedCells[randomIndex]

  // Отправляем запрос на подсказку (счетчик увеличивается на сервере)
  props.wsClient.sendHint(selectedCell.row, selectedCell.col)
}

const handleNewGame = () => {
  if (!props.wsClient?.isConnected()) return
  props.wsClient.sendNewGame()
  // Сбрасываем время начала игры и изменение рейтинга
  // Подсказки сбрасываются на сервере при создании новой игры
  gameStartTime.value = null
  ratingChange.value = null
  // Список игроков не сбрасываем, так как они остаются в комнате
}

const handleMessage = (msg: WebSocketMessage) => {
  if (msg.type === 'gameState' && msg.gameState) {
    const prevGameWon = gameState.value?.gw
    gameState.value = msg.gameState

    // Если игра только что завершилась победой, рассчитываем изменение рейтинга
    if (msg.gameState.gw && !prevGameWon && gameStartTime.value !== null && gameState.value) {
      const gameTime = (Date.now() - gameStartTime.value) / 1000 // время в секундах
      const currentRating = authStore.user?.rating || 0.0

      // Проверяем, достаточно ли сложности поля для получения рейтинга
      if (calculateDifficulty(gameState.value.c, gameState.value.r, gameState.value.m) > 0) {
        ratingChange.value = calculateDifficulty(gameState.value.c, gameState.value.r, gameState.value.m)
      } else {
        ratingChange.value = null // Поле слишком простое
      }
    }

    // Сбрасываем время начала игры при новой игре
    // Подсказки сбрасываются на сервере при создании новой игры
    if (!msg.gameState.gw && !msg.gameState.go && gameState.value.rv === 0) {
      gameStartTime.value = null
      ratingChange.value = null
    }
  } else if (msg.type === 'cursor' && msg.cursor) {
    // playerId может быть на верхнем уровне или внутри cursor (pid)
    const playerId = msg.playerId || msg.cursor.pid
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
  } else if (msg.type === 'players' && msg.players) {
    // Обновляем список игроков
    playersList.value = msg.players.map((p: any) => ({
      id: p.id || p.playerId || '',
      nickname: p.nickname || 'Игрок',
      color: p.color || '#667eea'
    }))
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

.sidebar-wrapper {
  order: 3;
  flex-shrink: 0;
  width: 300px;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.players-list-wrapper {
  background: var(--bg-primary);
  border-radius: 0.5rem;
  box-shadow: 0 2px 8px var(--shadow);
  overflow: hidden;
  display: flex;
  flex-direction: column;
  max-height: 300px;
}

.players-list-header {
  padding: 1rem;
  border-bottom: 2px solid var(--border-color);
  background: var(--bg-secondary);
}

.players-list-title {
  margin: 0;
  font-size: 1rem;
  font-weight: 600;
  color: var(--text-primary);
}

.players-list {
  overflow-y: auto;
  padding: 0.5rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.player-item {
  padding: 0.5rem;
  border-radius: 0.5rem;
  transition: background 0.2s;
}

.player-item:hover {
  background: var(--bg-secondary);
}

.player-item--own {
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.1) 0%, rgba(118, 75, 162, 0.1) 100%);
}

.player-link {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  text-decoration: none;
  color: var(--text-primary);
  transition: color 0.2s;
}

.player-link:hover {
  color: #667eea;
}

.player-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #667eea;
  color: white;
  font-weight: 700;
  font-size: 0.875rem;
  flex-shrink: 0;
}

.avatar-text {
  user-select: none;
}

.player-name {
  flex: 1;
  font-weight: 500;
  font-size: 0.875rem;
}

.player-badge {
  font-size: 0.75rem;
  color: #667eea;
  font-weight: 600;
  padding: 0.25rem 0.5rem;
  background: rgba(102, 126, 234, 0.1);
  border-radius: 0.25rem;
}

.players-list-empty {
  padding: 1rem;
  text-align: center;
  color: var(--text-secondary);
  font-size: 0.875rem;
}

.chat-wrapper {
  flex-shrink: 0;
  width: 100%;
  height: 500px;
  display: flex;
  flex-direction: column;
}

@media (max-width: 768px) {
  .sidebar-wrapper {
    width: 100%;
    order: 1;
  }

  .players-list-wrapper {
    max-height: 200px;
  }

  .chat-wrapper {
    width: 100%;
    height: 300px;
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

.game-actions {
  display: flex;
  gap: 0.75rem;
  align-items: center;
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

.hint-button {
  padding: 0.75rem 1.5rem;
  font-size: 1rem;
  font-weight: 600;
  color: white;
  background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%);
  border: none;
  border-radius: 0.5rem;
  cursor: pointer;
  transition: transform 0.2s, box-shadow 0.2s, opacity 0.2s;
}

.hint-button:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(245, 158, 11, 0.4);
}

.hint-button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.edit-room-button {
  padding: 0.75rem 1.5rem;
  font-size: 1rem;
  font-weight: 600;
  color: white;
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  border: none;
  border-radius: 0.5rem;
  cursor: pointer;
  transition: transform 0.2s, box-shadow 0.2s;
}

.edit-room-button:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(16, 185, 129, 0.4);
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

.cell:hover:not(.cell--revealed):not(.cell--flagged):not(.cell--blocked) {
  background: var(--border-color);
}

.cell--blocked {
  opacity: 0.5;
  cursor: not-allowed;
  pointer-events: none;
}

.cell--safe {
  position: relative;
}

.cell-safe-marker {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  pointer-events: none;
  z-index: 1;
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
  display: inline-block;
  vertical-align: middle;
  --flag-color: #dc2626;
}

.cell-flag .flag-cloth {
  fill: var(--flag-color, #dc2626);
}

.remote-cursor {
  position: absolute;
  pointer-events: none;
  z-index: 1000;
  left: 0;
  top: 0;
  will-change: transform;
  transition: opacity 0.2s ease;
}

.remote-cursor--hovered {
  opacity: 0.2;
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
  pointer-events: none;
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

.rating-change {
  margin: 1rem 0;
  padding: 0.75rem 1.5rem;
  border-radius: 0.5rem;
  font-size: 1.125rem;
  font-weight: 600;
  animation: slideDown 0.3s ease-out;
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

.rating-change__positive {
  color: #22c55e;
  background: rgba(34, 197, 94, 0.1);
  border: 2px solid rgba(34, 197, 94, 0.3);
}

.rating-change__negative {
  color: #ef4444;
  background: rgba(239, 68, 68, 0.1);
  border: 2px solid rgba(239, 68, 68, 0.3);
}

.rating-change__neutral {
  color: var(--text-secondary);
  background: rgba(107, 114, 128, 0.1);
  border: 2px solid rgba(107, 114, 128, 0.3);
}

.rating-change__hint {
  color: var(--text-secondary);
  font-size: 0.875rem;
  font-weight: 400;
  font-style: italic;
  background: rgba(107, 114, 128, 0.1);
  border: 2px solid rgba(107, 114, 128, 0.3);
}

.rating-change__note {
  margin-top: 0.5rem;
  font-size: 0.75rem;
  color: var(--text-secondary);
  font-style: italic;
  opacity: 0.8;
  text-align: center;
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

  .game-actions {
    flex-direction: column;
    width: 100%;
    gap: 0.5rem;
  }

  .hint-button,
  .new-game-button,
  .edit-room-button {
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

