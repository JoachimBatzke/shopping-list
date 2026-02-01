<script setup>
import { ref, computed, onMounted, onUnmounted, nextTick, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import Sortable from 'sortablejs'

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
import editGray from '@/assets/icons/edit_gray.svg'
import dragWhite from '@/assets/icons/drag_white.svg'
import dragGray from '@/assets/icons/drag_gray.svg'
import menuWhite from '@/assets/icons/menu_white.svg'
import menuBlack from '@/assets/icons/menu_black.svg'

const linkIcon = computed(() => isLightBackground.value ? linkBlack : linkWhite)
const addIcon = computed(() => isLightBackground.value ? addBlack : addWhite)
const menuIcon = computed(() => isLightBackground.value ? menuBlack : menuWhite)

// Edit state
const editingItemId = ref(null)
const editingItemName = ref('')

// Menu state
const showMenu = ref(false)

// Recommendations state
const recommendations = ref([])

// System dark mode detection for item icons
const isDarkMode = ref(window.matchMedia('(prefers-color-scheme: dark)').matches)
const editIcon = editGray
const dragIcon = computed(() => isDarkMode.value ? dragWhite : dragGray)

// Sortable reference
const itemsListRef = ref(null)
let sortableInstance = null
let isReordering = false

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

// Edit item functions
function startEditItem(item) {
  editingItemId.value = item.id
  editingItemName.value = item.name
}

function cancelEdit() {
  editingItemId.value = null
  editingItemName.value = ''
}

async function saveEditItem(item) {
  if (!editingItemName.value.trim()) {
    cancelEdit()
    return
  }
  try {
    const response = await fetch(`${API_URL}/api/lists/${props.id}/items/${item.id}`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name: editingItemName.value.trim() })
    })
    if (!response.ok) throw new Error('Failed to update item')
    item.name = editingItemName.value.trim()
    cancelEdit()
  } catch (e) {
    error.value = e.message
  }
}

// Reorder items
async function reorderItems(newOrder) {
  try {
    const response = await fetch(`${API_URL}/api/lists/${props.id}/items/reorder`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ item_ids: newOrder })
    })
    if (!response.ok) throw new Error('Failed to reorder items')
  } catch (e) {
    error.value = e.message
    // Refetch items to restore order on error
    await fetchItems()
  }
}

// Initialize Sortable.js
function initSortable() {
  if (sortableInstance) {
    sortableInstance.destroy()
  }
  if (itemsListRef.value) {
    sortableInstance = Sortable.create(itemsListRef.value, {
      handle: '.drag-handle',
      animation: 150,
      ghostClass: 'ghost',
      onEnd: async (evt) => {
        // Prevent watcher from reinitializing during reorder
        isReordering = true
        // Reorder items array based on new positions
        const movedItem = items.value.splice(evt.oldIndex, 1)[0]
        items.value.splice(evt.newIndex, 0, movedItem)
        // Send new order to backend
        const newOrder = items.value.map(item => item.id)
        await reorderItems(newOrder)
        isReordering = false
      }
    })
  }
}

// Menu functions
function toggleMenu() {
  showMenu.value = !showMenu.value
}

function closeMenu() {
  showMenu.value = false
}

// Share functionality
function shareList() {
  const url = window.location.href
  navigator.clipboard.writeText(url)
  alert(t('link_copied'))
  closeMenu()
}

// Go back to create new list
function createNewList() {
  closeMenu()
  router.push('/')
}

// Refresh list
async function refreshList() {
  closeMenu()
  await fetchItems()
}

// Delete list
async function deleteList() {
  if (!confirm(t('delete_list_confirm') || 'Are you sure you want to delete this list?')) {
    return
  }
  try {
    const response = await fetch(`${API_URL}/api/lists/${props.id}`, {
      method: 'DELETE'
    })
    if (!response.ok) throw new Error('Failed to delete list')
    router.push('/')
  } catch (e) {
    error.value = e.message
  }
  closeMenu()
}

// Fetch recommendations
async function fetchRecommendations() {
  try {
    const response = await fetch(`${API_URL}/api/lists/${props.id}/recommendations`)
    if (!response.ok) return
    recommendations.value = await response.json()
  } catch (e) {
    // Silently fail - recommendations are optional
  }
}

// Add item from recommendation
async function addFromRecommendation(name) {
  try {
    const response = await fetch(`${API_URL}/api/lists/${props.id}/items`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name })
    })
    if (!response.ok) throw new Error('Failed to add item')
    const newItem = await response.json()
    items.value.push(newItem)
    // Remove from recommendations
    recommendations.value = recommendations.value.filter(r => r.name !== name)
  } catch (e) {
    error.value = e.message
  }
}

// Dismiss recommendation
async function dismissRecommendation(name) {
  try {
    await fetch(`${API_URL}/api/lists/${props.id}/recommendations/${encodeURIComponent(name)}/dismiss`, {
      method: 'POST'
    })
    recommendations.value = recommendations.value.filter(r => r.name !== name)
  } catch (e) {
    // Silently fail
  }
}

// PWA: Update manifest and meta tags for this specific list
function updatePWAForList(listData) {
  if (!listData) return

  // Update theme-color meta tag
  const themeColorMeta = document.querySelector('meta[name="theme-color"]')
  if (themeColorMeta) {
    themeColorMeta.setAttribute('content', `#${listData.hex_color}`)
  }

  // Update manifest link to point to dynamic manifest from backend
  let manifestLink = document.querySelector('link[rel="manifest"]')
  if (manifestLink) {
    manifestLink.setAttribute('href', `${API_URL}/api/lists/${listData.id}/manifest.webmanifest`)
  }

  // Update apple-touch-icon to dynamic icon from backend
  let appleTouchIcon = document.querySelector('link[rel="apple-touch-icon"]')
  if (appleTouchIcon) {
    appleTouchIcon.setAttribute('href', `${API_URL}/api/lists/${listData.id}/icon/180.png`)
  }

  // Update iOS app title (iOS uses this meta tag, not manifest short_name)
  const appleTitle = document.querySelector('meta[name="apple-mobile-web-app-title"]')
  if (appleTitle) {
    appleTitle.setAttribute('content', listData.name)
  }

  // Update page title
  document.title = `${listData.name} - JORLIST`
}

// PWA: Reset to defaults when leaving page
function resetPWADefaults() {
  const themeColorMeta = document.querySelector('meta[name="theme-color"]')
  if (themeColorMeta) {
    themeColorMeta.setAttribute('content', '#333333')
  }

  const manifestLink = document.querySelector('link[rel="manifest"]')
  if (manifestLink) {
    manifestLink.setAttribute('href', '/manifest.webmanifest')
  }

  const appleTouchIcon = document.querySelector('link[rel="apple-touch-icon"]')
  if (appleTouchIcon) {
    appleTouchIcon.setAttribute('href', '/apple-touch-icon.png')
  }

  const appleTitle = document.querySelector('meta[name="apple-mobile-web-app-title"]')
  if (appleTitle) {
    appleTitle.setAttribute('content', 'JORLIST')
  }

  document.title = 'JORLIST - Share a list with friends'
}

onMounted(async () => {
  // Listen for dark mode changes
  window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
    isDarkMode.value = e.matches
  })

  await fetchList()
  if (!notFound.value && list.value) {
    await fetchItems()
    await fetchRecommendations()
    // Update PWA manifest and icons for this list
    updatePWAForList(list.value)
  }
  // Set loading false first so the ul element is rendered
  loading.value = false

  // Initialize sortable after content is visible
  if (!notFound.value && list.value && items.value.length > 0) {
    await nextTick()
    initSortable()
  }
})

// Re-init sortable when items change (e.g., after adding item)
// Skip during reorder to avoid destroying Sortable mid-drag
watch(items, async () => {
  if (isReordering) return
  await nextTick()
  initSortable()
}, { deep: true })

// Reset PWA defaults when leaving the page
onUnmounted(() => {
  resetPWADefaults()
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
  <div v-else-if="list" class="container">
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
          @click="toggleMenu"
          class="btn-icon"
          :title="t('menu') || 'Menu'"
        >
          <img :src="menuIcon" alt="Menu" class="icon" />
        </button>
        <!-- Dropdown menu -->
        <div v-if="showMenu" class="menu-overlay" @click="closeMenu"></div>
        <div v-if="showMenu" class="dropdown-menu">
          <button @click="shareList" class="menu-item">
            <img :src="linkIcon" alt="" class="menu-icon" />
            <span>{{ t('share_btn') }}</span>
          </button>
          <button @click="refreshList" class="menu-item">
            <img src="@/assets/icons/refresh_white.svg" alt="" class="menu-icon" />
            <span>{{ t('refresh') || 'Refresh' }}</span>
          </button>
          <button @click="toggleLanguage" class="menu-item">
            <span class="menu-icon-text">{{ locale === 'en' ? 'DE' : 'EN' }}</span>
            <span>{{ t('language') || 'Language' }}</span>
          </button>
          <button @click="createNewList" class="menu-item">
            <img :src="addIcon" alt="" class="menu-icon" />
            <span>{{ t('new_list') }}</span>
          </button>
          <hr class="menu-divider" />
          <button @click="deleteList" class="menu-item menu-item-danger">
            <img src="@/assets/icons/trash_white.svg" alt="" class="menu-icon" />
            <span>{{ t('delete_list') || 'Delete List' }}</span>
          </button>
        </div>
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
        class="btn-add-circle"
        :style="{ backgroundColor: '#' + list.hex_color, color: headerTextColor }"
        :title="t('add_btn')"
      >
        +
      </button>
    </form>

    <!-- Recommendations -->
    <div v-if="recommendations.length > 0" class="recommendations">
      <div
        v-for="rec in recommendations"
        :key="rec.name"
        class="recommendation-pill"
        :style="{ backgroundColor: '#' + list.hex_color, color: headerTextColor }"
        @click="addFromRecommendation(rec.name)"
      >
        <span class="rec-name">{{ rec.name }}</span>
        <button
          class="rec-dismiss"
          @click.stop="dismissRecommendation(rec.name)"
          :style="{ color: headerTextColor }"
        >
          ×
        </button>
      </div>
    </div>

    <!-- Empty state -->
    <p v-if="items.length === 0" class="empty">
      {{ t('empty_list') }}
    </p>

    <!-- Items list -->
    <ul v-else class="items-list" ref="itemsListRef">
      <li
        v-for="item in items"
        :key="item.id"
        :class="{ checked: item.checked, separator: item.is_separator, editing: editingItemId === item.id }"
        class="item"
        :data-id="item.id"
      >
        <!-- Edit mode -->
        <div v-if="editingItemId === item.id" class="edit-form">
          <input
            v-model="editingItemName"
            type="text"
            class="edit-input"
            @keyup.enter="saveEditItem(item)"
            @keyup.escape="cancelEdit"
            @blur="saveEditItem(item)"
            ref="editInput"
            autofocus
          />
        </div>
        <!-- Normal mode -->
        <template v-else>
          <label v-if="!item.is_separator" class="item-label">
            <input
              type="checkbox"
              :checked="item.checked"
              @change="toggleItem(item)"
              class="checkbox-hidden"
            />
            <span
              class="checkmark"
              :class="{ checked: item.checked }"
              :style="{ '--list-color': '#' + list.hex_color }"
            ></span>
            <span class="item-name">{{ item.name }}</span>
          </label>
          <hr v-else class="separator-line" />
          <!-- Edit button for unchecked items -->
          <button
            v-if="!item.checked && !item.is_separator"
            @click="startEditItem(item)"
            class="btn-action"
            :title="t('edit_btn') || 'Edit'"
          >
            <img :src="editIcon" alt="Edit" class="icon-small" />
          </button>
          <!-- Delete button for checked items -->
          <button
            v-if="item.checked || item.is_separator"
            @click="deleteItem(item)"
            class="btn-delete"
            :title="t('delete_btn')"
          >
            <img src="@/assets/icons/delete_red.svg" alt="Delete" class="icon-small" />
          </button>
          <!-- Drag handle -->
          <span class="drag-handle">
            <img :src="dragIcon" alt="Drag" class="icon-small" />
          </span>
        </template>
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
  color: var(--text-primary);
  margin-bottom: 1rem;
}

.not-found p {
  color: var(--text-secondary);
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

/* Dropdown menu */
.menu-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 15;
}

.dropdown-menu {
  position: absolute;
  top: 100%;
  right: 0;
  margin-top: 0.5rem;
  min-width: 180px;
  background: var(--bg-item);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  z-index: 20;
  overflow: hidden;
  backdrop-filter: blur(10px);
}

@media (prefers-color-scheme: dark) {
  .dropdown-menu {
    background: rgba(40, 40, 40, 0.95);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.4);
  }
}

.menu-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  width: 100%;
  padding: 0.75rem 1rem;
  border: none;
  background: transparent;
  color: var(--text-primary);
  font-size: 0.9rem;
  text-align: left;
  cursor: pointer;
  transition: background 0.15s;
}

.menu-item:hover {
  background: rgba(128, 128, 128, 0.1);
}

.menu-item-danger {
  color: #e53935;
}

.menu-item-danger:hover {
  background: rgba(229, 57, 53, 0.1);
}

.menu-icon {
  width: 18px;
  height: 18px;
  opacity: 0.7;
}

.menu-icon-text {
  width: 18px;
  font-size: 0.7rem;
  font-weight: 600;
  text-align: center;
}

.menu-divider {
  border: none;
  border-top: 1px solid var(--border-color);
  margin: 0.25rem 0;
}

/* Recommendations */
.recommendations {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin-bottom: 1rem;
  padding: 0.5rem 0;
}

.recommendation-pill {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.4rem 0.5rem 0.4rem 0.75rem;
  border-radius: 50px;
  font-size: 0.85rem;
  cursor: pointer;
  transition: transform 0.15s, opacity 0.15s;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.recommendation-pill:hover {
  transform: scale(1.02);
  opacity: 0.95;
}

.rec-name {
  white-space: nowrap;
}

.rec-dismiss {
  background: none;
  border: none;
  font-size: 1.2rem;
  line-height: 1;
  cursor: pointer;
  opacity: 0.7;
  padding: 0 0.25rem;
  transition: opacity 0.15s;
}

.rec-dismiss:hover {
  opacity: 1;
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

@media (prefers-color-scheme: dark) {
  .error {
    background: #3d1f1f;
    border-color: #5c2b2b;
    color: #ff6b6b;
  }
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
  border: 1px solid var(--border-input);
  border-radius: 4px;
  background: var(--bg-input);
  color: var(--text-primary);
}

.input:focus {
  outline: none;
  border-color: #42b883;
}

.input::placeholder {
  color: var(--text-muted);
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

/* Circular add button */
.btn-add-circle {
  width: 3rem;
  height: 3rem;
  border-radius: 50%;
  border: none;
  font-size: 2rem;
  font-weight: 300;
  line-height: 1;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  transition: transform 0.15s, opacity 0.15s;
}

.btn-add-circle:hover {
  opacity: 0.9;
  transform: scale(1.05);
}

.btn-add-circle:active {
  transform: scale(0.95);
}

.btn-action,
.btn-delete {
  width: 36px;
  height: 36px;
  padding: 8px;
  background: transparent;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.2s;
  flex-shrink: 0;
}

.btn-action:hover {
  background: rgba(128, 128, 128, 0.1);
}

.btn-delete:hover {
  background: rgba(255, 0, 0, 0.1);
}

/* Edit form */
.edit-form {
  flex: 1;
  display: flex;
}

.edit-input {
  flex: 1;
  padding: 0.5rem;
  font-size: 1rem;
  border: 2px solid var(--list-color, #ff9800);
  border-radius: 4px;
  background: var(--bg-input);
  color: var(--text-primary);
  outline: none;
}

.item.editing {
  padding: 0.5rem 0.75rem;
}

/* Drag handle */
.drag-handle {
  cursor: grab;
  opacity: 0.5;
  padding: 0.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: opacity 0.2s;
}

.drag-handle:hover {
  opacity: 1;
}

.drag-handle:active {
  cursor: grabbing;
}

/* Sortable ghost (dragging placeholder) */
.ghost {
  opacity: 0.4;
  background: var(--bg-item-checked);
}

.empty {
  text-align: center;
  color: var(--text-secondary);
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
  background: var(--bg-item);
  border: 1px solid var(--border-color);
  border-radius: 4px;
  margin-bottom: 0.5rem;
}

.item.checked {
  background: var(--bg-item-checked);
}

.item.checked .item-name {
  text-decoration: line-through;
  color: var(--text-muted);
}

.item.separator {
  padding: 0.25rem 0.75rem;
}

.separator-line {
  flex: 1;
  border: none;
  border-top: 1px solid var(--border-color);
  margin: 0 0.5rem;
}

.item-label {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  cursor: pointer;
  flex: 1;
  position: relative;
}

/* Hide native checkbox */
.checkbox-hidden {
  position: absolute;
  opacity: 0;
  width: 0;
  height: 0;
}

/* Custom checkbox */
.checkmark {
  width: 1.5rem;
  height: 1.5rem;
  border: 2px solid var(--list-color, #ff9800);
  border-radius: 0.25rem;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  transition: background-color 0.15s, border-color 0.15s;
}

.checkmark:hover {
  background-color: rgba(255, 152, 0, 0.1);
}

/* Checked state */
.checkmark.checked {
  background-color: var(--list-color, #ff9800);
}

/* Checkmark icon (white tick) */
.checkmark.checked::after {
  content: '';
  width: 0.4rem;
  height: 0.7rem;
  border: solid white;
  border-width: 0 2.5px 2.5px 0;
  transform: rotate(45deg);
  margin-bottom: 2px;
}

.item-name {
  font-size: 1rem;
  color: var(--text-primary);
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
    background: var(--bg-item);
    box-shadow: var(--shadow-input);
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
