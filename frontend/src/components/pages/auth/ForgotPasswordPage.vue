<template>
  <div class="auth-page">
    <div class="auth-page__theme-toggle">
      <ThemeToggle />
    </div>
    <div class="auth-container">
      <h1 class="auth-title">Восстановление пароля</h1>
      <form @submit.prevent="handleSubmit" class="auth-form">
        <TextInput
          v-model="email"
          label="Email"
          placeholder="Введите email вашего аккаунта"
          name="email"
          type="email"
          :disabled="loading"
        />
        <div v-if="error" class="error-message">{{ error }}</div>
        <div v-if="success" class="success-message">
          <p>{{ success }}</p>
          <div v-if="resetToken" class="reset-token-info">
            <p><strong>Токен сброса пароля (только для разработки):</strong></p>
            <div class="token-display">{{ resetToken }}</div>
            <p class="token-note">В продакшене этот токен будет отправлен на email</p>
            <router-link :to="`/reset-password?token=${resetToken}`" class="reset-link">
              Перейти к сбросу пароля
            </router-link>
          </div>
        </div>
        <button type="submit" class="auth-button" :disabled="loading">
          {{ loading ? 'Отправка...' : 'Отправить запрос' }}
        </button>
        <div class="auth-footer">
          <span>Вспомнили пароль?</span>
          <router-link to="/login" class="auth-link">Войти</router-link>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { requestPasswordReset } from '@/api/auth'
import TextInput from '@/components/inputs/TextInput.vue'
import ThemeToggle from '@/components/ThemeToggle.vue'
import { getErrorMessage } from '@/utils/errorHandler'

const email = ref('')
const loading = ref(false)
const error = ref<string | null>(null)
const success = ref<string | null>(null)
const resetToken = ref<string | null>(null)

const handleSubmit = async () => {
  error.value = null
  success.value = null
  resetToken.value = null
  loading.value = true

  try {
    const response = await requestPasswordReset({ email: email.value })
    success.value = response.message
    // В разработке получаем токен напрямую (в продакшене это небезопасно!)
    if (response.resetToken) {
      resetToken.value = response.resetToken
    }
  } catch (err: any) {
    error.value = getErrorMessage(err, 'Ошибка запроса сброса пароля')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.auth-page {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  padding: 2rem;
  background: var(--bg-primary, #f9fafb);
  position: relative;
}

.auth-page__theme-toggle {
  position: absolute;
  top: 1rem;
  right: 1rem;
  z-index: 10;
}

.auth-container {
  width: 100%;
  max-width: 400px;
  background: var(--bg-secondary, #ffffff);
  padding: 2.5rem;
  border-radius: 1rem;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}

.auth-title {
  font-size: 2rem;
  font-weight: 700;
  margin-bottom: 2rem;
  text-align: center;
  color: var(--text-primary, #111827);
}

.auth-form {
  display: flex;
  flex-direction: column;
}

.error-message {
  color: #ef4444;
  font-size: 0.875rem;
  margin-bottom: 1rem;
  padding: 0.75rem;
  background: #fee2e2;
  border-radius: 0.5rem;
}

.success-message {
  color: #22c55e;
  font-size: 0.875rem;
  margin-bottom: 1rem;
  padding: 0.75rem;
  background: #dcfce7;
  border-radius: 0.5rem;
}

.reset-token-info {
  margin-top: 1rem;
  padding-top: 1rem;
  border-top: 1px solid rgba(34, 197, 94, 0.3);
}

.token-display {
  background: #f3f4f6;
  padding: 0.75rem;
  border-radius: 0.5rem;
  font-family: monospace;
  font-size: 0.75rem;
  word-break: break-all;
  margin: 0.5rem 0;
  color: #111827;
}

.token-note {
  font-size: 0.75rem;
  color: #6b7280;
  margin-top: 0.5rem;
  font-style: italic;
}

.reset-link {
  display: inline-block;
  margin-top: 0.75rem;
  color: #667eea;
  text-decoration: none;
  font-weight: 600;
  transition: color 0.2s;
}

.reset-link:hover {
  color: #764ba2;
}

.auth-button {
  width: 100%;
  padding: 0.875rem 1.5rem;
  font-size: 1rem;
  font-weight: 600;
  color: #ffffff;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
  border-radius: 0.5rem;
  cursor: pointer;
  transition: all 0.2s ease-in-out;
  margin-top: 1rem;
}

.auth-button:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
}

.auth-button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.auth-footer {
  margin-top: 1.5rem;
  text-align: center;
  font-size: 0.875rem;
  color: var(--text-secondary, #6b7280);
}

.auth-link {
  color: #667eea;
  text-decoration: none;
  font-weight: 600;
  margin-left: 0.5rem;
  transition: color 0.2s;
}

.auth-link:hover {
  color: #764ba2;
}
</style>

