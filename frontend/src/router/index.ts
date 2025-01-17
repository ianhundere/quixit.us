import { createRouter, createWebHistory } from 'vue-router';
import { useAuthStore } from '@/stores/auth';
import type { RouteLocationNormalized, NavigationGuardNext } from 'vue-router';

const router = createRouter({
    history: createWebHistory(),
    routes: [
        {
            path: '/',
            name: 'Home',
            component: () => import('@/views/Home.vue'),
            meta: { requiresAuth: true },
        },
        {
            path: '/login',
            name: 'Login',
            component: () => import('@/views/Login.vue'),
            meta: { guest: true },
        },
        {
            path: '/register',
            name: 'Register',
            component: () => import('@/views/Register.vue'),
            meta: { guest: true },
        },
        {
            path: '/packs/:id',
            name: 'pack-details',
            component: () => import('@/views/PackDetails.vue'),
            meta: { requiresAuth: true },
        },
    ],
});

// Navigation guard
router.beforeEach(async (to, from, next) => {
    const auth = useAuthStore();

    // Check if the route requires authentication
    if (to.meta.requiresAuth && !auth.token) {
        next('/login');
        return;
    }

    // Prevent authenticated users from accessing login/register pages
    if (to.meta.guest && auth.token) {
        next('/');
        return;
    }

    next();
});

export default router;
