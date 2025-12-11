<template>
  <div class="router-page">
    <nav class="nav" aria-label="Главная навигация">
      <div class="nav__left">
        <a 
          @click="handleMainClick" 
          class="nav__link" 
          :class="{ 'nav__link--active': $route.path === '/' || $route.path === '' }"
        >
          Главная
        </a>
      </div>
      <div class="nav__right">
        <ThemeToggle />
        <template v-if="authStore.isAuthenticated">
          <router-link to="/profile" class="nav__link nav__link--auth">
            {{ authStore.user?.username }}
          </router-link>
          <button @click="handleLogout" class="nav__logout">Выйти</button>
        </template>
        <template v-else>
          <router-link to="/login" class="nav__link nav__link--auth">Войти</router-link>
          <router-link to="/register" class="nav__link nav__link--auth">Регистрация</router-link>
        </template>
      </div>
    </nav>
    <router-view class="router-view"></router-view>
  </div>
</template>

<script setup lang="ts">
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import ThemeToggle from '@/components/ThemeToggle.vue'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const handleLogout = () => {
  authStore.logout()
  router.push('/login')
}

const handleMainClick = () => {
  // Если мы уже на главной странице - отправляем событие для сброса игры
  if (route.path === '/' || route.path === '') {
    window.dispatchEvent(new CustomEvent('reset-game'))
  } else {
    // Иначе просто переходим на главную
    router.push('/')
  }
}
</script>

<style scoped>
.router-page {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.nav {
  background: var(--bg-primary);
  padding: 1rem 2rem;
  box-shadow: 0 2px 4px var(--shadow);
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
  border-bottom: 2px solid var(--border-color);
  transition: background 0.3s ease, border-color 0.3s ease;
}

.nav__left {
  display: flex;
  gap: 1rem;
}

.nav__right {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.nav__user {
  color: var(--text-primary);
  font-weight: 600;
}

.nav__logout {
  padding: 0.5rem 1rem;
  background: #ef4444;
  color: white;
  border: none;
  border-radius: 0.5rem;
  cursor: pointer;
  font-weight: 600;
  transition: all 0.2s ease-in-out;
}

.nav__logout:hover {
  background: #dc2626;
  transform: translateY(-1px);
}

.nav__link {
  padding: 0.75rem 1.5rem;
  text-decoration: none;
  color: var(--text-secondary);
  font-weight: 600;
  border-radius: 0.5rem;
  transition: all 0.2s ease-in-out;
  position: relative;
  user-select: none;
  cursor: pointer;
}

.nav__link:hover {
  color: #667eea;
  background-color: var(--bg-tertiary);
}

.nav__link--active {
  color: #667eea;
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.1) 0%, rgba(118, 75, 162, 0.1) 100%);
}

.nav__link--auth {
  font-size: 0.875rem;
  padding: 0.5rem 1rem;
}

.nav__link--auth:hover {
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.1) 0%, rgba(118, 75, 162, 0.1) 100%);
}

.router-view {
  flex: 1;
  overflow: auto;
}

@media (max-width: 768px) {
  .nav {
    padding: 0.75rem 1rem;
    flex-wrap: wrap;
    gap: 0.5rem;
  }

  .nav__link {
    padding: 0.5rem 0.75rem;
    font-size: 0.875rem;
  }

  .nav__link--auth {
    padding: 0.4rem 0.6rem;
    font-size: 0.8rem;
  }

  .nav__logout {
    padding: 0.4rem 0.75rem;
    font-size: 0.8rem;
  }
}
</style>
