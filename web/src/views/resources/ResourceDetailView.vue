<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { useResourceStore } from '@/stores/resourceStore'
import { useConfirm } from '@/composables/useConfirm'
import { timeAgo } from '@/libs/date-time.helper'
import { fetchActivities } from '@/services/activityService'
import ResourceModal from '@/components/resources/ResourceModal.vue'
import ResponseTimeChart from '@/components/ResponseTimeChart.vue'
import type { Resource, MonitoringActivity } from '@/types'

const route = useRoute()
const router = useRouter()
const resourceStore = useResourceStore()

const resource = ref<Resource | null>(null)
const showModal = ref(false)
const activeTab = ref<'overview' | 'activity' | 'incidents' | 'settings'>('overview')

const activities = ref<MonitoringActivity[]>([])
const activitiesLoading = ref(true)

const statusColor = computed(() => {
  switch (resource.value?.status) {
    case 'up':
      return '#10B981'
    case 'down':
      return '#EF4444'
    case 'flapping':
      return '#F59E0B'
    case 'paused':
      return '#94A3B8'
    default:
      return '#94A3B8'
  }
})

const isPaused = computed(() => resource.value?.status === 'paused')

const targetSummary = computed(() => {
  const r = resource.value as unknown as { target?: string; type?: string } | null
  if (!r) return ''
  const t = r.target?.trim()
  const kind = (r.type ?? '').toUpperCase()
  return t ? `${kind} monitor for ${t}` : `${kind} monitor`
})

async function loadDetail() {
  const id = String(route.params.id)
  const r = await resourceStore.loadResource(id)
  resource.value = (r as Resource | undefined) ?? null
}

async function loadActivity() {
  if (!resource.value) return
  activitiesLoading.value = true
  try {
    activities.value = await fetchActivities(resource.value.id)
  } catch {
    activities.value = []
  } finally {
    activitiesLoading.value = false
  }
}

async function onModalSubmit() {
  showModal.value = false
  await loadDetail()
}

async function togglePause() {
  if (!resource.value) return
  if (isPaused.value) {
    await resourceStore.resumeMonitoring(resource.value.id)
  } else {
    await resourceStore.pauseMonitoring(resource.value.id)
  }
  await loadDetail()
}

async function onDelete() {
  if (!resource.value) return
  const ok = await useConfirm({
    kind: 'destructive',
    title: 'Delete monitor?',
    body: `${resource.value.name} will stop being checked immediately.`,
    ctaLabel: 'Delete',
  })
  if (ok) {
    await resourceStore.removeResource(resource.value.id)
    router.push('/resources')
  }
}

onMounted(async () => {
  await loadDetail()
  void loadActivity()
})

defineExpose({ resource, activeTab, loadDetail, loadActivity, togglePause, onDelete })
</script>

<template>
  <div class="bg-default text-default min-h-full">
    <div v-if="!resource" class="px-6 py-12 text-center text-sm text-slate-500">Loading…</div>
    <template v-else>
      <div class="flex items-center justify-between mb-5">
        <div class="flex items-center gap-2 text-sm text-slate-500">
          <RouterLink to="/resources" class="hover:text-slate-700">Resources</RouterLink>
          <UIcon name="i-lucide-chevron-right" class="size-3.5 text-slate-400" />
          <span class="text-slate-900 font-medium">{{ resource.name }}</span>
        </div>
        <div class="flex items-center gap-2">
          <UButton
            color="neutral"
            variant="outline"
            size="sm"
            icon="i-lucide-zap"
            @click="loadDetail"
          >
            Test now
          </UButton>
          <UButton
            color="neutral"
            variant="outline"
            size="sm"
            :icon="isPaused ? 'i-lucide-play' : 'i-lucide-pause'"
            @click="togglePause"
          >
            {{ isPaused ? 'Resume' : 'Pause' }}
          </UButton>
          <UButton
            color="neutral"
            variant="outline"
            size="sm"
            icon="i-lucide-pencil"
            @click="showModal = true"
          >
            Edit
          </UButton>
          <UButton
            color="error"
            variant="outline"
            size="sm"
            icon="i-lucide-trash-2"
            @click="onDelete"
          >
            Delete
          </UButton>
        </div>
      </div>

      <div class="flex items-center gap-4 mb-5">
        <span class="size-3 rounded-full" :style="{ backgroundColor: statusColor }" />
        <div class="flex flex-col gap-0.5">
          <h1 class="text-[22px] font-semibold font-mono text-slate-900 leading-tight">
            {{ resource.name }}
          </h1>
          <p class="text-sm text-slate-600">{{ targetSummary }}</p>
        </div>
      </div>

      <div class="flex items-center gap-1 border-b border-slate-200 mb-6">
        <button
          v-for="t in ['overview', 'activity', 'incidents', 'settings'] as const"
          :key="t"
          type="button"
          class="px-4 py-2 text-sm capitalize border-b-2 transition-colors"
          :class="
            activeTab === t
              ? 'border-primary-600 text-primary-600 font-medium'
              : 'border-transparent text-slate-600 hover:text-slate-900'
          "
          @click="activeTab = t"
        >
          {{ t }}
        </button>
      </div>

      <div v-if="activeTab === 'overview'" class="grid grid-cols-[1fr_320px] gap-6 items-start">
        <div class="flex flex-col gap-5">
          <div class="bg-white rounded-lg border border-slate-200 p-6">
            <h3 class="text-base font-semibold text-slate-900 mb-4">Response Time</h3>
            <ResponseTimeChart />
          </div>

          <div class="bg-white rounded-lg border border-slate-200 p-6">
            <div class="flex items-center justify-between mb-3">
              <h3 class="text-base font-semibold text-slate-900">Recent Activity</h3>
              <button
                type="button"
                class="text-xs text-primary-600"
                @click="activeTab = 'activity'"
              >
                View all
              </button>
            </div>
            <div v-if="activitiesLoading" class="text-sm text-slate-500 py-4 text-center">
              Loading…
            </div>
            <UEmpty
              v-else-if="activities.length === 0"
              variant="naked"
              size="sm"
              icon="i-lucide-inbox"
              title="No activity yet"
              description="Recent check results will appear here once the monitor runs."
            />
            <div v-else class="space-y-2">
              <div
                v-for="a in activities.slice(0, 10)"
                :key="a.id"
                class="flex items-center gap-3 py-2 border-b border-slate-100 last:border-0"
              >
                <span
                  class="size-1.5 rounded-full shrink-0"
                  :style="{ backgroundColor: a.success ? '#10B981' : '#EF4444' }"
                />
                <span class="text-sm text-slate-700 flex-1 truncate">
                  {{ a.message || (a.success ? 'Check passed' : 'Check failed') }}
                </span>
                <span class="text-xs text-slate-500 font-mono">{{ a.response_time }}ms</span>
                <span class="text-xs text-slate-400">{{ timeAgo(a.created_at) }}</span>
              </div>
            </div>
          </div>
        </div>

        <div class="flex flex-col gap-5">
          <div class="bg-white rounded-lg border border-slate-200 p-6">
            <h3 class="text-base font-semibold text-slate-900 mb-3">Monitor Details</h3>
            <dl class="space-y-2.5 text-sm">
              <div class="flex items-center justify-between py-2 border-b border-slate-100">
                <dt class="text-slate-500">Type</dt>
                <dd class="text-slate-900 font-medium uppercase">{{ resource.type }}</dd>
              </div>
              <div class="flex items-center justify-between py-2 border-b border-slate-100">
                <dt class="text-slate-500">Status</dt>
                <dd class="text-slate-900 font-medium capitalize">{{ resource.status }}</dd>
              </div>
              <div class="flex items-center justify-between py-2 border-b border-slate-100">
                <dt class="text-slate-500">Interval</dt>
                <dd class="text-slate-900 font-mono">
                  {{ (resource as { interval?: number }).interval ?? 60 }}s
                </dd>
              </div>
              <div
                v-if="resource.last_checked"
                class="flex items-center justify-between py-2 border-b border-slate-100"
              >
                <dt class="text-slate-500">Last check</dt>
                <dd class="text-slate-900">{{ timeAgo(resource.last_checked) }}</dd>
              </div>
              <div
                v-if="(resource as { created_at?: string }).created_at"
                class="flex items-center justify-between py-2"
              >
                <dt class="text-slate-500">Created</dt>
                <dd class="text-slate-900">
                  {{ timeAgo((resource as unknown as { created_at: string }).created_at) }}
                </dd>
              </div>
            </dl>
          </div>
        </div>
      </div>

      <div
        v-else-if="activeTab === 'activity'"
        class="bg-white rounded-lg border border-slate-200 p-6"
      >
        <h3 class="text-base font-semibold text-slate-900 mb-4">Activity log</h3>
        <div v-if="activitiesLoading" class="text-sm text-slate-500 py-8 text-center">Loading…</div>
        <UEmpty
          v-else-if="activities.length === 0"
          variant="naked"
          icon="i-lucide-inbox"
          title="No activity yet"
          description="Once this monitor starts running checks, the log will populate here."
        />
        <div v-else class="space-y-1">
          <div
            v-for="a in activities"
            :key="a.id"
            class="grid grid-cols-[20px_1fr_80px_120px] gap-3 items-center py-2 border-b border-slate-100 text-sm"
          >
            <span
              class="size-2 rounded-full"
              :style="{ backgroundColor: a.success ? '#10B981' : '#EF4444' }"
            />
            <span class="text-slate-700 truncate">{{
              a.message || (a.success ? 'Check passed' : 'Check failed')
            }}</span>
            <span class="text-slate-600 font-mono text-xs">{{ a.response_time }}ms</span>
            <span class="text-slate-400 text-xs">{{ timeAgo(a.created_at) }}</span>
          </div>
        </div>
      </div>

      <UEmpty
        v-else-if="activeTab === 'incidents'"
        class="bg-white rounded-lg border border-slate-200"
        icon="i-lucide-alert-triangle"
        title="Incidents coming with PRD 006"
        description="Per-resource incident filtering will land alongside the IncidentsView migration."
        :actions="[
          {
            label: 'See all incidents',
            icon: 'i-lucide-arrow-right',
            trailing: true,
            color: 'primary',
            variant: 'soft',
            to: '/incidents',
          },
        ]"
      />

      <div v-else class="bg-white rounded-lg border border-slate-200 p-6">
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-base font-semibold text-slate-900">Configuration</h3>
          <UButton color="primary" size="sm" icon="i-lucide-pencil" @click="showModal = true">
            Edit
          </UButton>
        </div>
        <dl class="space-y-2 text-sm">
          <div
            v-for="(value, key) in resource"
            :key="String(key)"
            class="flex items-baseline justify-between py-2 border-b border-slate-100 last:border-0"
          >
            <dt class="text-slate-500 capitalize">{{ String(key).replace(/_/g, ' ') }}</dt>
            <dd class="text-slate-900 font-mono text-xs max-w-[60%] truncate">
              {{ typeof value === 'object' ? JSON.stringify(value) : String(value ?? '—') }}
            </dd>
          </div>
        </dl>
      </div>

      <ResourceModal v-model:open="showModal" :resource="resource" @submit="onModalSubmit" />
    </template>
  </div>
</template>
