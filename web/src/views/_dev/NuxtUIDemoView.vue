<script setup lang="ts">
import { ref } from 'vue'
import { useToast } from '@nuxt/ui/composables/useToast'
import { useConfirm } from '@/composables/useConfirm'
import { useLicence } from '@/composables/useLicence'

const toast = useToast()
const { isEnterprise: isEE } = useLicence()

const inputValue = ref('')
const dateValue = ref<Date | null>(null)
const confirmResult = ref<string | null>(null)
const removedFilter = ref<string | null>(null)

function showToast() {
  toast.add({
    title: 'NuxtUI is wired',
    description: 'Tokens, plugin, toast composable all reachable.',
    color: 'success',
  })
}

async function tryConfirm() {
  const ok = await useConfirm({
    kind: 'destructive',
    title: 'Delete this monitor?',
    body: 'api.acme.com will stop being checked immediately.',
    ctaLabel: 'Delete',
  })
  confirmResult.value = ok ? 'confirmed' : 'dismissed'
}

const sampleDays90 = Array.from({ length: 90 }, (_, i) => {
  const r = (i * 31 + 7) % 10
  if (r === 0) return 'down' as const
  if (r === 1) return 'warning' as const
  if (r === 2) return 'nodata' as const
  return 'up' as const
})

const sampleDays31 = Array.from({ length: 31 }, (_, i) => {
  const r = (i * 13 + 3) % 9
  if (r === 0) return 'warning' as const
  if (r === 1) return 'down' as const
  return 'up' as const
})

const sampleSparkline = [10, 14, 12, 18, 22, 19, 25, 23, 28, 30, 27, 32]
</script>

<template>
  <div class="min-h-screen p-8 bg-default text-default font-sans">
    <header class="mb-8">
      <h1 class="text-2xl font-semibold">Shared components demo</h1>
      <p class="text-muted text-sm mt-1">
        Dev-only · Spec 053 + 055 · removed at Slice 6 (PRD 009).
      </p>
    </header>

    <section class="space-y-10 max-w-5xl">
      <!-- Foundation primitives (PR-1) -->
      <div>
        <h2 class="text-lg font-semibold mb-3">Foundation primitives (PR-1)</h2>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <UCard>
            <div class="flex flex-col gap-2">
              <UButton color="primary" @click="showToast">Trigger toast</UButton>
              <UIcon name="i-lucide-bell" class="size-6 text-primary-500" />
            </div>
          </UCard>
          <UCard>
            <UInput v-model="inputValue" placeholder="Type here..." />
          </UCard>
          <UCard>
            <UDatePicker v-model="dateValue" placeholder="Pick a date" />
          </UCard>
        </div>
      </div>

      <!-- US2 — shared library (PR-3) -->
      <div>
        <h2 class="text-lg font-semibold mb-3">Shared component library (PR-3)</h2>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <UCard>
            <p class="text-sm font-medium mb-2">UStatusBadge</p>
            <div class="flex flex-wrap gap-2">
              <UStatusBadge status="up" dot />
              <UStatusBadge status="down" dot />
              <UStatusBadge status="warning" dot />
              <UStatusBadge status="maintenance" dot />
              <UStatusBadge status="unknown" dot />
            </div>
          </UCard>

          <UCard>
            <p class="text-sm font-medium mb-2">UEditionBadge</p>
            <div class="flex gap-2">
              <UEditionBadge edition="ce" />
              <UEditionBadge edition="ee" />
            </div>
          </UCard>

          <UCard>
            <p class="text-sm font-medium mb-2">EE-gating pattern (live)</p>
            <UButton
              :disabled="!isEE"
              :ui="{ tooltip: !isEE ? 'Available on Enterprise — Upgrade' : undefined }"
              data-test="ee-gated-action"
            >
              Add team member
              <UEditionBadge v-if="!isEE" edition="ee" />
            </UButton>
            <p class="text-xs text-muted mt-2">
              Edition: <code class="font-mono">{{ isEE ? 'enterprise' : 'community' }}</code>
              · documented at docs/frontend/ee-gating.md
            </p>
          </UCard>

          <UCard>
            <p class="text-sm font-medium mb-2">UFilterChip</p>
            <div class="flex flex-wrap gap-2">
              <UFilterChip kind="tag" value="production" @remove="removedFilter = 'tag:production'" />
              <UFilterChip kind="component" value="api-cluster" @remove="removedFilter = 'component:api-cluster'" />
              <UFilterChip kind="status" value="down" @remove="removedFilter = 'status:down'" />
            </div>
            <p v-if="removedFilter" class="text-xs text-muted mt-2">
              Removed: <code>{{ removedFilter }}</code>
            </p>
          </UCard>

          <UCard>
            <p class="text-sm font-medium mb-2">UKbd</p>
            <div class="flex gap-3">
              <UKbd :keys="['⌘', 'K']" />
              <UKbd :keys="['⌘', '?']" />
              <UKbd :keys="['G', 'I']" />
            </div>
          </UCard>

          <UCard>
            <p class="text-sm font-medium mb-2">USkeleton</p>
            <div class="space-y-2">
              <USkeleton variant="text" />
              <div class="flex items-center gap-2">
                <USkeleton variant="circle" />
                <USkeleton variant="text" width="40%" />
              </div>
              <USkeleton variant="rect" height="40px" />
              <USkeleton variant="table-row" />
              <USkeleton variant="card" />
            </div>
          </UCard>

          <UCard>
            <p class="text-sm font-medium mb-2">UStepper</p>
            <div class="space-y-3">
              <UStepper :steps="['Profile', 'Verify', 'Done']" :active-step="1" />
              <UStepper :steps="['One', 'Two', 'Three']" :active-step="0" variant="dots" />
            </div>
          </UCard>

          <UCard>
            <p class="text-sm font-medium mb-2">UStatCard</p>
            <UStatCard
              label="Monitors"
              :value="14"
              subtitle="3 down"
              icon="i-lucide-radar"
              :sparkline="sampleSparkline"
            />
          </UCard>

          <UCard>
            <p class="text-sm font-medium mb-2">UConfirmModal (via useConfirm)</p>
            <UButton color="error" variant="soft" @click="tryConfirm">
              Open confirm
            </UButton>
            <p v-if="confirmResult" class="text-xs text-muted mt-2">
              Last result: <code>{{ confirmResult }}</code>
            </p>
          </UCard>

          <UCard>
            <p class="text-sm font-medium mb-2">UEmptyState</p>
            <UEmptyState
              icon="i-lucide-radar"
              title="No monitors yet"
              description="Add your first monitor to start receiving alerts."
            >
              <template #actions>
                <UButton size="sm">Add monitor</UButton>
              </template>
            </UEmptyState>
          </UCard>
        </div>

        <UCard class="mt-4">
          <p class="text-sm font-medium mb-2">UUptimeBar (90 days)</p>
          <UUptimeBar :days="sampleDays90" />
        </UCard>

        <UCard class="mt-4">
          <p class="text-sm font-medium mb-2">UUptimeCalendar</p>
          <div class="max-w-xs">
            <UUptimeCalendar :month="5" :year="2026" :days="sampleDays31" :uptime-pct="98.21" />
          </div>
        </UCard>
      </div>
    </section>

    <footer class="mt-10 text-xs text-muted">
      Strict isolation enforced for UStatusBadge / UUptimeBar / UUptimeCalendar (spec 055 Q2).
    </footer>
  </div>
</template>
