<template>
  <div class="input-wrapper">
    <label v-if="label" class="input__label">
      {{ label }}
      <span v-if="showValue" class="input__value">{{ displayValue }}</span>
    </label>
    <div class="range-container">
      <input
        v-model.number="inputValue"
        :name="name"
        class="input__target input__target--range"
        type="range"
        :min="min"
        :max="max"
        :step="step"
        :disabled="disabled"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps({
  modelValue: {
    type: [Number, String],
    default: 0
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
  min: {
    type: Number,
    default: 0
  },
  max: {
    type: Number,
    default: 1
  },
  step: {
    type: Number,
    default: 0.01
  },
  showValue: {
    type: Boolean,
    default: true
  }
})
const emit = defineEmits(['update:modelValue'])

// Двусторонняя привязка
const inputValue = computed({
  get: () => {
    const val = Number(props.modelValue)
    return isNaN(val) ? props.min : val
  },
  set: (value) => {
    emit('update:modelValue', value)
  }
})

const displayValue = computed(() => {
  const val = Number(inputValue.value)
  return isNaN(val) ? '0.00' : val.toFixed(2)
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
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.input__value {
  font-weight: 500;
  color: #667eea;
  font-size: 0.875rem;
}

.range-container {
  width: 100%;
  position: relative;
}

.input__target--range {
  width: 100%;
  height: 8px;
  border-radius: 4px;
  background: #e5e7eb;
  outline: none;
  -webkit-appearance: none;
  appearance: none;
  cursor: pointer;
}

.input__target--range::-webkit-slider-thumb {
  -webkit-appearance: none;
  appearance: none;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  cursor: pointer;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
  transition: all 0.2s ease-in-out;
}

.input__target--range::-webkit-slider-thumb:hover {
  transform: scale(1.1);
  box-shadow: 0 4px 8px rgba(102, 126, 234, 0.4);
}

.input__target--range::-moz-range-thumb {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  cursor: pointer;
  border: none;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
  transition: all 0.2s ease-in-out;
}

.input__target--range::-moz-range-thumb:hover {
  transform: scale(1.1);
  box-shadow: 0 4px 8px rgba(102, 126, 234, 0.4);
}

.input__target--range:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.input__target--range:disabled::-webkit-slider-thumb {
  cursor: not-allowed;
  transform: none !important;
}

.input__target--range:disabled::-moz-range-thumb {
  cursor: not-allowed;
  transform: none !important;
}
</style>

