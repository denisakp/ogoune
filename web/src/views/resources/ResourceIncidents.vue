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
const loadMoreIncidents = () => {
  incidentsToShow.value += 3
}

const getIncidentStatus = (incident: Incident) =>
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
  <a-card>
    <template #title>
      <div style="display: flex; justify-content: space-between; align-items: center">
        <span style="font-size: 14px; font-weight: 600">Recent incidents</span>
        <a-badge :count="sortedIncidents.length" :number-style="{ backgroundColor: '#52c41a' }" />
      </div>
    </template>
    <template v-if="sortedIncidents.length > 0">
      <a-timeline>
        <a-timeline-item
          v-for="incident in visibleIncidents"
          :key="incident.id"
          :color="incident.resolved_at ? 'green' : 'red'"
        >
          <template #dot>
            <a-icon-clock-circle
              v-if="!incident.resolved_at"
              style="font-size: 16px; color: #f5222d"
            />
            <a-icon-check-circle v-else style="font-size: 16px; color: #52c41a" />
          </template>
          <div style="padding-bottom: 16px">
            <div
              style="
                display: flex;
                justify-content: space-between;
                align-items: start;
                margin-bottom: 8px;
              "
            >
              <div style="flex: 1">
                <div style="display: flex; align-items: center; gap: 8px; margin-bottom: 4px">
                  <a-tag :color="getIncidentStatus(incident).color">{{
                    getIncidentStatus(incident).text
                  }}</a-tag>
                  <span style="font-size: 12px; color: rgba(0, 0, 0, 0.45)">
                    {{ formatDuration(incident.started_at, incident.resolved_at) }}
                  </span>
                </div>
                <div style="font-weight: 500; margin-bottom: 4px">{{ incident.reason }}</div>
                <div style="font-size: 12px; color: rgba(0, 0, 0, 0.65); margin-bottom: 4px">
                  <strong>Cause:</strong> {{ incident.cause }}
                </div>
              </div>
            </div>
            <div style="font-size: 12px; color: rgba(0, 0, 0, 0.45); margin-bottom: 8px">
              <div>
                <a-icon-calendar style="margin-right: 4px" />Started:
                {{ formatDate(incident.started_at) }}
              </div>
              <div v-if="incident.resolved_at" style="margin-top: 4px">
                <a-icon-check style="margin-right: 4px" />Resolved:
                {{ formatDate(incident.resolved_at) }}
              </div>
              <div v-else style="margin-top: 4px; color: #f5222d">
                <a-icon-exclamation-circle style="margin-right: 4px" />Still ongoing
              </div>
            </div>
            <a-collapse v-if="incident.details" ghost size="small">
              <a-collapse-panel key="1" header="Technical details">
                <div
                  style="
                    font-size: 12px;
                    font-family: monospace;
                    background: rgba(0, 0, 0, 0.02);
                    padding: 12px;
                    border-radius: 4px;
                    word-break: break-word;
                  "
                >
                  {{ decodeDetails(incident.details) }}
                </div>
              </a-collapse-panel>
            </a-collapse>
          </div>
        </a-timeline-item>
      </a-timeline>
      <div
        v-if="hasMoreIncidents"
        style="
          text-align: center;
          margin-top: 16px;
          padding-top: 16px;
          border-top: 1px solid rgba(0, 0, 0, 0.06);
        "
      >
        <a-button @click="loadMoreIncidents">
          <template #icon><a-icon-down /></template>
          Load more incidents ({{ sortedIncidents.length - incidentsToShow }} remaining)
        </a-button>
      </div>
    </template>
    <template v-else>
      <a-empty description="No incidents recorded">
        <template #image><a-icon-smile style="font-size: 48px; color: #52c41a" /></template>
      </a-empty>
    </template>
  </a-card>
</template>
