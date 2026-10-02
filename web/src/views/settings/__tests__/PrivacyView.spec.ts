import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import type { PrivacySummary } from '@/types'

const { getSummaryMock, exportMock, pushMock } = vi.hoisted(() => ({
  getSummaryMock: vi.fn(),
  exportMock: vi.fn(),
  pushMock: vi.fn(),
}))

vi.mock('@/services/privacyService', async () => {
  const actual = await vi.importActual<typeof import('@/services/privacyService')>(
    '@/services/privacyService',
  )
  return {
    ...actual,
    default: {
      getPrivacySummary: getSummaryMock,
      exportPersonalData: exportMock,
      listOtherAccounts: () => Promise.resolve([]),
    },
  }
})

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: pushMock, replace: vi.fn() }),
  useRoute: () => ({ path: '/settings/privacy', query: {}, params: {} }),
  RouterLink: { props: ['to'], template: '<a :href="to" v-bind="$attrs"><slot /></a>' },
}))

import { InvalidCredentialsError } from '@/services/privacyService'
import PrivacyView from '../PrivacyView.vue'

const stubs = {
  RouterLink: { props: ['to'], template: '<a :href="to" v-bind="$attrs"><slot /></a>' },
  UIcon: { template: '<span />' },
  // Rendered inline instead of teleported, so the form is findable.
  UModal: {
    props: ['open', 'title', 'description'],
    template: '<div v-if="open" data-test="modal"><slot name="body" /></div>',
  },
  Modal: {
    props: ['open', 'title', 'description'],
    template: '<div v-if="open" data-test="modal"><slot name="body" /></div>',
  },
}

function summary(over: Partial<PrivacySummary> = {}): PrivacySummary {
  return {
    generated_at: '2026-10-02T08:00:00Z',
    two_factor_enabled: false,
    categories: [
      { key: 'account', count: 1, manage_path: '/settings/account' },
      { key: 'sessions', count: 3, manage_path: '/settings/sessions' },
      { key: 'api_keys', count: 2, manage_path: '/api-keys' },
      { key: 'incident_updates', count: 0, manage_path: '/incidents' },
      { key: 'notification_channels', count: 1, manage_path: '/notifications' },
      { key: 'reports', count: 4, manage_path: '/reports' },
    ],
    not_personal_data: ['monitors', 'host_metrics'],
    unchecked_channels: [],
    ...over,
  }
}

async function build(s: PrivacySummary) {
  getSummaryMock.mockResolvedValue(s)
  const w = mount(PrivacyView, { global: { stubs } })
  await flushPromises()
  return w
}

beforeEach(() => {
  getSummaryMock.mockReset()
  exportMock.mockReset()
  pushMock.mockReset()
  URL.createObjectURL = vi.fn(() => 'blob:x')
  URL.revokeObjectURL = vi.fn()
})
afterEach(() => vi.restoreAllMocks())

describe('PrivacyView (spec 094)', () => {
  it('shows each category with its count and a link to where it is managed', async () => {
    const w = await build(summary())
    const sessions = w.find('[data-test="privacy-category-sessions"]')
    expect(sessions.attributes('href')).toBe('/settings/sessions')
    expect(sessions.find('[data-test="privacy-count"]').text()).toBe('3')
    expect(w.find('[data-test="privacy-category-reports"] [data-test="privacy-count"]').text()).toBe('4')
    expect(w.find('[data-test="privacy-not-personal"]').text()).toContain('host metrics')
  })

  it('names the channels that could not be checked, and only when there are some', async () => {
    const none = await build(summary())
    expect(none.find('[data-test="privacy-unchecked"]').exists()).toBe(false)

    const some = await build(summary({ unchecked_channels: [{ id: 'c1', name: 'Ops pager', type: 'slack' }] }))
    expect(some.find('[data-test="privacy-unchecked"]').text()).toContain('Ops pager')
  })

  it('asks for the two-factor code only when two-factor is on', async () => {
    const off = await build(summary())
    await off.find('[data-test="privacy-export-open"]').trigger('click')
    await flushPromises()
    expect(off.find('[data-test="privacy-password"]').exists()).toBe(true)
    expect(off.find('[data-test="privacy-code"]').exists()).toBe(false)

    const on = await build(summary({ two_factor_enabled: true }))
    await on.find('[data-test="privacy-export-open"]').trigger('click')
    await flushPromises()
    expect(on.find('[data-test="privacy-code"]').exists()).toBe(true)
  })

  it('accepts a backup code in the two-factor field (not just 6 digits)', async () => {
    const on = await build(summary({ two_factor_enabled: true }))
    await on.find('[data-test="privacy-export-open"]').trigger('click')
    await flushPromises()
    const input = on.find('[data-test="privacy-code"]')
    expect(Number(input.attributes('maxlength'))).toBeGreaterThanOrEqual('abcd-efgh-jkmn'.length)
    expect(input.attributes('inputmode')).not.toBe('numeric')

    exportMock.mockResolvedValue({ blob: new Blob(['{}']), filename: 'x.json' })
    const vm = on.vm as unknown as { form: { password: string; code: string }; onSubmit: () => Promise<void> }
    vm.form.password = 'pw'
    vm.form.code = 'abcd-efgh-jkmn'
    await vm.onSubmit()
    await flushPromises()
    expect(exportMock).toHaveBeenCalledWith({ password: 'pw', code: 'abcd-efgh-jkmn' })
  })

  it('submits the credentials and downloads the file', async () => {
    exportMock.mockResolvedValue({ blob: new Blob(['{}']), filename: 'ogoune-personal-data-2026-10-02.json' })
    const w = await build(summary({ two_factor_enabled: true }))
    const vm = w.vm as unknown as { form: { password: string; code: string }; onSubmit: () => Promise<void>; exportOpen: boolean }
    vm.exportOpen = true
    vm.form.password = 'pw'
    vm.form.code = '123456'
    await vm.onSubmit()
    await flushPromises()

    expect(exportMock).toHaveBeenCalledWith({ password: 'pw', code: '123456' })
    expect(URL.createObjectURL).toHaveBeenCalled()
    expect(vm.exportOpen).toBe(false)
    // The password does not linger in the form after submit.
    expect(vm.form.password).toBe('')
  })

  it('a wrong password shows one error, downloads nothing, and does not leave the page', async () => {
    exportMock.mockRejectedValue(new InvalidCredentialsError())
    const w = await build(summary())
    const vm = w.vm as unknown as { form: { password: string }; onSubmit: () => Promise<void>; exportOpen: boolean; exportError: string | null }
    vm.exportOpen = true
    vm.form.password = 'nope'
    await vm.onSubmit()
    await flushPromises()

    expect(vm.exportError).toBe('The password or the two-factor code is not correct.')
    expect(vm.exportOpen).toBe(true)
    expect(URL.createObjectURL).not.toHaveBeenCalled()
    expect(pushMock).not.toHaveBeenCalled()
  })
})
