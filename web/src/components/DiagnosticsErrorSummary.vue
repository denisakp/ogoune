<script setup lang="ts">
interface Props {
  errorSummary?: string | null
  failureType?: string | null
  errorMessage?: string | null
  rootCauseHint?: string | null
}

const props = withDefaults(defineProps<Props>(), {
  errorSummary: null,
  failureType: null,
  errorMessage: null,
  rootCauseHint: null,
})

const getFailureIcon = (
  failureType: string | null | undefined,
  hint: string | null | undefined,
) => {
  if (hint === 'host_unreachable') return { icon: 'i-lucide-unplug', color: '#ff4d4f' }
  if (hint === 'service_down') return { icon: 'i-lucide-triangle-alert', color: '#faad14' }
  if (hint === 'icmp_unavailable') return { icon: 'i-lucide-circle-alert', color: '#8c8c8c' }

  if (!failureType) return { icon: 'i-lucide-circle-alert', color: '#ff4d4f' }

  const type = failureType.toLowerCase()
  if (type.includes('timeout')) return { icon: 'i-lucide-clock', color: '#faad14' }
  if (type.includes('connection') || type.includes('refused'))
    return { icon: 'i-lucide-unplug', color: '#ff4d4f' }
  if (type.includes('dns')) return { icon: 'i-lucide-globe', color: '#fa541c' }
  if (type.includes('ssl') || type.includes('certificate'))
    return { icon: 'i-lucide-lock', color: '#fa541c' }
  if (type.includes('http') || type.includes('status'))
    return { icon: 'i-lucide-triangle-alert', color: '#faad14' }
  return { icon: 'i-lucide-circle-alert', color: '#ff4d4f' }
}

const failureIcon = getFailureIcon(props.failureType, props.rootCauseHint)
</script>

<template>
  <div>
    <div class="flex gap-3 items-start">
      <div class="shrink-0 mt-0.5">
        <UIcon :name="failureIcon.icon" :style="{ color: failureIcon.color }" class="size-5" />
      </div>

      <div class="flex-1">
        <div v-if="errorSummary" class="text-sm font-medium mb-3 leading-relaxed text-default">
          {{ errorSummary }}
        </div>

        <div v-if="failureType" class="text-xs text-muted mb-2">
          <strong>Type:</strong> {{ failureType }}
        </div>

        <div
          v-if="errorMessage"
          class="text-xs text-muted font-mono p-2 bg-slate-50 dark:bg-slate-900 border-l-4 border-red-500 rounded-sm"
        >
          {{ errorMessage }}
        </div>
      </div>
    </div>
  </div>
</template>
