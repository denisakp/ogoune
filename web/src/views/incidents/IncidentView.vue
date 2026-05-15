<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { message } from 'ant-design-vue'
import { ArrowLeftOutlined } from '@ant-design/icons-vue'

import { useIncidents } from '@/composables/useIncidents.ts'
import type { Incident, IncidentEventStep } from '@/types'
import DiagnosticsErrorSummary from '@/components/DiagnosticsErrorSummary.vue'
import DiagnosticsHeadersDisplay from '@/components/DiagnosticsHeadersDisplay.vue'
import DiagnosticsResponseBody from '@/components/DiagnosticsResponseBody.vue'
import DiagnosticsTimingBreakdown from '@/components/DiagnosticsTimingBreakdown.vue'

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

// Check if diagnostics are available
const hasDiagnostics = computed(() => {
  return incident.value?.diagnostics !== null && incident.value?.diagnostics !== undefined
})

// Check if ICMP network diagnostics are present (T041: hide when fields absent)
const hasICMPDiagnostics = computed(() => {
  const diag = incident.value?.diagnostics
  return diag != null && diag.icmp_available != null
})

// Check if keyword diagnostics are present (only for keyword monitor incidents)
const hasKeywordDiagnostics = computed(() => {
  const diag = incident.value?.diagnostics
  return diag != null && diag.keyword != null && diag.keyword !== ''
})

// Root cause hint label and badge color for ICMP network diagnostics
const rootCauseHintLabel = computed((): { label: string; color: string } => {
  const hint = incident.value?.diagnostics?.root_cause_hint
  switch (hint) {
    case 'host_unreachable':
      return { label: 'Host Unreachable', color: 'red' }
    case 'service_down':
      return { label: 'Service Down', color: 'orange' }
    case 'icmp_unavailable':
      return { label: 'ICMP Unavailable', color: 'default' }
    default:
      return { label: hint ?? 'Unknown', color: 'default' }
  }
})

// Copy URL to clipboard
const copyUrlToClipboard = async (url: string) => {
  try {
    await navigator.clipboard.writeText(url)
    message.success('URL copied to clipboard')
  } catch {
    message.error('Failed to copy URL')
  }
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

              <div v-if="hasDiagnostics && incident.diagnostics">
                <a-tabs>
                  <!-- URL Tab -->
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
                          margin-bottom: 16px;
                        "
                      >
                        {{ incident.diagnostics.request_method || 'HEAD' }}
                      </div>

                      <div style="font-size: 12px; color: rgba(0, 0, 0, 0.65); margin-bottom: 8px">
                        URL
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
                        <span>{{ incident.diagnostics.request_url }}</span>
                        <a-button
                          size="small"
                          type="text"
                          @click="copyUrlToClipboard(incident.diagnostics?.request_url || '')"
                        >
                          <template #icon>
                            <a-icon-copy />
                          </template>
                        </a-button>
                      </div>

                      <div v-if="incident.diagnostics.request_timeout" style="margin-top: 16px">
                        <div
                          style="font-size: 12px; color: rgba(0, 0, 0, 0.65); margin-bottom: 8px"
                        >
                          TIMEOUT
                        </div>
                        <div style="font-size: 14px; font-weight: 500">
                          {{ incident.diagnostics.request_timeout }}ms
                        </div>
                      </div>
                    </div>
                  </a-tab-pane>

                  <!-- Headers Tab -->
                  <a-tab-pane key="headers" tab="Headers">
                    <div style="margin: 16px 0">
                      <DiagnosticsHeadersDisplay
                        :headers="incident.diagnostics.request_headers"
                        title="Request Headers"
                        empty-message="No request headers available"
                      />
                    </div>
                  </a-tab-pane>
                </a-tabs>
              </div>

              <div v-else>
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
              </div>
            </a-card>

            <!-- Response Card -->
            <a-card style="margin-bottom: 16px">
              <template #title>
                <div style="font-size: 14px; font-weight: 600">Response</div>
              </template>

              <div v-if="hasDiagnostics && incident.diagnostics">
                <!-- Status Code -->
                <div v-if="incident.diagnostics.http_status_code" style="margin-bottom: 20px">
                  <div style="font-size: 12px; color: rgba(0, 0, 0, 0.65); margin-bottom: 8px">
                    HTTP STATUS CODE
                  </div>
                  <div
                    style="
                      font-size: 24px;
                      font-weight: bold;
                      padding: 12px;
                      border-radius: 4px;
                      display: inline-block;
                    "
                    :style="{
                      color:
                        incident.diagnostics.http_status_code >= 500
                          ? '#ff4d4f'
                          : incident.diagnostics.http_status_code >= 400
                            ? '#faad14'
                            : incident.diagnostics.http_status_code >= 300
                              ? '#1890ff'
                              : '#52c41a',
                      backgroundColor:
                        incident.diagnostics.http_status_code >= 500
                          ? '#fff1f0'
                          : incident.diagnostics.http_status_code >= 400
                            ? '#fffbe6'
                            : incident.diagnostics.http_status_code >= 300
                              ? '#e6f7ff'
                              : '#f6ffed',
                    }"
                  >
                    {{ incident.diagnostics.http_status_code }}
                  </div>
                </div>

                <a-tabs>
                  <!-- Body Tab -->
                  <a-tab-pane key="body" tab="Body">
                    <div style="margin: 16px 0">
                      <DiagnosticsResponseBody
                        :body="incident.diagnostics.response_body"
                        :is-encoded="incident.diagnostics.body_encoded"
                        :is-truncated="incident.diagnostics.body_truncated"
                        :response-size="incident.diagnostics.response_size"
                      />
                    </div>
                  </a-tab-pane>

                  <!-- Headers Tab -->
                  <a-tab-pane key="headers" tab="Headers">
                    <div style="margin: 16px 0">
                      <DiagnosticsHeadersDisplay
                        :headers="incident.diagnostics.response_headers"
                        title="Response Headers"
                        empty-message="No response headers available"
                      />
                    </div>
                  </a-tab-pane>
                </a-tabs>
              </div>

              <div v-else>
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
              </div>
            </a-card>

            <!-- Error Summary Card (when diagnostics available) -->
            <a-card v-if="hasDiagnostics && incident.diagnostics" style="margin-bottom: 16px">
              <template #title>
                <div style="font-size: 14px; font-weight: 600">Error Summary</div>
              </template>

              <DiagnosticsErrorSummary
                :error-summary="incident.diagnostics.error_summary"
                :failure-type="incident.diagnostics.failure_type"
                :error-message="incident.diagnostics.error_message"
              />
            </a-card>

            <!-- Timing Breakdown Card (when diagnostics available) -->
            <a-card v-if="hasDiagnostics && incident.diagnostics">
              <template #title>
                <div style="font-size: 14px; font-weight: 600">Performance Timing</div>
              </template>

              <DiagnosticsTimingBreakdown
                :total-duration="incident.diagnostics.total_duration"
                :dns-duration="incident.diagnostics.dns_duration"
                :tls-duration="incident.diagnostics.tls_duration"
                :first-byte-duration="incident.diagnostics.first_byte_duration"
              />
            </a-card>

            <a-card v-if="hasKeywordDiagnostics && incident.diagnostics" style="margin-top: 16px">
              <template #title>
                <div style="font-size: 14px; font-weight: 600">🔎 Keyword Diagnostics</div>
              </template>

              <div style="display: flex; flex-direction: column; gap: 12px">
                <div>
                  <div style="font-size: 12px; color: rgba(0,0,0,0.65); margin-bottom: 6px">KEYWORD</div>
                  <code style="background: #f5f5f5; padding: 2px 6px; border-radius: 4px">{{ incident.diagnostics.keyword }}</code>
                </div>
                <div>
                  <div style="font-size: 12px; color: rgba(0,0,0,0.65); margin-bottom: 6px">MATCH MODE</div>
                  <a-tag>{{ incident.diagnostics.keyword_mode }}</a-tag>
                </div>
                <div>
                  <div style="font-size: 12px; color: rgba(0,0,0,0.65); margin-bottom: 6px">KEYWORD FOUND</div>
                  <a-tag :color="incident.diagnostics.keyword_found ? 'green' : 'red'">
                    {{ incident.diagnostics.keyword_found ? 'Yes' : 'No' }}
                  </a-tag>
                </div>
                <div v-if="incident.diagnostics.response_body">
                  <div style="font-size: 12px; color: rgba(0,0,0,0.65); margin-bottom: 6px">
                    BODY EXCERPT
                    <a-tag v-if="incident.diagnostics.body_truncated" color="warning" style="margin-left: 6px">Truncated</a-tag>
                  </div>
                  <pre style="background: #f5f5f5; padding: 8px; border-radius: 4px; font-size: 12px; white-space: pre-wrap; word-break: break-all; max-height: 200px; overflow-y: auto">{{ incident.diagnostics.response_body }}</pre>
                </div>
                <div v-if="incident.diagnostics.response_size">
                  <div style="font-size: 12px; color: rgba(0,0,0,0.65); margin-bottom: 6px">BODY SIZE</div>
                  <div style="font-size: 14px">{{ incident.diagnostics.response_size }} bytes</div>
                </div>
              </div>
            </a-card>

            <a-card v-if="hasICMPDiagnostics && incident.diagnostics" style="margin-top: 16px">
              <template #title>
                <div style="font-size: 14px; font-weight: 600">🔍 Network Diagnostics (ICMP Ping)</div>
              </template>

              <div style="display: flex; flex-direction: column; gap: 12px">
                <!-- Root cause hint badge -->
                <div v-if="incident.diagnostics.root_cause_hint">
                  <div style="font-size: 12px; color: rgba(0,0,0,0.65); margin-bottom: 6px">ROOT CAUSE HINT</div>
                  <a-tag :color="rootCauseHintLabel.color">{{ rootCauseHintLabel.label }}</a-tag>
                </div>

                <!-- ICMP reachability -->
                <div>
                  <div style="font-size: 12px; color: rgba(0,0,0,0.65); margin-bottom: 6px">ICMP REACHABLE</div>
                  <a-tag :color="incident.diagnostics.icmp_reachable ? 'green' : 'red'">
                    {{ incident.diagnostics.icmp_reachable ? 'Yes' : 'No' }}
                  </a-tag>
                </div>

                <!-- RTT (only when host was reachable) -->
                <div v-if="incident.diagnostics.icmp_rtt_ms != null">
                  <div style="font-size: 12px; color: rgba(0,0,0,0.65); margin-bottom: 6px">RTT</div>
                  <div style="font-size: 16px; font-weight: 600">
                    {{ incident.diagnostics.icmp_rtt_ms }} ms
                  </div>
                </div>
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
