import { vi } from 'vitest'
import { config } from '@vue/test-utils'

// Mock environment variables
vi.stubGlobal('import.meta.env', {
  VITE_API_URL: 'http://localhost:8080',
})

// Global test utils configuration
config.global.mocks = {
  // Add any global mocks here
}
