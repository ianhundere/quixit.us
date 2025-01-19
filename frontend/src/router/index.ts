import { createRouter, createWebHistory } from 'vue-router';
import { useAuthStore } from '@/stores/auth';

const router = createRouter({
    history: createWebHistory(),
    routes: [
        {
            path: '/',
            name: 'home',
            component: () => import('@/views/Home.vue'),
            meta: { requiresAuth: true }
        },
        {
            path: '/login',
            component: () => import('@/views/Login.vue'),
            meta: { guest: true }
        },
        {
            // OAuth provider callback (e.g. Discord)
            path: '/auth/:provider/callback',
            component: () => import('@/views/OAuthCallback.vue'),
            props: route => ({
                code: route.query.code,
                provider: route.params.provider
            })
        },
        {
            // Dev login callback
            path: '/auth/dev/login',
            redirect: to => {
                const { state } = to.query;
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
            path: '/auth/callback',
            name: 'auth-callback',
            component: () => import('@/views/OAuthCallback.vue'),
            props: route => ({
                code: route.query.code,
                token: route.query.token,
                provider: route.query.provider || 'dev'
            })
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
        },
        {
            path: '/samples/upload',
            name: 'upload-samples',
            component: () => import('@/views/SampleUpload.vue'),
            meta: { requiresAuth: true }
        },
        {
            path: '/tracks/submit',
            name: 'submit-track',
            component: () => import('@/views/TrackSubmission.vue'),
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
            await auth.init()
            if (!auth.user) {
                // Save intended destination
                localStorage.setItem('redirect_after_login', to.fullPath)
                return next('/auth/callback')
            }
        } catch (e) {
            console.error('Auth check failed:', e)
            return next('/auth/callback')
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
