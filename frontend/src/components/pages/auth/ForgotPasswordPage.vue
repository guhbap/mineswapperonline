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
          <p><strong>{{ success }}</strong></p>
          <div class="email-instructions">
            <p>Для восстановления пароля отправьте письмо на адрес:</p>
            <p class="admin-email">
              <a href="mailto:guhbap@gmail.com" class="email-link">guhbap@gmail.com</a>
            </p>
            <p class="instructions-text">
              <strong>Важно:</strong> Письмо должно быть отправлено с email-адреса, на который зарегистрирован ваш аккаунт.
            </p>
            <p class="instructions-text">
              В письме укажите ваш username или email аккаунта, и мы поможем вам восстановить доступ.
            </p>
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

const handleSubmit = async () => {
  error.value = null
  success.value = null
  loading.value = true

  try {
    const response = await requestPasswordReset({ email: email.value })
    success.value = response.message
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

.email-instructions {
  margin-top: 1rem;
  padding-top: 1rem;
  border-top: 1px solid rgba(34, 197, 94, 0.3);
}

.admin-email {
  margin: 1rem 0;
  text-align: center;
}

.email-link {
  display: inline-block;
  padding: 0.75rem 1.5rem;
  background: #667eea;
  color: #ffffff;
  text-decoration: none;
  border-radius: 0.5rem;
  font-weight: 600;
  font-size: 1rem;
  transition: all 0.2s;
}

.email-link:hover {
  background: #764ba2;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
}

.instructions-text {
  margin-top: 0.75rem;
  font-size: 0.875rem;
  line-height: 1.5;
  color: #374151;
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

