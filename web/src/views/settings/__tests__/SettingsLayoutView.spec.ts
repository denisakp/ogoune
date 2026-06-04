import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'

const routePath: { value: string } = { value: '/settings/account' }
vi.mock('vue-router', () => ({
  useRoute: () => ({
    get path() {
      return routePath.value
    },
    params: {},
    name: 'SettingsAccount',
  }),
  RouterLink: { template: '<a :href="to" class="rl"><slot /></a>', props: ['to'] },
  RouterView: { template: '<div data-test="router-view"></div>' },
}))

import SettingsLayoutView from '../SettingsLayoutView.vue'

function build() {
  return mount(SettingsLayoutView, {
    global: {
      stubs: {
        UIcon: { template: '<span class="u-icon" :data-name="name"></span>', props: ['name'] },
      },
    },
  })
}

describe('SettingsLayoutView', () => {
  it('renders the 3 section labels — PROFILE / SECURITY / ORGANIZATION', () => {
    const w = build()
    const labels = w.findAll('.uppercase').map((n) => n.text())
    expect(labels).toContain('PROFILE')
    expect(labels).toContain('SECURITY')
    expect(labels).toContain('ORGANIZATION')
  })

  it('lists the 5 in-layout entries (Notifications/Escalation/API Keys are top-level)', () => {
    const w = build()
    const labels = w.findAll('.rl').map((n) => n.text().trim())
    for (const expected of ['Account', 'Two-Factor Auth', 'Sessions', 'General', 'Status Page']) {
      expect(labels.some((t) => t.includes(expected))).toBe(true)
    }
  })

  it('does NOT expose Notifications / Escalation / API Keys / Tags / Maintenance inside the sub-nav', () => {
    const w = build()
    const hrefs = w.findAll('.rl').map((a) => a.attributes('href') ?? '')
    expect(hrefs.some((h) => h.includes('/settings/notifications'))).toBe(false)
    expect(hrefs.some((h) => h.includes('/settings/escalation'))).toBe(false)
    expect(hrefs.some((h) => h.includes('/settings/api-keys'))).toBe(false)
    expect(hrefs.some((h) => h.includes('/settings/tags'))).toBe(false)
    expect(hrefs.some((h) => h.includes('/maintenance'))).toBe(false)
  })

  it('renders <RouterView/> for the active child route', () => {
    const w = build()
    expect(w.find('[data-test="router-view"]').exists()).toBe(true)
  })

  it('highlights the active sub-nav entry from the current route', () => {
    routePath.value = '/settings/org/status-page'
    const w = build()
    const active = w
      .findAll('a.rl')
      .filter((a) => a.attributes('href') === '/settings/org/status-page')
    expect(active.length).toBe(1)
    expect(active[0]!.classes().join(' ')).toContain('bg-elevated')
  })
})
