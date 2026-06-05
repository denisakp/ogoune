<script setup lang="ts">
// Public bundle is intentionally light-only — operators rely on a predictable
// white background for screenshots, embeds, and brand consistency. The wrapper
// scopes color-scheme to light so any leaked Tailwind `dark:` variant stays
// inert.
</script>

<template>
  <div class="status-app">
    <router-view />
  </div>
</template>

<style>
.status-app {
  color-scheme: light;
  background-color: #ffffff;
  color: #0f172a;
  min-height: 100vh;
}
</style>
