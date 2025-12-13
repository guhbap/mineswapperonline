<template>
  <div v-if="show" class="modal-overlay" @click.self="handleOverlayClick">
    <div class="modal">
      <h2 class="modal__title">Редактировать комнату</h2>

      <div class="modal__form">
        <RoomForm
          v-model="form"
          v-model:has-password="hasPassword"
          :error="error"
          :show-advanced-options="true"
          :show-all-game-modes="false"
          :auto-generate-name="false"
        >
          <template #warning>
            <div class="form-warning">
              ⚠️ Внимание: изменение параметров комнаты пересоздаст игровое поле. Текущий прогресс будет потерян.
            </div>
          </template>
        </RoomForm>

        <div class="modal__actions">
          <button @click="handleCancel" class="btn btn-secondary">Отмена</button>
          <button @click="handleSubmit" class="btn btn-primary" :disabled="!isValid || loading">
            {{ loading ? 'Сохранение...' : 'Сохранить' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import RoomForm, { type RoomFormData } from './RoomForm.vue'
import { updateRoom, type Room } from '@/api/rooms'

const props = defineProps<{
  show: boolean
  room: Room | null
}>()

const emit = defineEmits<{
  submit: [room: Room]
  cancel: []
}>()

const form = ref<RoomFormData>({
  name: '',
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
const loading = ref(false)

// Загружаем данные комнаты при открытии модалки или изменении комнаты
watch([() => props.show, () => props.room], ([isShowing, room]) => {
  if (isShowing && room) {
    form.value = {
      name: room.name,
      rows: room.rows,
      cols: room.cols,
      mines: room.mines,
      password: '',
      gameMode: (room.gameMode ?? 'classic') as 'classic' | 'training' | 'fair',
      quickStart: room.quickStart ?? false,
      chording: room.chording ?? false,
    }
    hasPassword.value = room.hasPassword
    error.value = ''
  }
}, { immediate: true })

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

const handleSubmit = async () => {
  if (!isValid.value || !props.room) {
    error.value = 'Заполните все поля корректно'
    return
  }

  error.value = ''
  loading.value = true

  try {
    const data: any = {
      name: form.value.name.trim(),
      rows: form.value.rows,
      cols: form.value.cols,
      mines: form.value.mines,
      gameMode: form.value.gameMode,
      quickStart: form.value.quickStart,
      chording: form.value.chording,
    }

    // Обрабатываем пароль:
    // - Если пользователь убрал галочку пароля - отправляем пустую строку (удаляем пароль)
    // - Если пользователь установил галочку и ввел пароль - отправляем новый пароль
    // - Если пароль не менялся (галочка осталась, но пароль не вводился) - не отправляем поле
    if (!hasPassword.value) {
      // Пользователь убрал пароль
      data.password = ''
    } else if (form.value.password) {
      // Пользователь установил новый пароль
      data.password = form.value.password
    }
    // Если hasPassword.value === true, но password пустой - не отправляем (не меняем пароль)

    const updatedRoom = await updateRoom(props.room.id, data)
    emit('submit', updatedRoom)
    error.value = ''
  } catch (err: any) {
    error.value = err.response?.data?.error || err.message || 'Ошибка обновления комнаты'
  } finally {
    loading.value = false
  }
}

const handleCancel = () => {
  emit('cancel')
  error.value = ''
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

.form-warning {
  padding: 0.75rem;
  background: rgba(255, 193, 7, 0.1);
  color: #ffc107;
  border-radius: 0.5rem;
  font-size: 0.875rem;
  border: 1px solid rgba(255, 193, 7, 0.3);
  margin-top: 0.5rem;
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

