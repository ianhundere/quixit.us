import { defineConfig, loadEnv } from 'vite';
import vue from '@vitejs/plugin-vue';
import path from 'path';

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
    // Load env from root directory
    const rootEnv = loadEnv(mode, path.resolve(__dirname, '..'), '');
    const isProd = mode === 'production';
    
    // Determine host and port for API URL
    const hostDomain = isProd ? 'quixit.us' : (rootEnv.HOST_DOMAIN || 'localhost');
    const hostPort = isProd ? '' : `:${rootEnv.HOST_PORT || '3000'}`;
    const protocol = isProd ? 'https' : 'http';
    
    // In production, we don't include the port in the URL
    const apiUrl = isProd 
        ? `https://quixit.us/api`
        : `http://${hostDomain}${hostPort}/api`;
    
    return {
        plugins: [vue()],
        resolve: {
            alias: {
                '@': path.resolve(__dirname, './src'),
            },
        },
        server: {
            port: parseInt(rootEnv.HOST_PORT || '3000'),
            host: hostDomain,
            proxy: isProd ? {} : {
                '/api': {
                    target: `http://${hostDomain}${hostPort}`,
                    changeOrigin: true,
                    rewrite: (path) => path.replace(/^\/api/, '/api'),
                },
            },
        },
        define: {
            'window.__DEV_BYPASS_TIME_WINDOWS__': false,
            'globalThis.__DEV_BYPASS_TIME_WINDOWS__': false,
            'import.meta.env.VITE_API_URL': JSON.stringify(apiUrl)
        }
    };
});
