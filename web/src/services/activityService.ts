import { getAuthenticatedClient, request } from '@/core/http/client'
import type { MonitoringActivity } from '@/types'

/**
 * Fetch all monitoring activities with optional resource filter.
 * `resource_id` is omitted from the query string when undefined.
 */
export const fetchActivities = async (
  resourceId?: string,
): Promise<MonitoringActivity[]> => {
  const searchParams = resourceId ? { resource_id: resourceId } : undefined
  return await request<MonitoringActivity[]>(
    getAuthenticatedClient(),
    'monitoring-activities',
    { searchParams },
  )
}

/**
 * Fetch activities for a specific resource.
 */
export const fetchActivityByResource = async (
  resourceId: string,
): Promise<MonitoringActivity[]> => {
  return fetchActivities(resourceId)
}
