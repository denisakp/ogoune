import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import UStatCard from './UStatCard.vue'

describe('UStatCard', () => {
  it('renders label + value', () => {
    const wrapper = mount(UStatCard, {
      props: { label: 'Monitors', value: 14, icon: 'i-lucide-radar' },
      global: { stubs: { UCard: { template: '<div><slot /></div>' }, UIcon: true } },
    })
    expect(wrapper.text()).toContain('Monitors')
    expect(wrapper.text()).toContain('14')
  })

  it('renders a sparkline svg when sparkline is provided', () => {
    const wrapper = mount(UStatCard, {
      props: {
        label: 'p95',
        value: '120ms',
        icon: 'i-lucide-activity',
        sparkline: [10, 20, 15, 30, 25],
      },
      global: { stubs: { UCard: { template: '<div><slot /></div>' }, UIcon: true } },
    })
    expect(wrapper.find('svg').exists()).toBe(true)
  })
})
