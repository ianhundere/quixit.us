import { defineStore } from 'pinia'
import type { SamplePack } from '@/types'
import * as api from '@/api'

export interface PackState {
  currentPack: SamplePack | null
  pastPacks: SamplePack[]
}

export const usePackStore = defineStore('pack', {
  state: () => ({
    currentPack: null as SamplePack | null,
    pastPacks: [] as SamplePack[],
    error: null as string | null,
    loading: false
  }),

  actions: {
    async fetchPacks() {
      this.loading = true
      this.error = null
      try {
        console.log('Fetching packs...')
        const { data } = await api.packs.list()
        console.log('Fetched packs response:', {
          currentPack: data.currentPack ? {
            id: data.currentPack.ID,
            title: data.currentPack.title,
            isActive: data.currentPack.isActive
          } : null,
          pastPacksCount: data.pastPacks?.length
        })
        
        this.currentPack = data.currentPack
        this.pastPacks = data.pastPacks
      } catch (e: any) {
        console.error('Failed to fetch packs:', e)
        this.error = e.response?.data?.error || 'Failed to load packs'
        throw e
      } finally {
        this.loading = false
      }
    }
  }
}) 