<script setup lang="ts">
defineProps<{
  title?: string
  message: string
}>()
</script>

<template>
  <a-alert
    type="error"
    show-icon
  >
    <template #message v-if="title">
      {{ title }}
    </template>
    <template #description>
      {{ message }}
    </template>
  </a-alert>
</template>

<style scoped></style>
