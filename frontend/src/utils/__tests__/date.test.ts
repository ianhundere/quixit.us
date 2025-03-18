import { describe, it, expect } from 'vitest'
import { formatDate } from '../date'

describe('date utils', () => {
  it('formats date string correctly', () => {
    const date = '2024-01-26T10:30:00Z'
    const formatted = formatDate(date)
    expect(formatted).toMatch(/\d{1,2}\/\d{1,2}\/\d{4}.*\d{1,2}:\d{2}.*/)
  })

  it('formats Date object correctly', () => {
    const date = new Date('2024-01-26T10:30:00Z')
    const formatted = formatDate(date)
    expect(formatted).toMatch(/\d{1,2}\/\d{1,2}\/\d{4}.*\d{1,2}:\d{2}.*/)
  })

  it('handles invalid dates', () => {
    const invalid = 'not-a-date'
    expect(() => formatDate(invalid)).not.toThrow()
  })
})
