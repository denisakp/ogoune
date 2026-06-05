<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useStatusPublic } from '@/composables/useStatusPublic'
import PublicQuickNav from '@/components/status/PublicQuickNav.vue'
import CalendarMonth from '@/components/status/CalendarMonth.vue'
import CalendarRangeNavigator from '@/components/status/CalendarRangeNavigator.vue'
import PublicPageFooter from '@/components/status/PublicPageFooter.vue'

const { uptime, summary, error, loadSummary, loadUptime } = useStatusPublic()

const componentID = ref('')
// Default: 3-month window ending on the current month.
const today = new Date()
const startYear = ref(today.getUTCFullYear())
const startMonth = ref(today.getUTCMonth() - 1) // start = month - 2 (1-indexed below)

if (startMonth.value <= 0) {
  startMonth.value += 12
  startYear.value -= 1
}

const visibleMonths = computed(() => {
  const out: { year: number; month: number }[] = []
  for (let i = 0; i < 3; i++) {
    const idx = startMonth.value - 1 + i
    const y = startYear.value + Math.floor(idx / 12)
    let m = idx % 12
    if (m < 0) m += 12
    out.push({ year: y, month: m + 1 })
  }
  return out
})

function fmtDate(y: number, m: number, d: number): string {
  return `${y}-${String(m).padStart(2, '0')}-${String(d).padStart(2, '0')}`
}

const rangeBounds = computed(() => {
  const first = visibleMonths.value[0]
  const last = visibleMonths.value[2]
  if (!first || !last) return { from: '', to: '' }
  const lastDay = new Date(Date.UTC(last.year, last.month, 0)).getUTCDate()
  return {
    from: fmtDate(first.year, first.month, 1),
    to: fmtDate(last.year, last.month, lastDay),
  }
})

async function refresh() {
  const { from, to } = rangeBounds.value
  if (!from || !to) return
  await loadUptime({ from, to, component_id: componentID.value || undefined })
}

onMounted(async () => {
  await loadSummary()
  await refresh()
})

watch([startYear, startMonth, componentID], () => {
  refresh()
})

function shift(delta: number) {
  const idx = (startMonth.value - 1) + delta
  startYear.value += Math.floor(idx / 12)
  let m = idx % 12
  if (m < 0) m += 12
  startMonth.value = m + 1
}

const componentOptions = computed(() => summary.value?.components ?? [])
const daysInWindow = computed(() => uptime.value?.days ?? [])
</script>

<template>
  <main class="max-w-4xl mx-auto px-4 py-8 space-y-6" data-testid="status-uptime-view">
    <div class="flex justify-end">
      <PublicQuickNav />
    </div>

    <header class="space-y-3">
      <h1 class="text-2xl font-semibold">Uptime</h1>
      <div class="flex flex-wrap items-end gap-3">
        <label class="text-sm">
          <span class="block text-xs text-gray-500 mb-1">Component</span>
          <select
            v-model="componentID"
            class="rounded-md border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 px-2 py-1 text-sm"
            data-testid="filter-component"
          >
            <option value="">All</option>
            <option v-for="c in componentOptions" :key="c.id" :value="c.id">{{ c.name }}</option>
          </select>
        </label>
        <CalendarRangeNavigator
          class="ml-auto"
          :start-year="startYear"
          :start-month="startMonth"
          @shift="shift"
        />
      </div>
    </header>

    <div
      v-if="error && !uptime"
      class="rounded-xl border border-red-300 bg-red-50 dark:bg-red-950/40 p-6 text-red-700 dark:text-red-300"
    >
      <p class="font-semibold mb-1">Could not load uptime</p>
      <p class="text-sm opacity-80">{{ error.message }}</p>
    </div>

    <section
      class="grid gap-6 sm:grid-cols-2 lg:grid-cols-3"
      data-testid="calendars"
    >
      <div
        v-for="m in visibleMonths"
        :key="`${m.year}-${m.month}`"
        class="rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 p-4"
        :data-month="`${m.year}-${m.month}`"
      >
        <CalendarMonth :year="m.year" :month="m.month" :days="daysInWindow" />
      </div>
    </section>

    <footer
      class="text-xs text-gray-500 flex flex-wrap items-center gap-4"
      data-testid="legend"
    >
      <span class="flex items-center gap-1"><span class="size-3 rounded-sm bg-emerald-500" /> Operational</span>
      <span class="flex items-center gap-1"><span class="size-3 rounded-sm bg-yellow-400" /> Minor</span>
      <span class="flex items-center gap-1"><span class="size-3 rounded-sm bg-orange-500" /> Major</span>
      <span class="flex items-center gap-1"><span class="size-3 rounded-sm bg-red-500" /> Outage</span>
      <span class="flex items-center gap-1"><span class="size-3 rounded-sm bg-slate-200 dark:bg-slate-700" /> No data</span>
    </footer>

    <PublicPageFooter />
  </main>
</template>
