<script setup lang="ts">
import { computed } from 'vue'
import type { Resource } from '@/types'
import { formatDate } from '@/utils/formatters'

const props = defineProps<{ resource: Resource; nowTs: number }>()

const isHeartbeat = computed(() => props.resource.type === 'heartbeat')

const effectiveStatus = computed(() =>
  props.resource.waiting ? 'waiting' : (props.resource.status ?? 'unknown'),
)

const isConfirming = computed(() =>
  props.resource.status === 'down' &&
  props.resource.failure_count > 0 &&
  props.resource.failure_count < props.resource.confirmation_checks,
)

const confirmationProgress = computed(() =>
  `${props.resource.failure_count}/${props.resource.confirmation_checks}`,
)

const nextConfirmationCountdown = computed(() => {
  if (!props.resource.last_checked) return 'n/a'
  const nextTs = new Date(props.resource.last_checked).getTime() + props.resource.confirmation_interval * 1000
  const remainingSec = Math.max(0, Math.ceil((nextTs - props.nowTs) / 1000))
  return `${remainingSec}s`
})

const isFlapping = computed(() => props.resource.status === 'flapping')

const flappingDuration = computed(() => {
  if (!props.resource.flap_started_at) return ''
  const diff = props.nowTs - new Date(props.resource.flap_started_at).getTime()
  const hours = Math.floor(diff / (1000 * 60 * 60))
  const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60))
  if (hours > 0) return `${hours}h ${minutes}m`
  return `${minutes}m`
})

const flappingTransitionText = computed(() => {
  const threshold = props.resource.flap_threshold
  if (!threshold || threshold < 1) return 'multiple status transitions'
  return `${threshold}+ status transitions`
})

const statusColor = computed(() => {
  const s = effectiveStatus.value
  if (s === 'down') return '#ff4d4f'
  if (s === 'paused') return '#d9d9d9'
  if (s === 'waiting') return '#8c8c8c'
  return '#52c41a'
})

const statusTexts: Record<string, string> = {
  up: 'Up', down: 'Down', paused: 'Paused', pending: 'Pending', error: 'Error', waiting: 'Waiting',
}
</script>

<template>
  <a-card style="margin-bottom: 16px">
    <template #title>
      <div style="font-size: 14px; font-weight: 600">Current status</div>
    </template>
    <a-alert v-if="isConfirming" style="margin-bottom: 16px" type="warning" show-icon
      :message="`Confirming outage: ${confirmationProgress}`"
      :description="`Next confirmation check in ${nextConfirmationCountdown}`" />
    <a-alert v-if="isFlapping" style="margin-bottom: 16px" type="warning" show-icon
      message="Service is flapping"
      :description="`${flappingTransitionText}${flappingDuration ? ` over ${flappingDuration}` : ''}. Alerts suppressed until service stabilizes.`" />
    <a-alert v-if="isHeartbeat && resource.waiting" style="margin-bottom: 16px" type="info" show-icon
      data-testid="heartbeat-waiting-alert"
      message="Waiting for first ping"
      description="This monitor will transition to UP as soon as it receives its first ping." />
    <a-row :gutter="16">
      <a-col :xs="12" :sm="8">
        <div style="text-align: center">
          <div style="font-size: 28px; font-weight: bold" :style="{ color: statusColor }">
            {{ statusTexts[effectiveStatus] || effectiveStatus }}
          </div>
          <div style="font-size: 12px; color: rgba(0, 0, 0, 0.65); margin-top: 8px">
            Currently {{ resource.is_active ? 'active' : 'inactive' }}
          </div>
        </div>
      </a-col>
      <a-col :xs="12" :sm="8">
        <div style="text-align: center">
          <div style="font-size: 24px; font-weight: bold; color: #faad14">{{ resource.failure_count }}</div>
          <div style="font-size: 12px; color: rgba(0, 0, 0, 0.65); margin-top: 8px">Failures</div>
        </div>
      </a-col>
      <a-col :xs="24" :sm="8">
        <div style="text-align: center">
          <div style="font-size: 12px; color: rgba(0, 0, 0, 0.65); margin-bottom: 8px">Last checked</div>
          <div style="font-size: 12px; font-weight: 600">
            {{ resource.last_checked ? formatDate(resource.last_checked) : 'Never' }}
          </div>
        </div>
      </a-col>
    </a-row>
  </a-card>
</template>
