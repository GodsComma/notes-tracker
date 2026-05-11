import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '@/views/HomeView.vue'
import ClassView from '@/views/ClassView.vue'
import SourceView from '@/views/SourceView.vue'
import SearchView from '@/views/SearchView.vue'
import SemesterCalendarView from '@/views/SemesterCalendarView.vue'
import DeepReadView from '@/views/DeepReadView.vue'

export default createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', component: HomeView },
    { path: '/calendar', component: SemesterCalendarView },
    { path: '/library', component: DeepReadView },
    { path: '/deepread', redirect: '/library' },
    { path: '/classes/:id', component: ClassView },
    { path: '/sources/:id', component: SourceView },
    { path: '/search', component: SearchView }
  ]
})
