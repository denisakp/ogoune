<script setup lang="ts">
import { computed } from 'vue'
import type { PublicVerdict } from '@/types'

const props = defineProps<{
  verdict: PublicVerdict
  secondsAgo: number | null
}>()

const palette = computed(() => {
  switch (props.verdict.status) {
    case 'operational':
      return 'bg-emerald-500 text-white'
    case 'partial_degradation':
      return 'bg-orange-500 text-white'
    case 'major_outage':
      return 'bg-red-600 text-white'
    default:
      return 'bg-gray-400 text-white'
  }
})

const updatedLabel = computed(() => {
  if (props.secondsAgo === null) return 'Updated just now'
  if (props.secondsAgo < 5) return 'Updated just now'
  if (props.secondsAgo < 60) return `Updated ${props.secondsAgo}s ago`
  const minutes = Math.floor(props.secondsAgo / 60)
  return `Updated ${minutes}m ago`
})
</script>

<template>
  <div
    :class="['rounded-xl px-6 py-5 flex items-center justify-between gap-4', palette]"
    role="status"
    :data-status="verdict.status"
  >
    <div class="flex items-center gap-3">
      <span class="inline-block size-3 rounded-full bg-white/90" />
      <h1 class="text-xl font-semibold leading-none">{{ verdict.label }}</h1>
    </div>
    <span class="text-sm/none opacity-90 font-mono" data-testid="updated-label">{{ updatedLabel }}</span>
  </div>
</template>
