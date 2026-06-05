<script setup lang="ts">
import { computed } from 'vue'
interface UptimeBar {
  status: 'up' | 'down' | 'degraded' | 'no_data'
  tooltip?: string
}

interface Props {
  name: string
  uptimePercentage: number
  status: 'Operational' | 'Down' | 'Partial Outage'
  uptimeData?: UptimeBar[]
  showUptimePercentage?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  uptimeData: () => [],
  showUptimePercentage: true,
})

const uptimeBars = computed(() => props.uptimeData)

const statusColor = computed(() => {
  switch (props.status) {
    case 'Operational':
      return '#52c41a'
    case 'Down':
      return '#ff4d4f'
    case 'Partial Outage':
      return '#faad14'
    default:
      return '#d9d9d9'
  }
})

const getBarColor = (bar: UptimeBar): string => {
  switch (bar.status) {
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

const getBarTooltipText = (bar: UptimeBar): string => {
  switch (bar.status) {
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

const barTitle = (bar: UptimeBar): string => {
  const base = getBarTooltipText(bar)
  return bar.tooltip ? `${base}\n${bar.tooltip}` : base
}
</script>

<template>
  <div class="service-status-item">
    <div class="service-header">
      <div class="service-name">
        <span>{{ name }}</span>
        <UIcon name="i-lucide-arrow-right" class="arrow-icon" />
      </div>

      <div class="service-stats">
        <div v-if="showUptimePercentage" class="uptime-percentage">
          {{ uptimePercentage.toFixed(3) }}%
        </div>

        <div class="status-indicator">
          <span class="status-dot" :style="{ backgroundColor: statusColor }"></span>
          <span class="status-text">{{ status }}</span>
        </div>
      </div>
    </div>

    <div class="uptime-bar-container">
      <div
        v-for="(bar, index) in uptimeBars"
        :key="index"
        class="uptime-bar"
        :title="barTitle(bar)"
        :style="{ backgroundColor: getBarColor(bar) }"
      ></div>
    </div>
  </div>
</template>

<style scoped>
.service-status-item {
  padding: 20px;
  background: #ffffff;
  border: 1px solid #f0f0f0;
  border-radius: 8px;
  margin-bottom: 16px;
  transition: all 0.3s ease;
}

.service-status-item:hover {
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  border-color: #d9d9d9;
}

.service-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.service-name {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 600;
  color: rgba(0, 0, 0, 0.85);
}

.arrow-icon {
  color: rgba(0, 0, 0, 0.45);
  font-size: 14px;
  transition: transform 0.3s ease;
}

.service-status-item:hover .arrow-icon {
  transform: translateX(4px);
}

.service-stats {
  display: flex;
  align-items: center;
  gap: 24px;
}

.uptime-percentage {
  font-size: 14px;
  font-weight: 600;
  color: rgba(0, 0, 0, 0.65);
  min-width: 65px;
  text-align: right;
}

.status-indicator {
  display: flex;
  align-items: center;
  gap: 8px;
}

.status-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  display: inline-block;
}

.status-text {
  font-size: 14px;
  font-weight: 500;
  color: rgba(0, 0, 0, 0.85);
  min-width: 120px;
}

.uptime-bar-container {
  display: flex;
  gap: 2px;
  height: 40px;
  align-items: flex-end;
}

.uptime-bar {
  flex: 1;
  height: 100%;
  border-radius: 2px;
  transition: opacity 0.2s ease;
  min-width: 3px;
  cursor: default;
}

.uptime-bar:hover {
  opacity: 0.8;
}

@media (max-width: 768px) {
  .service-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }

  .service-stats {
    width: 100%;
    justify-content: space-between;
  }

  .uptime-percentage {
    min-width: auto;
  }

  .status-text {
    min-width: auto;
  }
}
</style>
