<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useResourceStore } from '@/stores/resourceStore'

import HeroCard from '@/components/overview/HeroCard.vue'
import SecondaryStats from '@/components/overview/SecondaryStats.vue'
import StatusBreakdown from '@/components/overview/StatusBreakdown.vue'
import RecentActivity from '@/components/overview/RecentActivity.vue'
import ResponseTimeChart from '@/components/ResponseTimeChart.vue'

const resourceStore = useResourceStore()
const router = useRouter()

onMounted(() => {
  if (resourceStore.resources.length === 0) {
    void resourceStore.loadResources()
  }
})
</script>

<template>
  <div class="bg-default text-default min-h-full">
    <div class="flex items-center justify-between mb-7">
      <div>
        <h1 class="text-2xl font-semibold text-slate-900">Overview</h1>
        <p class="text-sm text-slate-600 mt-1">Live view across all resources</p>
      </div>
      <div class="flex items-center gap-3">
        <UButton color="neutral" variant="outline" size="sm" icon="i-lucide-calendar">
          Last 24 hours
        </UButton>
        <UButton
          color="primary"
          size="sm"
          icon="i-lucide-plus"
          @click="router.push('/monitors')"
        >
          Add Resource
        </UButton>
      </div>
    </div>

    <div class="grid grid-cols-[1fr_320px] gap-4 mb-7">
      <HeroCard />
      <SecondaryStats />
    </div>

    <div class="grid grid-cols-[1fr_320px] gap-4 mb-7">
      <div class="bg-white rounded-lg border border-slate-200 p-6">
        <h3 class="text-base font-semibold text-slate-900 mb-4">Response Time</h3>
        <ResponseTimeChart />
      </div>
      <StatusBreakdown />
    </div>

    <RecentActivity />
  </div>
</template>
