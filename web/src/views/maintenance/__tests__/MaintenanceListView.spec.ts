/* eslint-disable @typescript-eslint/ban-ts-comment */
// @ts-nocheck — spec 059 polish debt: index-access narrowing
import { describe, expect, it, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'

const { fetchMock, createMock } = vi.hoisted(() => ({
  fetchMock: vi.fn(),
  createMock: vi.fn(),
}))

vi.mock('@/services/maintenanceService', () => ({
  fetchMaintenances: fetchMock,
  createMaintenance: createMock,
}))

vi.mock('@/components/maintenance/MaintenanceModal.vue', () => ({
  default: { name: 'MaintenanceModal', template: '<div data-testid="modal" />', props: ['open'] },
}))

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: vi.fn(), replace: vi.fn(), resolve: () => ({ href: '#' }) }),
  useRoute: () => ({ path: '/maintenance', params: {}, query: {}, name: 'Maintenance' }),
  useLink: () => ({ href: { value: '#' }, navigate: vi.fn(), isActive: { value: false } }),
  RouterLink: { template: '<a><slot /></a>' },
}))

import MaintenanceListView from '../MaintenanceListView.vue'
import type { Maintenance } from '@/types'

type Vm = {
  maintenances: Maintenance[]
  stats: { key: string; label: string; value: string; meta: string; icon: string; tint: string }[]
  filtered: Maintenance[]
  preset: 'all' | 'active' | 'scheduled' | 'finished'
  strategyFilter: 'all' | 'one_time' | 'cron'
  search: string
  load: () => Promise<void>
  onSubmit: (p: { title: string; strategy: string; resource_ids: string[] }) => Promise<void>
}

const mkOneTime = (overrides: Partial<Maintenance>): Maintenance => ({
  id: overrides.id ?? 'm1',
  title: overrides.title ?? 'DB Migration',
  strategy: 'one_time',
  status: overrides.status ?? 'scheduled',
  start_at: overrides.start_at ?? new Date(Date.now() + 86_400_000).toISOString(),
  end_at: overrides.end_at ?? new Date(Date.now() + 90_000_000).toISOString(),
  updated_at: overrides.updated_at,
})

beforeEach(() => {
  fetchMock.mockReset()
  createMock.mockReset()
})

describe('MaintenanceListView', () => {
  it('stats row aggregates active / upcoming / recurring / completed with meta text', async () => {
    fetchMock.mockResolvedValue([
      mkOneTime({ id: 'a', status: 'active', title: 'API Rate Limiter Update' }),
      mkOneTime({
        id: 'b',
        status: 'scheduled',
        title: 'SSL Renewal',
        start_at: new Date(Date.now() + 2 * 86_400_000).toISOString(),
      }),
      {
        id: 'c',
        title: 'Weekly Cache Purge',
        strategy: 'cron',
        status: 'scheduled',
        cron_expr: '0 4 * * 0',
        window_minutes: 30,
      },
      {
        id: 'd',
        title: 'API Rate Limiter Old',
        strategy: 'one_time',
        status: 'finished',
        updated_at: new Date(Date.now() - 86_400_000).toISOString(),
      },
    ] as Maintenance[])
    const w = mount(MaintenanceListView)
    const vm = w.vm as unknown as Vm
    await vm.load()
    await flushPromises()
    expect(vm.stats.map((s) => s.value)).toEqual(['1', '1', '1', '1'])
    expect(vm.stats[0].label).toBe('ACTIVE NOW')
    expect(vm.stats[0].meta).toContain('API Rate Limiter Update')
    expect(vm.stats[2].meta).toContain('weekly / monthly')
  })

  it('empty state CTA renders when list is empty', async () => {
    fetchMock.mockResolvedValue([])
    const w = mount(MaintenanceListView)
    const vm = w.vm as unknown as Vm
    await vm.load()
    await flushPromises()
    expect(w.text()).toContain('No maintenance windows yet')
  })

  it('preset filter narrows the rendered list to status=active', async () => {
    fetchMock.mockResolvedValue([
      mkOneTime({ id: 'a', status: 'active', title: 'A' }),
      mkOneTime({ id: 'b', status: 'scheduled', title: 'B' }),
    ])
    const w = mount(MaintenanceListView)
    const vm = w.vm as unknown as Vm
    await vm.load()
    await flushPromises()
    expect(vm.filtered.length).toBe(2)
    vm.preset = 'active'
    await flushPromises()
    expect(vm.filtered.map((m) => m.id)).toEqual(['a'])
  })

  it('search filter matches title substring', async () => {
    fetchMock.mockResolvedValue([
      mkOneTime({ id: 'a', title: 'Database Migration' }),
      mkOneTime({ id: 'b', title: 'SSL Renewal' }),
    ])
    const w = mount(MaintenanceListView)
    const vm = w.vm as unknown as Vm
    await vm.load()
    await flushPromises()
    vm.search = 'ssl'
    await flushPromises()
    expect(vm.filtered.map((m) => m.id)).toEqual(['b'])
  })

  it('onSubmit calls createMaintenance and reloads list', async () => {
    fetchMock.mockResolvedValue([])
    createMock.mockResolvedValue(mkOneTime({ id: 'new' }))
    const w = mount(MaintenanceListView)
    const vm = w.vm as unknown as Vm
    await vm.load()
    await flushPromises()
    await vm.onSubmit({ title: 'X', strategy: 'one_time', resource_ids: [] })
    expect(createMock).toHaveBeenCalled()
    expect(fetchMock.mock.calls.length).toBeGreaterThanOrEqual(2)
  })
})
