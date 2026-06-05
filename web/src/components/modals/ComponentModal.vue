<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { useComponentStore } from '@/stores/componentStore'
import type { Component, UpdateComponentPayload } from '@/types'

const props = defineProps<{
  visible: boolean
  editing: Component | null
}>()

const emit = defineEmits<{
  close: []
  submit: []
  'update:visible': [value: boolean]
}>()

const visibleModel = computed({
  get: () => props.visible,
  set: (val) => emit('update:visible', val),
})

const componentStore = useComponentStore()

const form = ref({
  name: '',
  description: '',
  groupingWindowSeconds: undefined as number | undefined,
})

const loading = ref(false)

watch(
  () => props.editing,
  (newVal) => {
    if (newVal) {
      form.value = {
        name: newVal.name,
        description: newVal.description || '',
        groupingWindowSeconds: newVal.grouping_window_seconds ?? undefined,
      }
    } else {
      form.value = {
        name: '',
        description: '',
        groupingWindowSeconds: undefined,
      }
    }
  },
  { immediate: true },
)

const handleSubmit = async () => {
  if (!form.value.name.trim()) return
  loading.value = true
  try {
    if (props.editing) {
      const payload: UpdateComponentPayload = {
        name: form.value.name,
        description: form.value.description || undefined,
        grouping_window_seconds: form.value.groupingWindowSeconds,
      }
      await componentStore.updateComponentData(props.editing.id, payload)
      emit('submit')
      emit('close')
    } else {
      emit('close')
    }
  } catch {
    /* errors surfaced by HTTP interceptor */
  } finally {
    loading.value = false
  }
}

const handleCancel = () => {
  emit('close')
}
</script>

<template>
  <UModal
    v-model:open="visibleModel"
    :title="editing ? 'Edit Component' : 'Create Component'"
    @update:open="(v: boolean) => !v && handleCancel()"
  >
    <template #body>
      <div class="space-y-4">
        <div>
          <label class="block text-sm font-medium mb-1">Component Name <span class="text-red-500">*</span></label>
          <UInput
            v-model="form.name"
            placeholder="e.g., Frontend Services"
            class="w-full"
            @keyup.enter="handleSubmit"
          />
        </div>

        <div>
          <label class="block text-sm font-medium mb-1">Description</label>
          <UTextarea v-model="form.description" placeholder="Optional description" :rows="3" class="w-full" />
        </div>

        <div>
          <label class="block text-sm font-medium mb-1">Alert Grouping Window (seconds, 0 = disabled)</label>
          <UInputNumber v-model="form.groupingWindowSeconds" :min="0" :max="300" placeholder="0" class="w-full" />
          <p class="mt-1 text-xs text-muted">
            Group multiple resource alerts into a single component notification within this window
            (10–300s). Set 0 to disable.
          </p>
        </div>
      </div>
    </template>
    <template #footer>
      <div class="flex justify-end gap-2 w-full">
        <UButton color="neutral" variant="soft" :disabled="loading" @click="handleCancel">Cancel</UButton>
        <UButton color="primary" :loading="loading" @click="handleSubmit">OK</UButton>
      </div>
    </template>
  </UModal>
</template>
