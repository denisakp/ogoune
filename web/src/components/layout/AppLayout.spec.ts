import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import AppLayout from './AppLayout.vue'

describe('AppLayout', () => {
  it('mounts and renders the default slot inside the main column', () => {
    const wrapper = mount(AppLayout, {
      slots: { default: '<div data-test="slot">page-content</div>' },
      global: {
        stubs: {
          AppSidebar: { template: '<aside data-test="sidebar" />' },
          AppTopbar: { template: '<header data-test="topbar" />' },
        },
      },
    })
    expect(wrapper.find('[data-test="sidebar"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="topbar"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="slot"]').text()).toBe('page-content')
  })
})
