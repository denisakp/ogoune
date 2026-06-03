<script setup lang="ts">
import { onMounted, onUnmounted, ref, computed, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { message } from 'ant-design-vue'
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

const resource = ref<Resource | null>(null)
const timeRange = ref<'24h' | '7d' | '30d' | '365d'>('24h')
const showEditModal = ref(false)
const nowTs = ref(Date.now())
let timer: number | undefined

const resourceId = computed(() => route.params.id as string)

const {
  liveData, isLoading: isLiveLoading, lastUpdated, error: liveError,
  isTerminated, refresh, startPolling, stopPolling,
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
  timer = window.setInterval(() => { nowTs.value = Date.now() }, 1000)
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
    ...resource.value, ...incoming,
    tags: incoming.tags ? [...incoming.tags] : resource.value.tags,
    incidents: incoming.incidents ? [...incoming.incidents] : resource.value.incidents,
    response_times: resource.value.response_times,
  }
})

const loadResource = async () => {
  if (!resourceId.value) { message.error('Resource ID not found'); return }
  try {
    const data = await store.loadResourceWithResponseTimes(resourceId.value, 50)
    if (data) resource.value = data
    else message.error('Failed to load resource')
  } catch (err) {
    message.error(err instanceof Error ? err.message : 'Failed to load resource')
  }
}

const handlePauseResource = async () => {
  if (!resource.value) return
  try { await store.pauseMonitoring(resource.value.id); await loadResource() } catch { /* interceptor handles */ }
}

const openEditModal = () => { showEditModal.value = true }
const handleEditSubmit = async () => { showEditModal.value = false; await loadResource() }
const goBack = () => { router.back() }
</script>

<template>
  <div style="padding: 24px">
    <a-spin :spinning="store.fetchLoading">
      <template v-if="resource">
        <a-button type="text" style="margin-bottom: 16px" @click="goBack">
          <template #icon><UIcon name="i-lucide-arrow-left" /></template>
          Monitoring
        </a-button>

        <!-- Header -->
        <div style="display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 24px">
          <div>
            <div style="display: flex; align-items: center; gap: 12px; margin-bottom: 8px">
              <a-avatar :size="40" style="background-color: #87d068">
                <template #icon><a-icon-api /></template>
              </a-avatar>
              <div>
                <h1 style="font-size: 24px; font-weight: bold; margin: 0">{{ resource.name }}</h1>
                <p style="margin: 0; font-size: 12px; color: rgba(0,0,0,0.45)">
                  {{ resource.type.toUpperCase() }} monitor{{ isHeartbeat ? '' : ' for ' + resource.target }}
                </p>
                <div style="display: flex; align-items: center; gap: 8px; margin-top: 6px">
                  <span style="width: 8px; height: 8px; border-radius: 50%; display: inline-block"
                    :style="{ backgroundColor: !isLiveLoading && liveData ? 'var(--color-text-success, #52c41a)' : 'var(--color-border-secondary, #d9d9d9)' }" />
                  <span v-if="lastUpdated" style="font-size: 12px; color: rgba(0,0,0,0.55)">Updated {{ lastUpdatedRelative }}</span>
                  <a-button size="small" :disabled="isLiveLoading" @click="refresh"> ↻ </a-button>
                </div>
              </div>
            </div>
          </div>
          <div style="display: flex; gap: 8px">
            <a-button @click="handlePauseResource"><template #icon><UIcon name="i-lucide-pause" /></template>Pause</a-button>
            <a-button @click="openEditModal"><template #icon><UIcon name="i-lucide-pencil" /></template>Edit</a-button>
            <a-dropdown>
              <template #overlay>
                <a-menu>
                  <a-menu-item>Edit</a-menu-item>
                  <a-menu-item>Duplicate</a-menu-item>
                  <a-menu-divider />
                  <a-menu-item danger>Delete</a-menu-item>
                </a-menu>
              </template>
              <a-button><template #icon><UIcon name="i-lucide-ellipsis" /></template></a-button>
            </a-dropdown>
          </div>
        </div>

        <!-- Alerts -->
        <a-alert v-if="liveError && !isTerminated" style="margin-bottom: 12px" type="warning" show-icon
          message="Could not refresh - showing last known data" />
        <a-alert v-if="isTerminated" style="margin-bottom: 12px" type="warning" show-icon
          message="This monitor no longer exists - showing last known data" />

        <a-row :gutter="24">
          <!-- Left Column -->
          <a-col :xs="24" :lg="16">
            <ResourceStatusCard :resource="resource" :now-ts="nowTs" />
            <ResourceHeartbeat v-if="isHeartbeat" :resource="resource" :now-ts="nowTs" />
            <ResourcePerformance :resource="resource" v-model:time-range="timeRange" />

            <!-- Protocol Info -->
            <a-card v-if="isProtocol" style="margin-bottom: 16px" data-testid="protocol-info-card">
              <template #title><div style="font-size: 14px; font-weight: 600">Protocol details</div></template>
              <a-descriptions :column="2" size="small">
                <a-descriptions-item label="Protocol">
                  <a-tag color="blue">{{ resource.protocol_type?.toUpperCase() ?? '—' }}</a-tag>
                </a-descriptions-item>
                <a-descriptions-item label="Port">{{ resource.protocol_port ?? 'Default' }}</a-descriptions-item>
              </a-descriptions>
            </a-card>

            <ResourceIncidents :incidents="resource.incidents ?? []" />
          </a-col>

          <!-- Right Column -->
          <a-col :xs="24" :lg="8">
            <ResourceDetails :resource="resource" />
          </a-col>
        </a-row>
      </template>

      <template v-else>
        <a-result status="404" title="Resource not found" sub-title="The requested resource does not exist.">
          <template #extra><a-button type="primary" @click="goBack">Go Back</a-button></template>
        </a-result>
      </template>
    </a-spin>

    <ResourceModal v-model:open="showEditModal" :resource="resource" @submit="handleEditSubmit" />
  </div>
</template>

<style scoped>
:deep(.ant-card) { border-radius: 8px; }
:deep(.ant-card-head) { border-bottom: 1px solid rgba(0,0,0,0.06); }
:deep(.ant-card-body) { padding: 24px; }
</style>
