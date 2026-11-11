import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '@/views/HomeView.vue'
import ClassView from '@/views/ClassView.vue'
import SourceView from '@/views/SourceView.vue'
import SearchView from '@/views/SearchView.vue'
import SemesterCalendarView from '@/views/SemesterCalendarView.vue'
import DeepReadView from '@/views/DeepReadView.vue'
import LoginView from '@/views/LoginView.vue'

export default createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', component: LoginView },
    { path: '/', component: HomeView, meta: { requiresAuth: true } },
    { path: '/calendar', component: SemesterCalendarView, meta: { requiresAuth: true } },
    { path: '/library', component: DeepReadView, meta: { requiresAuth: true } },
    { path: '/deepread', redirect: '/library' },
    { path: '/classes/:id', component: ClassView, meta: { requiresAuth: true } },
    { path: '/sources/:id', component: SourceView, meta: { requiresAuth: true } },
    { path: '/search', component: SearchView, meta: { requiresAuth: true } },
    // Dev-only Playwright harness (client/e2e). Tree-shaken from production builds.
    ...(import.meta.env.DEV
      ? [{ path: '/__e2e/pdf-viewer', component: () => import('@/views/__e2e/PdfViewerHarness.vue') }]
      : [])
  ]
})
