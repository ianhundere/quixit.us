import { describe, it, expect, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { createRouter, createWebHistory } from 'vue-router'
import { createPinia, setActivePinia } from 'pinia'
import App from '@/App.vue'
import { useAuthStore } from '@/stores/auth'

// Create a minimal router for testing
const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', component: { template: '<div>Home</div>' } },
    { path: '/packs', component: { template: '<div>Packs</div>' } }
  ]
})

describe('Navigation', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('shows all navigation items when authenticated', async () => {
    const auth = useAuthStore()
    auth.user = { ID: 1, email: 'test@example.com' }

    const wrapper = mount(App, {
      global: {
        plugins: [router],
        stubs: {
          RouterLink: {
            template: '<a><slot /></a>'
          },
          RouterView: {
            template: '<div><slot /></div>'
          }
        }
      }
    })

    const navItems = wrapper.findAll('a')
    expect(navItems.length).toBeGreaterThan(3) // Should have Home, Packs, Upload, etc.
    expect(navItems.some(item => item.text().includes('Sample Packs'))).toBe(true)
    expect(navItems.some(item => item.text().includes('Upload Samples'))).toBe(true)
  })

  it('shows only public navigation when not authenticated', () => {
    const wrapper = mount(App, {
      global: {
        plugins: [router],
        stubs: {
          RouterLink: {
            template: '<a><slot /></a>'
          },
          RouterView: {
            template: '<div><slot /></div>'
          }
        }
      }
    })

    const navItems = wrapper.findAll('a')
    expect(navItems.length).toBe(2) // Should only have Home and Sign In
    expect(navItems.some(item => item.text().includes('Sign In'))).toBe(true)
  })

  it('highlights current route', async () => {
    const auth = useAuthStore()
    auth.user = { ID: 1, email: 'test@example.com' }

    await router.push('/packs')

    const wrapper = mount(App, {
      global: {
        plugins: [router],
        stubs: {
          RouterLink: {
            template: '<a><slot /></a>'
          },
          RouterView: {
            template: '<div><slot /></div>'
          }
        }
      }
    })

    const activeNavItem = wrapper.find('.border-blue-500')
    expect(activeNavItem.exists()).toBe(true)
    expect(activeNavItem.text()).toContain('Sample Packs')
  })
})
