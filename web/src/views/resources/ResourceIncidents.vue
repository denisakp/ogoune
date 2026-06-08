<script setup lang="ts">
import { computed, ref } from 'vue'
import type { Incident } from '@/types'
import { formatDate, formatDuration } from '@/utils/formatters'

const props = defineProps<{ incidents: Incident[] }>()

const incidentsToShow = ref(3)

const sortedIncidents = computed(() =>
  [...props.incidents].sort(
    (a, b) => new Date(b.started_at).getTime() - new Date(a.started_at).getTime(),
  ),
)
const visibleIncidents = computed(() => sortedIncidents.value.slice(0, incidentsToShow.value))
const hasMoreIncidents = computed(() => sortedIncidents.value.length > incidentsToShow.value)
const loadMoreIncidents = () => {
  incidentsToShow.value += 3
}

const getIncidentStatus = (incident: Incident): { text: string; color: 'success' | 'error' } =>
  incident.resolved_at ? { text: 'Resolved', color: 'success' } : { text: 'Active', color: 'error' }

const decodeDetails = (details?: string): string => {
  if (!details) return 'No details available'
  try {
    return atob(details)
  } catch {
    return details
  }
}

type TimelineItem = {
  value: string
  color: 'success' | 'error'
  icon: string
  incident: Incident
}

const timelineItems = computed<TimelineItem[]>(() =>
  visibleIncidents.value.map((incident) => ({
    value: incident.id,
    color: incident.resolved_at ? 'success' : 'error',
    icon: incident.resolved_at ? 'i-lucide-check' : 'i-lucide-circle',
    incident,
  })),
)
</script>

<template>
  <UCard>
    <template #header>
      <div class="flex justify-between items-center">
        <span class="text-sm font-semibold">Recent incidents</span>
        <UBadge color="success" variant="subtle">{{ sortedIncidents.length }}</UBadge>
      </div>
    </template>
    <template v-if="sortedIncidents.length > 0">
      <UTimeline :items="timelineItems">
        <template #title="{ item }">
          <div class="flex items-center gap-2">
            <UBadge :color="getIncidentStatus(item.incident).color" variant="subtle">
              {{ getIncidentStatus(item.incident).text }}
            </UBadge>
            <span class="text-xs text-muted">
              {{ formatDuration(item.incident.started_at, item.incident.resolved_at) }}
            </span>
          </div>
          <div class="font-medium mt-1">{{ item.incident.reason }}</div>
        </template>
        <template #description="{ item }">
          <div class="text-xs text-muted mb-1">
            <strong>Cause:</strong> {{ item.incident.cause }}
          </div>
          <div class="text-xs text-muted mb-2">
            <div class="flex items-center gap-1">
              <UIcon name="i-lucide-calendar" class="size-3" />
              <span>Started: {{ formatDate(item.incident.started_at) }}</span>
            </div>
            <div v-if="item.incident.resolved_at" class="mt-1 flex items-center gap-1">
              <UIcon name="i-lucide-check" class="size-3" />
              <span>Resolved: {{ formatDate(item.incident.resolved_at) }}</span>
            </div>
            <div v-else class="mt-1 flex items-center gap-1 text-red-500">
              <UIcon name="i-lucide-circle-alert" class="size-3" />
              <span>Still ongoing</span>
            </div>
          </div>
          <UCollapsible v-if="item.incident.details">
            <template #default="{ open }">
              <UButton size="xs" color="neutral" variant="ghost">
                <UIcon
                  :name="open ? 'i-lucide-chevron-up' : 'i-lucide-chevron-down'"
                  class="size-3"
                />
                Technical details
              </UButton>
            </template>
            <template #content>
              <div
                class="mt-2 text-xs font-mono p-3 rounded bg-slate-50 dark:bg-slate-900 break-words"
              >
                {{ decodeDetails(item.incident.details) }}
              </div>
            </template>
          </UCollapsible>
        </template>
      </UTimeline>
      <div
        v-if="hasMoreIncidents"
        class="text-center mt-4 pt-4 border-t border-slate-100 dark:border-slate-800"
      >
        <UButton
          color="neutral"
          variant="soft"
          icon="i-lucide-chevron-down"
          @click="loadMoreIncidents"
        >
          Load more incidents ({{ sortedIncidents.length - incidentsToShow }} remaining)
        </UButton>
      </div>
    </template>
    <template v-else>
      <UEmpty icon="i-lucide-smile" title="No incidents recorded" />
    </template>
  </UCard>
</template>
