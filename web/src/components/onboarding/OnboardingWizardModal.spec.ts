import { describe, expect, it, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'

const markDoneMock = vi.fn().mockResolvedValue(undefined)
vi.mock('@/composables/useOnboardingState', () => ({
  useOnboardingState: () => ({ markDone: markDoneMock }),
}))

import OnboardingWizardModal from './OnboardingWizardModal.vue'

const stubs = {
  UModal: {
    template: '<div><slot name="content" /></div>',
    props: ['open', 'ui'],
    emits: ['update:open'],
  },
  UIcon: { template: '<span />' },
  UButton: { template: '<button><slot /></button>' },
  UInput: { template: '<input />' },
  USelect: { template: '<select />' },
}

function build() {
  return mount(OnboardingWizardModal, {
    global: { stubs },
    props: { open: true },
  })
}

beforeEach(() => {
  markDoneMock.mockClear()
})

describe('OnboardingWizardModal', () => {
  it('starts on step 0 (Welcome) by default', () => {
    const w = build()
    expect((w.vm as unknown as { activeStep: number }).activeStep).toBe(0)
  })

  it('next() advances activeStep from 0 → 1', async () => {
    const w = build()
    ;(w.vm as unknown as { next: () => void }).next()
    await w.vm.$nextTick()
    expect((w.vm as unknown as { activeStep: number }).activeStep).toBe(1)
  })

  it('skip() calls markDone() exactly once and emits close', async () => {
    const w = build()
    await (w.vm as unknown as { skip: () => Promise<void> }).skip()
    expect(markDoneMock).toHaveBeenCalledTimes(1)
    expect(w.emitted('close')).toBeTruthy()
  })

  it('finish() on summary calls markDone() and emits close', async () => {
    const w = build()
    ;(w.vm as unknown as { activeStep: number }).activeStep = 3
    await (w.vm as unknown as { finish: () => Promise<void> }).finish()
    expect(markDoneMock).toHaveBeenCalledTimes(1)
    expect(w.emitted('close')).toBeTruthy()
  })

  it('does not call markDone twice when skip then finish are called', async () => {
    const w = build()
    await (w.vm as unknown as { skip: () => Promise<void> }).skip()
    await (w.vm as unknown as { finish: () => Promise<void> }).finish()
    expect(markDoneMock).toHaveBeenCalledTimes(1)
  })
})
