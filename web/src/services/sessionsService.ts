import { getAuthenticatedClient, request } from '@/core/http/client'

const SKIP_SUCCESS = { headers: { 'x-skip-success-toast': '1' } }

export interface Session {
  id: string
  browser: string
  os: string
  ip: string
  location: string | null
  last_active_at: string
  is_current: boolean
  revoked_at: string | null
}

const sessionsService = {
  async list(): Promise<Session[]> {
    const r = await request<{ data: Session[] }>(
      getAuthenticatedClient(),
      'me/sessions',
      SKIP_SUCCESS,
    )
    return r.data
  },

  async revoke(id: string): Promise<void> {
    await request<void>(getAuthenticatedClient(), `me/sessions/${id}`, { method: 'DELETE' })
  },

  async revokeOthers(): Promise<void> {
    await request<void>(getAuthenticatedClient(), 'me/sessions/others', { method: 'DELETE' })
  },
}

export default sessionsService
