<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { api } from '@/api'

const router = useRouter()
const auth = useAuthStore()
const currentPack = ref<any>(null)
const loading = ref(true)

const isAuthenticated = computed(() => !!auth.user)
const isLoginPage = computed(() => router.currentRoute.value.path === '/login')

// determine if we're in the sample upload phase or track submission phase
const isInUploadPhase = computed(() => {
  if (!currentPack.value) return false
  if ((globalThis as any).__DEV_BYPASS_TIME_WINDOWS__) return true
  
  const now = new Date()
  const start = new Date(currentPack.value.uploadStart)
  const end = new Date(currentPack.value.uploadEnd)
  return now >= start && now <= end
})

const isInSubmissionPhase = computed(() => {
  if (!currentPack.value) return false
  if ((globalThis as any).__DEV_BYPASS_TIME_WINDOWS__) return true
  
  const now = new Date()
  const start = new Date(currentPack.value.startDate)
  const end = new Date(currentPack.value.endDate)
  return now >= start && now <= end
})

const showSubmitTrack = computed(() => {
  // allow submission when in submission phase or when the dev bypass is enabled
  return isInSubmissionPhase.value || (globalThis as any).__DEV_BYPASS_TIME_WINDOWS__
})

const logout = async () => {
  auth.logout(router)
}

// fetch current pack info on mount
onMounted(async () => {
  try {
    if (isAuthenticated.value) {
      const response = await api.packs.get(1) // assuming pack ID 1 is the current pack
      currentPack.value = response.data
    }
  } catch (e) {
    console.error('Failed to fetch current pack:', e)
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="min-h-screen bg-gray-50">
    <!-- Navigation -->
    <nav class="bg-white shadow">
      <div class="max-w-7xl mx-auto px-4">
        <div class="flex justify-between h-16">
          <div class="flex">
            <!-- Logo/Home -->
            <router-link to="/" class="flex-shrink-0 flex items-center">
              <span class="text-xl font-bold text-gray-900">Quixit</span>
            </router-link>
            
            <!-- Navigation Links -->
            <div v-if="isAuthenticated" class="hidden sm:ml-6 sm:flex sm:space-x-8">
              <router-link 
                to="/packs"
                class="border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700 inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium"
                :class="{ 'border-blue-500 text-gray-900': $route.path.startsWith('/packs') }"
              >
                Sample Packs
              </router-link>
              
              <router-link 
                to="/samples/upload"
                class="border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700 inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium"
                :class="{ 'border-blue-500 text-gray-900': $route.path.startsWith('/samples/upload') }"
              >
                Upload Samples
              </router-link>
              
              <!-- Only show Submit Track during submission phase -->
              <router-link 
                v-if="showSubmitTrack"
                to="/tracks/submit"
                class="border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700 inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium"
                :class="{ 'border-blue-500 text-gray-900': $route.path.startsWith('/tracks/submit') }"
              >
                Submit Track
              </router-link>

              <router-link 
                to="/tracks"
                class="border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700 inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium"
                :class="{ 'border-blue-500 text-gray-900': $route.path.startsWith('/tracks') }"
              >
                Past Tracks
              </router-link>
            </div>
          </div>
          
          <!-- Right side -->
          <div class="flex items-center">
            <router-link
              v-if="!isAuthenticated && !isLoginPage"
              to="/login"
              class="ml-3 inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700"
            >
              Sign In
            </router-link>
            <button
              v-else-if="isAuthenticated"
              @click="logout"
              class="ml-3 inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-gray-700 bg-gray-100 hover:bg-gray-200"
            >
              Logout
            </button>
          </div>
        </div>
      </div>
    </nav>

    <!-- Main Content -->
    <main>
      <router-view></router-view>
    </main>
  </div>
</template>

<style scoped>
.logo {
  height: 6em;
  padding: 1.5em;
  will-change: filter;
  transition: filter 300ms;
}

.logo:hover {
  filter: drop-shadow(0 0 2em #646cffaa);
}

.logo.vue:hover {
  filter: drop-shadow(0 0 2em #42b883aa);
}

.app-container {
  padding-bottom: 70px;
  /* Height of the audio player plus some extra space */
}
</style>
