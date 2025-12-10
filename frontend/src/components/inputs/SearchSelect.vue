<template>
  <div class="input-wrapper">
    <label v-if="label" class="input__label">{{ label }}</label>
    <div class="search-select" ref="containerRef">
      <div
        class="search-select__trigger"
        :class="{ 'search-select__trigger--open': isOpen, 'search-select__trigger--disabled': disabled }"
        @click="toggleDropdown"
      >
        <input
          v-model="searchQuery"
          type="text"
          class="search-select__input"
          :placeholder="selectedOption ? selectedOption.label : placeholder"
          @focus="openDropdown"
          @blur="handleBlur"
          @input="handleSearch"
          :disabled="disabled"
        />
        <svg
          class="search-select__arrow"
          :class="{ 'search-select__arrow--open': isOpen }"
          viewBox="0 0 12 12"
          fill="none"
        >
          <path d="M6 9L1 4h10z" fill="currentColor" />
        </svg>
      </div>
      <transition name="dropdown">
        <div v-if="isOpen" class="search-select__dropdown">
          <div
            v-for="option in filteredOptions"
            :key="option.value"
            class="search-select__option"
            :class="{ 'search-select__option--selected': option.value === modelValue }"
            @mousedown.prevent="selectOption(option)"
          >
            {{ option.label }}
          </div>
          <div v-if="filteredOptions.length === 0" class="search-select__no-results">
            Ничего не найдено
          </div>
        </div>
      </transition>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'

interface Option {
  value: string
  label: string
}

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
  options: {
    type: Array as () => Option[],
    required: true
  }
})

const emit = defineEmits(['update:modelValue'])

const isOpen = ref(false)
const searchQuery = ref('')
const containerRef = ref<HTMLElement | null>(null)

const selectedOption = computed(() => {
  return props.options.find((opt) => opt.value === props.modelValue)
})

const filteredOptions = computed(() => {
  if (!searchQuery.value.trim()) {
    return props.options
  }
  const query = searchQuery.value.toLowerCase()
  return props.options.filter((option) => option.label.toLowerCase().includes(query))
})

const toggleDropdown = () => {
  if (props.disabled) return
  isOpen.value = !isOpen.value
  if (isOpen.value) {
    searchQuery.value = ''
  }
}

const openDropdown = () => {
  if (props.disabled) return
  isOpen.value = true
  searchQuery.value = ''
}

const handleBlur = () => {
  // Задержка для обработки клика на опцию
  setTimeout(() => {
    isOpen.value = false
    searchQuery.value = ''
  }, 200)
}

const handleSearch = () => {
  if (!isOpen.value) {
    isOpen.value = true
  }
}

const selectOption = (option: Option) => {
  emit('update:modelValue', option.value)
  isOpen.value = false
  searchQuery.value = ''
}

const handleClickOutside = (event: MouseEvent) => {
  if (containerRef.value && !containerRef.value.contains(event.target as Node)) {
    isOpen.value = false
    searchQuery.value = ''
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
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

.search-select {
  position: relative;
  width: 100%;
}

.search-select__trigger {
  position: relative;
  display: flex;
  align-items: center;
  width: 100%;
  padding: 0.75rem 1rem;
  font-size: 1rem;
  line-height: 1.5;
  color: #111827;
  background-color: #ffffff;
  border: 2px solid #e5e7eb;
  border-radius: 0.5rem;
  transition: all 0.2s ease-in-out;
  cursor: pointer;
}

.search-select__trigger:hover:not(.search-select__trigger--disabled) {
  border-color: #d1d5db;
}

.search-select__trigger--open {
  border-color: #6366f1;
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.1);
}

.search-select__trigger--disabled {
  background-color: #f9fafb;
  color: #9ca3af;
  cursor: not-allowed;
  opacity: 0.6;
}

.search-select__input {
  flex: 1;
  border: none;
  outline: none;
  background: transparent;
  font-size: 1rem;
  color: #111827;
  font-family: inherit;
}

.search-select__input::placeholder {
  color: #9ca3af;
}

.search-select__input:disabled {
  cursor: not-allowed;
}

.search-select__arrow {
  width: 12px;
  height: 12px;
  color: #374151;
  transition: transform 0.2s ease-in-out;
  flex-shrink: 0;
  margin-left: 0.5rem;
}

.search-select__arrow--open {
  transform: rotate(180deg);
}

.search-select__dropdown {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  right: 0;
  background: white;
  border: 2px solid #e5e7eb;
  border-radius: 0.5rem;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
  max-height: 200px;
  overflow-y: auto;
  z-index: 1000;
}

.search-select__option {
  padding: 0.75rem 1rem;
  cursor: pointer;
  transition: background-color 0.2s ease-in-out;
  font-size: 1rem;
  color: #111827;
}

.search-select__option:hover {
  background-color: #f3f4f6;
}

.search-select__option--selected {
  background-color: rgba(102, 126, 234, 0.1);
  color: #667eea;
  font-weight: 600;
}

.search-select__no-results {
  padding: 0.75rem 1rem;
  text-align: center;
  color: #9ca3af;
  font-size: 0.875rem;
}

.dropdown-enter-active,
.dropdown-leave-active {
  transition: opacity 0.2s ease, transform 0.2s ease;
}

.dropdown-enter-from,
.dropdown-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}
</style>

