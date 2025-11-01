import axiosHelper from '../libs/axios.helper'
import type { PublicMonitorDetail } from '@/types'

/**
 * Fetch public monitor detail data from /status/:id endpoint
 * Returns comprehensive view of a single monitor with 90-day statistics
 */
export const fetchPublicMonitorDetail = async (id: string): Promise<PublicMonitorDetail> => {
  const { data } = await axiosHelper.get<PublicMonitorDetail>(`/status/${id}`)
  return data
}
