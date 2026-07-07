<script setup lang="ts">
import { computed, inject, onMounted, ref, type ComputedRef, type Ref } from 'vue'
import { useRouter } from 'vue-router'

import { useResourceStore } from '@/stores/resourceStore'
import { fetchPublicStatusSummary } from '@/services/statusPublicService'
import type { OverviewRange } from '@/composables/useOverviewMetrics'
import type { PublicStatusSummary } from '@/types'

const resourceStore = useResourceStore()
const router = useRouter()

const rangeInj = inject<{ range: Ref<OverviewRange> } | null>('overview.range', null)
const range = computed<OverviewRange>(() => rangeInj?.range.value ?? '24h')

interface MetricsShape {
  uptimePct: ComputedRef<number | null>
}
const metrics = inject<MetricsShape | null>('overview.metrics', null)

const RANGE_LABEL: Record<OverviewRange, string> = {
  '1h': '1-hour uptime',
  '6h': '6-hour uptime',
  '24h': '24-hour uptime',
  '7d': '7-day uptime',
  '30d': '30-day uptime',
}

const hasResources = computed(() => resourceStore.resources.length > 0)
const downCount = computed(() => resourceStore.resources.filter((r) => r.status === 'down').length)
const hasIncidents = computed(() => downCount.value > 0)

const summary = ref<PublicStatusSummary | null>(null)

onMounted(async () => {
  try {
    summary.value = await fetchPublicStatusSummary()
  } catch {
    summary.value = null
  }
})

// Aggregate every resource's 90-day ribbon → keep the last 30 entries,
// average each day across resources. Ribbons return ratio:null for days
// without data.
const dailyRatios = computed<(number | null)[]>(() => {
  const resources = [
    ...(summary.value?.components.flatMap((c) => c.resources) ?? []),
    ...(summary.value?.standalone_resources ?? []),
  ]
  if (resources.length === 0) return []
  const ribbons = resources.map((r) => r.uptime_ribbon ?? [])
  const len = ribbons.reduce((m, r) => Math.max(m, r.length), 0)
  if (len === 0) return []
  const start = Math.max(0, len - 30)
  const out: (number | null)[] = []
  for (let i = start; i < len; i++) {
    let sum = 0
    let count = 0
    for (const r of ribbons) {
      const entry = r[i]
      if (entry && entry.ratio !== null) {
        sum += entry.ratio
        count++
      }
    }
    out.push(count === 0 ? null : sum / count)
  }
  return out
})

const knownRatios = computed(() => dailyRatios.value.filter((v): v is number => v !== null))

const uptimePct = computed<number | null>(() => {
  // No resources yet → no uptime to report. Returning null lets the template
  // render a neutral "no data" state instead of a misleading 100 %.
  if (!hasResources.value) return null
  // Prefer the windowed value from useOverviewMetrics (activity-based, real
  // time-aware). Fall back to the 90-day ribbon mean when no activities
  // landed in the window yet.
  const windowed = metrics?.uptimePct.value
  if (typeof windowed === 'number') return Math.round(windowed * 100) / 100
  if (knownRatios.value.length === 0) {
    const up = resourceStore.resources.filter((r) => r.status === 'up').length
    return Math.round((up / resourceStore.resources.length) * 1000) / 10
  }
  const mean = knownRatios.value.reduce((a, b) => a + b, 0) / knownRatios.value.length
  return Math.round(mean * 1000) / 10
})

const hasUptime = computed(() => uptimePct.value !== null)
const uptimeWhole = computed(() => Math.floor(uptimePct.value ?? 0))
const uptimeDecimal = computed(() => {
  const v = uptimePct.value ?? 0
  const dec = Math.round((v - Math.floor(v)) * 100)
  return dec.toString().padStart(2, '0')
})

interface SparkBar {
  h: number
  band: 'operational' | 'minor' | 'major' | 'outage' | 'unknown'
}

function bandFor(ratio: number | null): SparkBar['band'] {
  if (ratio === null) return 'unknown'
  if (ratio >= 1) return 'operational'
  if (ratio >= 0.99) return 'minor'
  if (ratio >= 0.95) return 'major'
  return 'outage'
}

const sparkBars = computed<SparkBar[]>(() => {
  const ratios = dailyRatios.value
  if (ratios.length === 0) {
    return Array.from({ length: 30 }, () => ({ h: 30, band: 'unknown' as const }))
  }
  const padded: (number | null)[] = [
    ...Array(Math.max(0, 30 - ratios.length)).fill(null),
    ...ratios.slice(-30),
  ]
  return padded.map((r) => {
    const band = bandFor(r)
    let h = 30
    if (r !== null) {
      const v = Math.max(0.5, r)
      h = Math.round(30 + (v - 0.5) * 140)
      if (h > 100) h = 100
    }
    return { h, band }
  })
})

const BAND_COLORS: Record<SparkBar['band'], string> = {
  operational: '#4F46E5',
  minor: '#F59E0B',
  major: '#F97316',
  outage: '#EF4444',
  unknown: '#E2E8F0',
}

function openStatusPage() {
  window.open('/status.html', '_blank', 'noopener')
}
</script>

<template>
  <div class="bg-default rounded-lg border border-default p-5 flex gap-4 items-start">
    <div class="flex-1 min-w-0 flex flex-col gap-3.5">
      <div class="flex items-center gap-2.5 flex-wrap min-w-0">
        <div
          v-if="!hasResources"
          class="inline-flex items-center gap-2 px-2.5 py-0.5 rounded-full border border-default bg-muted"
        >
          <span class="size-1.5 rounded-full bg-dimmed" />
          <span class="text-[13px] font-semibold text-muted"> No resources yet </span>
        </div>
        <div
          v-else-if="hasIncidents"
          class="inline-flex items-center gap-2 px-2.5 py-0.5 rounded-full border"
          style="background-color: #fffbeb; border-color: #fcd34d"
        >
          <span class="size-1.5 rounded-full" style="background-color: #f59e0b" />
          <span class="text-[13px] font-semibold" style="color: #92400e">
            {{ downCount }} active incident{{ downCount > 1 ? 's' : '' }}
          </span>
        </div>
        <div
          v-else
          class="inline-flex items-center gap-2 px-2.5 py-0.5 rounded-full border"
          style="background-color: #ecfdf5; border-color: #6ee7b7"
        >
          <span class="size-1.5 rounded-full" style="background-color: #10b981" />
          <span class="text-[13px] font-semibold" style="color: #047857">
            All systems operational
          </span>
        </div>
        <template v-if="!hasResources">
          <span class="text-dimmed text-xs">·</span>
          <span class="text-[13px] text-muted">Add a resource to start monitoring</span>
        </template>
        <template v-else-if="hasIncidents">
          <span class="text-dimmed text-xs">·</span>
          <span class="text-[13px] text-muted">
            Detected just now, {{ downCount }} resource{{ downCount > 1 ? 's' : '' }} down
          </span>
        </template>
      </div>

      <div class="flex items-end gap-5">
        <div class="shrink-0">
          <div class="flex items-end gap-1">
            <template v-if="hasUptime">
              <span class="text-[44px] font-bold text-highlighted leading-none tracking-tight">
                {{ uptimeWhole }}.{{ uptimeDecimal }}
              </span>
              <span class="text-xl font-semibold text-muted leading-none pb-1">%</span>
            </template>
            <span
              v-else
              class="text-[44px] font-bold text-dimmed leading-none tracking-tight"
              aria-label="No data yet"
            >
              —
            </span>
          </div>
          <div class="text-[10px] font-medium text-dimmed uppercase tracking-wider mt-1">
            {{ hasUptime ? RANGE_LABEL[range] : 'No data yet' }}
          </div>
        </div>

        <div class="flex-1 min-w-0 flex justify-center">
          <div class="inline-flex flex-col">
            <div class="flex items-end gap-0.75 h-10">
              <span
                v-for="(b, i) in sparkBars"
                :key="i"
                class="w-2 rounded-[1px]"
                :style="{
                  height: `${b.h}%`,
                  backgroundColor: BAND_COLORS[b.band],
                  opacity: b.band === 'unknown' ? 0.4 : 1,
                }"
              />
            </div>
            <div class="flex justify-between mt-1 text-[10px] text-dimmed">
              <span>30 days ago</span>
              <span>now</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="flex flex-col gap-1.5 shrink-0 w-42.5">
      <button
        type="button"
        class="inline-flex items-center justify-between gap-1.5 px-2.5 py-1.5 rounded-md border border-default text-[12px] text-default hover:bg-muted"
        @click="router.push('/incidents')"
      >
        <span class="inline-flex items-center gap-1.5">
          <UIcon name="i-lucide-alert-circle" class="size-3.5" />
          <span v-if="hasIncidents">View {{ downCount }} active</span>
          <span v-else>View incidents</span>
        </span>
        <UIcon name="i-lucide-arrow-right" class="size-3" />
      </button>
      <button
        type="button"
        class="inline-flex items-center justify-between gap-1.5 px-2.5 py-1.5 rounded-md border border-default text-[12px] text-default hover:bg-muted"
        @click="openStatusPage"
      >
        <span class="inline-flex items-center gap-1.5">
          <UIcon name="i-lucide-external-link" class="size-3.5" />
          Status Page
        </span>
        <UIcon name="i-lucide-arrow-right" class="size-3" />
      </button>
    </div>
  </div>
</template>
