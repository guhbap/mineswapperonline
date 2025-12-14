<template>
  <svg :class="className" :style="style" viewBox="0 0 24 24" fill="currentColor" xmlns="http://www.w3.org/2000/svg">
    <!-- Зубчатое колесо с 8 зубцами -->
    <path d="M12 1 L13.4 4.6 L17 5.4 L13.6 6.2 L12 9.8 L10.4 6.2 L7 5.4 L10.6 4.6 L12 1 Z M12 14.2 L13.4 17.8 L17 18.6 L13.6 19.4 L12 23 L10.4 19.4 L7 18.6 L10.6 17.8 L12 14.2 Z M1 12 L4.6 10.6 L5.4 7 L6.2 10.4 L9.8 12 L6.2 13.6 L5.4 17 L4.6 13.4 L1 12 Z M14.2 12 L17.8 13.4 L18.6 17 L19.4 13.6 L23 12 L19.4 10.4 L18.6 7 L17.8 10.6 L14.2 12 Z M4.2 4.2 L6.4 6.4 L5.2 8.6 L3 6.4 L4.2 4.2 Z M19.8 4.2 L22 6.4 L20.8 8.6 L18.6 6.4 L19.8 4.2 Z M4.2 19.8 L6.4 17.6 L8.6 18.8 L6.4 21 L4.2 19.8 Z M19.8 19.8 L21 17.6 L18.8 15.4 L16.6 17.6 L19.8 19.8 Z"/>
    <!-- Внутренний круг -->
    <circle cx="12" cy="12" r="4.5" fill="currentColor"/>
    <!-- Горизонтальная линия -->
    <line x1="7.5" y1="12" x2="16.5" y2="12" stroke="white" stroke-width="2.5" stroke-linecap="round"/>
  </svg>
</template>

<script setup lang="ts">
defineProps<{
  className?: string
  style?: Record<string, string>
}>()
</script>

