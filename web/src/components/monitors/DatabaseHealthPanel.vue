<script setup lang="ts">
import { computed } from 'vue'
import type { DatabaseHealth } from '@/types'

interface Props {
  health: DatabaseHealth
}
const props = defineProps<Props>()

/**
 * Saturation needs both halves of the pair. When either is missing the ratio is
 * meaningless, so the row is not shown at all rather than shown as a bare count.
 */
const saturation = computed(() => {
  const { connections_active: active, connections_max: max } = props.health
  if (active == null || max == null || max <= 0) return null
  return { active, max, pct: Math.round((active / max) * 100) }
})

/** Only ever a hint of pressure — this panel is context, never an alert. */
const saturationTone = computed(() => {
  const s = saturation.value
  if (!s) return 'text-highlighted'
  if (s.pct >= 90) return 'text-error'
  if (s.pct >= 75) return 'text-warning'
  return 'text-highlighted'
})

function seconds(v: number): string {
  if (v < 1) return `${Math.round(v * 1000)} ms`
  if (v < 60) return `${v.toFixed(1)} s`
  return `${Math.round(v / 60)} min`
}

const collectedAt = computed(() => new Date(props.health.collected_at).toLocaleString())

/**
 * A missing grant and an unsupported server both leave fields empty, but they ask
 * the operator for different things: grant something, or upgrade something.
 */
const limitation = computed(() => {
  if (props.health.unsupported_version) {
    return {
      test: 'db-health-unsupported',
      text: 'Some figures need PostgreSQL 12+ or MySQL 8.0+. The monitor itself works on any version.',
    }
  }
  if (props.health.privilege_limited) {
    return {
      test: 'db-health-privilege',
      text: 'Grant pg_read_all_stats (PostgreSQL) or PROCESS (MySQL) to this credential for query age and replication lag. Entirely optional.',
    }
  }
  return null
})
</script>

<template>
  <div class="bg-default rounded-lg border border-default overflow-hidden">
    <div class="px-5 py-3 border-b border-default">
      <h3 class="text-sm font-semibold text-highlighted">Database health</h3>
    </div>

    <div class="px-5 py-3 space-y-3">
      <div v-if="saturation" data-test="db-health-connections">
        <div class="text-[10px] uppercase tracking-wider text-muted font-semibold mb-0.5">
          Connections
        </div>
        <div class="text-sm" :class="saturationTone">
          {{ saturation.active }} / {{ saturation.max }}
          <span class="text-muted">&middot; {{ saturation.pct }}%</span>
        </div>
      </div>

      <div v-if="health.longest_query_seconds != null" data-test="db-health-longest-query">
        <div class="text-[10px] uppercase tracking-wider text-muted font-semibold mb-0.5">
          Longest running query
        </div>
        <div class="text-sm text-highlighted">{{ seconds(health.longest_query_seconds) }}</div>
      </div>

      <div v-if="health.replication_lag_seconds != null" data-test="db-health-replication">
        <div class="text-[10px] uppercase tracking-wider text-muted font-semibold mb-0.5">
          Replication lag
        </div>
        <div class="text-sm text-highlighted">{{ seconds(health.replication_lag_seconds) }}</div>
      </div>
    </div>

    <div v-if="limitation" class="px-5 py-2 border-t border-default" :data-test="limitation.test">
      <p class="text-xs text-muted">{{ limitation.text }}</p>
    </div>

    <div class="px-5 py-2 border-t border-default text-xs text-muted" data-test="db-health-collected">
      Collected {{ collectedAt }}
    </div>
  </div>
</template>
