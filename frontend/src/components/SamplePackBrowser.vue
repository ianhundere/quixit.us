<template>
  <div class="browser-container">
    <h2>Sample Packs</h2>
    
    <div class="current-pack" v-if="currentPack">
      <h3>Current Pack</h3>
      <div class="pack-info">
        <p>Available until: {{ formatDate(currentPack.endDate) }}</p>
        <div class="samples-list">
          <div v-for="sample in currentPack.samples" 
               :key="sample.ID" 
               class="sample-item">
            <span>{{ sample.filename }}</span>
            <div class="sample-controls">
              <button @click="playSample(sample)">Play</button>
              <button @click="downloadSample(sample)">Download</button>
            </div>
          </div>
        </div>
      </div>
    </div>
    
    <div class="past-packs" v-if="pastPacks.length">
      <h3>Past Packs</h3>
      <div v-for="pack in pastPacks" 
           :key="pack.ID" 
           class="pack-item">
        <span>{{ formatDate(pack.startDate) }} - {{ formatDate(pack.endDate) }}</span>
        <button @click="viewPack(pack)">View Details</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { api } from '@/api'
import type { SamplePack, Sample } from '@/types'

const currentPack = ref<SamplePack | null>(null)
const pastPacks = ref<SamplePack[]>([])
const authToken = computed(() => localStorage.getItem('token'))

const formatDate = (date: string): string => {
  if (!date) return ''
  return new Date(date).toLocaleDateString()
}

const playSample = (sample: Sample) => {
  // TODO: Implement audio playback
  console.log('Playing sample:', sample.filename)
}

const downloadSample = async (sample: Sample) => {
  try {
    const response = await fetch(`/api/samples/download/${sample.ID}`, {
      headers: {
        Authorization: `Bearer ${authToken.value}`
      }
    })
    if (response.ok) {
      const blob = await response.blob()
      const url = window.URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = sample.filename
      document.body.appendChild(a)
      a.click()
      window.URL.revokeObjectURL(url)
      a.remove()
    }
  } catch (error) {
    console.error('Download failed:', error)
  }
}

const viewPack = (pack: SamplePack) => {
  // TODO: Implement pack detail view
  console.log('Viewing pack:', pack.title)
}

onMounted(async () => {
  try {
    const response = await fetch('/api/samples/packs', {
      headers: {
        Authorization: `Bearer ${localStorage.getItem('token')}`
      }
    })
    if (response.ok) {
      const data = await response.json()
      currentPack.value = data.currentPack
      pastPacks.value = data.pastPacks
    }
  } catch (error) {
    console.error('Failed to fetch packs:', error)
  }
})
</script>

<style scoped>
.browser-container {
  max-width: 800px;
  margin: 0 auto;
  padding: 20px;
}

.sample-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px;
  margin: 5px 0;
  background-color: #f5f5f5;
  border-radius: 4px;
}

.sample-controls button {
  margin-left: 10px;
  padding: 5px 10px;
  background-color: #42b983;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.pack-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px;
  margin: 5px 0;
  background-color: #f9f9f9;
  border-radius: 4px;
}
</style> 