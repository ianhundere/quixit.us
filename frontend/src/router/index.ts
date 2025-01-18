import { createRouter, createWebHistory } from 'vue-router';
import { useAuthStore } from '@/stores/auth';

const router = createRouter({
    history: createWebHistory(),
    routes: [
        {
            path: '/',
            component: () => import('@/views/Home.vue'),
            meta: { requiresAuth: true }
        },
        {
            path: '/login',
            component: () => import('@/views/Login.vue'),
            meta: { guest: true }
        },
        {
            path: '/auth/:provider/callback',
            component: () => import('@/views/OAuthCallback.vue'),
            props: route => ({
                code: route.query.code,
                provider: route.params.provider
            })
        },
        {
            path: '/auth/callback',
            component: () => import('@/views/OAuthCallback.vue'),
            props: route => ({
                code: route.query.code,
                provider: route.query.provider || 'dev'
            })
        },
        {
            path: '/auth/dev/login',
            redirect: to => {
                const { client_id, redirect_uri, response_type, scope, state } = to.query
                return {
                    path: '/auth/callback',
                    query: {
                        code: 'dev-code',
                        provider: 'dev',
                        state
                    }
                }
            }
        },
        {
            path: '/packs',
            component: () => import('@/views/SamplePacks.vue'),
            meta: { requiresAuth: true }
        },
        {
            path: '/packs/:id',
            component: () => import('@/views/SamplePack.vue'),
            meta: { requiresAuth: true }
        },
        {
            path: '/submissions',
            component: () => import('@/views/Submissions.vue'),
            meta: { requiresAuth: true }
        }
    ]
})

// Navigation guard
router.beforeEach(async (to, from, next) => {
    const auth = useAuthStore()
    const token = localStorage.getItem('token')

    // If route requires auth and no token exists, redirect to login
    if (to.meta.requiresAuth && !token) {
        next('/login')
        return
    }

    // If we have a token but no user data, try to get current user
    if (token && !auth.user) {
        try {
            await auth.getCurrentUser()
        } catch (error) {
            // If getting user fails, clear token and redirect to login
            localStorage.removeItem('token')
            next('/login')
            return
        }
    }

    // If guest route and user is authenticated, redirect to home
    if (to.meta.guest && token) {
        next('/')
        return
    }

    next()
})

export default router
