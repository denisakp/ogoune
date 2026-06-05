import { describe, expect, it } from 'vitest'
import { deriveCrumbs } from './breadcrumb-helpers'

describe('AppBreadcrumb / deriveCrumbs', () => {
  it('returns empty when no matched record carries a breadcrumbLabel', () => {
    expect(deriveCrumbs([{ meta: {}, path: '/login' }])).toEqual([])
  })

  it('caps at 2 levels and uses the breadcrumbLabel from matched records', () => {
    const crumbs = deriveCrumbs([
      { meta: { breadcrumbLabel: 'Root' }, path: '/' },
      { meta: { breadcrumbLabel: 'Resources' }, path: '/monitors' },
      { meta: { breadcrumbLabel: 'Monitor' }, path: '/monitors/abc' },
    ])
    expect(crumbs).toHaveLength(2)
    expect(crumbs.map((c) => c.label)).toEqual(['Resources', 'Monitor'])
    expect(crumbs.map((c) => c.to)).toEqual(['/monitors', '/monitors/abc'])
  })

  it('ignores records without a breadcrumbLabel mixed in the middle', () => {
    const crumbs = deriveCrumbs([
      { meta: { breadcrumbLabel: 'A' }, path: '/a' },
      { meta: {}, path: '/skip' },
      { meta: { breadcrumbLabel: 'B' }, path: '/b' },
    ])
    expect(crumbs.map((c) => c.label)).toEqual(['A', 'B'])
  })
})
