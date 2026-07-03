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

export interface AnnouncementInput {
  severity: Banner['severity']
  title: string
  description?: string
  dismissible: boolean
}

export interface AnnouncementsFeed {
  fetchActive(): Promise<Banner[]>
  create(input: AnnouncementInput): Promise<Banner>
  remove(id: string): Promise<void>
}

const successMsg = (m: string) => ({ headers: { 'x-success-message': m } })

export function createRemoteAnnouncementsFeed(): AnnouncementsFeed {
  const client = () => getAuthenticatedClient()
  return {
    async fetchActive(): Promise<Banner[]> {
      const res = await request<{ data: AnnouncementResponse[] }>(client(), 'v1/announcements')
      return (res?.data ?? []).map(toBanner)
    },
    async create(input: AnnouncementInput): Promise<Banner> {
      const res = await request<{ data: AnnouncementResponse }>(client(), 'v1/announcements', {
        method: 'POST',
        json: { ...input, description: input.description ?? '' },
        ...successMsg('Announcement published'),
      })
      return toBanner(res.data)
    },
    async remove(id: string): Promise<void> {
      await request<void>(client(), `v1/announcements/${id}`, {
        method: 'DELETE',
        ...successMsg('Announcement retracted'),
      })
    },
  }
}

let activeFeed: AnnouncementsFeed = createRemoteAnnouncementsFeed()

const announcementsService: AnnouncementsFeed = {
  fetchActive: () => activeFeed.fetchActive(),
  create: (input) => activeFeed.create(input),
  remove: (id) => activeFeed.remove(id),
}

export default announcementsService

// Test-only seam.
export function __setAnnouncementsFeedForTests(feed: AnnouncementsFeed): void {
  activeFeed = feed
}

export function __resetAnnouncementsForTests(): void {
  activeFeed = createRemoteAnnouncementsFeed()
}
