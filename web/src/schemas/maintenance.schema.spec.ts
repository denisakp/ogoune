import { describe, expect, it } from 'vitest'
import { maintenanceSchema, isValidCronExpression } from './maintenance.schema'

describe('maintenanceSchema — one_time', () => {
  it('accepts a future window where start < end', () => {
    const start = new Date(Date.now() + 3_600_000).toISOString()
    const end = new Date(Date.now() + 7_200_000).toISOString()
    const r = maintenanceSchema.safeParse({
      strategy: 'one_time',
      name: 'DB maintenance',
      start_at: start,
      end_at: end,
    })
    expect(r.success).toBe(true)
  })

  it('rejects a start in the past', () => {
    const past = new Date(Date.now() - 24 * 3_600_000).toISOString()
    const end = new Date(Date.now() + 3_600_000).toISOString()
    const r = maintenanceSchema.safeParse({
      strategy: 'one_time',
      name: 'X',
      start_at: past,
      end_at: end,
    })
    expect(r.success).toBe(false)
  })

  it('rejects end_at before start_at', () => {
    const start = new Date(Date.now() + 7_200_000).toISOString()
    const end = new Date(Date.now() + 3_600_000).toISOString()
    const r = maintenanceSchema.safeParse({
      strategy: 'one_time',
      name: 'X',
      start_at: start,
      end_at: end,
    })
    expect(r.success).toBe(false)
  })
})

describe('maintenanceSchema — recurring', () => {
  it('accepts a valid cron + duration', () => {
    const r = maintenanceSchema.safeParse({
      strategy: 'recurring',
      name: 'Nightly backup',
      cron: '0 2 * * 0',
      duration_minutes: 60,
    })
    expect(r.success).toBe(true)
  })

  it('rejects an invalid cron expression', () => {
    const r = maintenanceSchema.safeParse({
      strategy: 'recurring',
      name: 'X',
      cron: 'not a cron',
      duration_minutes: 60,
    })
    expect(r.success).toBe(false)
  })

  it('rejects duration < 5 or > 1440', () => {
    const base = { strategy: 'recurring' as const, name: 'X', cron: '0 2 * * *' }
    expect(maintenanceSchema.safeParse({ ...base, duration_minutes: 3 }).success).toBe(false)
    expect(maintenanceSchema.safeParse({ ...base, duration_minutes: 2000 }).success).toBe(false)
  })
})

describe('isValidCronExpression', () => {
  it('accepts a 5-field cron with digits, *, /', () => {
    expect(isValidCronExpression('*/15 * * * *')).toBe(true)
    expect(isValidCronExpression('0 2 * * 0')).toBe(true)
  })

  it('rejects expressions with the wrong number of fields', () => {
    expect(isValidCronExpression('0 2 * *')).toBe(false)
    expect(isValidCronExpression('0 2 * * * *')).toBe(false)
  })

  it('rejects expressions with disallowed characters', () => {
    expect(isValidCronExpression('zero two * * *')).toBe(false)
  })
})
