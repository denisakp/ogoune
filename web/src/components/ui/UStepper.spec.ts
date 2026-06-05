import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import UStepper from './UStepper.vue'

describe('UStepper', () => {
  it('marks the active step with data-active=true', () => {
    const wrapper = mount(UStepper, {
      props: { steps: ['One', 'Two', 'Three'], activeStep: 1 },
    })
    const items = wrapper.findAll('li')
    expect(items[0]?.attributes('data-active')).toBe('false')
    expect(items[1]?.attributes('data-active')).toBe('true')
    expect(items[2]?.attributes('data-active')).toBe('false')
  })

  it('renders step labels', () => {
    const wrapper = mount(UStepper, {
      props: { steps: ['Profile', 'Verify', 'Done'], activeStep: 0 },
    })
    expect(wrapper.text()).toContain('Profile')
    expect(wrapper.text()).toContain('Verify')
    expect(wrapper.text()).toContain('Done')
  })

  it('accepts variant=dots', () => {
    const wrapper = mount(UStepper, {
      props: { steps: ['a', 'b'], activeStep: 0, variant: 'dots' },
    })
    expect(wrapper.exists()).toBe(true)
  })
})
