export interface User {
  ID: number
  email: string
}

export interface Sample {
  ID: number
  filename: string
  fileSize: number
  fileUrl?: string
  uploadedAt: string
  userId: number
  user: User
}

export interface SamplePack {
  ID: number
  title: string
  description: string
  startDate: string
  endDate: string
  uploadStart: string
  uploadEnd: string
  isActive: boolean
  samples: Sample[]
}

export interface Submission {
  ID: number
  title: string
  description: string
  fileUrl?: string
  user: User
} 