<script setup lang="ts">
type Kind = 'tag' | 'component' | 'type' | 'status'

interface Props {
  kind: Kind
  value: string
}

defineProps<Props>()
const emit = defineEmits<{ remove: [] }>()
</script>

<template>
  <span
    class="inline-flex items-center gap-1.5 px-2 py-1 rounded-md bg-elevated text-xs font-medium text-default"
    :data-kind="kind"
  >
    <span class="text-muted">{{ kind }}:</span>
    {{ value }}
    <button
      type="button"
      class="ml-1 hover:text-error transition-colors"
      :aria-label="`Remove ${kind} filter ${value}`"
      @click="emit('remove')"
    >
      <UIcon name="i-lucide-x" class="size-3.5" />
    </button>
  </span>
</template>
