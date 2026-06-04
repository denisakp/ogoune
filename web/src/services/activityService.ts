import { getAuthenticatedClient, request } from '@/core/http/client'
import type { MonitoringActivity } from '@/types'

interface ActivitiesEnvelope {
  activities: MonitoringActivity[]
  limit: number
  offset: number
}

/**
 * Fetch monitoring activities, optionally filtered by resource ID.
 * Backend envelope: `{ activities, limit, offset }` — unwrap to the array.
 */
export const fetchActivities = async (
  resourceId?: string,
  limit?: number,
  offset?: number,
): Promise<MonitoringActivity[]> => {
  const searchParams: Record<string, string | number> = {}
  if (resourceId) searchParams.resource_id = resourceId
  if (limit != null) searchParams.limit = limit
  if (offset != null) searchParams.offset = offset

  const res = await request<ActivitiesEnvelope>(getAuthenticatedClient(), 'monitoring-activities', {
    searchParams: Object.keys(searchParams).length > 0 ? searchParams : undefined,
  })
  return res?.activities ?? []
}

/**
 * Fetch activities for a specific resource.
 */
export const fetchActivityByResource = async (
  resourceId: string,
): Promise<MonitoringActivity[]> => {
  return fetchActivities(resourceId)
}
