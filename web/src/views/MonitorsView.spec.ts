import { describe, expect, it, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
const pushMock = vi.fn()
const replaceMock = vi.fn()
const routeQuery = { value: {} as Record<string, string | undefined> }
vi.mock('vue-router', () => ({
  useRouter: () => ({ push: pushMock, replace: replaceMock }),
  useRoute: () => ({ get query() { return routeQuery.value } }),
}))

const useConfirmMock = vi.fn()
vi.mock('@/composables/useConfirm', () => ({
  useConfirm: (opts: unknown) => useConfirmMock(opts),
}))

const rows = [
  {
    id: 'r1',
    name: 'api',
    type: 'http',
    status: 'up',
    url: 'https://api.test',
    last_checked: new Date().toISOString(),
  },
  {
    id: 'r2',
    name: 'db',
    type: 'tcp',
    status: 'down',
    host: 'db.test',
    port: 5432,
    last_checked: null,
  },
]
const removeMock = vi.fn().mockResolvedValue(true)
const loadResourcesMock = vi.fn().mockResolvedValue(undefined)

vi.mock('@/stores/resourceStore', () => ({
  useResourceStore: () => ({
    resources: rows,
    loading: false,
    error: null,
    loadResources: loadResourcesMock,
    removeResource: removeMock,
  }),
}))

vi.mock('@/stores/componentStore', () => ({
  useComponentStore: () => ({
    loadComponents: vi.fn().mockResolvedValue(undefined),
  }),
}))

vi.mock('@/components/resources/ResourceModal.vue', () => ({
  default: { template: '<div />' },
}))

import MonitorsView from './MonitorsView.vue'

const stubs = {
  UButton: { template: '<button><slot /></button>' },
  UAlert: { template: '<div />' },
  UStatusBadge: { template: '<span><slot /></span>' },
  UDataTable: {
    name: 'UDataTable',
    template: '<div data-testid="dt" />',
    props: ['columns', 'rows', 'loading', 'filters', 'pagination', 'rowActions'],
  },
}

function build() {
  setActivePinia(createPinia())
  return mount(MonitorsView, { global: { stubs, mocks: { $route: { query: routeQuery.value } } } })
}

beforeEach(() => {
  pushMock.mockReset()
  replaceMock.mockReset()
  useConfirmMock.mockReset()
  removeMock.mockClear()
  loadResourcesMock.mockClear()
  routeQuery.value = {}
})

describe('MonitorsView', () => {
  it('mounts and passes 6 columns to UDataTable', () => {
    const w = build()
    const dt = w.findComponent({ name: 'UDataTable' })
    const cols = (dt.props('columns') as Array<{ key: string }>).map((c) => c.key)
    expect(cols).toEqual(['status', 'name', 'target', 'uptime', 'last_checked', 'actions'])
  })

  it('removeFilter clears the status filter and triggers router.replace (URL sync)', async () => {
    const w = build()
    ;(w.vm as unknown as { filterStatus: { value: string[] } }).filterStatus.value = ['down']
    await w.vm.$nextTick()
    ;(
      w.vm as unknown as {
        removeFilter: (f: { kind: string; value: string }) => void
      }
    ).removeFilter({ kind: 'status', value: 'down' })
    await w.vm.$nextTick()
    expect(replaceMock).toHaveBeenCalled()
    const lastCall = replaceMock.mock.calls.at(-1)?.[0] as { query: Record<string, unknown> }
    expect(lastCall.query.status).toBeUndefined()
  })

  it('Delete row action calls useConfirm with destructive kind and removeResource on confirm', async () => {
    useConfirmMock.mockResolvedValueOnce(true)
    const w = build()
    await (
      w.vm as unknown as {
        onAction: (p: { action: { label: string }; row: unknown }) => Promise<void>
      }
    ).onAction({ action: { label: 'Delete' }, row: rows[0] })
    expect(useConfirmMock).toHaveBeenCalledWith(
      expect.objectContaining({ kind: 'destructive', ctaLabel: 'Delete' }),
    )
    expect(removeMock).toHaveBeenCalledWith('r1')
  })

  it('Delete cancel does NOT call removeResource', async () => {
    useConfirmMock.mockResolvedValueOnce(false)
    const w = build()
    await (
      w.vm as unknown as {
        onAction: (p: { action: { label: string }; row: unknown }) => Promise<void>
      }
    ).onAction({ action: { label: 'Delete' }, row: rows[0] })
    expect(removeMock).not.toHaveBeenCalled()
  })

  it('View row action pushes to ResourceDetail', async () => {
    const w = build()
    await (
      w.vm as unknown as {
        onAction: (p: { action: { label: string }; row: unknown }) => Promise<void>
      }
    ).onAction({ action: { label: 'View' }, row: rows[0] })
    expect(pushMock).toHaveBeenCalledWith({ name: 'ResourceDetail', params: { id: 'r1' } })
  })

  it('dark-mode artifact check: root carries bg-default token (FR-017)', () => {
    document.documentElement.classList.add('dark')
    const w = build()
    expect(w.find('.bg-default').exists()).toBe(true)
    document.documentElement.classList.remove('dark')
  })
})
