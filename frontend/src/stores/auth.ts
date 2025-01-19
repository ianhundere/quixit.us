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
        if (!token) {
            console.log('No token found in localStorage');
            return null;
        }

        try {
            loading.value = true;
            error.value = null;

            // Set auth header
            api.defaults.headers.common['Authorization'] = `Bearer ${token}`;

            // Get current user
            const { data } = await api.auth.getCurrentUser();
            if (!data || !data.ID) {
                throw new Error('Invalid user data received');
            }

            user.value = data;
            console.log('User loaded:', user.value.ID);

            return user.value;
        } catch (e: any) {
            console.error('Failed to initialize auth:', e);
            // If getting user fails, clear token and user
            localStorage.removeItem('token');
            delete api.defaults.headers.common['Authorization'];
            user.value = null;
            error.value = e.response?.data?.error || e.message || 'Failed to initialize auth';
            throw error.value;
        } finally {
            loading.value = false;
        }
    };

    // Handle OAuth callback
    const handleOAuthCallback = async (code: string, provider: string, router: Router) => {
        try {
            loading.value = true;
            error.value = null;

            // Get token from OAuth callback
            const { data } = await api.auth.oauthCallback(code, provider);
            console.log('OAuth callback response:', data);

            if (!data || !data.token) {
                throw new Error('Invalid OAuth response - no token received');
            }

            // Handle the token
            await handleToken(data.token, router);
        } catch (e: any) {
            console.error('OAuth callback failed:', e);
            error.value = e.response?.data?.error || e.message || 'OAuth login failed';
            throw error.value;
        } finally {
            loading.value = false;
        }
    };

    // Handle direct token
    const handleToken = async (token: string, router: Router) => {
        try {
            loading.value = true;
            error.value = null;

            // Save token and set auth header
            localStorage.setItem('token', token);
            api.defaults.headers.common['Authorization'] = `Bearer ${token}`;

            // Get user data
            const { data } = await api.auth.getCurrentUser();
            console.log('Got user data:', data);

            if (!data || !data.ID) {
                throw new Error('Invalid user data received');
            }

            // Set user data
            user.value = data;
            console.log('Token handled successfully, user:', user.value.ID);

            // Redirect to home
            router.push('/');
        } catch (e: any) {
            console.error('Token handling failed:', e);
            // Clean up on failure
            localStorage.removeItem('token');
            delete api.defaults.headers.common['Authorization'];
            user.value = null;
            error.value = e.message || 'Failed to handle token';
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
            if (!data || !data.ID) {
                throw new Error('Invalid user data received');
            }

            user.value = data;
            return user.value;
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
        handleToken,
        getCurrentUser,
        logout
    };
});
