import { describe, expect, it } from 'vitest'
import { http, HttpResponse } from 'msw'
import { createRemoteAnnouncementsFeed } from './announcementsService'
import { server } from '@/test/msw/server'

describe('announcementsService (option 2 — real backend)', () => {
  it('fetchActive unwraps {data} and maps to banners', async () => {
    server.use(
      http.get('*/v1/announcements', () =>
        HttpResponse.json({
          data: [
            {
              id: 'a1',
              severity: 'warning',
              title: 'Maintenance',
              description: 'soon',
              dismissible: true,
              createdAt: '2026-07-03T10:00:00Z',
            },
            {
              id: 'a2',
              severity: 'info',
              title: 'Welcome',
              description: '',
              dismissible: false,
              createdAt: '2026-07-02T10:00:00Z',
            },
          ],
        }),
      ),
    )
    const banners = await createRemoteAnnouncementsFeed().fetchActive()
    expect(banners).toHaveLength(2)
    expect(banners[0]).toEqual({
      id: 'a1',
      severity: 'warning',
      title: 'Maintenance',
      description: 'soon',
      dismissible: true,
    })
    // empty description → undefined
    expect(banners[1]!.description).toBeUndefined()
  })

  it('returns [] when the envelope has no data', async () => {
    server.use(http.get('*/v1/announcements', () => HttpResponse.json({ data: null })))
    expect(await createRemoteAnnouncementsFeed().fetchActive()).toEqual([])
  })
})
