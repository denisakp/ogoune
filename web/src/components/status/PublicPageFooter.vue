<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { loadRuntimeConfig } from '@/composables/useRuntimeConfig'
import type { RuntimeConfig } from '@/services/runtimeConfigService'

const props = defineProps<{
  brandName?: string
  backHref?: string
  backLabel?: string
}>()

const runtime = ref<RuntimeConfig | null>(null)

onMounted(async () => {
  runtime.value = await loadRuntimeConfig()
})

const metaLicense = computed(() => {
  if (typeof document === 'undefined') return null
  const tag = document.querySelector('meta[name="x-ogoune-license"]')
  return tag?.getAttribute('content') ?? null
})

const showCredit = computed(() => {
  const fromMeta = metaLicense.value
  if (fromMeta === 'community') return true
  if (fromMeta === 'enterprise-suppressed') return false
  return runtime.value?.powered_by_required ?? true
})

const year = new Date().getUTCFullYear()
</script>

<template>
  <footer
    class="border-t border-gray-200 mt-12"
    data-testid="public-footer"
  >
    <div
      class="max-w-5xl mx-auto px-6 py-4 flex items-center justify-between text-xs text-gray-500"
    >
      <div class="flex items-center gap-4">
        <a
          v-if="backHref"
          :href="backHref"
          class="hover:text-gray-700 inline-flex items-center gap-1"
          data-testid="back-link"
        >
          ← {{ backLabel || 'Current Status' }}
        </a>
        <a
          v-if="showCredit"
          href="https://ogoune.dev"
          target="_blank"
          rel="noopener noreferrer"
          class="hover:text-gray-700"
          data-testid="powered-by"
        >
          Powered by Ogoune
        </a>
      </div>
      <slot name="right">
        <span v-if="brandName" class="text-gray-400" data-testid="copyright">
          © {{ year }} {{ brandName }}. All rights reserved.
        </span>
      </slot>
    </div>
  </footer>
</template>
