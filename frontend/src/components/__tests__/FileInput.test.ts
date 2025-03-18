import { describe, it, expect, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import FileInput from '../FileInput.vue'

describe('FileInput', () => {
  it('renders with default props', () => {
    const wrapper = mount(FileInput)
    expect(wrapper.find('input[type="file"]').exists()).toBe(true)
  })

  it('displays label when provided', () => {
    const label = 'Upload File'
    const wrapper = mount(FileInput, {
      props: { label }
    })
    expect(wrapper.find('label').text()).toBe(label)
  })

  it('shows error message when provided', () => {
    const error = 'Invalid file type'
    const wrapper = mount(FileInput, {
      props: { error }
    })
    expect(wrapper.find('.text-red-600').text()).toBe(error)
  })

  it('handles file selection', async () => {
    const wrapper = mount(FileInput)
    const file = new File(['test content'], 'test.txt', { type: 'text/plain' })

    // Create a mock event
    const event = {
      target: {
        files: [file]
      }
    } as unknown as Event

    // Call the handler directly
    await wrapper.vm.handleFileChange(event)

    expect(wrapper.emitted('fileSelected')?.[0][0]).toEqual(file)
  })

  it('displays selected file info', async () => {
    const file = new File(['test content'], 'test.txt', { type: 'text/plain' })
    const wrapper = mount(FileInput, {
      props: {
        selectedFile: file
      }
    })

    expect(wrapper.text()).toContain('test.txt')
    expect(wrapper.text()).toContain('12 Bytes')
  })

  it('respects accept prop', () => {
    const accept = '.mp3,.wav'
    const wrapper = mount(FileInput, {
      props: { accept }
    })
    expect(wrapper.find('input[type="file"]').attributes('accept')).toBe(accept)
  })

  it('handles disabled state', () => {
    const wrapper = mount(FileInput, {
      props: { disabled: true }
    })
    expect(wrapper.find('input[type="file"]').attributes('disabled')).toBeDefined()
  })

  it('formats file sizes correctly', () => {
    const wrapper = mount(FileInput)
    const cases = [
      { size: 500, expected: '500 Bytes' },
      { size: 1024, expected: '1 KB' },
      { size: 1024 * 1024, expected: '1 MB' }
    ]

    // Test the formatting function directly
    for (const { size, expected } of cases) {
      const formatted = wrapper.vm.formatFileSize(size)
      expect(formatted).toBe(expected)
    }
  })
})
