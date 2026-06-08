<script setup lang="ts">
import { computed } from 'vue'
import type { Resource } from '@/types'
import { formatDate } from '@/utils/formatters'

const props = defineProps<{ resource: Resource; nowTs: number }>()

const isHeartbeat = computed(() => props.resource.type === 'heartbeat')

const effectiveStatus = computed(() =>
  props.resource.waiting ? 'waiting' : (props.resource.status ?? 'unknown'),
)

const isConfirming = computed(
  () =>
    props.resource.status === 'down' &&
    props.resource.failure_count > 0 &&
    props.resource.failure_count < props.resource.confirmation_checks,
)

const confirmationProgress = computed(
  () => `${props.resource.failure_count}/${props.resource.confirmation_checks}`,
)

const nextConfirmationCountdown = computed(() => {
  if (!props.resource.last_checked) return 'n/a'
  const nextTs =
    new Date(props.resource.last_checked).getTime() + props.resource.confirmation_interval * 1000
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
  up: 'Up',
  down: 'Down',
  paused: 'Paused',
  pending: 'Pending',
  error: 'Error',
  waiting: 'Waiting',
}
</script>

<template>
  <UCard class="mb-4">
    <template #header>
      <div class="text-sm font-semibold">Current status</div>
    </template>
    <UAlert
      v-if="isConfirming"
      color="warning"
      variant="soft"
      icon="i-lucide-triangle-alert"
      class="mb-4"
      :title="`Confirming outage: ${confirmationProgress}`"
      :description="`Next confirmation check in ${nextConfirmationCountdown}`"
    />
    <UAlert
      v-if="isFlapping"
      color="warning"
      variant="soft"
      icon="i-lucide-triangle-alert"
      class="mb-4"
      title="Service is flapping"
      :description="`${flappingTransitionText}${flappingDuration ? ` over ${flappingDuration}` : ''}. Alerts suppressed until service stabilizes.`"
    />
    <UAlert
      v-if="isHeartbeat && resource.waiting"
      color="info"
      variant="soft"
      icon="i-lucide-info"
      class="mb-4"
      data-testid="heartbeat-waiting-alert"
      title="Waiting for first ping"
      description="This monitor will transition to UP as soon as it receives its first ping."
    />
    <div class="grid grid-cols-2 sm:grid-cols-3 gap-4">
      <div class="text-center">
        <div class="text-3xl font-bold" :style="{ color: statusColor }">
          {{ statusTexts[effectiveStatus] || effectiveStatus }}
        </div>
        <div class="text-xs text-muted mt-2">
          Currently {{ resource.is_active ? 'active' : 'inactive' }}
        </div>
      </div>
      <div class="text-center">
        <div class="text-2xl font-bold text-amber-500">{{ resource.failure_count }}</div>
        <div class="text-xs text-muted mt-2">Failures</div>
      </div>
      <div class="text-center col-span-2 sm:col-span-1">
        <div class="text-xs text-muted mb-2">Last checked</div>
        <div class="text-xs font-semibold">
          {{ resource.last_checked ? formatDate(resource.last_checked) : 'Never' }}
        </div>
      </div>
    </div>
  </UCard>
</template>
