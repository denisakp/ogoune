import { z } from 'zod'

export const escalationStepSchema = z.object({
  delay_minutes: z.number().int().min(1, 'At least 1 minute').max(1440, 'At most 1440 minutes'),
  channel_ids: z.array(z.string()).min(1, 'Pick at least one channel'),
})

const scopeSchema = z.discriminatedUnion('kind', [
  z.object({ kind: z.literal('component'), value: z.string().min(1, 'Required') }),
  z.object({ kind: z.literal('tag'), value: z.string().min(1, 'Required') }),
])

export const escalationPolicySchema = z.object({
  name: z.string().trim().min(1, 'Required').max(80, 'At most 80 characters'),
  scope: scopeSchema,
  is_active: z.boolean().default(true),
  steps: z.array(escalationStepSchema).min(1, 'At least one step').max(5, 'At most 5 steps'),
})

export type EscalationPolicyInput = z.infer<typeof escalationPolicySchema>
export type EscalationStepInput = z.infer<typeof escalationStepSchema>

export function emptyStep(): EscalationStepInput {
  return { delay_minutes: 5, channel_ids: [] }
}

export function emptyPolicy(): EscalationPolicyInput {
  return {
    name: '',
    scope: { kind: 'component', value: '' },
    is_active: true,
    steps: [emptyStep()],
  }
}
