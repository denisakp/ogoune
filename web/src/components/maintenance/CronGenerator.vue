<script setup lang="ts">
/* eslint-disable @typescript-eslint/ban-ts-comment */
// @ts-nocheck — spec 059 polish debt: NuxtUI v4 template-emit types
/**
 * Cron generator — input + human-readable preview + next 5 occurrences.
 * Spec 059 US7 / T101. NuxtUI-only, no AntDV.
 * Backed by `cronstrue` (preview) + `cron-parser` (next occurrences).
 */
import { computed, ref, watch } from 'vue'
import cronstrue from 'cronstrue'
import { CronExpressionParser } from 'cron-parser'

interface Props {
  modelValue: string
  durationMinutes?: number
}
const props = withDefaults(defineProps<Props>(), {
  durationMinutes: 60,
})
const emit = defineEmits<{
  (e: 'update:modelValue', v: string): void
  (e: 'update:valid', v: boolean): void
}>()

const expr = ref<string>(props.modelValue || '*/15 * * * *')

watch(
  () => props.modelValue,
  (v) => {
    if (v && v !== expr.value) expr.value = v
  },
)

const humanReadable = computed<string>(() => {
  try {
    return cronstrue.toString(expr.value, { use24HourTimeFormat: true })
  } catch {
    return 'Invalid cron expression'
  }
})

const isValid = computed<boolean>(() => {
  try {
    CronExpressionParser.parse(expr.value)
    return true
  } catch {
    return false
  }
})

const nextOccurrences = computed<string[]>(() => {
  try {
    const iter = CronExpressionParser.parse(expr.value, { tz: 'UTC' })
    const out: string[] = []
    for (let i = 0; i < 5; i++) out.push(iter.next().toISOString())
    return out
  } catch {
    return []
  }
})

function onInput(v: string | number) {
  const next = String(v).trim()
  expr.value = next
  emit('update:modelValue', next)
  emit('update:valid', isValid.value)
}

const PRESETS = [
  { label: 'Every 15 min', value: '*/15 * * * *' },
  { label: 'Hourly', value: '0 * * * *' },
  { label: 'Daily 2 AM', value: '0 2 * * *' },
  { label: 'Weekly Sun 2 AM', value: '0 2 * * 0' },
]

defineExpose({ expr, humanReadable, isValid, nextOccurrences, onInput })
</script>

<template>
  <div class="space-y-3">
    <UFormField label="Cron expression" :error="!isValid ? 'Invalid cron expression' : undefined">
      <UInput :model-value="expr" placeholder="*/15 * * * *" @update:model-value="onInput" />
    </UFormField>

    <div class="flex flex-wrap gap-2">
      <UButton
        v-for="p in PRESETS"
        :key="p.value"
        size="xs"
        variant="outline"
        @click="onInput(p.value)"
      >
        {{ p.label }}
      </UButton>
    </div>

    <div
      class="rounded-md border border-default/40 bg-elevated px-3 py-2 text-sm"
      :class="isValid ? 'text-default' : 'text-error'"
    >
      {{ humanReadable }}
    </div>

    <div v-if="isValid && nextOccurrences.length > 0">
      <p class="text-xs font-semibold uppercase tracking-wide text-muted mb-1">
        Next 5 occurrences (UTC)
      </p>
      <ul class="text-xs font-mono text-default space-y-0.5">
        <li v-for="o in nextOccurrences" :key="o">{{ o }}</li>
      </ul>
    </div>
  </div>
</template>
