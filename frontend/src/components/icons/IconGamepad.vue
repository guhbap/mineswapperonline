<template>
  <svg :class="className" :style="style" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
    <rect x="2" y="6" width="20" height="12" rx="2" stroke="currentColor" stroke-width="2"/>
    <path d="M6 10H18M6 14H14" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
  </svg>
</template>

<script setup lang="ts">
defineProps<{
  className?: string
  style?: Record<string, string>
}>()
</script>

