/* eslint-disable @typescript-eslint/ban-ts-comment */
// @ts-nocheck — spec 059 polish debt: index-access narrowing
import { describe, expect, it, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'

const listMock = vi.fn()
const revokeMock = vi.fn()
const revokeOthersMock = vi.fn()

vi.mock('@/services/sessionsService', () => ({
  default: {
    list: (...a: unknown[]) => listMock(...a),
    revoke: (...a: unknown[]) => revokeMock(...a),
    revokeOthers: (...a: unknown[]) => revokeOthersMock(...a),
  },
}))

const confirmMock = vi.fn()
vi.mock('@/composables/useConfirm', () => ({
  useConfirm: (opts: unknown) => confirmMock(opts),
}))

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: vi.fn(), replace: vi.fn(), resolve: () => ({ href: '#' }) }),
  useRoute: () => ({ path: '/settings/sessions', params: {}, query: {}, name: 'SettingsSessions' }),
  useLink: () => ({ href: { value: '#' }, navigate: vi.fn(), isActive: { value: false } }),
  RouterLink: { template: '<a><slot /></a>' },
}))

import SessionsView from '../SessionsView.vue'

type Sess = {
  id: string
  browser: string
  os: string
  ip: string
  location: string | null
  last_active_at: string
  is_current: boolean
  revoked_at: null
}

const sessions: Sess[] = [
  {
    id: 's1',
    browser: 'Chrome',
    os: 'macOS',
    ip: '1.1.1.1',
    location: 'Paris, FR',
    last_active_at: new Date().toISOString(),
    is_current: true,
    revoked_at: null,
  },
  {
    id: 's2',
    browser: 'Firefox',
    os: 'Windows',
    ip: '2.2.2.2',
    location: null,
    last_active_at: new Date().toISOString(),
    is_current: false,
    revoked_at: null,
  },
  {
    id: 's3',
    browser: 'Safari',
    os: 'iOS',
    ip: '3.3.3.3',
    location: null,
    last_active_at: new Date().toISOString(),
    is_current: false,
    revoked_at: null,
  },
]

type Vm = {
  sessions: Sess[]
  showRevokeAll: boolean
  onRevoke: (id: string) => Promise<void>
  onRevokeAllOthers: () => Promise<void>
}

beforeEach(() => {
  listMock.mockReset()
  revokeMock.mockReset()
  revokeOthersMock.mockReset()
  confirmMock.mockReset()
})

describe('SessionsView', () => {
  it('loads sessions on mount and exposes them', async () => {
    listMock.mockResolvedValue(sessions)
    const w = mount(SessionsView)
    await flushPromises()
    const vm = w.vm as unknown as Vm
    expect(vm.sessions.length).toBe(3)
    expect(vm.showRevokeAll).toBe(true)
  })

  it('confirm body for single-session revoke contains "Effective immediately."', async () => {
    listMock.mockResolvedValue(sessions)
    confirmMock.mockResolvedValue(true)
    revokeMock.mockResolvedValue(undefined)
    const w = mount(SessionsView)
    await flushPromises()
    const vm = w.vm as unknown as Vm
    await vm.onRevoke('s2')
    expect(confirmMock).toHaveBeenCalled()
    const opts = confirmMock.mock.calls[0]?.[0] as { body: string; ctaLabel: string }
    expect(opts.body).toContain('Effective immediately.')
    expect(revokeMock).toHaveBeenCalledWith('s2')
    expect(vm.sessions.find((s) => s.id === 's2')).toBeUndefined()
  })

  it('skips revoke service call when confirm is dismissed', async () => {
    listMock.mockResolvedValue(sessions)
    confirmMock.mockResolvedValue(false)
    const w = mount(SessionsView)
    await flushPromises()
    const vm = w.vm as unknown as Vm
    await vm.onRevoke('s2')
    expect(revokeMock).not.toHaveBeenCalled()
    expect(vm.sessions.length).toBe(3)
  })

  it('Revoke all others keeps only the current session locally', async () => {
    listMock.mockResolvedValue(sessions)
    confirmMock.mockResolvedValue(true)
    revokeOthersMock.mockResolvedValue(undefined)
    const w = mount(SessionsView)
    await flushPromises()
    const vm = w.vm as unknown as Vm
    await vm.onRevokeAllOthers()
    expect(revokeOthersMock).toHaveBeenCalled()
    expect(vm.sessions.length).toBe(1)
    expect(vm.sessions[0].is_current).toBe(true)
  })

  it('hides "Revoke all other sessions" when only the current session is active', async () => {
    listMock.mockResolvedValue([sessions[0]])
    const w = mount(SessionsView)
    await flushPromises()
    const vm = w.vm as unknown as Vm
    expect(vm.showRevokeAll).toBe(false)
  })
})
