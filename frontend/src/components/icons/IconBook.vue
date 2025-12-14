<template>
  <svg :class="className" :style="style" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
    <path d="M4 19.5C4 18.1193 5.11929 17 6.5 17H20" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
    <path d="M6.5 2C5.11929 2 4 3.11929 4 4.5V19.5C4 20.8807 5.11929 22 6.5 22H19.5C20.8807 22 22 20.8807 22 19.5V4.5C22 3.11929 20.8807 2 19.5 2H6.5Z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
    <path d="M9 7H15M9 11H15M9 15H13" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
  </svg>
</template>

<script setup lang="ts">
defineProps<{
  className?: string
  style?: Record<string, string>
}>()
</script>

