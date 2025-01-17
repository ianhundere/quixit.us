<template>
  <div class="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
    <div class="max-w-md w-full space-y-8">
      <div>
        <h2 class="mt-6 text-center text-3xl font-extrabold text-gray-900">
          {{ loading ? 'Logging you in...' : error ? 'Login Failed' : 'Login Successful' }}
        </h2>
        <p v-if="error" class="mt-2 text-center text-sm text-red-600">
          {{ error }}
        </p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

interface Props {
  code?: string
  provider?: string
}

const props = defineProps<Props>()
const router = useRouter()
const auth = useAuthStore()
const error = ref<string | null>(null)
const loading = ref(true)

onMounted(async () => {
  if (!props.code) {
    error.value = 'No authorization code provided'
    loading.value = false
    return
  }

  try {
    await auth.handleOAuthCallback(props.code, props.provider || 'dev', router)
  } catch (e: any) {
    error.value = e.message || 'Failed to complete login'
  } finally {
    loading.value = false
  }
})
</script> 