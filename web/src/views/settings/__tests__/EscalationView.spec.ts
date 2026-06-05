/* eslint-disable @typescript-eslint/ban-ts-comment */
// @ts-nocheck — spec 059 polish debt: index-access narrowing
import { describe, expect, it, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'

const listMock = vi.fn()
const createMock = vi.fn()
const updateMock = vi.fn()
const deleteMock = vi.fn()
const reorderMock = vi.fn()
vi.mock('@/services/escalationService', () => ({
  default: {
    list: (...a: unknown[]) => listMock(...a),
    create: (...a: unknown[]) => createMock(...a),
    update: (...a: unknown[]) => updateMock(...a),
    delete: (...a: unknown[]) => deleteMock(...a),
    reorder: (...a: unknown[]) => reorderMock(...a),
  },
}))

const fetchChannelsMock = vi.fn()
vi.mock('@/services/notificationChannelService', () => ({
  fetchChannels: (...a: unknown[]) => fetchChannelsMock(...a),
}))

const confirmMock = vi.fn()
vi.mock('@/composables/useConfirm', () => ({
  useConfirm: (opts: unknown) => confirmMock(opts),
}))

vi.mock('@/components/settings/escalation/PolicyCard.vue', () => ({
  default: {
    name: 'PolicyCard',
    template: '<div :data-testid="`card-${policy.id}`">{{ policy.name }}</div>',
    props: ['policy', 'canMoveUp', 'canMoveDown'],
  },
}))

vi.mock('@/components/settings/escalation/PolicyModal.vue', () => ({
  default: {
    name: 'PolicyModal',
    template: '<div data-testid="modal" />',
    props: ['open', 'initial', 'channels'],
  },
}))

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: vi.fn(), replace: vi.fn(), resolve: () => ({ href: '#' }) }),
  useRoute: () => ({
    path: '/settings/escalation',
    params: {},
    query: {},
    name: 'SettingsEscalation',
  }),
  useLink: () => ({ href: { value: '#' }, navigate: vi.fn(), isActive: { value: false } }),
  RouterLink: { template: '<a><slot /></a>' },
}))

import EscalationView from '../EscalationView.vue'
import type { EscalationPolicy } from '@/services/escalationService'

type Vm = {
  policies: EscalationPolicy[]
  stats: { label: string; value: string; tip: string }[]
  openCreate: () => void
  onSubmit: (p: {
    name: string
    scope: { kind: 'component'; value: string }
    is_active: boolean
    steps: { delay_minutes: number; channel_ids: string[] }[]
  }) => Promise<void>
  onDelete: (p: EscalationPolicy) => Promise<void>
  onToggle: (p: EscalationPolicy) => Promise<void>
  moveUp: (p: EscalationPolicy) => void
  moveDown: (p: EscalationPolicy) => void
}

const mkPolicy = (overrides: Partial<EscalationPolicy>): EscalationPolicy => ({
  id: overrides.id ?? 'p',
  name: overrides.name ?? 'X',
  scope: overrides.scope ?? { kind: 'component', value: 'c' },
  is_active: overrides.is_active ?? true,
  priority: overrides.priority ?? 1,
  steps: overrides.steps ?? [{ delay_minutes: 5, channel_ids: ['ch'] }],
})

beforeEach(() => {
  listMock.mockReset()
  createMock.mockReset()
  updateMock.mockReset()
  deleteMock.mockReset()
  reorderMock.mockReset()
  fetchChannelsMock.mockReset()
  confirmMock.mockReset()
  vi.useFakeTimers()
})

describe('EscalationView', () => {
  it('renders policies in priority order from stub', async () => {
    listMock.mockResolvedValue([
      mkPolicy({ id: 'a', name: 'A', priority: 1 }),
      mkPolicy({ id: 'b', name: 'B', priority: 2 }),
    ])
    fetchChannelsMock.mockResolvedValue([])
    const w = mount(EscalationView)
    await flushPromises()
    const vm = w.vm as unknown as Vm
    expect(vm.policies.map((p) => p.id)).toEqual(['a', 'b'])
  })

  it('stats render 4 KPIs with placeholders for backend-pending metrics', async () => {
    listMock.mockResolvedValue([])
    fetchChannelsMock.mockResolvedValue([])
    const w = mount(EscalationView)
    await flushPromises()
    const vm = w.vm as unknown as Vm
    expect(vm.stats.length).toBe(4)
    expect(vm.stats[0]).toEqual({ label: 'Policies', value: '0', tip: '' })
    expect(vm.stats[1].value).toBe('—')
    expect(vm.stats[1].tip).toContain('Backend metric pending')
  })

  it('moveUp swaps two cards and debounces reorder service call', async () => {
    listMock.mockResolvedValue([
      mkPolicy({ id: 'a', name: 'A', priority: 1 }),
      mkPolicy({ id: 'b', name: 'B', priority: 2 }),
    ])
    fetchChannelsMock.mockResolvedValue([])
    reorderMock.mockResolvedValue([
      mkPolicy({ id: 'b', name: 'B', priority: 1 }),
      mkPolicy({ id: 'a', name: 'A', priority: 2 }),
    ])
    const w = mount(EscalationView)
    await flushPromises()
    const vm = w.vm as unknown as Vm
    vm.moveUp(vm.policies[1])
    expect(vm.policies.map((p) => p.id)).toEqual(['b', 'a'])
    vi.advanceTimersByTime(600)
    await flushPromises()
    expect(reorderMock).toHaveBeenCalledWith(['b', 'a'])
  })

  it('disable toggle on active policy → confirm body mentions "Active incident escalations will continue"', async () => {
    const p = mkPolicy({ id: 'a', name: 'A', is_active: true })
    listMock.mockResolvedValue([p])
    fetchChannelsMock.mockResolvedValue([])
    confirmMock.mockResolvedValue(false)
    const w = mount(EscalationView)
    await flushPromises()
    const vm = w.vm as unknown as Vm
    await vm.onToggle(vm.policies[0])
    expect(confirmMock).toHaveBeenCalled()
    const opts = confirmMock.mock.calls[0]?.[0] as { body: string }
    expect(opts.body).toContain('Active incident escalations will continue')
  })

  it('onSubmit creates policy via service and reloads', async () => {
    listMock.mockResolvedValue([])
    fetchChannelsMock.mockResolvedValue([])
    createMock.mockResolvedValue(mkPolicy({ id: 'new' }))
    const w = mount(EscalationView)
    await flushPromises()
    const vm = w.vm as unknown as Vm
    await vm.onSubmit({
      name: 'New',
      scope: { kind: 'component', value: 'c' },
      is_active: true,
      steps: [{ delay_minutes: 5, channel_ids: ['ch'] }],
    })
    expect(createMock).toHaveBeenCalled()
    expect(listMock).toHaveBeenCalledTimes(2)
  })

  it('onDelete with confirm true → service.delete called + row removed', async () => {
    const p = mkPolicy({ id: 'a' })
    listMock.mockResolvedValue([p])
    fetchChannelsMock.mockResolvedValue([])
    confirmMock.mockResolvedValue(true)
    deleteMock.mockResolvedValue(undefined)
    const w = mount(EscalationView)
    await flushPromises()
    const vm = w.vm as unknown as Vm
    await vm.onDelete(p)
    expect(deleteMock).toHaveBeenCalledWith('a')
    expect(vm.policies).toEqual([])
  })
})
