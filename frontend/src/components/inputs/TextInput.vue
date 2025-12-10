<template>
  <div class="input-wrapper">
    <label v-if="label" class="input__label">{{ label }}</label>
    <input
      v-if="!multiline"
      v-model="inputValue"
      :name="name"
      class="input__target"
      type="text"
      :placeholder="placeholder"
      :disabled="disabled"
    />
    <textarea
      v-if="multiline"
      v-model="inputValue"
      :name="name"
      class="input__target input__target--textarea"
      :placeholder="placeholder"
      :disabled="disabled"
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps({
  modelValue: {
    type: String,
    default: ''
  },
  name: {
    type: String,
    default: ''
  },
  label: {
    type: String,
    default: ''
  },
  disabled: {
    type: Boolean,
    default: false
  },
  placeholder: {
    type: String,
    default: ''
  },
  multiline: {
    type: Boolean,
    default: false
  }
})
const emit = defineEmits(['update:modelValue'])

// Двусторонняя привязка с фильтрацией
const inputValue = computed({
  get: () => props.modelValue,
  set: (value) => {
    emit('update:modelValue', value)
  }
})
</script>

<style scoped>
.input-wrapper {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  margin-bottom: 1.25rem;
}

.input__label {
  font-size: 0.875rem;
  font-weight: 600;
  color: #374151;
  margin-bottom: 0.25rem;
  display: block;
}

.input__target {
  width: 100%;
  padding: 0.75rem 1rem;
  font-size: 1rem;
  line-height: 1.5;
  color: #111827;
  background-color: #ffffff;
  border: 2px solid #e5e7eb;
  border-radius: 0.5rem;
  transition: all 0.2s ease-in-out;
  outline: none;
  font-family: inherit;
}

.input__target:hover:not(:disabled) {
  border-color: #d1d5db;
}

.input__target:focus {
  border-color: #6366f1;
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.1);
}

.input__target:disabled {
  background-color: #f9fafb;
  color: #9ca3af;
  cursor: not-allowed;
  opacity: 0.6;
}

.input__target::placeholder {
  color: #9ca3af;
}

.input__target--textarea {
  min-height: 100px;
  resize: vertical;
  font-family: inherit;
}
</style>
