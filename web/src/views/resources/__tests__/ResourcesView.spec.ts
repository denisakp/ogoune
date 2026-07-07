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
    path: '/resources',
    name: 'resources',
  }),
  useLink: () => ({ href: { value: '#' }, navigate: vi.fn(), isActive: { value: false } }),
  RouterLink: { template: '<a><slot /></a>' },
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
    url: 'https://x',
    last_checked: new Date().toISOString(),
    component_id: null,
  },
  {
    id: 'r2',
    name: 'db',
    type: 'tcp',
    status: 'down',
    host: 'db',
    port: 5432,
    last_checked: null,
    component_id: 'c1',
  },
  {
    id: 'r3',
    name: 'web',
    type: 'http',
    status: 'up',
    url: 'https://w',
    last_checked: null,
    component_id: 'c1',
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
    pauseMonitoring: vi.fn().mockResolvedValue(true),
    resumeMonitoring: vi.fn().mockResolvedValue(true),
  }),
}))

vi.mock('@/stores/componentStore', () => ({
  useComponentStore: () => ({
    components: [{ id: 'c1', name: 'API Cluster' }],
    loadComponents: vi.fn().mockResolvedValue(undefined),
  }),
}))

vi.mock('@/components/resources/ResourceModal.vue', () => ({
  default: { template: '<div />' },
}))

import ResourcesView from '../ResourcesView.vue'

const stubs = {
  UInput: { template: '<input />' },
  UButton: { template: '<button><slot /></button>' },
  USelectMenu: { template: '<select />' },
  UTabs: { template: '<div data-testid="tabs" />', props: ['items', 'modelValue'] },
  UFilterChip: { template: '<span><slot /></span>', props: ['kind', 'value'] },
  UIcon: { template: '<span />' },
  UDropdownMenu: { template: '<div><slot /></div>' },
  ComponentGroupHeader: {
    name: 'ComponentGroupHeader',
    template: '<div data-testid="group-header" />',
    props: ['component', 'resources', 'collapsed'],
  },
  ResourceListItem: {
    name: 'ResourceListItem',
    template: '<div data-testid="row" />',
    props: ['resource'],
  },
}

function build() {
  setActivePinia(createPinia())
  return mount(ResourcesView, { global: { stubs } })
}

beforeEach(() => {
  pushMock.mockReset()
  replaceMock.mockReset()
  useConfirmMock.mockReset()
  removeMock.mockClear()
  loadResourcesMock.mockClear()
  routeQuery.value = {}
})

describe('ResourcesView', () => {
  it('default mode renders ComponentGroupHeader(s) + Standalone Resources group', async () => {
    const w = build()
    await w.vm.$nextTick()
    const headers = w.findAllComponents({ name: 'ComponentGroupHeader' })
    expect(headers.length).toBe(2) // API Cluster + Standalone Resources
  })

  it('flat mode renders flat rows, no group header', async () => {
    routeQuery.value = { view: 'flat' }
    const w = build()
    await w.vm.$nextTick()
    expect(w.findAllComponents({ name: 'ComponentGroupHeader' }).length).toBe(0)
    expect(w.findAllComponents({ name: 'ResourceListItem' }).length).toBe(3)
  })

  it('typing in Search updates URL ?search=', async () => {
    const w = build()
    ;(w.vm as unknown as { filters: { search: { value: string } } }).filters.search.value = 'ap'
    await w.vm.$nextTick()
    const last = replaceMock.mock.calls.at(-1)?.[0] as { query: Record<string, string> }
    expect(last.query.search).toBe('ap')
  })

  it('removeChip removes chip from URL', async () => {
    routeQuery.value = { status: 'down' }
    const w = build()
    await w.vm.$nextTick()
    const f = (
      w.vm as unknown as { filters: { removeChip: (c: { kind: string; value: string }) => void } }
    ).filters
    f.removeChip({ kind: 'status', value: 'down' })
    await w.vm.$nextTick()
    const last = replaceMock.mock.calls.at(-1)?.[0] as { query: Record<string, string> }
    expect(last.query.status).toBeUndefined()
  })

  it('Clear all resets URL', async () => {
    routeQuery.value = { status: 'down', type: 'http' }
    const w = build()
    await w.vm.$nextTick()
    ;(w.vm as unknown as { filters: { clear: () => void } }).filters.clear()
    await w.vm.$nextTick()
    const last = replaceMock.mock.calls.at(-1)?.[0] as { query: Record<string, string> }
    expect(last.query).toEqual({})
  })

  it('Delete row action calls useConfirm destructive (kind+title+body w/ resource name) then removeResource [FR-021, US5]', async () => {
    useConfirmMock.mockResolvedValueOnce(true)
    const w = build()
    await (
      w.vm as unknown as {
        onAction: (p: { kind: string; resource: { id: string; name: string } }) => Promise<void>
      }
    ).onAction({ kind: 'delete', resource: rows[0]! })
    expect(useConfirmMock).toHaveBeenCalledWith(
      expect.objectContaining({
        kind: 'destructive',
        title: 'Delete monitor?',
        ctaLabel: 'Delete',
      }),
    )
    const opts = useConfirmMock.mock.calls.at(-1)?.[0] as { body: string }
    expect(opts.body).toContain('api') // resource name appears in body
    expect(removeMock).toHaveBeenCalledWith('r1')
  })

  it('dark-mode artifact check: root carries bg-default token (FR-022)', () => {
    document.documentElement.classList.add('dark')
    const w = build()
    expect(w.find('.bg-default').exists()).toBe(true)
    document.documentElement.classList.remove('dark')
  })
})
