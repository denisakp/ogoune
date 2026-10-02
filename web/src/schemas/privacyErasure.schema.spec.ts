import { describe, expect, it } from 'vitest'
import { privacyErasureSchema } from './privacyErasure.schema'

const base = { confirm_email: 'Ada@Example.com ', password: 'pw' }

describe('privacyErasureSchema', () => {
  it('accepts the address regardless of case and surrounding spaces', () => {
    expect(privacyErasureSchema(false, 'ada@example.com').safeParse(base).success).toBe(true)
  })

  it('rejects a different address', () => {
    const r = privacyErasureSchema(false, 'bob@example.com').safeParse(base)
    expect(r.success).toBe(false)
    expect(r.error?.issues[0]?.message).toBe('Type the address exactly')
  })

  it('requires the password', () => {
    expect(privacyErasureSchema(false, 'ada@example.com').safeParse({ ...base, password: '' }).success).toBe(false)
  })

  it('needs no code without two-factor', () => {
    expect(privacyErasureSchema(false, 'ada@example.com').safeParse(base).success).toBe(true)
  })

  it('requires a code with two-factor, accepting 6 digits or a backup code, max 20 chars', () => {
    const s = privacyErasureSchema(true, 'ada@example.com')
    expect(s.safeParse(base).success).toBe(false)
    expect(s.safeParse({ ...base, code: '' }).success).toBe(false)
    expect(s.safeParse({ ...base, code: '123456' }).success).toBe(true)
    expect(s.safeParse({ ...base, code: 'abcd-efgh-jkmn' }).success).toBe(true)
    expect(s.safeParse({ ...base, code: 'x'.repeat(21) }).success).toBe(false)
  })
})
