import { getAuthenticatedClient, request } from '@/core/http/client'
import type { Banner } from '@/stores/announcementStore'

/**
 * Announcements feed — active operator banners from the v1 API (option 2).
 * The backend stores instance-wide banners; dismissals are per-user client-side.
 */
interface AnnouncementResponse {
  id: string
  severity: Banner['severity']
  title: string
  description: string
  dismissible: boolean
  createdAt: string
}

function toBanner(a: AnnouncementResponse): Banner {
  return {
    id: a.id,
    severity: a.severity,
    title: a.title,
    description: a.description || undefined,
    dismissible: a.dismissible,
  }
}

export interface AnnouncementsFeed {
  fetchActive(): Promise<Banner[]>
}

export function createRemoteAnnouncementsFeed(): AnnouncementsFeed {
  const client = () => getAuthenticatedClient()
  return {
    async fetchActive(): Promise<Banner[]> {
      const res = await request<{ data: AnnouncementResponse[] }>(client(), 'v1/announcements')
      return (res?.data ?? []).map(toBanner)
    },
  }
}

let activeFeed: AnnouncementsFeed = createRemoteAnnouncementsFeed()

const announcementsService: AnnouncementsFeed = {
  fetchActive: () => activeFeed.fetchActive(),
}

export default announcementsService

// Test-only seam.
export function __setAnnouncementsFeedForTests(feed: AnnouncementsFeed): void {
  activeFeed = feed
}

export function __resetAnnouncementsForTests(): void {
  activeFeed = createRemoteAnnouncementsFeed()
}
