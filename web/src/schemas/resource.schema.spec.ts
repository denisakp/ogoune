import { describe, expect, it } from 'vitest'
import { resourceSchema } from './resource.schema'

const base = { name: 'api', interval: 60 }

describe('resourceSchema — happy + failure per monitor type', () => {
  it('http accepts valid input', () => {
    const r = resourceSchema.safeParse({ type: 'http', ...base, url: 'https://x.test' })
    expect(r.success).toBe(true)
  })
  it('http rejects invalid url', () => {
    const r = resourceSchema.safeParse({ type: 'http', ...base, url: 'not-a-url' })
    expect(r.success).toBe(false)
  })

  it('tcp accepts valid input', () => {
    const r = resourceSchema.safeParse({ type: 'tcp', ...base, host: 'db.test', port: 5432 })
    expect(r.success).toBe(true)
  })
  it('tcp rejects out-of-range port', () => {
    const r = resourceSchema.safeParse({ type: 'tcp', ...base, host: 'db.test', port: 99999 })
    expect(r.success).toBe(false)
  })

  it('dns accepts valid input', () => {
    const r = resourceSchema.safeParse({
      type: 'dns',
      ...base,
      host: 'example.com',
      record_type: 'A',
    })
    expect(r.success).toBe(true)
  })
  it('dns rejects missing host', () => {
    const r = resourceSchema.safeParse({ type: 'dns', ...base, record_type: 'A' })
    expect(r.success).toBe(false)
  })

  it('icmp accepts valid input', () => {
    const r = resourceSchema.safeParse({ type: 'icmp', ...base, host: 'example.com' })
    expect(r.success).toBe(true)
  })
  it('icmp rejects missing host', () => {
    const r = resourceSchema.safeParse({ type: 'icmp', ...base })
    expect(r.success).toBe(false)
  })

  it('keyword accepts valid input', () => {
    const r = resourceSchema.safeParse({
      type: 'keyword',
      ...base,
      url: 'https://x.test',
      keyword: 'OK',
    })
    expect(r.success).toBe(true)
  })
  it('keyword rejects empty keyword', () => {
    const r = resourceSchema.safeParse({
      type: 'keyword',
      ...base,
      url: 'https://x.test',
      keyword: '',
    })
    expect(r.success).toBe(false)
  })

  it('heartbeat accepts valid input', () => {
    const r = resourceSchema.safeParse({ type: 'heartbeat', ...base, grace_seconds: 120 })
    expect(r.success).toBe(true)
  })
  it('heartbeat rejects grace_seconds below floor', () => {
    const r = resourceSchema.safeParse({ type: 'heartbeat', ...base, grace_seconds: 5 })
    expect(r.success).toBe(false)
  })

  it('protocol accepts valid input', () => {
    const r = resourceSchema.safeParse({
      type: 'protocol',
      ...base,
      protocol: 'ssh',
      host: 'db.test',
      port: 22,
    })
    expect(r.success).toBe(true)
  })
  it('protocol rejects unknown protocol', () => {
    const r = resourceSchema.safeParse({
      type: 'protocol',
      ...base,
      protocol: 'gopher',
      host: 'db.test',
      port: 22,
    })
    expect(r.success).toBe(false)
  })
})

describe('resourceSchema — base extras (tags + notification_channels)', () => {
  it('accepts tags as string array', () => {
    const r = resourceSchema.safeParse({
      type: 'http',
      ...base,
      url: 'https://x.test',
      tags: ['t1', 't2'],
    })
    expect(r.success).toBe(true)
  })
  it('accepts notification_channels as string array', () => {
    const r = resourceSchema.safeParse({
      type: 'http',
      ...base,
      url: 'https://x.test',
      notification_channels: ['c1'],
    })
    expect(r.success).toBe(true)
  })
})
