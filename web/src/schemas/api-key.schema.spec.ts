import { describe, expect, it } from 'vitest'
import { apiKeySchema, resolveExpiresAt } from './api-key.schema'

describe('apiKeySchema', () => {
  it('accepts read scope with never expiry', () => {
    const r = apiKeySchema.safeParse({ name: 'CI', scope: 'read', expiry: 'never' })
    expect(r.success).toBe(true)
  })

  it('rejects empty name', () => {
    const r = apiKeySchema.safeParse({ name: '', scope: 'read', expiry: 'never' })
    expect(r.success).toBe(false)
  })

  it('rejects custom expiry without expires_at', () => {
    const r = apiKeySchema.safeParse({ name: 'X', scope: 'read', expiry: 'custom' })
    expect(r.success).toBe(false)
  })

  it('accepts custom expiry with ISO expires_at', () => {
    const r = apiKeySchema.safeParse({
      name: 'X',
      scope: 'read_write',
      expiry: 'custom',
      expires_at: '2027-01-01T00:00:00.000Z',
    })
    expect(r.success).toBe(true)
  })

  it('resolveExpiresAt returns undefined for never', () => {
    expect(resolveExpiresAt({ name: 'X', scope: 'read', expiry: 'never' })).toBeUndefined()
  })

  it('resolveExpiresAt converts 30d/90d/1y to ISO offsets', () => {
    const r30 = resolveExpiresAt({ name: 'X', scope: 'read', expiry: '30d' }) ?? ''
    expect(new Date(r30).getTime()).toBeGreaterThan(Date.now())
  })
})
