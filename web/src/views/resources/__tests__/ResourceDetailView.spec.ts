import { describe, expect, it, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'

const pushMock = vi.fn()
vi.mock('vue-router', () => ({
  useRouter: () => ({ push: pushMock, replace: vi.fn(), resolve: () => ({ href: '#' }) }),
  useRoute: () => ({
    query: {},
    params: { id: 'r1' },
    path: '/resources/r1',
    name: 'ResourceDetail',
  }),
  useLink: () => ({ href: { value: '#' }, navigate: vi.fn(), isActive: { value: false } }),
  RouterLink: { template: '<a><slot /></a>' },
}))

const useConfirmMock = vi.fn()
vi.mock('@/composables/useConfirm', () => ({
  useConfirm: (opts: unknown) => useConfirmMock(opts),
}))

const resourceData = {
  id: 'r1',
  name: 'api.acme.com',
  type: 'http',
  status: 'up',
  url: 'https://api.acme.com/health',
  interval: 60,
  last_checked: new Date().toISOString(),
}

const loadResourceMock = vi.fn().mockResolvedValue(resourceData)
const removeMock = vi.fn().mockResolvedValue(true)
const pauseMock = vi.fn().mockResolvedValue(true)
const resumeMock = vi.fn().mockResolvedValue(true)

vi.mock('@/stores/resourceStore', () => ({
  useResourceStore: () => ({
    loadResource: loadResourceMock,
    removeResource: removeMock,
    pauseMonitoring: pauseMock,
    resumeMonitoring: resumeMock,
  }),
}))

const fetchActivitiesMock = vi.fn().mockResolvedValue([])
vi.mock('@/services/activityService', () => ({
  fetchActivities: () => fetchActivitiesMock(),
}))

vi.mock('@/components/resources/ResourceModal.vue', () => ({
  default: { template: '<div />' },
}))
vi.mock('@/components/ResponseTimeChart.vue', () => ({
  default: { template: '<div />' },
}))

import ResourceDetailView from '../ResourceDetailView.vue'

const stubs = {
  UButton: { template: '<button><slot /></button>' },
  UIcon: { template: '<span />' },
}

function build() {
  setActivePinia(createPinia())
  return mount(ResourceDetailView, { global: { stubs } })
}

beforeEach(() => {
  pushMock.mockReset()
  useConfirmMock.mockReset()
  loadResourceMock.mockClear()
  removeMock.mockClear()
  pauseMock.mockClear()
  resumeMock.mockClear()
  fetchActivitiesMock.mockClear()
})

describe('ResourceDetailView', () => {
  it('mount triggers loadResource(id) + renders header', async () => {
    const w = build()
    await flushPromises()
    expect(loadResourceMock).toHaveBeenCalled()
    expect(w.text()).toContain('api.acme.com')
  })

  it('default activeTab is overview', async () => {
    const w = build()
    await flushPromises()
    expect((w.vm as unknown as { activeTab: string }).activeTab).toBe('overview')
  })

  it('switching to activity tab fetches + shows activity log', async () => {
    const w = build()
    await flushPromises()
    ;(w.vm as unknown as { activeTab: string }).activeTab = 'activity'
    await w.vm.$nextTick()
    expect(w.text()).toContain('Activity log')
  })

  it('incidents tab shows PRD 006 placeholder', async () => {
    const w = build()
    await flushPromises()
    ;(w.vm as unknown as { activeTab: string }).activeTab = 'incidents'
    await w.vm.$nextTick()
    expect(w.text()).toContain('Incidents coming with PRD 006')
  })

  it('Delete action confirms via useConfirm then calls removeResource + navigates back', async () => {
    useConfirmMock.mockResolvedValueOnce(true)
    const w = build()
    await flushPromises()
    await (w.vm as unknown as { onDelete: () => Promise<void> }).onDelete()
    expect(useConfirmMock).toHaveBeenCalledWith(expect.objectContaining({ kind: 'destructive' }))
    expect(removeMock).toHaveBeenCalledWith('r1')
    expect(pushMock).toHaveBeenCalledWith('/resources')
  })

  it('togglePause calls pauseMonitoring when status not paused', async () => {
    const w = build()
    await flushPromises()
    await (w.vm as unknown as { togglePause: () => Promise<void> }).togglePause()
    expect(pauseMock).toHaveBeenCalledWith('r1')
  })

  it('dark-mode artifact check: root carries bg-default token (FR-022)', async () => {
    document.documentElement.classList.add('dark')
    const w = build()
    await flushPromises()
    expect(w.find('.bg-default').exists()).toBe(true)
    document.documentElement.classList.remove('dark')
  })
})
