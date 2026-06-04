<script setup lang="ts">
/**
 * Sessions view — list active devices + revoke flows.
 * Spec 059 US1 / FR-008/FR-009/FR-009a.
 */
import { ref, computed, onMounted } from 'vue'
import sessionsService, { type Session } from '@/services/sessionsService'
import { useConfirm } from '@/composables/useConfirm'
import SessionRow from '@/components/settings/sessions/SessionRow.vue'

const loading = ref(true)
const sessions = ref<Session[]>([])
const error = ref<string | null>(null)

const showRevokeAll = computed(() => sessions.value.length > 1)

async function load() {
  loading.value = true
  error.value = null
  try {
    sessions.value = await sessionsService.list()
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load sessions'
  } finally {
    loading.value = false
  }
}

async function onRevoke(id: string) {
  const target = sessions.value.find((s) => s.id === id)
  if (!target) return
  const ok = await useConfirm({
    kind: 'destructive',
    title: `Revoke ${target.browser} on ${target.os}?`,
    body: 'Effective immediately. The signed-out device will be required to log in again on its next action.',
    ctaLabel: 'Revoke session',
  })
  if (!ok) return
  await sessionsService.revoke(id)
  sessions.value = sessions.value.filter((s) => s.id !== id)
}

async function onRevokeAllOthers() {
  const ok = await useConfirm({
    kind: 'destructive',
    title: 'Revoke all other sessions?',
    body: 'Effective immediately. Every other device will be signed out on its next action.',
    ctaLabel: 'Revoke all others',
  })
  if (!ok) return
  await sessionsService.revokeOthers()
  sessions.value = sessions.value.filter((s) => s.is_current)
}

onMounted(load)

defineExpose({ sessions, loading, error, onRevoke, onRevokeAllOthers, showRevokeAll })
</script>

<template>
  <div class="space-y-6">
    <header>
      <h1 class="text-lg font-semibold text-default">Active Sessions</h1>
      <p class="text-sm text-muted">
        Devices and browsers connected to your account. Revoke anything unfamiliar.
      </p>
    </header>

    <USkeleton v-if="loading" class="h-32 w-full" />

    <UAlert
      v-else-if="error"
      color="error"
      variant="soft"
      :title="error"
      icon="i-lucide-triangle-alert"
    />

    <UEmpty
      v-else-if="sessions.length === 0"
      icon="i-lucide-monitor-off"
      title="No active sessions"
    />

    <template v-else>
      <ul class="flex flex-col gap-2">
        <SessionRow v-for="s in sessions" :key="s.id" :session="s" @revoke="onRevoke" />
      </ul>

      <div
        v-if="showRevokeAll"
        class="flex items-center gap-4 rounded-xl border border-error/40 bg-error/5 px-4 py-3"
      >
        <div class="flex-1 min-w-0">
          <p class="text-sm font-semibold text-error">Revoke all other sessions</p>
          <p class="text-xs text-error/80">Signs out everywhere except this device</p>
        </div>
        <UButton color="error" size="sm" @click="onRevokeAllOthers">Revoke all</UButton>
      </div>
    </template>
  </div>
</template>
