<script setup lang="ts">
/**
 * Escalation view — policies grid + reorder + stats.
 * Spec 059 US5 / FR-023..FR-026a / FR-039.
 */
import { computed, onMounted, ref } from 'vue'
import escalationService, { type EscalationPolicy } from '@/services/escalationService'
import { fetchChannels } from '@/services/notificationChannelService'
import type { NotificationChannel } from '@/types'
import { useConfirm } from '@/composables/useConfirm'
import type { EscalationPolicyInput } from '@/schemas/escalation-policy.schema'
import PolicyCard from '@/components/settings/escalation/PolicyCard.vue'
import PolicyModal from '@/components/settings/escalation/PolicyModal.vue'

const policies = ref<EscalationPolicy[]>([])
const channels = ref<NotificationChannel[]>([])
const loading = ref(true)
const modalOpen = ref(false)
const editing = ref<EscalationPolicy | null>(null)

const stats = computed(() => [
  { label: 'Policies', value: String(policies.value.length), tip: '' },
  { label: 'Resources covered', value: '—', tip: 'Backend metric pending' },
  { label: 'Escalated (30d)', value: '—', tip: 'Backend metric pending' },
  { label: 'Avg time to ack', value: '—', tip: 'Backend metric pending' },
])

let reorderTimer: ReturnType<typeof setTimeout> | null = null
const pendingReorder = ref(false)

function scheduleReorder() {
  if (reorderTimer) clearTimeout(reorderTimer)
  pendingReorder.value = true
  reorderTimer = setTimeout(async () => {
    try {
      const ids = policies.value.filter((p) => p.is_active).map((p) => p.id)
      const updated = await escalationService.reorder(ids)
      policies.value = updated
    } finally {
      pendingReorder.value = false
    }
  }, 500)
}

async function reload() {
  loading.value = true
  try {
    const [pol, ch] = await Promise.all([escalationService.list(), fetchChannels()])
    policies.value = pol
    channels.value = ch
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editing.value = null
  modalOpen.value = true
}

function openEdit(p: EscalationPolicy) {
  editing.value = p
  modalOpen.value = true
}

async function onSubmit(payload: EscalationPolicyInput) {
  if (editing.value) {
    await escalationService.update(editing.value.id, payload)
  } else {
    await escalationService.create(payload)
  }
  modalOpen.value = false
  await reload()
}

async function onDelete(p: EscalationPolicy) {
  const ok = await useConfirm({
    kind: 'destructive',
    title: `Delete policy "${p.name}"?`,
    body: 'Active incident escalations under this policy will continue, but no new escalations will fire.',
    ctaLabel: 'Delete policy',
  })
  if (!ok) return
  await escalationService.delete(p.id)
  policies.value = policies.value.filter((x) => x.id !== p.id)
}

async function onToggle(p: EscalationPolicy) {
  if (p.is_active) {
    const ok = await useConfirm({
      kind: 'default',
      title: `Disable "${p.name}"?`,
      body: 'Active incident escalations will continue under this policy. New incidents will not trigger this policy until re-enabled.',
      ctaLabel: 'Disable',
    })
    if (!ok) return
  }
  const next = await escalationService.update(p.id, {
    name: p.name,
    scope: p.scope,
    is_active: !p.is_active,
    steps: p.steps,
  })
  policies.value = policies.value.map((x) => (x.id === p.id ? next : x))
}

function moveUp(p: EscalationPolicy) {
  const i = policies.value.findIndex((x) => x.id === p.id)
  if (i <= 0) return
  const arr = [...policies.value]
  ;[arr[i - 1], arr[i]] = [arr[i]!, arr[i - 1]!]
  policies.value = arr
  scheduleReorder()
}

function moveDown(p: EscalationPolicy) {
  const i = policies.value.findIndex((x) => x.id === p.id)
  if (i < 0 || i >= policies.value.length - 1) return
  const arr = [...policies.value]
  ;[arr[i + 1], arr[i]] = [arr[i]!, arr[i + 1]!]
  policies.value = arr
  scheduleReorder()
}

const initialForModal = computed(() => {
  if (!editing.value) return undefined
  return {
    id: editing.value.id,
    name: editing.value.name,
    scope: editing.value.scope,
    is_active: editing.value.is_active,
    steps: editing.value.steps.map((s) => ({
      delay_minutes: s.delay_minutes,
      channel_ids: s.channel_ids,
    })),
  }
})

onMounted(reload)

defineExpose({
  policies,
  stats,
  pendingReorder,
  openCreate,
  onSubmit,
  onDelete,
  onToggle,
  moveUp,
  moveDown,
})
</script>

<template>
  <div class="space-y-6">
    <header class="flex items-start justify-between gap-4">
      <div>
        <h1 class="text-lg font-semibold text-default">Escalation policies</h1>
        <p class="text-sm text-muted">
          Rules that decide who gets paged, when, and via which channels.
        </p>
      </div>
      <UButton color="primary" icon="i-lucide-plus" @click="openCreate">New policy</UButton>
    </header>

    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3">
      <div
        v-for="s in stats"
        :key="s.label"
        class="rounded-xl border border-default/40 bg-default p-4 space-y-1"
      >
        <p class="text-xs text-muted uppercase tracking-wide">{{ s.label }}</p>
        <p class="text-xl font-semibold text-default">{{ s.value }}</p>
        <p v-if="s.tip" class="text-[10px] text-muted italic">{{ s.tip }}</p>
      </div>
    </div>

    <USkeleton v-if="loading" class="h-32 w-full" />

    <UEmpty
      v-else-if="policies.length === 0"
      icon="i-lucide-siren"
      title="No escalation policies yet"
      description="Create one to fan out alerts across teams and time."
    >
      <template #actions>
        <UButton color="primary" @click="openCreate">Create your first policy</UButton>
      </template>
    </UEmpty>

    <div v-else class="space-y-3">
      <div v-if="pendingReorder" class="text-xs text-muted flex items-center gap-1.5">
        <UIcon name="i-lucide-loader-2" class="size-3 animate-spin" />
        Saving new order…
      </div>
      <PolicyCard
        v-for="(p, i) in policies"
        :key="p.id"
        :policy="p"
        :can-move-up="i > 0"
        :can-move-down="i < policies.length - 1"
        @toggle="onToggle"
        @edit="openEdit"
        @delete="onDelete"
        @move-up="moveUp"
        @move-down="moveDown"
      />
    </div>

    <PolicyModal
      v-model:open="modalOpen"
      :initial="initialForModal as unknown as never"
      :channels="channels"
      @submit="onSubmit"
    />
  </div>
</template>
