<script setup lang="ts">
import { computed } from 'vue'
import { useResourceStore } from '@/stores/resourceStore'

const resourceStore = useResourceStore()

interface Row {
  label: string
  count: number
  color: string
  bg: string
}

const rows = computed<Row[]>(() => {
  const r = resourceStore.resources
  const total = r.length || 1
  return [
    {
      label: 'Operational',
      count: r.filter((x) => x.status === 'up').length,
      color: '#10B981',
      bg: '#ECFDF5',
    },
    {
      label: 'Degraded',
      count: r.filter((x) => (x as { status: string }).status === 'warning').length,
      color: '#F59E0B',
      bg: '#FFFBEB',
    },
    {
      label: 'Down',
      count: r.filter((x) => x.status === 'down').length,
      color: '#EF4444',
      bg: '#FEF2F2',
    },
    {
      label: 'Paused',
      count: r.filter((x) => x.status === 'paused').length,
      color: '#94A3B8',
      bg: '#F1F5F9',
    },
  ].map((row) => ({ ...row, _pct: total > 0 ? (row.count / total) * 100 : 0 })) as Row[]
})

const total = computed(() => resourceStore.resources.length || 1)
</script>

<template>
  <div class="bg-default rounded-lg border border-default p-6">
    <h3 class="text-base font-semibold text-highlighted mb-4">Status Breakdown</h3>
    <div class="flex flex-col gap-3.5">
      <div v-for="r in rows" :key="r.label" class="flex flex-col gap-1.5">
        <div class="flex items-center justify-between text-xs">
          <div class="flex items-center gap-2">
            <span class="size-1.5 rounded-full" :style="{ backgroundColor: r.color }" />
            <span class="text-default font-medium">{{ r.label }}</span>
          </div>
          <span class="text-muted font-mono">{{ r.count }}</span>
        </div>
        <div class="h-1.5 rounded-full bg-elevated overflow-hidden">
          <div
            class="h-full rounded-full"
            :style="{
              width: `${(r.count / total) * 100}%`,
              backgroundColor: r.color,
            }"
          />
        </div>
      </div>
    </div>
  </div>
</template>
