<script setup lang="ts">
import { computed } from 'vue'
import { CheckCircleOutlined, WarningOutlined, CloseCircleOutlined } from '@ant-design/icons-vue'

import ServiceStatusItem from './ServiceStatusItem.vue'
import type { GlobalStatus, ResourceStatusInfo, DailyStatus } from '@/types'

interface Service {
  id: string
  name: string
  uptimePercentage: number
  status: 'Operational' | 'Down' | 'Partial Outage'
  uptimeData: { status: 'up' | 'down' | 'degraded' | 'no_data' }[]
}

interface Props {
  globalStatus: GlobalStatus
  resources: ResourceStatusInfo[]
  loading?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
})

// Define emits
const emit = defineEmits<{
  'service-click': [serviceId: string]
}>()

// Map global status to display text
const overallStatus = computed<'Operational' | 'Some systems down' | 'All systems down'>(() => {
  if (props.globalStatus === 'all_systems_operational') {
    return 'Operational'
  }
  return 'Some systems down'
})

// Map API status to display status
const mapStatusToDisplay = (status: string): 'Operational' | 'Down' | 'Partial Outage' => {
  switch (status) {
    case 'up':
      return 'Operational'
    case 'down':
      return 'Down'
    case 'degraded':
      return 'Partial Outage'
    default:
      return 'Operational'
  }
}

// Map daily status to uptime bar format
const mapDailyStatusToBar = (
  dailyStatus: DailyStatus[],
): { status: 'up' | 'down' | 'degraded' | 'no_data' }[] => {
  return dailyStatus.map((status) => ({ status }))
}

// Convert resources to services format
const services = computed<Service[]>(() => {
  return props.resources.map((resource) => ({
    id: resource.id,
    name: resource.name,
    uptimePercentage: resource.uptime_percentage_last_90_days,
    status: mapStatusToDisplay(resource.current_status),
    uptimeData: mapDailyStatusToBar(resource.daily_status_last_90_days),
  }))
})

const getOverallStatusIcon = () => {
  switch (overallStatus.value) {
    case 'Operational':
      return CheckCircleOutlined
    case 'Some systems down':
      return WarningOutlined
    case 'All systems down':
      return CloseCircleOutlined
    default:
      return CheckCircleOutlined
  }
}

const getOverallStatusColor = () => {
  switch (overallStatus.value) {
    case 'Operational':
      return '#52c41a'
    case 'Some systems down':
      return '#faad14'
    case 'All systems down':
      return '#ff4d4f'
    default:
      return '#d9d9d9'
  }
}

const handleServiceClick = (serviceId: string) => {
  emit('service-click', serviceId)
}
</script>

<template>
  <div class="status-page">
    <div class="status-page-container">
      <!-- Header -->
      <div class="status-page-header">
        <h1 class="page-title">Status Page</h1>
        <p class="page-subtitle">Real-time service status and uptime information</p>
      </div>

      <!-- Overall Status Card -->
      <a-card class="overall-status-card" :bordered="false">
        <div class="overall-status-content">
          <component
            :is="getOverallStatusIcon()"
            class="status-icon"
            :style="{ color: getOverallStatusColor() }"
          />
          <div class="status-info">
            <h2 class="status-title">{{ overallStatus }}</h2>
            <p class="status-description">
              {{
                overallStatus === 'Operational'
                  ? 'All systems are functioning normally'
                  : 'Some services are experiencing issues'
              }}
            </p>
          </div>
        </div>
      </a-card>

      <!-- Services Section -->
      <div class="services-section">
        <div class="section-header">
          <h2 class="section-title">Services</h2>
          <p class="section-subtitle">Monitor the status of all services</p>
        </div>

        <a-spin :spinning="loading">
          <div v-if="!loading && services.length === 0" class="empty-state">
            <a-empty description="No services to monitor" />
          </div>
          <div v-else class="services-list">
            <div
              v-for="service in services"
              :key="service.id"
              class="service-item-wrapper"
              @click="handleServiceClick(service.id)"
            >
              <ServiceStatusItem
                :name="service.name"
                :uptime-percentage="service.uptimePercentage"
                :status="service.status"
                :uptime-data="service.uptimeData"
              />
            </div>
          </div>
        </a-spin>
      </div>

      <!-- Footer -->
      <div class="status-page-footer">
        <a-typography-text type="secondary">
          Last updated: {{ new Date().toLocaleString() }}
        </a-typography-text>
      </div>
    </div>
  </div>
</template>

<style scoped>
.status-page {
  min-height: 100vh;
  background: #f0f2f5;
  padding: 24px;
}

.status-page-container {
  max-width: 1200px;
  margin: 0 auto;
}

.status-page-header {
  margin-bottom: 32px;
  text-align: center;
}

.page-title {
  font-size: 32px;
  font-weight: 700;
  color: rgba(0, 0, 0, 0.85);
  margin-bottom: 8px;
}

.page-subtitle {
  font-size: 16px;
  color: rgba(0, 0, 0, 0.45);
  margin: 0;
}

.overall-status-card {
  margin-bottom: 32px;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.overall-status-content {
  display: flex;
  align-items: center;
  gap: 24px;
  padding: 24px;
}

.status-icon {
  font-size: 64px;
  flex-shrink: 0;
}

.status-info {
  flex: 1;
}

.status-title {
  font-size: 28px;
  font-weight: 700;
  color: rgba(0, 0, 0, 0.85);
  margin: 0 0 8px 0;
}

.status-description {
  font-size: 16px;
  color: rgba(0, 0, 0, 0.65);
  margin: 0;
}

.services-section {
  margin-bottom: 32px;
}

.section-header {
  margin-bottom: 24px;
}

.section-title {
  font-size: 24px;
  font-weight: 700;
  color: rgba(0, 0, 0, 0.85);
  margin: 0 0 8px 0;
}

.section-subtitle {
  font-size: 14px;
  color: rgba(0, 0, 0, 0.45);
  margin: 0;
}

.services-list {
  display: flex;
  flex-direction: column;
}

.service-item-wrapper {
  cursor: pointer;
  transition: transform 0.2s ease;
}

.service-item-wrapper:hover {
  transform: translateX(4px);
}

.empty-state {
  padding: 48px 0;
  text-align: center;
}

.status-page-footer {
  text-align: center;
  padding: 24px 0;
  border-top: 1px solid #f0f0f0;
  margin-top: 32px;
}

:deep(.ant-card-body) {
  padding: 0;
}

@media (max-width: 768px) {
  .status-page {
    padding: 16px;
  }

  .page-title {
    font-size: 24px;
  }

  .page-subtitle {
    font-size: 14px;
  }

  .overall-status-content {
    flex-direction: column;
    text-align: center;
    padding: 16px;
  }

  .status-icon {
    font-size: 48px;
  }

  .status-title {
    font-size: 22px;
  }

  .status-description {
    font-size: 14px;
  }

  .section-title {
    font-size: 20px;
  }
}
</style>
