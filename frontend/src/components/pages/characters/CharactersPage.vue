<template>
  <div class="characters-page">
    <div v-if="dataStore.charactersLoaded">
      <div class="characters-page__header">
        <h1 class="characters-page__title">Персонажи</h1>
        <button
          class="toggle-button"
          @click="toggleForm"
          :class="{ 'toggle-button--active': showForm }"
        >
          {{ showForm ? 'Скрыть форму' : 'Показать форму' }}
        </button>
      </div>

      <transition name="fade">
        <div v-if="showForm" class="characters-page__form">
          <AddCharacterForm
            :character-uuid="editingCharacterUuid"
            @saved="handleFormSaved"
            @cancelled="handleFormCancelled"
          />
        </div>
      </transition>

      <div class="characters-grid">
        <div
          v-for="character in Object.values(dataStore.characters)"
          :key="character.uuid"
          class="character-card"
        >
          <div class="character-card__content">
            <h2 class="character-card__name">{{ character.name_ru || 'Без имени' }}</h2>
            <p v-if="character.name_en" class="character-card__name-en">{{ character.name_en }}</p>
            <p v-if="character.name_original" class="character-card__name-original">
              {{ character.name_original }}
            </p>
            <div class="character-card__meta">
              <span class="character-card__badge character-card__badge--gender">{{
                getGenderLabel(character.gender)
              }}</span>
              <span v-if="character.age" class="character-card__badge"
                >{{ character.age }} лет</span
              >
              <span v-if="character.multiplier !== undefined" class="character-card__badge"
                >×{{ character.multiplier.toFixed(2) }}</span
              >
            </div>
            <div v-if="getTitleName(character.title_uuid)" class="character-card__title">
              <span class="character-card__title-label">Тайтл:</span>
              <span class="character-card__title-name">{{ getTitleName(character.title_uuid) }}</span>
            </div>
            <div class="character-card__actions">
              <button
                class="character-card__edit-button"
                @click.stop="editCharacter(character.uuid)"
                title="Редактировать"
              >
                <svg viewBox="0 0 20 20" fill="currentColor">
                  <path
                    d="M13.586 3.586a2 2 0 112.828 2.828l-.793.793-2.828-2.828.793-.793zM11.379 5.793L3 14.172V17h2.828l8.38-8.379-2.83-2.828z"
                  />
                </svg>
              </button>
            </div>
          </div>
        </div>
      </div>

      <div v-if="Object.keys(dataStore.characters).length === 0" class="empty-state">
        <p class="empty-state__text">Пока нет персонажей. Добавьте первого!</p>
      </div>
    </div>

    <div v-else class="loading-state">
      <div class="spinner"></div>
      <p class="loading-state__text">Загрузка...</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import AddCharacterForm from '@/components/forms/AddCharacterForm.vue'
import { useDataStore } from '@/stores/dataStore'
import { onMounted, ref } from 'vue'

const dataStore = useDataStore()

const showForm = ref(true)
const editingCharacterUuid = ref<string | undefined>(undefined)

const editCharacter = (uuid: string) => {
  editingCharacterUuid.value = uuid
  showForm.value = true
}

const handleFormSaved = () => {
  editingCharacterUuid.value = undefined
  showForm.value = false
}

const handleFormCancelled = () => {
  editingCharacterUuid.value = undefined
}

const toggleForm = () => {
  if (showForm.value && !editingCharacterUuid.value) {
    showForm.value = false
  } else {
    editingCharacterUuid.value = undefined
    showForm.value = true
  }
}

const getGenderLabel = (gender: string) => {
  const labels: Record<string, string> = {
    муж: 'Мужской',
    жен: 'Женский',
    другое: 'Другое'
  }
  return labels[gender] || gender
}

const getTitleName = (titleUuid: string) => {
  const title = dataStore.titles[titleUuid]
  if (!title) return null
  return title.name_ru || title.name_en || title.name_original || 'Без названия'
}

onMounted(() => {
  dataStore.loadCharacters()
  // Загружаем тайтлы, если они еще не загружены (для отображения названий)
  if (!dataStore.loaded) {
    dataStore.loadData()
  }
})
</script>

<style scoped>
.characters-page {
  max-width: 1400px;
  width: 100%;
  margin: 0 auto;
  padding: 2rem;
  box-sizing: border-box;
}

.characters-page__header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
  flex-wrap: wrap;
  gap: 1rem;
}

.characters-page__title {
  font-size: 2rem;
  font-weight: 700;
  color: #1f2937;
  margin: 0;
}

.toggle-button {
  padding: 0.75rem 1.5rem;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border: none;
  border-radius: 0.5rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease-in-out;
  box-shadow: 0 4px 6px rgba(102, 126, 234, 0.3);
}

.toggle-button:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 12px rgba(102, 126, 234, 0.4);
}

.toggle-button--active {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
}

.characters-page__form {
  margin-bottom: 2rem;
  width: 100%;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

.characters-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 1.5rem;
  margin-top: 2rem;
}

.character-card {
  background: white;
  border-radius: 0.75rem;
  padding: 1.5rem;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  transition: all 0.3s ease-in-out;
  border: 1px solid #e5e7eb;
  cursor: pointer;
}

.character-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.12);
  border-color: #667eea;
}

.character-card__content {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.character-card__name {
  font-size: 1.25rem;
  font-weight: 700;
  color: #1f2937;
  margin: 0;
  line-height: 1.4;
}

.character-card__name-en {
  font-size: 0.95rem;
  color: #6b7280;
  font-style: italic;
  margin: 0;
}

.character-card__name-original {
  font-size: 0.875rem;
  color: #9ca3af;
  margin: 0;
}

.character-card__meta {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
  margin-top: 0.5rem;
}

.character-card__badge {
  display: inline-block;
  padding: 0.25rem 0.75rem;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border-radius: 1rem;
  font-size: 0.75rem;
  font-weight: 600;
}

.character-card__badge--gender {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
}

.character-card__title {
  margin-top: 0.5rem;
  padding-top: 0.75rem;
  border-top: 1px solid #e5e7eb;
  display: flex;
  gap: 0.5rem;
  align-items: center;
}

.character-card__title-label {
  font-size: 0.75rem;
  color: #6b7280;
  font-weight: 600;
  text-transform: uppercase;
}

.character-card__title-name {
  font-size: 0.875rem;
  color: #374151;
  font-weight: 500;
}

.character-card__actions {
  margin-top: 1rem;
  padding-top: 0.75rem;
  border-top: 1px solid #e5e7eb;
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
}

.character-card__edit-button {
  padding: 0.5rem;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border: none;
  border-radius: 0.5rem;
  cursor: pointer;
  transition: all 0.2s ease-in-out;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 2px 4px rgba(102, 126, 234, 0.3);
}

.character-card__edit-button:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 8px rgba(102, 126, 234, 0.4);
}

.character-card__edit-button svg {
  width: 1rem;
  height: 1rem;
}

.empty-state {
  text-align: center;
  padding: 4rem 2rem;
  background: white;
  border-radius: 0.75rem;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.empty-state__text {
  font-size: 1.125rem;
  color: #6b7280;
  margin: 0;
}

.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 4rem 2rem;
  gap: 1rem;
}

.spinner {
  width: 48px;
  height: 48px;
  border: 4px solid #e5e7eb;
  border-top-color: #667eea;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.loading-state__text {
  font-size: 1.125rem;
  color: #6b7280;
  margin: 0;
}

@media (max-width: 768px) {
  .characters-page {
    padding: 1rem;
  }

  .characters-page__title {
    font-size: 1.5rem;
  }

  .characters-grid {
    grid-template-columns: 1fr;
  }
}
</style>
