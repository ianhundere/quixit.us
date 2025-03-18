import { describe, it, expect, beforeEach, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useAuthStore } from '../auth'
import { api } from '@/api'

// Mock the API
vi.mock('@/api', () => ({
  api: {
    defaults: {
      headers: {
        common: {}
      }
    },
    auth: {
      getCurrentUser: vi.fn(),
      oauthCallback: vi.fn()
    }
  }
}))

// Mock router
const mockRouter = {
  push: vi.fn()
}

describe('Auth Store', () => {
  beforeEach(() => {
    // Create a fresh pinia instance for each test
    setActivePinia(createPinia())

    // Clear all mocks
    vi.clearAllMocks()

    // Clear localStorage
    localStorage.clear()
  })

  it('initializes with default values', () => {
    const store = useAuthStore()
    expect(store.user).toBeNull()
    expect(store.loading).toBe(false)
    expect(store.error).toBeNull()
    expect(store.isAuthenticated).toBe(false)
  })

  it('handles successful initialization with token', async () => {
    // Setup
    localStorage.setItem('token', 'fake-token')
    const mockUser = { ID: 1, email: 'test@example.com' }
    vi.mocked(api.auth.getCurrentUser).mockResolvedValueOnce({ data: mockUser })

    // Execute
    const store = useAuthStore()
    await store.init()

    // Verify
    expect(store.user).toEqual(mockUser)
    expect(store.isAuthenticated).toBe(true)
    expect(store.error).toBeNull()
    expect(api.defaults.headers.common['Authorization']).toBe('Bearer fake-token')
  })

  it('handles failed initialization', async () => {
    // Setup
    localStorage.setItem('token', 'invalid-token')
    const mockError = new Error('Auth failed')
    vi.spyOn(console, 'error').mockImplementation(() => { })
    vi.mocked(api.auth.getCurrentUser).mockRejectedValueOnce(mockError)

    // Execute
    const store = useAuthStore()

    // Verify
    await expect(store.init()).rejects.toThrow('Auth failed')
    expect(store.user).toBeNull()
    expect(store.isAuthenticated).toBe(false)
    expect(localStorage.getItem('token')).toBeNull()
    expect(api.defaults.headers.common['Authorization']).toBeUndefined()
  })

  it('handles successful OAuth callback', async () => {
    // Setup
    const mockToken = 'new-token'
    const mockUser = { ID: 1, email: 'test@example.com' }
    vi.mocked(api.auth.oauthCallback).mockResolvedValueOnce({ data: { token: mockToken } })
    vi.mocked(api.auth.getCurrentUser).mockResolvedValueOnce({ data: mockUser })

    // Execute
    const store = useAuthStore()
    await store.handleOAuthCallback('auth-code', 'github', mockRouter as any)

    // Verify
    expect(store.user).toEqual(mockUser)
    expect(localStorage.getItem('token')).toBe(mockToken)
    expect(api.defaults.headers.common['Authorization']).toBe(`Bearer ${mockToken}`)
    expect(mockRouter.push).toHaveBeenCalledWith('/')
  })

  it('handles logout correctly', () => {
    // Setup
    localStorage.setItem('token', 'fake-token')
    api.defaults.headers.common['Authorization'] = 'Bearer fake-token'

    // Execute
    const store = useAuthStore()
    store.user = { ID: 1, email: 'test@example.com' } as any
    store.logout(mockRouter as any)

    // Verify
    expect(store.user).toBeNull()
    expect(store.isAuthenticated).toBe(false)
    expect(localStorage.getItem('token')).toBeNull()
    expect(api.defaults.headers.common['Authorization']).toBeUndefined()
    expect(mockRouter.push).toHaveBeenCalledWith('/login')
  })
})
