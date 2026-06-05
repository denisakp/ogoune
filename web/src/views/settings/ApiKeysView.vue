<script setup lang="ts">
/* eslint-disable @typescript-eslint/ban-ts-comment */
// @ts-nocheck — spec 059 polish debt: NuxtUI v4 template-emit types
/**
 * API keys view — design fidelity v2.
 * Header + reveal banner (green) + 4 KPI cards + table.
 */
import { computed, onMounted, ref } from 'vue'
import accountService, { type APIKey } from '@/services/accountService'
import { useApiKeyStore } from '@/stores/useApiKeyStore'
import { useConfirm } from '@/composables/useConfirm'
import { resolveExpiresAt, type ApiKeyInput } from '@/schemas/api-key.schema'
import CreateKeyModal from '@/components/settings/apikeys/CreateKeyModal.vue'
import OneShotRevealBanner from '@/components/settings/apikeys/OneShotRevealBanner.vue'

const store = useApiKeyStore()
const keys = ref<APIKey[]>([])
const loading = ref(true)
const modalOpen = ref(false)

const stats = computed(() => {
  const total = keys.value.length
  const active = keys.value.filter((k) => k.is_active).length
  const revoked = total - active
  const rw = keys.value.filter((k) => k.scope === 'read_write').length
  const ro = keys.value.filter((k) => k.scope === 'read').length

  return [
    {
      key: 'total',
      label: 'TOTAL KEYS',
      value: String(total),
      meta: total > 0 ? `${active} active, ${revoked} revoked` : 'no keys yet',
    },
    {
      key: 'rw',
      label: 'READ_WRITE',
      value: String(rw),
      meta: 'elevated scope — audit regularly',
    },
    {
      key: 'r',
      label: 'READ',
      value: String(ro),
      meta: 'safe for dashboards',
    },
    {
      key: 'requests',
      label: 'REQUESTS (30d)',
      value: '—',
      meta: 'Backend metric pending',
    },
  ]
})

function relativeTime(iso: string | null | undefined): string {
  if (!iso) return '—'
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return '—'
  const diff = Date.now() - d.getTime()
  const m = Math.floor(diff / 60_000)
  if (m < 1) return 'just now'
  if (m < 60) return `${m}m ago`
  const h = Math.floor(m / 60)
  if (h < 24) return `${h}h ago`
  const days = Math.floor(h / 24)
  if (days < 30) return `${days}d ago`
  const months = Math.floor(days / 30)
  if (months < 12) return `${months}mo ago`
  const years = Math.floor(months / 12)
  return `${years}y ago`
}

function expiresLabel(k: APIKey): { label: string; tone: 'default' | 'warn' | 'muted' } {
  if (!k.is_active) return { label: 'Revoked', tone: 'muted' }
  if (!k.expires_at) return { label: 'Never', tone: 'default' }
  const exp = new Date(k.expires_at).getTime()
  const diff = exp - Date.now()
  if (diff < 0) return { label: 'Expired', tone: 'warn' }
  const days = Math.floor(diff / 86_400_000)
  if (days < 30) return { label: `in ${days} days`, tone: 'warn' }
  const months = Math.floor(days / 30)
  if (months < 12) return { label: `in ${months} months`, tone: 'default' }
  return { label: `in ${Math.floor(months / 12)}y`, tone: 'default' }
}

function maskedPrefix(k: APIKey): string {
  return `${k.key_prefix}…${k.id.slice(-4)}`
}

async function reload() {
  loading.value = true
  try {
    keys.value = await accountService.listAPIKeys()
  } finally {
    loading.value = false
  }
}

async function onSubmit(payload: ApiKeyInput) {
  const expires = resolveExpiresAt(payload)
  const created = await accountService.createAPIKey({
    name: payload.name,
    scope: payload.scope,
    expires_at: expires,
  })
  store.set(created)
  modalOpen.value = false
  await reload()
}

async function onRevoke(k: APIKey) {
  const ok = await useConfirm({
    kind: 'destructive',
    title: `Revoke API key "${k.name}"?`,
    body: `This key will continue to work for ~60 seconds due to backend caching.`,
    ctaLabel: 'Revoke key',
  })
  if (!ok) return
  await accountService.revokeAPIKey(k.id)
  keys.value = keys.value.filter((x) => x.id !== k.id)
}

function dismissBanner() {
  store.clear()
}

onMounted(reload)

defineExpose({ keys, stats, store, onSubmit, onRevoke, dismissBanner })
</script>

<template>
  <div class="space-y-6">
    <header class="flex items-start justify-between gap-4">
      <div>
        <h1 class="text-2xl font-bold text-default">API Keys</h1>
        <p class="text-sm text-muted">Authenticate against the public v1 REST API</p>
      </div>
      <UButton color="primary" icon="i-lucide-plus" @click="modalOpen = true">Create Key</UButton>
    </header>

    <OneShotRevealBanner
      v-if="store.lastCreated"
      :payload="store.lastCreated"
      @dismiss="dismissBanner"
    />

    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <div
        v-for="s in stats"
        :key="s.key"
        class="rounded-xl border border-default bg-default p-4 space-y-1"
      >
        <p class="text-[11px] font-medium text-muted uppercase tracking-wide">{{ s.label }}</p>
        <p class="text-2xl font-bold text-default leading-tight">{{ s.value }}</p>
        <p class="text-xs text-muted">{{ s.meta }}</p>
      </div>
    </div>

    <USkeleton v-if="loading" class="h-64 w-full" />

    <UEmpty
      v-else-if="keys.length === 0"
      icon="i-lucide-key-round"
      title="No API keys yet"
      description="Create one to call the API from CI or scripts."
    >
      <template #actions>
        <UButton color="primary" @click="modalOpen = true">Create your first key</UButton>
      </template>
    </UEmpty>

    <div v-else class="overflow-hidden rounded-xl border border-default bg-default">
      <table class="w-full text-sm">
        <thead class="bg-elevated text-xs uppercase tracking-wide text-muted">
          <tr>
            <th class="px-4 py-2 text-left font-medium">Key</th>
            <th class="px-4 py-2 text-left font-medium">Scope</th>
            <th class="px-4 py-2 text-left font-medium">Created</th>
            <th class="px-4 py-2 text-left font-medium">Last used</th>
            <th class="px-4 py-2 text-left font-medium">Expires</th>
            <th class="px-4 py-2"></th>
          </tr>
        </thead>
        <tbody class="divide-y divide-default">
          <tr
            v-for="k in keys"
            :key="k.id"
            class="hover:bg-elevated/40 transition-colors"
            :class="!k.is_active ? 'opacity-60' : ''"
          >
            <td class="px-4 py-3">
              <div class="flex items-center gap-3">
                <div
                  class="size-9 shrink-0 rounded-md flex items-center justify-center bg-primary/10 text-primary"
                >
                  <UIcon
                    :name="!k.is_active ? 'i-lucide-ban' : 'i-lucide-key-round'"
                    class="size-4"
                  />
                </div>
                <div class="min-w-0 space-y-0.5">
                  <div class="flex items-center gap-2">
                    <span class="font-semibold text-default">{{ k.name }}</span>
                    <UBadge v-if="!k.is_active" color="neutral" variant="subtle" size="xs">
                      revoked
                    </UBadge>
                  </div>
                  <p class="text-xs text-muted font-mono">{{ maskedPrefix(k) }}</p>
                </div>
              </div>
            </td>
            <td class="px-4 py-3">
              <UBadge
                v-if="!k.is_active"
                color="neutral"
                variant="subtle"
                size="sm"
                icon="i-lucide-ban"
              >
                revoked
              </UBadge>
              <UBadge
                v-else-if="k.scope === 'read_write'"
                color="warning"
                variant="subtle"
                size="sm"
                icon="i-lucide-key"
              >
                read_write
              </UBadge>
              <UBadge v-else color="info" variant="subtle" size="sm" icon="i-lucide-eye">
                read
              </UBadge>
            </td>
            <td class="px-4 py-3">
              <p class="text-default">{{ relativeTime(k.created_at) }}</p>
              <p class="text-xs text-muted">by you</p>
            </td>
            <td class="px-4 py-3 text-default">
              {{ k.last_used_at ? relativeTime(k.last_used_at) : 'never' }}
            </td>
            <td
              class="px-4 py-3"
              :class="{
                'text-warning': expiresLabel(k).tone === 'warn',
                'text-muted': expiresLabel(k).tone === 'muted',
                'text-default': expiresLabel(k).tone === 'default',
              }"
            >
              {{ expiresLabel(k).label }}
            </td>
            <td class="px-4 py-3 text-right">
              <UDropdownMenu
                v-if="k.is_active"
                :items="[
                  {
                    label: 'Revoke',
                    icon: 'i-lucide-trash-2',
                    onSelect: () => onRevoke(k),
                  },
                ]"
              >
                <UButton variant="ghost" size="xs" icon="i-lucide-more-horizontal" />
              </UDropdownMenu>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <CreateKeyModal v-model:open="modalOpen" @submit="onSubmit" />
  </div>
</template>
