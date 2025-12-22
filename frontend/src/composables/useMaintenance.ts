import { ref } from 'vue'
import type { CreateMaintenance, Maintenance, UpdateMaintenance } from '@/types'
import * as maintenanceService from '@/services/maintenanceService'

export function useMaintenance() {
  const maintenances = ref<Maintenance[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const statusFilter = ref<string>('')

  const loadMaintenances = async (status?: string) => {
    loading.value = true
    error.value = null
    statusFilter.value = status ?? ''
    try {
      maintenances.value = await maintenanceService.fetchMaintenances(status)
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load maintenances'
      throw err
    } finally {
      loading.value = false
    }
  }

  const addMaintenance = async (payload: CreateMaintenance) => {
    const created = await maintenanceService.createMaintenance(payload)
    maintenances.value.unshift(created)
    return created
  }

  const updateMaintenance = async (id: string, payload: UpdateMaintenance) => {
    const updated = await maintenanceService.updateMaintenance(id, payload)
    const index = maintenances.value.findIndex((m) => m.id === id)
    if (index !== -1) {
      maintenances.value[index] = updated
    }
    return updated
  }

  const deleteMaintenance = async (id: string) => {
    await maintenanceService.deleteMaintenance(id)
    maintenances.value = maintenances.value.filter((m) => m.id !== id)
  }

  const finishMaintenance = async (id: string) => {
    const finished = await maintenanceService.finishMaintenance(id)
    const index = maintenances.value.findIndex((m) => m.id === id)
    if (index !== -1) {
      maintenances.value[index] = finished
    }
    return finished
  }

  return {
    maintenances,
    loading,
    error,
    statusFilter,
    loadMaintenances,
    addMaintenance,
    updateMaintenance,
    deleteMaintenance,
    finishMaintenance,
  }
}
