<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'

const props = defineProps({
  id: {
    type: String,
    required: true
  }
})

const router = useRouter()
const { t, locale } = useI18n()
const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080'

// State
const list = ref(null)
const items = ref([])
const newItemName = ref('')
const loading = ref(true)
const error = ref(null)
const notFound = ref(false)

// Dynamic text color based on background brightness (threshold 0.7 matches original)
const isLightBackground = computed(() => {
  if (!list.value?.hex_color) return false
  const hex = list.value.hex_color
  const r = parseInt(hex.substr(0, 2), 16) / 255
  const g = parseInt(hex.substr(2, 2), 16) / 255
  const b = parseInt(hex.substr(4, 2), 16) / 255
  const lightness = (Math.max(r, g, b) + Math.min(r, g, b)) / 2
  return lightness > 0.7
})

const headerTextColor = computed(() => {
  return isLightBackground.value ? '#333333' : '#ffffff'
})

// Icon imports for dynamic switching
import linkWhite from '@/assets/icons/link_white.svg'
import linkBlack from '@/assets/icons/link_black.svg'
import addWhite from '@/assets/icons/add_white.svg'
import addBlack from '@/assets/icons/add_black.svg'

const linkIcon = computed(() => isLightBackground.value ? linkBlack : linkWhite)
const addIcon = computed(() => isLightBackground.value ? addBlack : addWhite)

// Language toggle
function toggleLanguage() {
  const newLocale = locale.value === 'en' ? 'de' : 'en'
  locale.value = newLocale
  localStorage.setItem('jorlist-locale', newLocale)
}

// Fetch single list
async function fetchList() {
  try {
    const response = await fetch(`${API_URL}/api/lists/${props.id}`)
    if (response.status === 404) {
      notFound.value = true
      return
    }
    if (!response.ok) throw new Error('Failed to fetch list')
    list.value = await response.json()
  } catch (e) {
    error.value = e.message
  }
}

// Fetch items for this list
async function fetchItems() {
  try {
    const response = await fetch(`${API_URL}/api/lists/${props.id}/items`)
    if (!response.ok) throw new Error('Failed to fetch items')
    items.value = await response.json()
  } catch (e) {
    error.value = e.message
  }
}

// Item CRUD operations
async function addItem() {
  if (!newItemName.value.trim()) return
  try {
    const response = await fetch(`${API_URL}/api/lists/${props.id}/items`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name: newItemName.value.trim() })
    })
    if (!response.ok) throw new Error('Failed to add item')
    const newItem = await response.json()
    items.value.push(newItem)
    newItemName.value = ''
  } catch (e) {
    error.value = e.message
  }
}

async function toggleItem(item) {
  try {
    const response = await fetch(`${API_URL}/api/lists/${props.id}/items/${item.id}`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ checked: !item.checked })
    })
    if (!response.ok) throw new Error('Failed to update item')
    item.checked = !item.checked
  } catch (e) {
    error.value = e.message
  }
}

async function deleteItem(item) {
  try {
    const response = await fetch(`${API_URL}/api/lists/${props.id}/items/${item.id}`, {
      method: 'DELETE'
    })
    if (!response.ok) throw new Error('Failed to delete item')
    items.value = items.value.filter(i => i.id !== item.id)
  } catch (e) {
    error.value = e.message
  }
}

// Share functionality
function shareList() {
  const url = window.location.href
  navigator.clipboard.writeText(url)
  alert(t('link_copied'))
}

// Go back to create new list
function createNewList() {
  router.push('/')
}

onMounted(async () => {
  await fetchList()
  if (!notFound.value && list.value) {
    await fetchItems()
  }
  loading.value = false
})
</script>

<template>
  <!-- Loading state -->
  <div v-if="loading" class="loading-screen">
    <div class="spinner"></div>
    <p>{{ t('loading') }}</p>
  </div>

  <!-- Not found state -->
  <div v-else-if="notFound" class="not-found">
    <h1>{{ t('not_found_title') }}</h1>
    <p>{{ t('not_found_desc') }}</p>
    <button @click="createNewList" class="btn-primary">{{ t('create_btn') }}</button>
  </div>

  <!-- List content -->
  <div v-else class="container">
    <!-- Header with full-color background -->
    <header
      class="list-header"
      :style="{ backgroundColor: '#' + list.hex_color, color: headerTextColor }"
    >
      <div class="list-info">
        <span v-if="list.emoji" class="list-emoji">{{ list.emoji }}</span>
        <h1 :style="{ color: headerTextColor }">{{ list.name }}</h1>
      </div>
      <div class="header-actions">
        <button
          @click="shareList"
          class="btn-icon"
          :title="t('share_btn')"
        >
          <img :src="linkIcon" alt="Share" class="icon" />
        </button>
        <button
          @click="createNewList"
          class="btn-icon"
          :title="t('new_list')"
        >
          <img :src="addIcon" alt="New" class="icon" />
        </button>
        <button
          @click="toggleLanguage"
          class="btn-lang"
          :style="{ color: headerTextColor, borderColor: headerTextColor }"
        >
          {{ locale === 'en' ? 'DE' : 'EN' }}
        </button>
      </div>
    </header>

    <!-- Error message -->
    <div v-if="error" class="error">
      {{ error }}
      <button @click="error = null">Dismiss</button>
    </div>

    <!-- Add item form -->
    <form @submit.prevent="addItem" class="add-form">
      <input
        v-model="newItemName"
        type="text"
        :placeholder="t('add_item')"
        class="input"
      />
      <button
        type="submit"
        class="btn btn-primary"
        :style="{ backgroundColor: '#' + list.hex_color, color: headerTextColor }"
      >
        {{ t('add_btn') }}
      </button>
    </form>

    <!-- Empty state -->
    <p v-if="items.length === 0" class="empty">
      {{ t('empty_list') }}
    </p>

    <!-- Items list -->
    <ul v-else class="items-list">
      <li
        v-for="item in items"
        :key="item.id"
        :class="{ checked: item.checked, separator: item.is_separator }"
        class="item"
      >
        <label v-if="!item.is_separator" class="item-label">
          <input
            type="checkbox"
            :checked="item.checked"
            @change="toggleItem(item)"
          />
          <span class="item-name">{{ item.name }}</span>
        </label>
        <hr v-else class="separator-line" />
        <button @click="deleteItem(item)" class="btn-delete" :title="t('delete_btn')">
          <img src="@/assets/icons/delete_red.svg" alt="Delete" class="icon-small" />
        </button>
      </li>
    </ul>
  </div>
</template>

<style scoped>
.container {
  max-width: 500px;
  margin: 0 auto;
  padding: 0 1rem 1rem;
  font-family: system-ui, -apple-system, sans-serif;
  position: relative;
}

.loading-screen,
.not-found {
  text-align: center;
  padding: 4rem 1rem;
  font-family: system-ui, -apple-system, sans-serif;
}

.not-found h1 {
  color: #333;
  margin-bottom: 1rem;
}

.not-found p {
  color: #666;
  margin-bottom: 2rem;
}

/* Loading spinner */
.spinner {
  width: 40px;
  height: 40px;
  border: 3px solid #e0e0e0;
  border-top-color: #42b883;
  border-radius: 50%;
  margin: 0 auto 1rem;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* Header with full-color background */
.list-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem;
  margin: 0 -1rem 1.5rem;
  position: sticky;
  top: 0;
  z-index: 10;
}

.list-info {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.list-emoji {
  font-size: 1.75rem;
}

.list-header h1 {
  font-size: 1.25rem;
  margin: 0;
  font-weight: 600;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 0.25rem;
}

/* Icon buttons */
.btn-icon {
  width: 40px;
  height: 40px;
  padding: 8px;
  background: rgba(255, 255, 255, 0.2);
  border: none;
  border-radius: 8px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.2s;
}

.btn-icon:hover {
  background: rgba(255, 255, 255, 0.3);
}

.icon {
  width: 24px;
  height: 24px;
}

.icon-small {
  width: 20px;
  height: 20px;
}

/* Language button in header */
.btn-lang {
  padding: 0.35rem 0.75rem;
  font-size: 0.75rem;
  font-weight: 600;
  background: transparent;
  border: 1px solid;
  border-radius: 4px;
  cursor: pointer;
  margin-left: 0.25rem;
}

.btn-lang:hover {
  background: rgba(255, 255, 255, 0.1);
}

/* Error */
.error {
  background: #fee;
  border: 1px solid #fcc;
  color: #c00;
  padding: 0.75rem;
  border-radius: 4px;
  margin-bottom: 1rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

/* Add form */
.add-form {
  display: flex;
  gap: 0.5rem;
  margin-bottom: 1.5rem;
}

.input {
  flex: 1;
  padding: 0.75rem;
  font-size: 1rem;
  border: 1px solid #ddd;
  border-radius: 4px;
}

.input:focus {
  outline: none;
  border-color: #42b883;
}

.btn {
  padding: 0.75rem 1rem;
  font-size: 1rem;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.btn-primary {
  background: #42b883;
  color: white;
}

.btn-primary:hover {
  opacity: 0.9;
}

.btn-delete {
  width: 36px;
  height: 36px;
  padding: 8px;
  background: #fff0f0;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.2s;
  flex-shrink: 0;
}

.btn-delete:hover {
  background: #ffe0e0;
}

.empty {
  text-align: center;
  color: #666;
}

/* Items list */
.items-list {
  list-style: none;
  padding: 0;
  margin: 0;
}

.item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem;
  border: 1px solid #eee;
  border-radius: 4px;
  margin-bottom: 0.5rem;
}

.item.checked {
  background: #f9f9f9;
}

.item.checked .item-name {
  text-decoration: line-through;
  color: #999;
}

.item.separator {
  padding: 0.25rem 0.75rem;
}

.separator-line {
  flex: 1;
  border: none;
  border-top: 1px solid #ddd;
  margin: 0 0.5rem;
}

.item-label {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  cursor: pointer;
  flex: 1;
}

.item-label input[type="checkbox"] {
  width: 1.25rem;
  height: 1.25rem;
  cursor: pointer;
}

.item-name {
  font-size: 1rem;
}

/* Mobile responsive - Bottom input placement */
@media (max-width: 640px) {
  .container {
    padding-bottom: 80px; /* Space for fixed input */
  }

  .add-form {
    position: fixed;
    bottom: 0;
    left: 0;
    right: 0;
    margin: 0;
    padding: 12px 16px;
    background: white;
    box-shadow: 0 -2px 10px rgba(0, 0, 0, 0.1);
    border-radius: 0;
    z-index: 20;
  }

  .list-header h1 {
    font-size: 1.1rem;
  }

  .list-emoji {
    font-size: 1.5rem;
  }

  .btn-icon {
    width: 36px;
    height: 36px;
    padding: 6px;
  }

  .icon {
    width: 20px;
    height: 20px;
  }
}
</style>
