<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { message } from 'ant-design-vue'
import {
  PauseOutlined,
  PlayCircleOutlined,
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

import { useIncidents } from '@/composables/useIncidents'
import type { Incident, IncidentEventStep } from '@/types'

const router = useRouter()
const route = useRoute()

const { resolveIncident, getIncidentById } = useIncidents()

// State
const incident = ref<Incident | null>(null)
const loading = ref(false)
const isSubmitting = ref(false)

// Get incident ID from route
const incidentId = computed(() => route.params.id as string)

// Fetch incident on mount
onMounted(async () => {
  await loadIncident()
})

// Load incident
const loadIncident = async () => {
  if (!incidentId.value) {
    message.error('Incident ID not found')
    return
  }

  loading.value = true
  try {
    incident.value = await getIncidentById(incidentId.value)
  } catch (err) {
    const errorMsg = err instanceof Error ? err.message : 'Failed to load incident'
    message.error(errorMsg)
  } finally {
    loading.value = false
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
    timeZone: 'UTC',
  })
}

// Get status color
const getStatusColor = (incident: Incident): string => {
  return incident.resolved_at ? 'green' : 'red'
}

// Get status text
const getStatusText = (incident: Incident): string => {
  return incident.resolved_at ? 'Resolved' : 'Ongoing'
}

// Get cause badge color
const getCauseBadgeColor = (cause: string): string => {
  switch (cause) {
    case 'timeout':
    case 'Connection Timeout':
      return 'orange'
    case 'connection_error':
    case 'Connection Error':
      return 'red'
    case 'bad_status':
    case 'Bad Gateway':
      return 'volcano'
    default:
      return 'default'
  }
}

// Get event step icon and color
const getEventStepIcon = (step: IncidentEventStep) => {
  switch (step.step) {
    case 'detected':
      return { icon: 'exclamation-circle', color: 'red' }
    case 'alert_sent':
    case 'resource_down_alert':
      return { icon: 'bell', color: 'orange' }
    case 'resource_up_alert':
      return { icon: 'smile', color: 'blue' }
    case 'resolved':
      return { icon: 'check-circle', color: 'green' }
    default:
      return { icon: 'info-circle', color: 'blue' }
  }
}

// Format event step name
const formatEventStepName = (step: string): string => {
  return step
    .split('_')
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ')
}

// Calculate incident duration
const getIncidentDuration = (incident: Incident): string => {
  const startDate = new Date(incident.started_at)
  const endDate = incident.resolved_at ? new Date(incident.resolved_at) : new Date()
  const durationMs = endDate.getTime() - startDate.getTime()

  const days = Math.floor(durationMs / (1000 * 60 * 60 * 24))
  const hours = Math.floor((durationMs % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60))
  const minutes = Math.floor((durationMs % (1000 * 60 * 60)) / (1000 * 60))
  const seconds = Math.floor(durationMs % (1000 * 60)) / 1000

  if (days > 0) return `${days}d ${hours}h ${minutes}m ${seconds}s`
  if (hours > 0) return `${hours}h ${minutes}m ${seconds}s`
  if (minutes > 0) return `${minutes}m ${seconds}s`
  return `${seconds}s`
}

// Handle resolve incident
const handleResolveIncident = () => {
  if (!incident.value) return

  const { confirm } = window
  if (
    confirm(
      `Resolve incident for "${incident.value.resource?.name || incident.value.resource_id}"?`,
    )
  ) {
    isSubmitting.value = true
    resolveIncident(incident.value.id)
      .then((resolved) => {
        incident.value = resolved
        message.success('Incident resolved successfully')
      })
      .catch((err) => {
        const errorMsg = err instanceof Error ? err.message : 'Failed to resolve incident'
        message.error(errorMsg)
      })
      .finally(() => {
        isSubmitting.value = false
      })
  }
}

// Go back to incidents list
const goBack = () => {
  router.back()
}

// Handle download response
const handleDownloadResponse = () => {
  if (!incident.value || !incident.value.details) {
    message.info('No response data available to download')
    return
  }

  const element = document.createElement('a')
  const file = new Blob([incident.value.details], { type: 'application/json' })
  element.href = URL.createObjectURL(file)
  element.download = `incident-${incident.value.id}-response.json`
  document.body.appendChild(element)
  element.click()
  document.body.removeChild(element)
}
</script>

<template>
  <div style="padding: 24px">
    <a-spin :spinning="loading">
      <template v-if="incident">
        <!-- Header -->
        <div style="margin-bottom: 24px">
          <div style="display: flex; align-items: center; gap: 12px; margin-bottom: 16px">
            <a-button type="text" @click="goBack">
              <template #icon>
                <ArrowLeftOutlined />
              </template>
              Incidents
            </a-button>
          </div>

          <div style="display: flex; align-items: flex-start; gap: 16px">
            <!-- Status Badge -->
            <div
              :style="{
                width: '24px',
                height: '24px',
                borderRadius: '50%',
                backgroundColor: getStatusColor(incident) === 'green' ? '#52c41a' : '#ff4d4f',
              }"
            ></div>

            <!-- Title and Info -->
            <div style="flex: 1">
              <h1 style="font-size: 24px; font-weight: bold; margin: 0; margin-bottom: 8px">
                {{ getStatusText(incident) }} incident on
                {{ incident.resource?.name || incident.resource_id }}
              </h1>
              <p style="color: rgba(0, 0, 0, 0.65); margin: 0; margin-bottom: 4px">
                {{ incident.resource?.type?.toUpperCase() }} monitor for
                {{ incident.resource?.target }}
              </p>
              <a-tag color="blue">Included</a-tag>
            </div>

            <!-- Action Buttons -->
            <div style="display: flex; gap: 8px">
              <a-button @click="handleDownloadResponse">
                <template #icon>
                  <a-icon-download />
                </template>
                Download response
              </a-button>
              <a-button
                v-if="!incident.resolved_at"
                type="primary"
                :loading="isSubmitting"
                @click="handleResolveIncident"
              >
                Resolve
              </a-button>
            </div>
          </div>
        </div>

        <!-- Main Content -->
        <a-row :gutter="16" style="margin-bottom: 24px">
          <!-- Left Column -->
          <a-col :xs="24" :lg="12">
            <!-- Root Cause -->
            <a-card style="margin-bottom: 16px">
              <template #title>
                <div style="font-size: 14px; font-weight: 600">Root cause</div>
              </template>
              <div style="font-size: 18px; font-weight: bold">
                <a-tag :color="getCauseBadgeColor(incident.cause)">
                  {{ incident.cause }}
                </a-tag>
              </div>
            </a-card>

            <!-- Status and Duration -->
            <a-row :gutter="16" style="margin-bottom: 16px">
              <a-col :xs="24" :sm="12">
                <a-card>
                  <template #title>
                    <div style="font-size: 14px; font-weight: 600">Status</div>
                  </template>
                  <div style="font-size: 18px; font-weight: bold">
                    <a-tag :color="getStatusColor(incident)">
                      {{ getStatusText(incident) }}
                    </a-tag>
                  </div>
                  <p style="color: rgba(0, 0, 0, 0.65); margin-top: 8px; margin-bottom: 0">
                    Started at {{ formatDate(incident.started_at) }}
                  </p>
                </a-card>
              </a-col>

              <a-col :xs="24" :sm="12">
                <a-card>
                  <template #title>
                    <div style="font-size: 14px; font-weight: 600">Duration</div>
                  </template>
                  <div style="font-size: 18px; font-weight: bold">
                    {{ getIncidentDuration(incident) }}
                  </div>
                  <p
                    v-if="incident.resolved_at"
                    style="color: rgba(0, 0, 0, 0.65); margin-top: 8px; margin-bottom: 0"
                  >
                    Resolved at {{ formatDate(incident.resolved_at) }}
                  </p>
                </a-card>
              </a-col>
            </a-row>

            <!-- Activity Log (Event Steps) -->
            <a-card>
              <template #title>
                <div style="font-size: 14px; font-weight: 600">Activity log</div>
              </template>

              <a-empty
                v-if="!incident.event_steps || incident.event_steps.length === 0"
                description="No events recorded"
              />

              <div v-else style="display: flex; flex-direction: column; gap: 12px">
                <div
                  v-for="(step, index) in incident.event_steps"
                  :key="step.id"
                  style="display: flex; gap: 16px; padding-bottom: 12px"
                  :style="{
                    borderBottom:
                      index < incident.event_steps!.length - 1 ? '1px solid #f0f0f0' : 'none',
                  }"
                >
                  <!-- Icon -->
                  <div
                    style="
                      flex-shrink: 0;
                      display: flex;
                      flex-direction: column;
                      align-items: center;
                    "
                  >
                    <a-icon
                      :type="getEventStepIcon(step).icon"
                      :style="{ fontSize: '16px', color: getEventStepIcon(step).color }"
                    />
                  </div>

                  <!-- Content -->
                  <div style="flex: 1; min-width: 0">
                    <div
                      style="display: flex; justify-content: space-between; align-items: flex-start"
                    >
                      <div>
                        <div style="font-weight: 600; margin-bottom: 4px">
                          {{ formatEventStepName(step.step) }}
                        </div>
                        <p
                          v-if="step.message"
                          style="color: rgba(0, 0, 0, 0.65); margin: 0; white-space: pre-wrap"
                        >
                          {{ step.message }}
                        </p>
                      </div>
                      <a-tag style="white-space: nowrap; margin-left: 8px">
                        {{ step.step === 'resolved' ? 'SUCCESS' : 'PENDING' }}
                      </a-tag>
                    </div>
                    <p style="color: rgba(0, 0, 0, 0.45); font-size: 12px; margin: 8px 0 0 0">
                      {{ formatDate(step.created_at) }}
                    </p>
                  </div>
                </div>
              </div>
            </a-card>
          </a-col>

          <!-- Right Column -->
          <a-col :xs="24" :lg="12">
            <!-- Request Card -->
            <a-card style="margin-bottom: 16px">
              <template #title>
                <div style="font-size: 14px; font-weight: 600">Request</div>
              </template>

              <a-tabs>
                <a-tab-pane key="url" tab="URL">
                  <div style="margin: 16px 0">
                    <div style="font-size: 12px; color: rgba(0, 0, 0, 0.65); margin-bottom: 8px">
                      METHOD
                    </div>
                    <div
                      style="
                        background-color: #1f1f1f;
                        color: #d4d4d4;
                        padding: 12px;
                        border-radius: 4px;
                        font-family: monospace;
                        font-size: 12px;
                        word-break: break-all;
                        display: flex;
                        justify-content: space-between;
                        align-items: center;
                      "
                    >
                      <span
                        >{{ incident.resource?.type?.toUpperCase() }}
                        {{ incident.resource?.target }}</span
                      >
                      <a-button size="small" type="text">
                        <template #icon>
                          <a-icon-copy />
                        </template>
                      </a-button>
                    </div>
                  </div>
                </a-tab-pane>
                <a-tab-pane key="headers" tab="Headers">
                  <div style="margin: 16px 0; color: rgba(0, 0, 0, 0.45); font-size: 12px">
                    No headers data available
                  </div>
                </a-tab-pane>
              </a-tabs>
            </a-card>

            <!-- Response Card -->
            <a-card style="margin-bottom: 16px">
              <template #title>
                <div style="font-size: 14px; font-weight: 600">Response</div>
              </template>

              <a-tabs>
                <a-tab-pane key="body" tab="Body">
                  <div
                    v-if="incident.details"
                    style="
                      background-color: #1f1f1f;
                      color: #d4d4d4;
                      padding: 12px;
                      border-radius: 4px;
                      font-family: monospace;
                      font-size: 11px;
                      max-height: 300px;
                      overflow-y: auto;
                      white-space: pre-wrap;
                      word-break: break-all;
                      margin: 16px 0;
                    "
                  >
                    {{ incident.details }}
                  </div>
                  <div v-else style="margin: 16px 0; color: rgba(0, 0, 0, 0.45); font-size: 12px">
                    &lt;empty&gt;
                  </div>
                </a-tab-pane>
                <a-tab-pane key="headers" tab="Headers">
                  <div style="margin: 16px 0; color: rgba(0, 0, 0, 0.45); font-size: 12px">
                    No headers data available
                  </div>
                </a-tab-pane>
              </a-tabs>
            </a-card>

            <!-- Traceroute Card (Optional) -->
            <a-card>
              <template #title>
                <div style="font-size: 14px; font-weight: 600">Traceroute</div>
              </template>

              <div
                style="
                  display: flex;
                  justify-content: space-between;
                  align-items: center;
                  margin-bottom: 16px;
                "
              >
                <div></div>
                <a-button size="small">
                  <template #icon>
                    <a-icon-line-chart />
                  </template>
                  Trace analysis
                </a-button>
              </div>

              <div
                v-if="incident.details"
                style="
                  background-color: #1f1f1f;
                  color: #d4d4d4;
                  padding: 12px;
                  border-radius: 4px;
                  font-family: monospace;
                  font-size: 11px;
                  max-height: 200px;
                  overflow-y: auto;
                  white-space: pre-wrap;
                  word-break: break-all;
                "
              >
                Tracing route to {{ incident.resource?.target }}<br />
                hop no - node ip - ms<br />
                1 - 172.31.1.1(5 ms)<br />
                2 - 91.99.240.186(0 ms)<br />
                3 - 212.120.32.138(0 ms)
              </div>
              <div v-else style="color: rgba(0, 0, 0, 0.45); font-size: 12px">
                No traceroute data available
              </div>
            </a-card>
          </a-col>
        </a-row>
      </template>

      <template v-else>
        <a-empty description="Incident not found" />
      </template>
    </a-spin>
  </div>
</template>

<style scoped>
:deep(.ant-card-head) {
  border-bottom: 1px solid #f0f0f0;
}

:deep(.ant-card-body) {
  padding: 16px;
}

:deep(.ant-tabs-tab) {
  padding: 8px 0 !important;
}
</style>
