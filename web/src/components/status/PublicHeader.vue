<script setup lang="ts">
import { computed } from 'vue'
import type { PublicBranding } from '@/types'

const props = defineProps<{
  branding?: PublicBranding | null
}>()

const name = computed(() => props.branding?.name || 'Status Page')
const homepage = computed(() => props.branding?.homepage_url || '#')
const logoLight = computed(() => props.branding?.logo_url_light || '')
</script>

<template>
  <header class="border-b border-gray-200">
    <div class="max-w-5xl mx-auto px-6 h-16 flex items-center justify-between">
      <a :href="homepage" class="flex items-center gap-2 group">
        <img v-if="logoLight" :src="logoLight" :alt="name" class="h-7 w-auto" />
        <svg
          v-else
          class="h-7 w-7 text-indigo-500"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2.4"
          stroke-linecap="round"
          stroke-linejoin="round"
        >
          <path d="M3 12h3l3 -8 4 16 3 -8 3 0" />
        </svg>
        <span class="text-base font-semibold text-gray-900">{{ name }}</span>
      </a>
      <a
        href="#"
        class="text-sm text-indigo-600 hover:underline"
        data-testid="subscribe-link"
        @click.prevent
      >
        Subscribe to updates
      </a>
    </div>
  </header>
</template>
