import { beforeEach, describe, expect, it, vi } from 'vitest'

import authService from '@/services/authService'
import axiosHelper from '@/libs/axios.helper'

vi.mock('@/libs/axios.helper', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    patch: vi.fn(),
    delete: vi.fn(),
  },
}))

describe('authService', () => {
  const mockPost = vi.mocked(axiosHelper.post)
  const mockGet = vi.mocked(axiosHelper.get)

  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('login', () => {
    it('sends POST to /auth/login with email and password', async () => {
      const loginResponse = {
        token: 'jwt-token-123',
        email: 'user@example.com',
      }
      mockPost.mockResolvedValue({ data: loginResponse })

      const result = await authService.login('user@example.com', 'secret123')

      expect(mockPost).toHaveBeenCalledOnce()
      expect(mockPost).toHaveBeenCalledWith(
        '/auth/login',
        { email: 'user@example.com', password: 'secret123' },
        expect.objectContaining({
          skipSuccessToast: true,
          skipErrorToast: false,
        }),
      )
      expect(result).toEqual(loginResponse)
    })

    it('propagates errors from rejected requests', async () => {
      const error = new Error('Network Error')
      mockPost.mockRejectedValue(error)

      await expect(authService.login('user@example.com', 'bad')).rejects.toThrow('Network Error')
    })
  })

  describe('verify', () => {
    it('sends GET to /auth/verify', async () => {
      const verifyResponse = {
        email: 'user@example.com',
        user_id: 'usr-1',
        name: 'Test User',
        force_password_change: false,
        two_factor_enabled: false,
      }
      mockGet.mockResolvedValue({ data: verifyResponse })

      const result = await authService.verify()

      expect(mockGet).toHaveBeenCalledOnce()
      expect(mockGet).toHaveBeenCalledWith(
        '/auth/verify',
        expect.objectContaining({
          skipSuccessToast: true,
          skipErrorToast: true,
        }),
      )
      expect(result).toEqual(verifyResponse)
    })

    it('propagates errors from rejected requests', async () => {
      const error = new Error('Unauthorized')
      mockGet.mockRejectedValue(error)

      await expect(authService.verify()).rejects.toThrow('Unauthorized')
    })
  })

  describe('verify2FA', () => {
    it('sends POST to /auth/verify-2fa with payload', async () => {
      const payload = { email: 'user@example.com', otp: '123456' }
      const response = { token: 'jwt-2fa-token', email: 'user@example.com' }
      mockPost.mockResolvedValue({ data: response })

      const result = await authService.verify2FA(payload)

      expect(mockPost).toHaveBeenCalledOnce()
      expect(mockPost).toHaveBeenCalledWith(
        '/auth/verify-2fa',
        payload,
        expect.objectContaining({
          skipSuccessToast: true,
          skipErrorToast: false,
        }),
      )
      expect(result).toEqual(response)
    })

    it('propagates errors from rejected requests', async () => {
      const error = new Error('Invalid OTP')
      mockPost.mockRejectedValue(error)

      await expect(
        authService.verify2FA({ email: 'user@example.com', otp: 'wrong' }),
      ).rejects.toThrow('Invalid OTP')
    })
  })

  describe('initializePassword', () => {
    it('sends POST to /auth/initialize-password with email and new_password', async () => {
      const response = { token: 'jwt-init-token', email: 'user@example.com' }
      mockPost.mockResolvedValue({ data: response })

      const result = await authService.initializePassword('user@example.com', 'newPass123')

      expect(mockPost).toHaveBeenCalledOnce()
      expect(mockPost).toHaveBeenCalledWith(
        '/auth/initialize-password',
        { email: 'user@example.com', new_password: 'newPass123' },
        expect.objectContaining({
          skipSuccessToast: true,
          skipErrorToast: false,
        }),
      )
      expect(result).toEqual(response)
    })
  })
})
