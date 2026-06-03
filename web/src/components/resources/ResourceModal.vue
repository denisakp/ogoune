<script setup lang="ts">
import { computed } from 'vue'

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

const isOpen = computed({
  get: () => props.open,
  set: (v) => emit('update:open', v),
})

const modalTitle = computed(() => (props.resource ? 'Edit monitor' : 'New monitor'))

function onFormSubmit() {
  emit('submit')
  isOpen.value = false
}

function onFormCancel() {
  emit('cancel')
  isOpen.value = false
}
</script>

<template>
  <UModal v-model:open="isOpen" :title="modalTitle" :ui="{ content: 'sm:max-w-xl' }">
    <template #body>
      <ResourceForm :resource="resource" @submit="onFormSubmit" @cancel="onFormCancel" />
    </template>
  </UModal>
</template>
