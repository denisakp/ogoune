import { defineStore } from 'pinia'
import { ref } from 'vue'
import * as tagService from '@/services/tagService'
import type { Tag, CreateTag } from '@/types'

export const useTagStore = defineStore('tag', () => {
  const tags = ref<Tag[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  /**
   * Fetch all tags from the API
   */
  const fetchTags = async () => {
    loading.value = true
    error.value = null
    try {
      tags.value = await tagService.fetchTags()
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load tags'
      console.error('Error loading tags:', err)
      throw err
    } finally {
      loading.value = false
    }
  }

  /**
   * Add a new tag
   */
  const addTag = async (data: CreateTag) => {
    try {
      const newTag = await tagService.createTag(data)
      tags.value.push(newTag)
      return newTag
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to create tag'
      console.error('Error creating tag:', err)
      throw err
    }
  }

  /**
   * Update an existing tag
   */
  const updateTag = async (id: string, data: Partial<Tag>) => {
    try {
      const updated = await tagService.updateTag(id, data)
      const index = tags.value.findIndex((t) => t.id === id)
      if (index !== -1) {
        tags.value[index] = updated
      }
      return updated
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to update tag'
      console.error('Error updating tag:', err)
      throw err
    }
  }

  /**
   * Delete a tag
   */
  const deleteTag = async (id: string) => {
    try {
      await tagService.deleteTag(id)
      tags.value = tags.value.filter((t) => t.id !== id)
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to delete tag'
      console.error('Error deleting tag:', err)
      throw err
    }
  }

  return {
    tags,
    loading,
    error,
    fetchTags,
    addTag,
    updateTag,
    deleteTag,
  }
})
