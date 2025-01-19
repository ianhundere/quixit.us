import { defineConfig, loadEnv } from 'vite';
import vue from '@vitejs/plugin-vue';
import path from 'path';

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
    // Load env from root directory
    const rootEnv = loadEnv(mode, path.resolve(__dirname, '..'), '');
    console.log('Environment loaded:', {
        mode,
        cwd: process.cwd(),
        rootDir: path.resolve(__dirname, '..'),
        DEV_BYPASS_TIME_WINDOWS: rootEnv.VITE_DEV_BYPASS_TIME_WINDOWS,
        raw: rootEnv
    });
    
    const bypassTimeWindows = rootEnv.VITE_DEV_BYPASS_TIME_WINDOWS === 'true';
    console.log('Bypass time windows:', bypassTimeWindows);
    
    return {
        plugins: [vue()],
        resolve: {
            alias: {
                '@': path.resolve(__dirname, './src'),
            },
        },
        server: {
            port: 3000,
            host: rootEnv.HOST_DOMAIN || 'localhost',
            proxy: {
                '/api': {
                    target: 'http://localhost:8080',
                    changeOrigin: true,
                    rewrite: (path) => path.replace(/^\/api/, '/api'),
                },
            },
        },
        define: {
            'window.__DEV_BYPASS_TIME_WINDOWS__': bypassTimeWindows,
            'globalThis.__DEV_BYPASS_TIME_WINDOWS__': bypassTimeWindows
        }
    };
});
