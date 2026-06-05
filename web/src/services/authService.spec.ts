import { describe, expect, it } from 'vitest'
import { http, HttpResponse } from 'msw'

import authService from '@/services/authService'
import { UnauthorizedError } from '@/core/errors'
import { server } from '@/test/msw/server'

describe('authService', () => {
  describe('login', () => {
    it('sends POST to /auth/login with email and password', async () => {
      const expected = { token: 'jwt-token-123', email: 'user@example.com' }
      const cap: {
        received: { email: string; password: string } | null
        headers: Headers | null
      } = { received: null, headers: null }

      server.use(
        http.post('*/auth/login', async ({ request }) => {
          cap.received = (await request.json()) as typeof cap.received
          cap.headers = request.headers
          return HttpResponse.json(expected)
        }),
      )

      const result = await authService.login('user@example.com', 'secret123')

      expect(cap.received).toEqual({ email: 'user@example.com', password: 'secret123' })
      expect(cap.headers?.get('x-skip-success-toast')).toBe('1')
      expect(result).toEqual(expected)
    })

    it('propagates errors from rejected requests', async () => {
      server.use(http.post('*/auth/login', () => HttpResponse.json({}, { status: 401 })))
      await expect(authService.login('user@example.com', 'bad')).rejects.toBeInstanceOf(
        UnauthorizedError,
      )
    })
  })

  describe('verify', () => {
    it('sends GET to /auth/verify with skip-error-toast header', async () => {
      const expected = {
        email: 'user@example.com',
        user_id: 'usr-1',
        name: 'Test User',
        force_password_change: false,
        two_factor_enabled: false,
      }
      const captured: { headers: Headers | null } = { headers: null }
      server.use(
        http.get('*/auth/verify', ({ request }) => {
          captured.headers = request.headers
          return HttpResponse.json(expected)
        }),
      )

      const result = await authService.verify()

      expect(captured.headers?.get('x-skip-success-toast')).toBe('1')
      expect(captured.headers?.get('x-skip-error-toast')).toBe('1')
      expect(result).toEqual(expected)
    })

    it('propagates errors from rejected requests', async () => {
      server.use(http.get('*/auth/verify', () => HttpResponse.json({}, { status: 401 })))
      await expect(authService.verify()).rejects.toBeInstanceOf(UnauthorizedError)
    })
  })

  describe('verify2FA', () => {
    it('sends POST to /auth/verify-2fa with payload', async () => {
      const expected = { token: 'jwt-2fa-token', email: 'user@example.com' }
      const cap: { received: { email: string; otp: string } | null } = { received: null }
      server.use(
        http.post('*/auth/verify-2fa', async ({ request }) => {
          cap.received = (await request.json()) as typeof cap.received
          return HttpResponse.json(expected)
        }),
      )

      const result = await authService.verify2FA({
        email: 'user@example.com',
        otp: '123456',
      })

      expect(cap.received).toEqual({ email: 'user@example.com', otp: '123456' })
      expect(result).toEqual(expected)
    })

    it('propagates errors from rejected requests', async () => {
      server.use(
        http.post('*/auth/verify-2fa', () =>
          HttpResponse.json({ message: 'Invalid OTP' }, { status: 401 }),
        ),
      )
      await expect(
        authService.verify2FA({ email: 'user@example.com', otp: 'wrong' }),
      ).rejects.toBeInstanceOf(UnauthorizedError)
    })
  })

  describe('initializePassword', () => {
    it('sends POST to /auth/initialize-password with email and new_password', async () => {
      const expected = { token: 'jwt-init-token', email: 'user@example.com' }
      const cap: { received: { email: string; new_password: string } | null } = { received: null }
      server.use(
        http.post('*/auth/initialize-password', async ({ request }) => {
          cap.received = (await request.json()) as typeof cap.received
          return HttpResponse.json(expected)
        }),
      )

      const result = await authService.initializePassword('user@example.com', 'newPass123')

      expect(cap.received).toEqual({ email: 'user@example.com', new_password: 'newPass123' })
      expect(result).toEqual(expected)
    })
  })
})
