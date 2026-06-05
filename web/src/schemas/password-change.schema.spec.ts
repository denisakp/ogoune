import { describe, expect, it } from 'vitest'
import { passwordChangeSchema } from './password-change.schema'

describe('passwordChangeSchema', () => {
  it('accepts matching new + confirm ≥12 chars', () => {
    expect(
      passwordChangeSchema.safeParse({
        current: 'old',
        new: 'aaaaaaaaaaaa',
        confirm: 'aaaaaaaaaaaa',
      }).success,
    ).toBe(true)
  })

  it('rejects new password < 12 chars', () => {
    expect(
      passwordChangeSchema.safeParse({ current: 'x', new: 'short', confirm: 'short' }).success,
    ).toBe(false)
  })

  it('rejects mismatched confirm', () => {
    const r = passwordChangeSchema.safeParse({
      current: 'x',
      new: 'aaaaaaaaaaaa',
      confirm: 'bbbbbbbbbbbb',
    })
    expect(r.success).toBe(false)
  })

  it('rejects empty current', () => {
    expect(
      passwordChangeSchema.safeParse({
        current: '',
        new: 'aaaaaaaaaaaa',
        confirm: 'aaaaaaaaaaaa',
      }).success,
    ).toBe(false)
  })
})
