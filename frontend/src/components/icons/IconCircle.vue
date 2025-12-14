<template>
  <svg :class="className" :style="style" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
    <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="2"/>
  </svg>
</template>

<script setup lang="ts">
defineProps<{
  className?: string
  style?: Record<string, string>
}>()
</script>

