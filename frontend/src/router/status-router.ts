import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'

const StatusPageView = () => import('@/views/status-page/StatusPageView.vue')
const StatusPageDetailView = () => import('@/views/status-page/StatusPageDetailView.vue')

export const statusRoutes: RouteRecordRaw[] = [
  {
    path: '/status',
    name: 'StatusPage',
    component: StatusPageView,
  },
  {
    path: '/status/:id',
    name: 'StatusPageDetail',
    component: StatusPageDetailView,
  },
  {
    path: '/',
    redirect: '/status',
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/status',
  },
]

const statusRouter = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: statusRoutes,
})

export default statusRouter
