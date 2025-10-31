import { createRouter, createWebHistory } from 'vue-router'
import MonitorsView from '@/views/MonitorsView.vue'
import ResourceDetailView from '@/views/ResourceDetailView.vue'
import IntegrationsView from '@/views/IntegrationsView.vue'
import SettingsView from '@/views/SettingsView.vue'
import IncidentsView from '@/views/IncidentsView.vue'
import IncidentDetailView from '@/views/IncidentDetailView.vue'

const routes = [
  {
    path: '/',
    redirect: '/monitors',
  },
  {
    path: '/monitors',
    name: 'Monitors',
    component: MonitorsView,
  },
  {
    path: '/monitors/:id',
    name: 'ResourceDetail',
    component: ResourceDetailView,
  },
  {
    path: '/incidents',
    name: 'Incidents',
    component: IncidentsView,
  },
  {
    path: '/incidents/:id',
    name: 'IncidentDetail',
    component: IncidentDetailView,
  },
  {
    path: '/settings',
    name: 'Settings',
    component: SettingsView,
  },
  {
    path: '/integrations',
    name: 'Integrations',
    component: IntegrationsView,
  },
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})

export default router
