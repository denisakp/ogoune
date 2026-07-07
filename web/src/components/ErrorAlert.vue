<script setup lang="ts">
defineProps<{
  title?: string
  message: string
}>()
</script>

<template>
  <UAlert
    color="error"
    variant="soft"
    icon="i-lucide-circle-alert"
    :title="title"
    :description="message"
  />
</template>
