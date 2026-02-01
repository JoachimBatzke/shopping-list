<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import iro from '@jaames/iro'
import { emojicompact } from '@/data/emojis.js'

const router = useRouter()
const { t, locale } = useI18n()
const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080'

// Wizard state
const currentStep = ref(1)
const listName = ref('')
const selectedEmoji = ref('\u{1F6D2}')
const selectedColor = ref('42b883')
const isCreating = ref(false)
const error = ref(null)

// Color picker reference
const colorPickerEl = ref(null)
let colorPicker = null

// Dynamic text color based on brightness
const textColor = computed(() => {
  const hex = selectedColor.value
  const r = parseInt(hex.substr(0, 2), 16) / 255
  const g = parseInt(hex.substr(2, 2), 16) / 255
  const b = parseInt(hex.substr(4, 2), 16) / 255
  const lightness = (Math.max(r, g, b) + Math.min(r, g, b)) / 2
  return lightness > 0.7 ? '#333333' : '#ffffff'
})

// Language toggle
function toggleLanguage() {
  const newLocale = locale.value === 'en' ? 'de' : 'en'
  locale.value = newLocale
  localStorage.setItem('jorlist-locale', newLocale)
}

// Random emoji
function randomEmoji() {
  const randomIndex = Math.floor(Math.random() * emojicompact.length)
  selectedEmoji.value = emojicompact[randomIndex]
}

// Navigation
function nextStep() {
  if (currentStep.value === 1 && !listName.value.trim()) {
    error.value = t('error_name_required')
    return
  }
  error.value = null
  currentStep.value++

  // Initialize color picker when entering step 3
  if (currentStep.value === 3) {
    setTimeout(initColorPicker, 50)
  }
}

function prevStep() {
  error.value = null
  currentStep.value--
}

// Initialize iro.js color picker
function initColorPicker() {
  if (colorPicker) {
    colorPicker.off('color:change', onColorChange)
  }

  if (colorPickerEl.value) {
    colorPicker = new iro.ColorPicker(colorPickerEl.value, {
      width: 200,
      color: '#' + selectedColor.value,
      layout: [
        {
          component: iro.ui.Wheel,
          options: {}
        },
        {
          component: iro.ui.Slider,
          options: {
            sliderType: 'value'
          }
        }
      ]
    })

    colorPicker.on('color:change', onColorChange)
  }
}

function onColorChange(color) {
  selectedColor.value = color.hexString.substring(1)
}

// Create list and redirect
async function createList() {
  if (isCreating.value) return

  isCreating.value = true
  error.value = null

  try {
    const response = await fetch(`${API_URL}/api/lists`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        name: listName.value.trim(),
        emoji: selectedEmoji.value,
        hex_color: selectedColor.value
      })
    })

    if (!response.ok) throw new Error(t('error_create_failed'))

    const newList = await response.json()
    router.push(`/list/${newList.id}`)
  } catch (e) {
    error.value = e.message
    isCreating.value = false
  }
}

onUnmounted(() => {
  if (colorPicker) {
    colorPicker.off('color:change', onColorChange)
  }
})
</script>

<template>
  <div class="landing">
    <!-- Language toggle -->
    <button class="lang-toggle" @click="toggleLanguage">
      {{ locale === 'en' ? 'DE' : 'EN' }}
    </button>

    <!-- Hero Section -->
    <header class="hero">
      <h1 class="title font-heavitas">JØRLIST</h1>
      <p class="subtitle">{{ t('subtitle') }}</p>
      <p class="tagline">{{ t('tagline') }}</p>

      <!-- Phone Mockup -->
      <div class="phone-mockup">
        <img src="@/assets/images/frame.png" alt="Phone frame" class="phone-frame" />
        <img src="@/assets/images/jorlist.gif" alt="Jorlist demo" class="phone-screen" />
      </div>

      <!-- Feature Cards -->
      <div class="features">
        <div class="feature-card">
          <span class="feature-icon">🔓</span>
          <h3>{{ t('feature1_title') }}</h3>
          <p>{{ t('feature1_desc') }}</p>
        </div>
        <div class="feature-card">
          <span class="feature-icon">📱</span>
          <h3>{{ t('feature2_title') }}</h3>
          <p>{{ t('feature2_desc') }}</p>
        </div>
        <div class="feature-card">
          <span class="feature-icon">🔗</span>
          <h3>{{ t('feature3_title') }}</h3>
          <p>{{ t('feature3_desc') }}</p>
        </div>
      </div>
    </header>

    <!-- Creation Wizard -->
    <main class="wizard">
      <div class="wizard-container">
        <!-- Progress indicator -->
        <div class="progress">
          <span :class="{ active: currentStep >= 1 }">1</span>
          <span :class="{ active: currentStep >= 2 }">2</span>
          <span :class="{ active: currentStep >= 3 }">3</span>
        </div>

        <!-- Error message -->
        <div v-if="error" class="error">{{ error }}</div>

        <!-- Step 1: Name -->
        <div v-if="currentStep === 1" class="step">
          <h2>1. {{ t('step1_title') }}</h2>
          <p>{{ t('step1_desc') }}</p>
          <input
            v-model="listName"
            type="text"
            maxlength="15"
            :placeholder="t('step1_placeholder')"
            class="name-input"
            @keyup.enter="nextStep"
          />
          <button @click="nextStep" class="btn-next">{{ t('next') }}</button>
        </div>

        <!-- Step 2: Emoji -->
        <div v-if="currentStep === 2" class="step">
          <h2>2. {{ t('step2_title') }}</h2>
          <p>{{ t('step2_desc') }}</p>
          <div class="emoji-preview" :style="{ backgroundColor: '#' + selectedColor }">
            {{ selectedEmoji }}
          </div>
          <div class="emoji-input-row">
            <input
              v-model="selectedEmoji"
              type="text"
              :placeholder="t('step2_placeholder')"
              class="emoji-input"
              maxlength="4"
            />
            <button @click="randomEmoji" class="btn-random">{{ t('random_btn') }}</button>
          </div>
          <div class="step-buttons">
            <button @click="prevStep" class="btn-back">{{ t('back') }}</button>
            <button @click="nextStep" class="btn-next">{{ t('next') }}</button>
          </div>
        </div>

        <!-- Step 3: Color -->
        <div v-if="currentStep === 3" class="step">
          <h2>3. {{ t('step3_title') }}</h2>
          <p>{{ t('step3_desc') }}</p>
          <div class="color-preview" :style="{ backgroundColor: '#' + selectedColor }">
            <span :style="{ color: textColor }">{{ selectedEmoji }} {{ listName }}</span>
          </div>
          <div ref="colorPickerEl" class="color-picker"></div>
          <div class="step-buttons">
            <button @click="prevStep" class="btn-back">{{ t('back') }}</button>
            <button
              @click="createList"
              :disabled="isCreating"
              class="btn-create"
              :style="{ backgroundColor: '#' + selectedColor, color: textColor }"
            >
              {{ isCreating ? t('creating') : t('create_btn') }}
            </button>
          </div>
        </div>
      </div>
    </main>

    <!-- Footer -->
    <footer class="footer">
      <p>{{ t('footer') }}</p>
    </footer>
  </div>
</template>

<style scoped>
/* Landing page styles */
.landing {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  position: relative;
}

/* Language toggle */
.lang-toggle {
  position: absolute;
  top: 1rem;
  right: 1rem;
  padding: 0.5rem 1rem;
  background: rgba(255, 255, 255, 0.2);
  border: 1px solid rgba(255, 255, 255, 0.3);
  border-radius: 4px;
  color: white;
  cursor: pointer;
  font-weight: 600;
  z-index: 10;
}

.lang-toggle:hover {
  background: rgba(255, 255, 255, 0.3);
}

/* Hero section */
.hero {
  background: linear-gradient(135deg, #333 0%, #1a1a1a 100%);
  color: white;
  padding: 3rem 1rem 4rem;
  text-align: center;
}

.title {
  font-size: 4rem;
  letter-spacing: 4px;
  margin-bottom: 0.5rem;
}

.subtitle {
  font-size: 1.5rem;
  opacity: 0.9;
  font-weight: 300;
  margin-bottom: 0.25rem;
}

.tagline {
  font-size: 1rem;
  opacity: 0.7;
  margin-bottom: 2rem;
}

/* Phone Mockup */
.phone-mockup {
  position: relative;
  width: 180px;
  height: 360px;
  margin: 0 auto 2.5rem;
}

.phone-frame {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  z-index: 2;
  pointer-events: none;
}

.phone-screen {
  position: absolute;
  top: 12px;
  left: 12px;
  width: calc(100% - 24px);
  height: calc(100% - 24px);
  border-radius: 16px;
  object-fit: cover;
  z-index: 1;
}

/* Feature Cards */
.features {
  display: flex;
  justify-content: center;
  gap: 1rem;
  flex-wrap: wrap;
  max-width: 800px;
  margin: 0 auto;
}

.feature-card {
  background: rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  padding: 1.25rem 1rem;
  width: 200px;
  text-align: center;
  transition: transform 0.2s, background 0.2s;
}

.feature-card:hover {
  background: rgba(255, 255, 255, 0.15);
  transform: translateY(-2px);
}

.feature-icon {
  font-size: 2rem;
  display: block;
  margin-bottom: 0.5rem;
}

.feature-card h3 {
  font-size: 1rem;
  font-weight: 600;
  margin-bottom: 0.25rem;
}

.feature-card p {
  font-size: 0.8rem;
  opacity: 0.8;
  line-height: 1.4;
}

/* Wizard */
.wizard {
  flex: 1;
  display: flex;
  justify-content: center;
  align-items: flex-start;
  padding: 2rem 1rem;
  background: #f5f5f5;
}

.wizard-container {
  background: white;
  border-radius: 16px;
  padding: 2rem;
  max-width: 400px;
  width: 100%;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);
}

/* Progress indicator */
.progress {
  display: flex;
  justify-content: center;
  gap: 2rem;
  margin-bottom: 2rem;
}

.progress span {
  width: 2rem;
  height: 2rem;
  border-radius: 50%;
  background: #ddd;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: bold;
  transition: all 0.3s;
}

.progress span.active {
  background: #42b883;
  color: white;
}

/* Steps */
.step {
  text-align: center;
}

.step h2 {
  margin-bottom: 0.5rem;
  color: #333;
}

.step p {
  color: #666;
  margin-bottom: 1.5rem;
}

/* Name input */
.name-input {
  width: 100%;
  padding: 1rem;
  font-size: 1.25rem;
  border: 2px solid #ddd;
  border-radius: 8px;
  text-align: center;
  margin-bottom: 1.5rem;
}

.name-input:focus {
  outline: none;
  border-color: #42b883;
}

/* Emoji preview */
.emoji-preview {
  width: 100px;
  height: 100px;
  margin: 0 auto 1.5rem;
  border-radius: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 3rem;
  transition: background-color 0.3s;
}

/* Emoji input row */
.emoji-input-row {
  display: flex;
  gap: 0.5rem;
  margin-bottom: 1.5rem;
}

.emoji-input {
  flex: 1;
  padding: 1rem;
  font-size: 1.5rem;
  border: 2px solid #ddd;
  border-radius: 8px;
  text-align: center;
}

.emoji-input:focus {
  outline: none;
  border-color: #42b883;
}

.btn-random {
  padding: 1rem 1.5rem;
  font-size: 1rem;
  font-weight: 600;
  background: #f0f0f0;
  border: 2px solid #ddd;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-random:hover {
  background: #e8e8e8;
  border-color: #ccc;
}

/* Color preview */
.color-preview {
  padding: 1.5rem;
  border-radius: 12px;
  margin-bottom: 1.5rem;
  font-size: 1.25rem;
  font-weight: 600;
  transition: all 0.3s;
}

/* Color picker */
.color-picker {
  display: flex;
  justify-content: center;
  margin-bottom: 1.5rem;
}

/* Buttons */
.step-buttons {
  display: flex;
  gap: 1rem;
}

.btn-next,
.btn-back,
.btn-create {
  flex: 1;
  padding: 1rem;
  font-size: 1rem;
  font-weight: 600;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-next {
  background: #42b883;
  color: white;
}

.btn-next:hover {
  background: #3aa876;
}

.btn-back {
  background: #ddd;
  color: #333;
}

.btn-back:hover {
  background: #ccc;
}

.btn-create {
  box-shadow: 0 4px 15px rgba(0, 0, 0, 0.2);
}

.btn-create:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.25);
}

.btn-create:disabled {
  opacity: 0.7;
  cursor: not-allowed;
  transform: none;
}

/* Error */
.error {
  background: #fee;
  color: #c00;
  padding: 0.75rem;
  border-radius: 8px;
  margin-bottom: 1rem;
}

/* Footer */
.footer {
  background: #333;
  color: white;
  text-align: center;
  padding: 1rem;
  font-size: 0.875rem;
  opacity: 0.9;
}

/* Mobile responsive */
@media (max-width: 640px) {
  .title {
    font-size: 3rem;
    letter-spacing: 2px;
  }

  .subtitle {
    font-size: 1.25rem;
  }

  .phone-mockup {
    width: 150px;
    height: 300px;
    margin-bottom: 2rem;
  }

  .phone-screen {
    top: 10px;
    left: 10px;
    width: calc(100% - 20px);
    height: calc(100% - 20px);
  }

  .features {
    flex-direction: column;
    align-items: center;
    gap: 0.75rem;
  }

  .feature-card {
    width: 100%;
    max-width: 280px;
    padding: 1rem;
  }
}

@media (max-width: 480px) {
  .wizard-container {
    padding: 1.5rem;
  }

  .emoji-input-row {
    flex-direction: column;
  }

  .btn-random {
    width: 100%;
  }
}
</style>
