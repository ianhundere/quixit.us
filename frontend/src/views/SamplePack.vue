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
          <h3 class="font-semibold text-gray-700">Upload Period</h3>
          <p>{{ formatDate(pack?.uploadStart) }} - {{ formatDate(pack?.uploadEnd) }}</p>
        </div>
        <div>
          <h3 class="font-semibold text-gray-700">Active Period</h3>
          <p>{{ formatDate(pack?.startDate) }} - {{ formatDate(pack?.endDate) }}</p>
        </div>
      </div>
      
      <!-- Samples List -->
      <div class="mt-8">
        <h2 class="text-2xl font-bold mb-4">Samples</h2>
        <div class="space-y-4">
          <div v-for="sample in pack?.samples" :key="sample.ID" class="border rounded-lg p-4">
            <div class="flex justify-between items-center">
              <div>
                <p class="font-medium">{{ sample.filename }}</p>
                <p class="text-sm text-gray-600">Uploaded by {{ getUserDisplay(sample.user) }}</p>
              </div>
              <a 
                :href="sample.fileUrl" 
                target="_blank"
                class="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600"
              >
                Download
              </a>
            </div>
          </div>
        </div>
      </div>
      
      <!-- Download Pack Button -->
      <div class="mt-8">
        <button 
          @click="downloadPack"
          class="w-full bg-green-500 text-white px-6 py-3 rounded-lg font-semibold hover:bg-green-600 focus:outline-none focus:ring-2 focus:ring-green-500 focus:ring-offset-2"
          :disabled="downloading"
        >
          {{ downloading ? 'Downloading...' : 'Download All Samples' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute } from 'vue-router'
import * as api from '@/api'
import type { SamplePack, User } from '@/types'
import { getId } from '@/utils/id'

const route = useRoute()
const loading = ref(true)
const downloading = ref(false)
const error = ref<string | null>(null)
const pack = ref<SamplePack | null>(null)

const authToken = computed(() => localStorage.getItem('access_token'))

const formatDate = (dateString?: string) => {
  if (!dateString) return ''
  return new Date(dateString).toLocaleDateString()
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

onMounted(async () => {
  const packId = parseInt(route.params.id as string)
  try {
    const { data } = await api.packs.get(packId)
    // Add auth token to sample URLs
    if (data.samples) {
      data.samples.forEach((sample: any) => {
        const sampleId = getId(sample)
        if (sampleId) {
          sample.fileUrl = `/api/samples/download/${sampleId}?token=${authToken.value}`
        }
      })
    }
    pack.value = data
  } catch (e: any) {
    error.value = e.response?.data?.message || 'Failed to load sample pack'
  } finally {
    loading.value = false
  }
})
</script> 