import { defineStore } from 'pinia'
import { ref } from 'vue'

import * as integrationService from '@/services/integrationService'
import type { Integration, CreateIntegration } from '@/types'

export const useIntegrationStore = defineStore('integration', () => {
  const integrations = ref<Integration[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  /**
   * Fetch all integrations from the API
   */
  const fetchIntegrations = async () => {
    loading.value = true
    error.value = null
    try {
      integrations.value = await integrationService.fetchIntegrations()
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load integrations'
      console.error('Error loading integrations:', err)
      throw err
    } finally {
      loading.value = false
    }
  }

  /**
   * Add a new integration
   */
  const addIntegration = async (data: CreateIntegration) => {
    try {
      const newIntegration = await integrationService.createIntegration(data)
      integrations.value.push(newIntegration)
      return newIntegration
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to create integration'
      console.error('Error creating integration:', err)
      throw err
    }
  }

  /**
   * Update an existing integration
   */
  const updateIntegration = async (id: string, data: Partial<Integration>) => {
    try {
      const updated = await integrationService.updateIntegration(id, data)
      const index = integrations.value.findIndex((i) => i.id === id)
      if (index !== -1) {
        integrations.value[index] = updated
      }
      return updated
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to update integration'
      console.error('Error updating integration:', err)
      throw err
    }
  }

  /**
   * Delete an integration
   */
  const deleteIntegration = async (id: string) => {
    try {
      await integrationService.deleteIntegration(id)
      integrations.value = integrations.value.filter((i) => i.id !== id)
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to delete integration'
      console.error('Error deleting integration:', err)
      throw err
    }
  }

  return {
    integrations,
    loading,
    error,
    fetchIntegrations,
    addIntegration,
    updateIntegration,
    deleteIntegration,
  }
})
