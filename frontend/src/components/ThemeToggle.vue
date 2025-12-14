<template>
  <button @click="toggleTheme" class="theme-toggle" :title="theme === 'light' ? 'Переключить на темную тему' : 'Переключить на светлую тему'">
    <IconMoon v-if="theme === 'light'" class="theme-icon" />
    <IconSun v-else class="theme-icon" />
  </button>
</template>

<script setup lang="ts">
import { useTheme } from '@/composables/useTheme'
import IconMoon from '@/components/icons/IconMoon.vue'
import IconSun from '@/components/icons/IconSun.vue'

const { theme, toggleTheme } = useTheme()
</script>

<style scoped>
.theme-toggle {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  border: none;
  background: var(--bg-secondary);
  color: var(--text-primary);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  transition: all 0.3s ease;
  flex-shrink: 0;
}

.theme-toggle:hover {
  transform: scale(1.1);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
}

.theme-toggle:active {
  transform: scale(0.95);
}

.theme-icon {
  width: 20px;
  height: 20px;
  stroke: currentColor;
  fill: none;
}

@media (max-width: 768px) {
  .theme-toggle {
    width: 36px;
    height: 36px;
  }
  
  .theme-icon {
    width: 18px;
    height: 18px;
  }
}
</style>

