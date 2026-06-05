<script setup lang="ts">
/**
 * Spec 060 — monthly uptime calendar.
 * - 7-column grid with leading blank cells for the month's first weekday
 *   (Monday-first ISO week per FR-008).
 * - Future-month cells render as neutral "unknown" until backfilled.
 * - Thresholds shared with UUptimeBar (Atlassian-style).
 */
import { computed } from 'vue'

export interface UptimeCalendarEntry {
  day: string
  ratio: number | null
}

const props = withDefaults(
  defineProps<{
    year: number
    month: number // 1-12
    entries: UptimeCalendarEntry[]
    hideHeader?: boolean
  }>(),
  { hideHeader: false },
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

const MONTHS = [
  'January', 'February', 'March', 'April', 'May', 'June',
  'July', 'August', 'September', 'October', 'November', 'December',
]

const monthLabel = computed(() => `${MONTHS[props.month - 1]} ${props.year}`)

const daysInMonth = computed(() => new Date(props.year, props.month, 0).getDate())

// Monday = 0 ... Sunday = 6 — keeps ISO-style weeks for European users.
const leadingBlanks = computed(() => {
  const firstDow = new Date(props.year, props.month - 1, 1).getDay() // 0=Sun..6=Sat
  return (firstDow + 6) % 7
})

const entriesByDay = computed(() => {
  const map = new Map<string, number | null>()
  for (const e of props.entries) map.set(e.day, e.ratio)
  return map
})

interface Cell {
  key: string
  blank: boolean
  dayNum?: number
  band?: Band
  colorClass?: string
  tooltip?: string
}

const cells = computed<Cell[]>(() => {
  const out: Cell[] = []
  for (let i = 0; i < leadingBlanks.value; i++) {
    out.push({ key: `blank-${i}`, blank: true })
  }
  for (let d = 1; d <= daysInMonth.value; d++) {
    const isoDay = `${props.year}-${String(props.month).padStart(2, '0')}-${String(d).padStart(2, '0')}`
    const ratio = entriesByDay.value.has(isoDay) ? (entriesByDay.value.get(isoDay) ?? null) : null
    const band = bandFor(ratio)
    const pct = ratio === null ? null : Math.round(ratio * 10000) / 100
    out.push({
      key: isoDay,
      blank: false,
      dayNum: d,
      band,
      colorClass: bandClass[band],
      tooltip: pct === null ? `${isoDay} — no data` : `${isoDay} — ${pct}%`,
    })
  }
  return out
})

const monthlyRatio = computed(() => {
  const known = props.entries.filter((e) => e.ratio !== null && !Number.isNaN(e.ratio))
  if (known.length === 0) return null
  const sum = known.reduce((acc, e) => acc + (e.ratio as number), 0)
  return sum / known.length
})

const monthlyPctLabel = computed(() =>
  monthlyRatio.value === null ? '—' : `${(monthlyRatio.value * 100).toFixed(2)}%`,
)
</script>

<template>
  <div class="space-y-2">
    <div v-if="!hideHeader" class="flex items-baseline justify-between">
      <h4 class="text-sm font-medium">{{ monthLabel }}</h4>
      <span class="text-xs text-muted font-mono">{{ monthlyPctLabel }}</span>
    </div>
    <div class="grid grid-cols-7 gap-1.5">
      <span
        v-for="cell in cells"
        :key="cell.key"
        :class="['h-6 w-full rounded-md', cell.blank ? 'bg-transparent' : cell.colorClass]"
        :data-blank="cell.blank ? '1' : undefined"
        :data-band="cell.band"
        :data-day-num="cell.dayNum"
        :title="cell.tooltip"
        :aria-label="cell.blank ? 'empty' : `Day ${cell.dayNum}: ${cell.band}`"
      />
    </div>
  </div>
</template>
