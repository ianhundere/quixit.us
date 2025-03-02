import { createRouter, createWebHistory } from 'vue-router';
import { useAuthStore } from '@/stores/auth';
import { api } from '@/api';

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
            component: () => import('@/views/OAuthCallback.vue')
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
            component: () => import('@/views/OAuthCallback.vue')
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
        },
        {
            path: '/tracks',
            name: 'past-tracks',
            component: () => import('@/views/PastTracks.vue'),
            meta: { requiresAuth: true }
        }
    ]
})

// Navigation guard
router.beforeEach(async (to: any, from: any, next: any) => {
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

    // Additional guard for track submission route
    if (to.name === 'submit-track' && auth.user) {
        try {
            // Fail-safe: add a direct check/redirect at the router level for additional security
            if (window.location.pathname === "/tracks/submit") {
                // Check if bypass is explicitly false to always redirect
                if (
                    (window as any).__DEV_BYPASS_TIME_WINDOWS__ === false || 
                    (globalThis as any).__DEV_BYPASS_TIME_WINDOWS__ === false
                ) {
                    console.log('Track submission is not allowed - time windows bypass is false')
                    return next('/')
                }
            }
            
            // check if bypass is enabled, if yes allow
            if ((globalThis as any).__DEV_BYPASS_TIME_WINDOWS__ === true) {
                return next()
            }
            
            // otherwise, check if we're in submission phase
            const response = await api.packs.get(1) // assuming pack ID 1 is the current pack
            const pack = response.data
            
            const now = new Date()
            const start = new Date(pack.startDate)
            const end = new Date(pack.endDate)
            const isInSubmissionPhase = now >= start && now <= end
            
            if (!isInSubmissionPhase) {
                console.log('Track submission is not allowed outside of submission phase')
                return next('/')
            }
        } catch (e) {
            console.error('Failed to verify submission phase:', e)
        }
    }

    next()
})

export default router
