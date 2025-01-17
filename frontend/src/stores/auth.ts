import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import type { Router } from 'vue-router';
import type { User } from '@/types';
import { api } from '@/api';

export const useAuthStore = defineStore('auth', () => {
    const user = ref<User | null>(null);
    const loading = ref(false);
    const error = ref<string | null>(null);

    // Computed property for authentication status
    const isAuthenticated = computed(() => !!user.value);

    // Check if we have a token in localStorage
    const token = localStorage.getItem('token');
    if (token) {
        api.defaults.headers.common['Authorization'] = `Bearer ${token}`;
    }

    // Initialize auth state
    const init = async () => {
        const token = localStorage.getItem('token');
        if (token) {
            api.defaults.headers.common['Authorization'] = `Bearer ${token}`;
            try {
                await getCurrentUser();
            } catch (e) {
                // If getting user fails, clear token
                localStorage.removeItem('token');
                delete api.defaults.headers.common['Authorization'];
            }
        }
    };

    // Handle OAuth callback
    const handleOAuthCallback = async (code: string, provider: string, router: Router) => {
        try {
            loading.value = true;
            error.value = null;

            const { data } = await api.auth.oauthCallback(code, provider);
            const { token, user: userData } = data;

            // Save token and set auth header
            localStorage.setItem('token', token);
            api.defaults.headers.common['Authorization'] = `Bearer ${token}`;

            // Update user state
            user.value = userData;

            // Redirect to home
            router.push('/');
        } catch (e: any) {
            error.value = e.response?.data?.error || e.message || 'OAuth login failed';
            throw error.value;
        } finally {
            loading.value = false;
        }
    };

    // Get current user
    const getCurrentUser = async () => {
        try {
            loading.value = true;
            error.value = null;

            const { data } = await api.auth.getCurrentUser();
            user.value = data;
        } catch (e: any) {
            error.value = e.response?.data?.error || e.message || 'Failed to get user';
            throw error.value;
        } finally {
            loading.value = false;
        }
    };

    // Logout
    const logout = (router: Router) => {
        user.value = null;
        localStorage.removeItem('token');
        delete api.defaults.headers.common['Authorization'];
        router.push('/login');
    };

    return {
        user,
        loading,
        error,
        isAuthenticated,
        init,
        handleOAuthCallback,
        getCurrentUser,
        logout
    };
});
