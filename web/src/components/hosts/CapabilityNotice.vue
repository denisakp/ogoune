<script setup lang="ts">
import { computed } from 'vue'
import type { Capability, HostCapabilities } from '@/types'

/**
 * What an agent can observe on its machine, in words a reader can act on
 * (spec 093). The backend ships facts -- three capabilities, a reason, a state;
 * the sentence is made here.
 *
 * Two surfaces, one component:
 *  - `host`: three rows on the host page, current declaration.
 *  - `incident`: one sentence explaining an ABSENCE of kernel events, from the
 *    declaration frozen when the incident opened. Only meaningful when host
 *    context exists and no events do; the caller decides that.
 *
 * The rule behind every branch: unknown is never rendered as unavailable.
 */
interface Props {
  capabilities: HostCapabilities
  variant: 'host' | 'incident'
}
const props = defineProps<Props>()

const SYSCTL = 'sysctl -w debug.exception-trace=1'

interface Row {
  key: 'kmsg' | 'oom' | 'segfault'
  label: string
  available: boolean
  text: string
  /** A one-line action, only where there is one. */
  action?: string
}

const isDeclared = computed(() => props.capabilities.state === 'declared')

function kmsgText(c: Capability | undefined): string {
  if (c?.available) return 'Readable'
  if (c?.reason === 'platform') return 'Not supported on this platform'
  return 'Not readable in this install (containerised agents usually cannot read it)'
}

function oomText(detail: HostCapabilities['oom_detail']): string {
  switch (detail) {
    case 'with_process':
      return 'Detected, with the process name'
    case 'without_process':
      return 'Detected, without the process name'
    default:
      return 'Not detected'
  }
}

function segfaultText(c: Capability | undefined): { text: string; action?: string } {
  if (c?.available) return { text: 'Available' }
  switch (c?.reason) {
    case 'setting_off':
      return {
        text: 'Unavailable — the kernel is not reporting userspace faults. Best-effort signal.',
        action: SYSCTL,
      }
    case 'platform':
      return { text: 'Not supported on this platform' }
    default:
      return { text: 'Unavailable — needs the kernel log' }
  }
}

const rows = computed<Row[]>(() => {
  if (!isDeclared.value) return []
  const c = props.capabilities
  const seg = segfaultText(c.segfault)
  return [
    { key: 'kmsg', label: 'Kernel log', available: !!c.kmsg?.available, text: kmsgText(c.kmsg) },
    {
      key: 'oom',
      label: 'Out-of-memory kills',
      available: c.oom_detail === 'with_process' || c.oom_detail === 'without_process',
      text: oomText(c.oom_detail),
    },
    { key: 'segfault', label: 'Segfault capture', available: !!c.segfault?.available, ...seg },
  ]
})

/** Host page wording for the two not-declared states. */
const hostNotice = computed<{ title: string; description: string } | null>(() => {
  switch (props.capabilities.state) {
    case 'not_reported':
      return {
        title: 'This agent version does not report what it can observe',
        description: 'Upgrade the agent to see which kernel signals this host can capture.',
      }
    case 'not_known':
      return { title: 'Not yet known', description: 'The agent has not connected.' }
    default:
      return null
  }
})

/** Incident page: one sentence about what "no kernel events" meant AT THE TIME. */
const incidentSentence = computed<string | null>(() => {
  const c = props.capabilities
  switch (c.state) {
    case 'not_reported':
      return 'No kernel events in the window. This agent version did not report what it could observe.'
    case 'not_known':
      return 'No kernel events in the window. What this host could observe at the time is not known.'
    case 'no_machine':
      return null
    case 'declared':
      break
  }
  const kmsg = !!c.kmsg?.available
  const oomAny = c.oom_detail === 'with_process' || c.oom_detail === 'without_process'
  const seg = c.segfault
  if (!kmsg && oomAny) {
    return 'No kernel events in the window. At the time, this agent could not read the kernel log, so only cgroup out-of-memory kills could have been reported — and none were.'
  }
  if (!oomAny) {
    return 'No kernel events in the window — this agent could not observe any.'
  }
  if (seg?.available) {
    return 'No kernel events in the window. This agent could observe out-of-memory kills and segfaults; none were reported.'
  }
  if (seg?.reason === 'setting_off') {
    return 'No kernel events in the window. Out-of-memory kills were observable; segfault capture was off, so a segfault would not have been seen.'
  }
  return 'No kernel events in the window. Out-of-memory kills were observable; segfault capture was unavailable, so a segfault would not have been seen.'
})
</script>

<template>
  <!-- Incident: one sentence, or nothing. -->
  <p
    v-if="variant === 'incident' && incidentSentence"
    class="text-sm text-muted"
    data-test="capability-notice-incident"
    :data-state="capabilities.state"
  >
    {{ incidentSentence }}
  </p>

  <div
    v-else-if="variant === 'host'"
    class="bg-default rounded-lg border border-default overflow-hidden"
    data-test="capability-notice-host"
    :data-state="capabilities.state"
  >
    <div class="px-5 py-3 border-b border-default">
      <h3 class="text-sm font-semibold text-highlighted">What this agent can observe</h3>
    </div>

    <!-- Not declared: said as what it is, never as three unavailables. -->
    <UAlert
      v-if="hostNotice"
      :color="capabilities.state === 'not_reported' ? 'warning' : 'neutral'"
      variant="soft"
      :icon="capabilities.state === 'not_reported' ? 'i-lucide-arrow-up-circle' : 'i-lucide-help-circle'"
      class="rounded-none"
      :data-test="`capability-notice-${capabilities.state}`"
      :title="hostNotice.title"
      :description="hostNotice.description"
    />

    <ul v-if="rows.length" class="divide-y divide-default">
      <li
        v-for="r in rows"
        :key="r.key"
        class="px-5 py-3 flex items-start justify-between gap-3"
        :data-test="`capability-row-${r.key}`"
      >
        <div class="min-w-0">
          <div class="text-sm text-highlighted">{{ r.label }}</div>
          <div class="text-xs text-muted" data-test="capability-row-text">{{ r.text }}</div>
          <code
            v-if="r.action"
            class="mt-1 inline-block text-xs font-mono text-highlighted bg-elevated rounded px-1.5 py-0.5"
            data-test="capability-row-action"
            >{{ r.action }}</code
          >
        </div>
        <UBadge
          :color="r.available ? 'success' : 'neutral'"
          variant="subtle"
          size="sm"
          class="shrink-0"
          data-test="capability-row-badge"
        >
          {{ r.available ? 'available' : 'unavailable' }}
        </UBadge>
      </li>
    </ul>
    <p
      v-if="isDeclared && capabilities.declared_at"
      class="px-5 py-2 text-[11px] text-muted border-t border-default"
      data-test="capability-declared-at"
    >
      As reported by the agent at {{ new Date(capabilities.declared_at).toLocaleString() }}.
    </p>
  </div>
</template>
