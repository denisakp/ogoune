<script setup lang="ts">
import { computed, ref, watch, onMounted, onBeforeUnmount } from 'vue'
import { useStatusPublic } from '@/composables/useStatusPublic'
import UUptimeBar from '@/components/ui/UUptimeBar.vue'
import type { PublicResource } from '@/types'

const props = defineProps<{
  resource: PublicResource | null
  open: boolean
}>()

const emit = defineEmits<{ (e: 'close'): void }>()

const { resource: details, loadResourceWindows } = useStatusPublic()

const WINDOWS = [
  { key: '24h', label: 'Last 24h' },
  { key: '7d', label: 'Last 7 days' },
  { key: '30d', label: 'Last 30 days' },
  { key: '90d', label: 'Last 90 days' },
] as const

type WindowKey = (typeof WINDOWS)[number]['key']
const activeWindow = ref<WindowKey>('30d')

watch(
  () => [props.open, props.resource?.id] as const,
  async ([open, id]) => {
    if (open && id) {
      await loadResourceWindows(id)
      activeWindow.value = '30d'
    }
  },
  { immediate: true },
)

const stats = computed(
  () =>
    details.value?.windows ?? ({} as Record<string, { uptime_ratio: number; incidents: number }>),
)
const daily30 = computed(() =>
  (details.value?.daily_30d ?? []).map((d) => ({ day: d.day, ratio: d.ratio ?? null })),
)
const recent = computed(() => details.value?.recent_incidents ?? [])

const previousOverflow = ref('')
onMounted(() => {
  document.addEventListener('keydown', onKeydown)
})
onBeforeUnmount(() => {
  document.removeEventListener('keydown', onKeydown)
  document.body.style.overflow = previousOverflow.value
})

watch(
  () => props.open,
  (open) => {
    if (open) {
      previousOverflow.value = document.body.style.overflow
      document.body.style.overflow = 'hidden'
    } else {
      document.body.style.overflow = previousOverflow.value
    }
  },
)

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape' && props.open) emit('close')
}

function pct(v: number | undefined): string {
  if (v === undefined || Number.isNaN(v)) return '—'
  return `${(v * 100).toFixed(2)}%`
}
</script>

<template>
  <Teleport to="body">
    <transition name="overlay">
      <div
        v-if="open"
        class="fixed inset-0 bg-black/30 z-40"
        data-testid="overlay"
        @click="emit('close')"
      />
    </transition>
    <transition name="panel">
      <aside
        v-if="open && resource"
        class="fixed top-0 right-0 h-full w-full sm:max-w-3xl bg-white shadow-2xl z-50 overflow-y-auto"
        role="dialog"
        aria-modal="true"
        :aria-label="`Uptime detail for ${resource.name}`"
        data-testid="overall-uptime-panel"
      >
        <header class="flex items-start justify-between gap-4 px-6 py-5 border-b border-gray-200">
          <div class="min-w-0">
            <h2 class="text-lg font-semibold text-gray-900 truncate">{{ resource.name }}</h2>
            <p class="text-sm text-gray-500 font-mono truncate">{{ resource.host }}</p>
          </div>
          <button
            type="button"
            class="size-8 inline-flex items-center justify-center rounded-md text-gray-400 hover:text-gray-700 hover:bg-gray-100"
            aria-label="Close"
            data-testid="close-panel"
            @click="emit('close')"
          >
            ✕
          </button>
        </header>

        <section class="px-6 py-5 grid grid-cols-2 sm:grid-cols-4 gap-3" data-section="windows">
          <button
            v-for="w in WINDOWS"
            :key="w.key"
            type="button"
            :class="[
              'text-left rounded-lg border p-3 transition-colors',
              activeWindow === w.key
                ? 'border-indigo-300 bg-indigo-50'
                : 'border-gray-200 hover:border-gray-300',
            ]"
            :data-window="w.key"
            :data-active="activeWindow === w.key ? '1' : undefined"
            @click="activeWindow = w.key"
          >
            <p class="text-[10px] uppercase tracking-wider text-gray-500 font-semibold">
              {{ w.label }}
            </p>
            <p class="mt-1 text-2xl font-bold text-gray-900">
              {{ pct(stats[w.key]?.uptime_ratio) }}
            </p>
            <p class="text-xs text-gray-500">
              {{ stats[w.key]?.incidents ?? 0 }} incident{{
                (stats[w.key]?.incidents ?? 0) === 1 ? '' : 's'
              }}
            </p>
          </button>
        </section>

        <section class="px-6 py-5 border-t border-gray-200" data-section="daily-30">
          <div class="flex items-baseline justify-between mb-3">
            <h3 class="text-sm font-semibold text-gray-900">Daily uptime — last 30 days</h3>
            <div class="text-[10px] text-gray-500 flex items-center gap-3">
              <span class="inline-flex items-center gap-1"
                ><span class="size-2 rounded-sm bg-emerald-500" /> Up</span
              >
              <span class="inline-flex items-center gap-1"
                ><span class="size-2 rounded-sm bg-yellow-400" /> Partial</span
              >
              <span class="inline-flex items-center gap-1"
                ><span class="size-2 rounded-sm bg-red-500" /> Down</span
              >
            </div>
          </div>
          <UUptimeBar :entries="daily30" />
          <div class="mt-1 flex items-center justify-between text-[10px] text-gray-400 font-mono">
            <span>30 days ago</span>
            <span>Today</span>
          </div>
        </section>

        <section class="px-6 py-5 border-t border-gray-200" data-section="recent">
          <h3 class="text-sm font-semibold text-gray-900 mb-3">
            Recent incidents on {{ resource.name }}
          </h3>
          <p v-if="recent.length === 0" class="text-sm text-gray-500 italic">
            No recent incidents.
          </p>
          <ul v-else class="space-y-2">
            <li
              v-for="inc in recent"
              :key="inc.id"
              class="flex items-center gap-3 text-sm"
              :data-incident-id="inc.id"
            >
              <span
                :class="[
                  'size-2 rounded-full',
                  inc.resolved_at ? 'bg-emerald-500' : 'bg-orange-500',
                ]"
              />
              <a
                :href="`#/incidents/${encodeURIComponent(inc.id)}`"
                class="font-medium text-gray-900 hover:underline truncate flex-1"
              >
                {{ inc.title }}
              </a>
              <span class="text-xs text-gray-500 font-mono shrink-0">
                {{ new Date(inc.started_at).toLocaleDateString() }}
              </span>
            </li>
          </ul>
        </section>
      </aside>
    </transition>
  </Teleport>
</template>

<style scoped>
.panel-enter-active,
.panel-leave-active {
  transition: transform 200ms ease;
}
.panel-enter-from,
.panel-leave-to {
  transform: translateX(100%);
}
.overlay-enter-active,
.overlay-leave-active {
  transition: opacity 200ms ease;
}
.overlay-enter-from,
.overlay-leave-to {
  opacity: 0;
}
</style>
