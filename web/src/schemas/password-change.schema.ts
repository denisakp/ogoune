import { z } from 'zod'

export const passwordChangeSchema = z
  .object({
    current: z.string().min(1, 'Required'),
    new: z.string().min(12, 'At least 12 characters'),
    confirm: z.string().min(1, 'Required'),
  })
  .refine((d) => d.new === d.confirm, {
    path: ['confirm'],
    message: 'Passwords do not match',
  })

export type PasswordChangeInput = z.infer<typeof passwordChangeSchema>
