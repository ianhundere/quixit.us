import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import LoadingSpinner from '../LoadingSpinner.vue'

describe('LoadingSpinner', () => {
  it('renders correctly', () => {
    const wrapper = mount(LoadingSpinner)
    expect(wrapper.find('.animate-spin').exists()).toBe(true)
  })

  it('has correct styling classes', () => {
    const wrapper = mount(LoadingSpinner)
    const spinner = wrapper.find('.animate-spin')

    expect(spinner.classes()).toContain('rounded-full')
    expect(spinner.classes()).toContain('border-b-2')
    expect(spinner.classes()).toContain('border-indigo-600')
  })

  it('has correct dimensions', () => {
    const wrapper = mount(LoadingSpinner)
    const spinner = wrapper.find('.animate-spin')

    expect(spinner.classes()).toContain('h-8')
    expect(spinner.classes()).toContain('w-8')
  })
})
