<script setup lang="ts">
import { computed } from 'vue'

type StatusType = 'up' | 'down' | 'paused' | 'pending' | 'error' | 'unknown'

interface StatusConfig {
  color: string
  label: string
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
}

const statusInfo = computed<StatusConfig>(() => {
  return statusConfigMap[props.status] ?? statusConfigMap.unknown
})
</script>

<template>
  <a-tag :color="statusInfo.color">
    {{ statusInfo.label }}
  </a-tag>
</template>

<style scoped></style>
