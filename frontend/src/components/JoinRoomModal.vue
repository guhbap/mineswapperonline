<template>
  <div v-if="show" class="modal-overlay" @click.self="handleOverlayClick">
    <div class="modal">
      <h2 class="modal__title">Подключиться к комнате</h2>
      <p class="modal__subtitle">{{ room.name }}</p>

      <div class="modal__form">
        <div v-if="room.hasPassword" class="form-group">
          <label class="form-label">Пароль</label>
          <input
            v-model="password"
            type="password"
            class="form-input"
            placeholder="Введите пароль"
            @keyup.enter="handleSubmit"
            autofocus
          />
        </div>
        <div v-else class="form-info">
          Комната не защищена паролем
        </div>

        <div v-if="error" class="form-error">{{ error }}</div>

        <div class="modal__actions">
          <button @click="handleCancel" class="btn btn-secondary">Отмена</button>
          <button @click="handleSubmit" class="btn btn-primary">
            Подключиться
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { Room } from '@/api/rooms'
import { joinRoom } from '@/api/rooms'

const props = defineProps<{
  show: boolean
  room: Room | null
}>()

const emit = defineEmits<{
  submit: [room: Room]
  cancel: []
}>()

const password = ref('')
const error = ref('')

const handleSubmit = async () => {
  if (!props.room) return

  if (props.room.hasPassword && !password.value.trim()) {
    error.value = 'Введите пароль'
    return
  }

  try {
    const roomData = await joinRoom({
      roomId: props.room.id,
      password: props.room.hasPassword ? password.value : undefined,
    })
    emit('submit', roomData)
    error.value = ''
    password.value = ''
  } catch (err: any) {
    error.value = err.response?.data || 'Ошибка подключения к комнате'
  }
}

const handleCancel = () => {
  emit('cancel')
  error.value = ''
  password.value = ''
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
  min-width: 400px;
  max-width: 90vw;
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
  margin: 0 0 0.5rem 0;
  font-size: 1.5rem;
  color: var(--text-primary);
  text-align: center;
}

.modal__subtitle {
  margin: 0 0 1.5rem 0;
  font-size: 1rem;
  color: var(--text-secondary);
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

.form-info {
  padding: 1rem;
  background: var(--bg-secondary);
  border-radius: 0.5rem;
  color: var(--text-secondary);
  text-align: center;
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

.btn-primary:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
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

  .modal__subtitle {
    font-size: 1rem;
  }

  .modal__form {
    gap: 1rem;
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
}
</style>

