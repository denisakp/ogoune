import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import UConfirmModal from './UConfirmModal.vue'

interface Exposed {
  confirm: () => void
  dismiss: () => void
  ctaColor: string
  headerIcon: string
}

function mountModal(props: Partial<{ kind: 'default' | 'destructive'; title: string; body: string; ctaLabel: string }> = {}) {
  return mount(UConfirmModal, {
    props: {
      title: 'Delete?',
      body: 'This is permanent.',
      ctaLabel: 'Delete',
      ...props,
    },
    global: {
      stubs: {
        UModal: true,
        UIcon: true,
        UButton: true,
      },
    },
  })
}

describe('UConfirmModal', () => {
  it('emits close=true when confirm() is called', () => {
    const wrapper = mountModal()
    const exposed = wrapper.vm as unknown as Exposed
    exposed.confirm()
    expect(wrapper.emitted('close')?.[0]).toEqual([true])
  })

  it('emits close=false when dismiss() is called', () => {
    const wrapper = mountModal()
    const exposed = wrapper.vm as unknown as Exposed
    exposed.dismiss()
    expect(wrapper.emitted('close')?.[0]).toEqual([false])
  })

  it('uses error color + alert-triangle icon for kind=destructive', () => {
    const wrapper = mountModal({ kind: 'destructive' })
    const exposed = wrapper.vm as unknown as Exposed
    expect(exposed.ctaColor).toBe('error')
    expect(exposed.headerIcon).toBe('i-lucide-alert-triangle')
  })

  it('defaults to primary color + help-circle icon for kind=default', () => {
    const wrapper = mountModal({ kind: 'default' })
    const exposed = wrapper.vm as unknown as Exposed
    expect(exposed.ctaColor).toBe('primary')
    expect(exposed.headerIcon).toBe('i-lucide-help-circle')
  })
})
