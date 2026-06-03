import { getAuthenticatedClient, request } from '@/core/http/client'
import type {
  CreateResource,
  Resource,
  UpdateResource,
  HourlyUptimeStat,
  SystemCapabilities,
} from '@/types'

const successMsg = (m: string) => ({ headers: { 'x-success-message': m } })

export const fetchResources = async (): Promise<Resource[]> => {
  return await request<Resource[]>(getAuthenticatedClient(), 'resources')
}

export const fetchResource = async (id: string, limit?: number): Promise<Resource> => {
  const searchParams = limit ? { limit } : undefined
  return await request<Resource>(getAuthenticatedClient(), `resources/${id}`, { searchParams })
}

export const fetchUptimeStats = async (
  id: string,
): Promise<{ resource_id: string; stats: HourlyUptimeStat[] }> => {
  return await request<{ resource_id: string; stats: HourlyUptimeStat[] }>(
    getAuthenticatedClient(),
    `resources/${id}/uptime-stats`,
  )
}

export const createResource = async (resource: CreateResource): Promise<Resource> => {
  return await request<Resource>(getAuthenticatedClient(), 'resources', {
    method: 'POST',
    json: resource,
    ...successMsg('Monitor created successfully'),
  })
}

export const updateResource = async (id: string, resource: UpdateResource): Promise<Resource> => {
  return await request<Resource>(getAuthenticatedClient(), `resources/${id}`, {
    method: 'PATCH',
    json: resource,
    ...successMsg('Monitor updated successfully'),
  })
}

export const deleteResource = async (id: string): Promise<void> => {
  await request<void>(getAuthenticatedClient(), `resources/${id}`, {
    method: 'DELETE',
    ...successMsg('Monitor deleted successfully'),
  })
}

export const pauseResource = async (id: string): Promise<Resource> => {
  return await request<Resource>(getAuthenticatedClient(), `resources/${id}/pause`, {
    method: 'POST',
    json: {},
    ...successMsg('Monitoring paused'),
  })
}

export const resumeResource = async (id: string): Promise<Resource> => {
  return await request<Resource>(getAuthenticatedClient(), `resources/${id}/resume`, {
    method: 'POST',
    json: {},
    ...successMsg('Monitoring resumed'),
  })
}

export const addTagsToResource = async (resourceId: string, tagIds: string[]): Promise<void> => {
  await request<void>(getAuthenticatedClient(), `resources/${resourceId}/tags`, {
    method: 'POST',
    json: { tag_ids: tagIds },
  })
}

export const removeTagFromResource = async (
  resourceId: string,
  tagId: string,
): Promise<void> => {
  await request<void>(getAuthenticatedClient(), `resources/${resourceId}/tags/${tagId}`, {
    method: 'DELETE',
  })
}

export const fetchCapabilities = async (): Promise<SystemCapabilities> => {
  return await request<SystemCapabilities>(getAuthenticatedClient(), 'system/capabilities')
}
