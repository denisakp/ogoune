import { storeToRefs } from 'pinia'

import { useComponentStore } from '@/stores/componentStore'

export function useComponents() {
  const store = useComponentStore()
  const { components, loading, error } = storeToRefs(store)

  return {
    // Reactive state
    components,
    loading,
    error,
    // Store actions
    loadComponents: store.loadComponents,
    loadComponent: store.loadComponent,
    addComponent: store.addComponent,
    removeComponent: store.removeComponent,
    updateComponent: store.updateComponentData,
  }
}
