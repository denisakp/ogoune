<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import { useResourceStore } from '@/stores/resourceStore'
import { fetchPublicStatusSummary } from '@/services/statusPublicService'
import type { PublicStatusSummary } from '@/types'

const resourceStore = useResourceStore()
const router = useRouter()

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

const uptimePct = computed(() => {
  if (knownRatios.value.length === 0) {
    if (resourceStore.resources.length === 0) return 100
    const up = resourceStore.resources.filter((r) => r.status === 'up').length
    return Math.round((up / resourceStore.resources.length) * 1000) / 10
  }
  const mean = knownRatios.value.reduce((a, b) => a + b, 0) / knownRatios.value.length
  return Math.round(mean * 1000) / 10
})

const uptimeWhole = computed(() => Math.floor(uptimePct.value))
const uptimeDecimal = computed(() => {
  const dec = Math.round((uptimePct.value - Math.floor(uptimePct.value)) * 100)
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
  <div class="bg-white rounded-lg border border-slate-200 p-5 flex gap-4 items-start">
    <div class="flex-1 min-w-0 flex flex-col gap-3.5">
      <div class="flex items-center gap-2.5 flex-wrap min-w-0">
        <div
          v-if="hasIncidents"
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
        <template v-if="hasIncidents">
          <span class="text-slate-400 text-xs">·</span>
          <span class="text-[13px] text-slate-600">
            Detected just now, {{ downCount }} resource{{ downCount > 1 ? 's' : '' }} down
          </span>
        </template>
      </div>

      <div class="flex items-end gap-5">
        <div class="shrink-0">
          <div class="flex items-end gap-1">
            <span class="text-[44px] font-bold text-slate-900 leading-none tracking-tight">
              {{ uptimeWhole }}.{{ uptimeDecimal }}
            </span>
            <span class="text-xl font-semibold text-slate-500 leading-none pb-1">%</span>
          </div>
          <div class="text-[10px] font-medium text-slate-400 uppercase tracking-wider mt-1">
            30-day uptime
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
            <div class="flex justify-between mt-1 text-[10px] text-slate-400">
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
        class="inline-flex items-center justify-between gap-1.5 px-2.5 py-1.5 rounded-md border border-slate-200 text-[12px] text-slate-700 hover:bg-slate-50"
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
        class="inline-flex items-center justify-between gap-1.5 px-2.5 py-1.5 rounded-md border border-slate-200 text-[12px] text-slate-700 hover:bg-slate-50"
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
