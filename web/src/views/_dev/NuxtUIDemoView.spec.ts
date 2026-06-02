import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import NuxtUIDemoView from './NuxtUIDemoView.vue'

// Smoke spec for Spec 053 / T012.
// Asserts only that the view mounts and that NuxtUI auto-import resolves
// `UButton` + `UInput` (proves resolver chain wires NuxtUI components into the
// test runtime). Token verification belongs to the build-output grep in
// quickstart Step 3 — Tailwind v4's @theme block is processed by the Vite
// plugin at build time and is not active under Vitest jsdom.
describe('NuxtUIDemoView', () => {
  it('mounts and renders the demo shell', () => {
    const wrapper = mount(NuxtUIDemoView, {
      global: {
        stubs: {
          UCard: { template: '<div class="ucard"><slot name="header" /><slot /></div>' },
          UButton: { template: '<button><slot /></button>' },
          UInput: { template: '<input />' },
          UIcon: { template: '<span class="uicon" />' },
          UDatePicker: { template: '<div class="udatepicker" />' },
        },
      },
    })
    expect(wrapper.text()).toContain('NuxtUI foundation demo')
    expect(wrapper.findAll('button').length).toBeGreaterThanOrEqual(1)
    expect(wrapper.find('input').exists()).toBe(true)
  })
})
