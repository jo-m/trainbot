import { createRouter, createWebHashHistory } from 'vue-router'
import TrainsView from '@/views/TrainsView.vue'
import FavoritesView from '@/views/FavoritesView.vue'
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
          component: TrainsView
        },
        {
          path: 'favs',
          name: 'trainFavs',
          component: FavoritesView
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
