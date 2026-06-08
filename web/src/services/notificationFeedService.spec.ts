import { beforeEach, describe, expect, it } from 'vitest'
import { __createMockFeedForTests, __createRemoteStubForTests } from './notificationFeedService'

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

  describe('remote stub', () => {
    it('throws not-implemented on every operation', async () => {
      const feed = __createRemoteStubForTests()
      await expect(feed.fetch()).rejects.toThrow(/not implemented/)
      await expect(feed.markRead('x')).rejects.toThrow(/not implemented/)
      await expect(feed.markAllRead()).rejects.toThrow(/not implemented/)
    })
  })
})
