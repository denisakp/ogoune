<script setup lang="ts">
import { ref, computed } from 'vue'
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
const expandedDetails = ref<Set<string>>(new Set())
const loadMoreIncidents = () => {
  incidentsToShow.value += 3
}
const toggleDetails = (id: string) => {
  if (expandedDetails.value.has(id)) expandedDetails.value.delete(id)
  else expandedDetails.value.add(id)
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
      <ul class="space-y-4">
        <li
          v-for="incident in visibleIncidents"
          :key="incident.id"
          class="relative pl-6"
        >
          <span
            class="absolute left-0 top-1 size-3 rounded-full"
            :style="{ backgroundColor: incident.resolved_at ? '#52c41a' : '#f5222d' }"
          ></span>
          <div class="pb-4">
            <div class="flex justify-between items-start mb-2">
              <div class="flex-1">
                <div class="flex items-center gap-2 mb-1">
                  <UBadge :color="getIncidentStatus(incident).color" variant="subtle">
                    {{ getIncidentStatus(incident).text }}
                  </UBadge>
                  <span class="text-xs text-muted">
                    {{ formatDuration(incident.started_at, incident.resolved_at) }}
                  </span>
                </div>
                <div class="font-medium mb-1">{{ incident.reason }}</div>
                <div class="text-xs text-muted mb-1">
                  <strong>Cause:</strong> {{ incident.cause }}
                </div>
              </div>
            </div>
            <div class="text-xs text-muted mb-2">
              <div class="flex items-center gap-1">
                <UIcon name="i-lucide-calendar" class="size-3" />
                <span>Started: {{ formatDate(incident.started_at) }}</span>
              </div>
              <div v-if="incident.resolved_at" class="mt-1 flex items-center gap-1">
                <UIcon name="i-lucide-check" class="size-3" />
                <span>Resolved: {{ formatDate(incident.resolved_at) }}</span>
              </div>
              <div v-else class="mt-1 flex items-center gap-1 text-red-500">
                <UIcon name="i-lucide-circle-alert" class="size-3" />
                <span>Still ongoing</span>
              </div>
            </div>
            <div v-if="incident.details">
              <UButton size="xs" color="neutral" variant="ghost" @click="toggleDetails(incident.id)">
                <UIcon :name="expandedDetails.has(incident.id) ? 'i-lucide-chevron-up' : 'i-lucide-chevron-down'" class="size-3" />
                Technical details
              </UButton>
              <div
                v-if="expandedDetails.has(incident.id)"
                class="mt-2 text-xs font-mono p-3 rounded bg-slate-50 dark:bg-slate-900 break-words"
              >
                {{ decodeDetails(incident.details) }}
              </div>
            </div>
          </div>
        </li>
      </ul>
      <div
        v-if="hasMoreIncidents"
        class="text-center mt-4 pt-4 border-t border-slate-100 dark:border-slate-800"
      >
        <UButton color="neutral" variant="soft" icon="i-lucide-chevron-down" @click="loadMoreIncidents">
          Load more incidents ({{ sortedIncidents.length - incidentsToShow }} remaining)
        </UButton>
      </div>
    </template>
    <template v-else>
      <UEmpty icon="i-lucide-smile" title="No incidents recorded" />
    </template>
  </UCard>
</template>
