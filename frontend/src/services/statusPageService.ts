import axiosHelper from '../libs/axios.helper'
import type { StatusPageData, PublicMonitorDetail } from '@/types'

/**
 * Fetch status page data from /status endpoint
 * Returns comprehensive view of all monitored resources with 90-day uptime statistics
 */
export const fetchStatusPageData = async (): Promise<StatusPageData> => {
  const { data } = await axiosHelper.get<StatusPageData>('/status')
  return data
}

/**
 * Fetch public monitor detail data from /status/:id endpoint
 * Returns comprehensive view of a single monitor with 90-day statistics
 */
export const fetchStatusPageDataDetail = async (id: string): Promise<PublicMonitorDetail> => {
  const { data } = await axiosHelper.get<PublicMonitorDetail>(`/status/${id}`)
  return data
}
