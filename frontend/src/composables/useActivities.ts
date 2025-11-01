import { ref } from 'vue'

import * as activityService from '@/services/activityService'
import type { MonitoringActivity } from '@/types'

export function useActivities() {
  const activities = ref<MonitoringActivity[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  const loadActivities = async (resourceId?: string) => {
    loading.value = true
    error.value = null
    try {
      activities.value = await activityService.fetchActivities(resourceId)
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load activities'
      console.error('Error loading activities:', err)
    } finally {
      loading.value = false
    }
  }

  return {
    activities,
    loading,
    error,
    loadActivities,
  }
}
