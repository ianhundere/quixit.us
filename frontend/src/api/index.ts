import client from './client'
import type { SamplePack, Sample, Submission, User } from '@/types'

export const auth = {
    login: (email: string, password: string) => client.post('/api/auth/login', { email, password }),
    register: (email: string, password: string) => client.post('/api/auth/register', { email, password }),
    me: () => client.get('/api/auth/me')
}

export const packs = {
    list: () => client.get<{ currentPack: SamplePack; pastPacks: SamplePack[] }>('/api/samples/packs'),
    get: (id: number) => client.get<SamplePack>(`/api/samples/packs/${id}`),
    uploadSample: (packId: number, file: File) => {
        const formData = new FormData()
        formData.append('file', file)
        return client.post<Sample>(`/api/samples/packs/${packId}/upload`, formData, {
            headers: {
                'Content-Type': 'multipart/form-data'
            }
        })
    },
    downloadPack: async (packId: number): Promise<Blob> => {
        const response = await client.get(`/api/samples/packs/${packId}/download`, {
            responseType: 'blob'
        });
        return response.data;
    }
}

export const submissions = {
    list: (packId: number) => client.get<Submission[]>(`/api/submissions?pack_id=${packId}`),
    create: (data: {
        title: string,
        description: string,
        samplePackId: number,
        file: File
    }) => {
        const formData = new FormData()
        formData.append('title', data.title)
        formData.append('description', data.description)
        formData.append('samplePackId', String(data.samplePackId))
        formData.append('file', data.file)

        return client.post<Submission>('/api/submissions', formData, {
            headers: {
                'Content-Type': 'multipart/form-data'
            }
        })
    }
}
