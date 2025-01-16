import { defineStore } from 'pinia'
import type { User } from '@/types'
import * as api from '@/api'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    user: null as User | null,
    loading: false,
    error: null as string | null
  }),

  getters: {
    isAuthenticated: (state) => !!state.user
  },

  actions: {
    async login(email: string, password: string) {
      this.loading = true
      this.error = null
      try {
        const { data } = await api.auth.login(email, password)
        this.user = data.user
        localStorage.setItem('access_token', data.access_token)
        return data
      } catch (err) {
        this.error = 'Login failed'
        throw err
      } finally {
        this.loading = false
      }
    },

    logout() {
      this.user = null
      localStorage.removeItem('access_token')
    }
  }
}) 