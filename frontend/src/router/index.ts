import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import type { RouteLocationNormalized, NavigationGuardNext } from 'vue-router'

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
      path: '/packs/:id',
      name: 'pack-details',
      component: () => import('@/views/PackDetails.vue'),
      beforeEnter: (_to: RouteLocationNormalized, _from: RouteLocationNormalized, next: NavigationGuardNext) => {
        next()
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