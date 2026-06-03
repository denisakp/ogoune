import { getAuthenticatedClient, http, request } from '@/core/http/client'

export interface LoginRequest {
  email: string
  password: string
}

export interface LoginResponse {
  token: string
  email: string
  force_password_change?: boolean
  requires_2fa?: boolean
  password_initialized?: boolean
}

export interface Verify2FARequest {
  email: string
  otp: string
}

export interface VerifyResponse {
  email: string
  user_id: string
  name: string
  force_password_change: boolean
  two_factor_enabled: boolean
}

const SKIP_SUCCESS = { headers: { 'x-skip-success-toast': '1' } }
const SKIP_BOTH = {
  headers: { 'x-skip-success-toast': '1', 'x-skip-error-toast': '1' },
}

const authService = {
  /**
   * Authenticate user with email and password. Unauthenticated endpoint.
   */
  async login(email: string, password: string): Promise<LoginResponse> {
    return await request<LoginResponse>(http, 'auth/login', {
      method: 'POST',
      json: { email, password },
      ...SKIP_SUCCESS,
    })
  },

  /**
   * Verify current JWT token. Uses the authenticated client.
   * Errors are swallowed silently — the caller treats failure as "not authenticated".
   */
  async verify(): Promise<VerifyResponse> {
    return await request<VerifyResponse>(
      getAuthenticatedClient(),
      'auth/verify',
      SKIP_BOTH,
    )
  },

  /**
   * Verify 2FA OTP and issue token. Unauthenticated endpoint.
   */
  async verify2FA(payload: Verify2FARequest): Promise<LoginResponse> {
    return await request<LoginResponse>(http, 'auth/verify-2fa', {
      method: 'POST',
      json: payload,
      ...SKIP_SUCCESS,
    })
  },

  /**
   * Initialize password for first-time users. Unauthenticated endpoint.
   */
  async initializePassword(email: string, newPassword: string): Promise<LoginResponse> {
    return await request<LoginResponse>(http, 'auth/initialize-password', {
      method: 'POST',
      json: { email, new_password: newPassword },
      ...SKIP_SUCCESS,
    })
  },
}

export default authService
