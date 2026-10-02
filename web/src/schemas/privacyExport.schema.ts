import { z } from 'zod'

/**
 * Re-authentication before the personal-data export (spec 094): the current
 * password always, the 6-digit two-factor code when two-factor is on -- the
 * same factors sign-in asks for.
 */
export function privacyExportSchema(twoFactorEnabled: boolean) {
  return z.object({
    password: z.string().min(1, 'Enter your current password'),
    code: twoFactorEnabled
      ? z.string().regex(/^\d{6}$/, 'Enter the 6-digit code from your authenticator app')
      : z.string().optional(),
  })
}

export type PrivacyExportForm = z.infer<ReturnType<typeof privacyExportSchema>>
