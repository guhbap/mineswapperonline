<template>
  <div class="router-page">
    <nav class="nav" aria-label="Главная навигация">
      <div class="nav__container">
        <div class="nav__left">
          <a
            @click="handleMainClick"
            class="nav__logo"
            :class="{ 'nav__logo--active': $route.path === '/' || $route.path === '' }"
          >
            <span class="logo-icon">💣</span>
            <span class="logo-text">Сапер Онлайн</span>
          </a>
          <div class="nav__links">
            <a
              @click="handleMainClick"
              class="nav__link"
              :class="{ 'nav__link--active': $route.path === '/' || $route.path === '' }"
            >
              <span class="nav__link-icon">🏠</span>
              <span>Главная</span>
            </a>
            <router-link to="/rating" class="nav__link" :class="{ 'nav__link--active': $route.path === '/rating' }">
              <span class="nav__link-icon">🏆</span>
              <span>Рейтинг</span>
            </router-link>
            <router-link to="/faq" class="nav__link" :class="{ 'nav__link--active': $route.path === '/faq' }">
              <span class="nav__link-icon">❓</span>
              <span>FAQ</span>
            </router-link>
          </div>
        </div>
        <div class="nav__right">
          <ThemeToggle />
          <template v-if="authStore.isAuthenticated">
            <router-link to="/profile" class="nav__user">
              <span class="user-avatar" :style="authStore.user?.color ? { background: authStore.user.color } : {}">
                {{ authStore.user?.username?.[0]?.toUpperCase() || 'U' }}
              </span>
              <span class="user-name">{{ authStore.user?.username }}</span>
            </router-link>
            <button @click="handleLogout" class="nav__logout">
              <span class="logout-icon">🚪</span>
              <span>Выйти</span>
            </button>
          </template>
          <template v-else>
            <router-link to="/login" class="nav__link nav__link--auth">
              <span>Войти</span>
            </router-link>
            <router-link to="/register" class="nav__link nav__link--register">
              <span>Регистрация</span>
            </router-link>
          </template>
        </div>
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
  padding: 0;
  box-shadow: 0 4px 12px var(--shadow);
  border-bottom: 1px solid var(--border-color);
  transition: background 0.3s ease, border-color 0.3s ease;
  position: sticky;
  top: 0;
  z-index: 1000;
  backdrop-filter: blur(10px);
  background: rgba(255, 255, 255, 0.95);
}

[data-theme="dark"] .nav {
  background: rgba(31, 41, 55, 0.95);
}

.nav__container {
  max-width: 1400px;
  margin: 0 auto;
  padding: 1rem 2rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 2rem;
}

.nav__left {
  display: flex;
  align-items: center;
  gap: 2rem;
  flex: 1;
}

.nav__logo {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  text-decoration: none;
  color: var(--text-primary);
  font-weight: 700;
  font-size: 1.25rem;
  padding: 0.5rem 1rem;
  border-radius: 0.75rem;
  transition: all 0.2s ease-in-out;
  user-select: none;
  cursor: pointer;
}

.nav__logo:hover {
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.1) 0%, rgba(118, 75, 162, 0.1) 100%);
  transform: translateY(-1px);
}

.nav__logo--active {
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.15) 0%, rgba(118, 75, 162, 0.15) 100%);
}

.logo-icon {
  font-size: 1.5rem;
  line-height: 1;
}

.logo-text {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.nav__links {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}

.nav__right {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.nav__link {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem 1.25rem;
  text-decoration: none;
  color: var(--text-secondary);
  font-weight: 600;
  border-radius: 0.75rem;
  transition: all 0.2s ease-in-out;
  position: relative;
  user-select: none;
  cursor: pointer;
  font-size: 0.9375rem;
}

.nav__link:hover {
  color: #667eea;
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.1) 0%, rgba(118, 75, 162, 0.1) 100%);
  transform: translateY(-1px);
}

.nav__link--active {
  color: #667eea;
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.15) 0%, rgba(118, 75, 162, 0.15) 100%);
  box-shadow: 0 2px 8px rgba(102, 126, 234, 0.2);
}

.nav__link--active::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 60%;
  height: 3px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 2px;
}

.nav__link-icon {
  font-size: 1rem;
  line-height: 1;
  opacity: 0.8;
}

.nav__link--auth {
  font-size: 0.875rem;
  padding: 0.625rem 1rem;
  color: var(--text-secondary);
}

.nav__link--register {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #ffffff;
  padding: 0.625rem 1.25rem;
}

.nav__link--register:hover {
  background: linear-gradient(135deg, #764ba2 0%, #667eea 100%);
  color: #ffffff;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
}

.nav__user {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.5rem 1rem;
  text-decoration: none;
  color: var(--text-primary);
  font-weight: 600;
  border-radius: 0.75rem;
  transition: all 0.2s ease-in-out;
  background: var(--bg-secondary);
  border: 2px solid transparent;
}

.nav__user:hover {
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.1) 0%, rgba(118, 75, 162, 0.1) 100%);
  border-color: rgba(102, 126, 234, 0.3);
  transform: translateY(-1px);
}

.user-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  font-weight: 700;
  font-size: 0.875rem;
  flex-shrink: 0;
}

.user-name {
  font-size: 0.9375rem;
}

.nav__logout {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.625rem 1.25rem;
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
  color: white;
  border: none;
  border-radius: 0.75rem;
  cursor: pointer;
  font-weight: 600;
  font-size: 0.875rem;
  transition: all 0.2s ease-in-out;
  box-shadow: 0 2px 8px rgba(239, 68, 68, 0.3);
}

.nav__logout:hover {
  background: linear-gradient(135deg, #dc2626 0%, #b91c1c 100%);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(239, 68, 68, 0.4);
}

.nav__logout:active {
  transform: translateY(0);
}

.logout-icon {
  font-size: 1rem;
  line-height: 1;
}

.router-view {
  flex: 1;
  overflow: auto;
}

@media (max-width: 768px) {
  .nav__container {
    padding: 0.75rem 1rem;
    flex-wrap: wrap;
    gap: 1rem;
  }

  .nav__left {
    flex-direction: column;
    align-items: flex-start;
    gap: 1rem;
    width: 100%;
  }

  .nav__logo {
    font-size: 1.125rem;
    padding: 0.5rem 0.75rem;
  }

  .logo-icon {
    font-size: 1.25rem;
  }

  .nav__links {
    width: 100%;
    flex-wrap: wrap;
    gap: 0.5rem;
  }

  .nav__link {
    padding: 0.625rem 1rem;
    font-size: 0.875rem;
    flex: 1;
    min-width: calc(33.333% - 0.5rem);
    justify-content: center;
  }

  .nav__link-icon {
    font-size: 0.875rem;
  }

  .nav__right {
    width: 100%;
    justify-content: space-between;
    flex-wrap: wrap;
    gap: 0.75rem;
  }

  .nav__user {
    flex: 1;
    min-width: 120px;
  }

  .user-name {
    display: none;
  }

  .nav__link--auth,
  .nav__link--register {
    padding: 0.5rem 0.875rem;
    font-size: 0.8125rem;
  }

  .nav__logout {
    padding: 0.5rem 1rem;
    font-size: 0.8125rem;
  }
}

@media (max-width: 480px) {
  .nav__container {
    padding: 0.5rem;
  }

  .nav__logo {
    font-size: 1rem;
  }

  .logo-text {
    display: none;
  }

  .nav__links {
    width: 100%;
  }

  .nav__link {
    min-width: calc(50% - 0.25rem);
    padding: 0.5rem 0.75rem;
    font-size: 0.8125rem;
  }

  .nav__link span:not(.nav__link-icon) {
    display: none;
  }

  .nav__link-icon {
    font-size: 1.125rem;
  }
}
</style>
