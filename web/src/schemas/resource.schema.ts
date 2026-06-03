import { z } from 'zod'

/**
 * Reference schema for the Resource (monitor) entity. Demonstrates:
 *   - field-level validation (URL, port ranges, intervals)
 *   - composability (`baseResource`, `httpExtra`, `tcpExtra` exposed for `.merge()`)
 *   - conditional rules via `z.discriminatedUnion` keyed on `type`
 *
 * Contract: specs/055-slice-shared-components/contracts/form-pattern.md
 *
 * Slice 2 (ResourceForm migration) is the first real consumer; it MAY extend
 * the per-type extras as it discovers cases the oracle does not cover.
 */

export const monitorTypes = [
  'http',
  'tcp',
  'dns',
  'icmp',
  'heartbeat',
  'keyword',
  'protocol',
] as const

export const httpMethods = ['GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'HEAD'] as const

/** Common fields shared by every monitor type. */
export const baseResource = z.object({
  name: z.string().min(1, 'Required').max(120, 'At most 120 characters'),
  interval: z
    .number({ message: 'Required' })
    .int('Must be a whole number')
    .min(30, 'At least 30 seconds')
    .max(86_400, 'At most 24 hours'),
  confirmation_interval: z.number().int().optional(),
})

export const httpExtra = z.object({
  url: z.string().url('Must be a valid URL'),
  method: z.enum(httpMethods).default('GET'),
  expected_status: z
    .number()
    .int()
    .min(100)
    .max(599)
    .default(200),
  follow_redirects: z.boolean().default(true),
  headers: z.record(z.string(), z.string()).optional(),
})

export const tcpExtra = z.object({
  host: z.string().min(1, 'Required'),
  port: z.number().int().min(1).max(65_535),
})

export const dnsExtra = z.object({
  host: z.string().min(1, 'Required'),
  record_type: z.enum(['A', 'AAAA', 'CNAME', 'MX', 'TXT', 'NS']).default('A'),
})

export const icmpExtra = z.object({
  host: z.string().min(1, 'Required'),
})

export const heartbeatExtra = z.object({
  /** Grace period (seconds) before the heartbeat is considered missed. */
  grace_seconds: z.number().int().min(30).max(86_400),
})

export const keywordExtra = z.object({
  url: z.string().url('Must be a valid URL'),
  keyword: z.string().min(1, 'Required'),
  case_sensitive: z.boolean().default(false),
})

export const protocolExtra = z.object({
  protocol: z.enum(['imap', 'smtp', 'pop3', 'ssh', 'mysql', 'postgres']),
  host: z.string().min(1, 'Required'),
  port: z.number().int().min(1).max(65_535),
})

/** Discriminated union — type-safe per-monitor-type validation. */
export const resourceSchema = z.discriminatedUnion('type', [
  baseResource.merge(httpExtra).extend({ type: z.literal('http') }),
  baseResource.merge(tcpExtra).extend({ type: z.literal('tcp') }),
  baseResource.merge(dnsExtra).extend({ type: z.literal('dns') }),
  baseResource.merge(icmpExtra).extend({ type: z.literal('icmp') }),
  baseResource.merge(heartbeatExtra).extend({ type: z.literal('heartbeat') }),
  baseResource.merge(keywordExtra).extend({ type: z.literal('keyword') }),
  baseResource.merge(protocolExtra).extend({ type: z.literal('protocol') }),
])

export type ResourceInput = z.infer<typeof resourceSchema>
