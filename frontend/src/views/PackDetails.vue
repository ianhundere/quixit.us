<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute } from 'vue-router'
import { usePackStore } from '@/stores/index'
import * as api from '@/api'
import type { Sample, Submission } from '@/types'
import AudioPlayer from '@/components/AudioPlayer.vue'
import { downloadFile } from '@/api'
import FileInput from '@/components/FileInput.vue'

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
const uploadSuccess = ref(false)

// Submission state
const submissionTitle = ref('')
const submissionDescription = ref('')
const submissionFile = ref<File | null>(null)
const submissionError = ref('')
const isSubmitting = ref(false)
const submissionSuccess = ref(false)

const allowedTypes = ['.wav', '.mp3', '.aiff', '.flac']
const acceptString = allowedTypes.join(',')

// Add a computed property for the auth token
const authToken = computed(() => localStorage.getItem('access_token'))

const downloadingFiles = ref(new Set<string>())

// Add loading state
const isLoading = ref(true)

const formatDate = (date: string) => {
  return new Date(date).toLocaleString('en-US', {
    weekday: 'short',
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const validateFile = (file: File) => {
  const extension = file.name.substring(file.name.lastIndexOf('.')).toLowerCase()
  
  if (!allowedTypes.includes(extension)) {
    return `Invalid file type. Allowed types: ${allowedTypes.join(', ')}`
  }
  
  if (file.size > 50 * 1024 * 1024) {
    return 'File size must be less than 50MB'
  }
  
  return null
}

onMounted(async () => {
  try {
    isLoading.value = true
    if (isNaN(packId)) {
      error.value = 'Invalid pack ID'
      return
    }

    console.log('Fetching pack:', packId)
    const { data } = await api.packs.get(packId)
    packStore.currentPack = data

    // Generate file URLs for samples with auth token
    if (data.samples) {
      data.samples.forEach(sample => {
        sample.fileUrl = `/api/samples/download/${sample.ID}?token=${authToken.value}`
      })
    }
    
    // Fetch submissions and add auth token to URLs
    const submissionsResponse = await api.submissions.list(packId)
    submissions.value = submissionsResponse.data.map(submission => ({
      ...submission,
      fileUrl: `/api/submissions/${submission.ID}/download?token=${authToken.value}`
    }))
  } catch (e: any) {
    console.error('Failed to fetch pack details:', e)
    error.value = e.response?.data?.error || 'Failed to load pack details'
  } finally {
    isLoading.value = false
  }
})

// Sample playback
const playSample = (sample: Sample) => {
  currentSample.value = sample
  isPlaying.value = true
}

const stopSample = () => {
  isPlaying.value = false
  currentSample.value = null
}

// Sample upload
const handleUpload = async () => {
  if (!uploadFile.value) {
    uploadError.value = 'Please select a file'
    return
  }

  if (!packId) {
    uploadError.value = 'Invalid pack ID'
    return
  }

  isUploading.value = true
  uploadError.value = ''
  uploadSuccess.value = false

  try {
    console.log('Uploading file:', uploadFile.value.name, 'to pack:', packId)
    await api.packs.uploadSample(packId, uploadFile.value)
    uploadFile.value = null
    uploadSuccess.value = true
    
    // Refresh pack details to show new sample
    const { data } = await api.packs.get(packId)
    packStore.currentPack = data
  } catch (err: any) {
    console.error('Upload error:', err)
    uploadError.value = err.response?.data?.error || 'Upload failed'
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

  if (!submissionTitle.value.trim()) {
    submissionError.value = 'Title is required'
    return
  }

  isSubmitting.value = true
  submissionError.value = ''

  try {
    console.log('Creating submission:', {
      title: submissionTitle.value,
      description: submissionDescription.value,
      samplePackId: packId,
      file: submissionFile.value
    })

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
    submissions.value = data.map(submission => ({
      ...submission,
      fileUrl: `/api/submissions/${submission.ID}/download?token=${authToken.value}`
    }))
    submissionSuccess.value = true
    setTimeout(() => {
      submissionSuccess.value = false
    }, 3000)
  } catch (err: any) {
    console.error('Submission error:', err)
    submissionError.value = err.response?.data?.error || 'Submission failed'
  } finally {
    isSubmitting.value = false
  }
}

const handleFileSelected = (file: File | null) => {
  if (!file) {
    uploadError.value = ''
    uploadFile.value = null
    return
  }

  const error = validateFile(file)
  if (error) {
    uploadError.value = error
    uploadFile.value = null
    return
  }
  
  uploadFile.value = file
  uploadError.value = ''
}

const handleSubmissionFileSelected = (file: File | null) => {
  if (!file) {
    submissionError.value = ''
    submissionFile.value = null
    return
  }

  const error = validateFile(file)
  if (error) {
    submissionError.value = error
    submissionFile.value = null
    return
  }
  
  submissionFile.value = file
  submissionError.value = ''
}

// Add download handler
const handleDownload = async (url: string, filename: string) => {
  if (downloadingFiles.value.has(url)) return
  
  try {
    downloadingFiles.value.add(url)
    const blob = await downloadFile(url)
    const downloadUrl = window.URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = downloadUrl
    a.download = filename
    document.body.appendChild(a)
    a.click()
    window.URL.revokeObjectURL(downloadUrl)
    document.body.removeChild(a)
  } catch (err) {
    console.error('Download failed:', err)
  } finally {
    downloadingFiles.value.delete(url)
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
      <div v-if="packStore.currentPack?.ID" class="bg-white shadow rounded-lg p-6 mb-8">
        <h1 class="text-3xl font-bold mb-4">{{ packStore.currentPack.title }}</h1>
        <p class="text-gray-600 mb-6">{{ packStore.currentPack.description }}</p>

        <!-- Time Windows -->
        <div class="grid grid-cols-2 gap-4 text-sm text-gray-600 mb-8">
          <div>
            <p class="font-semibold">Upload Window:</p>
            <p>{{ formatDate(packStore.currentPack.uploadStart) }}</p>
            <p>to</p>
            <p>{{ formatDate(packStore.currentPack.uploadEnd) }}</p>
          </div>
          <div>
            <p class="font-semibold">Submission Window:</p>
            <p>{{ formatDate(packStore.currentPack.startDate) }}</p>
            <p>to</p>
            <p>{{ formatDate(packStore.currentPack.endDate) }}</p>
          </div>
        </div>

        <!-- Sample List -->
        <div class="mb-8">
          <h2 class="text-2xl font-bold mb-4">Samples</h2>
          <div class="space-y-4">
            <div v-if="!packStore.currentPack.samples?.length" class="text-gray-500 text-center py-4">
              No samples uploaded yet.
            </div>
            <div 
              v-for="sample in packStore.currentPack.samples" 
              :key="sample.ID"
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
                  @click="isPlaying && currentSample?.ID === sample.ID ? stopSample() : playSample(sample)"
                  class="text-indigo-600 hover:text-indigo-800"
                >
                  {{ isPlaying && currentSample?.ID === sample.ID ? 'Stop' : 'Play' }}
                </button>
                <a
                  href="#"
                  @click.prevent="handleDownload(sample.fileUrl, sample.filename)"
                  class="text-indigo-600 hover:text-indigo-800"
                  :class="{ 'opacity-50 cursor-wait': downloadingFiles.has(sample.fileUrl) }"
                >
                  {{ downloadingFiles.has(sample.fileUrl) ? 'Downloading...' : 'Download' }}
                </a>
              </div>
            </div>
          </div>
        </div>

        <!-- Upload Form -->
        <div v-if="packStore.currentPack.isActive" class="mb-8">
          <h2 class="text-2xl font-bold mb-4">Upload Sample</h2>
          <form @submit.prevent="handleUpload" class="space-y-4">
            <FileInput
              :accept="acceptString"
              label="Choose a sample file"
              :error="uploadError"
              :disabled="isUploading"
              :selectedFile="uploadFile"
              @file-selected="handleFileSelected"
            />
            <div v-if="uploadSuccess" class="text-green-500">Upload successful!</div>
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
            
            <FileInput
              :accept="acceptString"
              label="Choose a submission file"
              :error="submissionError"
              :disabled="isSubmitting"
              :selectedFile="submissionFile"
              @file-selected="handleSubmissionFileSelected"
            />
            
            <div v-if="submissionSuccess" class="text-green-500">
              Submission successful!
            </div>
            
            <button
              type="submit"
              :disabled="isSubmitting || !submissionFile || !submissionTitle.trim()"
              class="bg-indigo-600 text-white px-4 py-2 rounded hover:bg-indigo-700 disabled:opacity-50"
            >
              {{ isSubmitting ? 'Submitting...' : 'Submit' }}
            </button>
          </form>

          <!-- Submission List -->
          <div class="space-y-4">
            <div v-if="!submissions.length" class="text-gray-500 text-center py-4">
              No submissions yet.
            </div>
            <div 
              v-for="submission in submissions" 
              :key="submission.ID"
              class="p-4 bg-gray-50 rounded-lg"
            >
              <div class="flex justify-between items-start">
                <div>
                  <h3 class="font-medium">{{ submission.title }}</h3>
                  <p class="text-sm text-gray-500">by {{ submission.user.email }}</p>
                  <p class="mt-2">{{ submission.description }}</p>
                </div>
                <a
                  href="#"
                  @click.prevent="handleDownload(submission.fileUrl, `submission_${submission.ID}.wav`)"
                  class="text-indigo-600 hover:text-indigo-800"
                  :class="{ 'opacity-50 cursor-wait': downloadingFiles.has(submission.fileUrl) }"
                >
                  {{ downloadingFiles.has(submission.fileUrl) ? 'Downloading...' : 'Download' }}
                </a>
              </div>
            </div>
          </div>
        </div>
      </div>
      <div v-else-if="isLoading" class="flex justify-center py-8">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-600"></div>
      </div>
    </div>

    <!-- Audio Player -->
    <AudioPlayer 
      :sample="currentSample"
      :isPlaying="isPlaying"
      @playback-ended="isPlaying = false"
      @can-play="() => {}"
      @error="(message) => error = message"
    />
  </div>
</template> 