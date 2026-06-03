<script setup lang="ts">
/**
 * Monthly uptime calendar (Atlassian-style) — STRICT isolation (spec 055 Q2).
 */
import { computed } from 'vue'

type DayState = 'up' | 'warning' | 'down' | 'nodata'

interface Props {
  month: number // 1-12
  year: number
  days: DayState[]
  uptimePct: number
}

const props = defineProps<Props>()

const MONTHS = [
  'January', 'February', 'March', 'April', 'May', 'June',
  'July', 'August', 'September', 'October', 'November', 'December',
]

const monthLabel = computed(() => `${MONTHS[props.month - 1]} ${props.year}`)

const dayClass: Record<DayState, string> = {
  up: 'bg-emerald-500',
  warning: 'bg-amber-500',
  down: 'bg-red-500',
  nodata: 'bg-slate-200 dark:bg-slate-700',
}
</script>

<template>
  <div class="space-y-2">
    <div class="flex items-baseline justify-between">
      <h4 class="text-sm font-medium">{{ monthLabel }}</h4>
      <span class="text-xs text-muted font-mono">{{ uptimePct.toFixed(2) }}%</span>
    </div>
    <div class="grid grid-cols-7 gap-1">
      <span
        v-for="(d, i) in days"
        :key="i"
        :class="['size-6 rounded-sm', dayClass[d]]"
        :data-day="d"
        :aria-label="`Day ${i + 1}: ${d}`"
      />
    </div>
  </div>
</template>
