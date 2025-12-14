<template>
  <svg :class="className" :style="style" viewBox="0 0 32 32" xmlns="http://www.w3.org/2000/svg">
    <circle cx="16" cy="16" r="12" stroke="currentColor" stroke-width="2"/>
    <path d="M16 2V5" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="bevel"/>
    <path d="M16 27V30" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="bevel"/>
    <path d="M30.0049 15.9951L27.0049 15.9951" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="bevel"/>
    <path d="M19 16H13" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="bevel"/>
    <path d="M5.00488 15.9951L2.00488 15.9951" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="bevel"/>
    <path d="M25.9009 6.09717L23.7796 8.21849" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="bevel"/>
    <path d="M8.22314 23.7749L6.10182 25.8962" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="bevel"/>
    <path d="M25.9009 25.896L23.7796 23.7747" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="bevel"/>
    <path d="M8.22314 8.21826L6.10182 6.09694" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="bevel"/>
  </svg>
</template>

<script setup lang="ts">
defineProps<{
  className?: string
  style?: Record<string, string>
}>()
</script>

