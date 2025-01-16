import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'Home',
      component: () => import('@/views/Home.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/login',
      name: 'Login',
      component: () => import('@/views/Login.vue')
    },
    {
      path: '/register',
      name: 'Register',
      component: () => import('@/views/Register.vue')
    },
    {
      path: '/pack/:id',
      name: 'PackDetails',
      component: () => import('@/views/PackDetails.vue'),
      beforeEnter: (to, from, next) => {
        const id = parseInt(to.params.id as string)
        if (isNaN(id)) {
          console.error('Invalid pack ID:', to.params.id)
          next('/')
        } else {
          next()
        }
      }
    }
  ]
})

// Navigation guard
router.beforeEach((to, from, next) => {
  const auth = useAuthStore()
  const publicPages = ['/login', '/register']
  const requiresAuth = !publicPages.includes(to.path)

  if (requiresAuth && !auth.isAuthenticated) {
    next('/login')
  } else {
    next()
  }
})

export default router 