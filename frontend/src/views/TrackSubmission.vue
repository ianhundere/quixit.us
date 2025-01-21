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
        
        <!-- Submission Window Info -->
        <div class="p-4 rounded-lg" :class="timeWindowClass">
          <h2 class="font-semibold mb-2">Submission Window Status</h2>
          <p>{{ formatDateRange(currentPack?.startDate, currentPack?.endDate) }}</p>
          <p class="mt-2">{{ timeWindowMessage }}</p>
        </div>
      </div>
      
      <!-- Download Pack -->
      <div class="bg-white shadow rounded-lg p-6">
        <h2 class="text-xl font-bold mb-4">Sample Pack</h2>
        <p class="text-gray-600 mb-4">
          Download the current sample pack to create your track. Your track must use samples from this pack.
        </p>
        <button 
          @click="downloadPack"
          class="w-full bg-blue-500 text-white px-6 py-3 rounded-lg font-semibold hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
          :disabled="downloading || !currentPack?.samples?.length"
        >
          {{ downloading ? 'Downloading...' : 'Download Sample Pack' }}
        </button>
      </div>
      
      <!-- Submit Track -->
      <div v-if="isSubmissionAllowed" class="bg-white shadow rounded-lg p-6">
        <h2 class="text-xl font-bold mb-4">Submit Your Track</h2>
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700">Track Title</label>
            <input 
              v-model="trackTitle" 
              type="text"
              class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
              placeholder="Enter your track title"
            />
          </div>
          
          <div>
            <label class="block text-sm font-medium text-gray-700">Track File</label>
            <input
              type="file"
              ref="fileInput"
              @change="handleFileSelect"
              accept=".wav,.mp3,.aiff,.flac"
              class="mt-1 block w-full text-sm text-gray-500
                file:mr-4 file:py-2 file:px-4
                file:rounded-full file:border-0
                file:text-sm file:font-semibold
                file:bg-blue-50 file:text-blue-700
                hover:file:bg-blue-100"
            />
          </div>
          
          <div v-if="selectedFile" class="p-4 bg-gray-50 rounded">
            <div class="flex justify-between items-center">
              <span class="font-medium">{{ selectedFile.name }}</span>
              <span class="text-sm text-gray-600">{{ formatFileSize(selectedFile.size) }}</span>
            </div>
          </div>
          
          <button
            @click="submitTrack"
            :disabled="!canSubmit || submitting"
            class="w-full bg-green-500 text-white px-6 py-3 rounded-lg font-semibold 
                   hover:bg-green-600 focus:outline-none focus:ring-2 focus:ring-green-500 
                   focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {{ submitting ? 'Submitting...' : 'Submit Track' }}
          </button>
          
          <p v-if="submitError" class="mt-2 text-red-600">{{ submitError }}</p>
          <p v-if="submitSuccess" class="mt-2 text-green-600">{{ submitSuccess }}</p>
        </div>
      </div>
      <div v-else class="bg-amber-50 text-amber-700 p-4 rounded-lg">
        <p class="font-medium">Submission window is currently closed</p>
        <p class="text-sm mt-1">Please check back during the next submission window</p>
      </div>
      
      <!-- Previous Submissions -->
      <div class="bg-white shadow rounded-lg p-6">
        <h2 class="text-xl font-bold mb-4">Submissions</h2>
        <div class="space-y-4">
          <div v-for="submission in submissions" :key="submission.ID" 
               class="p-4 border rounded-lg hover:bg-gray-50">
            <div class="flex justify-between items-center">
              <div>
                <h3 class="font-medium">{{ submission.title }}</h3>
                <p class="text-sm text-gray-600">
                  Submitted on {{ formatDate(submission.createdAt) }}
                </p>
              </div>
              <button v-if="submission.fileURL"
                      @click="downloadSubmission(submission)"
                      class="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600">
                Download
              </button>
            </div>
          </div>
          <div v-if="submissions.length === 0" class="text-center py-4 text-gray-600">
            No submissions yet.
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { api } from '@/api'

const auth = useAuthStore()
const currentPack = ref<any>(null)
const submissions = ref<any[]>([])
const loading = ref(true)
const error = ref<string | null>(null)
const downloading = ref(false)
const submitting = ref(false)
const submitError = ref<string | null>(null)
const submitSuccess = ref<string | null>(null)

const trackTitle = ref('')
const selectedFile = ref<File | null>(null)
const fileInput = ref<HTMLInputElement | null>(null)

// Compute window status
const isSubmissionAllowed = computed(() => {
  if (!currentPack.value) return false
  if ((globalThis as any).__DEV_BYPASS_TIME_WINDOWS__) return true
  
  const now = new Date()
  const start = new Date(currentPack.value.startDate)
  const end = new Date(currentPack.value.endDate)
  return now >= start && now <= end
})

const timeWindowClass = computed(() => ({
  'bg-green-50 text-green-700': isSubmissionAllowed.value,
  'bg-red-50 text-red-700': !isSubmissionAllowed.value
}))

const timeWindowMessage = computed(() => {
  if ((globalThis as any).__DEV_BYPASS_TIME_WINDOWS__) {
    return 'Time windows bypassed in development mode'
  }

  if (isSubmissionAllowed.value) {
    const end = new Date(currentPack.value.endDate)
    const now = new Date()
    const hours = Math.round((end.getTime() - now.getTime()) / (1000 * 60 * 60))
    return `Submission window closes in approximately ${hours} hours`
  } else {
    const start = new Date(currentPack.value.startDate)
    const now = new Date()
    if (now < start) {
      const days = Math.round((start.getTime() - now.getTime()) / (1000 * 60 * 60 * 24))
      return `Submission window opens in ${days} days`
    } else {
      return 'Submission window has closed'
    }
  }
})

const canSubmit = computed(() => {
  return trackTitle.value.trim() && selectedFile.value
})

// Format helpers
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

const formatFileSize = (bytes: number) => {
  if (!bytes) return '0 Bytes'
  const k = 1024
  const sizes = ['Bytes', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

// Actions
const handleFileSelect = (event: Event) => {
  const input = event.target as HTMLInputElement
  if (input.files && input.files.length > 0) {
    selectedFile.value = input.files[0]
  }
}

const downloadPack = async () => {
  if (!currentPack.value?.ID) return
  
  try {
    downloading.value = true
    console.log('Starting download for pack:', currentPack.value.ID)
    const { data: blob } = await api.packs.download(currentPack.value.ID)
    console.log('Received blob:', blob)
    
    // Create download link
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `${currentPack.value.title}_samples.zip`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)
  } catch (e: any) {
    console.error('Download error:', e)
    error.value = e.response?.data?.error || 'Failed to download pack'
  } finally {
    downloading.value = false
  }
}

const submitTrack = async () => {
  if (!selectedFile.value || !trackTitle.value.trim()) return
  
  try {
    submitting.value = true
    submitError.value = null
    submitSuccess.value = null
    
    const formData = new FormData()
    formData.append('file', selectedFile.value)
    formData.append('title', trackTitle.value.trim())
    formData.append('sample_pack_id', currentPack.value.ID.toString())
    
    await api.post('/submissions', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    })
    
    // Reset form
    trackTitle.value = ''
    selectedFile.value = null
    if (fileInput.value) {
      fileInput.value.value = ''
    }
    
    // Show success message
    submitSuccess.value = 'Track submitted successfully!'
    
    // Refresh submissions list
    await fetchSubmissions()
  } catch (e: any) {
    console.error('Submit error:', e)
    submitError.value = e.response?.data?.error || 'Failed to submit track'
  } finally {
    submitting.value = false
  }
}

const downloadSubmission = async (submission: { fileURL?: string }) => {
  if (!submission.fileURL) return
  window.open(submission.fileURL, '_blank')
}

const fetchSubmissions = async () => {
  if (!currentPack.value?.ID) return
  
  try {
    const response = await api.submissions.list(currentPack.value.ID)
    submissions.value = response.data
  } catch (e) {
    console.error('Failed to fetch submissions:', e)
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
    
    await fetchSubmissions()
  } catch (e) {
    console.error('Failed to initialize:', e)
    error.value = 'Failed to load current pack'
  } finally {
    loading.value = false
  }
})
</script> 