import { getAuthenticatedClient, request } from '@/core/http/client'
import type { StatsSummary } from '@/types'

/**
 * Fetch statistics summary for a given time range.
 * @param range - Time range (2h, 24h, 7d, 30d)
 */
export const fetchStatsSummary = async (range: string): Promise<StatsSummary> => {
  return await request<StatsSummary>(getAuthenticatedClient(), 'stats/summary', {
    searchParams: { range },
  })
}
