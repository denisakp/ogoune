/* eslint-disable @typescript-eslint/ban-ts-comment */
// @ts-nocheck — spec 059 polish debt: index-access narrowing
import { describe, expect, it, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'

const { getMock, updateMock, verifyMock } = vi.hoisted(() => ({
  getMock: vi.fn(),
  updateMock: vi.fn(),
  verifyMock: vi.fn(),
}))

vi.mock('@/services/statusPageSettingsService', () => ({
  getStatusPageSettings: getMock,
  updateStatusPageSettings: updateMock,
  verifyStatusPageDomain: verifyMock,
}))

const { runtimeRef } = vi.hoisted(() => ({
  runtimeRef: { value: { ssl_provider: 'external', edition: 'community', version: 't' } },
}))
vi.mock('@/composables/useRuntimeConfig', () => ({
  useRuntimeConfig: () => runtimeRef.value,
}))

vi.mock('@/components/settings/domain/DnsRecordsTable.vue', () => ({
  default: {
    name: 'DnsRecordsTable',
    template: '<div data-testid="dns" />',
    props: ['records', 'rechecking'],
  },
}))

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: vi.fn(), replace: vi.fn(), resolve: () => ({ href: '#' }) }),
  useRoute: () => ({
    path: '/settings/org/status-page',
    params: {},
    query: {},
    name: 'SettingsStatusPage',
  }),
  useLink: () => ({ href: { value: '#' }, navigate: vi.fn(), isActive: { value: false } }),
  RouterLink: { template: '<a><slot /></a>' },
}))

import StatusPageSettingsView from '../StatusPageSettingsView.vue'

const baseline = {
  id: 'sp1',
  name: 'Status Page',
  homepage_url: '',
  custom_domain: '',
  umami_website_id: '',
  enable_details_page: true,
  show_uptime_percentage: true,
  hide_paused_monitors: true,
  show_incident_history: true,
  custom_domain_status: 'pending' as const,
  custom_domain_ssl_status: 'none' as const,
  custom_domain_dns_records: [],
  created_at: '',
  updated_at: '',
}

beforeEach(() => {
  getMock.mockReset()
  updateMock.mockReset()
  verifyMock.mockReset()
  runtimeRef.value = { ssl_provider: 'external', edition: 'community', version: 't' }
})

describe('StatusPageSettingsView', () => {
  it('loads settings on mount and exposes dirty=false initially', async () => {
    getMock.mockResolvedValue(baseline)
    const w = mount(StatusPageSettingsView)
    const vm = w.vm as unknown as { dirty: boolean; load: () => Promise<void> }
    await vm.load()
    await flushPromises()
    expect(getMock).toHaveBeenCalled()
    expect(vm.dirty).toBe(false)
  })

  it('editing name flips dirty=true', async () => {
    getMock.mockResolvedValue(baseline)
    const w = mount(StatusPageSettingsView)
    const vm = w.vm as unknown as {
      state: { name: string }
      dirty: boolean
      load: () => Promise<void>
    }
    await vm.load()
    await flushPromises()
    vm.state.name = 'Acme Status'
    await flushPromises()
    expect(vm.dirty).toBe(true)
  })

  it('save() calls updateStatusPageSettings + clears dirty', async () => {
    getMock.mockResolvedValue(baseline)
    updateMock.mockImplementation(async (p: { name: string }) => ({ ...baseline, ...p }))
    const w = mount(StatusPageSettingsView)
    const vm = w.vm as unknown as {
      state: { name: string }
      dirty: boolean
      load: () => Promise<void>
      save: () => Promise<void>
    }
    await vm.load()
    await flushPromises()
    vm.state.name = 'Acme'
    await vm.save()
    expect(updateMock).toHaveBeenCalled()
    expect(vm.dirty).toBe(false)
  })

  it('verify() calls verifyStatusPageDomain and refreshes state', async () => {
    getMock.mockResolvedValue({
      ...baseline,
      custom_domain: 'status.acme.com',
      custom_domain_dns_records: [
        { type: 'CNAME', host: 'status.acme.com', value: 'status.ogoune.app', status: 'pending' },
      ],
    })
    verifyMock.mockResolvedValue({
      ...baseline,
      custom_domain: 'status.acme.com',
      custom_domain_status: 'verified',
      custom_domain_dns_records: [
        { type: 'CNAME', host: 'status.acme.com', value: 'status.ogoune.app', status: 'verified' },
      ],
    })
    const w = mount(StatusPageSettingsView)
    const vm = w.vm as unknown as {
      state: { custom_domain_status: string }
      load: () => Promise<void>
      verify: () => Promise<void>
    }
    await vm.load()
    await flushPromises()
    await vm.verify()
    expect(verifyMock).toHaveBeenCalled()
    expect(vm.state.custom_domain_status).toBe('verified')
  })

  it('SSL_PROVIDER=letsencrypt + saved domain → SSL panel text is "Provisioning Let\'s Encrypt cert (~5 min)"', async () => {
    runtimeRef.value = { ssl_provider: 'letsencrypt', edition: 'community', version: 't' }
    getMock.mockResolvedValue({ ...baseline, custom_domain: 'status.acme.com' })
    const w = mount(StatusPageSettingsView)
    const vm = w.vm as unknown as { load: () => Promise<void>; sslPanelLabel: string }
    await vm.load()
    await flushPromises()
    expect(vm.sslPanelLabel).toContain("Provisioning Let's Encrypt cert (~5 min)")
  })
})
