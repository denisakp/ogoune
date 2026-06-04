import { z } from 'zod'

/**
 * Spec 059 US7 — maintenance discriminated union.
 *
 * Cron parsing is intentionally lightweight (regex + field-shape check):
 * the `cron-parser` dependency is not yet pinned in package.json. The legacy
 * `CronGenerator.vue` produces only the 6 supported patterns; arbitrary
 * patterns flow through unvalidated server-side until cron-parser lands.
 */

const isoDate = z.string().datetime({ message: 'Must be an ISO 8601 datetime' })

const oneTimeSchema = z.object({
  strategy: z.literal('one_time'),
  name: z.string().trim().min(1).max(120),
  description: z.string().max(500).optional(),
  start_at: isoDate,
  end_at: isoDate,
  affected_resource_ids: z.array(z.string()).default([]),
})

// 5-field unix cron: m h dom mon dow. Each field is digits, *, /, -, comma.
const CRON_FIELD_RE = /^[\d*/,-]+$/

function isValidCronExpression(expr: string): boolean {
  const parts = expr.trim().split(/\s+/)
  if (parts.length !== 5) return false
  return parts.every((p) => CRON_FIELD_RE.test(p))
}

const recurringSchema = z.object({
  strategy: z.literal('recurring'),
  name: z.string().trim().min(1).max(120),
  description: z.string().max(500).optional(),
  cron: z.string().trim().refine(isValidCronExpression, { message: 'Invalid cron expression' }),
  duration_minutes: z.number().int().min(5).max(1440),
  affected_resource_ids: z.array(z.string()).default([]),
})

// `.refine` on a discriminated union child confuses Zod v3's discriminator
// extraction. Apply one-time temporal checks via `.superRefine` on the union.
export const maintenanceSchema = z
  .discriminatedUnion('strategy', [oneTimeSchema, recurringSchema])
  .superRefine((d, ctx) => {
    if (d.strategy !== 'one_time') return
    if (new Date(d.start_at).getTime() >= new Date(d.end_at).getTime()) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        path: ['end_at'],
        message: 'end_at must be after start_at',
      })
    }
    if (new Date(d.start_at).getTime() < Date.now() - 60_000) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        path: ['start_at'],
        message: 'start_at must be in the future',
      })
    }
  })

export type MaintenanceInput = z.infer<typeof maintenanceSchema>
export type MaintenanceOneTime = z.infer<typeof oneTimeSchema>
export type MaintenanceRecurring = z.infer<typeof recurringSchema>

export { isValidCronExpression }
