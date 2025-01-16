<template>
  <div class="upload-container">
    <h2>Upload Samples</h2>
    <div class="upload-status" v-if="timeWindow">
      <p>Upload window is {{ timeWindow.isOpen ? 'OPEN' : 'CLOSED' }}</p>
      <p v-if="timeWindow.isOpen">Closes in: {{ timeWindow.remainingTime }}</p>
    </div>
    
    <div class="upload-form" v-if="timeWindow?.isOpen">
      <input 
        type="file" 
        ref="fileInput"
        @change="handleFileSelect"
        accept="audio/*"
        multiple
      >
      <div class="selected-files" v-if="selectedFiles.length">
        <div v-for="file in selectedFiles" :key="file.name" class="file-item">
          <span>{{ file.name }}</span>
          <span>{{ formatFileSize(file.size) }}</span>
        </div>
      </div>
      <button @click="uploadFiles" :disabled="!selectedFiles.length">
        Upload Files
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'

const fileInput = ref(null)
const selectedFiles = ref([])
const timeWindow = ref(null)

const handleFileSelect = (event) => {
  selectedFiles.value = Array.from(event.target.files)
}

const formatFileSize = (bytes) => {
  if (bytes === 0) return '0 Bytes'
  const k = 1024
  const sizes = ['Bytes', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const uploadFiles = async () => {
  const formData = new FormData()
  selectedFiles.value.forEach(file => {
    formData.append('files[]', file)
  })

  try {
    const response = await fetch('/api/samples/upload', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      },
      body: formData
    })
    if (response.ok) {
      selectedFiles.value = []
      // TODO: Show success message
    }
  } catch (error) {
    console.error('Upload failed:', error)
  }
}

onMounted(async () => {
  // TODO: Fetch time window status
})
</script>

<style scoped>
.upload-container {
  max-width: 600px;
  margin: 0 auto;
  padding: 20px;
}

.upload-status {
  margin: 20px 0;
  padding: 10px;
  border-radius: 4px;
  background-color: #f5f5f5;
}

.file-item {
  display: flex;
  justify-content: space-between;
  padding: 8px;
  margin: 4px 0;
  background-color: #f9f9f9;
  border-radius: 4px;
}

button {
  margin-top: 20px;
  padding: 10px 20px;
  background-color: #42b983;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

button:disabled {
  background-color: #cccccc;
  cursor: not-allowed;
}
</style> 