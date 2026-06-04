import { getAuthenticatedClient, getPublicClient, request } from '@/core/http/client'

const SKIP_SUCCESS = { headers: { 'x-skip-success-toast': '1' } }

export interface TwoFactorSetup {
  secret: string
  otpauth_url: string
}

export interface TwoFactorVerifyResponse {
  backup_codes: string[]
}

export interface TwoFactorResetResponse {
  token: string
  session_id: string
}

interface Envelope<T> {
  data: T
}

const twoFactorService = {
  async setup(): Promise<TwoFactorSetup> {
    const r = await request<Envelope<TwoFactorSetup>>(getAuthenticatedClient(), 'me/2fa/setup', {
      method: 'POST',
      json: {},
      ...SKIP_SUCCESS,
    })
    return r.data
  },

  async verify(code: string): Promise<TwoFactorVerifyResponse> {
    const r = await request<Envelope<TwoFactorVerifyResponse>>(
      getAuthenticatedClient(),
      'me/2fa/verify',
      { method: 'POST', json: { code } },
    )
    return r.data
  },

  async disable(code: string): Promise<void> {
    await request<void>(getAuthenticatedClient(), 'me/2fa/disable', {
      method: 'POST',
      json: { code },
    })
  },

  async requestReset(email: string): Promise<void> {
    await request<void>(getPublicClient(), 'auth/2fa/reset-request', {
      method: 'POST',
      json: { email },
      ...SKIP_SUCCESS,
    })
  },

  async confirmReset(token: string): Promise<TwoFactorResetResponse> {
    const r = await request<Envelope<TwoFactorResetResponse>>(getPublicClient(), 'auth/2fa/reset', {
      method: 'POST',
      json: { token },
    })
    return r.data
  },
}

export default twoFactorService
