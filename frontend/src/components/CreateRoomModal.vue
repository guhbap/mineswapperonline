<template>
  <div v-if="show" class="modal-overlay" @click.self="handleOverlayClick">
    <div class="modal">
      <h2 class="modal__title">Создать комнату</h2>

      <div class="modal__form">
        <div class="form-group">
          <label class="form-label">Название комнаты</label>
          <input
            v-model="form.name"
            type="text"
            class="form-input"
            placeholder="Название комнаты"
            maxlength="30"
          />
        </div>

        <div class="form-group">
          <label class="form-label">Размер поля</label>
          <div class="form-row">
            <div class="form-col">
              <label class="form-label-small">Строки</label>
              <input
                v-model.number="form.rows"
                type="number"
                class="form-input"
                min="5"
                max="30"
              />
            </div>
            <div class="form-col">
              <label class="form-label-small">Столбцы</label>
              <input
                v-model.number="form.cols"
                type="number"
                class="form-input"
                min="5"
                max="30"
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
import { ref, computed } from 'vue'

const props = defineProps<{
  show: boolean
}>()

const emit = defineEmits<{
  submit: [data: { name: string; password?: string; rows: number; cols: number; mines: number }]
  cancel: []
}>()

const form = ref({
  name: '',
  rows: 16,
  cols: 16,
  mines: 40,
  password: '',
})

const hasPassword = ref(false)
const error = ref('')

const maxMines = computed(() => {
  return form.value.rows * form.value.cols - 1
})

const isValid = computed(() => {
  return (
    form.value.name.trim().length > 0 &&
    form.value.rows >= 5 &&
    form.value.rows <= 30 &&
    form.value.cols >= 5 &&
    form.value.cols <= 30 &&
    form.value.mines >= 1 &&
    form.value.mines <= maxMines.value
  )
})

const handleSubmit = () => {
  if (!isValid.value) {
    error.value = 'Заполните все поля корректно'
    return
  }

  const data = {
    name: form.value.name.trim(),
    rows: form.value.rows,
    cols: form.value.cols,
    mines: form.value.mines,
    ...(hasPassword.value && form.value.password ? { password: form.value.password } : {}),
  }

  emit('submit', data)
  error.value = ''
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
  margin-bottom: 0.25rem;
}

.form-input {
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

.form-hint {
  font-size: 0.75rem;
  color: var(--text-secondary);
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
</style>

