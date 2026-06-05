import type { RouteRecordRaw } from 'vue-router'
import { defineComponent, h } from 'vue'

// Spec 060 — public status page routes.
// The three views are built incrementally:
//   US1 (T042) → StatusPageView.vue (current snapshot)
//   US2 (T050) → StatusHistoryView.vue (incident archive)
//   US3 (T057) → StatusUptimeView.vue (90-day uptime)
// Until each view ships, an inline placeholder defers gracefully so the
// router stays mountable and feature-flag-safe.

const StatusPublicView = () => import('@/views/status/StatusPublicView.vue')
const StatusHistoryView = () => import('@/views/status/StatusHistoryView.vue')
const StatusUptimeView = () => import('@/views/status/StatusUptimeView.vue')
const StatusPageDetailView = () => import('@/views/status-page/StatusPageDetailView.vue')

const PlaceholderView = (label: string) =>
  defineComponent({
    name: 'PublicStatusPlaceholder',
    setup() {
      return () =>
        h(
          'div',
          { class: 'min-h-screen flex items-center justify-center text-sm text-gray-500' },
          `${label} — coming soon`,
        )
    },
  })

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
    path: '/:pathMatch(.*)*',
    redirect: '/',
  },
]
