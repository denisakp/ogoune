import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { withStoreAction } from '@/utils/storeHelpers'
import * as componentService from '@/services/componentService'
import type { Component, CreateComponent, UpdateComponent } from '@/types'

export const useComponentStore = defineStore('component', () => {
  const components = ref<Component[]>([])
  const loadLoading = ref(false)
  const loadError = ref<string | null>(null)
  const addLoading = ref(false)
  const addError = ref<string | null>(null)
  const updateLoading = ref(false)
  const updateError = ref<string | null>(null)
  const removeLoading = ref(false)
  const removeError = ref<string | null>(null)
  const loading = computed(() => loadLoading.value || addLoading.value || updateLoading.value || removeLoading.value)
  const error = computed(() => loadError.value ?? addError.value ?? updateError.value ?? removeError.value)

  const loadComponents = () =>
    withStoreAction(loadLoading, loadError, async () => {
      components.value = await componentService.fetchComponents()
    })

  const loadComponent = (id: string) =>
    withStoreAction(loadLoading, loadError, async () => {
      const component = await componentService.fetchComponent(id)
      const index = components.value.findIndex((c) => c.id === id)
      if (index !== -1) components.value[index] = component
      else components.value.push(component)
      return component
    })

  const addComponent = (component: CreateComponent) =>
    withStoreAction(addLoading, addError, async () => {
      const newComponent = await componentService.createComponent(component)
      components.value.push(newComponent)
      return newComponent
    })

  const updateComponentData = (id: string, component: UpdateComponent) =>
    withStoreAction(updateLoading, updateError, async () => {
      const updated = await componentService.updateComponent(id, component)
      const index = components.value.findIndex((c) => c.id === id)
      if (index !== -1) components.value[index] = updated
      return updated
    })

  const removeComponent = (id: string) =>
    withStoreAction(removeLoading, removeError, async () => {
      await componentService.deleteComponent(id)
      const index = components.value.findIndex((c) => c.id === id)
      if (index !== -1) components.value.splice(index, 1)
    })

  return {
    components,
    loading,
    error,
    loadLoading,
    loadError,
    addLoading,
    addError,
    updateLoading,
    updateError,
    removeLoading,
    removeError,
    loadComponents,
    loadComponent,
    addComponent,
    updateComponentData,
    removeComponent,
  }
})
