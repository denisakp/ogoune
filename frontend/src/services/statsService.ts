import axiosHelper from '../libs/axios.helper'
import type { StatsSummary } from '@/types'

/**
 * Fetch statistics summary for a given time range
 * @param range - Time range (2h, 24h, 7d, 30d)
 */
export const fetchStatsSummary = async (range: string): Promise<StatsSummary> => {
  const { data } = await axiosHelper.get<StatsSummary>('/stats/summary', {
    params: { range },
  })
  return data
}
