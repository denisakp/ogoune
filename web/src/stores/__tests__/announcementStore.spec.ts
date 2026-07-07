import { describe, expect, it, beforeEach, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'

const STORAGE_KEY = 'ogoune_dismissed_banners'

function mockLocalStorage() {
  const store: Record<string, string> = {}
  vi.stubGlobal('localStorage', {
    getItem: (k: string) => store[k] ?? null,
    setItem: (k: string, v: string) => {
      store[k] = v
    },
    removeItem: (k: string) => {
      delete store[k]
    },
    clear: () => {
      Object.keys(store).forEach((k) => delete store[k])
    },
    get length() {
      return Object.keys(store).length
    },
    key: (i: number) => Object.keys(store)[i] ?? null,
  })
  return store
}

describe('announcementStore', () => {
  let store: ReturnType<typeof mockLocalStorage>

  beforeEach(() => {
    store = mockLocalStorage()
    setActivePinia(createPinia())
  })

  it('publishes a banner and exposes it via `active` when not dismissed', async () => {
    const { useAnnouncementStore } = await import('@/stores/announcementStore')
    const s = useAnnouncementStore()
    s.publish({
      id: 'maintenance-2026-06-10',
      severity: 'warning',
      title: 'Maintenance scheduled',
      dismissible: true,
    })
    expect(s.active?.id).toBe('maintenance-2026-06-10')
  })

  it('dismiss() writes the id to localStorage and hides the banner from `active`', async () => {
    const { useAnnouncementStore } = await import('@/stores/announcementStore')
    const s = useAnnouncementStore()
    s.publish({ id: 'b1', severity: 'info', title: 't', dismissible: true })
    s.dismiss('b1')
    expect(s.active).toBeNull()
    expect(JSON.parse(store[STORAGE_KEY] ?? '[]')).toContain('b1')
  })

  it('re-instantiating with a seeded localStorage suppresses the dismissed banner', async () => {
    store[STORAGE_KEY] = JSON.stringify(['seeded'])
    const { useAnnouncementStore } = await import('@/stores/announcementStore')
    const s = useAnnouncementStore()
    s.publish({ id: 'seeded', severity: 'info', title: 't', dismissible: true })
    s.publish({ id: 'fresh', severity: 'info', title: 't2', dismissible: true })
    expect(s.active?.id).toBe('fresh')
  })

  it('publish() is idempotent on the same id', async () => {
    const { useAnnouncementStore } = await import('@/stores/announcementStore')
    const s = useAnnouncementStore()
    s.publish({ id: 'b1', severity: 'info', title: 't', dismissible: true })
    s.publish({ id: 'b1', severity: 'info', title: 't', dismissible: true })
    expect(s.banners).toHaveLength(1)
  })
})
