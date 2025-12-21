import axiosHelper from '../libs/axios.helper'
import type { StatusPageSettingsRequest, StatusPageSettingsResponse } from '@/types'

/**
 * Fetch status page settings
 */
export const getStatusPageSettings = async (): Promise<StatusPageSettingsResponse> => {
  const { data } = await axiosHelper.get<StatusPageSettingsResponse>('/settings/statuspage')
  return data
}

/**
 * Update status page settings
 */
export const updateStatusPageSettings = async (
  settings: StatusPageSettingsRequest,
): Promise<StatusPageSettingsResponse> => {
  const { data } = await axiosHelper.put<StatusPageSettingsResponse>(
    '/settings/statuspage',
    settings,
  )
  return data
}
