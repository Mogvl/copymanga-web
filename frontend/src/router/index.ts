import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'home',
      component: () => import('../views/HomeView.vue')
    },
    {
      path: '/search',
      name: 'search',
      component: () => import('../views/SearchView.vue')
    },
    {
      path: '/comic/:pathWord',
      name: 'comic',
      component: () => import('../views/ComicView.vue')
    },
    {
      path: '/downloaded',
      name: 'downloaded',
      component: () => import('../views/DownloadedView.vue')
    }
  ]
})

export default router
