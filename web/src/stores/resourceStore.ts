import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { withStoreAction } from '@/utils/storeHelpers'
import * as resourceService from '@/services/resourceService'
import type { CreateResource, Resource, UpdateResource, SystemCapabilities } from '@/types'

export const useResourceStore = defineStore('resource', () => {
  const resources = ref<Resource[]>([])
  const capabilities = ref<SystemCapabilities | null>(null)
  const fetchLoading = ref(false)
  const fetchError = ref<string | null>(null)
  const mutateLoading = ref(false)
  const mutateError = ref<string | null>(null)
  const loading = computed(() => fetchLoading.value || mutateLoading.value)
  const error = computed(() => fetchError.value ?? mutateError.value)

  function upsert(resource: Resource) {
    const i = resources.value.findIndex((r) => r.id === resource.id)
    if (i !== -1) resources.value[i] = resource
    else resources.value.push(resource)
  }

  const loadResources = () =>
    withStoreAction(fetchLoading, fetchError, async () => {
      resources.value = await resourceService.fetchResources()
    })
  const loadResource = (id: string) =>
    withStoreAction(fetchLoading, fetchError, async () => {
      const r = await resourceService.fetchResource(id)
      upsert(r)
      return r
    })
  const loadResourceWithResponseTimes = (id: string, limit = 50) =>
    withStoreAction(fetchLoading, fetchError, async () => {
      const r = await resourceService.fetchResource(id, limit)
      upsert(r)
      return r
    })
  const loadUptimeStats = (id: string) =>
    withStoreAction(
      fetchLoading,
      fetchError,
      async () => (await resourceService.fetchUptimeStats(id)).stats,
    )

  const mutateAndReload = (fn: () => Promise<unknown>) =>
    withStoreAction(mutateLoading, mutateError, async () => {
      await fn()
      await loadResources()
      return true
    }).catch(() => false)

  const addResource = (resource: CreateResource) =>
    mutateAndReload(() => resourceService.createResource(resource))
  const removeResource = (id: string) => mutateAndReload(() => resourceService.deleteResource(id))
  const updateResourceData = (id: string, updates: UpdateResource) =>
    mutateAndReload(() => resourceService.updateResource(id, updates))
  const pauseMonitoring = (id: string) => mutateAndReload(() => resourceService.pauseResource(id))
  const resumeMonitoring = (id: string) => mutateAndReload(() => resourceService.resumeResource(id))

  const loadCapabilities = async () => {
    try {
      capabilities.value = await resourceService.fetchCapabilities()
    } catch {
      /* graceful degradation */
    }
  }

  const activeResources = computed(() =>
    resources.value.filter((r) => !(r.type === 'heartbeat' && r.waiting)),
  )
  const upCount = computed(() => activeResources.value.filter((r) => r.status === 'up').length)
  const downCount = computed(() => activeResources.value.filter((r) => r.status === 'down').length)
  const waitingCount = computed(
    () => resources.value.filter((r) => r.type === 'heartbeat' && r.waiting).length,
  )

  return {
    resources,
    loading,
    error,
    capabilities,
    fetchLoading,
    fetchError,
    mutateLoading,
    mutateError,
    upCount,
    downCount,
    waitingCount,
    loadResources,
    loadResource,
    loadResourceWithResponseTimes,
    loadUptimeStats,
    addResource,
    removeResource,
    updateResourceData,
    pauseMonitoring,
    resumeMonitoring,
    loadCapabilities,
  }
})
