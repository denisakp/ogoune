import axiosClient from '@/libs/axios.helper'
import type { CustomAxiosConfig } from '@/libs/axios.helper'

export interface UserProfile {
  email: string
  name: string
  user_id: string
  force_password_change: boolean
  two_factor_enabled: boolean
}

export interface UpdateProfileRequest {
  name: string
  email: string
}

export interface ChangePasswordRequest {
  current_password: string
  new_password: string
}

export interface ResetPasswordRequest {
  current_password: string
}

export interface Enable2FAResponse {
  secret: string
  qr_code: string
  backup_codes: string[]
}

export interface Confirm2FARequest {
  otp: string
  secret: string
}

export interface Disable2FARequest {
  otp: string
}

const accountService = {
  /**
   * Get user profile
   */
  async getProfile(): Promise<UserProfile> {
    const config: CustomAxiosConfig = {
      skipSuccessToast: true,
    }

    const response = await axiosClient.get<UserProfile>('/account/profile', config)
    return response.data
  },

  /**
   * Update user profile
   */
  async updateProfile(name: string, email: string): Promise<UserProfile> {
    const config: CustomAxiosConfig = {
      skipSuccessToast: false,
    }

    const response = await axiosClient.patch<UserProfile>(
      '/account/profile',
      { name, email },
      config,
    )
    return response.data
  },

  /**
   * Change password
   */
  async changePassword(currentPassword: string, newPassword: string): Promise<{ message: string }> {
    const config: CustomAxiosConfig = {
      skipSuccessToast: false,
    }

    const response = await axiosClient.post<{ message: string }>(
      '/account/change-password',
      { current_password: currentPassword, new_password: newPassword },
      config,
    )
    return response.data
  },

  /**
   * Reset password to default
   */
  async resetPassword(currentPassword: string): Promise<{ default_password: string }> {
    const config: CustomAxiosConfig = {
      skipSuccessToast: false,
    }

    const response = await axiosClient.post<{ default_password: string }>(
      '/account/reset-password',
      { current_password: currentPassword },
      config,
    )
    return response.data
  },

  /**
   * Enable 2FA and get QR code
   */
  async enable2FA(): Promise<Enable2FAResponse> {
    const config: CustomAxiosConfig = {
      skipSuccessToast: true,
    }

    const response = await axiosClient.post<Enable2FAResponse>('/account/2fa/enable', {}, config)
    return response.data
  },

  /**
   * Confirm 2FA setup with OTP
   */
  async confirm2FA({ otp, secret }: Confirm2FARequest): Promise<{ message: string }> {
    const config: CustomAxiosConfig = {
      skipSuccessToast: false,
    }

    const response = await axiosClient.post<{ message: string }>(
      '/account/2fa/confirm',
      { otp, secret },
      config,
    )
    return response.data
  },

  /**
   * Disable 2FA
   */
  async disable2FA(otp: string): Promise<{ message: string }> {
    const config: CustomAxiosConfig = {
      skipSuccessToast: false,
    }

    const response = await axiosClient.post<{ message: string }>(
      '/account/2fa/disable',
      { otp },
      config,
    )
    return response.data
  },
}

export default accountService
