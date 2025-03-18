/// <reference types="vite/client" />

declare module '*.vue' {
    import type { DefineComponent } from 'vue';
    const component: DefineComponent<{}, {}, any>;
    export default component;
}

interface ImportMetaEnv {
    readonly VITE_API_URL: string;
}

interface ImportMeta {
    readonly env: ImportMetaEnv;
}

declare global {
    interface Window {
        __DEV_BYPASS_TIME_WINDOWS__: boolean;
    }
    var __DEV_BYPASS_TIME_WINDOWS__: boolean;
}
