import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { withStoreAction } from '@/utils/storeHelpers'
import * as tagService from '@/services/tagService'
import type { Tag, CreateTag } from '@/types'

export const useTagStore = defineStore('tag', () => {
  const tags = ref<Tag[]>([])
  const fetchLoading = ref(false)
  const fetchError = ref<string | null>(null)
  const addLoading = ref(false)
  const addError = ref<string | null>(null)
  const updateLoading = ref(false)
  const updateError = ref<string | null>(null)
  const deleteLoading = ref(false)
  const deleteError = ref<string | null>(null)
  const loading = computed(() => fetchLoading.value || addLoading.value || updateLoading.value || deleteLoading.value)
  const error = computed(() => fetchError.value ?? addError.value ?? updateError.value ?? deleteError.value)

  const fetchTags = () =>
    withStoreAction(fetchLoading, fetchError, async () => {
      tags.value = await tagService.fetchTags()
    })

  const addTag = (data: CreateTag) =>
    withStoreAction(addLoading, addError, async () => {
      const newTag = await tagService.createTag(data)
      tags.value.push(newTag)
      return newTag
    })

  const updateTag = (id: string, data: Partial<Tag>) =>
    withStoreAction(updateLoading, updateError, async () => {
      const updated = await tagService.updateTag(id, data)
      const index = tags.value.findIndex((t) => t.id === id)
      if (index !== -1) tags.value[index] = updated
      return updated
    })

  const deleteTag = (id: string) =>
    withStoreAction(deleteLoading, deleteError, async () => {
      await tagService.deleteTag(id)
      tags.value = tags.value.filter((t) => t.id !== id)
    })

  return {
    tags,
    loading,
    error,
    fetchLoading,
    fetchError,
    addLoading,
    addError,
    updateLoading,
    updateError,
    deleteLoading,
    deleteError,
    fetchTags,
    addTag,
    updateTag,
    deleteTag,
  }
})
