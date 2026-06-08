<script setup lang="ts">
defineProps<{
  message?: string
}>()
</script>

<template>
  <div class="text-center py-12">
    <UIcon name="i-lucide-loader-circle" class="size-8 animate-spin text-primary-500" />
    <p v-if="message" class="mt-4 text-sm text-muted">{{ message }}</p>
  </div>
</template>
