<script setup lang="ts">
import dayjs from 'dayjs'
import type { Maintenance } from '@/types'

const props = defineProps<{
  maintenances: Maintenance[]
  loading?: boolean
}>()

const emit = defineEmits<{
  (e: 'edit', maintenance: Maintenance): void
  (e: 'finish', id: string): void
  (e: 'delete', id: string): void
}>()

const statusColor = (status: Maintenance['status']) => {
  switch (status) {
    case 'active':
      return 'green'
    case 'scheduled':
      return 'blue'
    case 'finished':
      return 'default'
    case 'cancelled':
      return 'red'
    default:
      return 'default'
  }
}

const formatDate = (value?: string | null) => {
  return value ? dayjs(value).format('YYYY-MM-DD HH:mm') : '—'
}
</script>

<template>
  <a-table
    :data-source="props.maintenances"
    :loading="props.loading"
    row-key="id"
    :pagination="false"
  >
    <a-table-column key="title" title="Title" data-index="title" />
    <a-table-column key="strategy" title="Strategy">
      <template #default="{ record }">
        {{ record.strategy === 'cron' ? 'Cron' : 'One-time' }}
      </template>
    </a-table-column>
    <a-table-column key="status" title="Status">
      <template #default="{ record }">
        <a-tag :color="statusColor(record.status)">{{ record.status }}</a-tag>
      </template>
    </a-table-column>
    <a-table-column key="schedule" title="Schedule">
      <template #default="{ record }">
        <div v-if="record.strategy === 'cron'">
          <div><strong>Cron:</strong> {{ record.cron_expr || '—' }}</div>
          <div><strong>Window:</strong> {{ record.window_minutes ?? '—' }} min</div>
        </div>
        <div v-else>
          <div><strong>Start:</strong> {{ formatDate(record.start_at) }}</div>
          <div><strong>End:</strong> {{ formatDate(record.end_at) }}</div>
        </div>
      </template>
    </a-table-column>
    <a-table-column key="resources" title="Resources">
      <template #default="{ record }">
        {{ record.resources ? record.resources.length : 0 }}
      </template>
    </a-table-column>
    <a-table-column key="actions" title="Actions">
      <template #default="{ record }">
        <div style="display: flex; gap: 8px; flex-wrap: wrap">
          <a-button size="small" @click="emit('edit', record)">Edit</a-button>
          <a-button
            size="small"
            type="default"
            :disabled="record.status === 'finished' || record.status === 'cancelled'"
            @click="emit('finish', record.id)"
          >
            Finish
          </a-button>
          <a-button danger size="small" @click="emit('delete', record.id)"> Delete </a-button>
        </div>
      </template>
    </a-table-column>
  </a-table>
</template>

<style scoped></style>
