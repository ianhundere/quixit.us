<script setup lang="ts">
import { ref, watch, computed, onUnmounted } from 'vue'
import type { Sample } from '@/types'

const props = defineProps<{
  sample: Sample | null
  isPlaying: boolean
}>()

const emit = defineEmits<{
  (e: 'playbackEnded'): void
  (e: 'canPlay'): void
  (e: 'error', error: string): void
}>()

const audioElement = ref<HTMLAudioElement | null>(null)
const authToken = computed(() => localStorage.getItem('access_token'))
const isLoading = ref(false)

// Clean up when component is unmounted
onUnmounted(() => {
  if (audioElement.value) {
    audioElement.value.pause()
    audioElement.value.src = ''
    audioElement.value.load()
  }
})

// Load and play audio
const setAudioSource = async (sample: Sample) => {
  try {
    isLoading.value = true

    if (!sample.fileUrl) {
      throw new Error('Invalid file URL')
    }

    // Reset audio element
    if (audioElement.value) {
      audioElement.value.pause()
      audioElement.value.src = ''
      audioElement.value.load()
    }

    // Fetch audio file with auth headers
    const response = await fetch(sample.fileUrl, {
      headers: {
        'Authorization': `Bearer ${authToken.value}`
      }
    })
    
    if (!response.ok) {
      throw new Error(`Failed to load audio file: ${response.statusText}`)
    }

    const blob = await response.blob()
    if (blob.size === 0) {
      throw new Error('Empty audio file')
    }

    // Convert blob to base64 data URL
    const reader = new FileReader()
    const dataUrl = await new Promise<string>((resolve, reject) => {
      reader.onload = () => resolve(reader.result as string)
      reader.onerror = () => reject(reader.error)
      reader.readAsDataURL(blob)
    })

    // Set the source and load the audio
    if (audioElement.value) {
      audioElement.value.src = dataUrl
      audioElement.value.load()
      
      if (props.isPlaying) {
        try {
          await audioElement.value.play()
        } catch (error) {
          console.error('Playback failed:', error)
          emit('error', 'Failed to play audio')
          emit('playbackEnded')
        }
      }
    }
  } catch (error) {
    console.error('Failed to load audio:', error)
    emit('error', error instanceof Error ? error.message : 'Failed to load audio')
    emit('playbackEnded')
  } finally {
    isLoading.value = false
  }
}

// Watch for sample changes
watch(() => props.sample, (newSample) => {
  if (newSample) {
    setAudioSource(newSample)
  } else if (audioElement.value) {
    audioElement.value.pause()
    audioElement.value.src = ''
    audioElement.value.load()
  }
}, { immediate: true })

// Watch for play/pause changes
watch(() => props.isPlaying, (playing) => {
  if (!audioElement.value) return

  if (playing) {
    const playPromise = audioElement.value.play()
    if (playPromise) {
      playPromise.catch(error => {
        console.error('Playback failed:', error)
        emit('error', 'Failed to play audio')
        emit('playbackEnded')
      })
    }
  } else {
    audioElement.value.pause()
    audioElement.value.currentTime = 0
  }
})

// Add a function to filter out known browser-specific errors
const handleError = (event: Event) => {
  const target = event.target as HTMLAudioElement
  const error = target?.error
  
  // Ignore known Firefox privacy errors
  if (error?.message?.includes('Invalid URI') || 
      error?.message?.includes('Failed to open media')) {
    console.debug('Audio player: Ignoring Firefox privacy error -', error.message, 
      '(This is expected when privacy.resistFingerprinting is enabled)')
    return
  }

  // Only emit actual errors
  console.error('Audio player error:', error?.message)
  emit('error', error?.message || 'Audio playback error')
}
</script>

<template>
  <div class="audio-player" v-if="sample">
    <div class="player-content">
      <div class="player-info">
        <span class="filename">
          <span v-if="isLoading" class="loading">Loading...</span>
          <span v-else>{{ sample.filename }}</span>
        </span>
      </div>
      <audio 
        ref="audioElement" 
        @ended="emit('playbackEnded')"
        @error="handleError"
        @canplay="emit('canPlay')"
        preload="auto"
        :controls="true"
      />
    </div>
  </div>
</template>

<style scoped>
.audio-player {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  background: white;
  padding: 0.5rem;
  box-shadow: 0 -2px 10px rgba(0, 0, 0, 0.1);
  z-index: 50;
  height: 60px;
}

.player-content {
  max-width: 800px;
  margin: 0 auto;
  display: flex;
  align-items: center;
  gap: 1rem;
  height: 100%;
}

.player-info {
  min-width: 200px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.filename {
  font-size: 0.875rem;
  color: #4B5563;
}

.audio-player audio {
  flex: 1;
  height: 32px;
}

/* Hide some default audio controls for a more compact look */
.audio-player audio::-webkit-media-controls-enclosure {
  border-radius: 4px;
}

.audio-player audio::-webkit-media-controls-panel {
  background-color: #F3F4F6;
}

.loading {
  color: #6B7280;
  font-style: italic;
}

/* Add some padding at the bottom of the page to account for the player */
:global(body) {
  padding-bottom: 70px;
}
</style> 