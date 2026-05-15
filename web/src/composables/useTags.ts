import { storeToRefs } from 'pinia'

import { useTagStore } from '@/stores/tagStore'

export function useTags() {
  const store = useTagStore()
  const { tags, loading, error } = storeToRefs(store)

  return {
    // Reactive state
    tags,
    loading,
    error,
    // Store actions
    loadTags: store.fetchTags,
    addTag: store.addTag,
    updateTag: store.updateTag,
    deleteTag: store.deleteTag,
  }
}
