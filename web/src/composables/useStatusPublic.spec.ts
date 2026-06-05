import { describe, it, expect, vi, beforeEach } from 'vitest'
import { useStatusPublic } from './useStatusPublic'
import type { PublicStatusSummary, PublicStatusUptimeRange } from '@/types'

vi.mock('@/services/statusPublicService', () => ({
  fetchPublicStatusSummary: vi.fn(),
  fetchPublicStatusIncidents: vi.fn(),
  fetchPublicStatusUptime: vi.fn(),
  fetchPublicStatusResourceWindows: vi.fn(),
}))

import * as svc from '@/services/statusPublicService'

const summaryFixture: PublicStatusSummary = {
  generated_at: '2026-06-04T17:42:11Z',
  verdict: { status: 'operational', label: 'All Systems Operational', color: 'green' },
  components: [],
  standalone_resources: [],
  current_month_incidents: [],
}

const uptimeFixture: PublicStatusUptimeRange = {
  generated_at: '2026-06-04T17:42:11Z',
  days: [{ day: '2026-06-04', uptime_ratio: 0.9999, samples: 1440, incidents: 0 }],
}

describe('useStatusPublic', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('loadSummary populates summary on success', async () => {
    vi.mocked(svc.fetchPublicStatusSummary).mockResolvedValue(summaryFixture)
    const s = useStatusPublic()
    const data = await s.loadSummary()
    expect(data).toEqual(summaryFixture)
    expect(s.summary.value).toEqual(summaryFixture)
    expect(s.loading.value).toBe(false)
    expect(s.error.value).toBeNull()
  })

  it('captures error and leaves summary null', async () => {
    vi.mocked(svc.fetchPublicStatusSummary).mockRejectedValue(new Error('boom'))
    const s = useStatusPublic()
    const data = await s.loadSummary()
    expect(data).toBeNull()
    expect(s.summary.value).toBeNull()
    expect(s.error.value?.message).toBe('boom')
  })

  it('forwards uptime query params', async () => {
    vi.mocked(svc.fetchPublicStatusUptime).mockResolvedValue(uptimeFixture)
    const s = useStatusPublic()
    await s.loadUptime({ from: '2026-05-01', to: '2026-06-04', component_id: 'comp_x' })
    expect(svc.fetchPublicStatusUptime).toHaveBeenCalledWith({
      from: '2026-05-01',
      to: '2026-06-04',
      component_id: 'comp_x',
    })
    expect(s.uptime.value).toEqual(uptimeFixture)
  })

  it('sets loading true while in-flight', async () => {
    let resolveFn!: (v: PublicStatusSummary) => void
    vi.mocked(svc.fetchPublicStatusSummary).mockReturnValue(
      new Promise((res) => {
        resolveFn = res
      }),
    )
    const s = useStatusPublic()
    const p = s.loadSummary()
    expect(s.loading.value).toBe(true)
    resolveFn(summaryFixture)
    await p
    expect(s.loading.value).toBe(false)
  })
})
