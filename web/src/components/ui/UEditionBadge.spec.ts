import { describe, expect, it, vi, beforeEach } from 'vitest'
import { computed, ref } from 'vue'
import { mount } from '@vue/test-utils'
import UEditionBadge from './UEditionBadge.vue'

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

describe('UEditionBadge', () => {
  beforeEach(() => {
    edition.value = 'community'
  })

  it('renders the EE pill when edition="ee" prop is provided', () => {
    const wrapper = mount(UEditionBadge, { props: { edition: 'ee' } })
    expect(wrapper.attributes('data-edition')).toBe('ee')
    expect(wrapper.text()).toBe('EE')
  })

  it('renders the CE pill when edition="ce" prop is provided', () => {
    const wrapper = mount(UEditionBadge, { props: { edition: 'ce' } })
    expect(wrapper.attributes('data-edition')).toBe('ce')
    expect(wrapper.text()).toBe('CE')
  })

  it('renders nothing in CE when no edition prop is provided (community licence)', () => {
    edition.value = 'community'
    const wrapper = mount(UEditionBadge)
    expect(wrapper.find('[data-edition]').exists()).toBe(false)
  })

  it('renders the EE pill in EE when no edition prop is provided (enterprise licence)', () => {
    edition.value = 'enterprise'
    const wrapper = mount(UEditionBadge)
    expect(wrapper.attributes('data-edition')).toBe('ee')
    expect(wrapper.text()).toBe('EE')
  })
})
