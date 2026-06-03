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
  color: string
  label: string
  pulse?: boolean
}

interface Props {
  status: StatusType
}

const props = defineProps<Props>()

const statusConfigMap: Record<StatusType, StatusConfig> = {
  up: { color: 'green', label: 'UP' },
  down: { color: 'red', label: 'DOWN' },
  paused: { color: 'orange', label: 'PAUSED' },
  pending: { color: 'blue', label: 'PENDING' },
  error: { color: 'red', label: 'ERROR' },
  unknown: { color: 'default', label: 'UNKNOWN' },
  flapping: { color: 'warning', label: 'FLAPPING', pulse: true },
  waiting: { color: 'default', label: 'WAITING' },
}

const statusInfo = computed<StatusConfig>(() => {
  return statusConfigMap[props.status] ?? statusConfigMap.unknown
})
</script>

<template>
  <a-tag :color="statusInfo.color" :class="{ 'flapping-pulse': statusInfo.pulse }">
    {{ statusInfo.label }}
  </a-tag>
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
