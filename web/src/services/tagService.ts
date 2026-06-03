import { getAuthenticatedClient, request } from '@/core/http/client'
import type { CreateTag, Tag } from '@/types'

export const fetchTags = async (): Promise<Tag[]> => {
  return await request<Tag[]>(getAuthenticatedClient(), 'tags')
}

export const fetchTag = async (id: string): Promise<Tag> => {
  return await request<Tag>(getAuthenticatedClient(), `tags/${id}`)
}

export const createTag = async (tag: CreateTag): Promise<Tag> => {
  return await request<Tag>(getAuthenticatedClient(), 'tags', {
    method: 'POST',
    json: tag,
  })
}

export const updateTag = async (id: string, tag: Partial<Tag>): Promise<Tag> => {
  return await request<Tag>(getAuthenticatedClient(), `tags/${id}`, {
    method: 'PATCH',
    json: tag,
  })
}

/**
 * Delete a tag. Server returns 204 No Content — `request<T>` returns undefined.
 */
export const deleteTag = async (id: string): Promise<void> => {
  await request<void>(getAuthenticatedClient(), `tags/${id}`, { method: 'DELETE' })
}
