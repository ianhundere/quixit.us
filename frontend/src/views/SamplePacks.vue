<template>
  <div class="container mx-auto px-4 py-8">
    <h1 class="text-3xl font-bold mb-8">Sample Packs</h1>
    
    <div v-if="loading" class="text-center">
      <p>Loading sample packs...</p>
    </div>
    
    <div v-else-if="error" class="text-red-600">
      {{ error }}
    </div>
    
    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      <!-- Current Pack -->
      <div v-if="currentPack" class="bg-white shadow-lg rounded-lg p-6 border-2 border-blue-500">
        <h2 class="text-xl font-bold mb-2">Current Pack</h2>
        <div class="mb-4">
          <h3 class="text-lg font-semibold">{{ currentPack.title }}</h3>
          <p class="text-gray-600">{{ currentPack.description }}</p>
        </div>
        <router-link 
          class="text-indigo-600 hover:text-indigo-800"
          :to="`/packs/${getId(currentPack)}`"
        >
          View Details
        </router-link>
      </div>

      <!-- Past Packs -->
      <template v-for="pack in pastPacks" :key="getId(pack)">
        <div class="bg-white shadow-lg rounded-lg p-6">
          <div class="mb-4">
            <h3 class="text-lg font-semibold">{{ pack.title }}</h3>
            <p class="text-gray-600">{{ pack.description }}</p>
          </div>
          <router-link 
            class="text-indigo-600 hover:text-indigo-800"
            :to="`/packs/${getId(pack)}`"
          >
            View Details
          </router-link>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/api'
import type { SamplePack } from '@/types'
import { getId } from '@/utils/id'

const loading = ref(true)
const error = ref<string | null>(null)
const currentPack = ref<SamplePack | null>(null)
const pastPacks = ref<SamplePack[]>([])

onMounted(async () => {
  try {
    const response = await api.get<{ currentPack: SamplePack; pastPacks: SamplePack[] }>('/samples/packs')
    currentPack.value = response.data.currentPack
    pastPacks.value = response.data.pastPacks
  } catch (e: any) {
    error.value = e.response?.data?.message || 'Failed to load sample packs'
  } finally {
    loading.value = false
  }
})
</script> 