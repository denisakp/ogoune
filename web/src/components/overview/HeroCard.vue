<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'

import { useResourceStore } from '@/stores/resourceStore'

const resourceStore = useResourceStore()
const router = useRouter()

const totalResources = computed(() => resourceStore.resources.length)
const downCount = computed(() => resourceStore.resources.filter((r) => r.status === 'down').length)
const upCount = computed(() => resourceStore.resources.filter((r) => r.status === 'up').length)

const uptimePct = computed(() => {
  if (totalResources.value === 0) return 100
  return Math.round((upCount.value / totalResources.value) * 1000) / 10
})
const uptimeWhole = computed(() => Math.floor(uptimePct.value))
const uptimeDecimal = computed(() => {
  const dec = Math.round((uptimePct.value - Math.floor(uptimePct.value)) * 100)
  return dec.toString().padStart(2, '0')
})

const hasIncidents = computed(() => downCount.value > 0)

const sparkBars = computed(() => {
  const heights = [
    52, 48, 60, 55, 62, 58, 50, 45, 53, 48, 56, 65, 70, 62, 58, 50, 48, 55, 72, 60, 52, 45, 50, 58,
    65, 55, 48, 42, 55, 60,
  ]
  return heights.map((h, i) => ({
    h,
    isDown: hasIncidents.value && i === 22,
  }))
})
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
                  backgroundColor: b.isDown ? '#EF4444' : '#4F46E5',
                  opacity: b.isDown ? 1 : 0.35,
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
        @click="router.push('/status')"
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
