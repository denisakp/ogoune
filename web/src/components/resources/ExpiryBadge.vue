<script setup lang="ts">
import { computed } from 'vue'

/**
 * ExpiryBadge — Shows SSL or domain expiry status inline.
 * Renders nothing when status is 'ok'.
 */
interface Props {
  /** 'ssl' shows 🔒, 'domain' shows 🌐 */
  type: 'ssl' | 'domain'
  /** Days until expiry (negative = already expired) */
  daysRemaining: number | null | undefined
  /** Backend-computed severity level */
  status: 'ok' | 'warning' | 'critical' | 'expired'
}

const props = defineProps<Props>()

const color = computed<'warning' | 'error' | 'neutral'>(() => {
  if (props.status === 'warning') return 'warning'
  if (props.status === 'critical') return 'error'
  return 'neutral'
})

const icon = props.type === 'ssl' ? '🔒' : '🌐'

const label = (() => {
  if (props.daysRemaining == null) return props.type === 'ssl' ? 'SSL' : 'Domain'
  if (props.daysRemaining <= 0) return props.type === 'ssl' ? 'SSL expired' : 'Domain expired'
  const d = props.daysRemaining
  return `${props.type === 'ssl' ? 'SSL' : 'Domain'} ${d}d`
})()
</script>

<template>
  <UBadge v-if="status !== 'ok'" :color="color" variant="subtle" size="xs" class="cursor-default">
    {{ icon }} {{ label }}
  </UBadge>
</template>
