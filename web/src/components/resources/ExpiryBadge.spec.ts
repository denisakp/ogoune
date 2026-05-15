import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import { defineComponent } from 'vue'

// Stub AntDV a-tag so we can inspect rendered text/color without full AntDV setup
const ATagStub = defineComponent({
  name: 'ATag',
  props: { color: String },
  template: '<span :data-color="color"><slot /></span>',
})

import ExpiryBadge from '@/components/resources/ExpiryBadge.vue'

const mountBadge = (props: InstanceType<typeof ExpiryBadge>['$props']) =>
  mount(ExpiryBadge, {
    props,
    global: {
      stubs: { 'a-tag': ATagStub },
    },
  })

describe('ExpiryBadge', () => {
  it('renders nothing when status is ok', () => {
    const wrapper = mountBadge({ type: 'ssl', daysRemaining: 20, status: 'ok' })
    expect(wrapper.find('span').exists()).toBe(false)
  })

  it('renders SSL warning badge with amber color', () => {
    const wrapper = mountBadge({ type: 'ssl', daysRemaining: 14, status: 'warning' })
    const tag = wrapper.find('span')
    expect(tag.exists()).toBe(true)
    expect(tag.attributes('data-color')).toBe('#faad14')
    expect(tag.text()).toContain('🔒')
    expect(tag.text()).toContain('SSL')
    expect(tag.text()).toContain('14d')
  })

  it('renders SSL critical badge with red color', () => {
    const wrapper = mountBadge({ type: 'ssl', daysRemaining: 3, status: 'critical' })
    const tag = wrapper.find('span')
    expect(tag.attributes('data-color')).toBe('#ff4d4f')
    expect(tag.text()).toContain('3d')
  })

  it('renders SSL expired badge with grey color', () => {
    const wrapper = mountBadge({ type: 'ssl', daysRemaining: -1, status: 'expired' })
    const tag = wrapper.find('span')
    expect(tag.attributes('data-color')).toBe('#8c8c8c')
    expect(tag.text()).toContain('SSL expired')
  })

  it('renders domain warning badge', () => {
    const wrapper = mountBadge({ type: 'domain', daysRemaining: 25, status: 'warning' })
    const tag = wrapper.find('span')
    expect(tag.exists()).toBe(true)
    expect(tag.text()).toContain('🌐')
    expect(tag.text()).toContain('Domain')
    expect(tag.text()).toContain('25d')
  })

  it('handles null daysRemaining gracefully', () => {
    const wrapper = mountBadge({ type: 'ssl', daysRemaining: null, status: 'critical' })
    const tag = wrapper.find('span')
    expect(tag.exists()).toBe(true)
    expect(tag.text()).toContain('SSL')
  })
})
