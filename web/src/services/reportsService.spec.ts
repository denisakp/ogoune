import { beforeEach, describe, expect, it } from 'vitest'
import {
  __createMockFeedForTests,
  __createRemoteStubForTests,
} from './reportsService'

describe('reportsService (spec 070 / US1)', () => {
  describe('mock mode', () => {
    let feed: ReturnType<typeof __createMockFeedForTests>

    beforeEach(() => {
      feed = __createMockFeedForTests()
    })

    it('returns a default disabled monthly report on first fetch', async () => {
      const m = await feed.fetchMonthly()
      expect(m.enabled).toBe(false)
      expect(m.schedule).toBe('monthly-1st')
      expect(m.scope).toBe('all-resources')
      expect(m.lastSentAt).toBeNull()
    })

    it('persists toggle ON across fetches', async () => {
      const initial = await feed.fetchMonthly()
      await feed.saveMonthly({ ...initial, enabled: true })
      const reloaded = await feed.fetchMonthly()
      expect(reloaded.enabled).toBe(true)
    })

    it('returns history sorted by sentAt desc', async () => {
      const h = await feed.fetchHistory()
      for (let i = 1; i < h.length; i++) {
        expect(new Date(h[i - 1]!.sentAt).getTime()).toBeGreaterThanOrEqual(
          new Date(h[i]!.sentAt).getTime(),
        )
      }
    })

    it('caps history length at the provided limit', async () => {
      const h = await feed.fetchHistory(2)
      expect(h.length).toBe(2)
    })

    it('fetchPendingPreview returns null in MVP', async () => {
      const p = await feed.fetchPendingPreview()
      expect(p).toBeNull()
    })
  })

  describe('remote stub', () => {
    it('throws not-implemented on every operation', async () => {
      const feed = __createRemoteStubForTests()
      await expect(feed.fetchMonthly()).rejects.toThrow(/not implemented/)
      await expect(
        feed.saveMonthly({
          enabled: true,
          recipientEmail: 'x@y',
          schedule: 'monthly-1st',
          scope: 'all-resources',
          lastSentAt: null,
        }),
      ).rejects.toThrow(/not implemented/)
      await expect(feed.fetchHistory()).rejects.toThrow(/not implemented/)
      await expect(feed.fetchPendingPreview()).rejects.toThrow(/not implemented/)
    })
  })
})
