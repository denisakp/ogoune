<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useReports } from '@/composables/useReports'
import MonthlyReportCard from '@/components/reports/MonthlyReportCard.vue'
import ReportHistoryList from '@/components/reports/ReportHistoryList.vue'
import ReportPreviewInline from '@/components/reports/ReportPreviewInline.vue'

const reports = useReports()

onMounted(() => {
  void reports.loadAll()
})

const selectedId = ref<string | null>(null)

const selectedEntry = computed(() => {
  if (selectedId.value) {
    const explicit = reports.history.value.find((h) => h.id === selectedId.value)
    if (explicit) return explicit
  }
  return reports.latestDelivered.value
})

function onSelect(id: string) {
  selectedId.value = id
}
</script>

<template>
  <div class="px-6 py-6 max-w-7xl mx-auto">
    <header class="mb-6">
      <h1 class="text-2xl font-bold text-default">Reports</h1>
      <p class="text-sm text-muted mt-1">Scheduled summaries of your monitoring activity.</p>
    </header>

    <div class="grid grid-cols-1 lg:grid-cols-[1fr_420px] gap-6">
      <div class="space-y-4">
        <MonthlyReportCard />

        <UAlert
          color="primary"
          variant="subtle"
          icon="i-lucide-sparkles"
          title="Custom Reports — Enterprise"
          description="Daily, weekly, or cron schedules · filterable scope · multiple recipients."
          data-testid="reports-ee-banner"
        >
          <template #actions>
            <UButton
              color="primary"
              size="sm"
              to="/settings/account?tab=plan"
              data-testid="reports-upgrade-cta"
            >
              Upgrade
            </UButton>
          </template>
        </UAlert>

        <ReportHistoryList
          :entries="reports.history.value"
          :selected-id="selectedEntry?.id"
          @select="onSelect"
        />
      </div>

      <aside data-testid="reports-preview-column">
        <div
          v-if="!selectedEntry"
          class="p-6 border border-dashed border-default rounded text-center bg-muted"
          data-testid="reports-preview-empty"
        >
          <UIcon name="i-lucide-file-search" class="size-8 text-muted mx-auto mb-2" />
          <p class="text-sm text-muted">No report sent yet.</p>
          <p class="text-xs text-muted mt-1">Toggle the monthly report on to schedule the first one.</p>
        </div>
        <ReportPreviewInline v-else :entry="selectedEntry" />
      </aside>
    </div>
  </div>
</template>
