import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'
import type { DashboardScope } from '@/types'

const resourcesRef = await vi.hoisted(async () => {
  const vue = await import('vue')
  return vue.ref<unknown[]>([])
})
const tagsRef = await vi.hoisted(async () => {
  const vue = await import('vue')
  return vue.ref<unknown[]>([])
})
const componentsRef = await vi.hoisted(async () => {
  const vue = await import('vue')
  return vue.ref<unknown[]>([])
})

vi.mock('@/stores/resourceStore', () => ({
  useResourceStore: () => ({ resources: resourcesRef.value }),
}))
vi.mock('@/stores/tagStore', () => ({
  useTagStore: () => ({ tags: tagsRef.value }),
}))
vi.mock('@/stores/componentStore', () => ({
  useComponentStore: () => ({ components: componentsRef.value }),
}))

import DashboardScopeResolver from './DashboardScopeResolver.vue'

const stubs = {
  UIcon: { template: '<span />', props: ['name'] },
}

const initialScope: DashboardScope = { mode: 'tag', payload: { tagIds: [] } }

describe('DashboardScopeResolver (spec 070 / US2)', () => {
  beforeEach(() => {
    resourcesRef.value = []
    tagsRef.value = [
      { id: 't1', name: 'prod' },
      { id: 't2', name: 'dev' },
    ]
    componentsRef.value = []
  })

  it('renders the 4 scope tabs (tag / component / type / manual)', async () => {
    const wrapper = mount(DashboardScopeResolver, {
      global: { stubs },
      props: { modelValue: initialScope },
    })
    await nextTick()
    for (const m of ['tag', 'component', 'type', 'manual'] as const) {
      expect(wrapper.find(`[data-testid="scope-tab-${m}"]`).exists()).toBe(true)
    }
  })

  it('default tab is tag-picker', async () => {
    const wrapper = mount(DashboardScopeResolver, {
      global: { stubs },
      props: { modelValue: initialScope },
    })
    await nextTick()
    expect(wrapper.find('[data-testid="scope-tag-picker"]').exists()).toBe(true)
  })

  it('clicking the type tab swaps the picker and emits scope.mode=type with empty tagIds', async () => {
    const wrapper = mount(DashboardScopeResolver, {
      global: { stubs },
      props: { modelValue: initialScope },
    })
    await wrapper.find('[data-testid="scope-tab-type"]').trigger('click')
    await nextTick()
    expect(wrapper.find('[data-testid="scope-type-picker"]').exists()).toBe(true)
    const emits = wrapper.emitted('update:modelValue') as DashboardScope[][]
    const last = emits[emits.length - 1]![0]
    expect(last.mode).toBe('type')
    expect(last.payload.tagIds ?? []).toEqual([])
  })

  it('reports zero match by default with empty payload', async () => {
    const wrapper = mount(DashboardScopeResolver, {
      global: { stubs },
      props: { modelValue: initialScope },
    })
    await nextTick()
    expect(wrapper.find('[data-testid="scope-match-count"]').text()).toContain('0')
  })
})
