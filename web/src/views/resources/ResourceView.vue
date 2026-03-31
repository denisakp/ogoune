<script setup lang="ts">
import { onMounted, onUnmounted, ref, computed, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { message } from 'ant-design-vue'
import {
  PauseOutlined,
  ArrowLeftOutlined,
  DashboardOutlined,
  RiseOutlined,
  FallOutlined,
  SafetyOutlined,
  GlobalOutlined,
  CalendarOutlined,
  CheckCircleOutlined,
  WarningOutlined,
  ClockCircleOutlined,
  EditOutlined,
  EllipsisOutlined,
} from '@ant-design/icons-vue'

import { useResources } from '@/composables/useResources.ts'
import { useDateTime } from '@/composables/useDateTime.ts'
import ResourceModal from '@/components/resources/ResourceModal.vue'
import ResponseTimeChart from '@/components/ResponseTimeChart.vue'
import ExpiryBadge from '@/components/resources/ExpiryBadge.vue'
import type { Resource, Incident, ExpirationStatus } from '@/types'
import { useMonitorLive } from '@/composables/useMonitorLive'

const router = useRouter()
const route = useRoute()

// Use the composable following the project architecture
const { loading, pauseResource, loadResourceWithResponseTimes } = useResources()

// Local state
const resource = ref<Resource | null>(null)
const timeRange = ref<'24h' | '7d' | '30d' | '365d'>('24h')
const incidentsToShow = ref(3)
const showEditModal = ref(false)
const nowTs = ref(Date.now())
let timer: number | undefined

// Get resource ID from route
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
} = useMonitorLive(resourceId.value, () => resource.value?.interval)

const { getTimeRangeCutoff } = useDateTime()

// Get incidents filtered by time range
const filteredIncidents = computed(() => {
  if (!resource.value?.incidents) return []

  const cutoffDate = getTimeRangeCutoff(timeRange.value)

  return resource.value.incidents.filter((incident) => {
    const startDate = new Date(incident.started_at)
    return startDate >= cutoffDate
  })
})

// Calculate real uptime based on time range
const calculateUptime = computed((): number => {
  if (!resource.value) return 0

  // If resource is pending (no checks completed yet), return null-like value
  // We'll handle this in the display with a special message
  if (resource.value.status === 'pending' || !resource.value.last_checked) {
    return -1 // Special marker for pending state
  }

  // If the resource has an overall uptime property, use it as baseline
  if (resource.value.uptime !== undefined && timeRange.value === '24h') {
    return Number(resource.value.uptime.toFixed(1))
  }

  // Calculate based on incidents in the time range
  const cutoffDate = getTimeRangeCutoff(timeRange.value)
  const now = new Date()
  const totalDuration = now.getTime() - cutoffDate.getTime()

  if (totalDuration <= 0) return 100

  // Calculate total downtime from incidents in this period
  let totalDowntime = 0

  filteredIncidents.value.forEach((incident) => {
    const startDate = new Date(incident.started_at)
    const endDate = incident.resolved_at ? new Date(incident.resolved_at) : now

    // Only count time within our range
    const effectiveStart = startDate > cutoffDate ? startDate : cutoffDate
    const downtime = endDate.getTime() - effectiveStart.getTime()

    if (downtime > 0) {
      totalDowntime += downtime
    }
  })

  const uptime = ((totalDuration - totalDowntime) / totalDuration) * 100
  return Number(Math.max(0, Math.min(100, uptime)).toFixed(1))
})

// Get current stats with real data
const currentStats = computed(() => {
  const uptime = calculateUptime.value
  return {
    uptime: uptime >= 0 ? uptime : null, // null for pending
    incidents: filteredIncidents.value.length,
  }
})

// Get sorted incidents (latest first)
const sortedIncidents = computed(() => {
  if (!resource.value?.incidents) return []
  return [...resource.value.incidents].sort((a, b) => {
    return new Date(b.started_at).getTime() - new Date(a.started_at).getTime()
  })
})

// Get visible incidents based on pagination
const visibleIncidents = computed(() => {
  return sortedIncidents.value.slice(0, incidentsToShow.value)
})

// Check if there are more incidents to load
const hasMoreIncidents = computed(() => {
  return sortedIncidents.value.length > incidentsToShow.value
})

// Load more incidents
const loadMoreIncidents = () => {
  incidentsToShow.value += 3
}

// Decode base64 details
const decodeDetails = (details?: string): string => {
  if (!details) return 'No details available'
  try {
    return atob(details)
  } catch {
    return details
  }
}

// Get incident status
const getIncidentStatus = (incident: Incident): { text: string; color: string } => {
  if (incident.resolved_at) {
    return { text: 'Resolved', color: 'success' }
  }
  return { text: 'Active', color: 'error' }
}

// Format duration
const formatDuration = (startDate: string, endDate?: string | null): string => {
  const start = new Date(startDate).getTime()
  const end = endDate ? new Date(endDate).getTime() : Date.now()
  const diff = end - start

  const hours = Math.floor(diff / (1000 * 60 * 60))
  const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60))

  if (hours > 0) {
    return `${hours}h ${minutes}m`
  }
  return `${minutes}m`
}

// Format expiration date for display
const formatExpirationDate = (dateString: string): string => {
  const date = new Date(dateString)
  return date.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })
}

// Calculate days until expiration
const getDaysUntilExpiration = (dateString: string): number => {
  const expirationDate = new Date(dateString)
  const now = new Date()
  const diff = expirationDate.getTime() - now.getTime()
  return Math.ceil(diff / (1000 * 60 * 60 * 24))
}

// Get expiration status with color and icon
const getExpirationStatus = (dateString?: string): ExpirationStatus => {
  if (!dateString) return { text: 'Unknown', color: '#d9d9d9', type: 'success' }

  const days = getDaysUntilExpiration(dateString)

  if (days < 0) {
    return { text: 'Expired', color: '#ff4d4f', type: 'danger' }
  } else if (days <= 7) {
    return {
      text: `Expires in ${days} day${days !== 1 ? 's' : ''}`,
      color: '#ff4d4f',
      type: 'danger',
    }
  } else if (days <= 30) {
    return { text: `Expires in ${days} days`, color: '#faad14', type: 'warning' }
  } else {
    return { text: `Expires in ${days} days`, color: '#52c41a', type: 'success' }
  }
}

// Check if metadata exists
const hasMetadata = computed(() => {
  return (
    resource.value?.metadata &&
    (resource.value.metadata.ssl_expiration_date ||
      resource.value.metadata.ssl_issuer ||
      resource.value.metadata.domain_expiration_date ||
      resource.value.metadata.domain_registrar)
  )
})

// Fetch resource on mount
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
  if (timer) {
    window.clearInterval(timer)
  }
})

watch(liveData, (snapshot) => {
  if (!snapshot || !resource.value) {
    return
  }

  const incoming = snapshot.resource as Resource
  resource.value = {
    ...resource.value,
    ...incoming,
    tags: incoming.tags ? [...incoming.tags] : resource.value.tags,
    incidents: incoming.incidents ? [...incoming.incidents] : resource.value.incidents,
    response_times: resource.value.response_times,
  }
})

const lastUpdatedRelative = computed(() => {
  if (!lastUpdated.value) {
    return ''
  }
  const deltaSeconds = Math.max(0, Math.floor((nowTs.value - lastUpdated.value.getTime()) / 1000))
  if (deltaSeconds < 5) {
    return 'just now'
  }
  return `${deltaSeconds}s ago`
})

const isConfirming = computed(() => {
  if (!resource.value) return false
  return (
    resource.value.status === 'down' &&
    resource.value.failure_count > 0 &&
    resource.value.failure_count < resource.value.confirmation_checks
  )
})

const confirmationProgress = computed(() => {
  if (!resource.value) return ''
  return `${resource.value.failure_count}/${resource.value.confirmation_checks}`
})

const nextConfirmationCountdown = computed(() => {
  if (!resource.value || !resource.value.last_checked) return 'n/a'
  const nextTs =
    new Date(resource.value.last_checked).getTime() + resource.value.confirmation_interval * 1000
  const remainingSec = Math.max(0, Math.ceil((nextTs - nowTs.value) / 1000))
  return `${remainingSec}s`
})

const isFlapping = computed(() => resource.value?.status === 'flapping')

const flappingDuration = computed(() => {
  if (!resource.value?.flap_started_at) return ''
  const start = new Date(resource.value.flap_started_at).getTime()
  const diff = nowTs.value - start
  const hours = Math.floor(diff / (1000 * 60 * 60))
  const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60))
  if (hours > 0) return `${hours}h ${minutes}m`
  return `${minutes}m`
})

const flappingTransitionText = computed(() => {
  if (!resource.value) return ''
  const threshold = resource.value.flap_threshold
  if (!threshold || threshold < 1) return 'multiple status transitions'
  return `${threshold}+ status transitions`
})

const isHeartbeat = computed(() => resource.value?.type === 'heartbeat')

const pingUrl = computed(() => {
  if (!resource.value?.heartbeat_slug) return ''
  return `${window.location.origin}/ping/${resource.value.heartbeat_slug}`
})

const lastPingAtFormatted = computed(() => {
  if (!resource.value?.last_ping_at) return 'Never'
  return formatDate(resource.value.last_ping_at)
})

const nextExpectedPingCountdown = computed((): string | null => {
  if (!resource.value?.last_ping_at) return null
  const interval = resource.value.heartbeat_interval ?? 0
  const grace = resource.value.heartbeat_grace ?? 0
  const deadline =
    new Date(resource.value.last_ping_at).getTime() + (interval + grace) * 1000
  const remaining = Math.max(0, Math.ceil((deadline - nowTs.value) / 1000))
  if (remaining === 0) return 'Overdue'
  const minutes = Math.floor(remaining / 60)
  const seconds = remaining % 60
  if (minutes > 0) return `${minutes}m ${seconds}s`
  return `${seconds}s`
})

const copyPingUrl = async () => {
  if (!pingUrl.value) return
  try {
    await navigator.clipboard.writeText(pingUrl.value)
    message.success('Ping URL copied!')
  } catch {
    message.error('Failed to copy')
  }
}

const heartbeatSnippet = computed(() => {
  const url = pingUrl.value || 'https://your-ogoune-host/ping/<slug>'
  return `curl -fsS "${url}" >/dev/null`
})

// Load resource using the store with response times
const loadResource = async () => {
  if (!resourceId.value) {
    message.error('Resource ID not found')
    return
  }

  try {
    // Use the store method to fetch resource with response times (limit 50)
    const data = await loadResourceWithResponseTimes(resourceId.value, 50)
    if (data) {
      resource.value = data
    } else {
      message.error('Failed to load resource')
    }
  } catch (err) {
    const errorMsg = err instanceof Error ? err.message : 'Failed to load resource'
    message.error(errorMsg)
    console.error('Error loading resource:', err)
  }
}

// Format date
const formatDate = (dateString: string): string => {
  const date = new Date(dateString)
  return date.toLocaleString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  })
}

// Get status text
const getStatusText = (status: string): string => {
  const texts: Record<string, string> = {
    up: 'Up',
    down: 'Down',
    paused: 'Paused',
    pending: 'Pending',
    error: 'Error',
    waiting: 'Waiting',
  }
  return texts[status] || status
}

// Pause resource
const handlePauseResource = async () => {
  if (!resource.value) return
  try {
    await pauseResource(resource.value.id)
    await loadResource()
  } catch {
    // axios interceptor will handle the error toast here
  }
}

// Handle edit modal
const openEditModal = () => {
  showEditModal.value = true
}

const handleEditSubmit = async () => {
  showEditModal.value = false
  await loadResource()
}

// Go back
const goBack = () => {
  router.back()
}
</script>

<template>
  <div style="padding: 24px">
    <a-spin :spinning="loading">
      <template v-if="resource">
        <!-- Back Button -->
        <a-button type="text" style="margin-bottom: 16px" @click="goBack">
          <template #icon>
            <ArrowLeftOutlined />
          </template>
          Monitoring
        </a-button>

        <!-- Header -->
        <div
          style="
            display: flex;
            justify-content: space-between;
            align-items: flex-start;
            margin-bottom: 24px;
          "
        >
          <div>
            <div style="display: flex; align-items: center; gap: 12px; margin-bottom: 8px">
              <a-avatar :size="40" style="background-color: #87d068">
                <template #icon>
                  <a-icon-api />
                </template>
              </a-avatar>
              <div>
                <h1 style="font-size: 24px; font-weight: bold; margin: 0">{{ resource.name }}</h1>
                <p style="margin: 0; font-size: 12px; color: rgba(0, 0, 0, 0.45)">
                  {{ resource.type.toUpperCase() }} monitor{{ isHeartbeat ? '' : ' for ' + resource.target }}
                </p>
                <div style="display: flex; align-items: center; gap: 8px; margin-top: 6px">
                  <span
                    style="width: 8px; height: 8px; border-radius: 50%; display: inline-block"
                    :style="{
                      backgroundColor:
                        !isLiveLoading && liveData
                          ? 'var(--color-text-success, #52c41a)'
                          : 'var(--color-border-secondary, #d9d9d9)',
                    }"
                  />
                  <span v-if="lastUpdated" style="font-size: 12px; color: rgba(0, 0, 0, 0.55)">
                    Updated {{ lastUpdatedRelative }}
                  </span>
                  <a-button size="small" :disabled="isLiveLoading" @click="refresh"> ↻ </a-button>
                </div>
              </div>
            </div>
          </div>
          <div style="display: flex; gap: 8px">
            <a-button @click="handlePauseResource">
              <template #icon>
                <PauseOutlined />
              </template>
              Pause
            </a-button>
            <a-button @click="openEditModal">
              <template #icon>
                <EditOutlined />
              </template>
              Edit
            </a-button>
            <a-dropdown>
              <template #overlay>
                <a-menu>
                  <a-menu-item>Edit</a-menu-item>
                  <a-menu-item>Duplicate</a-menu-item>
                  <a-menu-divider />
                  <a-menu-item danger>Delete</a-menu-item>
                </a-menu>
              </template>
              <a-button>
                <template #icon>
                  <EllipsisOutlined />
                </template>
              </a-button>
            </a-dropdown>
          </div>
        </div>

        <!-- Main Content -->
        <a-alert
          v-if="liveError && !isTerminated"
          style="margin-bottom: 12px"
          type="warning"
          show-icon
          message="Could not refresh - showing last known data"
        />
        <a-alert
          v-if="isTerminated"
          style="margin-bottom: 12px"
          type="warning"
          show-icon
          message="This monitor no longer exists - showing last known data"
        />
        <a-row :gutter="24">
          <!-- Left Column -->
          <a-col :xs="24" :lg="16">
            <!-- Current Status -->
            <a-card style="margin-bottom: 16px">
              <template #title>
                <div style="font-size: 14px; font-weight: 600">Current status</div>
              </template>
              <a-alert
                v-if="isConfirming"
                style="margin-bottom: 16px"
                type="warning"
                show-icon
                :message="`Confirming outage: ${confirmationProgress}`"
                :description="`Next confirmation check in ${nextConfirmationCountdown}`"
              />
              <a-alert
                v-if="isFlapping"
                style="margin-bottom: 16px"
                type="warning"
                show-icon
                message="Service is flapping"
                :description="`${flappingTransitionText}${flappingDuration ? ` over ${flappingDuration}` : ''}. Alerts suppressed until service stabilizes.`"
              />
              <a-alert
                v-if="isHeartbeat && resource.waiting"
                style="margin-bottom: 16px"
                type="info"
                show-icon
                data-testid="heartbeat-waiting-alert"
                message="Waiting for first ping"
                description="This monitor will transition to UP as soon as it receives its first ping."
              />
              <a-row :gutter="16">
                <a-col :xs="12" :sm="8">
                  <div style="text-align: center">
                    <div
                      style="font-size: 28px; font-weight: bold"
                      :style="{
                        color:
                          resource.status === 'down'
                            ? '#ff4d4f'
                            : resource.status === 'paused'
                              ? '#d9d9d9'
                              : '#52c41a',
                      }"
                    >
                      {{ getStatusText(resource.status) }}
                    </div>
                    <div style="font-size: 12px; color: rgba(0, 0, 0, 0.65); margin-top: 8px">
                      Currently {{ resource.is_active ? 'active' : 'inactive' }}
                    </div>
                  </div>
                </a-col>
                <a-col :xs="12" :sm="8">
                  <div style="text-align: center">
                    <div style="font-size: 24px; font-weight: bold; color: #faad14">
                      {{ resource.failure_count }}
                    </div>
                    <div style="font-size: 12px; color: rgba(0, 0, 0, 0.65); margin-top: 8px">
                      Failures
                    </div>
                  </div>
                </a-col>
                <a-col :xs="24" :sm="8">
                  <div style="text-align: center">
                    <div style="font-size: 12px; color: rgba(0, 0, 0, 0.65); margin-bottom: 8px">
                      Last checked
                    </div>
                    <div style="font-size: 12px; font-weight: 600">
                      {{ resource.last_checked ? formatDate(resource.last_checked) : 'Never' }}
                    </div>
                  </div>
                </a-col>
              </a-row>
            </a-card>

            <!-- Heartbeat Integration -->
            <a-card v-if="isHeartbeat" style="margin-bottom: 16px" data-testid="heartbeat-integration-card">
              <template #title>
                <div style="font-size: 14px; font-weight: 600">Heartbeat integration</div>
              </template>
              <div style="display: flex; flex-direction: column; gap: 16px">
                <!-- Ping URL -->
                <div>
                  <div style="font-size: 12px; color: rgba(0, 0, 0, 0.45); margin-bottom: 4px">
                    Ping URL
                  </div>
                  <div style="display: flex; align-items: center; gap: 8px">
                    <code
                      data-testid="ping-url"
                      style="
                        flex: 1;
                        font-size: 12px;
                        background: rgba(0, 0, 0, 0.04);
                        padding: 6px 10px;
                        border-radius: 4px;
                        word-break: break-all;
                      "
                    >{{ pingUrl }}</code>
                    <a-button size="small" @click="copyPingUrl">Copy</a-button>
                  </div>
                </div>

                <!-- Last ping -->
                <div>
                  <div style="font-size: 12px; color: rgba(0, 0, 0, 0.45); margin-bottom: 4px">
                    Last ping received
                  </div>
                  <div style="font-size: 14px; font-weight: 500" data-testid="last-ping-at">
                    {{ lastPingAtFormatted }}
                  </div>
                </div>

                <!-- Next expected ping countdown -->
                <div v-if="resource.last_ping_at">
                  <div style="font-size: 12px; color: rgba(0, 0, 0, 0.45); margin-bottom: 4px">
                    Next deadline
                  </div>
                  <div
                    style="font-size: 14px; font-weight: 500"
                    data-testid="next-ping-countdown"
                    :style="{ color: nextExpectedPingCountdown === 'Overdue' ? '#ff4d4f' : 'inherit' }"
                  >
                    {{ nextExpectedPingCountdown }}
                  </div>
                </div>

                <!-- Integration snippet -->
                <div>
                  <div style="font-size: 12px; color: rgba(0, 0, 0, 0.45); margin-bottom: 8px">
                    Add to your script
                  </div>
                  <div
                    data-testid="heartbeat-snippet"
                    style="
                      font-family: monospace;
                      font-size: 12px;
                      background: rgba(0, 0, 0, 0.04);
                      padding: 12px;
                      border-radius: 4px;
                      word-break: break-all;
                    "
                  >{{ heartbeatSnippet }}</div>
                </div>
              </div>
            </a-card>

            <!-- Performance Stats -->
            <a-card style="margin-bottom: 16px">
              <template #title>
                <div style="display: flex; justify-content: space-between; align-items: center">
                  <span style="font-size: 14px; font-weight: 600">Performance</span>
                  <a-radio-group v-model:value="timeRange" button-style="solid" size="small">
                    <a-radio-button value="24h">24h</a-radio-button>
                    <a-radio-button value="7d">7d</a-radio-button>
                    <a-radio-button value="30d">30d</a-radio-button>
                    <a-radio-button value="365d">1y</a-radio-button>
                  </a-radio-group>
                </div>
              </template>

              <!-- Uptime and Incidents Row -->
              <a-row :gutter="24" style="margin-bottom: 24px">
                <a-col :xs="24" :sm="12">
                  <div style="text-align: center; padding: 24px">
                    <div
                      style="font-size: 48px; font-weight: bold"
                      :style="{ color: currentStats.uptime === null ? '#d9d9d9' : '#52c41a' }"
                    >
                      {{ currentStats.uptime !== null ? currentStats.uptime + '%' : 'Pending' }}
                    </div>
                    <div style="font-size: 14px; color: rgba(0, 0, 0, 0.65); margin-top: 8px">
                      {{ currentStats.uptime === null ? 'Waiting for first check' : 'Uptime' }}
                    </div>
                  </div>
                </a-col>
                <a-col :xs="24" :sm="12">
                  <div style="text-align: center; padding: 24px">
                    <div style="font-size: 48px; font-weight: bold; color: #f5222d">
                      {{ currentStats.incidents }}
                    </div>
                    <div style="font-size: 14px; color: rgba(0, 0, 0, 0.65); margin-top: 8px">
                      Incidents
                    </div>
                  </div>
                </a-col>
              </a-row>

              <!-- Response Time Statistics -->
              <div
                v-if="resource.response_times && resource.response_times.length > 0"
                style="
                  display: grid;
                  grid-template-columns: repeat(3, 1fr);
                  gap: 16px;
                  padding: 16px;
                  background: rgba(0, 0, 0, 0.02);
                  border-radius: 8px;
                  margin-bottom: 16px;
                "
              >
                <!-- Average -->
                <div style="text-align: center">
                  <DashboardOutlined style="font-size: 24px; color: #1890ff; margin-bottom: 8px" />
                  <div
                    style="font-size: 20px; font-weight: 600; color: #1890ff; margin-bottom: 4px"
                  >
                    {{
                      (
                        resource.response_times.reduce((sum, r) => sum + r.response_time, 0) /
                        resource.response_times.length
                      ).toFixed(0)
                    }}ms
                  </div>
                  <div style="font-size: 12px; color: rgba(0, 0, 0, 0.45)">Average</div>
                </div>

                <!-- Min -->
                <div style="text-align: center">
                  <RiseOutlined style="font-size: 24px; color: #52c41a; margin-bottom: 8px" />
                  <div
                    style="font-size: 20px; font-weight: 600; color: #52c41a; margin-bottom: 4px"
                  >
                    {{ Math.min(...resource.response_times.map((r) => r.response_time)) }}ms
                  </div>
                  <div style="font-size: 12px; color: rgba(0, 0, 0, 0.45)">Min</div>
                </div>

                <!-- Max -->
                <div style="text-align: center">
                  <FallOutlined style="font-size: 24px; color: #ff4d4f; margin-bottom: 8px" />
                  <div
                    style="font-size: 20px; font-weight: 600; color: #ff4d4f; margin-bottom: 4px"
                  >
                    {{ Math.max(...resource.response_times.map((r) => r.response_time)) }}ms
                  </div>
                  <div style="font-size: 12px; color: rgba(0, 0, 0, 0.45)">Max</div>
                </div>
              </div>

              <!-- Response Time Chart -->
              <div>
                <ResponseTimeChart :data="resource.response_times" :height="300" />
              </div>
            </a-card>

            <!-- Recent Incidents -->
            <a-card>
              <template #title>
                <div style="display: flex; justify-content: space-between; align-items: center">
                  <span style="font-size: 14px; font-weight: 600">Recent incidents</span>
                  <a-badge
                    :count="sortedIncidents.length"
                    :number-style="{ backgroundColor: '#52c41a' }"
                  />
                </div>
              </template>

              <template v-if="sortedIncidents.length > 0">
                <a-timeline>
                  <a-timeline-item
                    v-for="incident in visibleIncidents"
                    :key="incident.id"
                    :color="incident.resolved_at ? 'green' : 'red'"
                  >
                    <template #dot>
                      <a-icon-clock-circle
                        v-if="!incident.resolved_at"
                        style="font-size: 16px; color: #f5222d"
                      />
                      <a-icon-check-circle v-else style="font-size: 16px; color: #52c41a" />
                    </template>

                    <div style="padding-bottom: 16px">
                      <!-- Incident Header -->
                      <div
                        style="
                          display: flex;
                          justify-content: space-between;
                          align-items: start;
                          margin-bottom: 8px;
                        "
                      >
                        <div style="flex: 1">
                          <div
                            style="display: flex; align-items: center; gap: 8px; margin-bottom: 4px"
                          >
                            <a-tag :color="getIncidentStatus(incident).color">
                              {{ getIncidentStatus(incident).text }}
                            </a-tag>
                            <span style="font-size: 12px; color: rgba(0, 0, 0, 0.45)">
                              {{ formatDuration(incident.started_at, incident.resolved_at) }}
                            </span>
                          </div>
                          <div style="font-weight: 500; margin-bottom: 4px">
                            {{ incident.reason }}
                          </div>
                          <div
                            style="font-size: 12px; color: rgba(0, 0, 0, 0.65); margin-bottom: 4px"
                          >
                            <strong>Cause:</strong> {{ incident.cause }}
                          </div>
                        </div>
                      </div>

                      <!-- Incident Times -->
                      <div style="font-size: 12px; color: rgba(0, 0, 0, 0.45); margin-bottom: 8px">
                        <div>
                          <a-icon-calendar style="margin-right: 4px" />
                          Started: {{ formatDate(incident.started_at) }}
                        </div>
                        <div v-if="incident.resolved_at" style="margin-top: 4px">
                          <a-icon-check style="margin-right: 4px" />
                          Resolved: {{ formatDate(incident.resolved_at) }}
                        </div>
                        <div v-else style="margin-top: 4px; color: #f5222d">
                          <a-icon-exclamation-circle style="margin-right: 4px" />
                          Still ongoing
                        </div>
                      </div>

                      <!-- Incident Details -->
                      <a-collapse v-if="incident.details" ghost size="small">
                        <a-collapse-panel key="1" header="Technical details">
                          <div
                            style="
                              font-size: 12px;
                              font-family: monospace;
                              background: rgba(0, 0, 0, 0.02);
                              padding: 12px;
                              border-radius: 4px;
                              word-break: break-word;
                            "
                          >
                            {{ decodeDetails(incident.details) }}
                          </div>
                        </a-collapse-panel>
                      </a-collapse>
                    </div>
                  </a-timeline-item>
                </a-timeline>

                <!-- Load More Button -->
                <div
                  v-if="hasMoreIncidents"
                  style="
                    text-align: center;
                    margin-top: 16px;
                    padding-top: 16px;
                    border-top: 1px solid rgba(0, 0, 0, 0.06);
                  "
                >
                  <a-button @click="loadMoreIncidents">
                    <template #icon>
                      <a-icon-down />
                    </template>
                    Load more incidents ({{ sortedIncidents.length - incidentsToShow }} remaining)
                  </a-button>
                </div>
              </template>

              <template v-else>
                <a-empty description="No incidents recorded">
                  <template #image>
                    <a-icon-smile style="font-size: 48px; color: #52c41a" />
                  </template>
                </a-empty>
              </template>
            </a-card>
          </a-col>

          <!-- Right Column -->
          <a-col :xs="24" :lg="8">
            <!-- Monitor Details -->
            <a-card style="margin-bottom: 16px">
              <template #title>
                <div style="font-size: 14px; font-weight: 600">Monitor details</div>
              </template>
              <div style="display: flex; flex-direction: column; gap: 16px">
                <!-- Type -->
                <div>
                  <div style="font-size: 12px; color: rgba(0, 0, 0, 0.45); margin-bottom: 4px">
                    Type
                  </div>
                  <a-tag color="blue">{{ resource.type.toUpperCase() }}</a-tag>
                </div>

                <!-- Target -->
                <div>
                  <div style="font-size: 12px; color: rgba(0, 0, 0, 0.45); margin-bottom: 4px">
                    Target
                  </div>
                  <div style="font-size: 14px; word-break: break-all">{{ resource.target }}</div>
                </div>

                <!-- Check Interval -->
                <div>
                  <div style="font-size: 12px; color: rgba(0, 0, 0, 0.45); margin-bottom: 4px">
                    Check interval
                  </div>
                  <div style="font-size: 14px">Every {{ resource.interval }} seconds</div>
                </div>

                <!-- Timeout -->
                <div>
                  <div style="font-size: 12px; color: rgba(0, 0, 0, 0.45); margin-bottom: 4px">
                    Timeout
                  </div>
                  <div style="font-size: 14px">{{ resource.timeout }} seconds</div>
                </div>

                <!-- Created -->
                <div>
                  <div style="font-size: 12px; color: rgba(0, 0, 0, 0.45); margin-bottom: 4px">
                    Created
                  </div>
                  <div style="font-size: 14px">{{ formatDate(resource.created_at) }}</div>
                </div>

                <!-- Updated -->
                <div>
                  <div style="font-size: 12px; color: rgba(0, 0, 0, 0.45); margin-bottom: 4px">
                    Last updated
                  </div>
                  <div style="font-size: 14px">{{ formatDate(resource.updated_at) }}</div>
                </div>
              </div>
            </a-card>

            <!-- Tags -->
            <a-card style="margin-bottom: 16px">
              <template #title>
                <div style="font-size: 14px; font-weight: 600">Tags</div>
              </template>
              <div style="display: flex; flex-wrap: gap; gap: 8px">
                <a-tag
                  v-for="tag in resource.tags"
                  :key="tag.id"
                  :style="{
                    margin: '0',
                    backgroundColor: tag.color || '#f0f0f0',
                    color: '#000',
                    borderColor: 'transparent',
                  }"
                >
                  {{ tag.name }}
                </a-tag>
                <a-tag v-if="!resource.tags || resource.tags.length === 0" style="margin: 0">
                  No tags
                </a-tag>
              </div>
            </a-card>

            <!-- Additional Info -->
            <a-card>
              <template #title>
                <div style="font-size: 14px; font-weight: 600">Additional info</div>
              </template>

              <template v-if="hasMetadata">
                <div style="display: flex; flex-direction: column; gap: 20px">
                  <!-- SSL Certificate Information -->
                  <div
                    v-if="resource.metadata?.ssl_expiration_date || resource.metadata?.ssl_issuer"
                    style="
                      padding: 16px;
                      background: rgba(24, 144, 255, 0.05);
                      border-radius: 8px;
                      border-left: 3px solid #1890ff;
                    "
                  >
                    <div
                      style="
                        display: flex;
                        align-items: center;
                        gap: 8px;
                        margin-bottom: 12px;
                        font-weight: 600;
                        color: #1890ff;
                      "
                    >
                      <SafetyOutlined style="font-size: 18px" />
                      <span>SSL Certificate</span>
                    </div>

                    <!-- SSL Issuer -->
                    <div v-if="resource.metadata?.ssl_issuer" style="margin-bottom: 12px">
                      <div style="font-size: 12px; color: rgba(0, 0, 0, 0.45); margin-bottom: 4px">
                        Issuer
                      </div>
                      <div style="font-size: 14px; color: rgba(0, 0, 0, 0.85)">
                        {{ resource.metadata.ssl_issuer }}
                      </div>
                    </div>

                    <!-- SSL Expiration -->
                    <div v-if="resource.metadata?.ssl_expiration_date">
                      <div style="font-size: 12px; color: rgba(0, 0, 0, 0.45); margin-bottom: 4px">
                        Expiration Date
                      </div>
                      <div style="display: flex; align-items: center; gap: 8px">
                        <CalendarOutlined style="font-size: 14px; color: rgba(0, 0, 0, 0.45)" />
                        <span style="font-size: 14px; color: rgba(0, 0, 0, 0.85)">
                          {{ formatExpirationDate(resource.metadata.ssl_expiration_date) }}
                        </span>
                      </div>
                      <div
                        style="
                          margin-top: 8px;
                          display: flex;
                          align-items: center;
                          gap: 8px;
                          flex-wrap: wrap;
                        "
                      >
                        <a-tag
                          :color="getExpirationStatus(resource.metadata.ssl_expiration_date).color"
                        >
                          <template #icon>
                            <CheckCircleOutlined
                              v-if="
                                getExpirationStatus(resource.metadata.ssl_expiration_date).type ===
                                'success'
                              "
                            />
                            <WarningOutlined
                              v-else-if="
                                getExpirationStatus(resource.metadata.ssl_expiration_date).type ===
                                'warning'
                              "
                            />
                            <ClockCircleOutlined v-else />
                          </template>
                          {{ getExpirationStatus(resource.metadata.ssl_expiration_date).text }}
                        </a-tag>
                        <ExpiryBadge
                          v-if="
                            resource.expiry_status &&
                            resource.expiry_status !== 'ok' &&
                            resource.metadata?.ssl_days_remaining != null
                          "
                          type="ssl"
                          :days-remaining="resource.metadata.ssl_days_remaining"
                          :status="resource.expiry_status"
                        />
                      </div>
                    </div>
                  </div>

                  <!-- Domain Information -->
                  <div
                    v-if="
                      resource.metadata?.domain_expiration_date ||
                      resource.metadata?.domain_registrar
                    "
                    style="
                      padding: 16px;
                      background: rgba(82, 196, 26, 0.05);
                      border-radius: 8px;
                      border-left: 3px solid #52c41a;
                    "
                  >
                    <div
                      style="
                        display: flex;
                        align-items: center;
                        gap: 8px;
                        margin-bottom: 12px;
                        font-weight: 600;
                        color: #52c41a;
                      "
                    >
                      <GlobalOutlined style="font-size: 18px" />
                      <span>Domain</span>
                    </div>

                    <!-- Domain Registrar -->
                    <div v-if="resource.metadata?.domain_registrar" style="margin-bottom: 12px">
                      <div style="font-size: 12px; color: rgba(0, 0, 0, 0.45); margin-bottom: 4px">
                        Registrar
                      </div>
                      <div style="font-size: 14px; color: rgba(0, 0, 0, 0.85)">
                        {{ resource.metadata.domain_registrar }}
                      </div>
                    </div>

                    <!-- Domain Expiration -->
                    <div v-if="resource.metadata?.domain_expiration_date">
                      <div style="font-size: 12px; color: rgba(0, 0, 0, 0.45); margin-bottom: 4px">
                        Expiration Date
                      </div>
                      <div style="display: flex; align-items: center; gap: 8px">
                        <CalendarOutlined style="font-size: 14px; color: rgba(0, 0, 0, 0.45)" />
                        <span style="font-size: 14px; color: rgba(0, 0, 0, 0.85)">
                          {{ formatExpirationDate(resource.metadata.domain_expiration_date) }}
                        </span>
                      </div>
                      <div
                        style="
                          margin-top: 8px;
                          display: flex;
                          align-items: center;
                          gap: 8px;
                          flex-wrap: wrap;
                        "
                      >
                        <a-tag
                          :color="
                            getExpirationStatus(resource.metadata.domain_expiration_date).color
                          "
                        >
                          <template #icon>
                            <CheckCircleOutlined
                              v-if="
                                getExpirationStatus(resource.metadata.domain_expiration_date)
                                  .type === 'success'
                              "
                            />
                            <WarningOutlined
                              v-else-if="
                                getExpirationStatus(resource.metadata.domain_expiration_date)
                                  .type === 'warning'
                              "
                            />
                            <ClockCircleOutlined v-else />
                          </template>
                          {{ getExpirationStatus(resource.metadata.domain_expiration_date).text }}
                        </a-tag>
                        <ExpiryBadge
                          v-if="
                            resource.expiry_status &&
                            resource.expiry_status !== 'ok' &&
                            resource.metadata?.domain_days_remaining != null
                          "
                          type="domain"
                          :days-remaining="resource.metadata.domain_days_remaining"
                          :status="resource.expiry_status"
                        />
                      </div>
                    </div>
                  </div>
                </div>
              </template>

              <!-- Empty State -->
              <template v-else>
                <div style="text-align: center; padding: 32px 24px; color: rgba(0, 0, 0, 0.45)">
                  <a-icon-info-circle style="font-size: 40px; margin-bottom: 12px; opacity: 0.5" />
                  <div style="font-size: 14px; margin-bottom: 4px">No metadata available</div>
                  <div style="font-size: 12px">
                    SSL and domain information will appear here when available
                  </div>
                </div>
              </template>
            </a-card>
          </a-col>
        </a-row>
      </template>

      <template v-else>
        <a-result
          status="404"
          title="Resource not found"
          sub-title="The requested resource does not exist."
        >
          <template #extra>
            <a-button type="primary" @click="goBack">Go Back</a-button>
          </template>
        </a-result>
      </template>
    </a-spin>

    <!-- Edit Resource Modal -->
    <ResourceModal v-model:open="showEditModal" :resource="resource" @submit="handleEditSubmit" />
  </div>
</template>

<style scoped>
:deep(.ant-card) {
  border-radius: 8px;
}

:deep(.ant-card-head) {
  border-bottom: 1px solid rgba(0, 0, 0, 0.06);
}

:deep(.ant-card-body) {
  padding: 24px;
}
</style>
