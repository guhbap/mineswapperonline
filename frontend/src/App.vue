<template>
  <div class="app">
    <router-view />
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useTheme } from '@/composables/useTheme'
import { useUserActivity } from '@/composables/useUserActivity'

const { initTheme } = useTheme()

// Инициализируем отслеживание активности пользователя
useUserActivity()

onMounted(() => {
  initTheme()
})
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
  font-family: 'Nunito Sans', -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', 'Oxygen',
    'Ubuntu', 'Cantarell', 'Fira Sans', 'Droid Sans', 'Helvetica Neue',
    sans-serif;
  font-weight: 400;
}

:root {
  /* Светлая тема */
  --bg-primary: #ffffff;
  --bg-secondary: #f5f7fa;
  --bg-tertiary: #e5e7eb;
  --text-primary: #1f2937;
  --text-secondary: #6b7280;
  --border-color: #d1d5db;
  --shadow: rgba(0, 0, 0, 0.1);
  --shadow-hover: rgba(0, 0, 0, 0.15);
  --gradient-start: #f5f7fa;
  --gradient-end: #c3cfe2;
}

[data-theme="dark"] {
  /* Темная тема */
  --bg-primary: #1f2937;
  --bg-secondary: #111827;
  --bg-tertiary: #374151;
  --text-primary: #f9fafb;
  --text-secondary: #d1d5db;
  --border-color: #4b5563;
  --shadow: rgba(0, 0, 0, 0.3);
  --shadow-hover: rgba(0, 0, 0, 0.4);
  --gradient-start: #111827;
  --gradient-end: #1f2937;
}

body {
  font-family: 'Nunito Sans', -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', 'Oxygen',
    'Ubuntu', 'Cantarell', 'Fira Sans', 'Droid Sans', 'Helvetica Neue',
    sans-serif;
  font-weight: 400;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  background: linear-gradient(135deg, var(--gradient-start) 0%, var(--gradient-end) 100%);
  min-height: 100vh;
  color: var(--text-primary);
  transition: background 0.3s ease, color 0.3s ease;
}

.app {
  min-height: 100vh;
}

#app {
  min-height: 100vh;
}

code {
  font-family: source-code-pro, Menlo, Monaco, Consolas, 'Courier New',
    monospace;
}

/* Глобальные стили для гиперссылок */
a {
  color: #667eea;
  text-decoration: none;
  transition: color 0.2s ease, opacity 0.2s ease;
  font-weight: 500;
}

a:hover {
  color: #764ba2;
  text-decoration: underline;
}

a:visited {
  color: #667eea;
}

a:active {
  color: #764ba2;
  opacity: 0.8;
}

/* Стили для ссылок в темной теме */
[data-theme="dark"] a {
  color: #818cf8;
}

[data-theme="dark"] a:hover {
  color: #a78bfa;
}

[data-theme="dark"] a:visited {
  color: #818cf8;
}

[data-theme="dark"] a:active {
  color: #a78bfa;
}
</style>