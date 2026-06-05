<script setup lang="ts">
/**
 * One-shot reveal banner for a newly created API key.
 * Spec 059 US4 — the secret is shown ONCE. Persisted only in the
 * Pinia transient store; cleared on dismiss / reload.
 */
import { ref } from 'vue'
import type { CreateAPIKeyResponse } from '@/services/accountService'

interface Props {
  payload: CreateAPIKeyResponse
}
const props = defineProps<Props>()
const emit = defineEmits<{ (e: 'dismiss'): void }>()

const copied = ref(false)

async function copy() {
  try {
    await navigator.clipboard.writeText(props.payload.key)
    copied.value = true
    setTimeout(() => (copied.value = false), 2000)
  } catch {
    // best-effort
  }
}

function download() {
  const content =
    `Ogoune API key\n` +
    `Name: ${props.payload.name}\n` +
    `Prefix: ${props.payload.key_prefix}\n` +
    `Scope: ${props.payload.scope}\n` +
    `Created: ${props.payload.created_at}\n` +
    `Expires: ${props.payload.expires_at ?? 'never'}\n\n` +
    `Secret (shown only once):\n${props.payload.key}\n\n` +
    `Treat this like a password.\n`
  const blob = new Blob([content], { type: 'text/plain' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `ogoune-api-key-${props.payload.key_prefix}.txt`
  document.body.appendChild(a)
  a.click()
  a.remove()
  URL.revokeObjectURL(url)
}

defineExpose({ copy, download, copied })
</script>

<template>
  <div class="rounded-xl border border-success/40 bg-success/5 p-4 space-y-3">
    <div class="flex items-center gap-2">
      <UIcon name="i-lucide-check-circle-2" class="size-5 text-success shrink-0" />
      <p class="text-sm font-semibold text-default flex-1">
        Key created — copy it now, you won't see it again
      </p>
      <UButton size="xs" variant="ghost" icon="i-lucide-x" @click="emit('dismiss')" />
    </div>

    <div class="flex items-center gap-2">
      <code
        class="flex-1 font-mono text-sm text-default break-all rounded-md bg-default border border-default px-3 py-2"
        data-testid="key-secret"
      >
        {{ payload.key }}
      </code>
      <UButton color="neutral" variant="outline" size="sm" icon="i-lucide-copy" @click="copy">
        {{ copied ? 'Copied' : 'Copy' }}
      </UButton>
    </div>
  </div>
</template>
