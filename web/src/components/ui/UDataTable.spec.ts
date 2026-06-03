import { describe, expect, it } from 'vitest'
import { computeTotalPages } from './data-table-helpers'

describe('UDataTable / computeTotalPages', () => {
  it('returns 1 when no pagination is provided', () => {
    expect(computeTotalPages()).toBe(1)
  })

  it('returns ceil(total / perPage)', () => {
    expect(computeTotalPages({ page: 1, perPage: 10, total: 25 })).toBe(3)
    expect(computeTotalPages({ page: 1, perPage: 10, total: 10 })).toBe(1)
    expect(computeTotalPages({ page: 1, perPage: 10, total: 1 })).toBe(1)
  })

  it('never returns less than 1', () => {
    expect(computeTotalPages({ page: 1, perPage: 10, total: 0 })).toBe(1)
  })
})
