<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { fetchActivities } from '@/services/activityService'
import { timeAgo } from '@/libs/date-time.helper'
import { useResourceStore } from '@/stores/resourceStore'
import type { MonitoringActivity } from '@/types'

const activities = ref<MonitoringActivity[]>([])
const loading = ref(true)
const resourceStore = useResourceStore()

onMounted(async () => {
  try {
    if (resourceStore.resources.length === 0) {
      await resourceStore.loadResources()
    }
    activities.value = await fetchActivities()
  } catch {
    activities.value = []
  } finally {
    loading.value = false
  }
})

const resourceNameById = computed(() => {
  const map = new Map<string, string>()
  for (const r of resourceStore.resources) map.set(r.id, r.name)
  return map
})

function resourceLabel(a: MonitoringActivity): string {
  return resourceNameById.value.get(a.resource_id) ?? a.resource_id
}

const rows = computed(() => activities.value.slice(0, 10))

function dotColor(success: boolean): string {
  return success ? '#10B981' : '#EF4444'
}
function statusBadge(success: boolean) {
  return success
    ? { bg: '#ECFDF5', color: '#047857', text: 'Up' }
    : { bg: '#FEF2F2', color: '#B91C1C', text: 'Down' }
}
</script>

<template>
  <div class="bg-default rounded-lg border border-default overflow-hidden">
    <div class="flex items-center justify-between px-6 py-4">
      <h3 class="text-base font-semibold text-highlighted">Recent Activity</h3>
      <RouterLink to="/incidents" class="text-[13px] font-medium text-primary-600 hover:underline">
        View all
      </RouterLink>
    </div>
    <div
      class="border-t border-default bg-muted px-6 py-2.5 grid grid-cols-[1fr_180px_120px_140px] gap-2"
    >
      <span class="text-xs font-medium text-muted">Resource</span>
      <span class="text-xs font-medium text-muted">Event</span>
      <span class="text-xs font-medium text-muted">Status</span>
      <span class="text-xs font-medium text-muted">Time</span>
    </div>

    <div v-if="loading" class="px-6 py-4 space-y-3">
      <USkeleton v-for="i in 4" :key="i" class="h-6 w-full" />
    </div>
    <UEmpty
      v-else-if="rows.length === 0"
      variant="naked"
      icon="i-lucide-inbox"
      title="No activity yet"
      description="Your recent monitor checks will appear here."
    />
    <div v-else>
      <div
        v-for="a in rows"
        :key="a.id"
        class="px-6 py-3 grid grid-cols-[1fr_180px_120px_140px] gap-2 items-center border-t border-default first:border-t-0"
      >
        <div class="flex items-center gap-2 min-w-0">
          <span
            class="size-2 rounded-full shrink-0"
            :style="{ backgroundColor: dotColor(a.success) }"
          />
          <span class="text-sm text-highlighted truncate">{{ resourceLabel(a) }}</span>
        </div>
        <span class="text-sm text-muted truncate">{{
          a.message || (a.success ? 'Check passed' : 'Check failed')
        }}</span>
        <div>
          <span
            class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium"
            :style="{
              backgroundColor: statusBadge(a.success).bg,
              color: statusBadge(a.success).color,
            }"
          >
            {{ statusBadge(a.success).text }}
          </span>
        </div>
        <span class="text-xs text-muted">{{ timeAgo(a.created_at) }}</span>
      </div>
    </div>
  </div>
</template>
