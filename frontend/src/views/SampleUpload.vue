<template>
  <div class="max-w-4xl mx-auto px-4 py-8">
    <div v-if="loading" class="text-center">
      <p>Loading current pack...</p>
    </div>
    
    <div v-else-if="error" class="text-red-600">
      {{ error }}
    </div>
    
    <div v-else class="space-y-8">
      <!-- Pack Info -->
      <div class="bg-white shadow rounded-lg p-6">
        <h1 class="text-3xl font-bold mb-4">{{ currentPack?.title }}</h1>
        <p class="text-gray-600 mb-6">{{ currentPack?.description }}</p>
        
        <!-- Upload Window Info -->
        <div class="p-4 rounded-lg" :class="timeWindowClass">
          <h2 class="font-semibold mb-2">Upload Window Status</h2>
          <p>{{ formatDateRange(currentPack?.uploadStart, currentPack?.uploadEnd) }}</p>
          <p class="mt-2">{{ timeWindowMessage }}</p>
        </div>
      </div>
      
      <!-- Upload Component -->
      <div v-if="isUploadAllowed">
        <SampleUpload
          :packId="currentPack?.ID"
          :uploadStart="currentPack?.uploadStart"
          :uploadEnd="currentPack?.uploadEnd"
          :samples="currentPack?.samples || []"
          @upload-complete="refreshPack"
        />
      </div>
      <div v-else class="bg-amber-50 text-amber-700 p-4 rounded-lg">
        <p class="font-medium">Upload window is currently closed</p>
        <p class="text-sm mt-1">Please check back during the next upload window</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { api } from '@/api'
import SampleUpload from '@/components/SampleUpload.vue'

const auth = useAuthStore()
const currentPack = ref<any>(null)
const loading = ref(true)
const error = ref<string | null>(null)

// Compute window status
const isUploadAllowed = computed(() => {
  if (!currentPack.value) return false
  if (__DEV_BYPASS_TIME_WINDOWS__) return true
  
  const now = new Date()
  const start = new Date(currentPack.value.uploadStart)
  const end = new Date(currentPack.value.uploadEnd)
  return now >= start && now <= end
})

const timeWindowClass = computed(() => ({
  'bg-green-50 text-green-700': isUploadAllowed.value,
  'bg-red-50 text-red-700': !isUploadAllowed.value
}))

const timeWindowMessage = computed(() => {
  if (__DEV_BYPASS_TIME_WINDOWS__) {
    return 'Time windows bypassed in development mode'
  }

  if (isUploadAllowed.value) {
    const end = new Date(currentPack.value.uploadEnd)
    const now = new Date()
    const hours = Math.round((end.getTime() - now.getTime()) / (1000 * 60 * 60))
    return `Upload window closes in approximately ${hours} hours`
  } else {
    const start = new Date(currentPack.value.uploadStart)
    const now = new Date()
    if (now < start) {
      const days = Math.round((start.getTime() - now.getTime()) / (1000 * 60 * 60 * 24))
      return `Upload window opens in ${days} days`
    } else {
      return 'Upload window has closed'
    }
  }
})

// Format date helpers
const formatDate = (date: string) => {
  return new Date(date).toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const formatDateRange = (start: string, end: string) => {
  if (!start || !end) return ''
  return `${formatDate(start)} to ${formatDate(end)}`
}

const refreshPack = async () => {
  try {
    const response = await api.packs.get(currentPack.value.ID)
    currentPack.value = response.data
    console.log('Fetched pack:', currentPack.value)
  } catch (e) {
    console.error('Failed to refresh pack:', e)
  }
}

// Initialize
onMounted(async () => {
  try {
    await auth.init()
    console.log('Auth initialized with user:', auth.user?.ID)
    if (!auth.user) {
      throw new Error('Not authenticated')
    }
    
    const response = await api.packs.get(1) // TODO: Get current pack ID from API
    currentPack.value = response.data
    console.log('Fetched pack:', currentPack.value)
  } catch (e) {
    console.error('Failed to initialize:', e)
    error.value = 'Failed to load current pack'
  } finally {
    loading.value = false
  }
})
</script> 