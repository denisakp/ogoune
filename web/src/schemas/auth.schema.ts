import { z } from 'zod'

/**
 * Auth form schemas. Single-file exception to the per-entity-file rule —
 * the four auth screens form one semantic surface.
 *
 * Contract: specs/056-slice-auth-flows/contracts/form-schemas.md
 */

export const loginSchema = z.object({
  email: z.string().email('Must be a valid email'),
  password: z.string().min(1, 'Required'),
})

export const signupSchema = z
  .object({
    email: z.string().email('Must be a valid email'),
    password: z.string().min(12, 'At least 12 characters'),
    confirmPassword: z.string().min(1, 'Required'),
    newsletter: z.boolean().default(false),
  })
  .refine((d) => d.password === d.confirmPassword, {
    path: ['confirmPassword'],
    message: 'Passwords do not match',
  })

export const forgotPasswordSchema = z.object({
  email: z.string().email('Must be a valid email'),
})

export const resetPasswordSchema = z
  .object({
    token: z.string().min(1, 'Required'),
    password: z.string().min(12, 'At least 12 characters'),
    confirmPassword: z.string().min(1, 'Required'),
  })
  .refine((d) => d.password === d.confirmPassword, {
    path: ['confirmPassword'],
    message: 'Passwords do not match',
  })

export type LoginInput = z.infer<typeof loginSchema>
export type SignupInput = z.infer<typeof signupSchema>
export type ForgotPasswordInput = z.infer<typeof forgotPasswordSchema>
export type ResetPasswordInput = z.infer<typeof resetPasswordSchema>
