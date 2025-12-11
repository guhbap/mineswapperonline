<template>
  <div class="router-page">
    <nav class="nav">
      <div class="nav__left">
        <router-link to="/main" class="nav__link" active-class="nav__link--active">
          Главная
        </router-link>
      </div>
      <div class="nav__right" v-if="authStore.isAuthenticated">
        <span class="nav__user">{{ authStore.user?.username }}</span>
        <button @click="handleLogout" class="nav__logout">Выйти</button>
      </div>
    </nav>
    <router-view class="router-view"></router-view>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const handleLogout = () => {
  authStore.logout()
  router.push('/login')
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
}

.nav__link:hover {
  color: #667eea;
  background-color: var(--bg-tertiary);
}

.nav__link--active {
  color: #667eea;
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.1) 0%, rgba(118, 75, 162, 0.1) 100%);
}

.router-view {
  flex: 1;
  overflow: auto;
}

@media (max-width: 768px) {
  .nav {
    padding: 1rem;
    flex-wrap: wrap;
  }

  .nav__link {
    padding: 0.5rem 1rem;
    font-size: 0.875rem;
  }
}
</style>
