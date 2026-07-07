<script setup lang="ts">
import { computed } from 'vue'

type StatusType =
  | 'up'
  | 'down'
  | 'paused'
  | 'pending'
  | 'error'
  | 'unknown'
  | 'flapping'
  | 'waiting'

interface StatusConfig {
  color: 'success' | 'error' | 'warning' | 'info' | 'neutral'
  label: string
  pulse?: boolean
}

interface Props {
  status: StatusType
}

const props = defineProps<Props>()

const statusConfigMap: Record<StatusType, StatusConfig> = {
  up: { color: 'success', label: 'UP' },
  down: { color: 'error', label: 'DOWN' },
  paused: { color: 'warning', label: 'PAUSED' },
  pending: { color: 'info', label: 'PENDING' },
  error: { color: 'error', label: 'ERROR' },
  unknown: { color: 'neutral', label: 'UNKNOWN' },
  flapping: { color: 'warning', label: 'FLAPPING', pulse: true },
  waiting: { color: 'neutral', label: 'WAITING' },
}

const statusInfo = computed<StatusConfig>(() => {
  return statusConfigMap[props.status] ?? statusConfigMap.unknown
})
</script>

<template>
  <UBadge
    :color="statusInfo.color"
    variant="subtle"
    size="sm"
    :class="{ 'flapping-pulse': statusInfo.pulse }"
  >
    {{ statusInfo.label }}
  </UBadge>
</template>

<style scoped>
.flapping-pulse {
  animation: flapping-blink 1.4s ease-in-out infinite;
}

@keyframes flapping-blink {
  0%,
  100% {
    opacity: 1;
  }
  50% {
    opacity: 0.55;
  }
}
</style>
