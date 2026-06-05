<script setup lang="ts">
/**
 * Linear uptime bar — STRICT isolation (spec 055 Q2).
 */
type DayState = 'up' | 'warning' | 'down' | 'nodata'

interface Props {
  days: DayState[]
  compact?: boolean
}

withDefaults(defineProps<Props>(), {
  compact: false,
})

const dayClass: Record<DayState, string> = {
  up: 'bg-emerald-500',
  warning: 'bg-amber-500',
  down: 'bg-red-500',
  nodata: 'bg-slate-200 dark:bg-slate-700',
}
</script>

<template>
  <div
    :class="['flex w-full items-stretch', compact ? 'h-1.5 gap-px' : 'h-3 gap-0.5']"
    role="img"
    :aria-label="`Uptime over ${days.length} days`"
  >
    <span
      v-for="(d, i) in days"
      :key="i"
      :class="['flex-1 rounded-sm', dayClass[d]]"
      :data-day="d"
    />
  </div>
</template>
