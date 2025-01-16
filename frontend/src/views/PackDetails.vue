<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { usePackStore } from '@/stores/index'
import * as api from '@/api'
import type { Sample, Submission } from '@/types'

const route = useRoute()
const packStore = usePackStore()
const packId = parseInt(route.params.id as string)

// State
const submissions = ref<Submission[]>([])
const currentSample = ref<Sample | null>(null)
const audioPlayer = ref<HTMLAudioElement | null>(null)
const isPlaying = ref(false)
const error = ref('')

// Upload state
const uploadFile = ref<File | null>(null)
const uploadError = ref('')
const isUploading = ref(false)

// Submission state
const submissionTitle = ref('')
const submissionDescription = ref('')
const submissionFile = ref<File | null>(null)
const submissionError = ref('')
const isSubmitting = ref(false)

onMounted(async () => {
  try {
    if (isNaN(packId)) {
      error.value = 'Invalid pack ID'
      return
    }

    console.log('Fetching pack:', packId)
    const { data } = await api.packs.get(packId)
    packStore.currentPack = data
    
    // Fetch submissions
    const submissionsResponse = await api.submissions.list(packId)
    submissions.value = submissionsResponse.data
  } catch (e: any) {
    console.error('Failed to fetch pack details:', e)
    error.value = e.response?.data?.error || 'Failed to load pack details'
  }
})

// Sample playback
const playSample = (sample: Sample) => {
  if (!audioPlayer.value) return
  
  currentSample.value = sample
  audioPlayer.value.src = sample.fileUrl
  audioPlayer.value.play()
  isPlaying.value = true
}

const stopSample = () => {
  if (!audioPlayer.value) return
  
  audioPlayer.value.pause()
  audioPlayer.value.currentTime = 0
  isPlaying.value = false
}

// Sample upload
const handleUpload = async () => {
  if (!uploadFile.value) {
    uploadError.value = 'Please select a file'
    return
  }

  isUploading.value = true
  uploadError.value = ''

  try {
    await api.packs.uploadSample(uploadFile.value)
    uploadFile.value = null
    // Refresh pack details to show new sample
    const { data } = await api.packs.get(packId)
    packStore.currentPack = data
  } catch (err) {
    uploadError.value = 'Upload failed'
    console.error('Upload error:', err)
  } finally {
    isUploading.value = false
  }
}

// Submission creation
const handleSubmit = async () => {
  if (!submissionFile.value) {
    submissionError.value = 'Please select a file'
    return
  }

  isSubmitting.value = true
  submissionError.value = ''

  try {
    await api.submissions.create({
      title: submissionTitle.value,
      description: submissionDescription.value,
      samplePackId: packId,
      file: submissionFile.value
    })
    
    // Reset form
    submissionTitle.value = ''
    submissionDescription.value = ''
    submissionFile.value = null
    
    // Refresh submissions
    const { data } = await api.submissions.list(packId)
    submissions.value = data
  } catch (err) {
    submissionError.value = 'Submission failed'
    console.error('Submission error:', err)
  } finally {
    isSubmitting.value = false
  }
}

// Add type for file input event
const handleFileUpload = (e: Event) => {
  const target = e.target as HTMLInputElement
  if (target.files) {
    uploadFile.value = target.files[0]
  }
}

const handleSubmissionFile = (e: Event) => {
  const target = e.target as HTMLInputElement
  if (target.files) {
    submissionFile.value = target.files[0]
  }
}
</script>

<template>
  <div class="min-h-screen bg-gray-50 py-8">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
      <!-- Error message -->
      <div v-if="error" class="bg-red-50 text-red-600 p-4 rounded-lg mb-4">
        {{ error }}
      </div>

      <!-- Pack Details -->
      <div v-if="packStore.currentPack" class="bg-white shadow rounded-lg p-6 mb-8">
        <h1 class="text-3xl font-bold mb-4">{{ packStore.currentPack.title }}</h1>
        <p class="text-gray-600 mb-6">{{ packStore.currentPack.description }}</p>

        <!-- Time Windows -->
        <div class="grid grid-cols-2 gap-4 text-sm text-gray-600 mb-8">
          <div>
            <p class="font-semibold">Upload Window:</p>
            <p>{{ new Date(packStore.currentPack.uploadStart).toLocaleString() }}</p>
            <p>to</p>
            <p>{{ new Date(packStore.currentPack.uploadEnd).toLocaleString() }}</p>
          </div>
          <div>
            <p class="font-semibold">Submission Window:</p>
            <p>{{ new Date(packStore.currentPack.startDate).toLocaleString() }}</p>
            <p>to</p>
            <p>{{ new Date(packStore.currentPack.endDate).toLocaleString() }}</p>
          </div>
        </div>

        <!-- Sample List -->
        <div class="mb-8">
          <h2 class="text-2xl font-bold mb-4">Samples</h2>
          <div class="space-y-4">
            <div 
              v-for="sample in packStore.currentPack.samples" 
              :key="sample.id"
              class="flex items-center justify-between p-4 bg-gray-50 rounded-lg"
            >
              <div>
                <p class="font-medium">{{ sample.filename }}</p>
                <p class="text-sm text-gray-500">
                  Uploaded by {{ sample.user.email }}
                </p>
              </div>
              <div class="flex items-center space-x-4">
                <button
                  @click="isPlaying && currentSample?.id === sample.id ? stopSample() : playSample(sample)"
                  class="text-indigo-600 hover:text-indigo-800"
                >
                  {{ isPlaying && currentSample?.id === sample.id ? 'Stop' : 'Play' }}
                </button>
                <a
                  :href="sample.fileUrl"
                  download
                  class="text-indigo-600 hover:text-indigo-800"
                >
                  Download
                </a>
              </div>
            </div>
          </div>
        </div>

        <!-- Upload Form -->
        <div v-if="packStore.currentPack.isActive" class="mb-8">
          <h2 class="text-2xl font-bold mb-4">Upload Sample</h2>
          <form @submit.prevent="handleUpload" class="space-y-4">
            <div>
              <input
                type="file"
                accept=".wav,.mp3,.aiff,.flac"
                @change="handleFileUpload"
                class="block w-full text-sm text-gray-500 file:mr-4 file:py-2 file:px-4 file:rounded-full file:border-0 file:text-sm file:font-semibold file:bg-indigo-50 file:text-indigo-700 hover:file:bg-indigo-100"
              />
            </div>
            <div v-if="uploadError" class="text-red-500">{{ uploadError }}</div>
            <button
              type="submit"
              :disabled="isUploading || !uploadFile"
              class="bg-indigo-600 text-white px-4 py-2 rounded hover:bg-indigo-700 disabled:opacity-50"
            >
              {{ isUploading ? 'Uploading...' : 'Upload' }}
            </button>
          </form>
        </div>

        <!-- Submissions -->
        <div>
          <h2 class="text-2xl font-bold mb-4">Submissions</h2>
          
          <!-- Submission Form -->
          <form v-if="packStore.currentPack.isActive" @submit.prevent="handleSubmit" class="mb-8 space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700">Title</label>
              <input
                v-model="submissionTitle"
                type="text"
                required
                class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
              />
            </div>
            
            <div>
              <label class="block text-sm font-medium text-gray-700">Description</label>
              <textarea
                v-model="submissionDescription"
                rows="3"
                class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
              />
            </div>
            
            <div>
              <label class="block text-sm font-medium text-gray-700">File</label>
              <input
                type="file"
                accept=".wav,.mp3,.aiff,.flac"
                @change="handleSubmissionFile"
                class="mt-1 block w-full text-sm text-gray-500 file:mr-4 file:py-2 file:px-4 file:rounded-full file:border-0 file:text-sm file:font-semibold file:bg-indigo-50 file:text-indigo-700 hover:file:bg-indigo-100"
              />
            </div>
            
            <div v-if="submissionError" class="text-red-500">{{ submissionError }}</div>
            
            <button
              type="submit"
              :disabled="isSubmitting || !submissionFile"
              class="bg-indigo-600 text-white px-4 py-2 rounded hover:bg-indigo-700 disabled:opacity-50"
            >
              {{ isSubmitting ? 'Submitting...' : 'Submit' }}
            </button>
          </form>

          <!-- Submission List -->
          <div class="space-y-4">
            <div 
              v-for="submission in submissions" 
              :key="submission.id"
              class="p-4 bg-gray-50 rounded-lg"
            >
              <div class="flex justify-between items-start">
                <div>
                  <h3 class="font-medium">{{ submission.title }}</h3>
                  <p class="text-sm text-gray-500">by {{ submission.user.email }}</p>
                  <p class="mt-2">{{ submission.description }}</p>
                </div>
                <a
                  :href="submission.fileUrl"
                  download
                  class="text-indigo-600 hover:text-indigo-800"
                >
                  Download
                </a>
              </div>
            </div>
          </div>
        </div>
      </div>
      <div v-else-if="!error" class="flex justify-center py-8">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-600"></div>
      </div>
    </div>

    <!-- Audio Player -->
    <audio ref="audioPlayer" @ended="isPlaying = false" />
  </div>
</template> 