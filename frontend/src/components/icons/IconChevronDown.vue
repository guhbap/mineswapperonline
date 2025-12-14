<template>
  <svg
    :class="className"
    :style="style"
    viewBox="0 0 12 12"
    fill="none"
  >
    <path d="M6 9L1 4h10z" fill="currentColor" />
  </svg>
</template>

<script setup lang="ts">
defineProps<{
  className?: string
  style?: Record<string, string>
}>()
</script>

