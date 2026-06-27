import { describe, expect, it } from 'vitest'
import { sslCheckSchema } from '@/schemas/toolbox-ssl.schema'

describe('sslCheckSchema', () => {
  it('accepts a valid domain + port', () => {
    expect(sslCheckSchema.safeParse({ domain: 'example.com', port: 443 }).success).toBe(true)
  })
  it('rejects an invalid domain', () => {
    expect(sslCheckSchema.safeParse({ domain: 'bad domain', port: 443 }).success).toBe(false)
  })
  it('rejects an out-of-range port', () => {
    expect(sslCheckSchema.safeParse({ domain: 'example.com', port: 70000 }).success).toBe(false)
  })
})
