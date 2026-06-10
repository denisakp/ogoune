import type { NotificationFeedItem } from '@/types'

/**
 * Mocked feed for the notification bell — spec 069.
 * Replaced by a real backend in a future PRD; the UI imports through
 * `notificationFeedService` so swapping the source is a one-file change.
 */

function relativeIso(minutesAgo: number): string {
  return new Date(Date.now() - minutesAgo * 60_000).toISOString()
}

export const NOTIFICATIONS_FIXTURE: NotificationFeedItem[] = [
  {
    id: 'ntf-001',
    category: 'incident',
    severity: 'error',
    title: 'Resource down: api.example.com',
    description: 'HTTP check failed 3 times in a row.',
    occurredAt: relativeIso(3),
    deepLink: '/incidents',
    unread: true,
  },
  {
    id: 'ntf-002',
    category: 'incident',
    severity: 'warning',
    title: 'Component degraded: Public API',
    description: 'One of 3 resources is failing.',
    occurredAt: relativeIso(12),
    deepLink: '/incidents',
    unread: true,
  },
  {
    id: 'ntf-003',
    category: 'system',
    severity: 'info',
    title: 'Maintenance window scheduled',
    description: 'Tomorrow 02:00 → 02:30 UTC.',
    occurredAt: relativeIso(45),
    deepLink: '/maintenance',
    unread: true,
  },
  {
    id: 'ntf-004',
    category: 'incident',
    severity: 'success',
    title: 'Incident resolved: db-primary',
    description: 'Service restored after 4m 12s.',
    occurredAt: relativeIso(120),
    deepLink: '/incidents',
    unread: false,
  },
  {
    id: 'ntf-005',
    category: 'system',
    severity: 'warning',
    title: 'Notification channel auth failed',
    description: 'Slack webhook returned 401 — re-authorize in Settings.',
    occurredAt: relativeIso(180),
    deepLink: '/notifications',
    unread: false,
  },
  {
    id: 'ntf-006',
    category: 'general',
    severity: 'info',
    title: 'New feature: keyboard shortcuts',
    description: 'Press ? to see all available shortcuts.',
    occurredAt: relativeIso(60 * 24),
    unread: false,
  },
  {
    id: 'ntf-007',
    category: 'system',
    severity: 'info',
    title: 'SSL certificate renewed',
    description: 'cert for api.example.com renewed automatically.',
    occurredAt: relativeIso(60 * 26),
    unread: false,
  },
  {
    id: 'ntf-008',
    category: 'incident',
    severity: 'success',
    title: 'Incident resolved: cdn-edge',
    description: 'Cleared after 1m 03s.',
    occurredAt: relativeIso(60 * 48),
    deepLink: '/incidents',
    unread: false,
  },
]
