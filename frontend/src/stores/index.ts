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
        const { data } = await api.packs.list()
        this.currentPack = data.currentPack
        this.pastPacks = data.pastPacks
      } catch (e: any) {
        this.error = e.response?.data?.error || 'Failed to load packs'
        throw e
      } finally {
        this.loading = false
      }
    }
  }
}) 