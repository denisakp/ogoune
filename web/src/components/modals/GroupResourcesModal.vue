<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'

import { useComponentStore } from '@/stores/componentStore'
import { bulkAssignToComponent, createComponent } from '@/services/componentService'

interface Props {
  open?: boolean
  selectedIds: string[]
}

const props = withDefaults(defineProps<Props>(), { open: false })
const emit = defineEmits<{
  'update:open': [boolean]
  success: []
  cancel: []
}>()

const componentStore = useComponentStore()

const mode = ref<'pick' | 'create'>('pick')
const pickedComponentId = ref<string | null>(null)
const newComponentName = ref('')
const newComponentDescription = ref('')
const submitting = ref(false)
const error = ref<string | null>(null)

const isOpen = computed({
  get: () => props.open,
  set: (v) => emit('update:open', v),
})

onMounted(async () => {
  if (componentStore.components.length === 0) {
    await componentStore.loadComponents()
  }
})

watch(
  () => props.open,
  (v) => {
    if (v) {
      mode.value = 'pick'
      pickedComponentId.value = null
      newComponentName.value = ''
      newComponentDescription.value = ''
      error.value = null
    }
  },
)

const canSubmit = computed(() => {
  if (props.selectedIds.length === 0) return false
  if (mode.value === 'pick') return !!pickedComponentId.value
  return newComponentName.value.trim().length > 0
})

async function onSubmit() {
  if (!canSubmit.value) return
  submitting.value = true
  error.value = null
  try {
    let targetId: string
    if (mode.value === 'create') {
      const created = await createComponent({
        name: newComponentName.value.trim(),
        description: newComponentDescription.value.trim() || undefined,
        resource_ids: props.selectedIds,
      })
      targetId = created.id
    } else {
      targetId = pickedComponentId.value!
      await bulkAssignToComponent(targetId, { resource_ids: props.selectedIds })
    }
    emit('success')
    isOpen.value = false
  } catch (e) {
    error.value = (e as Error).message || 'Failed to assign resources'
  } finally {
    submitting.value = false
  }
}

function onCancel() {
  emit('cancel')
  isOpen.value = false
}

defineExpose({ mode, pickedComponentId, newComponentName, onSubmit, canSubmit })
</script>

<template>
  <UModal
    v-model:open="isOpen"
    title="Group resources into component"
    :ui="{ content: 'sm:max-w-md' }"
  >
    <template #body>
      <div class="space-y-4">
        <div class="text-xs text-slate-600">
          {{ selectedIds.length }} resource{{ selectedIds.length > 1 ? 's' : '' }} selected
        </div>

        <UTabs
          v-model="mode"
          :items="[
            { label: 'Existing component', value: 'pick', icon: 'i-lucide-folder' },
            { label: '+ New component', value: 'create', icon: 'i-lucide-folder-plus' },
          ]"
          size="sm"
        />

        <div v-if="mode === 'pick'" class="space-y-2 max-h-60 overflow-y-auto">
          <div
            v-if="componentStore.components.length === 0"
            class="text-xs text-slate-500 py-6 text-center"
          >
            No components yet. Switch to "+ New component".
          </div>
          <label
            v-for="c in componentStore.components"
            :key="c.id"
            class="flex items-center gap-3 px-3 py-2.5 rounded-md border cursor-pointer transition-colors"
            :class="
              pickedComponentId === c.id
                ? 'border-primary-600 bg-primary-50'
                : 'border-slate-200 hover:border-slate-300'
            "
          >
            <input
              v-model="pickedComponentId"
              type="radio"
              :value="c.id"
              class="accent-primary-600"
            />
            <UIcon name="i-lucide-folder" class="size-4 text-slate-500" />
            <span class="text-sm font-medium text-slate-900">{{ c.name }}</span>
          </label>
        </div>

        <div v-else class="space-y-3">
          <div class="space-y-1.5">
            <label class="text-xs font-medium text-slate-900">Component name</label>
            <UInput
              v-model="newComponentName"
              placeholder="e.g. Payment Systems"
              size="md"
              class="w-full"
            />
          </div>
          <div class="space-y-1.5">
            <label class="text-xs font-medium text-slate-900">Description (optional)</label>
            <UInput
              v-model="newComponentDescription"
              placeholder="A short description"
              size="md"
              class="w-full"
            />
          </div>
        </div>

        <p v-if="error" class="text-xs text-red-600">{{ error }}</p>

        <div class="flex justify-end gap-2 pt-2 border-t border-slate-200">
          <UButton color="neutral" variant="ghost" @click="onCancel">Cancel</UButton>
          <UButton color="primary" :loading="submitting" :disabled="!canSubmit" @click="onSubmit">
            Group {{ selectedIds.length }} resource{{ selectedIds.length > 1 ? 's' : '' }}
          </UButton>
        </div>
      </div>
    </template>
  </UModal>
</template>
