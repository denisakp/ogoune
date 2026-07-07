import { z } from 'zod'

export const accountSchema = z.object({
  first_name: z.string().min(1, 'Required').max(120, 'At most 120 characters'),
  last_name: z.string().min(1, 'Required').max(120, 'At most 120 characters'),
  email: z.string().email('Must be a valid email'),
  timezone: z.string().min(1, 'Required'),
})

export type AccountInput = z.infer<typeof accountSchema>
