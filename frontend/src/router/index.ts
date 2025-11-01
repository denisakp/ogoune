import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'

const MonitorsView = () => import('@/views/MonitorsView.vue')
const ResourceDetailView = () => import('@/views/ResourceDetailView.vue')
const SettingsView = () => import('@/views/SettingsView.vue')
const IncidentsView = () => import('@/views/IncidentsView.vue')
const IncidentDetailView = () => import('@/views/IncidentDetailView.vue')
const StatusPageView = () => import('@/views/StatusPageView.vue')
const StatusPageDetailView = () => import('@/views/StatusPageDetailView.vue')

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: '/monitors',
  },
  {
    path: '/monitors',
    name: 'Monitors',
    component: MonitorsView,
    meta: { requiresLayout: true },
  },
  {
    path: '/monitors/:id',
    name: 'ResourceDetail',
    component: ResourceDetailView,
    meta: { requiresLayout: true },
  },
  {
    path: '/incidents',
    name: 'Incidents',
    component: IncidentsView,
    meta: { requiresLayout: true },
  },
  {
    path: '/incidents/:id',
    name: 'IncidentDetail',
    component: IncidentDetailView,
    meta: { requiresLayout: true },
  },
  {
    path: '/settings',
    name: 'Settings',
    component: SettingsView,
    meta: { requiresLayout: true },
  },
  // Public status page routes (no app layout)
  {
    path: '/status',
    name: 'StatusPage',
    component: StatusPageView,
    meta: { public: true, requiresLayout: false },
  },
  {
    path: '/status/:id',
    name: 'StatusPageDetail',
    component: StatusPageDetailView,
    meta: { public: true, requiresLayout: false },
  },
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})

export default router
