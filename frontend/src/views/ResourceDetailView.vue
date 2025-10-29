<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { message } from 'ant-design-vue'

import { useResources } from '@/composables/useResources'
import ResourceModal from '@/components/ResourceModal.vue'
import type { Resource, Incident } from '@/types'
import {
  PauseOutlined,
  PlayCircleOutlined,
  ArrowLeftOutlined,
  EditOutlined,
  EllipsisOutlined,
} from '@ant-design/icons-vue'

const router = useRouter()
const route = useRoute()

// Use the composable following the project architecture
const {
  loading,
  pauseResource,
  loadResource: loadResourceFromStore,
  testNotification,
} = useResources()

// Local state
const resource = ref<Resource | null>(null)
const timeRange = ref<'24h' | '7d' | '30d' | '365d'>('24h')
const incidentsToShow = ref(3)
const showEditModal = ref(false)

// Get resource ID from route
const resourceId = computed(() => route.params.id as string)

// Generate mock stats based on timeRange
const generateMockStats = (range: string) => {
  const ranges = {
    '24h': { uptime: 100, incidents: 0 },
    '7d': { uptime: 100, incidents: 0 },
    '30d': { uptime: 99.5, incidents: 1 },
    '365d': { uptime: 98.2, incidents: 5 },
  }
  return ranges[range as keyof typeof ranges] || ranges['24h']
}

// Get current stats
const currentStats = computed(() => {
  return generateMockStats(timeRange.value)
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

// Fetch resource on mount
onMounted(async () => {
  await loadResource()
})

// Load resource using the store
const loadResource = async () => {
  if (!resourceId.value) {
    message.error('Resource ID not found')
    return
  }

  try {
    const data = await loadResourceFromStore(resourceId.value)
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
  }
  return texts[status] || status
}

// Pause resource
const handlePauseResource = async () => {
  if (!resource.value) return
  try {
    await pauseResource(resource.value.id)
    message.success('Resource paused successfully')
    await loadResource()
  } catch (err) {
    const errorMsg = err instanceof Error ? err.message : 'Failed to pause resource'
    message.error(errorMsg)
    console.error('Error pausing resource:', err)
  }
}

// Test notification
const handleTestNotification = async () => {
  if (!resource.value) return

  try {
    await testNotification(resource.value.id)
    message.success('Test notification sent successfully')
  } catch (err) {
    const errorMsg = err instanceof Error ? err.message : 'Failed to send test notification'
    message.error(errorMsg)
  }
}

// Handle edit modal
const openEditModal = () => {
  showEditModal.value = true
}

const handleEditSubmit = async () => {
  showEditModal.value = false
  await loadResource()
  message.success('Monitor updated successfully')
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
                  {{ resource.type.toUpperCase() }} monitor for {{ resource.target }}
                </p>
              </div>
            </div>
          </div>
          <div style="display: flex; gap: 8px">
            <a-button @click="handleTestNotification">
              <template #icon>
                <PlayCircleOutlined />
              </template>
              Test Notification
            </a-button>
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
        <a-row :gutter="24">
          <!-- Left Column -->
          <a-col :xs="24" :lg="16">
            <!-- Current Status -->
            <a-card style="margin-bottom: 16px">
              <template #title>
                <div style="font-size: 14px; font-weight: 600">Current status</div>
              </template>
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

              <a-row :gutter="24">
                <a-col :xs="24" :sm="12">
                  <div style="text-align: center; padding: 24px">
                    <div style="font-size: 48px; font-weight: bold; color: #52c41a">
                      {{ currentStats.uptime }}%
                    </div>
                    <div style="font-size: 14px; color: rgba(0, 0, 0, 0.65); margin-top: 8px">
                      Uptime
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

              <!-- Response Time Chart Placeholder -->
              <div
                style="
                  height: 200px;
                  background: linear-gradient(180deg, rgba(24, 144, 255, 0.1) 0%, transparent 100%);
                  border-radius: 8px;
                  display: flex;
                  align-items: center;
                  justify-content: center;
                  margin-top: 16px;
                "
              >
                <div style="text-align: center; color: rgba(0, 0, 0, 0.45)">
                  <a-icon-line-chart style="font-size: 48px; margin-bottom: 8px" />
                  <div>Response time chart</div>
                </div>
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
                <a-tag v-for="tag in resource.tags" :key="tag.id" color="blue" style="margin: 0">
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
              <div style="text-align: center; padding: 24px; color: rgba(0, 0, 0, 0.45)">
                <a-icon-info-circle style="font-size: 32px; margin-bottom: 8px" />
                <div>Monitor is {{ resource.is_active ? 'active' : 'inactive' }}</div>
              </div>
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
