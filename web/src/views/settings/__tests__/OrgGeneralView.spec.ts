import { describe, expect, it, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'

const getGeneralMock = vi.fn()
const updateGeneralMock = vi.fn()
const uploadLogoMock = vi.fn()
vi.mock('@/services/orgService', () => ({
  default: {
    getGeneral: (...a: unknown[]) => getGeneralMock(...a),
    updateGeneral: (...a: unknown[]) => updateGeneralMock(...a),
    uploadLogo: (...a: unknown[]) => uploadLogoMock(...a),
  },
}))

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: vi.fn(), replace: vi.fn(), resolve: () => ({ href: '#' }) }),
  useRoute: () => ({
    path: '/settings/org/general',
    params: {},
    query: {},
    name: 'SettingsOrgGeneral',
  }),
  useLink: () => ({ href: { value: '#' }, navigate: vi.fn(), isActive: { value: false } }),
  RouterLink: { template: '<a><slot /></a>' },
}))

import OrgGeneralView from '../OrgGeneralView.vue'

type Vm = {
  state: {
    name: string
    timezone: string
    date_format: string
    logo_url: string | null
  }
  initial: { name: string } | null
  dirty: boolean
  save: () => Promise<void>
  reset: () => void
}

const baseline = {
  name: 'Ogoune',
  logo_url: null,
  timezone: 'UTC',
  date_format: 'YYYY-MM-DD',
}

beforeEach(() => {
  getGeneralMock.mockReset()
  updateGeneralMock.mockReset()
  uploadLogoMock.mockReset()
})

describe('OrgGeneralView', () => {
  it('editing name enables the save bar (dirty=true)', async () => {
    getGeneralMock.mockResolvedValue(baseline)
    const w = mount(OrgGeneralView)
    await flushPromises()
    const vm = w.vm as unknown as Vm
    expect(vm.dirty).toBe(false)
    vm.state.name = 'Acme'
    await flushPromises()
    expect(vm.dirty).toBe(true)
  })

  it('save() calls updateGeneral and re-baselines initial state', async () => {
    getGeneralMock.mockResolvedValue(baseline)
    updateGeneralMock.mockImplementation(async (p: { name: string }) => ({ ...baseline, ...p }))
    const w = mount(OrgGeneralView)
    await flushPromises()
    const vm = w.vm as unknown as Vm
    vm.state.name = 'Acme'
    await vm.save()
    await flushPromises()
    expect(updateGeneralMock).toHaveBeenCalled()
    expect(vm.dirty).toBe(false)
    expect(vm.initial?.name).toBe('Acme')
  })

  it('reset() restores baseline values and clears dirty', async () => {
    getGeneralMock.mockResolvedValue(baseline)
    const w = mount(OrgGeneralView)
    await flushPromises()
    const vm = w.vm as unknown as Vm
    vm.state.timezone = 'Europe/Paris'
    expect(vm.dirty).toBe(true)
    vm.reset()
    await flushPromises()
    expect(vm.state.timezone).toBe('UTC')
    expect(vm.dirty).toBe(false)
  })
})
