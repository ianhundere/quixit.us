import { defineStore } from 'pinia'
import type { User } from '@/types'
import * as api from '@/api'

interface AuthResponse {
  access_token: string
  refresh_token: string
  user: User
}

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
        console.log('Auth store: Attempting login')
        const { data } = await api.auth.login(email, password)
        console.log('Auth store: Login response:', data)
        
        if (!data.access_token || !data.user) {
          throw new Error('Invalid response format')
        }

        this.user = data.user
        localStorage.setItem('access_token', data.access_token)
        return data
      } catch (err: any) {
        console.error('Auth store: Login error:', err)
        this.error = err.response?.data?.error || err.message || 'Login failed'
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