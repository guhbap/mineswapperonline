<template>
  <svg :class="className" :style="style" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
    <circle cx="12" cy="12" r="10" fill="#ffd700" stroke="#ffed4e" stroke-width="2"/>
    <path d="M8 8L10 6M16 8L14 6M10 6L12 4L14 6" stroke="white" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
    <circle cx="12" cy="12" r="2" fill="white"/>
  </svg>
</template>

<script setup lang="ts">
defineProps<{
  className?: string
  style?: Record<string, string>
}>()
</script>

