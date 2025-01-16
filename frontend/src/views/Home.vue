<script setup lang="ts">
import { onMounted } from 'vue'
import { usePackStore } from '@/stores/index'
import { useAuthStore } from '@/stores/auth'

const packStore = usePackStore()
const authStore = useAuthStore()

onMounted(() => {
  packStore.fetchPacks()
})
</script>

<template>
  <div class="min-h-screen bg-gray-50 py-8">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
      <!-- Loading State -->
      <div v-if="packStore.loading" class="flex justify-center py-8">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-600"></div>
      </div>

      <!-- Error State -->
      <div v-else-if="packStore.error" class="bg-red-50 text-red-600 p-4 rounded-lg mb-4">
        {{ packStore.error }}
      </div>

      <!-- Content -->
      <div v-else>
        <div class="text-right mb-4">
          <button
            @click="authStore.logout"
            class="text-sm text-gray-600 hover:text-gray-900"
          >
            Logout
          </button>
        </div>

        <!-- Current Pack -->
        <div v-if="packStore.currentPack?.ID" class="bg-white shadow rounded-lg p-6 mb-8">
          <h2 class="text-2xl font-bold mb-4">Current Pack</h2>
          <div class="space-y-4">
            <h3 class="text-xl">{{ packStore.currentPack.title }}</h3>
            <p class="text-gray-600">{{ packStore.currentPack.description }}</p>
            
            <!-- Time Windows -->
            <div class="grid grid-cols-2 gap-4 text-sm text-gray-600">
              <div>
                <p>Upload Window:</p>
                <p>{{ new Date(packStore.currentPack.uploadStart).toLocaleString() }}</p>
                <p>to</p>
                <p>{{ new Date(packStore.currentPack.uploadEnd).toLocaleString() }}</p>
              </div>
              <div>
                <p>Submission Window:</p>
                <p>{{ new Date(packStore.currentPack.startDate).toLocaleString() }}</p>
                <p>to</p>
                <p>{{ new Date(packStore.currentPack.endDate).toLocaleString() }}</p>
              </div>
            </div>

            <router-link 
              :to="`/packs/${packStore.currentPack.ID}`"
              class="inline-block bg-indigo-600 text-white px-4 py-2 rounded hover:bg-indigo-700"
            >
              View Details
            </router-link>
          </div>
        </div>
        <div v-else class="bg-white shadow rounded-lg p-6 mb-8">
          <p class="text-gray-600">No active sample pack at the moment.</p>
        </div>

        <!-- Past Packs -->
        <div v-if="packStore.pastPacks.length > 0" class="bg-white shadow rounded-lg p-6">
          <h2 class="text-2xl font-bold mb-4">Past Packs</h2>
          <div class="space-y-4">
            <div 
              v-for="pack in packStore.pastPacks" 
              :key="pack.ID"
              class="border-b pb-4 last:border-b-0"
            >
              <h3 class="text-xl">{{ pack.title }}</h3>
              <p class="text-gray-600">{{ pack.description }}</p>
              <router-link 
                :to="`/packs/${pack.ID}`"
                class="text-indigo-600 hover:text-indigo-800"
              >
                View Details →
              </router-link>
            </div>
          </div>
        </div>
        <div v-else class="bg-white shadow rounded-lg p-6">
          <p class="text-gray-600">No past sample packs available.</p>
        </div>
      </div>
    </div>
  </div>
</template> 