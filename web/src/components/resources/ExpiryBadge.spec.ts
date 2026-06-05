import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'

import ExpiryBadge from '@/components/resources/ExpiryBadge.vue'

const mountBadge = (props: InstanceType<typeof ExpiryBadge>['$props']) =>
  mount(ExpiryBadge, { props })

const hasColor = (wrapper: ReturnType<typeof mount>, color: string): boolean =>
  wrapper.html().includes(`text-${color}`)

describe('ExpiryBadge', () => {
  it('renders nothing when status is ok', () => {
    const wrapper = mountBadge({ type: 'ssl', daysRemaining: 20, status: 'ok' })
    // v-if hides the entire badge; rendered HTML is a Vue comment
    expect(wrapper.html()).not.toContain('SSL')
  })

  it('renders SSL warning badge with warning color', () => {
    const wrapper = mountBadge({ type: 'ssl', daysRemaining: 14, status: 'warning' })
    expect(hasColor(wrapper, 'warning')).toBe(true)
    expect(wrapper.text()).toContain('🔒')
    expect(wrapper.text()).toContain('SSL')
    expect(wrapper.text()).toContain('14d')
  })

  it('renders SSL critical badge with error color', () => {
    const wrapper = mountBadge({ type: 'ssl', daysRemaining: 3, status: 'critical' })
    expect(hasColor(wrapper, 'error')).toBe(true)
    expect(wrapper.text()).toContain('3d')
  })

  it('renders SSL expired badge with neutral color', () => {
    const wrapper = mountBadge({ type: 'ssl', daysRemaining: -1, status: 'expired' })
    // neutral may render as `text-inverted` / `text-default`; verify it's not success/warning/error
    expect(hasColor(wrapper, 'success')).toBe(false)
    expect(hasColor(wrapper, 'warning')).toBe(false)
    expect(hasColor(wrapper, 'error')).toBe(false)
    expect(wrapper.text()).toContain('SSL expired')
  })

  it('renders domain warning badge', () => {
    const wrapper = mountBadge({ type: 'domain', daysRemaining: 25, status: 'warning' })
    expect(hasColor(wrapper, 'warning')).toBe(true)
    expect(wrapper.text()).toContain('🌐')
    expect(wrapper.text()).toContain('Domain')
    expect(wrapper.text()).toContain('25d')
  })

  it('handles null daysRemaining gracefully', () => {
    const wrapper = mountBadge({ type: 'ssl', daysRemaining: null, status: 'critical' })
    expect(wrapper.text()).toContain('SSL')
  })
})
