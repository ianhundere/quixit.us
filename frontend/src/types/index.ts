export interface User {
  id: number
  email: string
}

export interface SamplePack {
  id: number
  title: string
  description: string
  startDate: string
  endDate: string
  uploadStart: string
  uploadEnd: string
  isActive: boolean
  samples: Sample[]
}

export interface Sample {
  id: number
  filename: string
  fileSize: number
  fileUrl: string
  uploadedAt: string
  userId: number
  user: User
}

export interface Submission {
  id: number
  title: string
  description: string
  fileUrl: string
  fileSize: number
  submittedAt: string
  userId: number
  user: User
  samplePackId: number
  samplePack: SamplePack
} 