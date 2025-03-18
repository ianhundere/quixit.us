import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest'
import axios from 'axios'

// Mock axios before importing the client
vi.mock('axios', () => ({
  default: {
    create: vi.fn(() => ({
      interceptors: {
        response: {
          use: vi.fn()
        }
      },
      defaults: {
        headers: {
          common: {}
        }
      }
    }))
  }
}))

describe('API Client', () => {
  beforeEach(() => {
    // Clear localStorage and reset modules before each test
    localStorage.clear()
    vi.clearAllMocks()
    vi.resetModules()
  })

  afterEach(() => {
    localStorage.clear()
  })

  it('should create axios instance with correct base config', async () => {
    // Import client after setting up the test
    await import('../client')

    expect(axios.create).toHaveBeenCalledWith({
      baseURL: 'http://localhost:8080',
      headers: {
        'Content-Type': 'application/json',
      },
    })
  })

  it('should set authorization header if token exists in localStorage', async () => {
    localStorage.setItem('token', 'test-token')

    // Import client after setting token
    const { api } = await import('../client')

    expect(api.defaults.headers.common['Authorization']).toBe('Bearer test-token')
  })

  it('should not set authorization header if token does not exist', async () => {
    const { api } = await import('../client')
    expect(api.defaults.headers.common['Authorization']).toBeUndefined()
  })
})
