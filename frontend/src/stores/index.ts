import { defineStore } from 'pinia';
import type { SamplePack } from '@/types';
import * as api from '@/api';

interface PackState {
    currentPack: SamplePack | null;
    pastPacks: SamplePack[];
    error: string | null;
    loading: boolean;
}

export const usePackStore = defineStore('pack', {
    state: (): PackState => ({
        currentPack: null,
        pastPacks: [],
        error: null,
        loading: false,
    }),

    actions: {
        async fetchPacks() {
            this.loading = true;
            this.error = null;
            try {
                const response = await api.packs.list();
                console.log('API Response:', response);
                console.log('Current Pack:', response.data.currentPack);
                console.log('Past Packs:', response.data.pastPacks);
                this.currentPack = response.data.currentPack;
                this.pastPacks = response.data.pastPacks;
            } catch (e: any) {
                console.error('Error fetching packs:', e);
                this.error = e.response?.data?.error || 'Failed to load packs';
                throw e;
            } finally {
                this.loading = false;
            }
        },
    },
});
