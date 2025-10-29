import { defineStore } from 'pinia'
import { ref } from 'vue'

import * as resourceService from '@/services/resourceService'
import type { CreateResource, Resource, UpdateResource, HourlyUptimeStat } from '@/types'

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

  const loadResource = async (id: string): Promise<Resource | null> => {
    loading.value = true
    error.value = null
    try {
      // Asynchronously fetch the single resource data from the service using its ID.
      const resource = await resourceService.fetchResource(id)

      // Find the index of the resource in our local `resources` array.
      // This is to check if we already have a version of this resource stored locally.
      const index = resources.value.findIndex((r) => r.id === id)

      // If the resource is found in the local array (index is not -1)...
      if (index !== -1) {
        // ...update the item at that index with the newly fetched resource data.
        // This keeps our local cache in sync.
        resources.value[index] = resource
      } else {
        // If the resource is not in the local array...
        // ...add the newly fetched resource to the end of the array.
        resources.value.push(resource)
      }

      // Return the fetched resource so the calling component can use it.
      return resource
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load resource'
      console.error('Error loading resource:', err)
      return null
    } finally {
      loading.value = false
    }
  }

  const loadResourceWithResponseTimes = async (
    id: string,
    limit = 50,
  ): Promise<Resource | null> => {
    loading.value = true
    error.value = null
    try {
      // Fetch resource with response times included
      const resource = await resourceService.fetchResource(id, limit)

      // Update or add to local resources array
      const index = resources.value.findIndex((r) => r.id === id)
      if (index !== -1) {
        resources.value[index] = resource
      } else {
        resources.value.push(resource)
      }

      return resource
    } catch (err) {
      error.value =
        err instanceof Error ? err.message : 'Failed to load resource with response times'
      console.error('Error loading resource with response times:', err)
      return null
    } finally {
      loading.value = false
    }
  }

  const loadUptimeStats = async (id: string): Promise<HourlyUptimeStat[] | null> => {
    loading.value = true
    error.value = null
    try {
      const response = await resourceService.fetchUptimeStats(id)
      return response.stats
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load uptime stats'
      console.error('Error loading uptime stats:', err)
      return null
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

  const testNotification = async (resourceId: string) => {
    try {
      await resourceService.testNotification(resourceId)
      return true
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to send test notification'
      console.error('Error sending test notification:', err)
      return false
    }
  }

  return {
    resources,
    loading,
    error,
    loadResources,
    loadResource,
    loadResourceWithResponseTimes,
    loadUptimeStats,
    addResource,
    removeResource,
    updateResourceData,
    pauseMonitoring,
    resumeMonitoring,
    testNotification,
  }
})
