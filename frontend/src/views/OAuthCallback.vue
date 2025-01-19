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
  token?: string
  provider?: string
}

const props = defineProps<Props>()
const router = useRouter()
const auth = useAuthStore()
const error = ref<string | null>(null)
const loading = ref(true)

onMounted(async () => {
  try {
    // Handle direct token (from dev login)
    if (props.token) {
      await auth.handleToken(props.token, router)
      return
    }

    // Handle OAuth code
    if (props.code) {
      await auth.handleOAuthCallback(props.code, props.provider || 'dev', router)
      return
    }

    error.value = 'No token or authorization code provided'
  } catch (e: any) {
    console.error('OAuth callback error:', e)
    error.value = e.response?.data?.error || e.message || 'Failed to complete login'
  } finally {
    loading.value = false
  }
})
</script> 