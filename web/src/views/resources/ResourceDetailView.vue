<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { useResourceStore } from '@/stores/resourceStore'
import { useConfirm } from '@/composables/useConfirm'
import { timeAgo } from '@/libs/date-time.helper'
import { fetchActivities } from '@/services/activityService'
import { fetchUptimeStats } from '@/services/resourceService'
import ResourceModal from '@/components/resources/ResourceModal.vue'
import type { Resource, MonitoringActivity, HourlyUptimeStat } from '@/types'

const route = useRoute()
const router = useRouter()
const resourceStore = useResourceStore()

const resource = ref<Resource | null>(null)
const showModal = ref(false)
const activeTab = ref<'overview' | 'activity' | 'incidents' | 'settings'>('overview')

const ACTIVITIES_PAGE_SIZE = 10
const STRIP_TARGET_BUCKETS = 20

const activities = ref<MonitoringActivity[]>([])
const activitiesLoading = ref(true)
const activitiesLoadingMore = ref(false)
const activitiesHasMore = ref(true)
const activityView = ref<'timeline' | 'strip'>('timeline')
const selectedActivityId = ref<string | null>(null)
const selectedBucketIndex = ref<number | null>(null)

const hourlyStats = ref<HourlyUptimeStat[]>([])
const chartRange = ref<'24h' | '7d' | '30d'>('24h')

interface MetadataLike {
  ssl_issuer?: string
  ssl_expiration_date?: string
  ssl_days_remaining?: number
  domain_registrar?: string
  domain_expiration_date?: string
  domain_days_remaining?: number
}

const metadata = computed<MetadataLike | null>(
  () => (resource.value as unknown as { metadata?: MetadataLike } | null)?.metadata ?? null,
)
const expiryStatus = computed(
  () => (resource.value as unknown as { expiry_status?: string } | null)?.expiry_status ?? 'ok',
)
const sslColor = computed(() => {
  switch (expiryStatus.value) {
    case 'critical':
    case 'expired':
      return { bg: '#FEF2F2', border: '#FCA5A5', icon: '#B91C1C' }
    case 'warning':
      return { bg: '#FFFBEB', border: '#FCD34D', icon: '#B45309' }
    default:
      return { bg: '#ECFDF5', border: '#6EE7B7', icon: '#047857' }
  }
})

interface ActivityGroup {
  label: string
  items: MonitoringActivity[]
}

function dayLabel(d: Date): string {
  const now = new Date()
  const today = new Date(now.getFullYear(), now.getMonth(), now.getDate())
  const yesterday = new Date(today.getTime() - 86_400_000)
  const start = new Date(d.getFullYear(), d.getMonth(), d.getDate())
  if (start.getTime() === today.getTime()) return 'Today'
  if (start.getTime() === yesterday.getTime()) return 'Yesterday'
  const daysAgo = Math.floor((today.getTime() - start.getTime()) / 86_400_000)
  if (daysAgo < 7) return `${daysAgo} days ago`
  return d.toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' })
}

function timeOfDay(iso: string): string {
  const d = new Date(iso)
  return d.toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit', second: '2-digit' })
}

interface StripBucket {
  index: number
  startTs: string
  endTs: string
  totalCount: number
  failedCount: number
  items: MonitoringActivity[]
  tone: 'up' | 'mixed' | 'down'
}

const stripBuckets = computed<StripBucket[]>(() => {
  const sorted = [...activities.value].sort(
    (a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime(),
  )
  if (sorted.length === 0) return []

  const size = Math.max(1, Math.ceil(sorted.length / STRIP_TARGET_BUCKETS))
  const out: StripBucket[] = []
  for (let i = 0; i < sorted.length; i += size) {
    const slice = sorted.slice(i, i + size)
    const failedCount = slice.filter((s) => !s.success).length
    out.push({
      index: out.length,
      startTs: slice[0]!.created_at,
      endTs: slice[slice.length - 1]!.created_at,
      totalCount: slice.length,
      failedCount,
      items: slice,
      tone: failedCount === 0 ? 'up' : failedCount === slice.length ? 'down' : 'mixed',
    })
  }
  return out
})

const selectedBucket = computed(() =>
  selectedBucketIndex.value != null
    ? stripBuckets.value[selectedBucketIndex.value] ?? null
    : null,
)

const selectedActivity = computed(() =>
  selectedActivityId.value
    ? activities.value.find((a) => a.id === selectedActivityId.value) ?? null
    : null,
)

function bucketColor(tone: StripBucket['tone']): string {
  if (tone === 'down') return '#EF4444'
  if (tone === 'mixed') return '#F59E0B'
  return '#10B981'
}

const activityGroups = computed<ActivityGroup[]>(() => {
  const map = new Map<string, ActivityGroup>()
  for (const a of activities.value) {
    if (!a.created_at) continue
    const d = new Date(a.created_at)
    const key = `${d.getFullYear()}-${d.getMonth()}-${d.getDate()}`
    if (!map.has(key)) map.set(key, { label: dayLabel(d), items: [] })
    map.get(key)!.items.push(a)
  }
  return [...map.values()].map((g) => ({
    label: g.label,
    items: g.items.sort(
      (a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime(),
    ),
  }))
})

const responseTimes = computed(
  () =>
    (resource.value as unknown as { response_times?: { timestamp: string; response_time: number }[] } | null)
      ?.response_times ?? [],
)

const filteredResponseTimes = computed(() => {
  const hoursByRange: Record<typeof chartRange.value, number> = {
    '24h': 24,
    '7d': 24 * 7,
    '30d': 24 * 30,
  }
  const cutoff = Date.now() - hoursByRange[chartRange.value] * 3_600_000
  return responseTimes.value
    .filter((p) => new Date(p.timestamp).getTime() >= cutoff)
    .sort((a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime())
})

const chartAvg = computed(() => {
  const values = filteredResponseTimes.value.map((p) => p.response_time)
  if (values.length === 0) return 0
  return values.reduce((s, v) => s + v, 0) / values.length
})

const TARGET_BARS = 60

const chartBars = computed(() => {
  const data = filteredResponseTimes.value
  if (data.length === 0) return []

  const bucketSize = Math.max(1, Math.ceil(data.length / TARGET_BARS))
  const buckets: { timestamp: string; value: number }[] = []
  for (let i = 0; i < data.length; i += bucketSize) {
    const slice = data.slice(i, i + bucketSize)
    const avg = slice.reduce((s, p) => s + p.response_time, 0) / slice.length
    buckets.push({ timestamp: slice[0]!.timestamp, value: Math.round(avg) })
  }

  const max = Math.max(...buckets.map((b) => b.value), 1)
  return buckets.map((b) => {
    const isAnomaly = chartAvg.value > 0 && b.value >= chartAvg.value * 2
    return {
      timestamp: b.timestamp,
      value: b.value,
      heightPct: Math.max(8, (b.value / max) * 100),
      color: isAnomaly ? '#EF4444' : '#A5B4FC',
    }
  })
})

const failureCount = computed(
  () => (resource.value as unknown as { failure_count?: number } | null)?.failure_count ?? 0,
)

const uptimeWindows = computed(() => {
  const stats = hourlyStats.value
  const now = Date.now()
  const ranges: Array<{ key: string; hours: number }> = [
    { key: '1h', hours: 1 },
    { key: '24h', hours: 24 },
    { key: '7d', hours: 24 * 7 },
    { key: '30d', hours: 24 * 30 },
    { key: '90d', hours: 24 * 90 },
  ]
  return ranges.map((r) => {
    const cutoff = now - r.hours * 3_600_000
    const inWindow = stats.filter((s) => new Date(s.hour).getTime() >= cutoff)
    const total = inWindow.reduce((s, x) => s + x.total_count, 0)
    const success = inWindow.reduce((s, x) => s + x.successful_count, 0)
    if (total === 0) return { key: r.key, value: '—', tone: 'neutral' as const }
    const pct = (success / total) * 100
    const tone: 'good' | 'warning' | 'bad' = pct >= 99.9 ? 'good' : pct >= 99 ? 'warning' : 'bad'
    return { key: r.key, value: `${pct.toFixed(2)}%`, tone }
  })
})

const uptime30d = computed(() => uptimeWindows.value.find((w) => w.key === '30d') ?? null)

const responseTimeStats = computed(() => {
  const rt = filteredResponseTimes.value
  if (rt.length === 0) return null
  const values = rt.map((r) => r.response_time)
  const sum = values.reduce((s, v) => s + v, 0)
  return {
    avg: Math.round(sum / values.length),
    min: Math.min(...values),
    max: Math.max(...values),
  }
})

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

const limitByRange: Record<typeof chartRange.value, number> = {
  '24h': 200,
  '7d': 800,
  '30d': 2000,
}

async function loadDetail() {
  const id = String(route.params.id)
  const r = await resourceStore.loadResourceWithResponseTimes(id, limitByRange[chartRange.value])
  resource.value = (r as Resource | undefined) ?? null
}

watch(chartRange, () => {
  void loadDetail()
})

async function loadActivity() {
  if (!resource.value) return
  activitiesLoading.value = true
  activitiesHasMore.value = true
  selectedActivityId.value = null
  selectedBucketIndex.value = null
  try {
    const page = await fetchActivities(resource.value.id, ACTIVITIES_PAGE_SIZE, 0)
    activities.value = page
    activitiesHasMore.value = page.length === ACTIVITIES_PAGE_SIZE
  } catch {
    activities.value = []
    activitiesHasMore.value = false
  } finally {
    activitiesLoading.value = false
  }
}

async function loadMoreActivity() {
  if (!resource.value || activitiesLoadingMore.value || !activitiesHasMore.value) return
  activitiesLoadingMore.value = true
  try {
    const page = await fetchActivities(
      resource.value.id,
      ACTIVITIES_PAGE_SIZE,
      activities.value.length,
    )
    activities.value = [...activities.value, ...page]
    activitiesHasMore.value = page.length === ACTIVITIES_PAGE_SIZE
  } catch {
    activitiesHasMore.value = false
  } finally {
    activitiesLoadingMore.value = false
  }
}

async function loadUptimeWindows() {
  if (!resource.value) return
  try {
    const r = await fetchUptimeStats(resource.value.id)
    hourlyStats.value = r.stats ?? []
  } catch {
    hourlyStats.value = []
  }
}

function formatDate(iso?: string): string {
  if (!iso) return '—'
  return new Date(iso).toLocaleDateString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })
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
  void loadUptimeWindows()
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
        <div class="flex flex-col gap-5 min-w-0">
          <div class="grid grid-cols-4 gap-3">
            <div class="bg-white rounded-lg border border-slate-200 p-4">
              <div class="text-xs text-slate-500 mb-1">Status</div>
              <div
                class="text-xl font-semibold capitalize"
                :style="{ color: statusColor }"
              >
                {{ resource.status }}
              </div>
            </div>
            <div class="bg-white rounded-lg border border-slate-200 p-4">
              <div class="text-xs text-slate-500 mb-1">Uptime (30d)</div>
              <div class="text-xl font-semibold text-slate-900">
                {{ uptime30d?.value ?? '—' }}
              </div>
            </div>
            <div class="bg-white rounded-lg border border-slate-200 p-4">
              <div class="text-xs text-slate-500 mb-1">Failures (30d)</div>
              <div class="text-xl font-semibold text-slate-900">{{ failureCount }}</div>
            </div>
            <div class="bg-white rounded-lg border border-slate-200 p-4">
              <div class="text-xs text-slate-500 mb-1">Last Checked</div>
              <div class="text-xl font-semibold text-slate-900">
                {{ resource.last_checked ? timeAgo(resource.last_checked) : '—' }}
              </div>
            </div>
          </div>

          <div class="bg-white rounded-lg border border-slate-200 p-5">
            <h3 class="text-base font-semibold text-slate-900 mb-3">Uptime by Time Window</h3>
            <div class="grid grid-cols-5 gap-3">
              <div
                v-for="w in uptimeWindows"
                :key="w.key"
                class="bg-slate-50 rounded-md p-3 text-center"
              >
                <div class="text-[11px] text-slate-500 mb-1">{{ w.key }}</div>
                <div
                  class="text-base font-semibold"
                  :class="{
                    'text-emerald-600': w.tone === 'good',
                    'text-amber-600': w.tone === 'warning',
                    'text-red-600': w.tone === 'bad',
                    'text-slate-400': w.tone === 'neutral',
                  }"
                >
                  {{ w.value }}
                </div>
              </div>
            </div>
          </div>

          <div class="bg-white rounded-lg border border-slate-200 p-5">
            <div class="flex items-center justify-between mb-3">
              <h3 class="text-base font-semibold text-slate-900">Response Time</h3>
              <div class="flex p-0.5 rounded-md bg-slate-50">
                <button
                  v-for="r in (['24h', '7d', '30d'] as const)"
                  :key="r"
                  type="button"
                  class="px-3 py-1 rounded text-xs font-medium transition-colors"
                  :class="
                    chartRange === r
                      ? 'bg-white text-slate-900 shadow-sm'
                      : 'text-slate-500 hover:text-slate-700'
                  "
                  @click="chartRange = r"
                >
                  {{ r }}
                </button>
              </div>
            </div>
            <div
              v-if="responseTimeStats"
              class="flex items-center gap-5 text-xs text-slate-600 mb-3"
            >
              <span class="inline-flex items-center gap-1.5">
                <span class="size-2 rounded-full" style="background-color: #4f46e5" />
                Average <strong class="text-slate-900 font-semibold">{{ responseTimeStats.avg }}ms</strong>
              </span>
              <span class="inline-flex items-center gap-1.5">
                <span class="size-2 rounded-full" style="background-color: #10b981" />
                Min <strong class="text-slate-900 font-semibold">{{ responseTimeStats.min }}ms</strong>
              </span>
              <span class="inline-flex items-center gap-1.5">
                <span class="size-2 rounded-full" style="background-color: #ef4444" />
                Max <strong class="text-slate-900 font-semibold">{{ responseTimeStats.max }}ms</strong>
              </span>
            </div>
            <div v-if="chartBars.length === 0" class="text-sm text-slate-500 text-center py-12">
              No response time data for this window.
            </div>
            <div v-else class="flex items-end gap-1.5 h-44">
              <div
                v-for="(b, i) in chartBars"
                :key="i"
                class="flex-1 rounded-sm group relative"
                :style="{ height: `${b.heightPct}%`, backgroundColor: b.color }"
              >
                <div
                  class="absolute bottom-full left-1/2 -translate-x-1/2 mb-1 hidden group-hover:flex flex-col items-center px-2 py-1 rounded bg-slate-900 text-white text-[10px] whitespace-nowrap"
                >
                  <span class="font-mono">{{ b.value }}ms</span>
                  <span class="text-slate-300">{{ new Date(b.timestamp).toLocaleString() }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="flex flex-col gap-5">
          <div class="bg-white rounded-lg border border-slate-200 p-5">
            <h3 class="text-base font-semibold text-slate-900 mb-4">Monitor Details</h3>
            <dl class="space-y-3.5 text-sm">
              <div>
                <dt class="text-[11px] text-slate-500 uppercase tracking-wide mb-0.5">Type</dt>
                <dd class="text-slate-900 font-medium">{{ String(resource.type).toUpperCase() }}</dd>
              </div>
              <div>
                <dt class="text-[11px] text-slate-500 uppercase tracking-wide mb-0.5">Target</dt>
                <dd class="text-slate-900 font-mono text-xs break-all">
                  {{ (resource as unknown as { target?: string }).target ?? '—' }}
                </dd>
              </div>
              <div v-if="resource.type === 'http'">
                <dt class="text-[11px] text-slate-500 uppercase tracking-wide mb-0.5">Method</dt>
                <dd class="text-slate-900 font-mono text-xs">
                  {{ (resource as unknown as { method?: string }).method ?? 'GET' }}
                </dd>
              </div>
              <div>
                <dt class="text-[11px] text-slate-500 uppercase tracking-wide mb-0.5">Interval</dt>
                <dd class="text-slate-900">
                  Every {{ (resource as unknown as { interval?: number }).interval ?? 60 }} seconds
                </dd>
              </div>
              <div v-if="(resource as unknown as { timeout?: number }).timeout != null">
                <dt class="text-[11px] text-slate-500 uppercase tracking-wide mb-0.5">Timeout</dt>
                <dd class="text-slate-900">
                  {{ (resource as unknown as { timeout: number }).timeout }} seconds
                </dd>
              </div>
              <div v-if="(resource as unknown as { created_at?: string }).created_at">
                <dt class="text-[11px] text-slate-500 uppercase tracking-wide mb-0.5">Created</dt>
                <dd class="text-slate-900">
                  {{ formatDate((resource as unknown as { created_at: string }).created_at) }}
                </dd>
              </div>
              <div v-if="(resource as unknown as { updated_at?: string }).updated_at">
                <dt class="text-[11px] text-slate-500 uppercase tracking-wide mb-0.5">Last Updated</dt>
                <dd class="text-slate-900">
                  {{ formatDate((resource as unknown as { updated_at: string }).updated_at) }}
                </dd>
              </div>
              <div v-if="resource.tags && resource.tags.length > 0">
                <dt class="text-[11px] text-slate-500 uppercase tracking-wide mb-1.5">Tags</dt>
                <dd class="flex flex-wrap gap-1.5">
                  <span
                    v-for="t in resource.tags"
                    :key="(t as unknown as { id?: string }).id ?? String(t)"
                    class="inline-block px-2 py-0.5 rounded-md text-xs bg-slate-100 text-slate-700"
                  >
                    {{ (t as unknown as { name?: string }).name ?? String(t) }}
                  </span>
                </dd>
              </div>
            </dl>
          </div>

          <div
            v-if="metadata?.ssl_issuer || metadata?.ssl_expiration_date"
            class="rounded-lg border p-5"
            :style="{ backgroundColor: sslColor.bg, borderColor: sslColor.border }"
          >
            <div class="flex items-center gap-2 mb-3">
              <UIcon name="i-lucide-shield-check" class="size-4" :style="{ color: sslColor.icon }" />
              <h3 class="text-sm font-semibold" :style="{ color: sslColor.icon }">SSL Certificate</h3>
            </div>
            <div class="space-y-1.5 text-xs">
              <div v-if="metadata?.ssl_issuer" class="text-slate-700">
                Issuer: <span class="font-medium">{{ metadata.ssl_issuer }}</span>
              </div>
              <div v-if="metadata?.ssl_expiration_date" class="text-slate-700">
                Expires: <span class="font-medium">{{ formatDate(metadata.ssl_expiration_date) }}</span>
                <span v-if="metadata.ssl_days_remaining != null" class="text-slate-500">
                  ({{ metadata.ssl_days_remaining }} days)
                </span>
              </div>
            </div>
          </div>

          <div
            v-if="metadata?.domain_registrar || metadata?.domain_expiration_date"
            class="rounded-lg border border-slate-200 bg-white p-5"
          >
            <div class="flex items-center gap-2 mb-3">
              <UIcon name="i-lucide-globe" class="size-4 text-slate-600" />
              <h3 class="text-sm font-semibold text-slate-900">Domain</h3>
            </div>
            <div class="space-y-1.5 text-xs">
              <div v-if="metadata?.domain_registrar" class="text-slate-700">
                Registrar: <span class="font-medium">{{ metadata.domain_registrar }}</span>
              </div>
              <div v-if="metadata?.domain_expiration_date" class="text-slate-700">
                Expires: <span class="font-medium">{{ formatDate(metadata.domain_expiration_date) }}</span>
                <span v-if="metadata.domain_days_remaining != null" class="text-slate-500">
                  ({{ metadata.domain_days_remaining }} days)
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div
        v-else-if="activeTab === 'activity'"
        class="bg-white rounded-lg border border-slate-200 overflow-hidden"
      >
        <div class="flex items-center justify-between gap-3 px-6 py-4 border-b border-slate-200">
          <h3 class="text-base font-semibold text-slate-900">Activity log</h3>
          <div class="flex items-center gap-3">
            <span v-if="activities.length > 0" class="text-xs text-slate-500">
              {{ activities.length }} check{{ activities.length > 1 ? 's' : '' }}
            </span>
            <div class="flex p-0.5 rounded-md bg-slate-50">
              <button
                type="button"
                class="px-2.5 py-1 rounded text-[11px] font-medium inline-flex items-center gap-1.5"
                :class="
                  activityView === 'timeline'
                    ? 'bg-white text-slate-900 shadow-sm'
                    : 'text-slate-500 hover:text-slate-700'
                "
                @click="activityView = 'timeline'"
              >
                <UIcon name="i-lucide-list" class="size-3" />
                Timeline
              </button>
              <button
                type="button"
                class="px-2.5 py-1 rounded text-[11px] font-medium inline-flex items-center gap-1.5"
                :class="
                  activityView === 'strip'
                    ? 'bg-white text-slate-900 shadow-sm'
                    : 'text-slate-500 hover:text-slate-700'
                "
                @click="activityView = 'strip'"
              >
                <UIcon name="i-lucide-bar-chart-3" class="size-3" />
                Strip
              </button>
            </div>
          </div>
        </div>

        <div v-if="activitiesLoading" class="px-6 py-4 space-y-2">
          <USkeleton v-for="i in 6" :key="i" class="h-6 w-full" />
        </div>
        <UEmpty
          v-else-if="activityGroups.length === 0"
          variant="naked"
          icon="i-lucide-inbox"
          title="No activity yet"
          description="Once this monitor starts running checks, the log will populate here."
        />
        <div v-else-if="activityView === 'timeline'" class="px-6 py-4">
          <div v-for="g in activityGroups" :key="g.label" class="mb-6 last:mb-0">
            <div
              class="text-[10px] font-semibold tracking-wider text-slate-400 uppercase mb-2 pl-6"
            >
              {{ g.label }}
            </div>
            <div class="relative pl-6 border-l border-slate-200">
              <div
                v-for="a in g.items"
                :key="a.id"
                class="relative -ml-[27px] flex items-center gap-3 py-2 pl-6 pr-2 rounded-md hover:bg-slate-50"
              >
                <span
                  class="absolute left-0 size-2 rounded-full ring-2 ring-white shrink-0"
                  :style="{ backgroundColor: a.success ? '#10B981' : '#EF4444' }"
                />
                <span class="text-xs font-mono text-slate-500 w-[80px] shrink-0">
                  {{ timeOfDay(a.created_at) }}
                </span>
                <span
                  class="inline-flex items-center px-2 py-0.5 rounded-full text-[11px] font-medium shrink-0"
                  :style="
                    a.success
                      ? { backgroundColor: '#ECFDF5', color: '#047857' }
                      : { backgroundColor: '#FEF2F2', color: '#B91C1C' }
                  "
                >
                  {{ a.success ? 'Up' : 'Down' }}
                </span>
                <span class="text-sm text-slate-700 flex-1 truncate min-w-0">
                  {{ a.message || (a.success ? 'Check passed' : 'Check failed') }}
                </span>
                <span class="text-xs font-mono text-slate-500 shrink-0">
                  {{ a.response_time != null ? `${a.response_time}ms` : '—' }}
                </span>
              </div>
            </div>
          </div>
          <div v-if="activitiesHasMore" class="flex justify-center pt-2 pb-1">
            <UButton
              color="neutral"
              variant="outline"
              size="sm"
              :loading="activitiesLoadingMore"
              icon="i-lucide-chevron-down"
              @click="loadMoreActivity"
            >
              Load {{ ACTIVITIES_PAGE_SIZE }} more
            </UButton>
          </div>
          <div
            v-else-if="activities.length > 0"
            class="text-center text-[11px] text-slate-400 pt-2 pb-1"
          >
            End of log
          </div>
        </div>

        <div v-else class="px-6 py-5">
          <div class="flex items-end gap-px h-12 mb-2">
            <button
              v-for="b in stripBuckets"
              :key="b.index"
              type="button"
              class="flex-1 min-w-[3px] rounded-sm transition-opacity"
              :class="{
                'opacity-100': selectedBucketIndex === b.index,
                'opacity-90 hover:opacity-100': selectedBucketIndex !== b.index,
                'ring-2 ring-primary-600 ring-offset-1': selectedBucketIndex === b.index,
              }"
              :style="{ backgroundColor: bucketColor(b.tone), height: '100%' }"
              :title="`${new Date(b.startTs).toLocaleString()} → ${new Date(b.endTs).toLocaleString()} · ${b.totalCount} checks (${b.failedCount} failed)`"
              @click="selectedBucketIndex = selectedBucketIndex === b.index ? null : b.index"
            />
          </div>
          <div class="flex items-center justify-between text-[10px] text-slate-400 mb-4">
            <span v-if="stripBuckets.length > 0">
              {{ new Date(stripBuckets[0]!.startTs).toLocaleString() }}
            </span>
            <span class="text-slate-500">
              {{ activities.length }} checks
              <template v-if="stripBuckets.length < activities.length">
                · bucketed into {{ stripBuckets.length }} groups
              </template>
            </span>
            <span v-if="stripBuckets.length > 0">
              {{ new Date(stripBuckets[stripBuckets.length - 1]!.endTs).toLocaleString() }}
            </span>
          </div>

          <div
            v-if="selectedBucket"
            class="rounded-md border border-slate-200 bg-slate-50 p-4 space-y-3"
          >
            <div class="flex items-center justify-between gap-3">
              <div class="flex items-center gap-2 flex-wrap">
                <span
                  class="size-2 rounded-full"
                  :style="{ backgroundColor: bucketColor(selectedBucket.tone) }"
                />
                <span
                  class="inline-flex items-center px-2 py-0.5 rounded-full text-[11px] font-medium"
                  :style="
                    selectedBucket.tone === 'up'
                      ? { backgroundColor: '#ECFDF5', color: '#047857' }
                      : selectedBucket.tone === 'down'
                        ? { backgroundColor: '#FEF2F2', color: '#B91C1C' }
                        : { backgroundColor: '#FFFBEB', color: '#92400E' }
                  "
                >
                  {{
                    selectedBucket.tone === 'up'
                      ? 'All Up'
                      : selectedBucket.tone === 'down'
                        ? 'All Down'
                        : 'Mixed'
                  }}
                </span>
                <span class="text-xs text-slate-600">
                  {{ selectedBucket.totalCount }} check{{
                    selectedBucket.totalCount > 1 ? 's' : ''
                  }}
                  <template v-if="selectedBucket.failedCount > 0">
                    · <strong>{{ selectedBucket.failedCount }} failed</strong>
                  </template>
                </span>
              </div>
              <button
                type="button"
                class="text-slate-400 hover:text-slate-700"
                @click="selectedBucketIndex = null"
              >
                <UIcon name="i-lucide-x" class="size-4" />
              </button>
            </div>
            <div class="text-xs text-slate-600">
              From <span class="font-mono">{{ new Date(selectedBucket.startTs).toLocaleString() }}</span>
              to <span class="font-mono">{{ new Date(selectedBucket.endTs).toLocaleString() }}</span>
            </div>
            <div v-if="selectedBucket.items.length > 0" class="space-y-1 max-h-60 overflow-auto">
              <button
                v-for="a in selectedBucket.items"
                :key="a.id"
                type="button"
                class="w-full text-left flex items-center gap-3 px-2 py-1.5 rounded hover:bg-white text-xs"
                :class="{ 'bg-white ring-1 ring-primary-200': selectedActivityId === a.id }"
                @click="selectedActivityId = selectedActivityId === a.id ? null : a.id"
              >
                <span
                  class="size-1.5 rounded-full shrink-0"
                  :style="{ backgroundColor: a.success ? '#10B981' : '#EF4444' }"
                />
                <span class="font-mono text-slate-600 w-[80px] shrink-0">
                  {{ timeOfDay(a.created_at) }}
                </span>
                <span class="text-slate-700 flex-1 truncate min-w-0">
                  {{ a.message || (a.success ? 'Check passed' : 'Check failed') }}
                </span>
                <span class="text-slate-500 font-mono shrink-0">
                  {{ a.response_time != null ? `${a.response_time}ms` : '—' }}
                </span>
              </button>
            </div>
            <div
              v-if="selectedActivity && (selectedActivity as unknown as { response_data?: string }).response_data"
              class="text-xs"
            >
              <div class="text-[10px] text-slate-500 uppercase tracking-wider mb-1">
                Response data
              </div>
              <pre
                class="bg-white border border-slate-200 rounded p-2 font-mono text-[11px] text-slate-700 max-h-32 overflow-auto"
                >{{ (selectedActivity as unknown as { response_data: string }).response_data }}</pre
              >
            </div>
          </div>
          <div
            v-else
            class="text-xs text-slate-500 text-center py-3 border border-dashed border-slate-200 rounded-md"
          >
            Hover a bar for a quick preview · Click for the bucket's check list.
          </div>

          <div v-if="activitiesHasMore" class="flex justify-center pt-4">
            <UButton
              color="neutral"
              variant="outline"
              size="sm"
              :loading="activitiesLoadingMore"
              icon="i-lucide-chevron-down"
              @click="loadMoreActivity"
            >
              Load {{ ACTIVITIES_PAGE_SIZE }} more
            </UButton>
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
