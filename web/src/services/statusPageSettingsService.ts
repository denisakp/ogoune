import { getAuthenticatedClient, request } from '@/core/http/client'
import type { StatusPageLogoSlot, StatusPageSettingsRequest, StatusPageSettingsResponse } from '@/types'

export const getStatusPageSettings = async (): Promise<StatusPageSettingsResponse> => {
  return await request<StatusPageSettingsResponse>(getAuthenticatedClient(), 'settings/statuspage')
}

export const updateStatusPageSettings = async (
  settings: StatusPageSettingsRequest,
): Promise<StatusPageSettingsResponse> => {
  return await request<StatusPageSettingsResponse>(
    getAuthenticatedClient(),
    'settings/statuspage',
    { method: 'PUT', json: settings },
  )
}

export const verifyStatusPageDomain = async (): Promise<StatusPageSettingsResponse> => {
  return await request<StatusPageSettingsResponse>(
    getAuthenticatedClient(),
    'settings/statuspage/verify-domain',
    { method: 'POST', json: {} },
  )
}

export const uploadStatusPageLogo = async (
  slot: StatusPageLogoSlot,
  file: File,
): Promise<StatusPageSettingsResponse> => {
  const body = new FormData()
  body.append('file', file)
  return await request<StatusPageSettingsResponse>(
    getAuthenticatedClient(),
    `settings/statuspage/logo?slot=${slot}`,
    { method: 'POST', body },
  )
}

export const deleteStatusPageLogo = async (slot: StatusPageLogoSlot): Promise<void> => {
  await request<unknown>(
    getAuthenticatedClient(),
    `settings/statuspage/logo?slot=${slot}`,
    { method: 'DELETE' },
  )
}
