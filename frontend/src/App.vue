<script setup>
// Vue 3 Composition API - this is the modern way to write Vue components
// Everything in <script setup> is automatically available in the template

import { ref, onMounted } from 'vue'

// API URL - we'll set this via environment variable later
const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080'

// ref() creates "reactive" state - when it changes, the UI updates automatically
// This is similar to useState() in React, or just variables in vanilla JS that auto-update the DOM
const items = ref([])           // Our list of shopping items
const newItemName = ref('')     // The text in the input field
const loading = ref(true)       // Show loading state
const error = ref(null)         // Store any errors

// Functions to interact with the API

async function fetchItems() {
  try {
    loading.value = true
    error.value = null

    const response = await fetch(`${API_URL}/api/items`)
    if (!response.ok) throw new Error('Failed to fetch items')

    items.value = await response.json()
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

async function addItem() {
  // Don't add empty items
  if (!newItemName.value.trim()) return

  try {
    const response = await fetch(`${API_URL}/api/items`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name: newItemName.value.trim() })
    })

    if (!response.ok) throw new Error('Failed to add item')

    const newItem = await response.json()
    // Add to the beginning of the list (newest first)
    items.value.unshift(newItem)
    // Clear the input
    newItemName.value = ''
  } catch (e) {
    error.value = e.message
  }
}

async function toggleItem(item) {
  try {
    const response = await fetch(`${API_URL}/api/items/${item.id}`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ checked: !item.checked })
    })

    if (!response.ok) throw new Error('Failed to update item')

    // Update the item in our local list
    item.checked = !item.checked
  } catch (e) {
    error.value = e.message
  }
}

async function deleteItem(item) {
  try {
    const response = await fetch(`${API_URL}/api/items/${item.id}`, {
      method: 'DELETE'
    })

    if (!response.ok) throw new Error('Failed to delete item')

    // Remove from local list
    items.value = items.value.filter(i => i.id !== item.id)
  } catch (e) {
    error.value = e.message
  }
}

// onMounted runs once when the component first appears on screen
// Similar to componentDidMount in React, or DOMContentLoaded in vanilla JS
onMounted(() => {
  fetchItems()
})
</script>

<template>
  <!-- Vue templates use a special syntax for dynamic content -->
  <!-- {{ }} outputs values, v-if shows/hides, v-for loops, @ handles events -->

  <div class="container">
    <h1>Shopping List</h1>

    <!-- Error message -->
    <div v-if="error" class="error">
      {{ error }}
      <button @click="error = null">Dismiss</button>
    </div>

    <!-- Add item form -->
    <!-- v-model creates two-way binding: typing updates newItemName, and vice versa -->
    <!-- @submit.prevent handles form submit and prevents page reload -->
    <form @submit.prevent="addItem" class="add-form">
      <input
        v-model="newItemName"
        type="text"
        placeholder="Add an item..."
        class="input"
      />
      <button type="submit" class="btn btn-primary">Add</button>
    </form>

    <!-- Loading state -->
    <p v-if="loading" class="loading">Loading...</p>

    <!-- Empty state -->
    <p v-else-if="items.length === 0" class="empty">
      Your shopping list is empty. Add some items!
    </p>

    <!-- Items list -->
    <!-- v-for loops through the array. :key helps Vue track which items changed -->
    <ul v-else class="items-list">
      <li
        v-for="item in items"
        :key="item.id"
        :class="{ checked: item.checked }"
        class="item"
      >
        <label class="item-label">
          <input
            type="checkbox"
            :checked="item.checked"
            @change="toggleItem(item)"
          />
          <span class="item-name">{{ item.name }}</span>
        </label>
        <button @click="deleteItem(item)" class="btn btn-delete">Delete</button>
      </li>
    </ul>
  </div>
</template>

<style scoped>
/* scoped means these styles only apply to this component */

.container {
  max-width: 500px;
  margin: 2rem auto;
  padding: 0 1rem;
  font-family: system-ui, -apple-system, sans-serif;
}

h1 {
  text-align: center;
  color: #333;
}

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
  background: #3aa876;
}

.btn-delete {
  background: #ff6b6b;
  color: white;
  padding: 0.5rem 0.75rem;
  font-size: 0.875rem;
}

.btn-delete:hover {
  background: #ee5a5a;
}

.loading, .empty {
  text-align: center;
  color: #666;
}

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
</style>
