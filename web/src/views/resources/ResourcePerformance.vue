<script setup lang="ts">
import { computed } from 'vue'
import { DashboardOutlined, RiseOutlined, FallOutlined } from '@ant-design/icons-vue'
import ResponseTimeChart from '@/components/ResponseTimeChart.vue'
import type { Resource } from '@/types'
import { getTimeRangeCutoff } from '@/libs/date-time.helper'

const props = defineProps<{ resource: Resource }>()
const timeRange = defineModel<'24h' | '7d' | '30d' | '365d'>('timeRange', { required: true })

const filteredIncidents = computed(() => {
  if (!props.resource.incidents) return []
  const cutoff = getTimeRangeCutoff(timeRange.value)
  return props.resource.incidents.filter((i) => new Date(i.started_at) >= cutoff)
})

const calculateUptime = computed((): number => {
  if (!props.resource || props.resource.status === 'pending' || !props.resource.last_checked) return -1
  if (props.resource.uptime !== undefined && timeRange.value === '24h') return Number(props.resource.uptime.toFixed(1))
  const cutoff = getTimeRangeCutoff(timeRange.value)
  const now = new Date()
  const totalDuration = now.getTime() - cutoff.getTime()
  if (totalDuration <= 0) return 100
  let totalDowntime = 0
  filteredIncidents.value.forEach((incident) => {
    const start = new Date(incident.started_at)
    const end = incident.resolved_at ? new Date(incident.resolved_at) : now
    const effectiveStart = start > cutoff ? start : cutoff
    const downtime = end.getTime() - effectiveStart.getTime()
    if (downtime > 0) totalDowntime += downtime
  })
  const uptime = ((totalDuration - totalDowntime) / totalDuration) * 100
  return Number(Math.max(0, Math.min(100, uptime)).toFixed(1))
})

const currentStats = computed(() => ({
  uptime: calculateUptime.value >= 0 ? calculateUptime.value : null,
  incidents: filteredIncidents.value.length,
}))
</script>

<template>
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
    <a-row :gutter="24" style="margin-bottom: 24px">
      <a-col :xs="24" :sm="12">
        <div style="text-align: center; padding: 24px">
          <div style="font-size: 48px; font-weight: bold" :style="{ color: currentStats.uptime === null ? '#d9d9d9' : '#52c41a' }">
            {{ currentStats.uptime !== null ? currentStats.uptime + '%' : 'Pending' }}
          </div>
          <div style="font-size: 14px; color: rgba(0,0,0,0.65); margin-top: 8px">
            {{ currentStats.uptime === null ? 'Waiting for first check' : 'Uptime' }}
          </div>
        </div>
      </a-col>
      <a-col :xs="24" :sm="12">
        <div style="text-align: center; padding: 24px">
          <div style="font-size: 48px; font-weight: bold; color: #f5222d">{{ currentStats.incidents }}</div>
          <div style="font-size: 14px; color: rgba(0,0,0,0.65); margin-top: 8px">Incidents</div>
        </div>
      </a-col>
    </a-row>
    <div v-if="resource.response_times && resource.response_times.length > 0"
      style="display: grid; grid-template-columns: repeat(3, 1fr); gap: 16px; padding: 16px; background: rgba(0,0,0,0.02); border-radius: 8px; margin-bottom: 16px">
      <div style="text-align: center">
        <DashboardOutlined style="font-size: 24px; color: #1890ff; margin-bottom: 8px" />
        <div style="font-size: 20px; font-weight: 600; color: #1890ff; margin-bottom: 4px">
          {{ (resource.response_times.reduce((sum, r) => sum + r.response_time, 0) / resource.response_times.length).toFixed(0) }}ms
        </div>
        <div style="font-size: 12px; color: rgba(0,0,0,0.45)">Average</div>
      </div>
      <div style="text-align: center">
        <RiseOutlined style="font-size: 24px; color: #52c41a; margin-bottom: 8px" />
        <div style="font-size: 20px; font-weight: 600; color: #52c41a; margin-bottom: 4px">
          {{ Math.min(...resource.response_times.map((r) => r.response_time)) }}ms
        </div>
        <div style="font-size: 12px; color: rgba(0,0,0,0.45)">Min</div>
      </div>
      <div style="text-align: center">
        <FallOutlined style="font-size: 24px; color: #ff4d4f; margin-bottom: 8px" />
        <div style="font-size: 20px; font-weight: 600; color: #ff4d4f; margin-bottom: 4px">
          {{ Math.max(...resource.response_times.map((r) => r.response_time)) }}ms
        </div>
        <div style="font-size: 12px; color: rgba(0,0,0,0.45)">Max</div>
      </div>
    </div>
    <div><ResponseTimeChart :data="resource.response_times" :height="300" /></div>
  </a-card>
</template>
