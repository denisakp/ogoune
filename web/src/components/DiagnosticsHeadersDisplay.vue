<script setup lang="ts">
import { computed, h } from 'vue'
import type { TableColumn } from '@nuxt/ui'

interface Props {
  headers?: Record<string, string> | null
  title?: string
  emptyMessage?: string
}

interface HeaderRow {
  key: string
  value: string
}

const props = withDefaults(defineProps<Props>(), {
  title: 'Headers',
  emptyMessage: 'No headers available',
})

const headersList = computed<HeaderRow[]>(() => {
  if (!props.headers || typeof props.headers !== 'object') return []
  return Object.entries(props.headers).map(([key, value]) => ({
    key,
    value: String(value),
  }))
})

const columns: TableColumn<HeaderRow>[] = [
  {
    id: 'name',
    accessorKey: 'key',
    header: 'Name',
    cell: ({ row }) =>
      h('span', { class: 'font-mono text-xs text-muted' }, row.original.key),
  },
  {
    id: 'value',
    accessorKey: 'value',
    header: 'Value',
    cell: ({ row }) =>
      h(
        'span',
        { class: 'font-mono text-xs text-default break-all' },
        row.original.value,
      ),
  },
]
</script>

<template>
  <div>
    <div v-if="headersList.length === 0" class="text-xs text-muted">
      {{ emptyMessage }}
    </div>
    <UTable v-else :data="headersList" :columns="columns" />
  </div>
</template>
