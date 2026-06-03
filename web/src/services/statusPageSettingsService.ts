import { getAuthenticatedClient, request } from '@/core/http/client'
import type { StatusPageSettingsRequest, StatusPageSettingsResponse } from '@/types'

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
