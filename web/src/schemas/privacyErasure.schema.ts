import { z } from 'zod'

const norm = (v: string) => v.trim().toLowerCase()

/**
 * Confirming an erasure (spec 095): the address typed back exactly, the current
 * password, and the two-factor code (6 digits or a backup code) when it is on.
 */
export function privacyErasureSchema(twoFactor: boolean, expectedEmail: string) {
  return z.object({
    confirm_email: z
      .string()
      .refine((v) => norm(v) === norm(expectedEmail), 'Type the address exactly'),
    password: z.string().min(1, 'Enter your current password'),
    code: twoFactor
      ? z
          .string()
          .min(1, 'Enter your two-factor code or a backup code')
          .max(20, 'That code is too long')
      : z.string().optional(),
  })
}

export type PrivacyErasureForm = z.infer<ReturnType<typeof privacyErasureSchema>>
