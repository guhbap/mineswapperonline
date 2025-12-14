<template>
  <div class="checkbox-wrapper">
    <label class="checkbox">
      <input v-model="localValue" class="checkbox__input" type="checkbox" :disabled="disabled" />
      <span class="checkbox__icon">
        <IconCheckmark class="checkbox__checkmark" />
      </span>
      <span v-if="label" :class="{ 'checkbox__text--disabled': disabled }" class="checkbox__text">
        {{ label }}
      </span>
    </label>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import IconCheckmark from '@/components/icons/IconCheckmark.vue'

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  label: {
    type: String,
    default: ''
  },
  disabled: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['update:modelValue'])

const localValue = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val)
})
</script>

<style scoped>
.checkbox-wrapper {
  margin-bottom: 1.25rem;
}

.checkbox {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  cursor: pointer;
  user-select: none;
  position: relative;
}

.checkbox__input {
  position: absolute;
  opacity: 0;
  width: 0;
  height: 0;
}

.checkbox__icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 1.5rem;
  height: 1.5rem;
  border: 2px solid #d1d5db;
  border-radius: 0.375rem;
  background-color: #ffffff;
  transition: all 0.2s ease-in-out;
  flex-shrink: 0;
  position: relative;
}

.checkbox__checkmark {
  width: 1rem;
  height: 1rem;
  color: white;
  opacity: 0;
  transform: scale(0);
  transition: all 0.2s ease-in-out;
}

.checkbox__input:checked + .checkbox__icon {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-color: #667eea;
  box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
}

.checkbox__input:checked + .checkbox__icon .checkbox__checkmark {
  opacity: 1;
  transform: scale(1);
}

.checkbox__input:focus + .checkbox__icon {
  outline: 2px solid #667eea;
  outline-offset: 2px;
}

.checkbox__input:hover:not(:disabled) + .checkbox__icon {
  border-color: #9ca3af;
}

.checkbox__input:disabled + .checkbox__icon {
  background-color: #f3f4f6;
  border-color: #e5e7eb;
  cursor: not-allowed;
  opacity: 0.6;
}

.checkbox__input:disabled ~ .checkbox__text {
  opacity: 0.6;
  cursor: not-allowed;
}

.checkbox__text {
  font-size: 1rem;
  color: #374151;
  font-weight: 500;
  line-height: 1.5;
}

.checkbox__text--disabled {
  color: #9ca3af;
}

.checkbox:has(.checkbox__input:disabled) {
  cursor: not-allowed;
}
</style>
