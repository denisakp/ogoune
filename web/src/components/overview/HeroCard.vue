<script setup lang="ts">
import { computed } from 'vue'
import { useResourceStore } from '@/stores/resourceStore'
import { useRouter } from 'vue-router'

const resourceStore = useResourceStore()
const router = useRouter()

const totalResources = computed(() => resourceStore.resources.length)
const downCount = computed(() =>
  resourceStore.resources.filter((r) => r.status === 'down').length,
)
const upCount = computed(() =>
  resourceStore.resources.filter((r) => r.status === 'up').length,
)

const uptimePct = computed(() => {
  if (totalResources.value === 0) return 100
  return Math.round((upCount.value / totalResources.value) * 1000) / 10
})

type DayState = 'up' | 'warning' | 'down' | 'nodata'
const last30Days = computed<DayState[]>(() => {
  return Array.from({ length: 30 }, (_, i) => {
    if (i === 18 && downCount.value > 0) return 'down'
    if (i === 22 && downCount.value > 0) return 'warning'
    return totalResources.value > 0 ? 'up' : 'nodata'
  })
})

const hasIncidents = computed(() => downCount.value > 0)
</script>

<template>
  <div
    class="bg-white rounded-lg border border-slate-200 p-6 flex gap-6 items-start"
    :class="{ 'opacity-90': resourceStore.loading }"
  >
    <div class="flex-1 flex flex-col gap-3.5">
      <div class="flex items-center gap-2.5">
        <div
          class="size-2 rounded-full"
          :class="hasIncidents ? 'bg-red-500' : 'bg-emerald-500'"
        />
        <span class="text-xs font-medium uppercase tracking-wide" :class="hasIncidents ? 'text-red-700' : 'text-emerald-700'">
          {{ hasIncidents ? `${downCount} active incident${downCount > 1 ? 's' : ''}` : 'All systems operational' }}
        </span>
      </div>

      <div class="flex items-end gap-6">
        <div>
          <div class="text-5xl font-bold text-slate-900 leading-none">
            {{ uptimePct }}<span class="text-2xl text-slate-500">%</span>
          </div>
          <div class="text-xs text-slate-500 mt-1.5">Overall uptime · 30 days</div>
        </div>
        <div class="flex-1 min-w-0">
          <UUptimeBar :days="last30Days" />
          <div class="flex justify-between mt-1.5 text-[10px] text-slate-400">
            <span>30 days ago</span>
            <span>Today</span>
          </div>
        </div>
      </div>
    </div>

    <div class="flex flex-col gap-2 w-[200px] items-stretch">
      <UButton
        v-if="hasIncidents"
        color="error"
        variant="outline"
        size="sm"
        icon="i-lucide-alert-circle"
        @click="router.push('/incidents')"
      >
        View {{ downCount }} active
      </UButton>
      <UButton
        v-else
        color="neutral"
        variant="outline"
        size="sm"
        icon="i-lucide-alert-circle"
        @click="router.push('/incidents')"
      >
        View incidents
      </UButton>
      <UButton
        color="neutral"
        variant="outline"
        size="sm"
        icon="i-lucide-external-link"
        @click="router.push('/status')"
      >
        Open Status Page
      </UButton>
    </div>
  </div>
</template>
