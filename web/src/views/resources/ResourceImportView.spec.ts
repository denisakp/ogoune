import { describe, expect, it, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'

const pushMock = vi.fn()
vi.mock('vue-router', async (importActual) => ({
  ...(await importActual<typeof import('vue-router')>()),
  useRouter: () => ({ push: pushMock }),
}))

const dryRunMock = vi.fn()
const importMock = vi.fn()
vi.mock('@/services/resourceImportService', () => ({
  dryRunImport: (...args: unknown[]) => dryRunMock(...args),
  importManifest: (...args: unknown[]) => importMock(...args),
}))

import ResourceImportView from './ResourceImportView.vue'

const stubs = {
  UButton: {
    props: ['disabled'],
    template: '<button :disabled="disabled" @click="$emit(\'click\')"><slot /></button>',
  },
  UTextarea: {
    props: ['modelValue'],
    template:
      '<textarea :value="modelValue" @input="$emit(\'update:modelValue\', $event.target.value)"></textarea>',
  },
  USelect: { props: ['modelValue', 'items'], template: '<select />' },
  UAlert: { props: ['title'], template: '<div class="alert">{{ title }}</div>' },
  UBadge: { template: '<span class="badge"><slot /></span>' },
}

function mountView() {
  return mount(ResourceImportView, { global: { stubs } })
}

describe('ResourceImportView', () => {
  beforeEach(() => {
    dryRunMock.mockReset()
    importMock.mockReset()
    pushMock.mockReset()
  })

  it('renders the dry-run preview rows', async () => {
    dryRunMock.mockResolvedValue({
      dry_run: true,
      total: 2,
      created: 0,
      skipped: 0,
      failed: 1,
      rows: [
        { index: 0, name: 'Good', valid: true, action: 'create' },
        { index: 1, name: 'Bad', valid: false, action: 'error', errors: ['target is required'] },
      ],
    })

    const wrapper = mountView()
    await wrapper.find('textarea').setValue('version: 1')
    await wrapper.findAll('button').find((b) => b.text().includes('Preview'))!.trigger('click')
    await flushPromises()

    expect(dryRunMock).toHaveBeenCalledOnce()
    const rows = wrapper.findAll('tbody tr')
    expect(rows).toHaveLength(2)
    expect(wrapper.text()).toContain('target is required')
    // Errors present → an error alert is shown.
    expect(wrapper.text()).toContain('nothing was imported')
  })

  it('confirm calls importManifest after a clean preview', async () => {
    dryRunMock.mockResolvedValue({
      dry_run: true,
      total: 1,
      created: 0,
      skipped: 0,
      failed: 0,
      rows: [{ index: 0, name: 'Good', valid: true, action: 'create' }],
    })
    importMock.mockResolvedValue({
      dry_run: false,
      total: 1,
      created: 1,
      skipped: 0,
      failed: 0,
      rows: [{ index: 0, name: 'Good', valid: true, action: 'create' }],
    })

    const wrapper = mountView()
    await wrapper.find('textarea').setValue('version: 1')
    await wrapper.findAll('button').find((b) => b.text().includes('Preview'))!.trigger('click')
    await flushPromises()

    await wrapper.findAll('button').find((b) => b.text().includes('Confirm'))!.trigger('click')
    await flushPromises()

    expect(importMock).toHaveBeenCalledOnce()
    expect(wrapper.text()).toContain('Imported 1 monitor')
  })
})
