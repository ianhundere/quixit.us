import axios from 'axios'
import type { SamplePack, User, Submission } from '@/types'

// Create axios instance with base URL
export const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || 'http://localhost:8080/api',
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
    api.post<User>('/auth/register', { email, password }),
  login: (email: string, password: string) =>
    api.post<{ token: string; user: User }>('/auth/login', { email, password }),
  getCurrentUser: () =>
    api.get<User>('/auth/current-user'),
  oauthCallback: (code: string, provider: string) =>
    api.get<{ token: string; user: User }>(`/auth/oauth/${provider}/callback`, { params: { code } })
}

export const packs = {
  list: () => api.get('/samples/packs'),
  get: (id: number) => api.get<SamplePack & { submissions: Submission[] }>(`/samples/packs/${id}`),
  uploadSample: (packId: number, file: File) => {
    const formData = new FormData();
    formData.append('file', file);
    return api.post(`/samples/packs/${packId}/upload`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    });
  },
  download: (id: number) => api.get(`/samples/packs/${id}/download`, {
    responseType: 'blob'
  })
}

export const submissions = {
  list: (packId: number) => api.get('/submissions', { params: { pack_id: packId } }),
  get: (id: number) => api.get(`/submissions/${id}`),
  create: (submission: any) => api.post('/submissions', submission),
  download: (id: number) => api.get(`/submissions/${id}/download`)
}

// Add type for the API instance
declare module 'axios' {
  interface AxiosInstance {
    packs: typeof packs;
    auth: typeof auth;
    submissions: typeof submissions;
  }
}

// Attach namespaces to api instance
api.packs = packs;
api.auth = auth;
api.submissions = submissions;
