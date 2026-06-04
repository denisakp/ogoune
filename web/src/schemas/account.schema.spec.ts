import { describe, expect, it } from 'vitest'
import { accountSchema } from './account.schema'

describe('accountSchema', () => {
  it('accepts a valid profile', () => {
    expect(
      accountSchema.safeParse({
        first_name: 'Ada',
        last_name: 'Lovelace',
        email: 'ada@x.test',
        timezone: 'Europe/Paris',
      }).success,
    ).toBe(true)
  })

  it('rejects empty first_name', () => {
    expect(
      accountSchema.safeParse({
        first_name: '',
        last_name: 'L',
        email: 'a@b.co',
        timezone: 'UTC',
      }).success,
    ).toBe(false)
  })

  it('rejects invalid email', () => {
    expect(
      accountSchema.safeParse({
        first_name: 'A',
        last_name: 'L',
        email: 'nope',
        timezone: 'UTC',
      }).success,
    ).toBe(false)
  })

  it('rejects empty timezone', () => {
    const r = accountSchema.safeParse({
      first_name: 'A',
      last_name: 'L',
      email: 'a@b.co',
      timezone: '',
    })
    expect(r.success).toBe(false)
  })
})
