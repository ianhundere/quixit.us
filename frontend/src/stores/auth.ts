import { defineStore } from 'pinia';
import type { User } from '@/types';
import * as api from '@/api';
import client from '@/api/client';

export const useAuthStore = defineStore('auth', {
    state: () => {
        const token = localStorage.getItem('access_token');
        return {
            user: null as User | null,
            token: token,
            loading: false,
            error: null as string | null,
            initialized: false,
        }
    },

    getters: {
        isAuthenticated: (state) => !!state.token && state.initialized,
    },

    actions: {
        async init() {
            if (this.token) {
                this.loading = true;
                try {
                    const { data } = await client.get('/auth/me');
                    this.user = data.user;
                } catch (err) {
                    this.logout();
                } finally {
                    this.loading = false;
                }
            }
            this.initialized = true;
        },

        async login(email: string, password: string) {
            this.loading = true;
            this.error = null;
            try {
                const { data } = await api.auth.login(email, password);

                if (!data.access_token || !data.user) {
                    throw new Error('Invalid response format');
                }

                this.user = data.user;
                this.token = data.access_token;
                localStorage.setItem('access_token', data.access_token);
                this.initialized = true;
                return data;
            } catch (err: any) {
                this.error =
                    err.response?.data?.error || err.message || 'Login failed';
                throw err;
            } finally {
                this.loading = false;
            }
        },

        logout() {
            this.user = null;
            this.token = null;
            localStorage.removeItem('access_token');
            this.initialized = false;
        },
    },
});
