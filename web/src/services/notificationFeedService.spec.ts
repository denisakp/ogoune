import { beforeEach, describe, expect, it } from 'vitest'
import { http, HttpResponse } from 'msw'
import { __createMockFeedForTests, __createRemoteFeedForTests } from './notificationFeedService'
import { server } from '@/test/msw/server'

describe('notificationFeedService (spec 069 / US4)', () => {
  describe('mock mode', () => {
    let feed: ReturnType<typeof __createMockFeedForTests>

    beforeEach(() => {
      feed = __createMockFeedForTests()
    })

    it('returns fixture items sorted by occurredAt desc', async () => {
      const items = await feed.fetch()
      expect(items.length).toBeGreaterThan(0)
      for (let i = 1; i < items.length; i++) {
        const prev = new Date(items[i - 1]!.occurredAt).getTime()
        const curr = new Date(items[i]!.occurredAt).getTime()
        expect(prev).toBeGreaterThanOrEqual(curr)
      }
    })

    it('markRead flips unread for the matching id and is idempotent', async () => {
      const before = await feed.fetch()
      const unread = before.find((n) => n.unread)!
      await feed.markRead(unread.id)
      await feed.markRead(unread.id)
      const after = await feed.fetch()
      expect(after.find((n) => n.id === unread.id)!.unread).toBe(false)
    })

    it('markAllRead clears the unread flag on every item', async () => {
      await feed.markAllRead()
      const items = await feed.fetch()
      expect(items.every((n) => n.unread === false)).toBe(true)
    })

    it('markRead on an unknown id does not throw', async () => {
      await expect(feed.markRead('does-not-exist')).resolves.toBeUndefined()
    })
  })

  describe('remote feed (spec 072)', () => {
    it('fetch unwraps the v1 {data} envelope into items', async () => {
      server.use(
        http.get('*/v1/notifications', () =>
          HttpResponse.json({
            data: [
              { id: 'n1', category: 'incident', severity: 'error', title: 'down', occurredAt: '2026-06-27T10:00:00Z', deepLink: '/incidents/n1', unread: true },
            ],
            meta: { page: 1, per_page: 50, total: 1 },
          }),
        ),
      )
      const feed = __createRemoteFeedForTests()
      const items = await feed.fetch()
      expect(items).toHaveLength(1)
      expect(items[0]!.id).toBe('n1')
      expect(items[0]!.unread).toBe(true)
    })

    it('markRead POSTs to /{id}/read', async () => {
      let hit = ''
      server.use(
        http.post('*/v1/notifications/:id/read', ({ params }) => {
          hit = String(params.id)
          return new HttpResponse(null, { status: 204 })
        }),
      )
      await __createRemoteFeedForTests().markRead('abc')
      expect(hit).toBe('abc')
    })

    it('markAllRead POSTs to /read-all', async () => {
      let called = false
      server.use(
        http.post('*/v1/notifications/read-all', () => {
          called = true
          return HttpResponse.json({ data: { marked: 3 } })
        }),
      )
      await __createRemoteFeedForTests().markAllRead()
      expect(called).toBe(true)
    })
  })
})
