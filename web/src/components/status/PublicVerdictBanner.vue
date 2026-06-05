<script setup lang="ts">
import { computed } from 'vue'
import type { PublicVerdict } from '@/types'

const props = defineProps<{
  verdict: PublicVerdict
  generatedAt: Date | string | null
  secondsAgo: number | null
}>()

const palette = computed(() => {
  switch (props.verdict.status) {
    case 'operational':
      return {
        pill: 'bg-emerald-50 text-emerald-700 border-emerald-200 dark:bg-emerald-950/40 dark:text-emerald-300 dark:border-emerald-800',
        dot: 'text-emerald-500',
      }
    case 'partial_degradation':
      return {
        pill: 'bg-orange-50 text-orange-700 border-orange-200 dark:bg-orange-950/40 dark:text-orange-300 dark:border-orange-800',
        dot: 'text-orange-500',
      }
    case 'major_outage':
      return {
        pill: 'bg-red-50 text-red-700 border-red-200 dark:bg-red-950/40 dark:text-red-300 dark:border-red-800',
        dot: 'text-red-500',
      }
    default:
      return {
        pill: 'bg-gray-50 text-gray-700 border-gray-200 dark:bg-gray-900/40 dark:text-gray-300 dark:border-gray-800',
        dot: 'text-gray-500',
      }
  }
})

const updatedLabel = computed(() => {
  if (!props.generatedAt) return 'Last updated: just now'
  const date = props.generatedAt instanceof Date ? props.generatedAt : new Date(props.generatedAt)
  try {
    const fmt = date.toLocaleString('en-US', {
      month: 'long',
      day: 'numeric',
      year: 'numeric',
      hour: 'numeric',
      minute: '2-digit',
      hour12: true,
      timeZone: 'UTC',
    })
    return `Last updated: ${fmt} UTC`
  } catch {
    return `Last updated: ${date.toISOString()}`
  }
})
</script>

<template>
  <section
    class="flex flex-col items-center gap-3 py-12"
    role="status"
    :data-status="verdict.status"
    data-testid="verdict-banner"
  >
    <div
      :class="['inline-flex items-center gap-2 rounded-2xl border px-5 py-3 text-base font-semibold shadow-sm', palette.pill]"
    >
      <svg
        v-if="verdict.status === 'operational'"
        class="size-5"
        :class="palette.dot"
        viewBox="0 0 20 20"
        fill="currentColor"
      >
        <path
          fill-rule="evenodd"
          d="M16.704 5.29a1 1 0 010 1.42l-7.5 7.5a1 1 0 01-1.414 0l-3.5-3.5a1 1 0 011.414-1.42l2.793 2.794 6.793-6.794a1 1 0 011.414 0z"
          clip-rule="evenodd"
        />
      </svg>
      <svg
        v-else
        class="size-5"
        :class="palette.dot"
        viewBox="0 0 20 20"
        fill="currentColor"
      >
        <path
          fill-rule="evenodd"
          d="M8.485 2.495c.673-1.167 2.357-1.167 3.03 0l6.28 10.88c.673 1.167-.17 2.625-1.516 2.625H3.72c-1.347 0-2.189-1.458-1.515-2.625L8.485 2.495zM10 5a1 1 0 011 1v3a1 1 0 11-2 0V6a1 1 0 011-1zm0 7a1 1 0 100 2 1 1 0 000-2z"
          clip-rule="evenodd"
        />
      </svg>
      <span>{{ verdict.label }}</span>
    </div>
    <p class="text-xs text-gray-500 font-mono" data-testid="updated-label">
      {{ updatedLabel }}
    </p>
  </section>
</template>
