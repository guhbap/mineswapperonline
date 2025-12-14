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
        <div v-if="currentRating !== null && !gameState?.go && !gameState?.gw" class="info-item info-item--rating">
          <span class="info-label">Рейтинг:</span>
          <span class="info-value">{{ Math.round(currentRating) }}</span>
        </div>
      </div>
      <div class="game-actions">
        <button
          v-if="room?.gameMode === 'classic'"
          @click="handleHint"
          class="hint-button"
          :disabled="(gameState?.hu ?? 0) >= 3 || !gameState || gameState.go || gameState.gw || !hasClosedCells"
          :title="(gameState?.hu ?? 0) >= 3 ? 'Подсказки закончились' : `Подсказки: ${3 - (gameState?.hu ?? 0)}/3`"
        >
          <svg class="hint-button-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M12 2V4M12 20V22M4 12H2M6.31412 6.31412L4.8999 4.8999M17.6859 6.31412L19.1001 4.8999M6.31412 17.69L4.8999 19.1042M17.6859 17.69L19.1001 19.1042M22 12H20M17 12C17 14.7614 14.7614 17 12 17C9.23858 17 7 14.7614 7 12C7 9.23858 9.23858 7 12 7C14.7614 7 17 9.23858 17 12Z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
          Подсказка ({{ 3 - (gameState?.hu ?? 0) }})
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
          <svg class="edit-room-button-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M12 15C13.6569 15 15 13.6569 15 12C15 10.3431 13.6569 9 12 9C10.3431 9 9 10.3431 9 12C9 13.6569 10.3431 15 12 15Z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            <path d="M19.4 15C19.2669 15.3016 19.2272 15.6362 19.286 15.9606C19.3448 16.285 19.4995 16.5843 19.73 16.82L19.79 16.88C19.976 17.0657 20.1235 17.2863 20.2241 17.5291C20.3248 17.7719 20.3766 18.0322 20.3766 18.295C20.3766 18.5578 20.3248 18.8181 20.2241 19.0609C20.1235 19.3037 19.976 19.5243 19.79 19.71C19.6043 19.896 19.3837 20.0435 19.1409 20.1441C18.8981 20.2448 18.6378 20.2966 18.375 20.2966C18.1122 20.2966 17.8519 20.2448 17.6091 20.1441C17.3663 20.0435 17.1457 19.896 16.96 19.71L16.9 19.65C16.6643 19.4195 16.365 19.2648 16.0406 19.206C15.7162 19.1472 15.3816 19.1869 15.08 19.32C14.7842 19.4468 14.532 19.6572 14.3543 19.9255C14.1766 20.1938 14.0813 20.5082 14.08 20.83V21C14.08 21.5304 13.8693 22.0391 13.4942 22.4142C13.1191 22.7893 12.6104 23 12.08 23C11.5496 23 11.0409 22.7893 10.6658 22.4142C10.2907 22.0391 10.08 21.5304 10.08 21V20.91C10.0723 20.579 9.96512 20.258 9.77251 19.9887C9.5799 19.7194 9.31074 19.5143 9 19.4C8.69838 19.2669 8.36381 19.2272 8.03941 19.286C7.71502 19.3448 7.41568 19.4995 7.18 19.73L7.12 19.79C6.93425 19.976 6.71368 20.1235 6.47088 20.2241C6.22808 20.3248 5.96783 20.3766 5.705 20.3766C5.44217 20.3766 5.18192 20.3248 4.93912 20.2241C4.69632 20.1235 4.47575 19.976 4.29 19.79C4.10405 19.6043 3.95653 19.3837 3.85588 19.1409C3.75523 18.8981 3.70343 18.6378 3.70343 18.375C3.70343 18.1122 3.75523 17.8519 3.85588 17.6091C3.95653 17.3663 4.10405 17.1457 4.29 16.96L4.35 16.9C4.58054 16.6643 4.73519 16.365 4.794 16.0406C4.85282 15.7162 4.81312 15.3816 4.68 15.08C4.55324 14.7842 4.34276 14.532 4.07447 14.3543C3.80618 14.1766 3.49179 14.0813 3.17 14.08H3C2.46957 14.08 1.96086 13.8693 1.58579 13.4942C1.21071 13.1191 1 12.6104 1 12.08C1 11.5496 1.21071 11.0409 1.58579 10.6658C1.96086 10.2907 2.46957 10.08 3 10.08H3.09C3.42099 10.0723 3.742 9.96512 4.01131 9.77251C4.28062 9.5799 4.48571 9.31074 4.6 9C4.73312 8.69838 4.77282 8.36381 4.714 8.03941C4.65519 7.71502 4.50054 7.41568 4.27 7.18L4.21 7.12C4.02405 6.93425 3.87653 6.71368 3.77588 6.47088C3.67523 6.22808 3.62343 5.96783 3.62343 5.705C3.62343 5.44217 3.67523 5.18192 3.77588 4.93912C3.87653 4.69632 4.02405 4.47575 4.21 4.29C4.39575 4.10405 4.61632 3.95653 4.85912 3.85588C5.10192 3.75523 5.36217 3.70343 5.625 3.70343C5.88783 3.70343 6.14808 3.75523 6.39088 3.85588C6.63368 3.95653 6.85425 4.10405 7.04 4.29L7.1 4.35C7.33568 4.58054 7.63502 4.73519 7.95941 4.794C8.28381 4.85282 8.61838 4.81312 8.92 4.68H9C9.29577 4.55324 9.54802 4.34276 9.72569 4.07447C9.90337 3.80618 9.99872 3.49179 10 3.17V3C10 2.46957 10.2107 1.96086 10.5858 1.58579C10.9609 1.21071 11.4696 1 12 1C12.5304 1 13.0391 1.21071 13.4142 1.58579C13.7893 1.96086 14 2.46957 14 3V3.09C14.0013 3.41179 14.0966 3.72618 14.2743 3.99447C14.452 4.26276 14.7042 4.47324 15 4.6C15.3016 4.73312 15.6362 4.77282 15.9606 4.714C16.285 4.65519 16.5843 4.50054 16.82 4.27L16.88 4.21C17.0657 4.02405 17.2863 3.87653 17.5291 3.77588C17.7719 3.67523 18.0322 3.62343 18.295 3.62343C18.5578 3.62343 18.8181 3.67523 19.0609 3.77588C19.3037 3.87653 19.5243 4.02405 19.71 4.21C19.896 4.39575 20.0435 4.61632 20.1441 4.85912C20.2448 5.10192 20.2966 5.36217 20.2966 5.625C20.2966 5.88783 20.2448 6.14808 20.1441 6.39088C20.0435 6.63368 19.896 6.85425 19.71 7.04L19.65 7.1C19.4195 7.33568 19.2648 7.63502 19.206 7.95941C19.1472 8.28381 19.1869 8.61838 19.32 8.92V9C19.4468 9.29577 19.6572 9.54802 19.9255 9.72569C20.1938 9.90337 20.5082 9.99872 20.83 10H21C21.5304 10 22.0391 10.2107 22.4142 10.5858C22.7893 10.9609 23 11.4696 23 12C23 12.5304 22.7893 13.0391 22.4142 13.4142C22.0391 13.7893 21.5304 14 21 14H20.91C20.5882 14.0013 20.2738 14.0966 20.0055 14.2743C19.7372 14.452 19.5268 14.7042 19.4 15Z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
          Настройки
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
              'cell--blocked': isCellBlocked(cellData.rowIndex, cellData.colIndex),
              'hint hint-mine': (room?.gameMode === 'training' || (room?.gameMode === 'fair' && gameState?.go)) && !cellData.cell.r && !cellData.cell.f && getCellHint(cellData.rowIndex, cellData.colIndex) === 'MINE',
              'hint hint-safe': (room?.gameMode === 'training' || (room?.gameMode === 'fair' && gameState?.go)) && !cellData.cell.r && !cellData.cell.f && getCellHint(cellData.rowIndex, cellData.colIndex) === 'SAFE',
              'hint hint-unknown': (room?.gameMode === 'training' || (room?.gameMode === 'fair' && gameState?.go)) && !cellData.cell.r && !cellData.cell.f && getCellHint(cellData.rowIndex, cellData.colIndex) === 'UNKNOWN',
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
          <svg v-else-if="cellData.cell.r && cellData.cell.m" class="cell-mine" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <!-- Основной корпус мины -->
            <circle cx="12" cy="14" r="7" fill="currentColor" stroke="#333" stroke-width="1"/>
            <!-- Сегменты мины -->
            <path d="M12 7L12 14M7 14L12 7M17 14L12 7" stroke="#333" stroke-width="1.5" stroke-linecap="round"/>
            <path d="M9 11L12 7L15 11M9 17L12 7L15 17" stroke="#333" stroke-width="1.5" stroke-linecap="round"/>
            <!-- Фитиль -->
            <line x1="12" y1="7" x2="12" y2="4" stroke="#8b5a3c" stroke-width="2" stroke-linecap="round"/>
            <circle cx="12" cy="3" r="1" fill="#ff6b6b"/>
            <!-- Центральная точка -->
            <circle cx="12" cy="14" r="2" fill="#333"/>
          </svg>
          <svg v-else-if="(gameState?.go || gameState?.gw) && cellData.cell.m && !cellData.cell.r" class="cell-mine" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <!-- Основной корпус мины -->
            <circle cx="12" cy="14" r="7" fill="currentColor" stroke="#333" stroke-width="1"/>
            <!-- Сегменты мины -->
            <path d="M12 7L12 14M7 14L12 7M17 14L12 7" stroke="#333" stroke-width="1.5" stroke-linecap="round"/>
            <path d="M9 11L12 7L15 11M9 17L12 7L15 17" stroke="#333" stroke-width="1.5" stroke-linecap="round"/>
            <!-- Фитиль -->
            <line x1="12" y1="7" x2="12" y2="4" stroke="#8b5a3c" stroke-width="2" stroke-linecap="round"/>
            <circle cx="12" cy="3" r="1" fill="#ff6b6b"/>
            <!-- Центральная точка -->
            <circle cx="12" cy="14" r="2" fill="#333"/>
          </svg>
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
        <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
          <path d="M1 12C1 12 5 4 12 4C19 4 23 12 23 12C23 12 19 20 12 20C5 20 1 12 1 12Z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
          <circle cx="12" cy="12" r="3" stroke="currentColor" stroke-width="2"/>
        </svg>
      </button>
      <h2>Игра окончена!</h2>
      <p v-if="gameState.ln">
        <router-link
          :to="`/profile/${gameState.ln}`"
          class="loser-link"
        >
          <strong>{{ gameState.ln }}</strong>
        </router-link> подорвался на мине
        <svg class="mine-inline-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
          <!-- Основной корпус мины -->
          <circle cx="12" cy="14" r="7" fill="currentColor" stroke="#333" stroke-width="1"/>
          <!-- Сегменты мины -->
          <path d="M12 7L12 14M7 14L12 7M17 14L12 7" stroke="#333" stroke-width="1.5" stroke-linecap="round"/>
          <path d="M9 11L12 7L15 11M9 17L12 7L15 17" stroke="#333" stroke-width="1.5" stroke-linecap="round"/>
          <!-- Фитиль -->
          <line x1="12" y1="7" x2="12" y2="4" stroke="#8b5a3c" stroke-width="2" stroke-linecap="round"/>
          <circle cx="12" cy="3" r="1" fill="#ff6b6b"/>
          <!-- Центральная точка -->
          <circle cx="12" cy="14" r="2" fill="#333"/>
        </svg>
      </p>
      <p v-else>
        Вы подорвались на мине
        <svg class="mine-inline-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
          <!-- Основной корпус мины -->
          <circle cx="12" cy="14" r="7" fill="currentColor" stroke="#333" stroke-width="1"/>
          <!-- Сегменты мины -->
          <path d="M12 7L12 14M7 14L12 7M17 14L12 7" stroke="#333" stroke-width="1.5" stroke-linecap="round"/>
          <path d="M9 11L12 7L15 11M9 17L12 7L15 17" stroke="#333" stroke-width="1.5" stroke-linecap="round"/>
          <!-- Фитиль -->
          <line x1="12" y1="7" x2="12" y2="4" stroke="#8b5a3c" stroke-width="2" stroke-linecap="round"/>
          <circle cx="12" cy="3" r="1" fill="#ff6b6b"/>
          <!-- Центральная точка -->
          <circle cx="12" cy="14" r="2" fill="#333"/>
        </svg>
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
        <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
          <path d="M1 12C1 12 5 4 12 4C19 4 23 12 23 12C23 12 19 20 12 20C5 20 1 12 1 12Z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
          <circle cx="12" cy="12" r="3" stroke="currentColor" stroke-width="2"/>
        </svg>
      </button>
      <h2>
        Победа!
        <svg class="celebration-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
          <path d="M12 2L15.09 8.26L22 9.27L17 14.14L18.18 21.02L12 17.77L5.82 21.02L7 14.14L2 9.27L8.91 8.26L12 2Z" fill="currentColor"/>
        </svg>
      </h2>
      <p>Все мины найдены!</p>
      <div v-if="ratingChange !== null" class="rating-change">
        <div class="rating-change__positive">
          Рейтинг за игру: {{ Math.round(ratingChange) }}
        </div>
        <div class="rating-change__note">
          Ваш рейтинг обновится, если это значение больше текущего
        </div>
      </div>
      <div v-else-if="gameState && gameStartTime" class="rating-change">
        <div class="rating-change__hint">
          Игра не дает рейтинг (плотность мин &lt; 10%)
        </div>
      </div>
      <button @click="handleNewGame" class="game-message__button">
        Новая игра
      </button>
    </div>
  </article>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed, watch } from 'vue'
import type { WebSocketMessage, IWebSocketClient } from '@/api/websocket'
import { useCursorAnimation } from '@/composables/useCursorAnimation'
import { useGameBoardZoom } from '@/composables/useGameBoardZoom'
import { useCellTouch } from '@/composables/useCellTouch'
import { useAuthStore } from '@/stores/auth'
import { calculateDifficulty, calculateGameRating, isRatingEligible } from '@/utils/ratingCalculator'
import Chat from '@/components/Chat.vue'

const props = defineProps<{
  wsClient: IWebSocketClient | null
  nickname: string
  room?: { id: string; creatorId?: number; gameMode?: string; chording?: boolean; quickStart?: boolean } | null
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
const currentGameTime = ref<number>(0) // Текущее время игры в секундах
const ratingUpdateInterval = ref<ReturnType<typeof setInterval> | null>(null) // Интервал для обновления рейтинга

// Вычисляем текущий рейтинг на основе времени игры
const currentRating = computed(() => {
  if (!gameState.value || !gameStartTime.value || gameState.value.go || gameState.value.gw) {
    return null
  }

  const gameTime = currentGameTime.value
  const chording = props.room?.chording ?? false
  const quickStart = props.room?.quickStart ?? false

  // Проверяем, может ли игра дать рейтинг
  if (!isRatingEligible(gameState.value.c, gameState.value.r, gameState.value.m, gameTime)) {
    return null
  }

  return calculateGameRating(gameState.value.c, gameState.value.r, gameState.value.m, gameTime, chording, quickStart)
})

// Функция для обновления времени игры
const updateGameTime = () => {
  if (gameStartTime.value && gameState.value && !gameState.value.go && !gameState.value.gw) {
    currentGameTime.value = (Date.now() - gameStartTime.value) / 1000
  }
}

// Запускаем обновление времени каждую секунду
const startRatingUpdate = () => {
  if (ratingUpdateInterval.value) {
    clearInterval(ratingUpdateInterval.value)
  }
  updateGameTime() // Обновляем сразу
  ratingUpdateInterval.value = setInterval(() => {
    updateGameTime()
  }, 10) // Обновляем каждую секунду
}

// Останавливаем обновление
const stopRatingUpdate = () => {
  if (ratingUpdateInterval.value) {
    clearInterval(ratingUpdateInterval.value)
    ratingUpdateInterval.value = null
  }
}

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

  // Проверка: нельзя открыть ячейку с флагом по левому клику
  if (!isRightClick && gameState.value?.b?.[row]?.[col]?.f) {
    return
  }

  // Блокируем клики на непомеченные ячейки в режиме без угадываний
  if (!isRightClick && isCellBlocked(row, col)) {
    return
  }

  // Запоминаем время начала игры при первом клике
  if (gameStartTime.value === null && !isRightClick) {
    gameStartTime.value = Date.now()
    currentGameTime.value = 0
    startRatingUpdate()
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

// Получаем тип подсказки для ячейки (показывается всегда в fairMode)
const getCellHint = (row: number, col: number): string | null => {
  if (!gameState.value?.hints) return null
  const hint = gameState.value.hints.find((h: { r: number; c: number; t: string }) => h.r === row && h.c === col)
  return hint ? hint.t : null
}

// В fairMode не блокируем клики - игра сама выберет худший сценарий
const isCellBlocked = (row: number, col: number): boolean => {
  return false
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
  currentGameTime.value = 0
  stopRatingUpdate()
  // Список игроков не сбрасываем, так как они остаются в комнате
}

const handleMessage = (msg: WebSocketMessage) => {
  const timestamp = new Date().toISOString()
  console.log(`[GAME MSG ${timestamp}] Получено сообщение:`, {
    type: msg.type,
    data: msg
  })

  if (msg.type === 'gameState' && msg.gameState) {
    console.log(`[GAME MSG ${timestamp}] Обработка gameState:`, {
      rows: msg.gameState.r,
      cols: msg.gameState.c,
      mines: msg.gameState.m,
      revealed: msg.gameState.rv,
      gameOver: msg.gameState.go,
      gameWon: msg.gameState.gw,
      hintsUsed: msg.gameState.hu
    })
    const prevGameWon = gameState.value?.gw
    const prevGameOver = gameState.value?.go
    gameState.value = msg.gameState

    // Если игра только что завершилась победой, рассчитываем изменение рейтинга
    if (msg.gameState.gw && !prevGameWon && gameStartTime.value !== null && gameState.value) {
      stopRatingUpdate() // Останавливаем обновление рейтинга
      const gameTime = currentGameTime.value // Используем уже вычисленное время
      const chording = props.room?.chording ?? false
      const quickStart = props.room?.quickStart ?? false

      // Проверяем, может ли игра дать рейтинг
      if (isRatingEligible(gameState.value.c, gameState.value.r, gameState.value.m, gameTime)) {
        ratingChange.value = calculateGameRating(gameState.value.c, gameState.value.r, gameState.value.m, gameTime, chording, quickStart)
      } else {
        ratingChange.value = null // Игра не дает рейтинг (плотность < 10%)
      }
    }

    // Если игра окончена (поражение), останавливаем обновление рейтинга
    if (msg.gameState.go && !prevGameOver && gameStartTime.value !== null) {
      stopRatingUpdate()
    }

    // Сбрасываем время начала игры при новой игре
    // Подсказки сбрасываются на сервере при создании новой игры
    if (!msg.gameState.gw && !msg.gameState.go && gameState.value.rv === 0) {
      gameStartTime.value = null
      ratingChange.value = null
    }
  } else if (msg.type === 'cellUpdate' && msg.cellUpdates && gameState.value) {
    // Обрабатываем обновления клеток
    console.log(`[GAME MSG ${timestamp}] Обработка cellUpdate:`, {
      updatesCount: msg.cellUpdates.length,
      gameOver: msg.gameOver,
      gameWon: msg.gameWon,
      revealed: msg.revealed,
      hintsUsed: msg.hintsUsed,
      updates: msg.cellUpdates.slice(0, 10) // Показываем первые 10 обновлений
    })
    const prevGameWon = gameState.value.gw
    const CellTypeClosed = 255  // Закрыта (изменено с 0 на 255)
    const CellTypeMine = 9
    const CellTypeSafe = 10
    const CellTypeUnknown = 11
    const CellTypeDanger = 12

    // Обновляем клетки
    for (const update of msg.cellUpdates) {
      const { row, col, type } = update
      if (gameState.value.b && gameState.value.b[row] && gameState.value.b[row][col]) {
        const cell = gameState.value.b[row][col]

        // Логика определения типа клетки:
        // - type 0-8: открытая клетка с количеством соседних мин (0-8)
        // - type 9: мина
        // - type 10-12: подсказки для режима обучения (закрытая клетка)
        // - type 255: закрытая клетка

        if (type >= 0 && type <= 8) {
          // Открытая клетка с количеством соседних мин (0-8)
          cell.r = true
          cell.m = false
          cell.n = type
        } else if (type === CellTypeMine) {
          // Мина (9)
          cell.r = true
          cell.m = true
          cell.n = 0
        } else if (type === CellTypeClosed) {
          // Закрытая клетка (255)
          cell.r = false
          cell.f = false
          cell.m = false
          cell.n = 0
        } else if (type === CellTypeSafe || type === CellTypeUnknown || type === CellTypeDanger) {
          // Подсказки для режима обучения - клетка остается закрытой
          cell.r = false
          cell.f = false
        } else {
          // Fallback: если type не распознан, считаем закрытой
          console.warn(`[GAME MSG] Неизвестный тип клетки: ${type} для клетки [${row}, ${col}]`)
          cell.r = false
          cell.f = false
          cell.m = false
          cell.n = 0
        }
      }
    }

    // Обновляем метаданные игры
    if (msg.gameOver !== undefined) {
      const prevGameOver = gameState.value?.go
      gameState.value.go = msg.gameOver
      // Если игра только что окончена (поражение), останавливаем обновление рейтинга
      if (msg.gameOver && !prevGameOver && gameStartTime.value !== null) {
        stopRatingUpdate()
      }
    }
    if (msg.gameWon !== undefined) {
      gameState.value.gw = msg.gameWon
    }
    if (msg.revealed !== undefined) {
      gameState.value.rv = msg.revealed
    }
    if (msg.hintsUsed !== undefined) {
      gameState.value.hu = msg.hintsUsed
    }
    if (msg.loserPlayerId !== undefined) {
      gameState.value.lpid = msg.loserPlayerId
    }
    if (msg.loserNickname !== undefined) {
      gameState.value.ln = msg.loserNickname
    }

    // Если игра только что завершилась победой, рассчитываем изменение рейтинга
    if (msg.gameWon && !prevGameWon && gameStartTime.value !== null && gameState.value) {
      const gameTime = (Date.now() - gameStartTime.value) / 1000
      const chording = props.room?.chording ?? false
      const quickStart = props.room?.quickStart ?? false
      // Проверяем, может ли игра дать рейтинг
      if (isRatingEligible(gameState.value.c, gameState.value.r, gameState.value.m, gameTime)) {
        ratingChange.value = calculateGameRating(gameState.value.c, gameState.value.r, gameState.value.m, gameTime, chording, quickStart)
      } else {
        ratingChange.value = null // Игра не дает рейтинг (плотность < 10%)
      }
    }
  } else if (msg.type === 'cursor' && msg.cursor) {
    console.log(`[GAME MSG ${timestamp}] Обработка cursor:`, {
      playerId: msg.playerId || msg.cursor.pid,
      x: msg.cursor.x,
      y: msg.cursor.y
    })
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
    console.log(`[GAME MSG ${timestamp}] Обработка players:`, {
      count: msg.players.length,
      players: msg.players
    })
    // Обновляем список игроков
    playersList.value = msg.players.map((p: any) => ({
      id: p.id || p.playerId || '',
      nickname: p.nickname || 'Игрок',
      color: p.color || '#667eea'
    }))
  } else if (msg.type === 'chat') {
    console.log(`[GAME MSG ${timestamp}] Обработка chat:`, {
      text: msg.chat?.text,
      isSystem: msg.chat?.isSystem,
      action: msg.chat?.action,
      playerId: msg.playerId,
      nickname: msg.nickname
    })
  } else {
    console.log(`[GAME MSG ${timestamp}] Необработанный тип сообщения:`, msg.type)
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
  stopRatingUpdate() // Очищаем интервал при размонтировании компонента
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
  width: 1.25rem;
  height: 1.25rem;
  flex-shrink: 0;
}

[data-theme="dark"] .cell-mine circle[fill="currentColor"] {
  fill: #dc2626;
}

[data-theme="dark"] .cell-mine path[stroke="#333"],
[data-theme="dark"] .cell-mine line[stroke="#333"],
[data-theme="dark"] .cell-mine circle[stroke="#333"],
[data-theme="dark"] .cell-mine circle[fill="#333"] {
  stroke: #fff;
  fill: #fff;
}

.hint-button-icon,
.edit-room-button-icon {
  width: 1rem;
  height: 1rem;
  display: inline-block;
  vertical-align: middle;
  margin-right: 0.5rem;
}

.mine-inline-icon {
  width: 1rem;
  height: 1rem;
  display: inline-block;
  vertical-align: middle;
  margin-left: 0.25rem;
}

.celebration-icon {
  width: 1.5rem;
  height: 1.5rem;
  display: inline-block;
  vertical-align: middle;
  margin-left: 0.5rem;
  color: #fbbf24;
}

.game-message__transparency-button svg {
  width: 1.25rem;
  height: 1.25rem;
  filter: invert(1);
}

.cell-flag {
  display: inline-block;
  vertical-align: middle;
  --flag-color: #dc2626;
}

.cell-flag .flag-cloth {
  fill: var(--flag-color, #dc2626);
}

/* Подсказки для ячеек (показываются при проигрыше в fairMode) */
.cell.hint {
  position: relative;
}

.cell.hint::before {
  content: "";
  position: absolute;
  top: 3px;
  left: 3px;
  bottom: 3px;
  right: 3px;
  border: 2px solid;
  line-height: 18px;
  font-weight: bold;
  font-size: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.cell.hint-safe::before {
  content: ".";
  color: #9c9;
  border-color: #9c9;
}

.cell.hint-unknown::before {
  content: "?";
  color: #da0;
  border-color: #da0;
}

.cell.hint-mine::before {
  content: "!";
  color: #e77;
  border-color: #e77;
}

[data-theme="dark"] .cell.hint-safe::before {
  color: #22c55e;
  border-color: #22c55e;
}

[data-theme="dark"] .cell.hint-unknown::before {
  color: #fbbf24;
  border-color: #fbbf24;
}

[data-theme="dark"] .cell.hint-mine::before {
  color: #ef4444;
  border-color: #ef4444;
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

  .info-item--rating {
    color: #667eea;
    font-weight: 600;
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

