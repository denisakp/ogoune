<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type {
  PublicMonitorDetail,
  MonitorRecentEvent,
  DailyStatus,
  ResourceCurrentStatus,
  MaintenanceBanner,
} from '@/types'

interface Props {
  monitorData: PublicMonitorDetail | null
  loading?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
})

const INITIAL_EVENTS_COUNT = 5
const visibleEventsCount = ref(INITIAL_EVENTS_COUNT)

const getStatusColor = (status: ResourceCurrentStatus) => {
  switch (status) {
    case 'up':
      return '#52c41a'
    case 'down':
      return '#ff4d4f'
    case 'degraded':
      return '#faad14'
    default:
      return '#d9d9d9'
  }
}

const getStatusText = (status: ResourceCurrentStatus) => {
  switch (status) {
    case 'up':
      return 'operational'
    case 'down':
      return 'down'
    case 'degraded':
      return 'experiencing issues'
    default:
      return 'unknown'
  }
}

const getBarColor = (status: DailyStatus): string => {
  switch (status) {
    case 'up':
      return '#52c41a'
    case 'down':
      return '#ff4d4f'
    case 'degraded':
      return '#faad14'
    case 'no_data':
      return '#d9d9d9'
    default:
      return '#d9d9d9'
  }
}

const getBarTooltipText = (status: DailyStatus): string => {
  switch (status) {
    case 'up':
      return 'Operational'
    case 'down':
      return 'Down'
    case 'degraded':
      return 'Partial Outage'
    case 'no_data':
      return 'Not Monitored'
    default:
      return 'Unknown'
  }
}

const formatTimestamp = (timestamp: string): string => {
  try {
    const date = new Date(timestamp)
    const monthNames = [
      'January',
      'February',
      'March',
      'April',
      'May',
      'June',
      'July',
      'August',
      'September',
      'October',
      'November',
      'December',
    ]
    const month = monthNames[date.getMonth()]
    const day = date.getDate()
    const year = date.getFullYear()
    const hours = String(date.getHours()).padStart(2, '0')
    const minutes = String(date.getMinutes()).padStart(2, '0')
    const timezoneOffset = -date.getTimezoneOffset()
    const offsetHours = String(Math.floor(Math.abs(timezoneOffset) / 60)).padStart(2, '0')
    const offsetMinutes = String(Math.abs(timezoneOffset) % 60).padStart(2, '0')
    const offsetSign = timezoneOffset >= 0 ? '+' : '-'
    const timezone = `${offsetSign}${offsetHours}:${offsetMinutes}`
    return `${month} ${day}, ${year} at ${hours}:${minutes} (${timezone})`
  } catch {
    return timestamp
  }
}

const getEventTitle = (event: MonitorRecentEvent): string => {
  if (event.type === 'up') {
    return event.reason || 'Running again'
  }
  if (event.duration) {
    return `Down for ${event.duration}`
  }
  return event.reason || 'Service down'
}

const maintenanceBanner = computed<MaintenanceBanner | null>(() => {
  return props.monitorData?.maintenance ?? null
})

const formatTimestampWithTimezone = (timestamp: string, tz?: string | null): string => {
  if (tz) {
    try {
      const date = new Date(timestamp)
      const monthNames = [
        'January',
        'February',
        'March',
        'April',
        'May',
        'June',
        'July',
        'August',
        'September',
        'October',
        'November',
        'December',
      ]
      const month = monthNames[date.getMonth()]
      const day = date.getDate()
      const year = date.getFullYear()
      const hours = String(date.getHours()).padStart(2, '0')
      const minutes = String(date.getMinutes()).padStart(2, '0')
      return `${month} ${day}, ${year} at ${hours}:${minutes} (${tz})`
    } catch {
      return `${timestamp} (${tz})`
    }
  }
  return formatTimestamp(timestamp)
}

const uptimeBars = computed(() => {
  if (!props.monitorData?.uptime_history_90_days) return []
  return props.monitorData.uptime_history_90_days
})

const visibleEvents = computed(() => {
  if (!props.monitorData?.recent_events) return []
  return props.monitorData.recent_events.slice(0, visibleEventsCount.value)
})

const hasMoreEvents = computed(() => {
  if (!props.monitorData?.recent_events) return false
  return visibleEventsCount.value < props.monitorData.recent_events.length
})

const allEventsShown = computed(() => {
  if (!props.monitorData?.recent_events) return false
  return (
    visibleEventsCount.value >= props.monitorData.recent_events.length &&
    props.monitorData.recent_events.length > 0
  )
})

const loadMoreEvents = () => {
  if (!props.monitorData?.recent_events) return
  const remaining = props.monitorData.recent_events.length - visibleEventsCount.value
  visibleEventsCount.value += Math.min(remaining, 5)
}

watch(
  () => props.monitorData?.id,
  () => {
    visibleEventsCount.value = INITIAL_EVENTS_COUNT
  },
)
</script>

<template>
  <div class="monitor-status-detail">
    <div class="detail-container">
      <!-- Loading State -->
      <div v-if="loading" class="loading-container">
        <UIcon name="i-lucide-loader-circle" class="size-8 animate-spin text-primary-500" />
      </div>

      <!-- Error State -->
      <div v-else-if="!monitorData" class="error-container">
        <UEmptyState icon="i-lucide-circle-alert" title="Failed to load monitor details. Please try again later." />
      </div>

      <!-- Content -->
      <div v-else>
        <!-- Maintenance Banner -->
        <div v-if="maintenanceBanner" class="maintenance-banner">
          <UAlert
            :color="maintenanceBanner.status === 'active' ? 'info' : 'warning'"
            variant="subtle"
            :icon="maintenanceBanner.status === 'active' ? 'i-lucide-wrench' : 'i-lucide-clock'"
            :title="
              maintenanceBanner.status === 'active'
                ? 'Maintenance in progress'
                : 'Scheduled maintenance'
            "
            :description="
              maintenanceBanner.status === 'active'
                ? maintenanceBanner.end_at
                  ? `Expected end: ${formatTimestampWithTimezone(maintenanceBanner.end_at, maintenanceBanner.timezone || undefined)}`
                  : 'Ongoing'
                : maintenanceBanner.start_at
                  ? `Planned for ${formatTimestampWithTimezone(maintenanceBanner.start_at, maintenanceBanner.timezone || undefined)}`
                  : ''
            "
          />
        </div>
        <!-- Header Card -->
        <div class="header-card">
          <div class="header-content">
            <div class="monitor-info">
              <h1 class="monitor-name">{{ monitorData.name }}</h1>
              <div class="monitor-status">
                <span>is</span>
                <span
                  class="status-text"
                  :style="{ color: getStatusColor(monitorData.current_status) }"
                >
                  {{ getStatusText(monitorData.current_status) }}
                </span>
              </div>
            </div>
          </div>
        </div>

        <!-- Uptime Section -->
        <div class="uptime-section">
          <div class="section-header">
            <h2 class="section-title">
              <UIcon name="i-lucide-layout-dashboard" style="margin-right: 8px" />
              Uptime
            </h2>
            <span class="section-subtitle">Last 90 days</span>
          </div>

          <div class="uptime-card">
            <div class="uptime-display">
              <div class="uptime-percentage-large">
                {{ monitorData.uptime_summary.last_90_days.toFixed(3) }}%
              </div>
              <div class="uptime-bar-container">
                <div
                  v-for="(bar, index) in uptimeBars"
                  :key="index"
                  class="uptime-bar"
                  :title="`Day ${index + 1} — ${getBarTooltipText(bar)}`"
                  :style="{ backgroundColor: getBarColor(bar) }"
                ></div>
              </div>
            </div>
          </div>
        </div>

        <!-- Overall Uptime Card -->
        <div class="overall-uptime-card">
          <div class="card-header">
            <span class="card-title">Overall Uptime</span>
          </div>

          <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
            <div class="stat-item">
              <div class="stat-value">
                {{ monitorData.uptime_summary.last_24_hours.toFixed(3) }}%
              </div>
              <div class="stat-label">Last 24 hours</div>
            </div>
            <div class="stat-item">
              <div class="stat-value">
                {{ monitorData.uptime_summary.last_7_days.toFixed(3) }}%
              </div>
              <div class="stat-label">Last 7 days</div>
            </div>
            <div class="stat-item">
              <div class="stat-value">
                {{ monitorData.uptime_summary.last_30_days.toFixed(3) }}%
              </div>
              <div class="stat-label">Last 30 days</div>
            </div>
            <div class="stat-item">
              <div class="stat-value">
                {{ monitorData.uptime_summary.last_90_days.toFixed(3) }}%
              </div>
              <div class="stat-label">Last 90 days</div>
            </div>
          </div>
        </div>

        <!-- Response Time Card -->
        <div class="response-time-card">
          <div class="card-header">
            <div class="card-title-with-subtitle">
              <span class="card-title">
                <UIcon name="i-lucide-clock" style="margin-right: 8px" />
                Response Time
              </span>
              <span class="card-subtitle">Last 7 days</span>
            </div>
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
            <div class="stat-item">
              <div class="stat-value">{{ monitorData.response_time_summary_7_days.avg_ms }}ms</div>
              <div class="stat-label">avg. response time</div>
            </div>
            <div class="stat-item">
              <div class="stat-value">{{ monitorData.response_time_summary_7_days.max_ms }}ms</div>
              <div class="stat-label">max. response time</div>
            </div>
            <div class="stat-item">
              <div class="stat-value">{{ monitorData.response_time_summary_7_days.min_ms }}ms</div>
              <div class="stat-label">min. response time</div>
            </div>
          </div>
        </div>

        <!-- Recent Events List -->
        <div v-if="monitorData.recent_events.length > 0" class="recent-events-section">
          <div class="section-header">
            <h2 class="section-title">Recent events</h2>
          </div>

          <div class="events-list">
            <div v-for="(event, index) in visibleEvents" :key="index" class="event-item">
              <div class="event-icon-wrapper">
                <div
                  class="event-icon"
                  :class="{
                    'event-icon-down': event.type === 'down',
                    'event-icon-up': event.type === 'up',
                  }"
                >
                  <UIcon v-if="event.type === 'down'" name="i-lucide-arrow-down" />
                  <UIcon v-if="event.type === 'up'" name="i-lucide-arrow-up" />
                </div>
                <div v-if="index !== visibleEvents.length - 1" class="event-line"></div>
              </div>

              <div class="event-content">
                <div class="event-title">{{ getEventTitle(event) }}</div>
                <div v-if="event.details" class="event-details">Details: {{ event.details }}</div>
                <div class="event-timestamp">{{ formatTimestamp(event.timestamp) }}</div>
              </div>
            </div>
          </div>

          <!-- Load More / All Events Shown -->
          <div class="events-footer">
            <UButton
              v-if="hasMoreEvents"
              color="primary"
              variant="link"
              class="load-more-button"
              @click="loadMoreEvents"
            >
              Load More Events
            </UButton>
            <div v-else-if="allEventsShown" class="all-events-message">That's all mate! 🎉</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.monitor-status-detail {
  min-height: 100vh;
  background: #f0f2f5;
  padding: 24px;
}

.detail-container {
  max-width: 1200px;
  margin: 0 auto;
}

.loading-container,
.error-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 400px;
}

.header-card,
.uptime-card,
.overall-uptime-card,
.response-time-card {
  background: #ffffff;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  padding: 24px;
  margin-bottom: 24px;
}

.maintenance-banner {
  margin-bottom: 16px;
}

.header-content {
  padding: 16px 0;
}

.monitor-info {
  text-align: center;
}

.monitor-name {
  font-size: 32px;
  font-weight: 700;
  color: rgba(0, 0, 0, 0.85);
  margin: 0 0 12px 0;
  word-break: break-word;
}

.monitor-status {
  font-size: 20px;
  color: rgba(0, 0, 0, 0.65);
}

.status-text {
  font-weight: 700;
  margin-left: 8px;
}

.uptime-section {
  margin-bottom: 24px;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.section-title {
  font-size: 20px;
  font-weight: 600;
  color: rgba(0, 0, 0, 0.85);
  margin: 0;
  display: flex;
  align-items: center;
}

.section-subtitle {
  font-size: 14px;
  color: rgba(0, 0, 0, 0.45);
}

.uptime-display {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.uptime-percentage-large {
  font-size: 48px;
  font-weight: 700;
  color: rgba(0, 0, 0, 0.85);
  text-align: center;
}

.uptime-bar-container {
  display: flex;
  gap: 2px;
  height: 60px;
  align-items: flex-end;
}

.uptime-bar {
  flex: 1;
  height: 100%;
  border-radius: 2px;
  transition: opacity 0.2s ease;
  min-width: 4px;
  cursor: default;
}

.uptime-bar:hover {
  opacity: 0.8;
}

.card-header {
  margin-bottom: 16px;
}

.card-title {
  font-size: 18px;
  font-weight: 600;
  color: rgba(0, 0, 0, 0.85);
  display: flex;
  align-items: center;
}

.card-title-with-subtitle {
  display: flex;
  align-items: center;
  gap: 12px;
}

.card-subtitle {
  font-size: 14px;
  font-weight: 400;
  color: rgba(0, 0, 0, 0.45);
}

.stat-item {
  text-align: center;
  padding: 16px 0;
}

.stat-value {
  font-size: 32px;
  font-weight: 700;
  color: rgba(0, 0, 0, 0.85);
  margin-bottom: 8px;
}

.stat-label {
  font-size: 14px;
  color: rgba(0, 0, 0, 0.45);
}

.recent-events-section {
  margin-bottom: 24px;
}

.events-list {
  background: #ffffff;
  border-radius: 8px;
  padding: 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.event-item {
  display: flex;
  gap: 16px;
  position: relative;
}

.event-item:not(:last-child) {
  margin-bottom: 24px;
}

.event-icon-wrapper {
  display: flex;
  flex-direction: column;
  align-items: center;
  flex-shrink: 0;
}

.event-icon {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  color: #ffffff;
}

.event-icon-down {
  background-color: #ff4d4f;
}

.event-icon-up {
  background-color: #52c41a;
}

.event-line {
  width: 2px;
  flex: 1;
  background-color: #f0f0f0;
  margin-top: 8px;
}

.event-content {
  flex: 1;
  padding-top: 8px;
}

.event-title {
  font-size: 16px;
  font-weight: 600;
  color: rgba(0, 0, 0, 0.85);
  margin-bottom: 4px;
}

.event-details {
  font-size: 14px;
  color: rgba(0, 0, 0, 0.65);
  margin-bottom: 8px;
  line-height: 1.5;
}

.event-timestamp {
  font-size: 12px;
  color: rgba(0, 0, 0, 0.45);
}

.events-footer {
  margin-top: 24px;
  text-align: center;
  padding: 16px;
}

.load-more-button {
  font-size: 14px;
  font-weight: 500;
}

.all-events-message {
  font-size: 14px;
  color: #000;
  font-weight: 500;
  padding: 8px 0;
}

@media (max-width: 768px) {
  .monitor-status-detail {
    padding: 16px;
  }

  .monitor-name {
    font-size: 24px;
  }

  .monitor-status {
    font-size: 16px;
  }

  .uptime-percentage-large {
    font-size: 36px;
  }

  .stat-value {
    font-size: 24px;
  }

  .stat-item {
    padding: 12px 0;
  }

  .section-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }
}
</style>
