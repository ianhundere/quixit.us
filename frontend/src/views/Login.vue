<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { testApi } from '@/api'

const auth = useAuthStore()
const router = useRouter()

const email = ref('')
const password = ref('')
const error = ref('')

const handleSubmit = async () => {
  try {
    error.value = ''
    console.log('Login attempt:', { email: email.value, password: password.value })
    const response = await auth.login(email.value, password.value)
    console.log('Login response:', response)
    router.push('/')
  } catch (e: any) {
    console.error('Login error details:', {
      message: e.message,
      response: e.response?.data,
      status: e.response?.status
    })
    error.value = e.response?.data?.error || e.message || 'Login failed'
  }
}

const testConnection = async () => {
  try {
    const response = await testApi()
    console.log('Health check response:', response.data)
  } catch (e) {
    console.error('Health check failed:', e)
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
    <div class="max-w-md w-full space-y-8">
      <h2 class="text-3xl font-bold text-center">Sign in</h2>
      
      <button 
        @click="testConnection"
        type="button"
        class="text-sm text-gray-600 hover:text-gray-900"
      >
        Test API Connection
      </button>
      
      <form @submit.prevent="handleSubmit" class="mt-8 space-y-6">
        <div v-if="error" class="text-red-500 text-center">{{ error }}</div>
        
        <div>
          <label for="email" class="sr-only">Email</label>
          <input
            id="email"
            v-model="email"
            type="email"
            required
            class="appearance-none rounded-md relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 focus:z-10 sm:text-sm"
            placeholder="Email address"
          />
        </div>
        
        <div>
          <label for="password" class="sr-only">Password</label>
          <input
            id="password"
            v-model="password"
            type="password"
            required
            class="appearance-none rounded-md relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 focus:z-10 sm:text-sm"
            placeholder="Password"
          />
        </div>

        <div>
          <button
            type="submit"
            :disabled="auth.loading"
            class="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50"
          >
            {{ auth.loading ? 'Logging in...' : 'Log in' }}
          </button>
        </div>
      </form>
      <div v-if="auth.error" class="text-red-600 text-center">
        {{ auth.error }}
      </div>
    </div>
  </div>
</template> 