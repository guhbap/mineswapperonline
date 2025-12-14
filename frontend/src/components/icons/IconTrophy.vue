<template>
  <svg :class="className" :style="style" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
    <path d="M6 9H4.5C3.67157 9 3 9.67157 3 10.5V19.5C3 20.3284 3.67157 21 4.5 21H19.5C20.3284 21 21 20.3284 21 19.5V10.5C21 9.67157 20.3284 9 19.5 9H18M6 9V6C6 4.34315 7.34315 3 9 3H15C16.6569 3 18 4.34315 18 6V9M6 9H18M12 9V21" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
    <path d="M9 12L12 9L15 12" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
  </svg>
</template>

<script setup lang="ts">
defineProps<{
  className?: string
  style?: Record<string, string>
}>()
</script>

