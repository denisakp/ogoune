import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import type { ErasurePreview, ErasureResult } from '@/types'

const { listMock, previewMock, eraseMock, pushMock } = vi.hoisted(() => ({
  listMock: vi.fn(),
  previewMock: vi.fn(),
  eraseMock: vi.fn(),
  pushMock: vi.fn(),
}))

vi.mock('@/services/privacyService', async () => {
  const actual = await vi.importActual<typeof import('@/services/privacyService')>(
    '@/services/privacyService',
  )
  return {
    ...actual,
    default: { listOtherAccounts: listMock, previewErasure: previewMock, erase: eraseMock },
  }
})

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: pushMock, replace: vi.fn() }),
  useRoute: () => ({ path: '/settings/privacy', query: {}, params: {} }),
  RouterLink: { props: ['to'], template: '<a :href="to" v-bind="$attrs"><slot /></a>' },
}))

import { CannotEraseSelfError, ErasureConflictError, InvalidCredentialsError } from '@/services/privacyService'
import ErasurePanel from '../ErasurePanel.vue'

const modal = {
  props: ['open', 'title', 'description'],
  template: '<div v-if="open" data-test="modal"><slot name="body" /></div>',
}
const form = { template: '<form v-bind="$attrs"><slot /></form>' }
const stubs = {
  RouterLink: { props: ['to'], template: '<a :href="to" v-bind="$attrs"><slot /></a>' },
  UModal: modal,
  Modal: modal,
  UForm: form,
  Form: form,
  UFormField: { template: '<div><slot /></div>' },
  UInput: { props: ['modelValue'], template: '<input v-bind="$attrs" :value="modelValue" />' },
  UButton: { template: '<button v-bind="$attrs"><slot /></button>' },
  UBadge: { template: '<span><slot /></span>' },
  UAlert: {
    props: ['title'],
    template: '<div v-bind="$attrs">{{ title }}<slot name="title" /><slot name="description" /></div>',
  },
  UEmpty: { props: ['title', 'description'], template: '<div v-bind="$attrs">{{ title }} {{ description }}</div>' },
}

function pv(over: Partial<ErasurePreview> = {}): ErasurePreview {
  return {
    kind: 'address',
    account: null,
    channels: [],
    report_recipient: false,
    reports_sent: 0,
    sessions: 0,
    api_keys: 0,
    updates_unlinked: 0,
    manual_review: [],
    previously_erased_at: null,
    transport_change: null,
    ...over,
  }
}

function res(over: Partial<ErasureResult> = {}): ErasureResult {
  return {
    record_id: 'r1',
    changes: { channels: 1 },
    disabled_channels: [],
    manual_review: [],
    transport_change: null,
    report_recipient_cleared: false,
    ...over,
  }
}

interface Vm {
  address: string
  form: { confirm_email: string; password: string; code: string }
  confirmOpen: boolean
  eraseError: string | null
  previewError: string | null
  result: ErasureResult | null
  previewAddress: () => void
  onErase: () => Promise<void>
  openConfirm: () => void
}

async function build(twoFactorEnabled = false, accounts: unknown[] = []) {
  listMock.mockResolvedValue(accounts)
  const w = mount(ErasurePanel, { props: { twoFactorEnabled }, global: { stubs } })
  await flushPromises()
  return { w, vm: w.vm as unknown as Vm }
}

async function previewFor(vm: Vm, w: Awaited<ReturnType<typeof build>>['w'], p: ErasurePreview) {
  previewMock.mockResolvedValue(p)
  vm.address = 'ada@example.com'
  vm.previewAddress()
  await flushPromises()
  expect(w.find('[data-test="erasure-preview-panel"]').exists()).toBe(true)
}

beforeEach(() => {
  listMock.mockReset()
  previewMock.mockReset()
  eraseMock.mockReset()
  pushMock.mockReset()
})

describe('ErasurePanel (spec 095)', () => {
  it('previews an address and renders removed / kept / review groups', async () => {
    const { w, vm } = await build()
    await previewFor(
      vm,
      w,
      pv({
        channels: [
          { id: 'c1', name: 'Ops mail', type: 'smtp', fields: ['recipient'], will_disable: true },
        ],
        report_recipient: true,
        reports_sent: 4,
        updates_unlinked: 2,
        manual_review: [
          { channel_id: 'c2', channel_name: 'Hook', channel_type: 'webhook', reason: 'address_in_url' },
          { channel_id: 'c3', channel_name: 'Locked', channel_type: 'slack', reason: 'undecryptable' },
        ],
      }),
    )
    expect(previewMock).toHaveBeenCalledWith({ email: 'ada@example.com' })
    const removed = w.find('[data-test="erasure-removed"]').text()
    expect(removed).toContain('Ops mail')
    expect(removed).toContain('recipient')
    expect(removed).toContain('will be disabled')
    expect(removed).toContain('Monthly report recipient')
    expect(removed).toContain('4 reports sent')
    expect(w.find('[data-test="erasure-kept"]').text()).toContain('kept, author removed')
    const review = w.find('[data-test="erasure-review"]').text()
    expect(review).toContain('address is inside a URL')
    expect(review).toContain('configuration could not be read')
  })

  it('shows the empty state, the previous erasure and the transport change', async () => {
    const { w, vm } = await build()
    await previewFor(
      vm,
      w,
      pv({
        previously_erased_at: '2026-09-01T10:00:00Z',
        transport_change: { from_channel: { id: 'a', name: 'Main SMTP' }, to_channel: null },
      }),
    )
    expect(w.find('[data-test="erasure-empty"]').text()).toContain('held nowhere')
    expect(w.find('[data-test="erasure-previous"]').text()).toContain('Erased before on')
    expect(w.find('[data-test="erasure-transport"]').text()).toContain('no email channel left')
  })

  it('names the replacement channel when there is one', async () => {
    const { w, vm } = await build()
    await previewFor(
      vm,
      w,
      pv({
        channels: [{ id: 'c1', name: 'A', type: 'smtp', fields: ['recipient'], will_disable: false }],
        transport_change: {
          from_channel: { id: 'a', name: 'Main SMTP' },
          to_channel: { id: 'b', name: 'Backup SMTP' },
        },
      }),
    )
    const t = w.find('[data-test="erasure-transport"]').text()
    expect(t).toContain('Main SMTP')
    expect(t).toContain('will use Backup SMTP')
  })

  it('lists other accounts and previews one by id with sessions, keys and author-removed updates', async () => {
    const { w } = await build(false, [{ id: 'u2', email: 'bob@example.com', last_login_at: null }])
    expect(w.find('[data-test="erasure-accounts"]').text()).toContain('bob@example.com')
    previewMock.mockResolvedValue(
      pv({
        kind: 'account',
        account: { id: 'u2', email: 'bob@example.com', last_login_at: null },
        sessions: 3,
        api_keys: 2,
        updates_unlinked: 5,
      }),
    )
    await w.find('[data-test="erasure-account-u2"]').trigger('click')
    await flushPromises()
    expect(previewMock).toHaveBeenCalledWith({ account_id: 'u2' })
    expect(w.find('[data-test="erasure-removed"]').text()).toContain('3 sessions')
    expect(w.find('[data-test="erasure-removed"]').text()).toContain('2 API keys')
    expect(w.find('[data-test="erasure-kept"]').text()).toContain('kept, author removed')
  })

  it('shows a clear message when asked about yourself', async () => {
    const { w, vm } = await build()
    previewMock.mockRejectedValue(new CannotEraseSelfError())
    vm.address = 'me@example.com'
    vm.previewAddress()
    await flushPromises()
    expect(w.find('[data-test="erasure-preview-error"]').text()).toContain('cannot erase your own account')
    expect(w.find('[data-test="erasure-preview-panel"]').exists()).toBe(false)
  })

  it('confirming calls erase with the subject, typed address, password and code', async () => {
    const { w, vm } = await build(true)
    await previewFor(vm, w, pv({ channels: [{ id: 'c1', name: 'A', type: 'smtp', fields: ['r'], will_disable: false }] }))
    vm.openConfirm()
    await flushPromises()
    expect(w.find('[data-test="erasure-code"]').exists()).toBe(true)
    eraseMock.mockResolvedValue(res())
    vm.form.confirm_email = 'ada@example.com'
    vm.form.password = 'pw'
    vm.form.code = 'abcd-efgh-jkmn'
    await vm.onErase()
    await flushPromises()
    expect(eraseMock).toHaveBeenCalledWith({
      email: 'ada@example.com',
      confirm_email: 'ada@example.com',
      password: 'pw',
      code: 'abcd-efgh-jkmn',
    })
    expect(vm.confirmOpen).toBe(false)
    expect(vm.form.password).toBe('')
    expect(vm.form.code).toBe('')
  })

  it('wrong credentials: one message, modal stays open, secrets cleared, no navigation', async () => {
    const { w, vm } = await build()
    await previewFor(vm, w, pv())
    vm.openConfirm()
    eraseMock.mockRejectedValue(new InvalidCredentialsError())
    vm.form.confirm_email = 'ada@example.com'
    vm.form.password = 'nope'
    await vm.onErase()
    await flushPromises()
    expect(vm.eraseError).toBe('The password or the two-factor code is not correct.')
    expect(vm.confirmOpen).toBe(true)
    expect(vm.form.password).toBe('')
    expect(w.find('[data-test="erasure-error"]').exists()).toBe(true)
    expect(pushMock).not.toHaveBeenCalled()
  })

  it('a conflict says nothing changed and keeps the modal open', async () => {
    const { w, vm } = await build()
    await previewFor(vm, w, pv())
    vm.openConfirm()
    eraseMock.mockRejectedValue(new ErasureConflictError())
    await vm.onErase()
    await flushPromises()
    expect(vm.eraseError).toContain('Nothing was changed')
    expect(vm.confirmOpen).toBe(true)
  })

  it('success shows disabled channels, the reports-off notice and links', async () => {
    const { w, vm } = await build()
    await previewFor(vm, w, pv())
    vm.openConfirm()
    eraseMock.mockResolvedValue(
      res({
        changes: { channels: 2, channels_disabled: 1, sessions: 0 },
        disabled_channels: [{ id: 'c1', name: 'Ops mail', type: 'smtp', fields: [], will_disable: true }],
        report_recipient_cleared: true,
        manual_review: [{ channel_id: 'c2', channel_name: 'Hook', channel_type: 'webhook', reason: 'address_in_url' }],
      }),
    )
    await vm.onErase()
    await flushPromises()
    const out = w.find('[data-test="erasure-result"]')
    expect(out.text()).toContain('2 channels cleaned')
    expect(out.text()).not.toContain('sessions removed')
    expect(out.find('[data-test="erasure-disabled"] a').attributes('href')).toBe('/notifications')
    expect(out.text()).toContain('add a recipient and enable it')
    expect(out.find('[data-test="erasure-reports-off"]').text()).toContain('Monthly reports are now off')
    expect(out.find('[data-test="erasure-reports-off"] a').attributes('href')).toBe('/reports')
    expect(out.find('[data-test="erasure-result-review"]').text()).toContain('Hook')
  })
})
