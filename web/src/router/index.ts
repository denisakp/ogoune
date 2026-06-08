import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/authStore'

const OverviewView = () => import('@/views/overview/OverviewView.vue')
const ResourcesView = () => import('@/views/resources/ResourcesView.vue')
const ResourceDetailView = () => import('@/views/resources/ResourceDetailView.vue')
const ComponentsView = () => import('@/views/ComponentsView.vue')
const SettingsLayoutView = () => import('@/views/settings/SettingsLayoutView.vue')
const AccountSettingsView = () => import('@/views/settings/AccountView.vue')
const SessionsSettingsView = () => import('@/views/settings/SessionsView.vue')
const TwoFactorSetupView = () => import('@/views/settings/TwoFactorSetupView.vue')
const NotificationsSettingsView = () => import('@/views/settings/NotificationsView.vue')
const ApiKeysSettingsView = () => import('@/views/settings/ApiKeysView.vue')
const EscalationSettingsView = () => import('@/views/settings/EscalationView.vue')
const OrgGeneralSettingsView = () => import('@/views/settings/OrgGeneralView.vue')
const StatusPageSettingsView = () => import('@/views/settings/StatusPageSettingsView.vue')
const TwoFactorRecoverView = () => import('@/views/auth/TwoFactorRecoverView.vue')
const TwoFactorResetView = () => import('@/views/auth/TwoFactorResetView.vue')
const Verify2FAView = () => import('@/views/auth/Verify2FAView.vue')
const IncidentsView = () => import('@/views/incidents/IncidentsView.vue')
const IncidentView = () => import('@/views/incidents/IncidentView.vue')
const LoginView = () => import('@/views/auth/LoginView.vue')
const RegisterView = () => import('@/views/auth/RegisterView.vue')
const ForgotPasswordView = () => import('@/views/auth/ForgotPasswordView.vue')
const ResetPasswordView = () => import('@/views/auth/ResetPasswordView.vue')
const InitializePasswordView = () => import('@/views/auth/InitializePasswordView.vue')
const MaintenanceView = () => import('@/views/maintenance/MaintenanceListView.vue')
const Error404View = () => import('@/views/errors/Error404View.vue')
const Error500View = () => import('@/views/errors/Error500View.vue')
const MaintenanceModeView = () => import('@/views/errors/MaintenanceModeView.vue')

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: '/overview',
  },
  {
    path: '/overview',
    name: 'Overview',
    component: OverviewView,
    meta: { requiresAuth: true, requiresLayout: true, breadcrumbLabel: 'Overview' },
  },
  {
    path: '/login',
    name: 'Login',
    component: LoginView,
    meta: { public: true, requiresLayout: false },
  },
  {
    path: '/register',
    name: 'Register',
    component: RegisterView,
    meta: { public: true, requiresLayout: false },
  },
  {
    path: '/forgot-password',
    name: 'ForgotPassword',
    component: ForgotPasswordView,
    meta: { public: true, requiresLayout: false },
  },
  {
    path: '/reset-password',
    name: 'ResetPassword',
    component: ResetPasswordView,
    meta: { public: true, requiresLayout: false },
  },
  {
    path: '/auth/initialize-password',
    name: 'InitializePassword',
    component: InitializePasswordView,
    meta: { public: true, requiresLayout: false },
  },
  {
    path: '/auth/verify-2fa',
    name: 'Verify2FA',
    component: Verify2FAView,
    meta: { public: true, requiresLayout: false },
  },
  {
    path: '/auth/2fa-recover',
    name: 'TwoFactorRecover',
    component: TwoFactorRecoverView,
    meta: { public: true, requiresLayout: false },
  },
  {
    path: '/2fa/reset',
    name: 'TwoFactorReset',
    component: TwoFactorResetView,
    meta: { public: true, requiresLayout: false },
  },
  {
    path: '/resources',
    name: 'Resources',
    component: ResourcesView,
    meta: { requiresAuth: true, requiresLayout: true, breadcrumbLabel: 'Resources' },
  },
  {
    path: '/resources/:id',
    name: 'ResourceDetail',
    component: ResourceDetailView,
    meta: { requiresAuth: true, requiresLayout: true, breadcrumbLabel: 'Resource' },
  },
  {
    path: '/monitors',
    name: 'Monitors',
    redirect: '/resources',
  },
  {
    path: '/monitors/:id',
    name: 'ResourceDetailLegacy',
    redirect: (to) => ({ name: 'ResourceDetail', params: { id: to.params.id } }),
  },
  {
    path: '/components',
    name: 'Components',
    component: ComponentsView,
    meta: { requiresAuth: true, requiresLayout: true, breadcrumbLabel: 'Components' },
  },
  {
    path: '/incidents',
    name: 'Incidents',
    component: IncidentsView,
    meta: { requiresAuth: true, requiresLayout: true, breadcrumbLabel: 'Incidents' },
  },
  {
    path: '/incidents/:id',
    name: 'IncidentDetail',
    component: IncidentView,
    meta: { requiresAuth: true, requiresLayout: true, breadcrumbLabel: 'Incident' },
  },
  {
    path: '/settings',
    component: SettingsLayoutView,
    meta: { requiresAuth: true, requiresLayout: true, breadcrumbLabel: 'Settings' },
    children: [
      { path: '', redirect: '/settings/account' },
      {
        path: 'account',
        name: 'SettingsAccount',
        component: AccountSettingsView,
        meta: { breadcrumbLabel: 'Account' },
      },
      {
        path: 'sessions',
        name: 'SettingsSessions',
        component: SessionsSettingsView,
        meta: { breadcrumbLabel: 'Sessions' },
      },
      {
        path: 'security/2fa',
        name: 'SettingsSecurity2FA',
        component: TwoFactorSetupView,
        meta: { breadcrumbLabel: 'Two-factor auth' },
      },
      {
        path: 'org/general',
        name: 'SettingsOrgGeneral',
        component: OrgGeneralSettingsView,
        meta: { breadcrumbLabel: 'General' },
      },
      {
        path: 'org/status-page',
        name: 'SettingsStatusPage',
        component: StatusPageSettingsView,
        meta: { breadcrumbLabel: 'Status Page' },
      },
    ],
  },
  {
    path: '/maintenance',
    name: 'Maintenance',
    component: MaintenanceView,
    meta: { requiresAuth: true, requiresLayout: true, breadcrumbLabel: 'Maintenance' },
  },
  // Top-level Settings entries (sidebar SETTINGS group).
  {
    path: '/notifications',
    name: 'Notifications',
    component: NotificationsSettingsView,
    meta: { requiresAuth: true, requiresLayout: true, breadcrumbLabel: 'Notifications' },
  },
  {
    path: '/escalation',
    name: 'Escalation',
    component: EscalationSettingsView,
    meta: { requiresAuth: true, requiresLayout: true, breadcrumbLabel: 'Escalation' },
  },
  {
    path: '/api-keys',
    name: 'ApiKeys',
    component: ApiKeysSettingsView,
    meta: { requiresAuth: true, requiresLayout: true, breadcrumbLabel: 'API keys' },
  },
  // Legacy /settings/* redirects (sidebar split per design).
  { path: '/settings/notifications', redirect: '/notifications' },
  { path: '/settings/escalation', redirect: '/escalation' },
  { path: '/settings/api-keys', redirect: '/api-keys' },
  // NB: the public status page lives in its own bundle (status.html → status-main.ts).
  // In dev: http://localhost:5173/status.html (and /status.html/uptime, /status.html/history).
  // In prod: served at status.<domain> or the custom_domain set in status page settings.

  // Spec 069 — branded error + maintenance surfaces. All public so anonymous
  // visitors hit them without an auth redirect (FR-004).
  {
    path: '/error-500',
    name: 'Error500',
    component: Error500View,
    meta: { public: true, requiresLayout: false },
  },
  {
    path: '/maintenance-mode',
    name: 'MaintenanceMode',
    component: MaintenanceModeView,
    meta: { public: true, requiresLayout: false },
  },
]

// Dev-only demo routes — build-time tree-shaken in production.
// Spec 053 FR-006, SC-007 + spec 055 FR-012 · contract: contracts/component-resolver.md
if (import.meta.env.DEV) {
  routes.push({
    path: '/_dev/nuxtui-demo',
    name: 'NuxtUIDemo',
    component: () => import('@/views/_dev/NuxtUIDemoView.vue'),
    meta: { public: true, requiresLayout: false },
  })
  routes.push({
    path: '/_dev/uform-example',
    name: 'UFormExample',
    component: () => import('@/views/_dev/UFormExampleView.vue'),
    meta: { public: true, requiresLayout: false },
  })
}

// Spec 069 — 404 catch-all MUST be the last route declared, otherwise dev
// demo routes (and any other later push) would never match.
routes.push({
  path: '/:pathMatch(.*)*',
  name: 'Error404',
  component: Error404View,
  meta: { public: true, requiresLayout: false },
})

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})

const MAINTENANCE_MODE_ENABLED =
  (import.meta.env.VITE_MAINTENANCE_MODE as string | undefined) === 'true'

// Spec 069 — maintenance gate runs BEFORE the auth guard so authenticated and
// anonymous visitors uniformly land on the branded MaintenanceMode screen.
router.beforeEach((to, _from, next) => {
  if (MAINTENANCE_MODE_ENABLED && to.name !== 'MaintenanceMode') {
    next({ name: 'MaintenanceMode' })
    return
  }
  next()
})

// Navigation guard for authentication
router.beforeEach(async (to, from, next) => {
  const authStore = useAuthStore()

  // Check if route requires authentication
  const requiresAuth = to.matched.some((record) => record.meta.requiresAuth)

  // If route is public, allow access
  if (to.meta.public) {
    // If already authenticated and trying to access login, redirect to overview
    if (to.path === '/login' && authStore.isAuthenticated) {
      next('/overview')
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
