import { describe, expect, it } from 'vitest'
import { whoisSchema } from '@/schemas/toolbox-whois.schema'

describe('whoisSchema', () => {
  it('accepts a valid domain', () => {
    expect(whoisSchema.safeParse({ domain: 'example.com' }).success).toBe(true)
  })
  it('rejects an empty domain', () => {
    expect(whoisSchema.safeParse({ domain: '' }).success).toBe(false)
  })
})
