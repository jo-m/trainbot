import { createRouter, createWebHistory } from 'vue-router'
import TrainsView from '@/views/TrainsView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'root',
      component: TrainsView
    }
  ]
})

export default router
