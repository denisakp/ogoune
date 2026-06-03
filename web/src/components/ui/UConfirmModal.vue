<script setup lang="ts">
/**
 * Internal modal used by `useConfirm()`. Not auto-imported by callers — always
 * reached through `useConfirm({...})`.
 *
 * Contract: specs/055-slice-shared-components/contracts/shared-components.md
 */
interface Props {
  kind?: 'default' | 'destructive'
  title: string
  body: string
  ctaLabel: string
}

const props = withDefaults(defineProps<Props>(), {
  kind: 'default',
})

const emit = defineEmits<{
  // `useOverlay().open()` resolves with whatever value is passed to `emit('close', value)`.
  close: [value: boolean]
}>()

const ctaColor = props.kind === 'destructive' ? 'error' : 'primary'
const headerIcon = props.kind === 'destructive' ? 'i-lucide-alert-triangle' : 'i-lucide-help-circle'

function confirm() {
  emit('close', true)
}

function dismiss() {
  emit('close', false)
}

defineExpose({ confirm, dismiss, ctaColor, headerIcon })
</script>

<template>
  <UModal :open="true" @update:open="(v: boolean) => !v && dismiss()">
    <template #content>
      <div class="p-6 max-w-md">
        <div class="flex items-start gap-3">
          <UIcon
            :name="headerIcon"
            :class="kind === 'destructive' ? 'text-error size-6' : 'text-primary size-6'"
          />
          <div class="flex-1">
            <h3 class="text-base font-semibold">{{ title }}</h3>
            <p class="text-sm text-muted mt-2">{{ body }}</p>
          </div>
        </div>
        <div class="flex justify-end gap-2 mt-6">
          <UButton color="neutral" variant="ghost" @click="dismiss">Cancel</UButton>
          <UButton :color="ctaColor" @click="confirm">{{ ctaLabel }}</UButton>
        </div>
      </div>
    </template>
  </UModal>
</template>
