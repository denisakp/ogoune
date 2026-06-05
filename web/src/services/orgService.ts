import { getAuthenticatedClient, request } from '@/core/http/client'

const SKIP_SUCCESS = { headers: { 'x-skip-success-toast': '1' } }

export interface OrgGeneral {
  name: string
  logo_url: string | null
  timezone: string
  date_format: string
}

interface Envelope<T> {
  data: T
}

const orgService = {
  async getGeneral(): Promise<OrgGeneral> {
    // Backend endpoint will be filled in a follow-up. Fall back to a safe
    // default so the view renders during the chantier.
    try {
      const r = await request<Envelope<OrgGeneral>>(
        getAuthenticatedClient(),
        'org/general',
        SKIP_SUCCESS,
      )
      return r.data
    } catch {
      return {
        name: 'Ogoune',
        logo_url: null,
        timezone: 'UTC',
        date_format: 'YYYY-MM-DD',
      }
    }
  },

  async updateGeneral(payload: Partial<OrgGeneral>): Promise<OrgGeneral> {
    const r = await request<Envelope<OrgGeneral>>(getAuthenticatedClient(), 'org/general', {
      method: 'PATCH',
      json: payload,
    })
    return r.data
  },

  async uploadLogo(file: File): Promise<{ logo_url: string }> {
    const form = new FormData()
    form.append('logo', file)
    const r = await request<Envelope<{ logo_url: string }>>(
      getAuthenticatedClient(),
      'org/general/logo',
      { method: 'POST', body: form },
    )
    return r.data
  },
}

export default orgService
