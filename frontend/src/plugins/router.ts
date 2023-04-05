import { createRouter, createWebHashHistory } from 'vue-router'
import TrainsListView from '@/views/TrainsListView.vue'
import TrainDetailView from '@/views/TrainDetailView.vue'
import TrainsDBProvider from '@/views/TrainsDBProvider.vue'
import NotFound from '@/views/NotFound.vue'

const router = createRouter({
  history: createWebHashHistory(),
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
          path: 'list',
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
