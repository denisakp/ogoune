<script setup lang="ts">
import { computed, ref } from 'vue'
import ServiceStatusItem from './ServiceStatusItem.vue'
import type { GlobalStatus, ResourceStatusInfo, ComponentStatusInfo, DailyStatus } from '@/types'

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
  components?: ComponentStatusInfo[]
  loading?: boolean
  showUptimePercentage?: boolean
  enableDetailsPage?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  components: () => [],
  loading: false,
  showUptimePercentage: true,
  enableDetailsPage: true,
})

const openGroups = ref<string[]>([])

const emit = defineEmits<{
  'service-click': [serviceId: string]
}>()

const handleServiceClick = (serviceId: string) => {
  if (props.enableDetailsPage) {
    emit('service-click', serviceId)
  }
}

const overallStatus = computed<'Operational' | 'Some systems down' | 'All systems down'>(() => {
  if (props.globalStatus === 'all_systems_operational') {
    return 'Operational'
  }
  return 'Some systems down'
})

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

const mapDailyStatusToBar = (
  dailyStatus: DailyStatus[],
): { status: 'up' | 'down' | 'degraded' | 'no_data' }[] => {
  return dailyStatus.map((status) => ({ status }))
}

const services = computed<Service[]>(() => {
  return props.resources.map((resource) => ({
    id: resource.id,
    name: resource.name,
    uptimePercentage: resource.uptime_percentage_last_90_days,
    status: mapStatusToDisplay(resource.current_status),
    uptimeData: mapDailyStatusToBar(resource.daily_status_last_90_days),
  }))
})

const componentGroups = computed(() => {
  if (!props.components || props.components.length === 0) {
    return []
  }

  return props.components.map((component) => ({
    id: component.id,
    name: component.name,
    status: mapStatusToDisplay(component.status),
    resources: component.resources.map((resource) => ({
      id: resource.id,
      name: resource.name,
      status: mapStatusToDisplay(resource.current_status),
      uptimePercentage: resource.uptime_percentage_last_90_days,
      uptimeData: mapDailyStatusToBar(resource.daily_status_last_90_days),
    })),
  }))
})

const getOverallStatusIcon = () => {
  switch (overallStatus.value) {
    case 'Operational':
      return 'i-lucide-check-circle'
    case 'Some systems down':
      return 'i-lucide-alert-triangle'
    case 'All systems down':
      return 'i-lucide-x-circle'
    default:
      return 'i-lucide-check-circle'
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
      <div class="overall-status-card">
        <div class="overall-status-content">
          <UIcon
            :name="getOverallStatusIcon()"
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
      </div>

      <!-- Components Section (if available) -->
      <div v-if="componentGroups.length > 0" class="components-section">
        <div class="section-header">
          <h2 class="section-title">Components</h2>
          <p class="section-subtitle">Service groups and their status</p>
        </div>

        <UAccordion
          v-model="openGroups"
          type="multiple"
          :items="componentGroups.map((g) => ({ value: g.id, label: g.name, _group: g }))"
          class="components-list"
        >
          <template #default="{ item, open }">
            <div class="component-header">
              <UIcon
                :name="open ? 'i-lucide-chevron-up' : 'i-lucide-chevron-down'"
                class="size-4 shrink-0"
              />
              <span
                class="status-dot"
                :style="{
                  backgroundColor:
                    item._group.status === 'Operational'
                      ? '#52c41a'
                      : item._group.status === 'Down'
                        ? '#ff4d4f'
                        : '#faad14',
                }"
              ></span>
              <h3 class="component-name">{{ item._group.name }}</h3>
              <span
                class="component-status"
                :class="item._group.status.toLowerCase().replace(' ', '-')"
              >
                {{ item._group.status }}
              </span>
              <span class="component-resource-count">
                {{ item._group.resources.length }} service(s)
              </span>
            </div>
          </template>
          <template #content="{ item }">
            <div class="component-resources">
              <div v-if="item._group.resources.length === 0" class="no-resources">
                <p>No services in this component</p>
              </div>
              <div
                v-for="resource in item._group.resources"
                :key="resource.id"
                class="resource-item-wrapper"
                :class="{ clickable: enableDetailsPage }"
                @click.stop="handleServiceClick(resource.id)"
              >
                <ServiceStatusItem
                  :name="resource.name"
                  :uptime-percentage="resource.uptimePercentage"
                  :status="resource.status"
                  :uptime-data="resource.uptimeData"
                  :show-uptime-percentage="showUptimePercentage"
                />
              </div>
            </div>
          </template>
        </UAccordion>
      </div>

      <!-- Services Section -->
      <div class="services-section">
        <div class="section-header">
          <h2 class="section-title">
            {{ componentGroups.length > 0 ? 'Other Services' : 'Services' }}
          </h2>
          <p class="section-subtitle">
            {{
              componentGroups.length > 0
                ? 'Standalone services (not grouped into components)'
                : 'Monitor the status of all services'
            }}
          </p>
        </div>

        <div v-if="!loading && services.length === 0" class="empty-state">
          <UEmpty icon="i-lucide-radar" title="No services to monitor" />
        </div>
        <div v-else class="services-list">
          <div
            v-for="service in services"
            :key="service.id"
            class="service-item-wrapper"
            :class="{ clickable: enableDetailsPage }"
            @click="handleServiceClick(service.id)"
          >
            <ServiceStatusItem
              :name="service.name"
              :uptime-percentage="service.uptimePercentage"
              :status="service.status"
              :uptime-data="service.uptimeData"
              :show-uptime-percentage="showUptimePercentage"
            />
          </div>
        </div>
      </div>

      <!-- Footer -->
      <div class="status-page-footer">
        <span class="text-sm text-muted"> Last updated: {{ new Date().toLocaleString() }} </span>
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
  background: #ffffff;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.overall-status-content {
  display: flex;
  align-items: center;
  gap: 24px;
  padding: 24px;
}

.status-icon {
  width: 64px;
  height: 64px;
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

.services-section,
.components-section {
  margin-bottom: 32px;
}

.components-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.component-card {
  border-radius: 12px;
  background: #ffffff;
  padding: 16px 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.component-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 0;
  cursor: pointer;
  transition: background-color 0.2s ease;
}

.component-header:hover {
  background-color: #fafafa;
}

.expand-btn {
  flex-shrink: 0;
}

.status-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  display: inline-block;
  flex-shrink: 0;
}

.component-name {
  flex: 1;
  font-size: 16px;
  font-weight: 600;
  margin: 0;
  color: rgba(0, 0, 0, 0.85);
  min-width: 0;
}

.component-status {
  padding: 4px 12px;
  border-radius: 4px;
  font-weight: 500;
  font-size: 12px;
  text-transform: uppercase;
  flex-shrink: 0;
}

.component-status.operational {
  background-color: #f6ffed;
  color: #52c41a;
}

.component-status.down {
  background-color: #fff1f0;
  color: #ff4d4f;
}

.component-status.partial-outage {
  background-color: #fffbe6;
  color: #faad14;
}

.component-resource-count {
  font-size: 12px;
  color: #666;
  flex-shrink: 0;
  margin-left: auto;
}

.component-resources {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid #f0f0f0;
}

.no-resources {
  text-align: center;
  color: #999;
  padding: 20px 0;
  font-size: 13px;
}

.resource-item-wrapper {
  padding: 12px;
  background-color: #fafafa;
  border-radius: 6px;
  transition: all 0.2s ease;
  margin-left: 12px;
  border-left: 3px solid #f0f0f0;
}

.resource-item-wrapper.clickable {
  cursor: pointer;
}

.resource-item-wrapper.clickable:hover {
  transform: translateX(4px);
  background-color: #f5f5f5;
  border-left-color: #1890ff;
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
  transition: transform 0.2s ease;
}

.service-item-wrapper.clickable {
  cursor: pointer;
}

.service-item-wrapper.clickable:hover {
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
    width: 48px;
    height: 48px;
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
