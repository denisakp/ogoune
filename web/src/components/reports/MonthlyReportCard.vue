<script setup lang="ts">
import { computed, ref } from 'vue'
import { useReports } from '@/composables/useReports'

const reports = useReports()

const saving = ref(false)
const inlineError = ref<string | null>(null)

const monthly = computed(() => reports.monthly.value)

const nextSendLabel = computed(() => {
  const now = new Date()
  const next = new Date(now.getFullYear(), now.getMonth() + 1, 1)
  return next.toLocaleString(undefined, { day: 'numeric', month: 'long', year: 'numeric' })
})

const lastSentLabel = computed(() => {
  if (!monthly.value?.lastSentAt) return 'Never'
  return new Date(monthly.value.lastSentAt).toLocaleString(undefined, {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
})

async function onToggle(next: boolean) {
  if (!monthly.value) return
  inlineError.value = null
  saving.value = true
  try {
    await reports.toggleMonthly(next)
  } catch (e) {
    if (e instanceof Error && e.message === 'NO_RESOURCES') {
      inlineError.value = 'Add a monitor first'
    } else {
      inlineError.value = e instanceof Error ? e.message : 'Failed to update'
    }
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <UCard
    :ui="{ body: 'p-0' }"
    class="bg-default border border-default"
    data-testid="monthly-report-card"
  >
    <div class="px-5 py-4 border-b border-default flex items-start justify-between gap-4">
      <div class="flex-1 min-w-0">
        <div class="flex items-center gap-2">
          <UIcon name="i-lucide-file-text" class="size-5 text-primary" />
          <h3 class="text-base font-semibold text-default">Monthly Health Report</h3>
          <UBadge color="primary" variant="subtle" size="sm">Community</UBadge>
        </div>
        <p class="text-sm text-muted mt-1">
          Automatic monthly summary of uptime, incidents, and downtime — sent to one recipient.
        </p>
      </div>
      <USwitch
        :model-value="monthly?.enabled ?? false"
        :disabled="saving || !monthly"
        :aria-label="monthly?.enabled ? 'Disable monthly report' : 'Enable monthly report'"
        data-testid="monthly-report-toggle"
        @update:model-value="(v: boolean) => onToggle(v)"
      />
    </div>

    <div v-if="inlineError" class="px-5 py-3 border-b border-default bg-error/10">
      <p class="text-sm text-error">{{ inlineError }}</p>
    </div>

    <div
      v-if="monthly"
      class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-px bg-default"
      data-testid="monthly-report-info"
    >
      <div class="px-5 py-4 border-default">
        <div class="text-[10px] font-semibold tracking-wider text-muted uppercase">Recipient</div>
        <div class="text-sm text-default mt-1 truncate">{{ monthly.recipientEmail }}</div>
      </div>
      <div class="px-5 py-4 border-l border-default">
        <div class="text-[10px] font-semibold tracking-wider text-muted uppercase">Schedule</div>
        <div class="text-sm text-default mt-1">
          {{ monthly.enabled ? `Next: ${nextSendLabel}` : '1st of every month' }}
        </div>
      </div>
      <div class="px-5 py-4 border-l border-default">
        <div class="text-[10px] font-semibold tracking-wider text-muted uppercase">Scope</div>
        <div class="text-sm text-default mt-1">All resources</div>
      </div>
      <div class="px-5 py-4 border-l border-default">
        <div class="text-[10px] font-semibold tracking-wider text-muted uppercase">Last Sent</div>
        <div class="text-sm text-default mt-1">{{ lastSentLabel }}</div>
      </div>
    </div>
  </UCard>
</template>
