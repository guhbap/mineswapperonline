<template>
  <div class="auth-page">
    <div class="auth-container">
      <h1 class="auth-title">Вход</h1>
      <form @submit.prevent="handleSubmit" class="auth-form">
        <TextInput
          v-model="username"
          label="Имя пользователя"
          placeholder="Введите имя пользователя"
          name="username"
          :disabled="loading"
        />
        <TextInput
          v-model="password"
          label="Пароль"
          placeholder="Введите пароль"
          name="password"
          type="password"
          :disabled="loading"
        />
        <div v-if="error" class="error-message">{{ error }}</div>
        <button type="submit" class="auth-button" :disabled="loading">
          {{ loading ? 'Вход...' : 'Войти' }}
        </button>
        <div class="auth-footer">
          <span>Нет аккаунта?</span>
          <router-link to="/register" class="auth-link">Зарегистрироваться</router-link>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import TextInput from '@/components/inputs/TextInput.vue'

const router = useRouter()
const authStore = useAuthStore()

const username = ref('')
const password = ref('')
const loading = ref(false)
const error = ref<string | null>(null)

const handleSubmit = async () => {
  error.value = null
  loading.value = true

  try {
    await authStore.loginUser(username.value, password.value)
    router.push('/main')
  } catch (err: any) {
    error.value = err.response?.data || 'Ошибка входа. Проверьте данные.'
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

