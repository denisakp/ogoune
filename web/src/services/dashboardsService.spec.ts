import { describe, expect, it } from 'vitest'
import {
  __createMockFeedForTests,
  __createRemoteStubForTests,
} from './dashboardsService'
import type { Dashboard, WidgetInstance } from '@/types'

function makeWidget(id: string, pos: number): WidgetInstance {
  return { id, widgetTypeId: 'uptime-stat', position: pos, config: {} }
}

describe('dashboardsService (spec 070 / US2)', () => {
  describe('mock mode', () => {
    it('list returns fixture sorted by updatedAt desc', async () => {
      const feed = __createMockFeedForTests('user-default')
      const list = await feed.list()
      expect(list.length).toBeGreaterThan(0)
      for (let i = 1; i < list.length; i++) {
        expect(new Date(list[i - 1]!.updatedAt).getTime()).toBeGreaterThanOrEqual(
          new Date(list[i]!.updatedAt).getTime(),
        )
      }
    })

    it('get returns null for unknown id', async () => {
      const feed = __createMockFeedForTests('user-default')
      expect(await feed.get('does-not-exist')).toBeNull()
    })

    it('create assigns id + timestamps + appends to store', async () => {
      const feed = __createMockFeedForTests('user-default')
      const before = (await feed.list()).length
      const created = await feed.create({
        name: 'test',
        scope: { mode: 'tag', payload: { tagIds: ['x'] } },
        widgets: [makeWidget('w1', 0)],
        defaultTimeRange: '24h',
        refreshInterval: '30s',
        visibility: 'private',
        ownerId: 'user-default',
        ownerName: 'Me',
      })
      expect(created.id).toMatch(/^dash-/)
      expect(typeof created.createdAt).toBe('string')
      expect(typeof created.updatedAt).toBe('string')
      const after = await feed.list()
      expect(after.length).toBe(before + 1)
    })

    it('update by non-owner throws FORBIDDEN', async () => {
      const feed = __createMockFeedForTests('user-default')
      const list = await feed.list()
      const otherOwned = list.find((d) => d.ownerId !== 'user-default')!
      await expect(
        feed.update(otherOwned.id, { name: 'hijacked' }),
      ).rejects.toThrow('FORBIDDEN')
    })

    it('update by owner bumps updatedAt only when state changed', async () => {
      const feed = __createMockFeedForTests('user-default')
      const list = await feed.list()
      const owned = list.find((d) => d.ownerId === 'user-default')!

      const sameWidgets: Dashboard = await feed.update(owned.id, { widgets: owned.widgets })
      expect(sameWidgets.updatedAt).toBe(owned.updatedAt)

      const changed = await feed.update(owned.id, { name: owned.name + ' v2' })
      expect(new Date(changed.updatedAt).getTime()).toBeGreaterThan(new Date(owned.updatedAt).getTime())
    })

    it('saveLayout is idempotent when widgets unchanged (FR-030)', async () => {
      const feed = __createMockFeedForTests('user-default')
      const list = await feed.list()
      const owned = list.find((d) => d.ownerId === 'user-default')!
      const same = await feed.saveLayout(owned.id, owned.widgets)
      expect(same.updatedAt).toBe(owned.updatedAt)
    })

    it('saveLayout bumps updatedAt when widgets changed', async () => {
      const feed = __createMockFeedForTests('user-default')
      const list = await feed.list()
      const owned = list.find((d) => d.ownerId === 'user-default')!
      const newLayout = [...owned.widgets, makeWidget('w-new', owned.widgets.length)]
      const result = await feed.saveLayout(owned.id, newLayout)
      expect(new Date(result.updatedAt).getTime()).toBeGreaterThan(new Date(owned.updatedAt).getTime())
    })

    it('remove by owner deletes', async () => {
      const feed = __createMockFeedForTests('user-default')
      const list = await feed.list()
      const owned = list.find((d) => d.ownerId === 'user-default')!
      await feed.remove(owned.id)
      expect(await feed.get(owned.id)).toBeNull()
    })

    it('remove by non-owner throws FORBIDDEN', async () => {
      const feed = __createMockFeedForTests('user-default')
      const list = await feed.list()
      const other = list.find((d) => d.ownerId !== 'user-default')!
      await expect(feed.remove(other.id)).rejects.toThrow('FORBIDDEN')
    })
  })

  describe('remote stub', () => {
    it('throws not-implemented on every operation', async () => {
      const feed = __createRemoteStubForTests()
      await expect(feed.list()).rejects.toThrow(/not implemented/)
      await expect(feed.get('x')).rejects.toThrow(/not implemented/)
    })
  })
})
