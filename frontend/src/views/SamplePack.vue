<template>
  <div class="container mx-auto px-4 py-8">
    <div v-if="loading" class="text-center">
      <p>Loading sample pack...</p>
    </div>
    
    <div v-else-if="error" class="text-red-600">
      {{ error }}
    </div>
    
    <div v-else class="bg-white shadow-lg rounded-lg p-6">
      <h1 class="text-3xl font-bold mb-4">{{ pack?.title }}</h1>
      <p class="text-gray-600 mb-6">{{ pack?.description }}</p>
      
      <!-- Date Information -->
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4 mb-8">
        <div>
          <h3 class="font-semibold text-gray-700">Upload Window:</h3>
          <p>{{ formatDate(pack?.uploadStart) }} to {{ formatDate(pack?.uploadEnd) }}</p>
        </div>
        <div>
          <h3 class="font-semibold text-gray-700">Submission Window:</h3>
          <p>{{ formatDate(pack?.startDate) }} to {{ formatDate(pack?.endDate) }}</p>
        </div>
      </div>
      
      <!-- Upload Section -->
      <div v-if="isUploadAllowed" class="mt-8">
        <SampleUpload 
          :pack-id="Number(getId(pack))"
          :upload-start="pack.uploadStart"
          :upload-end="pack.uploadEnd"
          :samples="samples"
          @upload-complete="refreshPack"
        />
      </div>
      
      <!-- Samples List -->
      <div class="mt-8">
        <h2 class="text-2xl font-bold mb-4">Samples ({{ sampleCount }})</h2>
        <div class="space-y-4">
          <div v-for="sample in samples" :key="sample.ID" class="border rounded-lg p-4">
            <div class="flex justify-between items-center">
              <div>
                <p class="font-medium">{{ sample.filename }}</p>
                <p class="text-sm text-gray-600">
                  Uploaded by {{ getUserDisplay(sample.user) }}
                  {{ sample.user?.ID === currentUser.value?.ID ? '(You)' : '' }}
                </p>
                <p class="text-sm text-gray-500">{{ formatFileSize(sample.fileSize) }}</p>
              </div>
            </div>
          </div>
          <div v-if="!sampleCount" class="text-center text-gray-600">
            No samples uploaded yet.
          </div>
        </div>
      </div>
      
      <!-- Download Pack Button -->
      <div class="mt-8">
        <button 
          @click="downloadPack"
          class="w-full bg-green-500 text-white px-6 py-3 rounded-lg font-semibold hover:bg-green-600 focus:outline-none focus:ring-2 focus:ring-green-500 focus:ring-offset-2"
          :disabled="downloading || !sampleCount"
        >
          {{ downloading ? 'Downloading...' : 'Download All Samples' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import * as api from '@/api'
import type { SamplePack, User, Sample } from '@/types'
import { getId } from '@/utils/id'
import SampleUpload from '@/components/SampleUpload.vue'

const route = useRoute()
const auth = useAuthStore()
const currentUser = computed(() => auth.user)
const loading = ref(true)
const downloading = ref(false)
const uploading = ref(false)
const error = ref<string | null>(null)
const uploadError = ref<string | null>(null)
const pack = ref<SamplePack | null>(null)
const selectedFile = ref<File | null>(null)
const fileInput = ref<HTMLInputElement | null>(null)

const authToken = computed(() => localStorage.getItem('access_token'))
const isUploadAllowed = computed(() => {
  if (!pack.value) return false
  const now = new Date()
  const uploadStart = new Date(pack.value.uploadStart)
  const uploadEnd = new Date(pack.value.uploadEnd)
  return now >= uploadStart && now <= uploadEnd
})

const formatDate = (dateString?: string) => {
  if (!dateString) return ''
  return new Date(dateString).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const handleFileSelect = (event: Event) => {
  const input = event.target as HTMLInputElement
  if (input.files && input.files.length > 0) {
    selectedFile.value = input.files[0]
  }
}

const uploadSample = async () => {
  if (!selectedFile.value || !pack.value) return
  
  try {
    uploading.value = true
    uploadError.value = null
    const packId = getId(pack.value)
    if (!packId) {
      throw new Error('Invalid pack ID')
    }

    // Create FormData and append file
    const formData = new FormData()
    formData.append('file', selectedFile.value)
    
    await api.packs.uploadSample(Number(packId), formData)
    
    // Refresh pack data to show new sample
    const { data } = await api.packs.get(Number(packId))
    pack.value = data
    
    // Reset file input
    selectedFile.value = null
    if (fileInput.value) {
      fileInput.value.value = ''
    }
  } catch (e: any) {
    console.error('Upload error:', e)
    uploadError.value = e.response?.data?.error || 'Failed to upload sample'
  } finally {
    uploading.value = false
  }
}

const downloadPack = async () => {
  const packId = getId(pack.value)
  if (!packId) return
  
  try {
    downloading.value = true
    console.log('Starting download for pack:', packId);
    const { data: blob } = await api.packs.downloadPack(Number(packId))
    console.log('Received blob:', blob);
    
    // Create download link
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `${pack.value?.title}.zip`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)
  } catch (e: any) {
    console.error('Download error:', e);
    error.value = e.response?.data?.error || 'Failed to download pack'
  } finally {
    downloading.value = false
  }
}

const getUserDisplay = (user: User) => {
  return user.email
}

const formatFileSize = (bytes: number) => {
  if (!bytes) return '0 Bytes'
  const k = 1024
  const sizes = ['Bytes', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const refreshPack = async () => {
  const packId = getId(pack.value)
  if (!packId) return
  
  try {
    console.log('Refreshing pack data...')
    const { data } = await api.packs.get(Number(packId))
    console.log('New pack data:', data)
    
    // Initialize samples array if not present
    if (!data.samples) {
      data.samples = []
    }
    
    // Add auth token to sample URLs
    data.samples = data.samples.map((sample: any) => {
      const sampleId = getId(sample)
      if (sampleId) {
        sample.fileUrl = `/api/samples/download/${sampleId}?token=${authToken.value}`
      }
      return sample
    })
    
    // Update the pack data
    pack.value = { ...data }
    console.log('Updated pack with samples:', pack.value.samples.length)
  } catch (e: any) {
    console.error('Failed to refresh pack:', e)
    error.value = e.response?.data?.message || 'Failed to refresh pack data'
  }
}

// Add reactive refs for samples
const samples = computed(() => pack.value?.samples || [])
const sampleCount = computed(() => samples.value.length)

onMounted(async () => {
  const packId = parseInt(route.params.id as string)
  try {
    loading.value = true
    const { data } = await api.packs.get(packId)
    
    // Initialize samples array if not present
    if (!data.samples) {
      data.samples = []
    }
    
    // Add auth token to sample URLs
    data.samples = data.samples.map((sample: any) => {
      const sampleId = getId(sample)
      if (sampleId) {
        sample.fileUrl = `/api/samples/download/${sampleId}?token=${authToken.value}`
      }
      return sample
    })
    
    pack.value = { ...data }
    console.log('Initial pack load with samples:', pack.value.samples.length)
  } catch (e: any) {
    console.error('Failed to load pack:', e)
    error.value = e.response?.data?.message || 'Failed to load sample pack'
  } finally {
    loading.value = false
  }
})
</script> 