<template>
  <div class="router-page">
    <nav class="nav">
      <router-link to="/main" class="nav__link" active-class="nav__link--active">
        Главная
      </router-link>
    </nav>
    <router-view class="router-view"></router-view>
  </div>
</template>

<script setup lang="ts"></script>

<style scoped>
.router-page {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.nav {
  background: white;
  padding: 1rem 2rem;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  display: flex;
  gap: 1rem;
  border-bottom: 2px solid #e5e7eb;
}

.nav__link {
  padding: 0.75rem 1.5rem;
  text-decoration: none;
  color: #6b7280;
  font-weight: 600;
  border-radius: 0.5rem;
  transition: all 0.2s ease-in-out;
  position: relative;
}

.nav__link:hover {
  color: #667eea;
  background-color: #f3f4f6;
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
