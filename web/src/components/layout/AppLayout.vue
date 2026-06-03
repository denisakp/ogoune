<script setup lang="ts">
/**
 * Authenticated app shell — sidebar + topbar + main slot.
 * Scope: authenticated bundle only (spec 055 clarification Q3).
 * Contract: specs/055-slice-shared-components/contracts/app-layout.md
 */
import AppSidebar from './AppSidebar.vue'
import AppTopbar from './AppTopbar.vue'
</script>

<template>
  <div class="flex min-h-screen bg-default text-default font-sans">
    <AppSidebar />
    <main class="flex-1 min-w-0 flex flex-col">
      <AppTopbar />
      <div class="flex-1 p-6">
        <slot />
      </div>
    </main>
  </div>
</template>
