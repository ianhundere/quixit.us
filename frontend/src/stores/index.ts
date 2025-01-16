import { defineStore } from 'pinia'
import type { SamplePack, Submission } from '@/types'
import * as api from '@/api'

export const usePackStore = defineStore('pack', {
  state: () => ({
    currentPack: null as SamplePack | null,
    pastPacks: [] as SamplePack[],
    loading: false,
    error: null as string | null
  }),

  actions: {
    async fetchPacks() {
      this.loading = true
      try {
        const { data } = await api.packs.list()
        this.currentPack = data.currentPack
        this.pastPacks = data.pastPacks
      } catch (err) {
        this.error = 'Failed to fetch packs'
        console.error(err)
      } finally {
        this.loading = false
      }
    }
  }
}) 