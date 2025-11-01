import { createRouter, createWebHistory } from 'vue-router'
import MonitorsView from '@/views/MonitorsView.vue'
import ResourceDetailView from '@/views/ResourceDetailView.vue'
import SettingsView from '@/views/SettingsView.vue'
import IncidentsView from '@/views/IncidentsView.vue'
import IncidentDetailView from '@/views/IncidentDetailView.vue'
import StatusPageView from '@/views/StatusPageView.vue'
import PublicMonitorDetailView from '@/views/PublicMonitorDetailView.vue'

const routes = [
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
    name: 'PublicMonitorDetail',
    component: PublicMonitorDetailView,
    meta: { public: true, requiresLayout: false },
  },
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})

export default router
