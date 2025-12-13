<template>
  <div v-if="show" class="modal-overlay" @click.self="handleOverlayClick">
    <div class="modal">
      <h2 class="modal__title">Создать комнату</h2>

      <div class="modal__form">
        <div class="form-group">
          <label class="form-label">Название комнаты</label>
          <div class="form-input-wrapper">
            <input
              v-model="form.name"
              type="text"
              class="form-input"
              placeholder="Название комнаты"
              maxlength="30"
            />
            <button
              type="button"
              @click="generateRoomName"
              class="form-input-button"
              title="Сгенерировать случайное название"
            >
              🎲
            </button>
          </div>
        </div>

        <div class="form-group">
          <label class="form-label">Размер поля</label>
          <div class="form-row">
            <div class="form-col">
              <label class="form-label-small">
                Строки: <span class="range-value">{{ form.rows }}</span>
              </label>
              <input
                v-model.number="form.rows"
                type="range"
                class="form-range"
                min="5"
                max="50"
                step="1"
              />
            </div>
            <div class="form-col">
              <label class="form-label-small">
                Столбцы: <span class="range-value">{{ form.cols }}</span>
              </label>
              <input
                v-model.number="form.cols"
                type="range"
                class="form-range"
                min="5"
                max="50"
                step="1"
              />
            </div>
          </div>
        </div>

        <div class="form-group">
          <label class="form-label">Количество мин</label>
          <input
            v-model.number="form.mines"
            type="number"
            class="form-input"
            :min="1"
            :max="maxMines"
          />
          <div class="form-hint">Максимум: {{ maxMines }}</div>
          <div class="difficulty-info">
            <span class="difficulty-label">Сложность поля:</span>
            <span class="difficulty-value">{{ difficulty.toFixed(2) }}</span>
          </div>
        </div>

        <div class="form-group rating-status" :class="{ 'rating-status--rated': isRatedGame, 'rating-status--unrated': !isRatedGame }">
          <div class="rating-status__icon">
            <span v-if="isRatedGame">⭐</span>
            <span v-else>⚪</span>
          </div>
          <div class="rating-status__content">
            <div class="rating-status__label">
              {{ isRatedGame ? 'Рейтинговая игра' : 'Нерейтинговая игра' }}
            </div>
            <div v-if="isRatedGame && maxRatingGain > 0" class="rating-status__gain">
              До +{{ Math.round(maxRatingGain) }} очков попытки
            </div>
            <div v-else-if="!isRatedGame" class="rating-status__hint">
              Поле слишком простое для получения рейтинга
            </div>
          </div>
        </div>

        <div class="form-group">
          <label class="form-label">
            <input
              v-model="hasPassword"
              type="checkbox"
              class="form-checkbox"
            />
            Защитить паролем
          </label>
          <input
            v-if="hasPassword"
            v-model="form.password"
            type="password"
            class="form-input"
            placeholder="Пароль"
            maxlength="20"
          />
        </div>

        <div class="form-group">
          <label class="form-label">
            <input
              v-model="form.quickStart"
              type="checkbox"
              class="form-checkbox"
            />
            Быстрый старт
          </label>
          <div class="form-hint">Первая кликнутая клетка всегда будет нулевой (без мин вокруг)</div>
        </div>

        <div class="form-group">
          <label class="form-label">
            <input
              v-model="form.chording"
              type="checkbox"
              class="form-checkbox"
            />
            Chording
          </label>
          <div class="form-hint">Клик на открытую клетку с цифрой открывает соседние клетки, если вокруг стоит нужное количество флагов</div>
        </div>

        <div class="form-group">
          <label class="form-label">Режим игры</label>
          <div class="game-mode-selector">
            <label class="game-mode-option" :class="{ 'game-mode-option--active': form.gameMode === 'classic' }">
              <input
                v-model="form.gameMode"
                type="radio"
                value="classic"
                class="game-mode-radio"
              />
              <div class="game-mode-content">
                <div class="game-mode-title">Классический</div>
                <div class="game-mode-description">Обычный режим сапера с заранее размещенными минами</div>
              </div>
            </label>
            <!-- <label class="game-mode-option" :class="{ 'game-mode-option--active': form.gameMode === 'training' }">
              <input
                v-model="form.gameMode"
                type="radio"
                value="training"
                class="game-mode-radio"
              />
              <div class="game-mode-content">
                <div class="game-mode-title">Обучение</div>
                <div class="game-mode-description">Режим с подсказками на границе для изучения логики игры</div>
              </div>
            </label>
            <label class="game-mode-option" :class="{ 'game-mode-option--active': form.gameMode === 'fair' }">
              <input
                v-model="form.gameMode"
                type="radio"
                value="fair"
                class="game-mode-radio"
              />
              <div class="game-mode-content">
                <div class="game-mode-title">Справедливый</div>
                <div class="game-mode-description">Мины размещаются динамически, игра всегда выбирает худший сценарий</div>
              </div>
            </label> -->
          </div>
        </div>

        <div v-if="error" class="form-error">{{ error }}</div>

        <div class="modal__actions">
          <button @click="handleCancel" class="btn btn-secondary">Отмена</button>
          <button @click="handleSubmit" class="btn btn-primary" :disabled="!isValid">
            Создать
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { generateRandomName } from '@/utils/nameGenerator'
import { calculateMaxRatingGain, isComplexitySufficient, calculateDifficulty } from '@/utils/ratingCalculator'
import { useAuthStore } from '@/stores/auth'

const props = defineProps<{
  show: boolean
}>()

const emit = defineEmits<{
  submit: [data: { name: string; password?: string; rows: number; cols: number; mines: number; gameMode: string; quickStart: boolean; chording: boolean }]
  cancel: []
}>()

const form = ref({
  name: '',
  rows: 16,
  cols: 16,
  mines: 40,
  password: '',
  gameMode: 'classic' as 'classic' | 'training' | 'fair',
  quickStart: false,
  chording: false,
})

const hasPassword = ref(false)
const error = ref('')
const authStore = useAuthStore()

// Генерируем случайное название при открытии модалки
watch(() => props.show, (isShowing) => {
  if (isShowing && !form.value.name.trim()) {
    form.value.name = generateRandomName()
  }
})

onMounted(() => {
  // Генерируем случайное название при монтировании, если поле пустое
  if (!form.value.name.trim()) {
    form.value.name = generateRandomName()
  }
})

const maxMines = computed(() => {
  return form.value.rows * form.value.cols - 15
})

const difficulty = computed(() => {
  return calculateDifficulty(form.value.cols, form.value.rows, form.value.mines)
})

const isRatedGame = computed(() => {
  return isComplexitySufficient(
    form.value.cols,
    form.value.rows,
    form.value.mines
  )
})

const maxRatingGain = computed(() => {
  if (!isRatedGame.value) return 0
  return calculateMaxRatingGain(
    form.value.cols,
    form.value.rows,
    form.value.mines
  )
})

const isValid = computed(() => {
  return (
    form.value.name.trim().length > 0 &&
    form.value.rows >= 5 &&
    form.value.rows <= 50 &&
    form.value.cols >= 5 &&
    form.value.cols <= 50 &&
    form.value.mines >= 1 &&
    form.value.mines <= maxMines.value
  )
})

const handleSubmit = () => {
  if (!isValid.value) {
    error.value = 'Заполните все поля корректно'
    return
  }

  error.value = ''

  const data = {
    name: form.value.name.trim(),
    rows: form.value.rows,
    cols: form.value.cols,
    mines: form.value.mines,
    gameMode: form.value.gameMode,
    quickStart: form.value.quickStart,
    chording: form.value.chording,
    ...(hasPassword.value && form.value.password ? { password: form.value.password } : {}),
  }

  emit('submit', data)
  error.value = ''
}

const generateRoomName = () => {
  form.value.name = generateRandomName()
}

const handleCancel = () => {
  emit('cancel')
  error.value = ''
  form.value = {
    name: '',
    rows: 16,
    cols: 16,
    mines: 40,
    password: '',
    gameMode: 'classic' as 'classic' | 'training' | 'fair',
    quickStart: false,
    chording: false,
  }
  hasPassword.value = false
}

const handleOverlayClick = () => {
  // Не закрываем при клике на overlay
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  backdrop-filter: blur(4px);
}

.modal {
  background: var(--bg-primary);
  padding: 2.5rem;
  border-radius: 1rem;
  box-shadow: 0 20px 60px var(--shadow);
  min-width: 500px;
  max-width: 90vw;
  max-height: 90vh;
  overflow-y: auto;
  animation: slideIn 0.3s ease-out;
}

@media (max-width: 768px) {
  .modal {
    min-width: auto;
    width: 95vw;
    max-width: 95vw;
    padding: 1.5rem;
    margin: 1rem;
  }
}

@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateY(-20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.modal__title {
  margin: 0 0 1.5rem 0;
  font-size: 1.5rem;
  color: var(--text-primary);
  text-align: center;
}

.modal__form {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.form-label {
  font-weight: 600;
  color: var(--text-primary);
  font-size: 0.875rem;
}

.form-label-small {
  font-weight: 500;
  color: var(--text-secondary);
  font-size: 0.75rem;
  margin-bottom: 0.5rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.range-value {
  font-weight: 600;
  color: #667eea;
  font-size: 0.875rem;
}

.form-input-wrapper {
  position: relative;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.form-input {
  flex: 1;
  padding: 0.75rem;
  font-size: 1rem;
  border: 2px solid var(--border-color);
  border-radius: 0.5rem;
  background: var(--bg-secondary);
  color: var(--text-primary);
  transition: border-color 0.2s;
  box-sizing: border-box;
}

.form-input:focus {
  outline: none;
  border-color: #667eea;
}

.form-input-button {
  flex-shrink: 0;
  width: 2.5rem;
  height: 2.5rem;
  padding: 0;
  border: 2px solid var(--border-color);
  border-radius: 0.5rem;
  background: var(--bg-secondary);
  color: var(--text-primary);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.25rem;
  transition: all 0.2s;
  box-sizing: border-box;
}

.form-input-button:hover {
  background: var(--bg-tertiary);
  border-color: #667eea;
  transform: scale(1.05);
}

.form-input-button:active {
  transform: scale(0.95);
}

.form-row {
  display: flex;
  gap: 1rem;
}

.form-col {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.form-checkbox {
  margin-right: 0.5rem;
}

.game-mode-selector {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.game-mode-option {
  display: flex;
  align-items: flex-start;
  padding: 1rem;
  border: 2px solid var(--border-color);
  border-radius: 0.75rem;
  background: var(--bg-secondary);
  cursor: pointer;
  transition: all 0.2s;
  gap: 0.75rem;
}

.game-mode-option:hover {
  border-color: #667eea;
  background: var(--bg-tertiary);
}

.game-mode-option--active {
  border-color: #667eea;
  background: var(--bg-tertiary);
  box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
}

.game-mode-radio {
  margin-top: 0.125rem;
  cursor: pointer;
}

.game-mode-content {
  flex: 1;
}

.game-mode-title {
  font-weight: 600;
  font-size: 1rem;
  color: var(--text-primary);
  margin-bottom: 0.25rem;
}

.game-mode-description {
  font-size: 0.875rem;
  color: var(--text-secondary);
  line-height: 1.4;
}

.form-hint {
  font-size: 0.75rem;
  color: var(--text-secondary);
}

.difficulty-info {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-top: 0.5rem;
  padding: 0.5rem;
  background: var(--bg-tertiary);
  border-radius: 0.5rem;
}

.difficulty-label {
  font-size: 0.875rem;
  color: var(--text-secondary);
  font-weight: 500;
}

.difficulty-value {
  font-size: 1rem;
  color: #667eea;
  font-weight: 700;
}

.form-range {
  width: 100%;
  height: 8px;
  border-radius: 4px;
  background: var(--bg-tertiary);
  outline: none;
  -webkit-appearance: none;
  appearance: none;
  cursor: pointer;
  margin: 0.5rem 0;
}

.form-range::-webkit-slider-thumb {
  -webkit-appearance: none;
  appearance: none;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  cursor: pointer;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
  transition: all 0.2s ease-in-out;
}

.form-range::-webkit-slider-thumb:hover {
  transform: scale(1.1);
  box-shadow: 0 4px 8px rgba(102, 126, 234, 0.4);
}

.form-range::-moz-range-thumb {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  cursor: pointer;
  border: none;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
  transition: all 0.2s ease-in-out;
}

.form-range::-moz-range-thumb:hover {
  transform: scale(1.1);
  box-shadow: 0 4px 8px rgba(102, 126, 234, 0.4);
}

.form-range:focus {
  outline: none;
}

.form-range:focus::-webkit-slider-thumb {
  box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.2);
}

.form-range:focus::-moz-range-thumb {
  box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.2);
}

.rating-status {
  padding: 1rem;
  border-radius: 0.5rem;
  border: 2px solid;
  display: flex;
  align-items: center;
  gap: 1rem;
}

.rating-status--rated {
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.1) 0%, rgba(118, 75, 162, 0.1) 100%);
  border-color: rgba(102, 126, 234, 0.3);
}

.rating-status--unrated {
  background: rgba(107, 114, 128, 0.1);
  border-color: rgba(107, 114, 128, 0.3);
}

.rating-status__icon {
  font-size: 2rem;
  line-height: 1;
  flex-shrink: 0;
}

.rating-status__content {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.rating-status__label {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--text-primary);
}

.rating-status--rated .rating-status__label {
  color: #667eea;
}

.rating-status--unrated .rating-status__label {
  color: var(--text-secondary);
}

.rating-status__gain {
  font-size: 0.875rem;
  color: #22c55e;
  font-weight: 500;
}

.rating-status__hint {
  font-size: 0.75rem;
  color: var(--text-secondary);
  font-style: italic;
}

.form-error {
  padding: 0.75rem;
  background: #fee2e2;
  color: #dc2626;
  border-radius: 0.5rem;
  font-size: 0.875rem;
}

.modal__actions {
  display: flex;
  gap: 1rem;
  margin-top: 1rem;
}

.btn {
  flex: 1;
  padding: 0.875rem;
  font-size: 1rem;
  font-weight: 600;
  border: none;
  border-radius: 0.5rem;
  cursor: pointer;
  transition: transform 0.2s, box-shadow 0.2s;
}

.btn-primary {
  color: white;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.btn-primary:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
}

.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-secondary {
  background: var(--bg-tertiary);
  color: var(--text-primary);
}

.btn-secondary:hover {
  background: var(--border-color);
}

@media (max-width: 768px) {
  .modal__title {
    font-size: 1.25rem;
    margin-bottom: 1rem;
  }

  .modal__form {
    gap: 1rem;
  }

  .form-row {
    gap: 0.75rem;
  }

  .form-input {
    padding: 0.625rem;
    font-size: 0.9375rem;
  }

  .modal__actions {
    flex-direction: column;
    gap: 0.75rem;
  }

  .btn {
    width: 100%;
  }
}

@media (max-width: 480px) {
  .modal {
    padding: 1rem;
    border-radius: 0.75rem;
  }

  .modal__title {
    font-size: 1.125rem;
  }

  .form-input {
    padding: 0.5rem;
    font-size: 0.875rem;
  }

  .form-input-button {
    width: 2rem;
    height: 2rem;
    font-size: 1rem;
  }
}
</style>

