import { describe, it, expect, vi, beforeEach } from 'vitest'
import { downloadFile } from '../download'

describe('download utils', () => {
  beforeEach(() => {
    vi.spyOn(Storage.prototype, 'getItem')
    localStorage.setItem('access_token', 'test-token')

    global.fetch = vi.fn()
  })

  it('sends authorization header', async () => {
    const mockResponse = new Response(new Blob(['test']))
    global.fetch = vi.fn().mockResolvedValue(mockResponse)

    await downloadFile('http://test.com/file')

    expect(fetch).toHaveBeenCalledWith('http://test.com/file', {
      headers: {
        Authorization: 'Bearer test-token'
      }
    })
  })

  it('returns blob on success', async () => {
    const expectedContent = 'test'
    const mockResponse = new Response(expectedContent)
    global.fetch = vi.fn().mockResolvedValue(mockResponse)

    const result = await downloadFile('http://test.com/file')

    const arrayBuffer = await result.arrayBuffer()
    const text = new TextDecoder().decode(arrayBuffer)
    expect(text).toBe(expectedContent)
  })

  it('throws error on failed response', async () => {
    const mockResponse = new Response(null, { status: 404 })
    global.fetch = vi.fn().mockResolvedValue(mockResponse)

    await expect(downloadFile('http://test.com/file')).rejects.toThrow('Download failed')
  })
})
