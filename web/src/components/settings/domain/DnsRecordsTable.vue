<script setup lang="ts">
/**
 * DNS records table — 2 rows (CNAME + TXT) + per-row badge + Re-check footer.
 * Spec 059 US6 / FR-029. Migrated to UTable in PRD-013 / FR-004.
 */
import { computed, h, resolveComponent } from 'vue'
import type { TableColumn } from '@nuxt/ui'
import type { StatusPageDNSRecord as DNSRecord } from '@/types'

interface Props {
  records: DNSRecord[]
  rechecking: boolean
}
const props = defineProps<Props>()
defineEmits<{ (e: 'recheck'): void }>()

function badgeColor(status: string): 'success' | 'error' | 'neutral' {
  if (status === 'verified') return 'success'
  if (status === 'failed') return 'error'
  return 'neutral'
}

const rows = computed(() => props.records)

const columns: TableColumn<DNSRecord>[] = [
  {
    id: 'type',
    accessorKey: 'type',
    header: 'Type',
    cell: ({ row }) => h('span', { class: 'font-mono text-xs text-default' }, row.original.type),
  },
  {
    id: 'host',
    accessorKey: 'host',
    header: 'Host',
    cell: ({ row }) =>
      h('span', { class: 'font-mono text-xs text-default break-all' }, row.original.host),
  },
  {
    id: 'value',
    accessorKey: 'value',
    header: 'Value',
    cell: ({ row }) =>
      h('span', { class: 'font-mono text-xs text-default break-all' }, row.original.value),
  },
  {
    id: 'status',
    header: 'Status',
    cell: ({ row }) => {
      const UBadge = resolveComponent('UBadge')
      const children = [
        h(
          UBadge,
          { color: badgeColor(row.original.status), variant: 'subtle', size: 'xs' },
          () => row.original.status,
        ),
      ]
      if (row.original.last_error) {
        children.push(h('p', { class: 'text-[10px] text-error mt-1' }, row.original.last_error))
      }
      return h('div', children)
    },
  },
]
</script>

<template>
  <div class="rounded-xl border border-default/40 bg-default overflow-hidden">
    <UTable :data="rows" :columns="columns" />
    <div class="flex justify-end gap-2 border-t border-default/40 px-3 py-2 bg-elevated">
      <UButton
        size="xs"
        color="primary"
        variant="outline"
        :loading="rechecking"
        @click="$emit('recheck')"
      >
        Re-check DNS
      </UButton>
    </div>
  </div>
</template>
