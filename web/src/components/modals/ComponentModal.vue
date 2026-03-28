<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { useComponents } from '@/composables/useComponents'
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

const { updateComponent } = useComponents()

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
  if (!form.value.name.trim()) {
    return
  }

  loading.value = true
  try {
    if (props.editing) {
      const payload: UpdateComponentPayload = {
        name: form.value.name,
        description: form.value.description || undefined,
        grouping_window_seconds: form.value.groupingWindowSeconds,
      }
      await updateComponent(props.editing.id, payload)
      emit('submit')
      emit('close')
    } else {
      // Create is not supported directly - components must be created via bulk grouping in MonitorsView
      console.warn('Components should be created via bulk grouping in Monitors')
      emit('close')
    }
  } catch (err) {
    console.error('Failed to save component:', err)
  } finally {
    loading.value = false
  }
}

const handleCancel = () => {
  emit('close')
}
</script>

<template>
  <a-modal
    :visible="visibleModel"
    :title="editing ? 'Edit Component' : 'Create Component'"
    :confirm-loading="loading"
    @ok="handleSubmit"
    @cancel="handleCancel"
    @update:visible="visibleModel = $event"
  >
    <a-form :model="form" layout="vertical">
      <a-form-item label="Component Name" required>
        <a-input
          v-model:value="form.name"
          placeholder="e.g., Frontend Services"
          @keyup.enter="handleSubmit"
        />
      </a-form-item>

      <a-form-item label="Description">
        <a-textarea v-model:value="form.description" placeholder="Optional description" :rows="3" />
      </a-form-item>

      <a-form-item label="Alert Grouping Window (seconds, 0 = disabled)">
        <a-input-number
          v-model:value="form.groupingWindowSeconds"
          :min="0"
          :max="300"
          style="width: 100%"
          placeholder="0"
        />
        <div style="margin-top: 4px; font-size: 12px; color: rgba(0, 0, 0, 0.45)">
          Group multiple resource alerts into a single component notification within this window
          (10–300s). Set 0 to disable.
        </div>
      </a-form-item>
    </a-form>
  </a-modal>
</template>
