<script setup lang="ts">
/**
 * Wrapper around NuxtUI `<UTable>` adding filter chips + pagination + actions.
 * Slice 2 (Resources/Incidents lists) is the first real consumer.
 */
import { computed } from 'vue'

export interface TableColumn {
  key: string
  label: string
}

export interface FilterChipDescriptor {
  kind: 'tag' | 'component' | 'type' | 'status'
  value: string
}

export interface RowAction {
  label: string
  icon?: string
  onSelect?: (row: unknown) => void
}

interface Props {
  columns: TableColumn[]
  rows: unknown[]
  loading?: boolean
  filters?: FilterChipDescriptor[]
  pagination?: { page: number; perPage: number; total: number }
  rowActions?: RowAction[]
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
  filters: () => [],
  rowActions: () => [],
})

const emit = defineEmits<{
  page: [page: number]
  'filter-remove': [filter: FilterChipDescriptor]
  action: [payload: { action: RowAction; row: unknown }]
}>()

const totalPages = computed(() =>
  props.pagination ? Math.max(1, Math.ceil(props.pagination.total / props.pagination.perPage)) : 1,
)

defineExpose({ totalPages, emit })
</script>

<template>
  <div class="space-y-3">
    <div v-if="filters.length > 0" class="flex flex-wrap gap-2">
      <UFilterChip
        v-for="f in filters"
        :key="`${f.kind}:${f.value}`"
        :kind="f.kind"
        :value="f.value"
        @remove="emit('filter-remove', f)"
      />
    </div>

    <UTable :columns="columns" :data="rows" :loading="loading">
      <template
        v-for="col in columns"
        :key="col.key"
        #[`cell-${col.key}`]="{ row }: { row: Record<string, unknown> }"
      >
        <slot :name="`cell-${col.key}`" :row="row">{{ row[col.key] }}</slot>
      </template>
      <template v-if="rowActions.length > 0" #cell-actions="{ row }: { row: unknown }">
        <slot name="actions" :row="row">
          <UButton
            v-for="a in rowActions"
            :key="a.label"
            :icon="a.icon"
            color="neutral"
            variant="ghost"
            size="xs"
            @click="emit('action', { action: a, row })"
          >
            {{ a.label }}
          </UButton>
        </slot>
      </template>
      <template #empty>
        <slot name="empty">
          <div class="text-center py-8 text-muted text-sm">No data</div>
        </slot>
      </template>
    </UTable>

    <div
      v-if="pagination && totalPages > 1"
      class="flex items-center justify-between text-xs text-muted"
    >
      <span>Page {{ pagination.page }} of {{ totalPages }}</span>
      <div class="flex gap-1">
        <UButton
          size="xs"
          color="neutral"
          variant="ghost"
          :disabled="pagination.page <= 1"
          icon="i-lucide-chevron-left"
          @click="emit('page', pagination.page - 1)"
        />
        <UButton
          size="xs"
          color="neutral"
          variant="ghost"
          :disabled="pagination.page >= totalPages"
          icon="i-lucide-chevron-right"
          @click="emit('page', pagination.page + 1)"
        />
      </div>
    </div>
  </div>
</template>
