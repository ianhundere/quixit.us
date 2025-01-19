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
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const error = ref<string | null>(null)
const loading = ref(true)

onMounted(async () => {
  try {
    const token = route.query.token as string
    const code = route.query.code as string
    const state = route.query.state as string
    const provider = route.query.provider as string || route.params.provider as string

    // Handle direct token (from OAuth callback or dev login)
    if (token) {
      await auth.handleToken(token, router)
      return
    }

    // Handle OAuth code by redirecting to backend
    if (code && provider) {
      // Get base URL without /api suffix
      const baseUrl = import.meta.env.VITE_API_URL?.replace('/api', '') || 'http://localhost:8080'
      // Redirect to backend OAuth callback with state parameter
      const callbackUrl = new URL(`${baseUrl}/api/auth/oauth/${provider}/callback`)
      callbackUrl.searchParams.set('code', code)
      if (state) {
        callbackUrl.searchParams.set('state', state)
      }
      window.location.href = callbackUrl.toString()
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