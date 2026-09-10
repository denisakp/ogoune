<script setup lang="ts">
import { computed } from 'vue'
import type { IncidentExplanation } from '@/types'

interface Props {
  explanation: IncidentExplanation | null | undefined
}
const props = defineProps<Props>()

/**
 * The causal narrative (spec 091). One sentence, produced by the API so the
 * wording exists in exactly one place, plus the two timestamps that let an
 * operator overrule it.
 *
 * Deliberately restrained: no confidence badge, no percentage, no "likely
 * cause" label. The sentence states that two things happened close together on
 * the same host. Decorating it with anything that looks like a score would
 * claim a certainty the data does not carry -- and operators calibrate on what
 * a monitoring tool tells them.
 *
 * Renders nothing at all when there is no explanation. Absence is silence, not
 * a panel saying nothing was found.
 */
const e = computed(() => props.explanation ?? null)

function when(iso: string): string {
  return new Date(iso).toLocaleString()
}

/** "13 seconds earlier" is already in the sentence; this labels the direction. */
const direction = computed(() => (e.value?.precedes ? 'before the failure' : 'after the failure'))
</script>

<template>
  <div
    v-if="e"
    class="bg-elevated rounded-lg border border-default p-5"
    data-test="incident-explanation"
  >
    <div class="flex items-start gap-3">
      <UIcon name="i-lucide-link" class="size-5 mt-0.5 text-muted shrink-0" />
      <div class="min-w-0">
        <h3 class="text-sm font-semibold text-highlighted mb-1">What happened around it</h3>
        <p class="text-sm text-default" data-test="incident-explanation-text">{{ e.text }}</p>

        <!-- Both times, always. They are what make the sentence checkable
             rather than something to be taken on trust. -->
        <dl class="mt-3 grid grid-cols-1 sm:grid-cols-2 gap-x-6 gap-y-1 text-xs">
          <div class="flex gap-2">
            <dt class="text-muted shrink-0">Check failed</dt>
            <dd class="text-default" data-test="incident-explanation-incident-at">
              {{ when(e.incident_at) }}
            </dd>
          </div>
          <div class="flex gap-2">
            <dt class="text-muted shrink-0">Kernel reported</dt>
            <dd class="text-default" data-test="incident-explanation-event-at">
              {{ when(e.event.occurredAt) }}
              <span class="text-muted">({{ direction }})</span>
            </dd>
          </div>
          <div class="flex gap-2">
            <dt class="text-muted shrink-0">Host</dt>
            <dd class="text-default">{{ e.host_name || e.host_id }}</dd>
          </div>
          <!-- "37 kills" and "1 kill" are different findings; so are "one event"
               and "one of four". Neither count is hidden behind a plural. -->
          <div v-if="e.other_events > 0" class="flex gap-2" data-test="incident-explanation-others">
            <dt class="text-muted shrink-0">Also in the window</dt>
            <dd class="text-default">
              {{ e.other_events }} other kernel
              {{ e.other_events === 1 ? 'event' : 'events' }}
            </dd>
          </div>
        </dl>

        <p class="mt-3 text-xs text-muted">
          Shown because both happened on the same host within the same window. That is a
          co-occurrence, not a proven cause.
        </p>
      </div>
    </div>
  </div>
</template>
