<script setup lang="ts">
defineProps<{
  saving?: boolean
  dirty?: boolean
}>()

const emit = defineEmits<{
  save: []
  cancel: []
}>()
</script>

<template>
  <div
    class="sticky top-0 z-20 mb-4 px-4 py-3 rounded-lg border border-warning/40 bg-warning/10 flex items-center justify-between gap-4 shadow-sm"
    role="region"
    aria-label="Edit mode"
    data-testid="dashboard-edit-banner"
  >
    <div class="flex items-center gap-3 min-w-0">
      <UIcon name="i-lucide-pencil" class="size-5 text-warning shrink-0" />
      <div class="min-w-0">
        <p class="text-sm font-semibold text-default">Edit Mode</p>
        <p class="text-xs text-muted">
          Drag widgets to reorder, remove with the trash icon, then save.
        </p>
      </div>
    </div>
    <div class="flex items-center gap-2 shrink-0">
      <UButton
        color="neutral"
        variant="ghost"
        size="sm"
        data-testid="edit-cancel"
        @click="emit('cancel')"
      >
        Cancel
      </UButton>
      <UButton
        color="primary"
        size="sm"
        :loading="saving"
        :disabled="saving"
        data-testid="edit-save"
        @click="emit('save')"
      >
        Save
      </UButton>
    </div>
  </div>
</template>
