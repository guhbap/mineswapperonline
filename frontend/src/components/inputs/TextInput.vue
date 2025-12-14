<template>
  <div class="input-wrapper">
    <label v-if="label" class="input__label">{{ label }}</label>
    <input
      v-if="!multiline"
      v-model="inputValue"
      :name="name"
      class="input__target"
      :type="type"
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
  },
  type: {
    type: String,
    default: 'text'
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
  color: var(--text-primary);
  margin-bottom: 0.25rem;
  display: block;
}

.input__target {
  width: 100%;
  padding: 0.75rem 1rem;
  font-size: 1rem;
  line-height: 1.5;
  color: var(--text-primary);
  background-color: var(--bg-primary);
  border: 2px solid var(--border-color);
  border-radius: 0.5rem;
  transition: all 0.2s ease-in-out;
  outline: none;
  font-family: inherit;
}

.input__target:hover:not(:disabled) {
  border-color: var(--text-secondary);
}

.input__target:focus {
  border-color: #6366f1;
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.1);
}

.input__target:disabled {
  background-color: var(--bg-secondary);
  color: var(--text-secondary);
  cursor: not-allowed;
  opacity: 0.6;
}

.input__target::placeholder {
  color: var(--text-secondary);
  opacity: 0.7;
}

.input__target--textarea {
  min-height: 100px;
  resize: vertical;
  font-family: inherit;
}
</style>
