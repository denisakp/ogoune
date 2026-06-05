<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  startYear: number
  startMonth: number
  // Bounds expressed as YYYY-MM (inclusive). When omitted, no clamping.
  minYearMonth?: string
  maxYearMonth?: string
}>()

const emit = defineEmits<{
  (e: 'shift', delta: number): void
}>()

const MONTHS_SHORT = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun',
  'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']

function shiftMonth(year: number, month: number, delta: number) {
  const idx = (month - 1) + delta
  const y = year + Math.floor(idx / 12)
  let m = idx % 12
  if (m < 0) m += 12
  return { year: y, month: m + 1 }
}

function key(year: number, month: number) {
  return `${year}-${String(month).padStart(2, '0')}`
}

const startKey = computed(() => key(props.startYear, props.startMonth))
const lastMonth = computed(() => shiftMonth(props.startYear, props.startMonth, 2))
const endKey = computed(() => key(lastMonth.value.year, lastMonth.value.month))

const label = computed(() => {
  const start = `${MONTHS_SHORT[props.startMonth - 1]} ${props.startYear}`
  const end = `${MONTHS_SHORT[lastMonth.value.month - 1]} ${lastMonth.value.year}`
  return `${start} to ${end}`
})

// Disable Prev when shifting back by 3 would land before minYearMonth.
const canGoPrev = computed(() => {
  if (!props.minYearMonth) return true
  const prev = shiftMonth(props.startYear, props.startMonth, -3)
  // We allow Prev as long as the new END (= start + 2) >= minYearMonth.
  // Practically: don't let the user navigate past the earliest known month.
  return key(prev.year, prev.month) >= props.minYearMonth
})

// Disable Next when shifting forward would land after maxYearMonth.
const canGoNext = computed(() => {
  if (!props.maxYearMonth) return true
  // Next is forbidden if endKey already equals or exceeds max.
  return endKey.value < props.maxYearMonth
})

function go(delta: number) {
  if (delta < 0 && !canGoPrev.value) return
  if (delta > 0 && !canGoNext.value) return
  emit('shift', delta)
}
</script>

<template>
  <div
    class="inline-flex items-center gap-1 rounded-md border border-gray-200 bg-white px-1 py-1 text-sm"
    data-testid="calendar-range-navigator"
    :data-start="startKey"
    :data-end="endKey"
  >
    <button
      type="button"
      class="size-7 rounded inline-flex items-center justify-center text-gray-500 hover:bg-gray-100 disabled:opacity-30 disabled:cursor-not-allowed disabled:hover:bg-transparent"
      aria-label="Previous 3 months"
      data-testid="nav-prev"
      :disabled="!canGoPrev"
      @click="go(-3)"
    >
      ‹
    </button>
    <span class="px-2 font-medium text-gray-900" data-testid="nav-label">{{ label }}</span>
    <button
      type="button"
      class="size-7 rounded inline-flex items-center justify-center text-gray-500 hover:bg-gray-100 disabled:opacity-30 disabled:cursor-not-allowed disabled:hover:bg-transparent"
      aria-label="Next 3 months"
      data-testid="nav-next"
      :disabled="!canGoNext"
      @click="go(3)"
    >
      ›
    </button>
  </div>
</template>
