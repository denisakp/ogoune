<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  startYear: number
  startMonth: number
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

const lastMonth = computed(() => shiftMonth(props.startYear, props.startMonth, 2))
const label = computed(() => {
  const start = `${MONTHS_SHORT[props.startMonth - 1]} ${props.startYear}`
  const end = `${MONTHS_SHORT[lastMonth.value.month - 1]} ${lastMonth.value.year}`
  return `${start} to ${end}`
})
</script>

<template>
  <div
    class="inline-flex items-center gap-1 rounded-md border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-900 px-1 py-1 text-sm"
    data-testid="calendar-range-navigator"
  >
    <button
      type="button"
      class="size-7 rounded hover:bg-gray-100 dark:hover:bg-gray-800 inline-flex items-center justify-center text-gray-500"
      aria-label="Previous 3 months"
      data-testid="nav-prev"
      @click="emit('shift', -3)"
    >
      ‹
    </button>
    <span class="px-2 font-medium text-gray-900 dark:text-gray-100" data-testid="nav-label">{{ label }}</span>
    <button
      type="button"
      class="size-7 rounded hover:bg-gray-100 dark:hover:bg-gray-800 inline-flex items-center justify-center text-gray-500"
      aria-label="Next 3 months"
      data-testid="nav-next"
      @click="emit('shift', 3)"
    >
      ›
    </button>
  </div>
</template>
