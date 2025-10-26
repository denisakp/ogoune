import axiosHelper from '../libs/axios.helper'
import type { MonitoringActivity } from '@/types'

/**
 * Fetch all monitoring activities with optional resource filter
 */
export const fetchActivities = async (resourceId?: string): Promise<MonitoringActivity[]> => {
  const params = resourceId ? { resource_id: resourceId } : {}
  const { data } = await axiosHelper.get<MonitoringActivity[]>('/monitoring-activities', {params})
  return data
}

/**
 * Fetch activities for a specific resource
 */
export const fetchActivityByResource = async (resourceId: string): Promise<MonitoringActivity[]> => {
  return fetchActivities(resourceId)
}
