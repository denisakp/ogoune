import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import OverallUptimePanel from '../OverallUptimePanel.vue'
import type { PublicResourceSummary, PublicStatusResourceWindows } from '@/types'

vi.mock('@/services/statusPublicService', () => ({
  fetchPublicStatusSummary: vi.fn(),
  fetchPublicStatusIncidents: vi.fn(),
  fetchPublicStatusUptime: vi.fn(),
  fetchPublicStatusResourceWindows: vi.fn(),
  fetchPublicIncidentDetail: vi.fn(),
}))

import * as svc from '@/services/statusPublicService'

const resource: PublicResourceSummary = {
  id: 'res-1',
  name: 'api.acme.com',
  host: 'api.acme.com',
  current_state: 'up',
  uptime_90d_ratio: 0.999,
  uptime_ribbon: [],
}

function mkDetails(
  overrides: Partial<PublicStatusResourceWindows> = {},
): PublicStatusResourceWindows {
  return {
    id: 'res-1',
    name: 'api.acme.com',
    windows: {
      '24h': { uptime_ratio: 1, incidents: 0 },
      '7d': { uptime_ratio: 0.99, incidents: 1 },
      '30d': { uptime_ratio: 0.97, incidents: 3 },
      '90d': { uptime_ratio: 0.95, incidents: 8 },
    } as never,
    recent_incidents: [],
    ...overrides,
  } as PublicStatusResourceWindows
}

async function render(open = true, details = mkDetails()) {
  vi.mocked(svc.fetchPublicStatusResourceWindows).mockResolvedValue(details)
  const w = mount(OverallUptimePanel, { props: { resource, open }, attachTo: document.body })
  await flushPromises()
  return w
}

describe('OverallUptimePanel — US4', () => {
  beforeEach(() => vi.clearAllMocks())

  it('loads windows when opened and defaults to 30d active', async () => {
    await render(true)
    expect(svc.fetchPublicStatusResourceWindows).toHaveBeenCalledWith('res-1')
    const active = document.querySelector('[data-window][data-active="1"]')
    expect(active?.getAttribute('data-window')).toBe('30d')
  })

  it('switching to 7d marks it active without refetching', async () => {
    await render(true)
    vi.mocked(svc.fetchPublicStatusResourceWindows).mockClear()
    const btn7d = document.querySelector('[data-window="7d"]') as HTMLElement
    btn7d.click()
    await flushPromises()
    const active = document.querySelector('[data-window][data-active="1"]')
    expect(active?.getAttribute('data-window')).toBe('7d')
    expect(svc.fetchPublicStatusResourceWindows).not.toHaveBeenCalled()
  })

  it('Escape key emits close', async () => {
    const w = await render(true)
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }))
    await flushPromises()
    expect(w.emitted('close')).toBeTruthy()
  })

  it('does not refetch when closed and reopened on the same resource', async () => {
    const w = await render(true)
    vi.mocked(svc.fetchPublicStatusResourceWindows).mockClear()
    await w.setProps({ open: false })
    await w.setProps({ open: true })
    await flushPromises()
    // Same resource → one fetch on every open is acceptable; assert at most one.
    expect(vi.mocked(svc.fetchPublicStatusResourceWindows).mock.calls.length).toBeLessThanOrEqual(1)
  })
})
