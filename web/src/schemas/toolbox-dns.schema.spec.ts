import { describe, expect, it } from 'vitest'
import { dnsLookupSchema } from '@/schemas/toolbox-dns.schema'

describe('dnsLookupSchema', () => {
  it('accepts a valid lookup', () => {
    const r = dnsLookupSchema.safeParse({
      domain: 'example.com',
      record_types: ['A', 'MX'],
      resolver: 'cloudflare',
    })
    expect(r.success).toBe(true)
  })

  it('rejects an empty domain', () => {
    const r = dnsLookupSchema.safeParse({ domain: '', record_types: ['A'], resolver: 'google' })
    expect(r.success).toBe(false)
  })

  it('rejects an invalid domain', () => {
    const r = dnsLookupSchema.safeParse({ domain: 'not a domain', record_types: ['A'], resolver: 'google' })
    expect(r.success).toBe(false)
  })

  it('requires at least one record type', () => {
    const r = dnsLookupSchema.safeParse({ domain: 'example.com', record_types: [], resolver: 'google' })
    expect(r.success).toBe(false)
  })

  it('requires custom_resolver when resolver is custom', () => {
    const r = dnsLookupSchema.safeParse({
      domain: 'example.com',
      record_types: ['A'],
      resolver: 'custom',
    })
    expect(r.success).toBe(false)
  })

  it('accepts custom resolver with an IP', () => {
    const r = dnsLookupSchema.safeParse({
      domain: 'example.com',
      record_types: ['A'],
      resolver: 'custom',
      custom_resolver: '192.0.2.1',
    })
    expect(r.success).toBe(true)
  })
})
