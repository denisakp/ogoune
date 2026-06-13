<script setup lang="ts">
import { computed, ref } from 'vue'
import type { IncidentDiagnostics } from '@/types'

interface Props {
  diagnostics?: IncidentDiagnostics | null
}
const props = defineProps<Props>()

interface Row {
  label: string
  value: string
  mono?: boolean
}

const TRUNCATE_AT = 5 * 1024
const showFullBody = ref(false)

const d = computed(() => props.diagnostics)

function nonNull(rows: Array<Row | null>): Row[] {
  return rows.filter((r): r is Row => r !== null)
}

const headerRows = computed<Row[]>(() => {
  const r = d.value
  if (!r) return []
  return nonNull([
    r.failure_type ? { label: 'Cause', value: r.failure_type, mono: true } : null,
    r.error_summary || r.error_message
      ? { label: 'Error', value: r.error_summary || r.error_message }
      : null,
    r.root_cause_hint ? { label: 'Root cause hint', value: r.root_cause_hint } : null,
  ])
})

const requestRows = computed<Row[]>(() => {
  const r = d.value
  if (!r) return []
  return nonNull([
    r.request_method && r.request_url
      ? { label: 'Request', value: `${r.request_method} ${r.request_url}`, mono: true }
      : null,
    r.http_status_code && r.http_status_code > 0
      ? { label: 'HTTP Status', value: String(r.http_status_code), mono: true }
      : null,
    r.request_timeout ? { label: 'Timeout', value: `${r.request_timeout}s`, mono: true } : null,
  ])
})

const timingRows = computed<Row[]>(() => {
  const r = d.value
  if (!r) return []
  return nonNull([
    r.total_duration ? { label: 'Total', value: `${r.total_duration} ms`, mono: true } : null,
    r.dns_duration ? { label: 'DNS', value: `${r.dns_duration} ms`, mono: true } : null,
    r.tls_duration ? { label: 'TLS', value: `${r.tls_duration} ms`, mono: true } : null,
    r.first_byte_duration
      ? { label: 'TTFB', value: `${r.first_byte_duration} ms`, mono: true }
      : null,
  ])
})

const icmpRows = computed<Row[]>(() => {
  const r = d.value
  if (!r || r.icmp_available == null) return []
  return nonNull([
    { label: 'ICMP available', value: r.icmp_available ? 'Yes' : 'No' },
    {
      label: 'ICMP reachable',
      value: r.icmp_reachable == null ? '—' : r.icmp_reachable ? 'Yes' : 'No',
    },
    r.icmp_rtt_ms != null ? { label: 'ICMP RTT', value: `${r.icmp_rtt_ms} ms`, mono: true } : null,
  ])
})

const keywordRows = computed<Row[]>(() => {
  const r = d.value
  if (!r || !r.keyword) return []
  return nonNull([
    { label: 'Keyword', value: r.keyword, mono: true },
    r.keyword_mode ? { label: 'Mode', value: r.keyword_mode } : null,
    {
      label: 'Found',
      value: r.keyword_found == null ? '—' : r.keyword_found ? 'Yes' : 'No',
    },
  ])
})

const responseBody = computed(() => d.value?.response_body ?? '')
const bodyTooLong = computed(() => responseBody.value.length > TRUNCATE_AT)
const visibleBody = computed(() => {
  if (!bodyTooLong.value || showFullBody.value) return responseBody.value
  return `${responseBody.value.slice(0, TRUNCATE_AT)}\n... (${(responseBody.value.length / 1024).toFixed(1)} KB total)`
})

const impactText = computed(() => d.value?.error_summary || d.value?.root_cause_hint || '')
const hasImpact = computed(() => !!impactText.value)
</script>

<template>
  <div class="space-y-4">
    <div class="bg-default rounded-lg border border-default overflow-hidden">
      <div class="px-5 py-3 border-b border-default">
        <h3 class="text-sm font-semibold text-highlighted">Diagnostics</h3>
      </div>
      <div v-if="!d" class="px-5 py-6 text-sm text-muted text-center">
        No diagnostics available.
      </div>
      <div v-else class="divide-y divide-slate-100">
        <div v-if="headerRows.length" class="px-5 py-3 space-y-3">
          <div v-for="row in headerRows" :key="row.label">
            <div class="text-[10px] uppercase tracking-wider text-muted font-semibold mb-0.5">
              {{ row.label }}
            </div>
            <div :class="row.mono ? 'font-mono text-xs text-highlighted' : 'text-sm text-default'">
              {{ row.value }}
            </div>
          </div>
        </div>

        <div v-if="requestRows.length" class="px-5 py-3 space-y-3">
          <div v-for="row in requestRows" :key="row.label">
            <div class="text-[10px] uppercase tracking-wider text-muted font-semibold mb-0.5">
              {{ row.label }}
            </div>
            <div
              :class="
                row.mono ? 'font-mono text-xs text-highlighted break-all' : 'text-sm text-default'
              "
            >
              {{ row.value }}
            </div>
          </div>
        </div>

        <div v-if="timingRows.length" class="px-5 py-3">
          <div class="text-[10px] uppercase tracking-wider text-muted font-semibold mb-2">
            Timing breakdown
          </div>
          <dl class="grid grid-cols-2 gap-y-1.5 gap-x-3 text-xs">
            <template v-for="row in timingRows" :key="row.label">
              <dt class="text-muted">{{ row.label }}</dt>
              <dd class="font-mono text-highlighted text-right">{{ row.value }}</dd>
            </template>
          </dl>
        </div>

        <div v-if="icmpRows.length" class="px-5 py-3">
          <div class="text-[10px] uppercase tracking-wider text-muted font-semibold mb-2">
            ICMP probe
          </div>
          <dl class="grid grid-cols-2 gap-y-1.5 gap-x-3 text-xs">
            <template v-for="row in icmpRows" :key="row.label">
              <dt class="text-muted">{{ row.label }}</dt>
              <dd class="text-highlighted text-right" :class="row.mono ? 'font-mono' : ''">
                {{ row.value }}
              </dd>
            </template>
          </dl>
        </div>

        <div v-if="keywordRows.length" class="px-5 py-3">
          <div class="text-[10px] uppercase tracking-wider text-muted font-semibold mb-2">
            Keyword check
          </div>
          <dl class="grid grid-cols-2 gap-y-1.5 gap-x-3 text-xs">
            <template v-for="row in keywordRows" :key="row.label">
              <dt class="text-muted">{{ row.label }}</dt>
              <dd class="text-highlighted text-right" :class="row.mono ? 'font-mono' : ''">
                {{ row.value }}
              </dd>
            </template>
          </dl>
        </div>

        <div v-if="responseBody" class="px-5 py-3">
          <div class="flex items-center justify-between mb-2">
            <div class="text-[10px] uppercase tracking-wider text-muted font-semibold">
              Response body
            </div>
            <button
              v-if="bodyTooLong"
              type="button"
              class="text-[11px] text-primary-600 hover:underline"
              @click="showFullBody = !showFullBody"
            >
              {{ showFullBody ? 'Show less' : 'Show full' }}
            </button>
          </div>
          <pre
            class="bg-muted rounded p-2 text-[11px] font-mono text-default whitespace-pre overflow-x-auto max-h-48"
            >{{ visibleBody }}</pre
          >
        </div>
      </div>
    </div>

    <div
      v-if="hasImpact"
      class="rounded-lg border p-4 flex gap-3"
      style="background-color: #fef2f2; border-color: #fca5a5"
    >
      <UIcon name="i-lucide-alert-triangle" class="size-4 shrink-0 mt-0.5" style="color: #b91c1c" />
      <div class="flex-1 min-w-0">
        <div class="text-sm font-semibold mb-1" style="color: #b91c1c">Impact</div>
        <p class="text-xs text-default">{{ impactText }}</p>
      </div>
    </div>
  </div>
</template>
