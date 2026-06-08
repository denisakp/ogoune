<script setup lang="ts">
import { useRouter } from 'vue-router'

const router = useRouter()

function goOverview() {
  router.push({ name: 'Overview' })
}

function goResources() {
  router.push({ name: 'Resources' })
}
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
        <div class="size-16 rounded-full bg-primary/10 flex items-center justify-center">
          <UIcon name="i-lucide-compass" class="size-8 text-primary" />
        </div>

        <div class="flex flex-col gap-2">
          <span class="text-[64px] font-bold leading-none text-primary">404</span>
          <h1 class="text-xl font-semibold text-default">Page introuvable</h1>
          <p class="text-sm text-muted">
            The page you were looking for doesn't exist or has been moved.
          </p>
        </div>

        <div class="flex flex-col sm:flex-row gap-3 w-full sm:w-auto">
          <UButton color="primary" size="md" block @click="goOverview">
            Back to Overview
          </UButton>
          <UButton color="neutral" variant="ghost" size="md" block @click="goResources">
            Browse Resources
          </UButton>
        </div>
      </div>
    </main>

    <footer class="px-6 py-5 text-center text-xs text-muted bg-default border-t border-default">
      If this keeps happening,
      <a href="mailto:hello@ogoune.com" class="text-primary hover:underline">hello@ogoune.com</a>
    </footer>
  </div>
</template>
