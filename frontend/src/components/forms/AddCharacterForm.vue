<template>
  <div class="form-container">
    <form class="form" @submit.prevent="handleSubmit">
      <h2 class="form__title">{{ isEditMode ? 'Редактировать персонажа' : 'Добавить нового персонажа' }}</h2>

      <div class="form__grid">
        <TextInput
          v-model="name_ru"
          label="Имя (RU)"
          placeholder="Введите имя на русском"
        />
        <TextInput v-model="name_en" label="Имя (EN)" placeholder="Enter name in English" />
        <TextInput
          v-model="name_original"
          label="Оригинальное имя"
          placeholder="Введите оригинальное имя"
        />
      </div>

      <div class="form__grid form__grid--title">
        <SearchSelect
          v-model="title_uuid"
          label="Тайтл"
          placeholder="Поиск тайтла..."
          :options="titleOptions"
        />
        <SelectInput
          v-model="gender"
          label="Пол"
          placeholder="Выберите пол"
          :options="genderOptions"
        />
      </div>

      <div class="form__grid">
        <NumberInput v-model="age" label="Возраст" placeholder="Введите возраст" :min="0" />
        <RangeInput
          v-model="multiplier"
          label="Множитель"
          :min="0"
          :max="1"
          :step="0.01"
        />
      </div>

      <PictureList v-model="pictures" />

      <div v-if="error" class="form__error">
        <svg class="form__error-icon" viewBox="0 0 20 20" fill="currentColor">
          <path
            fill-rule="evenodd"
            d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
            clip-rule="evenodd"
          />
        </svg>
        <span>{{ error }}</span>
      </div>

      <div v-if="success" class="form__success">
        <svg class="form__success-icon" viewBox="0 0 20 20" fill="currentColor">
          <path
            fill-rule="evenodd"
            d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
            clip-rule="evenodd"
          />
        </svg>
        <span>{{ isEditMode ? 'Персонаж успешно обновлен!' : 'Персонаж успешно добавлен!' }}</span>
      </div>

      <div class="form__actions">
        <button
          type="submit"
          class="form__button form__button--primary"
          :disabled="loading || !isFormValid"
        >
          <span v-if="loading" class="form__button-spinner"></span>
          <span>{{ loading ? 'Сохранение...' : 'Сохранить' }}</span>
        </button>
        <button
          type="button"
          class="form__button form__button--secondary"
          @click="handleReset"
          :disabled="loading"
        >
          {{ isEditMode ? 'Отмена' : 'Очистить' }}
        </button>
      </div>
    </form>
  </div>
</template>
<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import TextInput from '../inputs/TextInput.vue'
import SelectInput from '../inputs/SelectInput.vue'
import SearchSelect from '../inputs/SearchSelect.vue'
import NumberInput from '../inputs/NumberInput.vue'
import RangeInput from '../inputs/RangeInput.vue'
import PictureList from '../PictureList.vue'
import api from '@/api'
import { useDataStore } from '@/stores/dataStore'
import type { Character } from '@/stores/classes'
import { getErrorMessage } from '@/utils/errorHandler'

const props = defineProps<{
  characterUuid?: string
}>()

const emit = defineEmits<{
  saved: []
  cancelled: []
}>()

const dataStore = useDataStore()

const isEditMode = computed(() => !!props.characterUuid)

const name_ru = ref('')
const name_en = ref('')
const name_original = ref('')
const title_uuid = ref('')
const gender = ref('')
const age = ref<number | string>('')
const multiplier = ref<number | string>(0)
const pictures = ref('[]')

const loading = ref(false)
const error = ref('')
const success = ref(false)

const genderOptions = [
  { value: 'муж', label: 'Мужской' },
  { value: 'жен', label: 'Женский' },
  { value: 'другое', label: 'Другое' }
]

const titleOptions = computed(() => {
  return Object.values(dataStore.titles).map((title) => ({
    value: title.uuid,
    label: title.name_ru || title.name_en || title.name_original || 'Без названия'
  }))
})

// Валидация формы
const isFormValid = computed(() => {
  const multiplierNum = Number(multiplier.value)
  return (
    (name_ru.value.trim() !== '' ||
      name_en.value.trim() !== '' ||
      name_original.value.trim() !== '') &&
    title_uuid.value !== '' &&
    gender.value !== '' &&
    age.value !== '' &&
    !isNaN(multiplierNum) &&
    multiplierNum >= 0 &&
    multiplierNum <= 1
  )
})

// Валидация и преобразование данных перед отправкой
const validateAndPrepareData = () => {
  const errors: string[] = []

  // Проверка обязательных полей
  if (!name_ru.value.trim() && !name_en.value.trim() && !name_original.value.trim()) {
    errors.push('Необходимо указать хотя бы одно имя')
  }

  if (!title_uuid.value) {
    errors.push('Необходимо выбрать тайтл')
  }

  if (!gender.value) {
    errors.push('Необходимо выбрать пол')
  }

  // Валидация возраста
  const ageNum = Number(age.value)
  if (isNaN(ageNum) || ageNum < 0) {
    errors.push('Возраст должен быть положительным числом')
  }

  // Валидация множителя
  const multiplierNum = Number(multiplier.value)
  if (isNaN(multiplierNum) || multiplierNum < 0 || multiplierNum > 1) {
    errors.push('Множитель должен быть числом от 0 до 1')
  }

  if (errors.length > 0) {
    throw new Error(errors.join('. '))
  }

  // Подготовка данных для отправки
  const data: Record<string, unknown> = {
    name_ru: name_ru.value.trim() || '',
    name_en: name_en.value.trim() || '',
    name_original: name_original.value.trim() || '',
    title_uuid: title_uuid.value,
    gender: gender.value,
    age: ageNum,
    multiplier: multiplierNum
  }

  // Обработка pictures
  if (!isEditMode.value) {
    // При создании нового персонажа всегда передаем pictures: '[]'
    data.pictures = '[]'
  } else {
    // При редактировании передаем pictures (даже если это '[]' для очистки)
    if (pictures.value && pictures.value.trim() !== '') {
      data.pictures = pictures.value.trim()
    }
  }

  return data
}

const loadCharacterData = async (uuid: string) => {
  // Убеждаемся, что персонажи загружены
  if (!dataStore.charactersLoaded) {
    await dataStore.loadCharacters()
  }

  const character = dataStore.characters[uuid]
  if (!character) {
    error.value = 'Персонаж не найден'
    return
  }

  name_ru.value = character.name_ru || ''
  name_en.value = character.name_en || ''
  name_original.value = character.name_original || ''
  title_uuid.value = character.title_uuid || ''
  gender.value = character.gender || ''
  age.value = character.age || ''
  multiplier.value = character.multiplier || 0
  pictures.value = character.pictures || '[]'
}

const handleSubmit = async () => {
  if (!isFormValid.value) {
    error.value = 'Заполните все обязательные поля'
    return
  }

  loading.value = true
  error.value = ''
  success.value = false

  try {
    const data = validateAndPrepareData()

    let response
    if (isEditMode.value && props.characterUuid) {
      // Используем PATCH для обновления существующего элемента
      data.uuid = props.characterUuid
      response = await api.patch('/characters', data)
    } else {
      // Используем PUT для создания нового элемента
      response = await api.put('/characters', data)
    }

    // Проверяем успешный ответ (201 Created или 200 OK)
    if (response.status === 201 || response.status === 200) {
      success.value = true

      // Обновляем список персонажей
      await dataStore.loadCharacters(true)

      // Сбрасываем форму через небольшую задержку для показа сообщения об успехе
      setTimeout(() => {
        handleReset()
        emit('saved')
      }, 1500)
    } else {
      throw new Error(response.data?.message || 'Неизвестная ошибка')
    }
  } catch (err: unknown) {
    error.value = getErrorMessage(err, 'Произошла ошибка при сохранении')
    success.value = false
  } finally {
    loading.value = false
  }
}

const handleReset = () => {
  name_ru.value = ''
  name_en.value = ''
  name_original.value = ''
  title_uuid.value = ''
  gender.value = ''
  age.value = ''
  multiplier.value = 0
  pictures.value = '[]'
  error.value = ''
  success.value = false
  emit('cancelled')
}

// Загружаем данные персонажа при изменении characterUuid
watch(
  () => props.characterUuid,
  async (newUuid) => {
    if (newUuid) {
      // Убеждаемся, что персонажи загружены
      if (!dataStore.charactersLoaded) {
        await dataStore.loadCharacters()
      }
      await loadCharacterData(newUuid)
    } else {
      handleReset()
    }
  },
  { immediate: true }
)

onMounted(() => {
  // Загружаем тайтлы, если они еще не загружены
  if (!dataStore.loaded) {
    dataStore.loadData()
  }
  // Загружаем персонажей, если они еще не загружены
  if (!dataStore.charactersLoaded) {
    dataStore.loadCharacters()
  }
  // Если передан characterUuid, загружаем данные
  if (props.characterUuid) {
    loadCharacterData(props.characterUuid)
  }
})
</script>

<style scoped>
.form-container {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 2rem;
  border-radius: 1rem;
  margin-bottom: 2rem;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.1);
}

.form {
  background: white;
  padding: 2rem;
  border-radius: 0.75rem;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.05);
  max-width: 100%;
}

.form__title {
  font-size: 1.5rem;
  font-weight: 700;
  color: #1f2937;
  margin-bottom: 1.5rem;
  padding-bottom: 1rem;
  border-bottom: 2px solid #e5e7eb;
}

.form__grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 1.25rem;
  margin-bottom: 1.25rem;
}

@media (min-width: 1024px) {
  .form__grid {
    grid-template-columns: repeat(3, 1fr);
  }

  .form__grid--title {
    grid-template-columns: 2fr 1fr;
  }
}

.form__grid :deep(.input-wrapper) {
  margin-bottom: 0;
}

.form__grid--checkboxes {
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  align-items: start;
}

.form__actions {
  display: flex;
  gap: 1rem;
  margin-top: 2rem;
  padding-top: 1.5rem;
  border-top: 2px solid #e5e7eb;
  justify-content: flex-end;
}

.form__button {
  padding: 0.75rem 1.5rem;
  font-size: 1rem;
  font-weight: 600;
  border-radius: 0.5rem;
  border: none;
  cursor: pointer;
  transition: all 0.2s ease-in-out;
  font-family: inherit;
}

.form__button--primary {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  box-shadow: 0 4px 6px rgba(102, 126, 234, 0.3);
}

.form__button--primary:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 12px rgba(102, 126, 234, 0.4);
}

.form__button--primary:active {
  transform: translateY(0);
}

.form__button--secondary {
  background: #f3f4f6;
  color: #374151;
}

.form__button--secondary:hover {
  background: #e5e7eb;
}

.form__button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  transform: none !important;
}

.form__button-spinner {
  display: inline-block;
  width: 1rem;
  height: 1rem;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
  margin-right: 0.5rem;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.form__error {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 1rem;
  background-color: #fef2f2;
  border: 1px solid #fecaca;
  border-radius: 0.5rem;
  color: #dc2626;
  font-size: 0.875rem;
  margin-top: 1rem;
}

.form__error-icon {
  width: 1.25rem;
  height: 1.25rem;
  flex-shrink: 0;
}

.form__success {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 1rem;
  background-color: #f0fdf4;
  border: 1px solid #bbf7d0;
  border-radius: 0.5rem;
  color: #16a34a;
  font-size: 0.875rem;
  margin-top: 1rem;
}

.form__success-icon {
  width: 1.25rem;
  height: 1.25rem;
  flex-shrink: 0;
}

@media (min-width: 1024px) {
  .form-container {
    padding: 2rem 3rem;
  }

  .form {
    padding: 2rem 2.5rem;
  }
}

@media (min-width: 1440px) {
  .form-container {
    padding: 2rem 4rem;
  }

  .form {
    padding: 2rem 3rem;
  }
}

@media (max-width: 768px) {
  .form-container {
    padding: 1rem;
  }

  .form {
    padding: 1.5rem;
  }

  .form__grid {
    grid-template-columns: 1fr;
  }

  .form__actions {
    flex-direction: column;
  }

  .form__button {
    width: 100%;
  }
}
</style>
