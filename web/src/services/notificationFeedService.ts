import type { NotificationFeedItem } from '@/types'
import { NOTIFICATIONS_FIXTURE } from '@/mocks/notifications.fixture'
import { getAuthenticatedClient, request } from '@/core/http/client'

/**
 * Spec 069 — internal contract between the bell UI and its (mocked) feed.
 * Documented at `specs/069-cross-cutting-ui/contracts/notification-feed.contract.md`.
 *
 * UI components MUST import only the `NotificationFeed` interface and the
 * default export below. Swapping to a real backend in a future PRD is a
 * one-file change: route `'remote'` through `ky` and keep the same shape.
 */
export interface NotificationFeed {
  fetch(): Promise<NotificationFeedItem[]>
  markRead(id: string): Promise<void>
  markAllRead(): Promise<void>
}

function createMockFeed(): NotificationFeed {
  const readIds = new Set<string>()
  let allReadAt: number | null = null

  return {
    async fetch() {
      const snapshot = NOTIFICATIONS_FIXTURE.map((n) => {
        const fixtureUnread = n.unread === true
        const isLocallyRead =
          readIds.has(n.id) ||
          (allReadAt !== null && new Date(n.occurredAt).getTime() <= allReadAt)
        return { ...n, unread: fixtureUnread && !isLocallyRead }
      })
      // Sorted most-recent first.
      snapshot.sort((a, b) => new Date(b.occurredAt).getTime() - new Date(a.occurredAt).getTime())
      return snapshot
    },
    async markRead(id: string) {
      readIds.add(id)
    },
    async markAllRead() {
      allReadAt = Date.now()
      readIds.clear()
    },
  }
}

// Real backend feed (spec 072). v1 endpoints wrap the payload in a { data }
// envelope; the DTO already matches NotificationFeedItem (camelCase) 1:1.
function createRemoteFeed(): NotificationFeed {
  return {
    async fetch(): Promise<NotificationFeedItem[]> {
      const res = await request<{ data: NotificationFeedItem[] }>(
        getAuthenticatedClient(),
        'v1/notifications',
        { searchParams: { per_page: 50 } },
      )
      return res?.data ?? []
    },
    async markRead(id: string): Promise<void> {
      await request<void>(getAuthenticatedClient(), `v1/notifications/${id}/read`, { method: 'POST' })
    },
    async markAllRead(): Promise<void> {
      await request<void>(getAuthenticatedClient(), 'v1/notifications/read-all', {
        method: 'POST',
        json: {},
      })
    },
  }
}

const mode = (import.meta.env.VITE_NOTIFICATION_FEED_MODE as string | undefined) ?? 'mock'

const notificationFeedService: NotificationFeed =
  mode === 'remote' ? createRemoteFeed() : createMockFeed()

export default notificationFeedService

// Test-only: get a fresh mock implementation (the module singleton retains
// session-local read state across tests otherwise).
export function __createMockFeedForTests(): NotificationFeed {
  return createMockFeed()
}

export function __createRemoteFeedForTests(): NotificationFeed {
  return createRemoteFeed()
}
