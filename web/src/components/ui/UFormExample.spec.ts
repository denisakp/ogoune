import { describe, expect, it } from 'vitest'
import { resourceSchema } from '@/schemas/resource.schema'
import { ValidationError } from '@/core/errors'

/**
 * Specs for the reference form pattern. Vue mounting of `<UForm>` requires
 * NuxtUI's UForm runtime which is auto-imported; tests focus on the two
 * verifiable contracts:
 *   (a) the schema rejects/accepts inputs per spec 055 FR-011 (client-side)
 *   (b) the ValidationError carries fieldErrors in the documented shape
 *       per spec 055 FR-012 + PR-2 typed errors (server-side mapping)
 */

describe('UFormExample / client-side schema (FR-011, SC-005)', () => {
  it('rejects an empty name with a user-friendly message', () => {
    const result = resourceSchema.safeParse({
      type: 'http',
      name: '',
      interval: 60,
      url: 'https://example.com',
    })
    expect(result.success).toBe(false)
    if (!result.success) {
      const nameIssue = result.error.issues.find((i) => i.path[0] === 'name')
      expect(nameIssue?.message).toBe('Required')
    }
  })

  it('rejects an interval below 30 seconds', () => {
    const result = resourceSchema.safeParse({
      type: 'http',
      name: 'api',
      interval: 10,
      url: 'https://example.com',
    })
    expect(result.success).toBe(false)
  })

  it('rejects an invalid URL on an http monitor', () => {
    const result = resourceSchema.safeParse({
      type: 'http',
      name: 'api',
      interval: 60,
      url: 'not-a-url',
    })
    expect(result.success).toBe(false)
  })

  it('accepts a valid http monitor', () => {
    const result = resourceSchema.safeParse({
      type: 'http',
      name: 'api',
      interval: 60,
      url: 'https://example.com',
    })
    expect(result.success).toBe(true)
  })

  it('accepts a valid tcp monitor with host + port', () => {
    const result = resourceSchema.safeParse({
      type: 'tcp',
      name: 'db',
      interval: 60,
      host: 'db.local',
      port: 5432,
    })
    expect(result.success).toBe(true)
  })

  it('switches the required fields when type changes (discriminated union)', () => {
    // tcp branch requires host+port, NOT url
    const tcpResult = resourceSchema.safeParse({
      type: 'tcp',
      name: 'db',
      interval: 60,
      url: 'https://example.com', // url is irrelevant to tcp
    })
    expect(tcpResult.success).toBe(false) // missing host+port
  })
})

describe('UFormExample / server-side ValidationError (FR-012, SC-005)', () => {
  it('ValidationError exposes fieldErrors in Record<string, string[]> shape', () => {
    const err = new ValidationError('Validation failed', {
      name: ['Already taken'],
      url: ['Not reachable'],
    })
    expect(err.fieldErrors).toEqual({
      name: ['Already taken'],
      url: ['Not reachable'],
    })
  })

  it('fieldErrors entries can be mapped to formRef.setErrors() input shape', () => {
    const err = new ValidationError('Validation failed', {
      name: ['Already taken'],
      'config.url': ['Not reachable'],
    })
    const mapped = Object.entries(err.fieldErrors).map(([path, msgs]) => ({
      path,
      message: msgs[0] ?? 'Invalid',
    }))
    expect(mapped).toEqual([
      { path: 'name', message: 'Already taken' },
      { path: 'config.url', message: 'Not reachable' },
    ])
  })
})
