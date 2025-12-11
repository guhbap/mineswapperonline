<template>
  <div v-if="show" class="nickname-modal-overlay" @click.self="handleOverlayClick">
    <div class="nickname-modal">
      <h2 class="nickname-modal__title">Введите ваш никнейм</h2>
      <div class="nickname-input-wrapper">
        <input
          v-model="nickname"
          @keyup.enter="handleSubmit"
          type="text"
          class="nickname-modal__input"
          placeholder="Ваш никнейм"
          maxlength="20"
          autofocus
        />
        <button
          type="button"
          @click="generateNickname"
          class="nickname-generate-button"
          title="Сгенерировать случайный никнейм"
        >
          🎲
        </button>
      </div>
      <button @click="handleSubmit" class="nickname-modal__button">
        Войти в игру
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { generateRandomName } from '@/utils/nameGenerator'

const props = defineProps<{
  show: boolean
}>()

const emit = defineEmits<{
  submit: [nickname: string]
}>()

const nickname = ref('')

// Генерируем случайный никнейм при открытии модалки
watch(() => props.show, (isShowing) => {
  if (isShowing && !nickname.value.trim()) {
    nickname.value = generateRandomName()
  }
})

onMounted(() => {
  // Генерируем случайный никнейм при монтировании, если поле пустое
  if (!nickname.value.trim()) {
    nickname.value = generateRandomName()
  }
})

const generateNickname = () => {
  nickname.value = generateRandomName()
}

const handleSubmit = () => {
  const trimmed = nickname.value.trim()
  if (trimmed.length > 0) {
    emit('submit', trimmed)
  }
}

const handleOverlayClick = () => {
  // Не закрываем при клике на overlay, требуем ввод никнейма
}
</script>

<style scoped>
.nickname-modal-overlay {
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

.nickname-modal {
  background: var(--bg-primary);
  padding: 2.5rem;
  border-radius: 1rem;
  box-shadow: 0 20px 60px var(--shadow);
  min-width: 400px;
  max-width: 90vw;
  animation: slideIn 0.3s ease-out;
  transition: background 0.3s ease;
}

@media (max-width: 768px) {
  .nickname-modal {
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

.nickname-modal__title {
  margin: 0 0 1.5rem 0;
  font-size: 1.5rem;
  color: var(--text-primary);
  text-align: center;
  transition: color 0.3s ease;
}

.nickname-input-wrapper {
  position: relative;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 1.5rem;
}

.nickname-modal__input {
  flex: 1;
  padding: 0.875rem 1rem;
  font-size: 1rem;
  border: 2px solid var(--border-color);
  border-radius: 0.5rem;
  transition: border-color 0.2s, background 0.3s ease, color 0.3s ease;
  box-sizing: border-box;
  background: var(--bg-secondary);
  color: var(--text-primary);
}

.nickname-generate-button {
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

.nickname-generate-button:hover {
  background: var(--bg-tertiary);
  border-color: #667eea;
  transform: scale(1.05);
}

.nickname-generate-button:active {
  transform: scale(0.95);
}

.nickname-modal__input:focus {
  outline: none;
  border-color: #667eea;
}

.nickname-modal__button {
  width: 100%;
  padding: 0.875rem;
  font-size: 1rem;
  font-weight: 600;
  color: white;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
  border-radius: 0.5rem;
  cursor: pointer;
  transition: transform 0.2s, box-shadow 0.2s;
}

.nickname-modal__button:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
}

.nickname-modal__button:active {
  transform: translateY(0);
}

@media (max-width: 768px) {
  .nickname-modal__title {
    font-size: 1.25rem;
    margin-bottom: 1rem;
  }

  .nickname-modal__input {
    padding: 0.75rem;
    font-size: 0.9375rem;
    margin-bottom: 1rem;
  }

  .nickname-modal__button {
    padding: 0.75rem;
    font-size: 0.9375rem;
  }
}

@media (max-width: 480px) {
  .nickname-modal {
    padding: 1rem;
    border-radius: 0.75rem;
  }

  .nickname-modal__title {
    font-size: 1.125rem;
  }

  .nickname-modal__input {
    padding: 0.625rem;
    font-size: 0.875rem;
  }

  .nickname-generate-button {
    width: 2rem;
    height: 2rem;
    font-size: 1rem;
  }

  .nickname-modal__button {
    padding: 0.625rem;
    font-size: 0.875rem;
  }
}
</style>

