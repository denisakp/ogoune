import { getAuthenticatedClient, request } from '@/core/http/client'

const SKIP_SUCCESS = { headers: { 'x-skip-success-toast': '1' } }
const SKIP_BOTH = {
  headers: { 'x-skip-success-toast': '1', 'x-skip-error-toast': '1' },
}

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

export type APIKeyScope = 'read' | 'read_write'

export interface APIKey {
  id: string
  name: string
  key_prefix: string
  scope: APIKeyScope
  expires_at: string | null
  last_used_at: string | null
  last_used_ip: string
  is_active: boolean
  created_at: string
}

export interface CreateAPIKeyRequest {
  name: string
  scope: APIKeyScope
  expires_at?: string
}

export interface CreateAPIKeyResponse {
  id: string
  name: string
  key: string
  key_prefix: string
  scope: APIKeyScope
  expires_at: string | null
  created_at: string
}

const accountService = {
  async getProfile(): Promise<UserProfile> {
    return await request<UserProfile>(getAuthenticatedClient(), 'account/profile', SKIP_SUCCESS)
  },

  async updateProfile(name: string, email: string): Promise<UserProfile> {
    return await request<UserProfile>(getAuthenticatedClient(), 'account/profile', {
      method: 'PATCH',
      json: { name, email },
    })
  },

  async changePassword(currentPassword: string, newPassword: string): Promise<{ message: string }> {
    return await request<{ message: string }>(getAuthenticatedClient(), 'account/change-password', {
      method: 'POST',
      json: { current_password: currentPassword, new_password: newPassword },
    })
  },

  async resetPassword(currentPassword: string): Promise<{ default_password: string }> {
    return await request<{ default_password: string }>(
      getAuthenticatedClient(),
      'account/reset-password',
      { method: 'POST', json: { current_password: currentPassword } },
    )
  },

  async enable2FA(): Promise<Enable2FAResponse> {
    return await request<Enable2FAResponse>(getAuthenticatedClient(), 'account/2fa/enable', {
      method: 'POST',
      json: {},
      ...SKIP_SUCCESS,
    })
  },

  async confirm2FA({ otp, secret }: Confirm2FARequest): Promise<{ message: string }> {
    return await request<{ message: string }>(getAuthenticatedClient(), 'account/2fa/confirm', {
      method: 'POST',
      json: { otp, secret },
    })
  },

  async disable2FA(otp: string): Promise<{ message: string }> {
    return await request<{ message: string }>(getAuthenticatedClient(), 'account/2fa/disable', {
      method: 'POST',
      json: { otp },
    })
  },

  async createAPIKey(payload: CreateAPIKeyRequest): Promise<CreateAPIKeyResponse> {
    return await request<CreateAPIKeyResponse>(getAuthenticatedClient(), 'account/api-keys', {
      method: 'POST',
      json: payload,
    })
  },

  async listAPIKeys(): Promise<APIKey[]> {
    return await request<APIKey[]>(getAuthenticatedClient(), 'account/api-keys', SKIP_SUCCESS)
  },

  /**
   * Revoke an API key. Server returns a `{ message: string }` body on success
   * (Pattern B — body is consumed by caller, not a 204).
   */
  async revokeAPIKey(id: string): Promise<{ message: string }> {
    return await request<{ message: string }>(getAuthenticatedClient(), `account/api-keys/${id}`, {
      method: 'DELETE',
    })
  },

  async getOnboardingState(): Promise<{ status: 'pending' | 'done' }> {
    // Onboarding endpoint is best-effort: it may legitimately return 404 on
    // a fresh install (no row yet) — the composable defaults silently to
    // "no wizard". Silence the toast so reloads don't flash error modals.
    return await request<{ status: 'pending' | 'done' }>(
      getAuthenticatedClient(),
      'v1/me/onboarding-state',
      SKIP_BOTH,
    )
  },

  async deleteAccount(typedEmail: string): Promise<{ message: string }> {
    return await request<{ message: string }>(getAuthenticatedClient(), 'account', {
      method: 'DELETE',
      json: { confirm_email: typedEmail },
    })
  },

  async markOnboardingDone(): Promise<{ status: 'done' }> {
    return await request<{ status: 'done' }>(getAuthenticatedClient(), 'v1/me/onboarding-state', {
      method: 'PATCH',
      json: { status: 'done' },
      ...SKIP_SUCCESS,
    })
  },
}

export default accountService
