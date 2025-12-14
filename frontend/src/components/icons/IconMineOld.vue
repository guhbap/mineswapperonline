<template>
  <svg :class="className" :style="style" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
    <!-- Основной корпус мины -->
    <circle cx="12" cy="14" r="7" fill="currentColor" stroke="#333" stroke-width="1"/>
    <!-- Сегменты мины -->
    <path d="M12 7L12 14M7 14L12 7M17 14L12 7" stroke="#333" stroke-width="1.5" stroke-linecap="round"/>
    <path d="M9 11L12 7L15 11M9 17L12 7L15 17" stroke="#333" stroke-width="1.5" stroke-linecap="round"/>
    <!-- Фитиль -->
    <line x1="12" y1="7" x2="12" y2="4" stroke="#8b5a3c" stroke-width="2" stroke-linecap="round"/>
    <circle cx="12" cy="3" r="1" fill="#ff6b6b"/>
    <!-- Центральная точка -->
    <circle cx="12" cy="14" r="2" fill="#333"/>
  </svg>
</template>

<script setup lang="ts">
defineProps<{
  className?: string
  style?: Record<string, string>
}>()
</script>

