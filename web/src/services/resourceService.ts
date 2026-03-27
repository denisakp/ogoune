import type { AxiosRequestConfig } from 'axios'

import axiosHelper from '../libs/axios.helper'
import type { CreateResource, Resource, UpdateResource, HourlyUptimeStat } from '@/types'

interface CustomAxiosConfig extends AxiosRequestConfig {
  successMessage?: string
  skipSuccessToast?: boolean
  skipErrorToast?: boolean
}

/**
 * Fetch all resources (monitors)
 */
export const fetchResources = async (): Promise<Resource[]> => {
  const { data } = await axiosHelper.get<Resource[]>('/resources')
  return data
}

/**
 * Fetch a single resource by ID
 */
export const fetchResource = async (id: string, limit?: number): Promise<Resource> => {
  const params = limit ? { limit } : {}
  const { data } = await axiosHelper.get<Resource>(`/resources/${id}`, { params })
  return data
}

/**
 * Fetch uptime statistics for a resource
 */
export const fetchUptimeStats = async (
  id: string,
): Promise<{ resource_id: string; stats: HourlyUptimeStat[] }> => {
  const { data } = await axiosHelper.get<{ resource_id: string; stats: HourlyUptimeStat[] }>(
    `/resources/${id}/uptime-stats`,
  )
  return data
}

/**
 * Create a new resource
 */
export const createResource = async (resource: CreateResource): Promise<Resource> => {
  const config: CustomAxiosConfig = {
    successMessage: 'Monitor created successfully',
  }
  const { data } = await axiosHelper.post<Resource>('/resources', resource, config)
  return data
}

/**
 * Update an existing resource
 */
export const updateResource = async (id: string, resource: UpdateResource): Promise<Resource> => {
  const config: CustomAxiosConfig = {
    successMessage: 'Monitor updated successfully',
  }
  const { data } = await axiosHelper.patch<Resource>(`/resources/${id}`, resource, config)
  return data
}

/**
 * Delete a resource
 */
export const deleteResource = async (id: string): Promise<void> => {
  const config: CustomAxiosConfig = {
    successMessage: 'Monitor deleted successfully',
  }
  await axiosHelper.delete(`/resources/${id}`, config)
}

/**
 * Pause monitoring for a resource
 */
export const pauseResource = async (id: string): Promise<Resource> => {
  const config: CustomAxiosConfig = {
    successMessage: 'Monitoring paused',
  }
  const { data } = await axiosHelper.post<Resource>(`/resources/${id}/pause`, {}, config)
  return data
}

/**
 * Resume monitoring for a resource
 */
export const resumeResource = async (id: string): Promise<Resource> => {
  const config: CustomAxiosConfig = {
    successMessage: 'Monitoring resumed',
  }
  const { data } = await axiosHelper.post<Resource>(`/resources/${id}/resume`, {}, config)
  return data
}

/**
 * Add tags to a resource
 */
export const addTagsToResource = async (resourceId: string, tagIds: string[]): Promise<void> => {
  await axiosHelper.post(`/resources/${resourceId}/tags`, { tag_ids: tagIds })
}

/**
 * Remove a tag from a resource
 */
export const removeTagFromResource = async (resourceId: string, tagId: string): Promise<void> => {
  await axiosHelper.delete(`/resources/${resourceId}/tags/${tagId}`)
}
