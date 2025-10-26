<script setup lang="ts">
defineProps<{
  message?: string
}>()
</script>

<template>
  <div style="text-align: center; padding: 48px">
    <a-spin size="large" />
    <p v-if="message" style="margin-top: 16px; color: rgba(0,0,0,0.45)">{{ message }}</p>
  </div>
</template>

<style scoped></style>
