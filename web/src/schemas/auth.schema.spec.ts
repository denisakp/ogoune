import { describe, expect, it } from 'vitest'
import { loginSchema, signupSchema, forgotPasswordSchema, resetPasswordSchema } from './auth.schema'

describe('loginSchema', () => {
  it('accepts valid email + non-empty password', () => {
    const r = loginSchema.safeParse({ email: 'a@b.co', password: 'x' })
    expect(r.success).toBe(true)
  })
  it('rejects invalid email', () => {
    const r = loginSchema.safeParse({ email: 'nope', password: 'x' })
    expect(r.success).toBe(false)
  })
})

describe('signupSchema', () => {
  it('accepts valid input with matching passwords', () => {
    const r = signupSchema.safeParse({
      email: 'a@b.co',
      password: 'longenough12chars',
      confirmPassword: 'longenough12chars',
      newsletter: false,
    })
    expect(r.success).toBe(true)
  })
  it('rejects short password (< 12)', () => {
    const r = signupSchema.safeParse({
      email: 'a@b.co',
      password: 'short',
      confirmPassword: 'short',
      newsletter: false,
    })
    expect(r.success).toBe(false)
  })
  it('rejects mismatched confirmPassword with path = confirmPassword', () => {
    const r = signupSchema.safeParse({
      email: 'a@b.co',
      password: 'longenough12chars',
      confirmPassword: 'different1234',
      newsletter: false,
    })
    expect(r.success).toBe(false)
    if (!r.success) {
      const issue = r.error.issues.find((i) => i.path[0] === 'confirmPassword')
      expect(issue?.message).toBe('Passwords do not match')
    }
  })
})

describe('forgotPasswordSchema', () => {
  it('accepts valid email', () => {
    expect(forgotPasswordSchema.safeParse({ email: 'a@b.co' }).success).toBe(true)
  })
  it('rejects invalid email', () => {
    expect(forgotPasswordSchema.safeParse({ email: 'x' }).success).toBe(false)
  })
})

describe('resetPasswordSchema', () => {
  it('accepts valid token + matching strong passwords', () => {
    const r = resetPasswordSchema.safeParse({
      token: 'abc',
      password: 'longenough12chars',
      confirmPassword: 'longenough12chars',
    })
    expect(r.success).toBe(true)
  })
  it('rejects mismatched confirmPassword', () => {
    const r = resetPasswordSchema.safeParse({
      token: 'abc',
      password: 'longenough12chars',
      confirmPassword: 'mismatched111',
    })
    expect(r.success).toBe(false)
    if (!r.success) {
      const issue = r.error.issues.find((i) => i.path[0] === 'confirmPassword')
      expect(issue?.message).toBe('Passwords do not match')
    }
  })
})
