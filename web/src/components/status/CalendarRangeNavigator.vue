<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  // First month in the current 3-month window.
  startYear: number
  startMonth: number // 1-12
}>()

const emit = defineEmits<{
  (e: 'shift', delta: number): void
}>()

const MONTHS = [
  'January', 'February', 'March', 'April', 'May', 'June',
  'July', 'August', 'September', 'October', 'November', 'December',
]

function shiftMonth(year: number, month: number, delta: number): { year: number; month: number } {
  // month is 1..12; convert to 0..11, add delta, normalize back.
  const idx = (month - 1) + delta
  const y = year + Math.floor(idx / 12)
  let m = idx % 12
  if (m < 0) m += 12
  return { year: y, month: m + 1 }
}

const lastMonth = computed(() => shiftMonth(props.startYear, props.startMonth, 2))

const label = computed(() => {
  const startLabel = `${MONTHS[props.startMonth - 1]} ${props.startYear}`
  const endLabel = `${MONTHS[lastMonth.value.month - 1]} ${lastMonth.value.year}`
  return `${startLabel} — ${endLabel}`
})
</script>

<template>
  <div
    class="flex items-center justify-center gap-3 text-sm"
    data-testid="calendar-range-navigator"
  >
    <button
      type="button"
      class="px-2 py-1 rounded-md hover:bg-gray-100 dark:hover:bg-gray-700"
      aria-label="Previous 3 months"
      data-testid="nav-prev"
      @click="emit('shift', -3)"
    >
      ‹
    </button>
    <span class="font-medium" data-testid="nav-label">{{ label }}</span>
    <button
      type="button"
      class="px-2 py-1 rounded-md hover:bg-gray-100 dark:hover:bg-gray-700"
      aria-label="Next 3 months"
      data-testid="nav-next"
      @click="emit('shift', 3)"
    >
      ›
    </button>
  </div>
</template>
