import { describe, expect, it, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import type { Banner } from '@/stores/announcementStore'
import { __setAnnouncementsFeedForTests } from '@/services/announcementsService'

vi.mock('@/composables/useConfirm', () => ({ useConfirm: vi.fn(async () => true) }))

import AnnouncementsView from '../AnnouncementsView.vue'

const sample: Banner = { id: 'a1', severity: 'warning', title: 'Maintenance', description: 'soon', dismissible: true }

function makeFeed() {
  return {
    items: [sample] as Banner[],
    fetchActive: vi.fn(async function (this: { items: Banner[] }) {
      return feed.items
    }),
    create: vi.fn(async () => sample),
    remove: vi.fn(async () => {}),
  }
}
let feed: ReturnType<typeof makeFeed>

describe('AnnouncementsView (option 2 UI)', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    feed = makeFeed()
    __setAnnouncementsFeedForTests(feed)
  })

  it('lists active announcements', async () => {
    const w = mount(AnnouncementsView)
    await flushPromises()
    const items = w.findAll('[data-testid="active-item"]')
    expect(items).toHaveLength(1)
    expect(items[0]!.text()).toContain('Maintenance')
  })

  it('publishes a new announcement from the form', async () => {
    feed.items = []
    const w = mount(AnnouncementsView)
    await flushPromises()
    await w.find('input').setValue('New notice')
    await w.find('form').trigger('submit')
    await flushPromises()
    expect(feed.create).toHaveBeenCalledWith(
      expect.objectContaining({ title: 'New notice', severity: 'info' }),
    )
  })

  it('retracts an announcement after confirm', async () => {
    const w = mount(AnnouncementsView)
    await flushPromises()
    await w.find('[data-testid="active-item"] button[aria-label="Retract"]').trigger('click')
    await flushPromises()
    expect(feed.remove).toHaveBeenCalledWith('a1')
  })
})
