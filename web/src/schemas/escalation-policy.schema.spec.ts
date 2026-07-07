import { describe, expect, it } from 'vitest'
import { escalationPolicySchema, escalationStepSchema } from './escalation-policy.schema'

describe('escalationStepSchema', () => {
  it('accepts a step with delay 1..1440 and ≥1 channel', () => {
    expect(escalationStepSchema.safeParse({ delay_minutes: 5, channel_ids: ['c1'] }).success).toBe(
      true,
    )
  })

  it('rejects a step with zero channels', () => {
    expect(escalationStepSchema.safeParse({ delay_minutes: 5, channel_ids: [] }).success).toBe(
      false,
    )
  })

  it('rejects a step with delay out of range', () => {
    expect(escalationStepSchema.safeParse({ delay_minutes: 0, channel_ids: ['c'] }).success).toBe(
      false,
    )
    expect(
      escalationStepSchema.safeParse({ delay_minutes: 9999, channel_ids: ['c'] }).success,
    ).toBe(false)
  })
})

describe('escalationPolicySchema', () => {
  it('accepts a valid component-scoped policy', () => {
    const r = escalationPolicySchema.safeParse({
      name: 'Critical',
      scope: { kind: 'component', value: 'comp-1' },
      is_active: true,
      steps: [{ delay_minutes: 5, channel_ids: ['c1'] }],
    })
    expect(r.success).toBe(true)
  })

  it('accepts a valid tag-scoped policy', () => {
    const r = escalationPolicySchema.safeParse({
      name: 'Tag',
      scope: { kind: 'tag', value: 'prod' },
      is_active: true,
      steps: [{ delay_minutes: 10, channel_ids: ['c2'] }],
    })
    expect(r.success).toBe(true)
  })

  it('rejects name > 80 characters', () => {
    const r = escalationPolicySchema.safeParse({
      name: 'x'.repeat(81),
      scope: { kind: 'component', value: 'c' },
      steps: [{ delay_minutes: 5, channel_ids: ['c'] }],
    })
    expect(r.success).toBe(false)
  })

  it('rejects empty steps', () => {
    const r = escalationPolicySchema.safeParse({
      name: 'X',
      scope: { kind: 'component', value: 'c' },
      steps: [],
    })
    expect(r.success).toBe(false)
  })

  it('rejects > 5 steps', () => {
    const steps = Array.from({ length: 6 }, () => ({
      delay_minutes: 5,
      channel_ids: ['c'],
    }))
    const r = escalationPolicySchema.safeParse({
      name: 'X',
      scope: { kind: 'component', value: 'c' },
      steps,
    })
    expect(r.success).toBe(false)
  })

  it('rejects scope without value', () => {
    const r = escalationPolicySchema.safeParse({
      name: 'X',
      scope: { kind: 'tag', value: '' },
      steps: [{ delay_minutes: 5, channel_ids: ['c'] }],
    })
    expect(r.success).toBe(false)
  })
})
