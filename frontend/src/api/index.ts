import axios from 'axios'
import type { SamplePack, User, Submission } from '@/types'

// Create axios instance with default config
export const api = axios.create({
  baseURL: '/api',
  headers: {
    'Content-Type': 'application/json'
  }
})

// Add auth token to requests
api.interceptors.request.use(config => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// API namespaces
export const auth = {
  register: (email: string, password: string) =>
    api.post<User>('/auth/register', { email, password }, { baseURL: '/' }),
  login: (email: string, password: string) =>
    api.post<{ token: string; user: User }>('/auth/login', { email, password }, { baseURL: '/' }),
  getCurrentUser: () =>
    api.get<User>('/auth/me', { baseURL: '/' }),
  oauthCallback: (code: string, provider: string) =>
    api.get<{ token: string; user: User }>(`/auth/oauth/${provider}/callback`, {
      params: { code },
      baseURL: '/'
    })
}

export const packs = {
  list: () => api.get<{ currentPack: SamplePack; pastPacks: SamplePack[] }>('/samples/packs'),
  get: (id: number) => api.get<SamplePack & { submissions: Submission[] }>(`/samples/packs/${id}`),
  uploadSample: (packId: number, file: File) => {
    const formData = new FormData()
    formData.append('file', file)
    return api.post(`/samples/packs/${packId}/upload`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
  },
  downloadPack: (packId: number) => api.get(`/samples/packs/${packId}/download`, { responseType: 'blob' })
}

// Add type for the API instance
declare module 'axios' {
  interface AxiosInstance {
    packs: typeof packs;
    auth: typeof auth;
  }
}

// Attach namespaces to api instance
api.packs = packs;
api.auth = auth;
