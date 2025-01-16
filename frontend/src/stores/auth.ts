import { defineStore } from 'pinia'
import type { User } from '@/types'
import * as api from '@/api'

interface AuthState {
  user: User | null
  token: string | null
  isAuthenticated: boolean
  loading: boolean
  error: string | null
}

export const useAuthStore = defineStore('auth', {
  state: (): AuthState => ({
    user: null,
    token: null,
    isAuthenticated: false,
    loading: false,
    error: null
  }),

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
        this.token = data.access_token
        this.isAuthenticated = true
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
      this.token = null
      this.isAuthenticated = false
      localStorage.removeItem('access_token')
    }
  }
}) 