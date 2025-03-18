import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import SampleUpload from '../SampleUpload.vue'
import { api } from '@/api'

// Mock the API
vi.mock('@/api', () => ({
  api: {
    packs: {
      uploadSample: vi.fn()
    }
  }
}))

// Mock auth store
vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({
    user: { ID: 1, email: 'test@example.com' },
    init: vi.fn().mockResolvedValue({ ID: 1, email: 'test@example.com' })
  })
}))

describe('SampleUpload', () => {
  const mockProps = {
    packId: 1,
    uploadStart: new Date(Date.now() - 86400000).toISOString(), // Yesterday
    uploadEnd: new Date(Date.now() + 86400000).toISOString(), // Tomorrow
    samples: []
  }

  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('renders upload form when within time window', async () => {
    const wrapper = mount(SampleUpload, {
      props: mockProps
    })

    await flushPromises()

    expect(wrapper.find('input[type="file"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('Upload window is OPEN')
  })

  it('shows remaining upload count', async () => {
    const wrapper = mount(SampleUpload, {
      props: mockProps
    })

    await flushPromises()

    expect(wrapper.text()).toContain('You have 10 uploads remaining')
  })

  it('handles file selection', async () => {
    const wrapper = mount(SampleUpload, {
      props: mockProps
    })

    await flushPromises()

    const file = new File(['test audio'], 'test.wav', { type: 'audio/wav' })

    // Call the handler directly
    await wrapper.vm.handleFileSelect({
      target: {
        files: [file]
      }
    } as any)

    expect(wrapper.text()).toContain('test.wav')
  })

  it('handles successful upload', async () => {
    const mockResponse = {
      data: {
        ID: 1,
        filename: 'test.wav',
        fileSize: 1024
      }
    }
    vi.mocked(api.packs.uploadSample).mockResolvedValueOnce(mockResponse)

    const wrapper = mount(SampleUpload, {
      props: mockProps
    })

    await flushPromises()

    // Select file
    const file = new File(['test audio'], 'test.wav', { type: 'audio/wav' })
    await wrapper.vm.handleFileSelect({
      target: { files: [file] }
    } as any)

    // Trigger upload
    await wrapper.find('button').trigger('click')
    await flushPromises()

    expect(api.packs.uploadSample).toHaveBeenCalledWith(mockProps.packId, file)
    expect(wrapper.text()).toContain('Successfully uploaded')
  })

  it('handles upload error', async () => {
    const mockError = {
      response: {
        data: {
          error: 'Upload failed'
        }
      }
    }
    vi.spyOn(console, 'error').mockImplementation(() => { })
    vi.mocked(api.packs.uploadSample).mockRejectedValueOnce(mockError)

    const wrapper = mount(SampleUpload, {
      props: mockProps
    })

    await flushPromises()

    // Select file
    const file = new File(['test audio'], 'test.wav', { type: 'audio/wav' })
    await wrapper.vm.handleFileSelect({
      target: { files: [file] }
    } as any)

    // Trigger upload
    await wrapper.find('button').trigger('click')
    await flushPromises()

    expect(wrapper.find('[data-test="error-message"]').text()).toBe('Upload failed')
  })

  it('enforces file size limit', async () => {
    const wrapper = mount(SampleUpload, {
      props: mockProps
    })

    await flushPromises()

    const largeFile = new File(['test'], 'large.wav', { type: 'audio/wav' })
    Object.defineProperty(largeFile, 'size', { value: 26 * 1024 * 1024 })

    await wrapper.vm.handleFileSelect({
      target: { files: [largeFile] }
    } as any)

    expect(wrapper.text()).toContain('File size exceeds limit')
  })
})
