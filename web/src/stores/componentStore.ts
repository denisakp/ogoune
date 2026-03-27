import { defineStore } from 'pinia'
import { ref } from 'vue'

import * as componentService from '@/services/componentService'
import type { Component, CreateComponent, UpdateComponent } from '@/types'

export const useComponentStore = defineStore('component', () => {
  const components = ref<Component[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  const loadComponents = async () => {
    loading.value = true
    error.value = null
    try {
      components.value = await componentService.fetchComponents()
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load components'
      throw err
    } finally {
      loading.value = false
    }
  }

  const loadComponent = async (id: string): Promise<Component | null> => {
    loading.value = true
    error.value = null
    try {
      const component = await componentService.fetchComponent(id)

      const index = components.value.findIndex((c) => c.id === id)
      if (index !== -1) {
        components.value[index] = component
      } else {
        components.value.push(component)
      }

      return component
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load component'
      throw err
    } finally {
      loading.value = false
    }
  }

  const addComponent = async (component: CreateComponent): Promise<Component> => {
    loading.value = true
    error.value = null
    try {
      const newComponent = await componentService.createComponent(component)
      components.value.push(newComponent)
      return newComponent
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to create component'
      throw err
    } finally {
      loading.value = false
    }
  }

  const updateComponentData = async (
    id: string,
    component: UpdateComponent,
  ): Promise<Component> => {
    loading.value = true
    error.value = null
    try {
      const updated = await componentService.updateComponent(id, component)

      const index = components.value.findIndex((c) => c.id === id)
      if (index !== -1) {
        components.value[index] = updated
      }

      return updated
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to update component'
      throw err
    } finally {
      loading.value = false
    }
  }

  const removeComponent = async (id: string) => {
    loading.value = true
    error.value = null
    try {
      await componentService.deleteComponent(id)
      const index = components.value.findIndex((c) => c.id === id)
      if (index !== -1) {
        components.value.splice(index, 1)
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to delete component'
      throw err
    } finally {
      loading.value = false
    }
  }

  return {
    components,
    loading,
    error,
    loadComponents,
    loadComponent,
    addComponent,
    updateComponentData,
    removeComponent,
  }
})
