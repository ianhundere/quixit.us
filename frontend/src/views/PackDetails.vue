<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute } from 'vue-router'
import { api } from '@/api'
import type { SamplePack, Sample, Submission } from '@/types'
import { getId } from '@/utils/id'
import { usePackStore } from '@/stores'
import AudioPlayer from '@/components/AudioPlayer.vue'
import FileInput from '@/components/FileInput.vue'
import { downloadFile } from '@/utils/download'

const route = useRoute()
const packStore = usePackStore()
const loading = ref(false)
const error = ref<string | null>(null)
const pack = ref<SamplePack | null>(null)
const authToken = computed(() => localStorage.getItem('token'))

const fetchPack = async () => {
  try {
    loading.value = true
    const packId = getId(route.params)
    if (!packId) return
    
    const { data } = await api.packs.get(Number(packId))
    pack.value = data
    
    // Add download URLs to samples
    if (data.samples) {
      data.samples.forEach((sample: Sample) => {
        const sampleId = getId(sample)
        if (sampleId) {
          sample.fileUrl = `/api/samples/download/${sampleId}?token=${authToken.value}`
        }
      })
    }
    
    // Add download URLs to submissions
    if (data.submissions) {
      data.submissions.forEach((submission: Submission) => {
        const submissionId = getId(submission)
        if (submissionId) {
          submission.fileUrl = `/api/submissions/${submissionId}/download?token=${authToken.value}`
        }
      })
    }
  } catch (e: any) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

// ... rest of the file remains unchanged ...