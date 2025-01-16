import client from './client'
import type { SamplePack, Sample, Submission } from '@/types'

export const auth = {
  login: (email: string, password: string) => 
    client.post('/auth/login', { email, password }),
  register: (email: string, password: string) => 
    client.post('/auth/register', { email, password })
}

export const packs = {
  list: () => client.get<{ currentPack: SamplePack, pastPacks: SamplePack[] }>('/samples/packs'),
  get: (id: number) => client.get<SamplePack>(`/samples/packs/${id}`),
  uploadSample: (file: File) => {
    const formData = new FormData()
    formData.append('file', file)
    return client.post<Sample>('/samples/upload', formData)
  }
}

export const submissions = {
  create: (data: { title: string, description: string, samplePackId: number, file: File }) => {
    const formData = new FormData()
    formData.append('title', data.title)
    formData.append('description', data.description)
    formData.append('samplePackId', String(data.samplePackId))
    formData.append('file', data.file)
    return client.post<Submission>('/submissions', formData)
  },
  list: (packId: number) => 
    client.get<Submission[]>(`/submissions?pack_id=${packId}`)
} 