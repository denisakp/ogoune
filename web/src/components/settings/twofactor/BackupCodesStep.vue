<script setup lang="ts">
/**
 * 2FA setup — step 3: backup codes 2×5 + Copy + Download.
 * Spec 059 US2 / FR-010 — codes shown once.
 */
import { ref } from 'vue'

interface Props {
  codes: string[]
}
const props = defineProps<Props>()
defineEmits<{ (e: 'done'): void }>()

const copied = ref(false)

async function copy() {
  try {
    await navigator.clipboard.writeText(props.codes.join('\n'))
    copied.value = true
    setTimeout(() => (copied.value = false), 2000)
  } catch {
    // best-effort — fall through silently
  }
}

function download() {
  const blob = new Blob([props.codes.join('\n')], { type: 'text/plain' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = 'ogoune-2fa-backup-codes.txt'
  document.body.appendChild(a)
  a.click()
  a.remove()
  URL.revokeObjectURL(url)
}

defineExpose({ copy, download, copied })
</script>

<template>
  <div class="space-y-4">
    <header>
      <h2 class="text-base font-semibold text-default">Save your backup codes</h2>
      <p class="text-sm text-muted">
        Each code works once. Use one if you lose access to your authenticator app.
      </p>
    </header>

    <UAlert
      color="warning"
      variant="soft"
      icon="i-lucide-triangle-alert"
      title="Save these codes — they won't be shown again."
    />

    <div class="grid grid-cols-2 gap-2 rounded-md border border-default/40 p-3 bg-elevated">
      <code v-for="c in codes" :key="c" class="font-mono text-sm text-default px-2 py-1">
        {{ c }}
      </code>
    </div>

    <div class="flex flex-wrap items-center gap-2 pt-2">
      <UButton color="primary" variant="outline" icon="i-lucide-copy" @click="copy">
        {{ copied ? 'Copied' : 'Copy' }}
      </UButton>
      <UButton color="primary" variant="outline" icon="i-lucide-download" @click="download">
        Download .txt
      </UButton>
      <UButton color="primary" class="ml-auto" @click="$emit('done')">Done</UButton>
    </div>
  </div>
</template>
