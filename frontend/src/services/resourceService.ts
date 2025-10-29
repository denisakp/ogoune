import axiosHelper from '../libs/axios.helper'
import type { CreateResource, Resource, UpdateResource } from '@/types'

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
export const fetchResource = async (id: string): Promise<Resource> => {
  const { data } = await axiosHelper.get<Resource>(`/resources/${id}`)
  return data
}

/**
 * Create a new resource
 */
export const createResource = async (resource: CreateResource): Promise<Resource> => {
  const { data } = await axiosHelper.post<Resource>('/resources', resource)
  return data
}

/**
 * Update an existing resource
 */
export const updateResource = async (id: string, resource: UpdateResource): Promise<Resource> => {
  const { data } = await axiosHelper.patch<Resource>(`/resources/${id}`, resource)
  return data
}

/**
 * Delete a resource
 */
export const deleteResource = async (id: string): Promise<void> => {
  await axiosHelper.delete(`/resources/${id}`)
}

/**
 * Pause monitoring for a resource
 */
export const pauseResource = async (id: string): Promise<Resource> => {
  const { data } = await axiosHelper.post<Resource>(`/resources/${id}/pause`, {})
  return data
}

/**
 * Resume monitoring for a resource
 */
export const resumeResource = async (id: string): Promise<Resource> => {
  const { data } = await axiosHelper.post<Resource>(`/resources/${id}/resume`, {})
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

/**
 *  send test notification for a resource
 * @param resourceId
 */
export const testNotification = async (resourceId: string) => {
  const { data } = await axiosHelper.post(`/notifications/test`, { resource_id: resourceId })
  return data
}
