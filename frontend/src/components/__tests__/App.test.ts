import { describe, it, expect, vi, beforeEach } from 'vitest'
import { nextTick } from 'vue'
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
    { path: '/login', component: { template: '<div>Login</div>' } },
    { path: '/packs', component: { template: '<div>Packs</div>' } }
  ]
})

describe('App', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('renders navigation when authenticated', async () => {
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

    expect(wrapper.text()).toContain('Quixit')
    expect(wrapper.text()).toContain('Sample Packs')
    expect(wrapper.text()).toContain('Upload Samples')
    expect(wrapper.find('button').text()).toBe('Logout')
  })

  it('shows login button when not authenticated', async () => {
    const wrapper = mount(App, {
      global: {
        plugins: [router],
        stubs: {
          RouterLink: {
            template: '<a :href="$props.to"><slot /></a>',
            props: ['to']
          },
          RouterView: {
            template: '<div><slot /></div>'
          }
        }
      }
    })

    await nextTick()
    const loginLink = wrapper.find('a[href="/login"]')
    expect(loginLink.exists()).toBe(true)
    expect(loginLink.text()).toBe('Sign In')
    expect(wrapper.text()).not.toContain('Sample Packs')
    expect(wrapper.text()).not.toContain('Upload Samples')
  })

  it('handles logout', async () => {
    const auth = useAuthStore()
    auth.user = { ID: 1, email: 'test@example.com' }
    auth.logout = vi.fn()

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

    await wrapper.find('button').trigger('click')
    expect(auth.logout).toHaveBeenCalledWith(router)
  })

})
