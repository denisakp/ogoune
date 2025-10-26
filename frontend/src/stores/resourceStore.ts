import { defineStore } from 'pinia'
import { ref } from 'vue'

import * as resourceService from '@/services/resourceService'
import type { CreateResource, Resource, UpdateResource } from '@/types'

export const useResourceStore = defineStore('resource', () => { 
  const resources = ref<Resource[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  
  const loadResources = async () => { 
    loading.value = true
    error.value = null
    try { 
      resources.value = await resourceService.fetchResources()
    } catch (err) { 
      error.value = err instanceof Error ? err.message : 'Failed to load resources'
      console.error('Error loading resources:', err)
    } finally { 
      loading.value = false
    } 
  }
  
  const addResource = async (resource: CreateResource) => {
    try {
      await resourceService.createResource(resource)
      await loadResources()
      return true
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to create resource'
      console.error('Error creating resource:', err)
      return false
    }
  }
  
  const removeResource = async (id: string) => {
    try {
      await resourceService.deleteResource(id)
      await loadResources()
      return true
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to delete resource'
      console.error('Error deleting resource:', err)
      return false
    }
  }
  
  const updateResourceData = async (id: string, updates: UpdateResource) => {
    try {
      await resourceService.updateResource(id, updates)
      await loadResources()
      return true
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to update resource'
      console.error('Error updating resource:', err)
      return false
    }
  }
  
  const pauseMonitoring = async (id: string) => {
    try {
      await resourceService.pauseResource(id)
      await loadResources()
      return true
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to pause monitoring'
      console.error('Error pausing monitoring:', err)
      return false
    }
  }
  
  const resumeMonitoring = async (id: string) => {
    try {
      await resourceService.resumeResource(id)
      await loadResources()
      return true
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to resume monitoring'
      console.error('Error resuming monitoring:', err)
      return false
    }
  }
  
  return {
    resources,
    loading,
    error,
    loadResources,
    addResource,
    removeResource,
    updateResourceData,
    pauseMonitoring,
    resumeMonitoring,
  }
});
