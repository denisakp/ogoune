<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useResourceStore } from '@/stores/resourceStore'
import { fetchStatsSummary } from '@/services/statsService'

const resourceStore = useResourceStore()

const totalResources = computed(() => resourceStore.resources.length)

const avgResponse = computed(() => {
  const resources = resourceStore.resources
  if (resources.length === 0) return 0
  const sum = resources.reduce((acc, r) => {
    const rt = (r as { response_time?: number }).response_time ?? 0
    return acc + rt
  }, 0)
  return Math.round(sum / resources.length)
})

const checksToday = ref<number | null>(null)
onMounted(async () => {
  try {
    const summary = await fetchStatsSummary('24h')
    checksToday.value = (summary as unknown as { total_checks_24h?: number }).total_checks_24h ?? 0
  } catch {
    checksToday.value = 0
  }
})

const cards = computed(() => [
  {
    label: 'Resources',
    value: String(totalResources.value),
    icon: 'i-lucide-server',
    iconBg: '#4F46E514',
    iconColor: '#4F46E5',
  },
  {
    label: 'Avg response',
    value: `${avgResponse.value} ms`,
    icon: 'i-lucide-zap',
    iconBg: '#0EA5E914',
    iconColor: '#0EA5E9',
  },
  {
    label: 'Checks today',
    value: checksToday.value === null ? '—' : String(checksToday.value),
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
        class="size-9 rounded-lg flex items-center justify-center"
        :style="{ backgroundColor: c.iconBg }"
      >
        <UIcon :name="c.icon" class="size-4" :style="{ color: c.iconColor }" />
      </div>
      <div class="flex flex-col">
        <span class="text-[10px] uppercase font-semibold tracking-wider text-slate-500">
          {{ c.label }}
        </span>
        <span class="text-lg font-bold text-slate-900 leading-tight">{{ c.value }}</span>
      </div>
    </div>
  </div>
</template>
