<script setup lang="ts">
import { computed } from 'vue'

/**
 * Spec 060 — 90-day uptime ribbon.
 * Thresholds per FR-004 (Atlassian-style):
 *   1.0    → operational (green)
 *   ≥ 0.99 → minor       (yellow)
 *   ≥ 0.95 → major       (orange)
 *   < 0.95 → outage      (red)
 *   null / NaN → unknown (neutral)
 */

export interface UptimeBarEntry {
  day: string
  ratio: number | null
}

const props = withDefaults(
  defineProps<{
    entries: UptimeBarEntry[]
    compact?: boolean
  }>(),
  { compact: false },
)

type Band = 'operational' | 'minor' | 'major' | 'outage' | 'unknown'

function bandFor(ratio: number | null): Band {
  if (ratio === null || Number.isNaN(ratio)) return 'unknown'
  if (ratio >= 1) return 'operational'
  if (ratio >= 0.99) return 'minor'
  if (ratio >= 0.95) return 'major'
  return 'outage'
}

const bandClass: Record<Band, string> = {
  operational: 'bg-emerald-500',
  minor: 'bg-yellow-400',
  major: 'bg-orange-500',
  outage: 'bg-red-500',
  unknown: 'bg-slate-200 dark:bg-slate-700',
}

const cells = computed(() =>
  props.entries.map((e) => {
    const band = bandFor(e.ratio)
    const pct = e.ratio === null ? null : Math.round(e.ratio * 10000) / 100
    return {
      day: e.day,
      band,
      colorClass: bandClass[band],
      tooltip: pct === null ? `${e.day} — no data` : `${e.day} — ${pct}%`,
    }
  }),
)
</script>

<template>
  <div
    :class="['flex w-full items-stretch', compact ? 'h-1.5 gap-px' : 'h-3 gap-0.5']"
    role="img"
    :aria-label="`Uptime over ${entries.length} days`"
  >
    <span
      v-for="cell in cells"
      :key="cell.day"
      :class="['flex-1 rounded-sm', cell.colorClass]"
      :data-band="cell.band"
      :title="cell.tooltip"
    />
  </div>
</template>
