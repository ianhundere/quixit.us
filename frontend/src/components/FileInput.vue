<script setup lang="ts">
const props = defineProps<{
  accept?: string
  label?: string
  error?: string
  disabled?: boolean
  selectedFile?: File | null
}>()

const emit = defineEmits<{
  (e: 'fileSelected', file: File | null): void
}>()

const handleFileChange = (e: Event) => {
  const target = e.target as HTMLInputElement
  const file = target.files?.[0] || null
  emit('fileSelected', file)
}

const formatFileSize = (bytes: number) => {
  if (bytes === 0) return '0 Bytes'
  const k = 1024
  const sizes = ['Bytes', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}
</script>

<template>
  <div class="file-input">
    <label v-if="label" class="block text-sm font-medium text-gray-700 mb-1">
      {{ label }}
    </label>
    <div class="relative">
      <input
        type="file"
        :accept="accept"
        :disabled="disabled"
        @change="handleFileChange"
        class="block w-full text-sm text-gray-500
          file:mr-4 file:py-2 file:px-4 
          file:rounded-full file:border-0 
          file:text-sm file:font-semibold
          file:bg-indigo-50 file:text-indigo-700 
          hover:file:bg-indigo-100
          disabled:opacity-50 disabled:cursor-not-allowed"
      />
      <div v-if="selectedFile" class="mt-1 text-sm text-gray-600">
        Selected: {{ selectedFile.name }} 
        ({{ formatFileSize(selectedFile.size) }})
      </div>
    </div>
    <p v-if="error" class="mt-1 text-sm text-red-600">{{ error }}</p>
  </div>
</template> 