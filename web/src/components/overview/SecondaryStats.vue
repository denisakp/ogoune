<script setup lang="ts">
import { computed, inject, onMounted, ref, watch, type Ref } from 'vue'
import { useResourceStore } from '@/stores/resourceStore'
import { fetchStatsSummary } from '@/services/statsService'

const resourceStore = useResourceStore()

type Range = '1h' | '6h' | '24h' | '7d' | '30d'
const injected = inject<{ range: Ref<Range> } | null>('overview.range', null)
const range = computed<Range>(() => injected?.range.value ?? '24h')

const totalResources = computed(() => resourceStore.resources.length)
const downOrDegraded = computed(
  () =>
    resourceStore.resources.filter(
      (r) => r.status === 'down' || (r as { status: string }).status === 'warning',
    ).length,
)

const avgResponse = computed(() => {
  const resources = resourceStore.resources
  if (resources.length === 0) return 0
  const sum = resources.reduce((acc, r) => {
    const rt = (r as { response_time?: number }).response_time ?? 0
    return acc + rt
  }, 0)
  return Math.round(sum / resources.length)
})

const checksCount = ref<number | null>(null)
const checksDelta = ref<string>('')

async function refresh() {
  try {
    const summary = await fetchStatsSummary(range.value)
    // The API still uses the legacy `total_checks_24h` key. Treat it as
    // "total over the selected range" and rename the card label
    // dynamically below.
    const count =
      (summary as unknown as { total_checks_24h?: number; total_checks?: number }).total_checks ??
      (summary as unknown as { total_checks_24h?: number }).total_checks_24h ??
      0
    checksCount.value = count
    const delta = (summary as unknown as { delta_pct?: number }).delta_pct
    checksDelta.value = typeof delta === 'number' ? `${delta >= 0 ? '+' : ''}${delta}% vs prev` : ''
  } catch {
    checksCount.value = 0
    checksDelta.value = ''
  }
}

onMounted(refresh)
watch(range, refresh)

const RANGE_LABELS: Record<Range, string> = {
  '1h': 'Checks (1h)',
  '6h': 'Checks (6h)',
  '24h': 'Checks today',
  '7d': 'Checks (7d)',
  '30d': 'Checks (30d)',
}

const cards = computed(() => [
  {
    label: 'Resources',
    value: String(totalResources.value),
    sub: downOrDegraded.value > 0 ? `${downOrDegraded.value} down/degraded` : 'all healthy',
    icon: 'i-lucide-server',
    iconBg: '#4F46E514',
    iconColor: '#4F46E5',
  },
  {
    label: 'Avg response',
    value: `${avgResponse.value}ms`,
    sub: '',
    icon: 'i-lucide-zap',
    iconBg: '#0EA5E914',
    iconColor: '#0EA5E9',
  },
  {
    label: RANGE_LABELS[range.value],
    value: checksCount.value === null ? '—' : checksCount.value.toLocaleString(),
    sub: checksDelta.value,
    icon: 'i-lucide-activity',
    iconBg: '#10B98114',
    iconColor: '#10B981',
  },
])
</script>

<template>
  <div class="flex flex-col gap-3">
    <div
      v-for="c in cards"
      :key="c.label"
      class="bg-white rounded-lg border border-slate-200 p-4 flex items-center gap-3.5"
    >
      <div
        class="size-9 rounded-lg flex items-center justify-center shrink-0"
        :style="{ backgroundColor: c.iconBg }"
      >
        <UIcon :name="c.icon" class="size-4" :style="{ color: c.iconColor }" />
      </div>
      <div class="flex flex-col min-w-0 flex-1">
        <span class="text-[10px] uppercase font-semibold tracking-wider text-slate-500">
          {{ c.label }}
        </span>
        <div class="flex items-baseline gap-2">
          <span class="text-xl font-bold text-slate-900 leading-tight">{{ c.value }}</span>
          <span v-if="c.sub" class="text-[11px] text-slate-400 truncate">{{ c.sub }}</span>
        </div>
      </div>
    </div>
  </div>
</template>
