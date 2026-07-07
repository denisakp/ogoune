import { describe, expect, it, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'

const pushMock = vi.fn()
const replaceMock = vi.fn()
const routeQuery: { value: Record<string, string | undefined> } = { value: {} }
vi.mock('vue-router', () => ({
  useRouter: () => ({ push: pushMock, replace: replaceMock, resolve: () => ({ href: '#' }) }),
  useRoute: () => ({
    get query() {
      return routeQuery.value
    },
    params: {},
    path: '/incidents',
    name: 'incidents',
  }),
  useLink: () => ({ href: { value: '#' }, navigate: vi.fn(), isActive: { value: false } }),
  RouterLink: { template: '<a><slot /></a>' },
}))

const now = Date.now()
const day = 24 * 3_600_000
const incidents = [
  {
    id: 'i1',
    resource_id: 'r1',
    resource: { id: 'r1', name: 'api', type: 'http' },
    reason: 'r',
    cause: 'HTTP 500',
    started_at: new Date(now - day / 2).toISOString(),
    resolved_at: null,
    created_at: '',
    updated_at: '',
  },
  {
    id: 'i2',
    resource_id: 'r2',
    resource: { id: 'r2', name: 'db', type: 'tcp' },
    reason: 'r',
    cause: 'connection refused',
    started_at: new Date(now - day).toISOString(),
    resolved_at: new Date(now - day + 600_000).toISOString(),
    created_at: '',
    updated_at: '',
  },
]
const fetchMock = vi.fn().mockResolvedValue(undefined)

vi.mock('@/stores/incidentStore', () => ({
  useIncidentStore: () => ({
    get incidents() {
      return incidents
    },
    fetchIncidents: fetchMock,
  }),
}))

vi.mock('@/components/incidents/IncidentStatsRow.vue', () => ({
  default: {
    name: 'IncidentStatsRow',
    template: '<div data-testid="stats" />',
    props: ['incidents'],
  },
}))

import IncidentsView from '../IncidentsView.vue'

const stubs = {
  UInput: { template: '<input />' },
  USelectMenu: { template: '<select />' },
  UTabs: { template: '<div />', props: ['items', 'modelValue'] },
  UFilterChip: { template: '<span />' },
  UIcon: { template: '<span />' },
  UEmpty: { template: '<div />' },
  IncidentStatsRow: { template: '<div data-testid="stats" />' },
}

function build() {
  setActivePinia(createPinia())
  return mount(IncidentsView, { global: { stubs } })
}

beforeEach(() => {
  pushMock.mockReset()
  replaceMock.mockReset()
  fetchMock.mockClear()
  routeQuery.value = {}
})

describe('IncidentsView', () => {
  it('default preset shows all incidents', async () => {
    const w = build()
    await w.vm.$nextTick()
    const vm = w.vm as unknown as { filtered: Array<{ id: string }> }
    expect(vm.filtered.length).toBe(2)
  })

  it('preset Active filters to incidents with resolved_at == null', async () => {
    routeQuery.value = { preset: 'active' }
    const w = build()
    await w.vm.$nextTick()
    const vm = w.vm as unknown as { filtered: Array<{ id: string }> }
    expect(vm.filtered.length).toBe(1)
    expect(vm.filtered[0]?.id).toBe('i1')
  })

  it('preset Resolved filters to resolved incidents only', async () => {
    routeQuery.value = { preset: 'resolved' }
    const w = build()
    await w.vm.$nextTick()
    const vm = w.vm as unknown as { filtered: Array<{ id: string }> }
    expect(vm.filtered.length).toBe(1)
    expect(vm.filtered[0]?.id).toBe('i2')
  })

  it('search filters by cause / resource name', async () => {
    const w = build()
    await w.vm.$nextTick()
    const vm = w.vm as unknown as {
      filters: { search: { value: string } }
      filtered: Array<{ id: string }>
    }
    vm.filters.search.value = 'connection'
    await w.vm.$nextTick()
    expect(vm.filtered.length).toBe(1)
    expect(vm.filtered[0]?.id).toBe('i2')
  })

  it('Clear all resets URL', async () => {
    routeQuery.value = { preset: 'active', type: 'http' }
    const w = build()
    await w.vm.$nextTick()
    ;(w.vm as unknown as { filters: { clear: () => void } }).filters.clear()
    await w.vm.$nextTick()
    const last = replaceMock.mock.calls.at(-1)?.[0] as { query: Record<string, string> }
    expect(last.query).toEqual({})
  })

  it('dark-mode artifact check: root carries bg-default (FR-020)', () => {
    document.documentElement.classList.add('dark')
    const w = build()
    expect(w.find('.bg-default').exists()).toBe(true)
    document.documentElement.classList.remove('dark')
  })
})
