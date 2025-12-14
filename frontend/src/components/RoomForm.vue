<template>
  <div class="room-form">
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
          <IconDice />
        </button>
      </div>
    </div>

    <div class="form-group">
      <label class="form-label">Сложность</label>
      <div class="difficulty-templates">
        <button
          type="button"
          @click="applyTemplate('easy')"
          class="difficulty-template"
          :class="{ 'difficulty-template--active': currentTemplate === 'easy' }"
        >
          <div class="difficulty-template__title">Легкий</div>
          <div class="difficulty-template__params">9×9, 10 мин</div>
        </button>
        <button
          type="button"
          @click="applyTemplate('medium')"
          class="difficulty-template"
          :class="{ 'difficulty-template--active': currentTemplate === 'medium' }"
        >
          <div class="difficulty-template__title">Нормальный</div>
          <div class="difficulty-template__params">16×16, 40 мин</div>
        </button>
        <button
          type="button"
          @click="applyTemplate('hard')"
          class="difficulty-template"
          :class="{ 'difficulty-template--active': currentTemplate === 'hard' }"
        >
          <div class="difficulty-template__title">Сложный</div>
          <div class="difficulty-template__params">16×30, 99 мин</div>
        </button>
        <button
          type="button"
          @click="applyTemplate('custom')"
          class="difficulty-template"
          :class="{ 'difficulty-template--active': currentTemplate === 'custom' }"
        >
          <div class="difficulty-template__title">Свой</div>
          <div class="difficulty-template__params">Настроить вручную</div>
        </button>
      </div>
    </div>

    <div v-if="currentTemplate === 'custom'" class="form-group">
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
            @input="checkTemplate"
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
            @input="checkTemplate"
          />
        </div>
      </div>
    </div>

    <div v-if="currentTemplate === 'custom'" class="form-group">
      <label class="form-label">
        Количество мин: <span class="range-value">{{ form.mines }}</span>
      </label>
      <input
        v-model.number="form.mines"
        type="range"
        class="form-range"
        :min="1"
        :max="maxMines"
        step="1"
        @input="checkTemplate"
      />
      <div class="form-hint">Максимум: {{ maxMines }}</div>
      <div class="difficulty-info">
        <span class="difficulty-label">Сложность поля:</span>
        <span class="difficulty-value">{{ difficulty.toFixed(2) }}</span>
      </div>
    </div>

    <div v-else class="form-group">
      <!-- <div class="difficulty-info">
        <span class="difficulty-label">Текущие параметры:</span>
        <span class="difficulty-value">{{ form.rows }}×{{ form.cols }}, {{ form.mines }} мин</span>
      </div> -->
      <div class="difficulty-info">
        <span class="difficulty-label">Сложность поля:</span>
        <span class="difficulty-value">{{ difficulty.toFixed(2) }}</span>
      </div>
    </div>

    <div class="form-group rating-status" :class="{ 'rating-status--rated': isRatedGame, 'rating-status--unrated': !isRatedGame }">
      <div class="rating-status__icon">
        <IconStar v-if="isRatedGame" class="rating-status-icon" />
        <IconCircle v-else class="rating-status-icon" />
      </div>
      <div class="rating-status__content">
        <div class="rating-status__label">
          {{ isRatedGame ? 'Рейтинговая игра' : 'Нерейтинговая игра' }}
        </div>
        <div v-if="isRatedGame && maxRatingGain > 0" class="rating-status__gain">
          Макс. рейтинг: {{ Math.round(maxRatingGain) }}
        </div>
        <div v-else-if="!isRatedGame && form.seed != null && form.seed !== ''" class="rating-status__hint">
          Указан seed - игра нерейтинговая
        </div>
        <div v-else-if="!isRatedGame" class="rating-status__hint">
          Плотность мин &lt; 10% (минимальное требование для рейтинга)
        </div>
      </div>
    </div>

    <div class="form-group">
      <label class="form-label">Пароль (опционально)</label>
      <input
        v-model="form.password"
        type="password"
        class="form-input"
        placeholder="Оставьте пустым, чтобы не защищать паролем"
        maxlength="20"
      />
      <div class="form-hint">Пароль будет установлен только если поле заполнено</div>
    </div>

    <div class="form-group">
      <button
        type="button"
        @click="showAdvanced = !showAdvanced"
        class="advanced-toggle"
        :class="{ 'advanced-toggle--open': showAdvanced }"
      >
        <span class="advanced-toggle__icon">{{ showAdvanced ? '▼' : '▶' }}</span>
        <span class="advanced-toggle__text">Продвинутые настройки</span>
      </button>

      <div v-if="showAdvanced" class="advanced-options">
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
          <label class="form-label">Seed (UUID, опционально)</label>
          <input
            v-model="form.seed"
            type="text"
            class="form-input"
            placeholder="Оставьте пустым для случайной генерации"
            pattern="[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}"
          />
          <div class="form-hint">
            Укажите seed для воспроизводимой генерации поля. Если указан - игра будет нерейтинговой.
          </div>
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
        <label v-if="showAllGameModes" class="game-mode-option" :class="{ 'game-mode-option--active': form.gameMode === 'training' }">
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
        <label v-if="showAllGameModes" class="game-mode-option" :class="{ 'game-mode-option--active': form.gameMode === 'fair' }">
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
        </label>
      </div>
    </div>
      </div>
    </div>



    <slot name="warning"></slot>

    <div v-if="error" class="form-error">{{ error }}</div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { generateRandomName } from '@/utils/nameGenerator'
import { calculateMaxRating, isRatingEligible, calculateDifficulty } from '@/utils/ratingCalculator'
import IconDice from '@/components/icons/IconDice.vue'
import IconStar from '@/components/icons/IconStar.vue'
import IconCircle from '@/components/icons/IconCircle.vue'

export interface RoomFormData {
  name: string
  rows: number
  cols: number
  mines: number
  password: string
  gameMode: 'classic' | 'training' | 'fair'
  quickStart: boolean
  chording: boolean
  seed?: string | null
}

const props = defineProps<{
  modelValue: RoomFormData
  hasPassword?: boolean // Оставляем для обратной совместимости, но не используем
  error?: string
  showAdvancedOptions?: boolean // Больше не используется, оставлено для обратной совместимости
  showAllGameModes?: boolean
  autoGenerateName?: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: RoomFormData]
  'update:hasPassword': [value: boolean] // Оставляем для обратной совместимости
  'generate-name': []
}>()

const form = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const showAdvanced = ref(false)

const error = computed(() => props.error)

// Генерируем случайное название при необходимости
watch(() => props.autoGenerateName, (shouldGenerate) => {
  if (shouldGenerate && !form.value.name.trim()) {
    generateRoomName()
  }
}, { immediate: true })

// Проверяем шаблон при инициализации
watch(() => form.value, () => {
  checkTemplate()
}, { immediate: true, deep: true })

const maxMines = computed(() => {
  return form.value.rows * form.value.cols - 15
})

const difficulty = computed(() => {
  return calculateDifficulty(form.value.cols, form.value.rows, form.value.mines)
})

// Проверяем, может ли игра дать рейтинг (проверяем только плотность, время проверится при завершении)
// Если указан seed - игра нерейтинговая
const isRatedGame = computed(() => {
  // Если указан seed - игра нерейтинговая
  if (form.value.seed != null && form.value.seed !== '') return false
  // Проверяем минимальную плотность мин (10%)
  return isRatingEligible(form.value.cols, form.value.rows, form.value.mines, 20)
})

// Максимальный рейтинг при минимальном времени (0.1 секунды для расчета)
const maxRatingGain = computed(() => {
  if (!isRatedGame.value) return 0
  return calculateMaxRating(
    form.value.cols,
    form.value.rows,
    form.value.mines,
    form.value.chording,
    form.value.quickStart
  )
})

const generateRoomName = () => {
  form.value = { ...form.value, name: generateRandomName() }
  emit('generate-name')
}

// Шаблоны сложности
const templates = {
  easy: { rows: 9, cols: 9, mines: 10 },
  medium: { rows: 16, cols: 16, mines: 40 },
  hard: { rows: 16, cols: 30, mines: 99 },
}

const currentTemplate = ref<'easy' | 'medium' | 'hard' | 'custom'>('medium')

// Применяем шаблон
const applyTemplate = (template: 'easy' | 'medium' | 'hard' | 'custom') => {
  if (template === 'custom') {
    currentTemplate.value = 'custom'
    return
  }

  const templateData = templates[template]
  form.value = {
    ...form.value,
    rows: templateData.rows,
    cols: templateData.cols,
    mines: templateData.mines,
  }
  currentTemplate.value = template
}

// Проверяем, соответствует ли текущая конфигурация какому-либо шаблону
const checkTemplate = () => {
  const { rows, cols, mines } = form.value

  for (const [key, template] of Object.entries(templates)) {
    if (template.rows === rows && template.cols === cols && template.mines === mines) {
      currentTemplate.value = key as 'easy' | 'medium' | 'hard'
      return
    }
  }

  currentTemplate.value = 'custom'
}
</script>

<style scoped>
.room-form {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  width: 100%;
  min-width: 0;
  max-width: 100%;
  box-sizing: border-box;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  min-width: 0;
  max-width: 100%;
  box-sizing: border-box;
}

.form-label {
  font-weight: 600;
  color: var(--text-primary);
  font-size: 0.875rem;
  word-wrap: break-word;
  overflow-wrap: break-word;
  min-width: 0;
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
  transition: all 0.2s;
  box-sizing: border-box;
}

.form-input-button svg {
  width: 1.25rem;
  height: 1.25rem;
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

.advanced-toggle {
  width: 100%;
  padding: 0.75rem 1rem;
  background: var(--bg-secondary);
  border: 2px solid var(--border-color);
  border-radius: 0.5rem;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--text-primary);
  transition: all 0.2s;
  text-align: left;
}

.advanced-toggle:hover {
  background: var(--bg-tertiary);
  border-color: #667eea;
}

.advanced-toggle--open {
  border-color: #667eea;
  background: var(--bg-tertiary);
}

.advanced-toggle__icon {
  font-size: 0.75rem;
  color: var(--text-secondary);
  transition: transform 0.2s;
  flex-shrink: 0;
}

.advanced-toggle--open .advanced-toggle__icon {
  transform: rotate(0deg);
}

.advanced-toggle__text {
  flex: 1;
}

.advanced-options {
  margin-top: 1rem;
  padding: 1rem;
  background: var(--bg-secondary);
  border: 2px solid var(--border-color);
  border-radius: 0.5rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
  animation: slideDown 0.2s ease-out;
  min-width: 0;
  max-width: 100%;
  box-sizing: border-box;
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

.game-mode-selector {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  min-width: 0;
  max-width: 100%;
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
  min-width: 0;
  max-width: 100%;
  box-sizing: border-box;
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
  min-width: 0;
  max-width: 100%;
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
  word-wrap: break-word;
  overflow-wrap: break-word;
  min-width: 0;
}

.form-hint {
  font-size: 0.75rem;
  color: var(--text-secondary);
  word-wrap: break-word;
  overflow-wrap: break-word;
  word-break: break-word;
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

.difficulty-templates {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
  gap: 0.75rem;
  margin-bottom: 1rem;
}

.difficulty-template {
  padding: 0.875rem;
  border: 2px solid var(--border-color);
  border-radius: 0.5rem;
  background: var(--bg-secondary);
  cursor: pointer;
  transition: all 0.2s;
  text-align: center;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.difficulty-template:hover {
  border-color: #667eea;
  background: var(--bg-tertiary);
  transform: translateY(-2px);
}

.difficulty-template--active {
  border-color: #667eea;
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.1) 0%, rgba(118, 75, 162, 0.1) 100%);
  box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
}

.difficulty-template__title {
  font-weight: 600;
  font-size: 0.875rem;
  color: var(--text-primary);
}

.difficulty-template__params {
  font-size: 0.75rem;
  color: var(--text-secondary);
}

.difficulty-template--active .difficulty-template__title {
  color: #667eea;
}

@media (max-width: 768px) {
  .difficulty-templates {
    grid-template-columns: repeat(2, 1fr);
    gap: 0.5rem;
  }

  .difficulty-template {
    padding: 0.75rem 0.5rem;
  }

  .difficulty-template__title {
    font-size: 0.8125rem;
  }

  .difficulty-template__params {
    font-size: 0.6875rem;
  }
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
  width: 2rem;
  height: 2rem;
  flex-shrink: 0;
}

.rating-status-icon {
  width: 100%;
  height: 100%;
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

@media (max-width: 768px) {
  .room-form {
    gap: 1rem;
  }

  .form-row {
    gap: 0.75rem;
  }

  .form-input {
    padding: 0.625rem;
    font-size: 0.9375rem;
  }
}

@media (max-width: 480px) {
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

