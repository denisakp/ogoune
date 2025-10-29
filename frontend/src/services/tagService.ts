import axiosHelper from '../libs/axios.helper'
import type { CreateTag, Tag } from '@/types'

/**
 * Fetch all tags
 */
export const fetchTags = async (): Promise<Tag[]> => {
  const { data } = await axiosHelper.get<Tag[]>('/tags')
  return data
}

/**
 * Fetch a single tag by ID
 */
export const fetchTag = async (id: string): Promise<Tag> => {
  const { data } = await axiosHelper.get<Tag>(`/tags/${id}`)
  return data
}

/**
 * Create a new tag
 */
export const createTag = async (tag: CreateTag): Promise<Tag> => {
  const { data } = await axiosHelper.post<Tag>('/tags', tag)
  return data
}

/**
 * Update an existing tag
 */
export const updateTag = async (id: string, tag: Partial<Tag>): Promise<Tag> => {
  const { data } = await axiosHelper.patch<Tag>(`/tags/${id}`, tag)
  return data
}

/**
 * Delete a tag
 */
export const deleteTag = async (id: string): Promise<void> => {
  await axiosHelper.delete(`/tags/${id}`)
}
