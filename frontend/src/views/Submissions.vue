<template>
  <div class="container mx-auto px-4 py-8">
    <h1 class="text-3xl font-bold mb-8">Your Submissions</h1>
    
    <div v-if="loading" class="text-center">
      <p>Loading submissions...</p>
    </div>
    
    <div v-else-if="error" class="text-red-600">
      {{ error }}
    </div>
    
    <div v-else class="space-y-6">
      <div v-for="submission in submissions" :key="getId(submission)" class="bg-white shadow-lg rounded-lg p-6">
        <div class="flex justify-between items-start">
          <div>
            <h2 class="text-xl font-bold">{{ submission.title }}</h2>
            <p class="text-gray-600 mt-2">{{ submission.description }}</p>
          </div>
          <a 
            :href="submission.fileUrl" 
            target="_blank"
            class="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600"
          >
            Download
          </a>
        </div>
      </div>
      
      <div v-if="submissions.length === 0" class="text-center text-gray-600">
        <p>You haven't made any submissions yet.</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/api'
import type { Submission } from '@/types'
import { getId } from '@/utils/id'

const loading = ref(true)
const error = ref<string | null>(null)
const submissions = ref<Submission[]>([])

onMounted(async () => {
  try {
    const response = await api.get<Submission[]>('/submissions')
    submissions.value = response.data
  } catch (e: any) {
    error.value = e.response?.data?.message || 'Failed to load submissions'
  } finally {
    loading.value = false
  }
})
</script> 