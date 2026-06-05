import { createRouter, createWebHashHistory, type RouteRecordRaw } from 'vue-router'

// Spec 060 — public status page routes.
// Mounted by status-main.ts (separate bundle, status.html entry).

const StatusPublicView = () => import('@/views/status/StatusPublicView.vue')
const StatusHistoryView = () => import('@/views/status/StatusHistoryView.vue')
const StatusUptimeView = () => import('@/views/status/StatusUptimeView.vue')
const StatusIncidentView = () => import('@/views/status/StatusIncidentView.vue')
const StatusPageDetailView = () => import('@/views/status-page/StatusPageDetailView.vue')

export const statusRoutes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'PublicStatusCurrent',
    component: StatusPublicView,
  },
  {
    path: '/history',
    name: 'PublicStatusHistory',
    component: StatusHistoryView,
  },
  {
    path: '/uptime',
    name: 'PublicStatusUptime',
    component: StatusUptimeView,
  },
  {
    path: '/resource/:id',
    name: 'PublicStatusResource',
    component: StatusPageDetailView,
  },
  {
    path: '/incidents/:id',
    name: 'PublicStatusIncident',
    component: StatusIncidentView,
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/',
  },
]

// Hash history → the public bundle works from any entry path (status.html in
// dev, root path under a custom domain in prod) without requiring SPA fallback
// from the server. Once the Host router lands (US5), we can revisit.
const statusRouter = createRouter({
  history: createWebHashHistory(),
  routes: statusRoutes,
})

export default statusRouter
