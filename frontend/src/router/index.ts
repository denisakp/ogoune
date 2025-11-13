import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/authStore'

const MonitorsView = () => import('@/views/MonitorsView.vue')
const ResourceDetailView = () => import('@/views/ResourceDetailView.vue')
const SettingsView = () => import('@/views/SettingsView.vue')
const IncidentsView = () => import('@/views/IncidentsView.vue')
const IncidentDetailView = () => import('@/views/IncidentDetailView.vue')
const StatusPageView = () => import('@/views/StatusPageView.vue')
const StatusPageDetailView = () => import('@/views/StatusPageDetailView.vue')
const LoginView = () => import('@/views/LoginView.vue')

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: '/monitors',
  },
  {
    path: '/login',
    name: 'Login',
    component: LoginView,
    meta: { public: true, requiresLayout: false },
  },
  {
    path: '/monitors',
    name: 'Monitors',
    component: MonitorsView,
    meta: { requiresAuth: true, requiresLayout: true },
  },
  {
    path: '/monitors/:id',
    name: 'ResourceDetail',
    component: ResourceDetailView,
    meta: { requiresAuth: true, requiresLayout: true },
  },
  {
    path: '/incidents',
    name: 'Incidents',
    component: IncidentsView,
    meta: { requiresAuth: true, requiresLayout: true },
  },
  {
    path: '/incidents/:id',
    name: 'IncidentDetail',
    component: IncidentDetailView,
    meta: { requiresAuth: true, requiresLayout: true },
  },
  {
    path: '/settings',
    name: 'Settings',
    component: SettingsView,
    meta: { requiresAuth: true, requiresLayout: true },
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

// Navigation guard for authentication
router.beforeEach(async (to, from, next) => {
  const authStore = useAuthStore()

  // Check if route requires authentication
  const requiresAuth = to.matched.some((record) => record.meta.requiresAuth)

  // If route is public, allow access
  if (to.meta.public) {
    // If already authenticated and trying to access login, redirect to monitors
    if (to.path === '/login' && authStore.isAuthenticated) {
      next('/monitors')
      return
    }
    next()
    return
  }

  // If route requires auth and user is not authenticated
  if (requiresAuth && !authStore.isAuthenticated) {
    next('/login')
    return
  }

  // If authenticated but token hasn't been verified yet, verify it
  if (requiresAuth && authStore.isAuthenticated) {
    const isValid = await authStore.verify()
    if (!isValid) {
      next('/login')
      return
    }
  }

  next()
})

export default router
