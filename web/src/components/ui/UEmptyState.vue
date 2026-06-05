<script setup lang="ts">
interface Props {
  icon: string
  title: string
  description?: string
}

defineProps<Props>()
</script>

<template>
  <div class="flex flex-col items-center text-center py-12 px-6">
    <UIcon :name="icon" class="size-12 text-muted" />
    <h3 class="text-base font-semibold mt-4">{{ title }}</h3>
    <p v-if="description" class="text-sm text-muted mt-2 max-w-md">{{ description }}</p>
    <div class="mt-6 flex gap-2">
      <slot name="actions" />
    </div>
  </div>
</template>
