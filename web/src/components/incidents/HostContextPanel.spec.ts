import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import HostContextPanel from './HostContextPanel.vue'
import type { HostContext } from '@/types'

const stubs = {
  UIcon: { template: '<span />' },
  RouterLink: { template: '<a><slot /></a>', props: ['to'] },
}

const ctx = (overrides: Partial<HostContext> = {}): HostContext => ({
  host_id: 'host-1',
  host_name: 'web-01',
  peak_cpu_pct: 97.4,
  peak_mem_pct: 99.1,
  worst_disk: { mount: '/var', used_pct: 96.2 },
  sample_count: 36,
  resolution: 'full',
  window_from: '2026-03-10T11:53:00Z',
  window_to: '2026-03-10T11:59:00Z',
  ...overrides,
})

function build(context: HostContext) {
  return mount(HostContextPanel, { props: { context }, global: { stubs } })
}

describe('HostContextPanel', () => {
  // T018 — the populated case renders every figure.
  it('renders the peaks, the busiest mount and the host name', () => {
    const w = build(ctx())
    expect(w.get('[data-test="host-context-cpu"]').text()).toContain('97.4%')
    expect(w.get('[data-test="host-context-mem"]').text()).toContain('99.1%')
    expect(w.get('[data-test="host-context-disk"]').text()).toContain('/var')
    expect(w.get('[data-test="host-context-disk"]').text()).toContain('96.2%')
    expect(w.get('[data-test="host-context-link"]').text()).toContain('web-01')
  })

  // T018 — a null worst_disk hides only the disk row. One missing signal must not
  // suppress the others.
  it('hides only the disk row when the host reported no mounts', () => {
    const w = build(ctx({ worst_disk: null }))
    expect(w.find('[data-test="host-context-disk"]').exists()).toBe(false)
    expect(w.get('[data-test="host-context-cpu"]').text()).toContain('97.4%')
    expect(w.get('[data-test="host-context-mem"]').text()).toContain('99.1%')
  })

  // T018 — the sample count is always visible, so a thin aggregate is never read
  // as a full one.
  it('always surfaces the sample count', () => {
    expect(build(ctx({ sample_count: 2 })).get('[data-test="host-context-samples"]').text()).toContain(
      '2 samples',
    )
    expect(build(ctx({ sample_count: 1 })).get('[data-test="host-context-samples"]').text()).toContain(
      '1 sample,',
    )
  })

  // T031 — a reduced context is presented as minute-level, never as exact peaks.
  it('presents a reduced context as approximate', () => {
    const w = build(ctx({ resolution: 'reduced' }))
    expect(w.find('[data-test="host-context-reduced"]').exists()).toBe(true)
    expect(w.find('[data-test="host-context-full"]').exists()).toBe(false)
    expect(w.get('[data-test="host-context-reduced"]').text()).toContain('Minute-level')
  })

  // T031 — a full context is not hedged.
  it('presents a full context as exact', () => {
    const w = build(ctx({ resolution: 'full' }))
    expect(w.find('[data-test="host-context-full"]').exists()).toBe(true)
    expect(w.find('[data-test="host-context-reduced"]').exists()).toBe(false)
  })

  // T031 — an unknown marker from a newer backend must read as reduced, the
  // conservative value, rather than being presented as an exact peak.
  it('treats an unknown resolution as reduced', () => {
    const w = build(ctx({ resolution: 'something-new' as HostContext['resolution'] }))
    expect(w.find('[data-test="host-context-reduced"]').exists()).toBe(true)
  })
})
