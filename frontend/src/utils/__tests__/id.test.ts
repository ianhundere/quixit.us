import { describe, it, expect } from 'vitest'
import { getId } from '../id'

describe('id utils', () => {
  it('gets ID from uppercase property', () => {
    const obj = { ID: '123' }
    expect(getId(obj)).toBe('123')
  })

  it('gets id from lowercase property', () => {
    const obj = { id: '123' }
    expect(getId(obj)).toBe('123')
  })

  it('prefers uppercase ID over lowercase id', () => {
    const obj = { ID: '123', id: '456' }
    expect(getId(obj)).toBe('123')
  })

  it('returns undefined for null/undefined input', () => {
    expect(getId(null)).toBeUndefined()
    expect(getId(undefined)).toBeUndefined()
  })

  it('returns undefined when no id exists', () => {
    const obj = { foo: 'bar' }
    expect(getId(obj)).toBeUndefined()
  })
})
