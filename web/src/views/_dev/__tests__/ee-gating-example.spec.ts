import { describe, expect, it, vi } from 'vitest'
import { computed, ref } from 'vue'
import { mount } from '@vue/test-utils'
import NuxtUIDemoView from '../NuxtUIDemoView.vue'

const edition = ref<'community' | 'enterprise'>('community')
const isEnterprise = computed(() => edition.value === 'enterprise')

vi.mock('@/composables/useLicence', () => ({
  useLicence: () => ({
    edition,
    isEnterprise,
    isLoaded: ref(true),
    version: ref('1.0.0'),
    load: async () => {},
  }),
}))

const stubs = {
  UCard: { template: '<div><slot /></div>' },
  UButton: {
    props: ['disabled'],
    template: '<button :disabled="disabled"><slot /></button>',
  },
  UEditionBadge: true,
  UInput: true,
  UIcon: true,
  UStatusBadge: true,
  UFilterChip: true,
  UKbd: true,
  USkeleton: true,
  UStepper: true,
  UStatCard: true,
  UEmpty: true,
  UUptimeBar: true,
  UUptimeCalendar: true,
}

describe('EE-gating live example in NuxtUIDemoView (spec 055 SC-006)', () => {
  it('disables the "Add team member" button when the edition is community', () => {
    edition.value = 'community'
    const wrapper = mount(NuxtUIDemoView, { global: { stubs } })
    const button = wrapper.find('[data-test="ee-gated-action"]')
    expect(button.exists()).toBe(true)
    expect(button.attributes('disabled')).toBeDefined()
  })

  it('enables the "Add team member" button when the edition is enterprise', () => {
    edition.value = 'enterprise'
    const wrapper = mount(NuxtUIDemoView, { global: { stubs } })
    const button = wrapper.find('[data-test="ee-gated-action"]')
    expect(button.exists()).toBe(true)
    expect(button.attributes('disabled')).toBeUndefined()
  })
})
