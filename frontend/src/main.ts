import { createApp } from 'vue';
import { createPinia } from 'pinia';
import App from './App.vue';
import router from './router';
import './style.css';
import { useAuthStore } from './stores/auth';

async function initApp() {
    const app = createApp(App);
    const pinia = createPinia();
    
    app.use(pinia);
    
    const authStore = useAuthStore(pinia);
    await authStore.init();
    
    app.use(router);
    
    // Wait for router to be ready before mounting
    await router.isReady();
    
    app.mount('#app');
}

initApp().catch(console.error);
