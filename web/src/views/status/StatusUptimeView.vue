<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useStatusPublic } from '@/composables/useStatusPublic'
import PublicHeader from '@/components/status/PublicHeader.vue'
import PublicHistoryUptimeTabs from '@/components/status/PublicHistoryUptimeTabs.vue'
import CalendarMonth from '@/components/status/CalendarMonth.vue'
import CalendarRangeNavigator from '@/components/status/CalendarRangeNavigator.vue'
import PublicPageFooter from '@/components/status/PublicPageFooter.vue'

const { uptime, summary, error, loadSummary, loadUptime } = useStatusPublic()

const MONTHS = [
  'January',
  'February',
  'March',
  'April',
  'May',
  'June',
  'July',
  'August',
  'September',
  'October',
  'November',
  'December',
]

const componentID = ref('')
const today = new Date()
const startYear = ref(today.getUTCFullYear())
const startMonth = ref(today.getUTCMonth() - 1)
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

const rangeBounds = computed(() => {
  const first = visibleMonths.value[0]
  const last = visibleMonths.value[2]
  if (!first || !last) return { from: '', to: '' }
  const lastDay = new Date(Date.UTC(last.year, last.month, 0)).getUTCDate()
  return {
    from: `${first.year}-${String(first.month).padStart(2, '0')}-01`,
    to: `${last.year}-${String(last.month).padStart(2, '0')}-${String(lastDay).padStart(2, '0')}`,
  }
})

async function refresh() {
  const { from, to } = rangeBounds.value
  if (!from || !to) return
  await loadUptime({ from, to, component_id: componentID.value || undefined })
}

onMounted(async () => {
  await loadSummary()
  clampToBounds()
  await refresh()
})

function clampToBounds() {
  // Clamp the visible window to [earliestKey, currentMonthKey].
  // Start with the latest 3-month window ending at the current month.
  const now = new Date()
  const curY = now.getUTCFullYear()
  const curM = now.getUTCMonth() + 1
  // End of visible window = startMonth + 2. We want endMonth ≤ currentMonth.
  const endIdx = startMonth.value - 1 + 2
  const endY = startYear.value + Math.floor(endIdx / 12)
  const endM = (((endIdx % 12) + 12) % 12) + 1
  const curKey = `${curY}-${String(curM).padStart(2, '0')}`
  const endKey = `${endY}-${String(endM).padStart(2, '0')}`
  if (endKey > curKey) {
    // Force end == current month.
    const startIdx = curM - 1 - 2
    startYear.value = curY + Math.floor(startIdx / 12)
    let m = startIdx % 12
    if (m < 0) m += 12
    startMonth.value = m + 1
  }
}

watch([startYear, startMonth, componentID], () => {
  refresh()
})

function shift(delta: number) {
  const idx = startMonth.value - 1 + delta
  startYear.value += Math.floor(idx / 12)
  let m = idx % 12
  if (m < 0) m += 12
  startMonth.value = m + 1
}

const componentOptions = computed(() => summary.value?.components ?? [])
const daysInWindow = computed(() => uptime.value?.days ?? [])

function monthUptimePct(year: number, month: number): string {
  const prefix = `${year}-${String(month).padStart(2, '0')}`
  const subset = daysInWindow.value.filter((d) => d.day.startsWith(prefix) && d.samples > 0)
  if (subset.length === 0) return '—'
  const mean = subset.reduce((acc, d) => acc + d.uptime_ratio, 0) / subset.length
  return `${(mean * 100).toFixed(2)}%`
}

const branding = computed(() => summary.value?.branding ?? null)
const brandName = computed(() => branding.value?.name ?? 'Status Page')

const earliestYearMonth = computed(() => {
  const iso = summary.value?.uptime_window?.earliest_day
  if (!iso) {
    // No data yet — pin the lower bound to the current month so user can't roam.
    const now = new Date()
    return `${now.getUTCFullYear()}-${String(now.getUTCMonth() + 1).padStart(2, '0')}`
  }
  return iso.slice(0, 7)
})

const latestYearMonth = computed(() => {
  // Always cap to the current month — uptime of future months doesn't exist.
  const now = new Date()
  return `${now.getUTCFullYear()}-${String(now.getUTCMonth() + 1).padStart(2, '0')}`
})
</script>

<template>
  <div class="min-h-screen bg-white">
    <PublicHeader :branding="branding" />
    <PublicHistoryUptimeTabs />

    <main class="max-w-5xl mx-auto px-6 py-6 space-y-6" data-testid="status-uptime-view">
      <header class="flex flex-wrap items-center gap-3">
        <div
          class="relative inline-flex items-center gap-2 rounded-md border border-gray-200 bg-white px-3 py-1.5 text-sm"
        >
          <svg
            class="size-4 text-gray-400"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <path
              d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"
            />
          </svg>
          <select
            v-model="componentID"
            class="bg-transparent outline-none pr-6"
            data-testid="filter-component"
          >
            <option value="">All Components</option>
            <option v-for="c in componentOptions" :key="c.id" :value="c.id">{{ c.name }}</option>
          </select>
        </div>
        <CalendarRangeNavigator
          class="ml-auto"
          :start-year="startYear"
          :start-month="startMonth"
          :min-year-month="earliestYearMonth"
          :max-year-month="latestYearMonth"
          @shift="shift"
        />
      </header>

      <div
        v-if="error && !uptime"
        class="rounded-xl border border-red-200 bg-red-50 p-6 text-red-700"
      >
        <p class="font-semibold mb-1">Could not load uptime</p>
        <p class="text-sm opacity-80">{{ error.message }}</p>
      </div>

      <section class="grid gap-8 sm:grid-cols-2 lg:grid-cols-3" data-testid="calendars">
        <div
          v-for="m in visibleMonths"
          :key="`${m.year}-${m.month}`"
          :data-month="`${m.year}-${m.month}`"
          class="space-y-2"
        >
          <div class="flex items-baseline justify-between">
            <h3 class="text-sm font-semibold text-gray-900">
              {{ MONTHS[m.month - 1] }} {{ m.year }}
            </h3>
            <span class="text-xs font-mono text-gray-500" :data-month-pct="`${m.year}-${m.month}`">
              {{ monthUptimePct(m.year, m.month) }}
            </span>
          </div>
          <CalendarMonth :year="m.year" :month="m.month" :days="daysInWindow" />
        </div>
      </section>
    </main>

    <PublicPageFooter :brand-name="brandName" back-href="#/" back-label="Current Status">
      <template #right>
        <div class="flex items-center gap-3 text-xs">
          <span class="inline-flex items-center gap-1"
            ><span class="size-2 rounded-sm bg-emerald-500" /> Operational</span
          >
          <span class="inline-flex items-center gap-1"
            ><span class="size-2 rounded-sm bg-yellow-400" /> Minor</span
          >
          <span class="inline-flex items-center gap-1"
            ><span class="size-2 rounded-sm bg-orange-500" /> Major</span
          >
          <span class="inline-flex items-center gap-1"
            ><span class="size-2 rounded-sm bg-red-500" /> Outage</span
          >
        </div>
      </template>
    </PublicPageFooter>
  </div>
</template>
