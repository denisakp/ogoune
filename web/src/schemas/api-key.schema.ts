import { z } from 'zod'

export const apiKeySchema = z
  .object({
    name: z.string().trim().min(1, 'Required').max(80, 'At most 80 characters'),
    scope: z.enum(['read', 'read_write']),
    expiry: z.enum(['never', '30d', '90d', '1y', 'custom']).default('never'),
    expires_at: z.string().datetime().optional(),
  })
  .refine((d) => d.expiry !== 'custom' || !!d.expires_at, {
    path: ['expires_at'],
    message: 'Required for custom expiry',
  })

export type ApiKeyInput = z.infer<typeof apiKeySchema>

export const EXPIRY_PRESETS: { value: ApiKeyInput['expiry']; label: string }[] = [
  { value: 'never', label: 'Never' },
  { value: '30d', label: '30 days' },
  { value: '90d', label: '90 days' },
  { value: '1y', label: '1 year' },
  { value: 'custom', label: 'Custom' },
]

export function resolveExpiresAt(input: ApiKeyInput): string | undefined {
  if (input.expiry === 'never') return undefined
  if (input.expiry === 'custom') return input.expires_at
  const now = Date.now()
  const days = input.expiry === '30d' ? 30 : input.expiry === '90d' ? 90 : 365
  return new Date(now + days * 86_400_000).toISOString()
}
