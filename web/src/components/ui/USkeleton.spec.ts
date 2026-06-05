import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import USkeleton from './USkeleton.vue'

describe('USkeleton', () => {
  it.each(['text', 'circle', 'rect', 'table-row', 'card'] as const)(
    'renders variant=%s',
    (variant) => {
      const wrapper = mount(USkeleton, { props: { variant } })
      expect(wrapper.attributes('data-variant')).toBe(variant)
      expect(wrapper.classes().some((c) => c.includes('animate-pulse'))).toBe(true)
    },
  )

  it('accepts custom width / height', () => {
    const wrapper = mount(USkeleton, {
      props: { variant: 'rect', width: '100px', height: '20px' },
    })
    const style = wrapper.attributes('style') ?? ''
    expect(style).toContain('width: 100px')
    expect(style).toContain('height: 20px')
  })
})
