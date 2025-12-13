<template>
  <div v-if="show" class="modal-overlay" @click.self="handleOverlayClick">
    <div class="modal">
      <h2 class="modal__title">Создать комнату</h2>

      <div class="modal__form">
        <RoomForm
          v-model="form"
          v-model:has-password="hasPassword"
          :error="error"
          :show-advanced-options="true"
          :show-all-game-modes="false"
          :auto-generate-name="false"
        />

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
import { generateRandomName } from '@/utils/nameGenerator'
import RoomForm, { type RoomFormData } from './RoomForm.vue'

const props = defineProps<{
  show: boolean
}>()

const emit = defineEmits<{
  submit: [data: { name: string; password?: string; rows: number; cols: number; mines: number; gameMode: string; quickStart: boolean; chording: boolean }]
  cancel: []
}>()

const form = ref<RoomFormData>({
  name: generateRandomName(),
  rows: 16,
  cols: 16,
  mines: 40,
  password: '',
  gameMode: 'classic',
  quickStart: false,
  chording: false,
})

const hasPassword = ref(false)
const error = ref('')

const maxMines = computed(() => {
  return form.value.rows * form.value.cols - 15
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

const handleCancel = () => {
  emit('cancel')
  error.value = ''
  form.value = {
    name: generateRandomName(),
    rows: 16,
    cols: 16,
    mines: 40,
    password: '',
    gameMode: 'classic',
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
}
</style>

