import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import AudioPlayer from '../AudioPlayer.vue'
import { api } from '@/api'

// Mock the API
vi.mock('@/api', () => ({
  api: {
    get: vi.fn()
  }
}))

// Mock URL.createObjectURL and URL.revokeObjectURL
const mockObjectUrl = 'blob:mock-url'
global.URL.createObjectURL = vi.fn(() => mockObjectUrl)
global.URL.revokeObjectURL = vi.fn()

// Mock HTMLMediaElement
window.HTMLMediaElement.prototype.load = vi.fn()
window.HTMLMediaElement.prototype.play = vi.fn()
window.HTMLMediaElement.prototype.pause = vi.fn()

// Sample test data
const mockSample = {
  ID: 1,
  filename: 'test-audio.mp3',
  fileUrl: '/api/samples/1/audio',
  user: {
    ID: 1,
    name: 'Test User'
  }
}

describe('AudioPlayer', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.clear()
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('mounts with sample data', async () => {
    vi.mocked(api.get).mockResolvedValueOnce({
      data: new Blob(['mock audio data'])
    })

    const wrapper = mount(AudioPlayer, {
      props: {
        sample: mockSample,
        isPlaying: false
      }
    })

    // Initially shows loading
    expect(wrapper.find('.loading').exists()).toBe(true)

    // Wait for API call to resolve
    await flushPromises()

    expect(wrapper.find('.filename').text()).toContain(mockSample.filename)
    expect(wrapper.find('audio').exists()).toBe(true)
  })

  it('shows loading state while fetching audio', async () => {
    let resolvePromise: (value: any) => void
    const promise = new Promise(resolve => {
      resolvePromise = resolve
    })

    vi.mocked(api.get).mockImplementationOnce(() => promise)

    const wrapper = mount(AudioPlayer, {
      props: {
        sample: mockSample,
        isPlaying: false
      }
    })

    expect(wrapper.find('.loading').exists()).toBe(true)
    expect(wrapper.find('.loading').text()).toBe('Loading...')

    // Resolve the API call
    resolvePromise!({ data: new Blob(['mock audio data']) })
    await flushPromises()

    expect(wrapper.find('.loading').exists()).toBe(false)
  })

  it('handles audio load success', async () => {
    const mockBlob = new Blob(['mock audio data'])
    vi.mocked(api.get).mockResolvedValueOnce({
      data: mockBlob
    })

    const wrapper = mount(AudioPlayer, {
      props: {
        sample: mockSample,
        isPlaying: false
      }
    })

    await flushPromises()

    expect(api.get).toHaveBeenCalledWith(
      mockSample.fileUrl,
      expect.objectContaining({ responseType: 'blob' })
    )
    expect(URL.createObjectURL).toHaveBeenCalledWith(mockBlob)
  })

  it('handles audio load error', async () => {
    const mockError = new Error('Network error')
    vi.spyOn(console, 'error').mockImplementation(() => { })
    vi.mocked(api.get).mockRejectedValueOnce(mockError)

    const wrapper = mount(AudioPlayer, {
      props: {
        sample: mockSample,
        isPlaying: false
      }
    })

    await flushPromises()

    const emittedErrors = wrapper.emitted('error')
    expect(emittedErrors).toBeTruthy()
    expect(emittedErrors![0]).toEqual(['Failed to load audio'])
    expect(wrapper.emitted('playbackEnded')).toBeTruthy()
  })

  it('cleans up resources when unmounted', async () => {
    vi.mocked(api.get).mockResolvedValueOnce({
      data: new Blob(['mock audio data'])
    })

    const wrapper = mount(AudioPlayer, {
      props: {
        sample: mockSample,
        isPlaying: false
      }
    })

    await flushPromises()
    wrapper.unmount()

    expect(URL.revokeObjectURL).toHaveBeenCalledWith(mockObjectUrl)
  })

  it('responds to isPlaying prop changes', async () => {
    vi.mocked(api.get).mockResolvedValueOnce({
      data: new Blob(['mock audio data'])
    })

    const wrapper = mount(AudioPlayer, {
      props: {
        sample: mockSample,
        isPlaying: false
      }
    })

    await flushPromises()

    // Test play
    await wrapper.setProps({ isPlaying: true })
    expect(HTMLMediaElement.prototype.play).toHaveBeenCalled()

    // Test pause
    await wrapper.setProps({ isPlaying: false })
    expect(HTMLMediaElement.prototype.pause).toHaveBeenCalled()
  })

  it('emits playbackEnded event when audio ends', async () => {
    vi.mocked(api.get).mockResolvedValueOnce({
      data: new Blob(['mock audio data'])
    })

    const wrapper = mount(AudioPlayer, {
      props: {
        sample: mockSample,
        isPlaying: false
      }
    })

    await flushPromises()
    await wrapper.find('audio').trigger('ended')

    expect(wrapper.emitted('playbackEnded')).toBeTruthy()
  })
})
