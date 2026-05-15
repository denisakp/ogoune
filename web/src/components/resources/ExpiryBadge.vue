<script setup lang="ts">
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

const COLOR: Record<string, string> = {
  warning: '#faad14',
  critical: '#ff4d4f',
  expired: '#8c8c8c',
}

const icon = props.type === 'ssl' ? '🔒' : '🌐'

const label = (() => {
  if (props.daysRemaining == null) return props.type === 'ssl' ? 'SSL' : 'Domain'
  if (props.daysRemaining <= 0) return props.type === 'ssl' ? 'SSL expired' : 'Domain expired'
  const d = props.daysRemaining
  return `${props.type === 'ssl' ? 'SSL' : 'Domain'} ${d}d`
})()
</script>

<template>
  <a-tag
    v-if="status !== 'ok'"
    :color="COLOR[status]"
    style="margin: 0; font-size: 11px; line-height: 18px; padding: 0 6px; cursor: default"
  >
    {{ icon }} {{ label }}
  </a-tag>
</template>
