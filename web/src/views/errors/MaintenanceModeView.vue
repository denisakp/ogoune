<script setup lang="ts">
import { onBeforeUnmount, onMounted } from 'vue'

const eta = (import.meta.env.VITE_MAINTENANCE_ETA as string | undefined) ?? ''
const message =
  (import.meta.env.VITE_MAINTENANCE_MESSAGE as string | undefined) ?? 'On bricole sous le capot'

// Force the dark theme on this view regardless of user preference.
onMounted(() => {
  document.documentElement.classList.add('dark')
})
onBeforeUnmount(() => {
  document.documentElement.classList.remove('dark')
})
</script>

<template>
  <div class="min-h-screen flex flex-col bg-default text-default">
    <header
      class="flex items-center justify-between px-6 py-5 border-b border-default bg-default"
    >
      <div class="flex items-center gap-2">
        <UIcon name="i-lucide-activity" class="size-6 text-primary" />
        <span class="text-lg font-bold text-default">Ogoune</span>
      </div>
      <a
        href="/status.html"
        class="text-sm text-muted hover:text-primary"
      >
        status.ogoune.com
      </a>
    </header>

    <main class="flex-1 flex items-center justify-center px-6 py-12 bg-default">
      <div class="max-w-md w-full text-center flex flex-col items-center gap-6">
        <div class="size-16 rounded-full bg-warning/10 flex items-center justify-center">
          <UIcon name="i-lucide-wrench" class="size-8 text-warning" />
        </div>

        <div class="flex flex-col gap-2">
          <span class="text-[40px] font-bold leading-tight tracking-wide text-warning">
            MAINTENANCE
          </span>
          <h1 class="text-xl font-semibold text-default">{{ message }}</h1>
        </div>

        <div
          v-if="eta"
          class="inline-flex items-center gap-2 px-3 py-1.5 rounded-full bg-warning/10 border border-warning/20"
        >
          <span class="size-1.5 rounded-full bg-warning" aria-hidden="true"></span>
          <span class="text-xs font-medium text-warning">Scheduled maintenance · {{ eta }}</span>
        </div>

        <div class="flex flex-col sm:flex-row gap-3 w-full sm:w-auto">
          <UButton color="primary" size="md" block to="/status.html" external>
            View Status Page
          </UButton>
          <UButton
            color="neutral"
            variant="ghost"
            size="md"
            block
            to="mailto:hello@ogoune.com?subject=Notify me when Ogoune is back"
            external
          >
            Notify me when back
          </UButton>
        </div>
      </div>
    </main>

    <footer class="px-6 py-5 text-center text-xs text-muted bg-default border-t border-default">
      We'll be back shortly. Thanks for your patience.
    </footer>
  </div>
</template>
