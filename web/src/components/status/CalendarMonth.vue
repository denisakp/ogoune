<script setup lang="ts">
import UUptimeCalendar from '@/components/ui/UUptimeCalendar.vue'
import type { PublicUptimeDay } from '@/types'

const props = defineProps<{
  year: number
  month: number
  days: PublicUptimeDay[]
}>()

// Map service-shape days → UUptimeCalendar entries.
function toEntries() {
  return props.days
    .filter((d) => d.day.startsWith(`${props.year}-${String(props.month).padStart(2, '0')}`))
    .map((d) => ({ day: d.day, ratio: d.samples === 0 ? null : d.uptime_ratio }))
}
</script>

<template>
  <UUptimeCalendar :year="year" :month="month" :entries="toEntries()" hide-header />
</template>
