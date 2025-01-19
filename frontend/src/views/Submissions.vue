<template>
  <div class="max-w-4xl mx-auto px-4 py-8">
    <h1 class="text-3xl font-bold mb-6">{{ currentPack?.title }}</h1>
    <p class="text-gray-600 mb-8">{{ currentPack?.description }}</p>

    <!-- Phase Timeline -->
    <div class="mb-8 border rounded-lg overflow-hidden">
      <!-- Sample Upload Phase -->
      <div class="p-6" :class="{'bg-blue-50': isInUploadPhase, 'bg-gray-50': !isInUploadPhase}">
        <div class="flex items-center justify-between">
          <div>
            <h2 class="text-xl font-semibold mb-2">Phase 1: Sample Uploads</h2>
            <p class="text-sm text-gray-600">
              {{ formatDateRange(currentPack?.uploadStart, currentPack?.uploadEnd) }}
            </p>
          </div>
          <div class="text-right">
            <div v-if="isInUploadPhase" class="text-blue-600 font-medium">
              Currently Active
            </div>
            <div v-else-if="isPastUploadPhase" class="text-gray-600">
              Closed
            </div>
            <div v-else class="text-gray-600">
              Opens {{ formatDate(currentPack?.uploadStart) }}
            </div>
          </div>
        </div>
        
        <!-- Sample Upload Section -->
        <div v-if="isInUploadPhase || __DEV_BYPASS_TIME_WINDOWS__" class="mt-6">
          <SampleUpload
            :packId="currentPack?.ID"
            :uploadStart="currentPack?.uploadStart"
            :uploadEnd="currentPack?.uploadEnd"
            :samples="currentPack?.samples || []"
            @upload-complete="refreshPack"
          />
        </div>
      </div>

      <!-- Song Creation Phase -->
      <div class="p-6 border-t" :class="{'bg-blue-50': isInCreationPhase, 'bg-gray-50': !isInCreationPhase}">
        <div class="flex items-center justify-between">
          <div>
            <h2 class="text-xl font-semibold mb-2">Phase 2: Song Submissions</h2>
            <p class="text-sm text-gray-600">
              {{ formatDateRange(currentPack?.startDate, currentPack?.endDate) }}
            </p>
          </div>
          <div class="text-right">
            <div v-if="isInCreationPhase" class="text-blue-600 font-medium">
              Currently Active
            </div>
            <div v-else-if="isPastCreationPhase" class="text-gray-600">
              Closed
            </div>
            <div v-else class="text-gray-600">
              Opens {{ formatDate(currentPack?.startDate) }}
            </div>
          </div>
        </div>

        <!-- Song Submission Section -->
        <div v-if="isInCreationPhase || __DEV_BYPASS_TIME_WINDOWS__" class="mt-6">
          <div v-if="loading" class="text-center py-4">
            Loading submissions...
          </div>
          <div v-else-if="error" class="text-red-600 py-4">
            {{ error }}
          </div>
          <div v-else>
            <!-- Download Pack Button -->
            <div class="mb-6">
              <button 
                @click="downloadPack"
                class="w-full bg-blue-500 text-white px-6 py-3 rounded-lg font-semibold hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
                :disabled="downloading || !currentPack?.samples?.length"
              >
                {{ downloading ? 'Downloading...' : 'Download Sample Pack' }}
              </button>
            </div>

            <!-- Submissions List -->
            <div class="space-y-4">
              <div v-for="submission in submissions" :key="submission.ID" 
                   class="p-4 border rounded-lg hover:bg-gray-50">
                <div class="flex justify-between items-center">
                  <div>
                    <h3 class="font-medium">{{ submission.title }}</h3>
                    <p class="text-sm text-gray-600">
                      By {{ submission.user?.email }} on {{ formatDate(submission.createdAt) }}
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
const submissions = ref<any[]>([])
const loading = ref(true)
const error = ref<string | null>(null)
const downloading = ref(false)

// Compute current phase
const isInUploadPhase = computed(() => {
  if (!currentPack.value) return false
  if (__DEV_BYPASS_TIME_WINDOWS__) return true
  
  const now = new Date()
  const start = new Date(currentPack.value.uploadStart)
  const end = new Date(currentPack.value.uploadEnd)
  return now >= start && now <= end
})

const isInCreationPhase = computed(() => {
  if (!currentPack.value) return false
  if (__DEV_BYPASS_TIME_WINDOWS__) return true
  
  const now = new Date()
  const start = new Date(currentPack.value.startDate)
  const end = new Date(currentPack.value.endDate)
  return now >= start && now <= end
})

const isPastUploadPhase = computed(() => {
  if (!currentPack.value) return false
  if (__DEV_BYPASS_TIME_WINDOWS__) return false
  
  const now = new Date()
  const end = new Date(currentPack.value.uploadEnd)
  return now > end
})

const isPastCreationPhase = computed(() => {
  if (!currentPack.value) return false
  if (__DEV_BYPASS_TIME_WINDOWS__) return false
  
  const now = new Date()
  const end = new Date(currentPack.value.endDate)
  return now > end
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

const fetchCurrentPack = async () => {
  try {
    const response = await api.packs.get(1) // TODO: Get current pack ID from route
    currentPack.value = response.data
    console.log('Fetched pack:', currentPack.value)
  } catch (e) {
    console.error('Failed to fetch current pack:', e)
    error.value = 'Failed to load pack data'
  }
}

const fetchSubmissions = async () => {
  if (!currentPack.value?.ID) return
  
  try {
    const response = await api.get('/submissions', {
      params: { pack_id: currentPack.value.ID }
    })
    submissions.value = response.data
  } catch (e) {
    console.error('Failed to fetch submissions:', e)
    error.value = 'Failed to load submissions'
  }
}

const downloadSubmission = async (submission: any) => {
  if (!submission.fileURL) return
  window.open(submission.fileURL, '_blank')
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

// Initialize
onMounted(async () => {
  try {
    await auth.init()
    console.log('Auth initialized with user:', auth.user?.ID)
    if (!auth.user) {
      throw new Error('Not authenticated')
    }
    
    await fetchCurrentPack()
    await fetchSubmissions()
  } catch (e) {
    console.error('Failed to initialize:', e)
    error.value = 'Please log in to continue'
  } finally {
    loading.value = false
  }
})
</script> 