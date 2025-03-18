import { describe, it, expect } from 'vitest'
import type { SamplePack, Sample, User, Submission } from '..'

describe('Types', () => {
  describe('SamplePack', () => {
    it('should match the expected structure', () => {
      const samplePack: SamplePack = {
        ID: '1',
        title: 'Test Pack',
        description: 'Test Description',
        uploadStart: '2023-01-01',
        uploadEnd: '2023-01-03',
        startDate: '2023-01-04',
        endDate: '2023-01-10',
        isActive: true,
        samples: [],
        submissions: [],
        createdAt: '2023-01-01',
        updatedAt: '2023-01-01'
      }

      expect(samplePack).toHaveProperty('ID')
      expect(samplePack).toHaveProperty('title')
      expect(samplePack).toHaveProperty('description')
      expect(samplePack).toHaveProperty('uploadStart')
      expect(samplePack).toHaveProperty('uploadEnd')
      expect(samplePack).toHaveProperty('startDate')
      expect(samplePack).toHaveProperty('endDate')
      expect(samplePack).toHaveProperty('isActive')
      expect(samplePack).toHaveProperty('samples')
    })
  })

  describe('Sample', () => {
    it('should match the expected structure', () => {
      const sample: Sample = {
        ID: '1',
        title: 'Test Sample',
        description: 'Test Description',
        fileUrl: 'http://example.com/test.mp3',
        createdAt: '2023-01-01',
        updatedAt: '2023-01-01'
      }

      expect(sample).toHaveProperty('ID')
      expect(sample).toHaveProperty('title')
      expect(sample).toHaveProperty('description')
      expect(sample).toHaveProperty('fileUrl')
    })
  })

  describe('User', () => {
    it('should match the expected structure', () => {
      const user: User = {
        ID: '1',
        username: 'testuser',
        email: 'test@example.com',
        createdAt: '2023-01-01',
        updatedAt: '2023-01-01'
      }

      expect(user).toHaveProperty('ID')
      expect(user).toHaveProperty('email')
    })
  })

  describe('Submission', () => {
    it('should match the expected structure', () => {
      const submission: Submission = {
        ID: '1',
        title: 'Test Submission',
        description: 'Test Description',
        fileUrl: 'http://example.com/submission.mp3',
        userId: '1',
        packId: '1',
        user: {
          ID: '1',
          username: 'testuser',
          email: 'test@example.com',
          createdAt: '2023-01-01',
          updatedAt: '2023-01-01'
        },
        createdAt: '2023-01-01',
        updatedAt: '2023-01-01'
      }

      expect(submission).toHaveProperty('ID')
      expect(submission).toHaveProperty('title')
      expect(submission).toHaveProperty('description')
      expect(submission).toHaveProperty('fileUrl')
      expect(submission).toHaveProperty('user')
    })
  })
})
