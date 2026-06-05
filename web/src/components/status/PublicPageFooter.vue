<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { loadRuntimeConfig } from '@/composables/useRuntimeConfig'
import type { RuntimeConfig } from '@/services/runtimeConfigService'

const runtime = ref<RuntimeConfig | null>(null)

onMounted(async () => {
  runtime.value = await loadRuntimeConfig()
})

// Server-injected meta tag — defense in depth. The runtime API can be
// disabled on hardened deployments; if the meta tag asserts "true" we
// still render the credit. Fall back to runtime API when meta is absent.
const metaLicense = computed(() => {
  if (typeof document === 'undefined') return null
  const tag = document.querySelector('meta[name="x-ogoune-license"]')
  return tag?.getAttribute('content') ?? null
})

const showCredit = computed(() => {
  const fromMeta = metaLicense.value
  if (fromMeta === 'community') return true
  if (fromMeta === 'enterprise-suppressed') return false
  // Fall through to runtime config (defaults to required = true).
  return runtime.value?.powered_by_required ?? true
})
</script>

<template>
  <footer
    class="text-center py-6 text-sm text-gray-500"
    data-testid="public-footer"
  >
    <a
      v-if="showCredit"
      href="https://ogoune.dev"
      target="_blank"
      rel="noopener noreferrer"
      class="hover:text-gray-700 dark:hover:text-gray-300"
      data-testid="powered-by"
    >
      Powered by Ogoune
    </a>
  </footer>
</template>
