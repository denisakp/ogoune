import type { NotificationFeedItem } from '@/types'
import { NOTIFICATIONS_FIXTURE } from '@/mocks/notifications.fixture'

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

function createRemoteStub(): NotificationFeed {
  const notImplemented = (op: string) => async () => {
    throw new Error(`notificationFeedService: 'remote' mode not implemented yet (${op})`)
  }
  return {
    fetch: notImplemented('fetch'),
    markRead: notImplemented('markRead'),
    markAllRead: notImplemented('markAllRead'),
  }
}

const mode = (import.meta.env.VITE_NOTIFICATION_FEED_MODE as string | undefined) ?? 'mock'

const notificationFeedService: NotificationFeed =
  mode === 'remote' ? createRemoteStub() : createMockFeed()

export default notificationFeedService

// Test-only: get a fresh mock implementation (the module singleton retains
// session-local read state across tests otherwise).
export function __createMockFeedForTests(): NotificationFeed {
  return createMockFeed()
}

export function __createRemoteStubForTests(): NotificationFeed {
  return createRemoteStub()
}
