import axiosClient from '@/libs/axios.helper'
import type { CustomRequestConfig } from '@/libs/axios.helper'

export interface LoginRequest {
  email: string
  password: string
}

export interface LoginResponse {
  token: string
  email: string
}

export interface VerifyResponse {
  email: string
}

const authService = {
  /**
   * Authenticate user with email and password
   */
  async login(email: string, password: string): Promise<LoginResponse> {
    const config: CustomRequestConfig = {
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
    const config: CustomRequestConfig = {
      skipSuccessToast: true,
      skipErrorToast: true, // Don't show error toast for token verification
    }

    const response = await axiosClient.get<VerifyResponse>('/auth/verify', config)
    return response.data
  },
}

export default authService
