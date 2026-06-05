<script setup lang="ts">
import { onMounted, onUnmounted, ref, computed, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useToast } from '@nuxt/ui/composables/useToast'
import { useResourceStore } from '@/stores/resourceStore'
import { useMonitorLive } from '@/composables/useMonitorLive'
import ResourceModal from '@/components/resources/ResourceModal.vue'
import ResourceStatusCard from './ResourceStatusCard.vue'
import ResourceHeartbeat from './ResourceHeartbeat.vue'
import ResourcePerformance from './ResourcePerformance.vue'
import ResourceIncidents from './ResourceIncidents.vue'
import ResourceDetails from './ResourceDetails.vue'
import type { Resource } from '@/types'

const router = useRouter()
const route = useRoute()
const store = useResourceStore()
const toast = useToast()

const resource = ref<Resource | null>(null)
const timeRange = ref<'24h' | '7d' | '30d' | '365d'>('24h')
const showEditModal = ref(false)
const nowTs = ref(Date.now())
let timer: number | undefined

const resourceId = computed(() => route.params.id as string)

const {
  liveData,
  isLoading: isLiveLoading,
  lastUpdated,
  error: liveError,
  isTerminated,
  refresh,
  startPolling,
  stopPolling,
} = useMonitorLive(
  resourceId.value,
  () => resource.value?.interval,
  () => resource.value?.waiting === true,
)

const lastUpdatedRelative = computed(() => {
  if (!lastUpdated.value) return ''
  const delta = Math.max(0, Math.floor((nowTs.value - lastUpdated.value.getTime()) / 1000))
  return delta < 5 ? 'just now' : `${delta}s ago`
})

const isHeartbeat = computed(() => resource.value?.type === 'heartbeat')
const isProtocol = computed(() => resource.value?.type === 'protocol')

onMounted(async () => {
  timer = window.setInterval(() => {
    nowTs.value = Date.now()
  }, 1000)
  stopPolling()
  await loadResource()
  await refresh()
  startPolling()
})

onUnmounted(() => {
  stopPolling()
  if (timer) window.clearInterval(timer)
})

watch(liveData, (snapshot) => {
  if (!snapshot || !resource.value) return
  const incoming = snapshot.resource as Resource
  resource.value = {
    ...resource.value,
    ...incoming,
    tags: incoming.tags ? [...incoming.tags] : resource.value.tags,
    incidents: incoming.incidents ? [...incoming.incidents] : resource.value.incidents,
    response_times: resource.value.response_times,
  }
})

const loadResource = async () => {
  if (!resourceId.value) {
    toast.add({ title: 'Resource ID not found', color: 'error' })
    return
  }
  try {
    const data = await store.loadResourceWithResponseTimes(resourceId.value, 50)
    if (data) resource.value = data
    else toast.add({ title: 'Failed to load resource', color: 'error' })
  } catch (err) {
    toast.add({
      title: err instanceof Error ? err.message : 'Failed to load resource',
      color: 'error',
    })
  }
}

const handlePauseResource = async () => {
  if (!resource.value) return
  try {
    await store.pauseMonitoring(resource.value.id)
    await loadResource()
  } catch {
    /* interceptor handles */
  }
}

const openEditModal = () => {
  showEditModal.value = true
}
const handleEditSubmit = async () => {
  showEditModal.value = false
  await loadResource()
}
const goBack = () => {
  router.back()
}
</script>

<template>
  <div class="p-6">
    <div v-if="store.fetchLoading && !resource" class="text-center py-12">
      <UIcon name="i-lucide-loader-circle" class="size-8 animate-spin text-primary-500" />
    </div>

    <template v-else-if="resource">
      <UButton color="neutral" variant="ghost" icon="i-lucide-arrow-left" class="mb-4" @click="goBack">
        Monitoring
      </UButton>

      <!-- Header -->
      <div class="flex justify-between items-start mb-6 gap-4">
        <div>
          <div class="flex items-center gap-3 mb-2">
            <div class="flex items-center justify-center size-10 rounded-full" style="background-color: #87d068">
              <UIcon name="i-lucide-radar" class="size-5 text-white" />
            </div>
            <div>
              <h1 class="text-2xl font-bold m-0">{{ resource.name }}</h1>
              <p class="m-0 text-xs text-muted">
                {{ resource.type.toUpperCase() }} monitor{{
                  isHeartbeat ? '' : ' for ' + resource.target
                }}
              </p>
              <div class="flex items-center gap-2 mt-1.5">
                <span
                  class="inline-block size-2 rounded-full"
                  :style="{
                    backgroundColor:
                      !isLiveLoading && liveData
                        ? 'var(--color-text-success, #52c41a)'
                        : 'var(--color-border-secondary, #d9d9d9)',
                  }"
                />
                <span v-if="lastUpdated" class="text-xs text-muted">
                  Updated {{ lastUpdatedRelative }}
                </span>
                <UButton size="xs" color="neutral" variant="soft" :disabled="isLiveLoading" @click="refresh">
                  ↻
                </UButton>
              </div>
            </div>
          </div>
        </div>
        <div class="flex gap-2">
          <UButton color="neutral" variant="soft" icon="i-lucide-pause" @click="handlePauseResource">
            Pause
          </UButton>
          <UButton color="neutral" variant="soft" icon="i-lucide-pencil" @click="openEditModal">
            Edit
          </UButton>
          <UDropdownMenu
            :items="[
              [
                { label: 'Edit', onSelect: openEditModal },
                { label: 'Duplicate', disabled: true },
              ],
              [{ label: 'Delete', color: 'error', disabled: true }],
            ]"
          >
            <UButton color="neutral" variant="soft" icon="i-lucide-ellipsis" />
          </UDropdownMenu>
        </div>
      </div>

      <!-- Alerts -->
      <UAlert
        v-if="liveError && !isTerminated"
        color="warning"
        variant="soft"
        icon="i-lucide-triangle-alert"
        class="mb-3"
        title="Could not refresh - showing last known data"
      />
      <UAlert
        v-if="isTerminated"
        color="warning"
        variant="soft"
        icon="i-lucide-triangle-alert"
        class="mb-3"
        title="This monitor no longer exists - showing last known data"
      />

      <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <!-- Left Column -->
        <div class="lg:col-span-2 space-y-4">
          <ResourceStatusCard :resource="resource" :now-ts="nowTs" />
          <ResourceHeartbeat v-if="isHeartbeat" :resource="resource" :now-ts="nowTs" />
          <ResourcePerformance :resource="resource" v-model:time-range="timeRange" />

          <!-- Protocol Info -->
          <UCard v-if="isProtocol" data-testid="protocol-info-card">
            <template #header>
              <div class="text-sm font-semibold">Protocol details</div>
            </template>
            <div class="grid grid-cols-2 gap-4 text-sm">
              <div>
                <div class="text-xs text-muted mb-1">Protocol</div>
                <UBadge color="info" variant="subtle">{{ resource.protocol_type?.toUpperCase() ?? '—' }}</UBadge>
              </div>
              <div>
                <div class="text-xs text-muted mb-1">Port</div>
                <div>{{ resource.protocol_port ?? 'Default' }}</div>
              </div>
            </div>
          </UCard>

          <ResourceIncidents :incidents="resource.incidents ?? []" />
        </div>

        <!-- Right Column -->
        <div>
          <ResourceDetails :resource="resource" />
        </div>
      </div>
    </template>

    <template v-else>
      <UEmptyState
        icon="i-lucide-search-x"
        title="Resource not found"
        description="The requested resource does not exist."
      >
        <template #actions>
          <UButton color="primary" @click="goBack">Go Back</UButton>
        </template>
      </UEmptyState>
    </template>

    <ResourceModal v-model:open="showEditModal" :resource="resource" @submit="handleEditSubmit" />
  </div>
</template>
