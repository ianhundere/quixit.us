<template>
  <div v-if="authInitialized" class="mt-8 p-6 border-2 border-dashed border-gray-300 rounded-lg">
    <h2 class="text-2xl font-bold mb-4">Upload Sample</h2>
    
    <!-- Upload Limits Info -->
    <div class="mb-4 p-4 bg-blue-50 text-blue-700 rounded">
      <p class="font-medium">Upload Limits:</p>
      <ul class="mt-1 text-sm list-disc list-inside">
        <li>Maximum {{ MAX_SAMPLES_PER_USER }} samples per user</li>
        <li>Maximum file size: {{ formatFileSize(MAX_FILE_SIZE) }}</li>
        <li>Accepted formats: .wav, .mp3, .aiff, .flac</li>
      </ul>
      <p class="mt-2 font-medium">
        You have {{ remainingUploads }} upload{{ remainingUploads === 1 ? '' : 's' }} remaining
        <span v-if="currentUploads.length > 0" class="text-sm">
          ({{ currentUploads.length }} uploaded)
        </span>
      </p>
    </div>

    <!-- Current Uploads -->
    <div v-if="currentUploads.length > 0" class="mb-4 p-4 bg-gray-50 rounded">
      <h3 class="font-medium mb-2">Your Current Uploads ({{ currentUploads.length }}/{{ MAX_SAMPLES_PER_USER }}):</h3>
      <ul class="space-y-2">
        <li v-for="upload in currentUploads" :key="upload.ID" class="flex justify-between items-center text-sm">
          <span>{{ upload.filename }}</span>
          <span class="text-gray-600">{{ formatFileSize(upload.fileSize) }}</span>
        </li>
      </ul>
      <p class="mt-2 text-sm text-gray-600">
        Total size: {{ formatFileSize(totalUploadSize) }}
      </p>
    </div>
    
    <div class="mb-4 p-4 rounded" :class="timeWindowClass">
      <p class="font-semibold">Upload window is {{ isUploadAllowed ? 'OPEN' : 'CLOSED' }}</p>
      <p class="text-sm mt-1">{{ timeWindowMessage }}</p>
    </div>
    
    <div v-if="isUploadAllowed && remainingUploads > 0" class="space-y-4">
      <input
        type="file"
        ref="fileInput"
        @change="handleFileSelect"
        accept=".wav,.mp3,.aiff,.flac"
        class="block w-full text-sm text-gray-500
          file:mr-4 file:py-2 file:px-4
          file:rounded-full file:border-0
          file:text-sm file:font-semibold
          file:bg-blue-50 file:text-blue-700
          hover:file:bg-blue-100"
      />
      <div v-if="selectedFile" class="p-4 bg-gray-50 rounded">
        <div class="flex justify-between items-center">
          <span class="font-medium">{{ selectedFile.name }}</span>
          <span class="text-sm text-gray-600">{{ formatFileSize(selectedFile.size) }}</span>
        </div>
      </div>
      <button
        @click="uploadFile"
        :disabled="!selectedFile || uploading"
        class="w-full bg-blue-500 text-white px-6 py-3 rounded-lg font-semibold 
               hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500 
               focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed"
      >
        {{ uploading ? 'Uploading...' : 'Upload Sample' }}
      </button>
    </div>
    <p v-else-if="remainingUploads <= 0" class="text-amber-600">
      You have reached the maximum number of uploads for this pack.
    </p>
    <p v-if="error" class="mt-2 text-red-600">{{ error }}</p>
    <p v-if="successMessage" class="mt-2 text-green-600">{{ successMessage }}</p>
  </div>
  <div v-else>
    Loading...
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { api } from '@/api'

const props = defineProps<{
  packId: number
  uploadStart: string
  uploadEnd: string
  currentSampleCount?: number
  samples?: Array<{
    ID: number
    filename: string
    fileSize: number
    user: { ID: number }
  }>
}>()

const emit = defineEmits<{
  (e: 'upload-complete'): void
}>()

const auth = useAuthStore()
const currentUser = computed(() => auth.user)

const MAX_FILE_SIZE = 25 * 1024 * 1024; // 25MB in bytes
const MAX_SAMPLES_PER_USER = 10;

const fileInput = ref<HTMLInputElement | null>(null)
const selectedFile = ref<File | null>(null)
const uploading = ref(false)
const error = ref<string | null>(null)
const successMessage = ref('')

// Add loading state
const authInitialized = ref(false);

// Update currentUploads computed to handle undefined user
const currentUploads = computed(() => {
  if (!props.samples || !currentUser.value) {
    console.log('No samples or user not initialized yet');
    return [];
  }
  
  console.log('Current user ID:', currentUser.value.ID);
  console.log('Total samples:', props.samples.length);
  
  const userUploads = props.samples.filter(s => {
    const sampleUserId = s.user?.ID;
    const currentUserId = currentUser.value?.ID;
    console.log(`Comparing sample user ID ${sampleUserId} with current user ID ${currentUserId}`);
    return sampleUserId === currentUserId;
  });
  
  console.log('User uploads found:', userUploads.length);
  console.log('Sample details:', userUploads.map(s => ({ id: s.ID, filename: s.filename, userId: s.user?.ID })));
  
  return userUploads;
})

const totalUploadSize = computed(() => {
  return currentUploads.value.reduce((total, sample) => total + (sample.fileSize || 0), 0)
})

const remainingUploads = computed(() => {
  const current = currentUploads.value.length
  console.log('Current uploads:', current, 'Max:', MAX_SAMPLES_PER_USER)
  return Math.max(0, MAX_SAMPLES_PER_USER - current)
})

const isUploadAllowed = computed(() => {
  // Always allow uploads in dev mode
  if (__DEV_BYPASS_TIME_WINDOWS__) {
    return true
  }
  
  const now = new Date()
  const start = new Date(props.uploadStart)
  const end = new Date(props.uploadEnd)
  return now >= start && now <= end
})

const timeWindowClass = computed(() => ({
  'bg-green-50 text-green-700': isUploadAllowed.value,
  'bg-red-50 text-red-700': !isUploadAllowed.value
}))

const timeWindowMessage = computed(() => {
  // Show bypass message in dev mode
  if (__DEV_BYPASS_TIME_WINDOWS__) {
    return 'Time windows bypassed in development mode'
  }

  if (isUploadAllowed.value) {
    const end = new Date(props.uploadEnd)
    const now = new Date()
    const hours = Math.round((end.getTime() - now.getTime()) / (1000 * 60 * 60))
    return `Upload window closes in approximately ${hours} hours`
  } else {
    const start = new Date(props.uploadStart)
    const now = new Date()
    if (now < start) {
      const days = Math.round((start.getTime() - now.getTime()) / (1000 * 60 * 60 * 24))
      return `Upload window opens in ${days} days`
    } else {
      return 'Upload window has closed'
    }
  }
})

const handleFileSelect = (event: Event) => {
  successMessage.value = ''
  const input = event.target as HTMLInputElement
  if (input.files && input.files.length > 0) {
    const file = input.files[0]
    
    // Check file size
    if (file.size > MAX_FILE_SIZE) {
      error.value = `File size exceeds limit of ${formatFileSize(MAX_FILE_SIZE)}`
      input.value = ''
      return
    }
    
    // Check remaining uploads
    if (remainingUploads.value <= 0) {
      error.value = 'You have reached the maximum number of uploads for this pack'
      input.value = ''
      return
    }
    
    selectedFile.value = file
    error.value = null
  }
}

const formatFileSize = (bytes: number) => {
  if (bytes === 0) return '0 Bytes'
  const k = 1024
  const sizes = ['Bytes', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const uploadFile = async () => {
  if (!selectedFile.value) return
  
  try {
    uploading.value = true
    error.value = null
    
    // Make the upload request with the file directly
    const response = await api.packs.uploadSample(props.packId, selectedFile.value)
    console.log('Upload response:', response)
    
    // Reset form
    selectedFile.value = null
    if (fileInput.value) {
      fileInput.value.value = ''
    }
    
    // Show success message
    const newSample = response.data
    if (newSample) {
      successMessage.value = `Successfully uploaded ${newSample.filename}`
    }
    
    // Notify parent to refresh pack data
    emit('upload-complete')
    
  } catch (e: any) {
    console.error('Upload error:', e)
    error.value = e.response?.data?.error || 'Failed to upload sample'
  } finally {
    uploading.value = false
  }
}

// Initialize auth in onMounted
onMounted(async () => {
  try {
    const user = await auth.init();
    console.log('Auth initialized with user:', user?.ID);
    if (!user) {
      throw new Error('Failed to load user');
    }
    authInitialized.value = true;
  } catch (e) {
    console.error('Failed to initialize auth:', e);
  }
})

// Debug auth state changes
watch(() => currentUser.value, (newUser) => {
  console.log('Auth user updated in SampleUpload:', newUser?.ID)
  if (newUser) {
    console.log('Current uploads after auth update:', currentUploads.value.length)
  }
}, { immediate: true })

// Add watch for samples prop to update UI
watch(() => props.samples, (newSamples) => {
  console.log('Samples prop updated:', newSamples?.length, 'samples')
  if (currentUser.value) {
    console.log('Current user uploads:', currentUploads.value.length, 'for user:', currentUser.value.ID)
  }
  console.log('Remaining uploads:', remainingUploads.value)
}, { immediate: true, deep: true })
</script> 