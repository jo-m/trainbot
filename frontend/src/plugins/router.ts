import { createRouter, createWebHistory } from 'vue-router'
import TrainsListView from '@/views/TrainsListView.vue'
import TrainDetailView from '@/views/TrainDetailView.vue'
import TrainsDBProvider from '@/views/TrainsDBProvider.vue'
import NotFound from '@/views/NotFound.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'root',
      redirect: { name: 'trainsList' }
    },
    {
      path: '/trains',
      component: TrainsDBProvider,
      redirect: { name: 'trainsList' },
      children: [
        {
          path: 'list/:filter?',
          name: 'trainsList',
          component: TrainsListView
        },
        {
          path: ':id',
          name: 'trainDetail',
          component: TrainDetailView
        }
      ]
    },
    { path: '/:pathMatch(.*)*', name: 'notFound', component: NotFound }
  ]
})

export default router
