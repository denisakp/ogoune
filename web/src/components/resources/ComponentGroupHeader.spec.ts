import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'

import ComponentGroupHeader from './ComponentGroupHeader.vue'

const stubs = { UIcon: { template: '<span />' } }

describe('ComponentGroupHeader', () => {
  it('renders component name + total count', () => {
    const w = mount(ComponentGroupHeader, {
      global: { stubs },
      props: {
        component: { id: 'c1', name: 'API Cluster' },
        resources: [
          { id: 'r1', status: 'up' },
          { id: 'r2', status: 'down' },
        ],
      },
    })
    expect(w.text()).toContain('API Cluster')
    expect(w.text()).toContain('2')
  })

  it('renders only status dots with count > 0', () => {
    const w = mount(ComponentGroupHeader, {
      global: { stubs },
      props: {
        component: { id: 'c1', name: 'X' },
        resources: [
          { id: 'r1', status: 'up' },
          { id: 'r2', status: 'up' },
          { id: 'r3', status: 'down' },
        ],
      },
    })
    expect(w.html()).toContain('bg-emerald-500')
    expect(w.html()).toContain('bg-red-500')
    expect(w.html()).not.toContain('bg-amber-500')
  })

  it('renders "Standalone Resources" when component is null', () => {
    const w = mount(ComponentGroupHeader, {
      global: { stubs },
      props: { component: null, resources: [] },
    })
    expect(w.text()).toContain('Standalone Resources')
  })

  it('emits update:collapsed on click', async () => {
    const w = mount(ComponentGroupHeader, {
      global: { stubs },
      props: { component: { id: 'c1', name: 'X' }, resources: [], collapsed: false },
    })
    await w.find('button').trigger('click')
    expect(w.emitted('update:collapsed')).toEqual([[true]])
  })
})
