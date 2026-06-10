<script setup lang="ts">
// Placeholder — implemented in spec 070 / US3 (T042) + US4 (T050).
// Existence required so the router (Foundational T007) can resolve its lazy import.
defineProps<{
  id: string
  editMode?: boolean
}>()
</script>

<template>
  <div class="px-6 py-6 max-w-7xl mx-auto">
    <h1 class="text-2xl font-bold text-default">Dashboard {{ id }}</h1>
    <p class="text-sm text-muted mt-2">
      {{ editMode ? 'Edit mode' : 'Read mode' }} — coming in spec 070 / US3.
    </p>
  </div>
</template>
