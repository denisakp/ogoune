import { ref } from 'vue'
import {
  listIncidentUpdates,
  createIncidentUpdate,
  updateIncidentUpdate,
  deleteIncidentUpdate,
  type IncidentUpdate,
  type IncidentUpdatePayload,
} from '@/services/incidentUpdateService'

export function useIncidentUpdates(incidentID: string) {
  const updates = ref<IncidentUpdate[]>([])
  const loading = ref(false)
  const error = ref<Error | null>(null)

  async function refresh() {
    loading.value = true
    error.value = null
    try {
      updates.value = await listIncidentUpdates(incidentID)
    } catch (e) {
      error.value = e instanceof Error ? e : new Error(String(e))
    } finally {
      loading.value = false
    }
  }

  async function add(payload: IncidentUpdatePayload) {
    await createIncidentUpdate(incidentID, payload)
    await refresh()
  }

  async function edit(updateID: string, payload: IncidentUpdatePayload) {
    await updateIncidentUpdate(incidentID, updateID, payload)
    await refresh()
  }

  async function remove(updateID: string) {
    await deleteIncidentUpdate(incidentID, updateID)
    await refresh()
  }

  return { updates, loading, error, refresh, add, edit, remove }
}
