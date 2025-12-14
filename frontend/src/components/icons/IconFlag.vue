<template>
  <svg
    :class="className"
    :style="style"
    viewBox="0 0 24 24"
    :width="width"
    :height="height"
  >
    <!-- Древко флага -->
    <line
      x1="4"
      y1="2"
      x2="4"
      y2="22"
      stroke="#8b5a3c"
      stroke-width="2"
      stroke-linecap="round"
    />
    <!-- Полотнище флага -->
    <path
      d="M4 2 L18 8 L4 14 Z"
      class="flag-cloth"
      :style="flagColor ? { fill: flagColor } : {}"
      stroke="#000"
      stroke-width="0.5"
    />
  </svg>
</template>

<script setup lang="ts">
defineProps<{
  className?: string
  style?: Record<string, string>
  width?: string | number
  height?: string | number
  flagColor?: string
}>()
</script>

