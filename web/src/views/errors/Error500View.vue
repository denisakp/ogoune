<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { createSyntheticIncident, type SyntheticIncidentRef } from './syntheticIncident'

const router = useRouter()

const incident = ref<SyntheticIncidentRef>(readIncidentFromState() ?? createSyntheticIncident(''))

function readIncidentFromState(): SyntheticIncidentRef | null {
  const state = window.history.state as
    | { incidentId?: string; occurredAt?: string; originalMessage?: string }
    | null
    | undefined
  if (!state?.incidentId || !state.occurredAt) return null
  return {
    id: state.incidentId,
    at: new Date(state.occurredAt),
    originalMessage: state.originalMessage ?? '',
  }
}

const relativeTime = ref('just now')
let timer: ReturnType<typeof setInterval> | undefined

function updateRelativeTime() {
  const diffMs = Date.now() - incident.value.at.getTime()
  const diffMin = Math.floor(diffMs / 60_000)
  if (diffMin < 1) relativeTime.value = 'just now'
  else if (diffMin === 1) relativeTime.value = '1 min ago'
  else if (diffMin < 60) relativeTime.value = `${diffMin} min ago`
  else {
    const hours = Math.floor(diffMin / 60)
    relativeTime.value = hours === 1 ? '1 hour ago' : `${hours} hours ago`
  }
}

onMounted(() => {
  updateRelativeTime()
  timer = setInterval(updateRelativeTime, 30_000)
})
onUnmounted(() => {
  if (timer) clearInterval(timer)
})

const pillLabel = computed(() => `Incident ID ${incident.value.id} · ${relativeTime.value}`)

function tryAgain() {
  router.replace({ name: 'Overview' })
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
        <div class="size-16 rounded-full bg-error/10 flex items-center justify-center">
          <UIcon name="i-lucide-circle-alert" class="size-8 text-error" />
        </div>

        <div class="flex flex-col gap-2">
          <span class="text-[64px] font-bold leading-none text-error">500</span>
          <h1 class="text-xl font-semibold text-default">Quelque chose s'est mal passé</h1>
          <p class="text-sm text-muted">
            An unexpected error happened. We've been notified — quote the incident below if you contact support.
          </p>
        </div>

        <div
          class="inline-flex items-center gap-2 px-3 py-1.5 rounded-full bg-error/10 border border-error/20"
        >
          <span class="size-1.5 rounded-full bg-error" aria-hidden="true"></span>
          <span class="text-xs font-medium text-error">{{ pillLabel }}</span>
        </div>

        <div class="flex flex-col sm:flex-row gap-3 w-full sm:w-auto">
          <UButton color="primary" size="md" block @click="tryAgain"> Try again </UButton>
          <UButton color="neutral" variant="ghost" size="md" block to="/status.html" external>
            View Status Page
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
