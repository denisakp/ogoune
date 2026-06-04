import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'

vi.mock('@/components/settings/account/ProfileSection.vue', () => ({
  default: { name: 'ProfileSection', template: '<div data-testid="profile" />' },
}))
vi.mock('@/components/settings/account/ChangePasswordSection.vue', () => ({
  default: { name: 'ChangePasswordSection', template: '<div data-testid="pwd" />' },
}))
vi.mock('@/components/settings/account/DangerZoneSection.vue', () => ({
  default: { name: 'DangerZoneSection', template: '<div data-testid="danger" />' },
}))

import AccountView from '../AccountView.vue'

describe('AccountView', () => {
  it('renders the 3 sub-sections', () => {
    const w = mount(AccountView)
    expect(w.find('[data-testid="profile"]').exists()).toBe(true)
    expect(w.find('[data-testid="pwd"]').exists()).toBe(true)
    expect(w.find('[data-testid="danger"]').exists()).toBe(true)
  })
})
