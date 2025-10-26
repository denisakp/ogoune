import { createRouter, createWebHistory } from 'vue-router'
import MonitorsView from '@/views/MonitorsView.vue'
import IntegrationsView from '@/views/IntegrationsView.vue'
import ActivitiesView from '@/views/ActivitiesView.vue'
import SettingsView from '@/views/SettingsView.vue'

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
    path: '/settings',
    name: 'Settings',
    component: SettingsView,
  },
  {
    path: '/integrations',
    name: 'Integrations',
    component: IntegrationsView,
  },
  {
    path: '/activities',
    name: 'Activities',
    component: ActivitiesView,
  },
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})

export default router
