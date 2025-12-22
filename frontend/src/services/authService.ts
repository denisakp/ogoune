import axiosClient from '@/libs/axios.helper'
import type { CustomAxiosConfig } from '@/libs/axios.helper'

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

const authService = {
  /**
   * Authenticate user with email and password
   */
  async login(email: string, password: string): Promise<LoginResponse> {
    const config: CustomAxiosConfig = {
      skipSuccessToast: true, // Don't show success toast for login
      skipErrorToast: false,
    }

    const response = await axiosClient.post<LoginResponse>(
      '/auth/login',
      { email, password },
      config,
    )

    return response.data
  },

  /**
   * Verify current JWT token
   */
  async verify(): Promise<VerifyResponse> {
    const config: CustomAxiosConfig = {
      skipSuccessToast: true,
      skipErrorToast: true, // Don't show error toast for token verification
    }

    const response = await axiosClient.get<VerifyResponse>('/auth/verify', config)
    return response.data
  },

  /**
   * Verify 2FA OTP and issue token
   */
  async verify2FA(payload: Verify2FARequest): Promise<LoginResponse> {
    const config: CustomAxiosConfig = {
      skipSuccessToast: true,
      skipErrorToast: false,
    }

    const response = await axiosClient.post<LoginResponse>('/auth/verify-2fa', payload, config)
    return response.data
  },

  /**
   * Initialize password for first-time users
   */
  async initializePassword(email: string, newPassword: string): Promise<LoginResponse> {
    const config: CustomAxiosConfig = {
      skipSuccessToast: true,
      skipErrorToast: false,
    }

    const response = await axiosClient.post<LoginResponse>(
      '/auth/initialize-password',
      { email, new_password: newPassword },
      config,
    )

    return response.data
  },
}

export default authService
