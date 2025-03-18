import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import SamplePackBrowser from '../SamplePackBrowser.vue'

// Mock fetch globally
const mockFetch = vi.fn()
global.fetch = mockFetch

// Mock localStorage
const mockLocalStorage = {
  getItem: vi.fn()
}
Object.defineProperty(window, 'localStorage', {
  value: mockLocalStorage
})

// Mock URL methods
const mockUrl = 'blob:mock-url'
global.URL.createObjectURL = vi.fn(() => mockUrl)
global.URL.revokeObjectURL = vi.fn()

// Sample test data
const mockCurrentPack = {
  ID: 1,
  title: 'Current Pack',
  startDate: '2025-01-01',
  endDate: '2025-01-31',
  samples: [
    {
      ID: 1,
      filename: 'test1.wav',
      fileUrl: '/api/samples/1'
    }
  ]
}

const mockPastPacks = [
  {
    ID: 2,
    title: 'Past Pack',
    startDate: '2024-12-01',
    endDate: '2024-12-31',
    samples: []
  }
]

describe('SamplePackBrowser', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockLocalStorage.getItem.mockReturnValue('fake-token')
  })

  it('renders current pack when available', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({
        currentPack: mockCurrentPack,
        pastPacks: []
      })
    })

    const wrapper = mount(SamplePackBrowser)
    await flushPromises()

    expect(wrapper.text()).toContain('Current Pack')
    expect(wrapper.text()).toContain(mockCurrentPack.samples[0].filename)
  })

  it('renders past packs list', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({
        currentPack: null,
        pastPacks: mockPastPacks
      })
    })

    const wrapper = mount(SamplePackBrowser)
    await flushPromises()

    expect(wrapper.text()).toContain('Past Packs')
    expect(wrapper.text()).toContain(new Date(mockPastPacks[0].startDate).toLocaleDateString())
  })

  it('handles sample download', async () => {
    const mockBlob = new Blob(['test audio'], { type: 'audio/wav' })
    mockFetch
      .mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({
          currentPack: mockCurrentPack,
          pastPacks: []
        })
      })
      .mockResolvedValueOnce({
        ok: true,
        blob: () => Promise.resolve(mockBlob)
      })

    const wrapper = mount(SamplePackBrowser)
    await flushPromises()

    // Mock createElement
    const mockLink = document.createElement('a')
    const mockClickFn = vi.fn()
    const mockRemoveFn = vi.fn()
    mockLink.click = mockClickFn
    mockLink.remove = mockRemoveFn
    vi.spyOn(document, 'createElement').mockReturnValue(mockLink)

    // Trigger download
    await wrapper.find('button:last-child').trigger('click')
    await flushPromises()

    expect(mockFetch).toHaveBeenCalledWith(
      `/api/samples/download/${mockCurrentPack.samples[0].ID}`,
      expect.objectContaining({
        headers: {
          Authorization: 'Bearer fake-token'
        }
      })
    )
    expect(URL.createObjectURL).toHaveBeenCalledWith(mockBlob)
    expect(mockClickFn).toHaveBeenCalled()
    expect(mockRemoveFn).toHaveBeenCalled()
  })

})
