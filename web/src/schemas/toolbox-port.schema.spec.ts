import { describe, expect, it } from 'vitest'
import { portScanSchema, portPresetValues } from '@/schemas/toolbox-port.schema'

describe('portScanSchema', () => {
  it('accepts a valid scan', () => {
    const r = portScanSchema.safeParse({ target: 'db-01', preset: 'common', ports: [22, 80], timeout_ms: 1000 })
    expect(r.success).toBe(true)
  })
  it('rejects > 100 ports', () => {
    const ports = Array.from({ length: 101 }, (_, i) => i + 1)
    expect(portScanSchema.safeParse({ target: 'h', preset: 'custom', ports, timeout_ms: 500 }).success).toBe(false)
  })
  it('rejects an out-of-range port', () => {
    expect(portScanSchema.safeParse({ target: 'h', preset: 'custom', ports: [70000], timeout_ms: 500 }).success).toBe(false)
  })
  it('rejects timeout outside 100–2000', () => {
    expect(portScanSchema.safeParse({ target: 'h', preset: 'common', ports: [80], timeout_ms: 50 }).success).toBe(false)
    expect(portScanSchema.safeParse({ target: 'h', preset: 'common', ports: [80], timeout_ms: 9000 }).success).toBe(false)
  })
  it('exposes preset port lists', () => {
    expect(portPresetValues.web).toContain(443)
  })
})
