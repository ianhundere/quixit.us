<template>
  <div class="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
    <div class="max-w-md w-full space-y-8">
      <div>
        <h2 class="mt-6 text-center text-3xl font-extrabold text-gray-900">
          Sign in to your account
        </h2>
      </div>
      <div class="mt-8 space-y-6">
        <div class="rounded-md shadow-sm space-y-4">
          <!-- OAuth Providers -->
          <button
            @click="loginWithProvider('github')"
            class="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-gray-800 hover:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-gray-500"
          >
            <span class="absolute left-0 inset-y-0 flex items-center pl-3">
              <i class="fab fa-github"></i>
            </span>
            Sign in with GitHub
          </button>

          <button
            @click="loginWithProvider('google')"
            class="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-red-600 hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500"
          >
            <span class="absolute left-0 inset-y-0 flex items-center pl-3">
              <i class="fab fa-google"></i>
            </span>
            Sign in with Google
          </button>

          <button
            @click="loginWithProvider('discord')"
            class="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
          >
            <span class="absolute left-0 inset-y-0 flex items-center pl-3">
              <i class="fab fa-discord"></i>
            </span>
            Sign in with Discord
          </button>

          <!-- Development Mode Login -->
          <button
            v-if="isDev"
            @click="loginWithProvider('dev')"
            class="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-green-600 hover:bg-green-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-green-500"
          >
            <span class="absolute left-0 inset-y-0 flex items-center pl-3">
              <i class="fas fa-code"></i>
            </span>
            Development Login
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const isDev = import.meta.env.DEV
const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080'

function loginWithProvider(provider: string) {
  window.location.href = `${apiUrl}/auth/oauth/${provider}`
}
</script>