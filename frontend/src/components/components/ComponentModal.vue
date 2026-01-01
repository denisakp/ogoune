<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { useComponents } from '@/composables/useComponents'
import type { Component, UpdateComponent } from '@/types'

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
})

const loading = ref(false)

watch(
  () => props.editing,
  (newVal) => {
    if (newVal) {
      form.value = {
        name: newVal.name,
        description: newVal.description || '',
      }
    } else {
      form.value = {
        name: '',
        description: '',
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
      const payload: UpdateComponent = {
        name: form.value.name,
        description: form.value.description || undefined,
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
    </a-form>
  </a-modal>
</template>
