<script setup lang="ts">
import { ref, watch } from 'vue'

interface Props {
  modelValue?: Record<string, string>
}
const props = withDefaults(defineProps<Props>(), {
  modelValue: () => ({}),
})
const emit = defineEmits<{ 'update:modelValue': [Record<string, string>] }>()

interface Row {
  id: string
  name: string
  value: string
}

function rowsFromHeaders(h: Record<string, string>): Row[] {
  return Object.entries(h).map(([name, value]) => ({
    id: `h_${Math.random().toString(36).slice(2, 10)}`,
    name,
    value,
  }))
}

const rows = ref<Row[]>(rowsFromHeaders(props.modelValue))

watch(
  () => props.modelValue,
  (v) => {
    if (rowsToHeaders(rows.value) === JSON.stringify(v)) return
    rows.value = rowsFromHeaders(v)
  },
  { deep: true },
)

function rowsToHeaders(r: Row[]): string {
  const obj: Record<string, string> = {}
  for (const row of r) {
    if (row.name.trim()) obj[row.name.trim()] = row.value
  }
  return JSON.stringify(obj)
}

function emitChange() {
  const obj: Record<string, string> = {}
  for (const r of rows.value) {
    if (r.name.trim()) obj[r.name.trim()] = r.value
  }
  emit('update:modelValue', obj)
}

function addRow() {
  rows.value.push({
    id: `h_${Math.random().toString(36).slice(2, 10)}`,
    name: '',
    value: '',
  })
}

function removeRow(id: string) {
  rows.value = rows.value.filter((r) => r.id !== id)
  emitChange()
}

defineExpose({ rows, addRow, removeRow, emitChange })
</script>

<template>
  <div class="space-y-2">
    <div v-for="row in rows" :key="row.id" class="flex items-center gap-2">
      <UInput
        v-model="row.name"
        placeholder="Header name"
        size="sm"
        class="flex-1"
        @blur="emitChange"
      />
      <UInput
        v-model="row.value"
        placeholder="Header value"
        size="sm"
        class="flex-1"
        @blur="emitChange"
      />
      <UButton
        color="neutral"
        variant="ghost"
        icon="i-lucide-x"
        size="sm"
        @click="removeRow(row.id)"
      />
    </div>

    <UButton color="neutral" variant="outline" size="sm" icon="i-lucide-plus" @click="addRow">
      Add header
    </UButton>
  </div>
</template>
