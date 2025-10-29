<script setup lang="ts">
import { ref, watch, computed } from 'vue'

import type { Resource } from '@/types'
import ResourceForm from './ResourceForm.vue'

interface Props {
  open?: boolean
  resource?: Resource | null
}

const props = withDefaults(defineProps<Props>(), {
  open: false,
  resource: null,
})

const emit = defineEmits<{
  'update:open': [value: boolean]
  submit: []
  cancel: []
}>()

// Local state to manage modal visibility
const isVisible = ref(props.open)

// Watch for external changes to open prop
watch(
  () => props.open,
  (newValue) => {
    isVisible.value = newValue
  },
)

// Watch local visibility and emit changes
watch(isVisible, (newValue) => {
  emit('update:open', newValue)
})

// Handle form submission
const handleSubmit = () => {
  isVisible.value = false
  emit('submit')
}

// Handle cancel
const handleCancel = () => {
  isVisible.value = false
  emit('cancel')
}

// Compute modal title
const modalTitle = computed(() => {
  return props.resource ? 'Edit Monitor' : 'New Monitor'
})
</script>

<template>
  <a-modal
    v-model:open="isVisible"
    :title="modalTitle"
    :footer="null"
    width="600px"
    @cancel="handleCancel"
  >
    <ResourceForm :resource="resource || undefined" @submit="handleSubmit" />
  </a-modal>
</template>

<style scoped></style>
