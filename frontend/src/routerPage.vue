<template>
  <div class="router-page">
    <nav class="nav" aria-label="Главная навигация">
      <div class="nav__container">
        <div class="nav__left">
          <div class="nav__links">
            <a
              @click="handleMainClick"
              class="nav__link"
              :class="{ 'nav__link--active': $route.path === '/' || $route.path === '' }"
            >
              <svg class="nav__link-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M3 12L5 10M5 10L12 3L19 10M5 10V20C5 20.5523 5.44772 21 6 21H9M19 10L21 12M19 10V20C19 20.5523 18.5523 21 18 21H15M9 21C9.55228 21 10 20.5523 10 20V16C10 15.4477 10.4477 15 11 15H13C13.5523 15 14 15.4477 14 16V20C14 20.5523 14.4477 21 15 21M9 21H15" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
              </svg>
              <span>Главная</span>
            </a>
            <router-link to="/rating" class="nav__link" :class="{ 'nav__link--active': $route.path === '/rating' }">
              <svg class="nav__link-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M6 9H4.5C3.67157 9 3 9.67157 3 10.5V19.5C3 20.3284 3.67157 21 4.5 21H19.5C20.3284 21 21 20.3284 21 19.5V10.5C21 9.67157 20.3284 9 19.5 9H18M6 9V6C6 4.34315 7.34315 3 9 3H15C16.6569 3 18 4.34315 18 6V9M6 9H18M12 9V21" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                <path d="M9 12L12 9L15 12" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
              </svg>
              <span>Рейтинг</span>
            </router-link>
            <router-link to="/faq" class="nav__link" :class="{ 'nav__link--active': $route.path === '/faq' }">
              <svg class="nav__link-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="2"/>
                <path d="M9.09 9C9.3251 8.33167 9.78915 7.76811 10.4 7.40913C11.0108 7.05016 11.7289 6.91894 12.4272 7.03871C13.1255 7.15849 13.7588 7.52152 14.2151 8.06353C14.6713 8.60553 14.9211 9.29152 14.92 10C14.92 12 11.92 13 11.92 13" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                <path d="M12 17H12.01" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
              </svg>
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
              <svg class="logout-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M9 21H5C4.46957 21 3.96086 20.7893 3.58579 20.4142C3.21071 20.0391 3 19.5304 3 19V5C3 4.46957 3.21071 3.96086 3.58579 3.58579C3.96086 3.21071 4.46957 3 5 3H9" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                <path d="M16 17L21 12L16 7" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                <path d="M21 12H9" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
              </svg>
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
  width: 1.5rem;
  height: 1.5rem;
  flex-shrink: 0;
  color: #667eea;
}

.logo-text {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

[data-theme="dark"] .logo-text {
  background: linear-gradient(135deg, #818cf8 0%, #a78bfa 100%);
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

/* Улучшаем читаемость в темной теме */
[data-theme="dark"] .nav__link {
  color: var(--text-primary);
}

.nav__link:hover {
  color: #667eea;
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.1) 0%, rgba(118, 75, 162, 0.1) 100%);
  transform: translateY(-1px);
}

[data-theme="dark"] .nav__link:hover {
  color: #818cf8;
}

.nav__link--active {
  color: #667eea;
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.15) 0%, rgba(118, 75, 162, 0.15) 100%);
  box-shadow: 0 2px 8px rgba(102, 126, 234, 0.2);
}

[data-theme="dark"] .nav__link--active {
  color: #818cf8;
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

[data-theme="dark"] .nav__link--active::after {
  background: linear-gradient(135deg, #818cf8 0%, #a78bfa 100%);
}

.nav__link-icon {
  width: 1rem;
  height: 1rem;
  flex-shrink: 0;
  opacity: 0.8;
}

.nav__link--auth {
  font-size: 0.875rem;
  padding: 0.625rem 1rem;
  color: var(--text-secondary);
}

[data-theme="dark"] .nav__link--auth {
  color: var(--text-primary);
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
  width: 1rem;
  height: 1rem;
  flex-shrink: 0;
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
    width: 1.25rem;
    height: 1.25rem;
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
    width: 0.875rem;
    height: 0.875rem;
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
    width: 1.125rem;
    height: 1.125rem;
  }
}
</style>
