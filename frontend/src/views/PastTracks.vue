<template>
  <div class="max-w-4xl mx-auto px-4 py-8">
    <div v-if="loading" class="text-center">
      <p>Loading past tracks...</p>
    </div>
    
    <div v-else-if="error" class="text-red-600">
      {{ error }}
    </div>
    
    <div v-else class="space-y-8">
      <h1 class="text-3xl font-bold mb-6">Past Tracks</h1>
      
      <!-- Past Quixits with Submissions -->
      <div v-for="pack in packs" :key="pack.ID" class="bg-white shadow rounded-lg p-6">
        <div class="border-b pb-4 mb-6">
          <h2 class="text-2xl font-bold mb-2">{{ pack.title.replace('Pack', 'Quixit') }}</h2>
          <p class="text-sm text-gray-500">
            Track submission period: {{ formatDateRange(pack.startDate, pack.endDate) }}
          </p>
        </div>
        
        <!-- Track submissions -->
        <div class="space-y-4">
          <div v-for="submission in pack.submissions" :key="submission.ID" 
               class="p-4 border rounded-lg hover:bg-gray-50 cursor-pointer"
               @click="playTrack(submission)">
            <div class="flex justify-between items-start">
              <div>
                <h3 class="font-medium text-lg flex items-center gap-2">
                  {{ submission.title }}
                  <span v-if="currentTrack?.ID === submission.ID" 
                        class="text-sm text-blue-500">
                    {{ isPlaying ? '▶️ Playing' : '⏸️ Paused' }}
                  </span>
                </h3>
                <p class="text-sm text-gray-600 mt-1">
                  Created by {{ submission.user?.email }}
                </p>
                <p class="text-xs text-gray-500 mt-1">
                  {{ formatDate(submission.createdAt) }}
                </p>
              </div>
              <div class="flex items-center space-x-2">
                <button v-if="submission.fileURL"
                        @click.stop="downloadSubmission(submission)"
                        class="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 flex items-center">
                  <span>Download Track</span>
                </button>
              </div>
            </div>
          </div>
          <div v-if="!pack.submissions?.length" class="text-center py-8 bg-gray-50 rounded-lg">
            <p class="text-gray-600 font-medium">No tracks submitted yet</p>
            <p class="text-sm text-gray-500 mt-1">Check back later for new tracks</p>
          </div>
        </div>
      </div>
      
      <div v-if="!packs.length" class="text-center py-8 text-gray-600">
        No past Quixits found.
      </div>
    </div>

    <!-- Audio Player -->
    <AudioPlayer 
      :sample="currentTrack" 
      :is-playing="isPlaying"
      @playback-ended="handlePlaybackEnded"
      @error="error = $event"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { api } from '@/api'
import AudioPlayer from '@/components/AudioPlayer.vue'

const auth = useAuthStore()
const packs = ref<any[]>([])
const loading = ref(true)
const error = ref('')
const currentTrack = ref<any>(null)
const isPlaying = ref(false)

onMounted(async () => {
  try {
    // Get current pack first
    const currentPackResponse = await api.packs.get(1)
    const currentPack = currentPackResponse.data
    console.log('Current pack:', currentPack)

    // Get submissions for current pack
    const submissionsResponse = await api.get('/submissions', { 
      params: { pack_id: currentPack.ID }
    })
    currentPack.submissions = submissionsResponse.data
    console.log('Submissions:', currentPack.submissions)

    packs.value = [currentPack]
  } catch (err) {
    console.error('Failed to load past tracks:', err)
    error.value = 'Failed to load past tracks. Please try again later.'
  } finally {
    loading.value = false
  }
})

const formatDate = (date: string) => {
  return new Date(date).toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const formatDateRange = (startDate: string, endDate: string) => {
  const start = new Date(startDate).toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
  const end = new Date(endDate).toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
  return `${start} - ${end}`
}

const downloadSubmission = async (submission: { ID: number, title: string }) => {
  try {
    const response = await api.get(`/submissions/${submission.ID}/download`, { 
      responseType: 'blob',
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('access_token')}`
      }
    })
    const url = window.URL.createObjectURL(new Blob([response.data]))
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', `${submission.title}.wav`)
    document.body.appendChild(link)
    link.click()
    link.remove()
    window.URL.revokeObjectURL(url)
  } catch (err) {
    console.error('Failed to download submission:', err)
    error.value = 'Failed to download submission. Please try again later.'
  }
}

const playTrack = (submission: any) => {
  // Convert submission to sample format expected by AudioPlayer
  const trackAsSample = {
    ID: submission.ID,
    filename: submission.title,
    fileUrl: `/submissions/${submission.ID}/download`
  }
  
  if (currentTrack.value?.ID === submission.ID) {
    isPlaying.value = !isPlaying.value
  } else {
    currentTrack.value = trackAsSample
    isPlaying.value = true
  }
}

const handlePlaybackEnded = () => {
  isPlaying.value = false
}
</script> 