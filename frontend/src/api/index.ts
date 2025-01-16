import axios from 'axios'
import type { SamplePack, Sample, Submission } from '@/types'

const baseURL = import.meta.env.VITE_API_URL
console.log('API Base URL:', baseURL)

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL,
  headers: {
    'Content-Type': 'application/json',
    'Accept': 'application/json'
  },
  withCredentials: true
})

// Add auth token to requests
api.interceptors.request.use(config => {
  const token = localStorage.getItem('access_token')
  if (token) {
    console.log('Setting auth token:', token.substring(0, 20) + '...')
    config.headers.Authorization = `Bearer ${token}`
  } else {
    console.log('No auth token found')
  }
  return config
}, error => {
  return Promise.reject(error)
})

// Debug interceptors
api.interceptors.request.use(
  config => {
    console.log('API Request:', {
      url: config.url,
      method: config.method,
      data: config.data,
      headers: config.headers,
      baseURL: config.baseURL
    })
    return config
  },
  error => {
    console.error('API Request Error:', error)
    return Promise.reject(error)
  }
)

// Add better error logging
api.interceptors.response.use(
  response => response,
  error => {
    if (error.response) {
      // The request was made and the server responded with a status code
      // that falls out of the range of 2xx
      console.error('Response Error:', {
        data: error.response.data,
        status: error.response.status,
        headers: error.response.headers
      })
    } else if (error.request) {
      // The request was made but no response was received
      console.error('Request Error:', error.request)
    } else {
      // Something happened in setting up the request that triggered an Error
      console.error('Error:', error.message)
    }
    return Promise.reject(error)
  }
)

export const auth = {
  login: (email: string, password: string) => 
    api.post('/auth/login', { email, password }),
  register: (email: string, password: string) => 
    api.post('/auth/register', { email, password })
}

export const packs = {
  list: () => api.get<{ currentPack: SamplePack, pastPacks: SamplePack[] }>('/samples/packs'),
  get: (id: number) => {
    console.log('Fetching pack:', id)
    return api.get<SamplePack>(`/samples/packs/${id}`)
  },
  uploadSample: (file: File) => {
    const formData = new FormData()
    formData.append('file', file)
    return api.post<Sample>('/samples/upload', formData)
  }
}

export const submissions = {
  create: (data: { title: string, description: string, samplePackId: number, file: File }) => {
    const formData = new FormData()
    formData.append('title', data.title)
    formData.append('description', data.description)
    formData.append('samplePackId', String(data.samplePackId))
    formData.append('file', data.file)
    return api.post<Submission>('/submissions', formData)
  },
  list: (packId: number) => 
    api.get<Submission[]>(`/submissions?pack_id=${packId}`)
}

// Add this temporarily to test API connection
export const testApi = () => api.get('/health') 