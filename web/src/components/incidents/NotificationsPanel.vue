<script setup lang="ts">
import { computed } from 'vue'
import type { IncidentEventStep } from '@/types'

interface Props {
  events: IncidentEventStep[]
}
const props = defineProps<Props>()

const alertEvents = computed(() =>
  props.events
    .filter((e) => e.step === 'alert_sent')
    .sort((a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime()),
)

function timeOfDay(iso: string): string {
  return new Date(iso).toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit' })
}
</script>

<template>
  <div class="bg-default rounded-lg border border-default overflow-hidden">
    <div class="px-5 py-3 border-b border-default flex items-center justify-between">
      <h3 class="text-sm font-semibold text-highlighted">Notifications Sent</h3>
      <span class="text-xs text-muted">{{ alertEvents.length }}</span>
    </div>
    <div v-if="alertEvents.length === 0" class="px-5 py-6 text-sm text-muted text-center">
      No notifications were dispatched for this incident.
    </div>
    <div v-else>
      <div
        v-for="e in alertEvents"
        :key="e.id"
        class="flex items-center gap-3 px-5 py-2.5 border-b border-muted last:border-0 text-sm"
      >
        <UIcon name="i-lucide-send" class="size-3.5 text-blue-500 shrink-0" />
        <span class="text-default flex-1 truncate">{{ e.message || 'Alert dispatched' }}</span>
        <span class="text-xs text-muted font-mono">{{ timeOfDay(e.created_at) }}</span>
        <span
          class="inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-medium"
          style="background-color: #ecfdf5; color: #047857"
        >
          Delivered
        </span>
      </div>
    </div>
  </div>
</template>
