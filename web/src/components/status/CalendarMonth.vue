<script setup lang="ts">
import { computed, ref } from 'vue'
import CalendarDayTooltip from './CalendarDayTooltip.vue'
import type { PublicUptimeDay } from '@/types'

const props = defineProps<{
  year: number
  month: number // 1-12
  days: PublicUptimeDay[]
}>()

type Band = 'operational' | 'minor' | 'major' | 'outage' | 'unknown'

function bandFor(d: PublicUptimeDay | null): Band {
  if (!d || d.samples === 0) return 'unknown'
  const r = d.uptime_ratio
  if (r >= 1) return 'operational'
  if (r >= 0.99) return 'minor'
  if (r >= 0.95) return 'major'
  return 'outage'
}

const BAND_CLASS: Record<Band, string> = {
  operational: 'bg-emerald-500',
  minor: 'bg-yellow-400',
  major: 'bg-orange-500',
  outage: 'bg-red-500',
  unknown: 'bg-slate-200',
}

const daysInMonth = computed(() => new Date(props.year, props.month, 0).getDate())

// Monday-first weeks per FR-008.
const leadingBlanks = computed(() => {
  const firstDow = new Date(props.year, props.month - 1, 1).getDay()
  return (firstDow + 6) % 7
})

const byDay = computed(() => {
  const map = new Map<string, PublicUptimeDay>()
  for (const d of props.days) map.set(d.day, d)
  return map
})

interface Cell {
  key: string
  blank: boolean
  dayNum?: number
  band?: Band
  data?: PublicUptimeDay | null
}

const cells = computed<Cell[]>(() => {
  const out: Cell[] = []
  for (let i = 0; i < leadingBlanks.value; i++) {
    out.push({ key: `blank-${i}`, blank: true })
  }
  for (let d = 1; d <= daysInMonth.value; d++) {
    const iso = `${props.year}-${String(props.month).padStart(2, '0')}-${String(d).padStart(2, '0')}`
    const data = byDay.value.get(iso) ?? null
    const band = bandFor(data)
    out.push({
      key: iso,
      blank: false,
      dayNum: d,
      band,
      data,
    })
  }
  return out
})

const hoveredKey = ref<string | null>(null)
const hoveredCellEl = ref<HTMLElement | null>(null)

function onEnter(e: MouseEvent, cell: Cell) {
  if (cell.blank || !cell.data) return
  hoveredKey.value = cell.key
  hoveredCellEl.value = e.currentTarget as HTMLElement
}

function onLeave() {
  hoveredKey.value = null
  hoveredCellEl.value = null
}

const hoveredDay = computed<PublicUptimeDay | null>(() => {
  if (!hoveredKey.value) return null
  const cell = cells.value.find((c) => c.key === hoveredKey.value)
  return cell?.data ?? null
})

// Position the tooltip below the hovered cell, centered horizontally.
const tooltipPosition = computed(() => {
  if (!hoveredCellEl.value) return { top: '0px', left: '0px', display: 'none' }
  const rect = hoveredCellEl.value.getBoundingClientRect()
  const tooltipWidth = 288
  let left = rect.left + rect.width / 2 - tooltipWidth / 2
  left = Math.max(8, Math.min(left, window.innerWidth - tooltipWidth - 8))
  return {
    top: `${rect.bottom + window.scrollY + 8}px`,
    left: `${left + window.scrollX}px`,
  }
})
</script>

<template>
  <div class="space-y-2">
    <div class="grid grid-cols-7 gap-1.5">
      <span
        v-for="cell in cells"
        :key="cell.key"
        :class="[
          'h-6 w-full rounded-md',
          cell.blank ? 'bg-transparent' : BAND_CLASS[cell.band as Band],
          !cell.blank && cell.data ? 'cursor-pointer hover:ring-2 hover:ring-offset-1 hover:ring-gray-300' : '',
        ]"
        :data-blank="cell.blank ? '1' : undefined"
        :data-band="cell.band"
        :data-day-num="cell.dayNum"
        :data-day-iso="cell.key.startsWith('blank-') ? undefined : cell.key"
        :aria-label="cell.blank ? 'empty' : `Day ${cell.dayNum}: ${cell.band}`"
        @mouseenter="onEnter($event, cell)"
        @mouseleave="onLeave"
      />
    </div>

    <Teleport to="body">
      <div
        v-if="hoveredDay"
        class="fixed z-50 pointer-events-auto"
        :style="tooltipPosition"
        @mouseenter="hoveredKey = hoveredKey"
        @mouseleave="onLeave"
      >
        <CalendarDayTooltip :day="hoveredDay" />
      </div>
    </Teleport>
  </div>
</template>
