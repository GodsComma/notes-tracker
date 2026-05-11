import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import './assets/main.css'
import { initTheme } from './composables/useTheme'
import { useAuthStore } from './stores/auth'

initTheme()

const app = createApp(App)
const pinia = createPinia()
app.use(pinia)
app.use(router)

// Resolve auth state before mounting so the router guard has user info
const auth = useAuthStore()
auth.init().then(() => {
  router.beforeEach((to) => {
    if (to.meta.requiresAuth && !auth.user) return '/login'
    if (to.path === '/login' && auth.user) return '/'
  })
  app.mount('#app')
})
