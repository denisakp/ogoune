<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useToast } from '@nuxt/ui/composables/useToast'

import { useAuthStore } from '@/stores/authStore'
import { useDashboards } from '@/composables/useDashboards'
import { useConfirm } from '@/composables/useConfirm'
import { listWidgets } from '@/widgets/widgetCatalog'
import DashboardScopeResolver from './DashboardScopeResolver.vue'
import type {
  Dashboard,
  DashboardRefreshInterval,
  DashboardScope,
  DashboardTimeRange,
  DashboardVisibility,
  WidgetInstance,
  WidgetTypeId,
} from '@/types'

const props = defineProps<{
  open: boolean
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
}>()

const router = useRouter()
const authStore = useAuthStore()
const dashboardsState = useDashboards()
const toast = useToast()

const step = ref<1 | 2 | 3>(1)
const dirty = ref(false)
const saving = ref(false)

// Step 1
const name = ref('')
const scope = ref<DashboardScope>({ mode: 'tag', payload: { tagIds: [] } })
const matchCount = ref(0)

// Step 2
const selectedTypes = ref<WidgetTypeId[]>([])

// Step 3
const defaultTimeRange = ref<DashboardTimeRange>('24h')
const refreshInterval = ref<DashboardRefreshInterval>('30s')
const visibility = ref<DashboardVisibility>('private')

const widgetsAvailable = computed(() => listWidgets())

watch([name, scope, selectedTypes, defaultTimeRange, refreshInterval], () => {
  dirty.value = true
}, { deep: true })

function reset() {
  step.value = 1
  name.value = ''
  scope.value = { mode: 'tag', payload: { tagIds: [] } }
  matchCount.value = 0
  selectedTypes.value = []
  defaultTimeRange.value = '24h'
  refreshInterval.value = '30s'
  visibility.value = 'private'
  dirty.value = false
}

watch(
  () => props.open,
  (now) => {
    if (now) reset()
  },
)

const canContinueStep1 = computed(() => name.value.trim().length > 0 && matchCount.value > 0)
const canContinueStep2 = computed(() => selectedTypes.value.length > 0)

function goNext() {
  if (step.value === 1 && canContinueStep1.value) step.value = 2
  else if (step.value === 2 && canContinueStep2.value) step.value = 3
}

function goBack() {
  if (step.value === 2) step.value = 1
  else if (step.value === 3) step.value = 2
}

function toggleWidget(id: WidgetTypeId) {
  if (selectedTypes.value.includes(id)) {
    selectedTypes.value = selectedTypes.value.filter((x) => x !== id)
  } else {
    selectedTypes.value = [...selectedTypes.value, id]
  }
}

async function tryClose() {
  if (!dirty.value) {
    emit('update:open', false)
    return
  }
  const ok = await useConfirm({
    kind: 'default',
    title: 'Discard new dashboard?',
    body: "Your selections won't be saved.",
    ctaLabel: 'Discard',
  })
  if (ok) emit('update:open', false)
}

async function submit() {
  if (saving.value) return
  saving.value = true
  try {
    const widgets: WidgetInstance[] = selectedTypes.value.map((typeId, i) => ({
      id: `w-${typeId}-${Date.now()}-${i}`,
      widgetTypeId: typeId,
      position: i,
      config: {},
    }))
    const input: Omit<Dashboard, 'id' | 'createdAt' | 'updatedAt'> = {
      name: name.value.trim(),
      scope: scope.value,
      widgets,
      defaultTimeRange: defaultTimeRange.value,
      refreshInterval: refreshInterval.value,
      visibility: visibility.value,
      ownerId: authStore.userId ?? 'anonymous',
      ownerName: authStore.email ?? 'You',
    }
    const created = await dashboardsState.create(input)
    toast.add({
      title: 'Dashboard created',
      color: 'success',
      icon: 'i-lucide-check-circle',
    })
    emit('update:open', false)
    router.push({ name: 'DashboardDetail', params: { id: created.id } })
  } catch (e) {
    toast.add({
      title: "Couldn't create dashboard",
      description: e instanceof Error ? e.message : 'Unknown error',
      color: 'error',
      icon: 'i-lucide-circle-alert',
    })
  } finally {
    saving.value = false
  }
}

const timeRanges: { value: DashboardTimeRange; label: string }[] = [
  { value: '24h', label: 'Last 24 hours' },
  { value: '7d', label: 'Last 7 days' },
  { value: '30d', label: 'Last 30 days' },
  { value: '90d', label: 'Last 90 days' },
]

const refreshOptions: { value: DashboardRefreshInterval; label: string }[] = [
  { value: 'off', label: 'Off' },
  { value: '30s', label: '30 s' },
  { value: '1m', label: '1 min' },
  { value: '5m', label: '5 min' },
]
</script>

<template>
  <UModal :open="open" :ui="{ content: 'max-w-2xl' }" @update:open="(v: boolean) => v || tryClose()">
    <template #content>
      <div class="bg-default" data-testid="wizard-modal">
        <header
          class="flex items-center justify-between px-5 py-4 border-b border-default"
        >
          <div>
            <h2 class="text-base font-semibold text-default">New dashboard</h2>
            <p class="text-xs text-muted mt-0.5">Step {{ step }} of 3</p>
          </div>
          <button
            type="button"
            class="size-7 rounded hover:bg-muted flex items-center justify-center"
            aria-label="Close"
            data-testid="wizard-close"
            @click="tryClose"
          >
            <UIcon name="i-lucide-x" class="size-4 text-muted" />
          </button>
        </header>

        <div class="px-5 py-4 min-h-70">
          <section v-if="step === 1" data-testid="wizard-step-1" class="space-y-4">
            <div>
              <label class="block text-xs font-medium text-default mb-1">Name</label>
              <input
                v-model="name"
                type="text"
                placeholder="e.g. Production health"
                class="w-full px-3 py-2 text-sm border border-default rounded bg-default text-default"
                data-testid="wizard-name-input"
              />
            </div>
            <div>
              <label class="block text-xs font-medium text-default mb-1">Scope</label>
              <DashboardScopeResolver
                v-model="scope"
                @update:match-count="matchCount = $event"
              />
            </div>
          </section>

          <section v-else-if="step === 2" data-testid="wizard-step-2" class="space-y-4">
            <p class="text-xs text-muted">
              <span data-testid="wizard-widget-counter">
                {{ selectedTypes.length }} of {{ widgetsAvailable.length }} selected
              </span>
            </p>
            <div class="grid grid-cols-2 gap-3">
              <button
                v-for="w in widgetsAvailable"
                :key="w.id"
                type="button"
                class="flex items-start gap-3 p-3 border rounded text-left transition-colors"
                :class="
                  selectedTypes.includes(w.id)
                    ? 'border-primary bg-primary/5'
                    : 'border-default hover:bg-muted'
                "
                :data-testid="`wizard-widget-${w.id}`"
                @click="toggleWidget(w.id)"
              >
                <UIcon
                  :name="w.icon"
                  class="size-5 mt-0.5"
                  :class="selectedTypes.includes(w.id) ? 'text-primary' : 'text-muted'"
                />
                <div class="flex-1 min-w-0">
                  <div class="text-sm font-medium text-default">{{ w.name }}</div>
                  <div class="text-[11px] text-muted capitalize">{{ w.archetype }}</div>
                </div>
                <UIcon
                  v-if="selectedTypes.includes(w.id)"
                  name="i-lucide-check"
                  class="size-4 text-primary"
                />
              </button>
            </div>
          </section>

          <section v-else data-testid="wizard-step-3" class="space-y-5">
            <div>
              <label class="block text-xs font-medium text-default mb-2">Default time range</label>
              <select
                v-model="defaultTimeRange"
                class="w-full px-3 py-2 text-sm border border-default rounded bg-default text-default"
                data-testid="wizard-time-range"
              >
                <option v-for="r in timeRanges" :key="r.value" :value="r.value">{{ r.label }}</option>
              </select>
            </div>
            <div>
              <label class="block text-xs font-medium text-default mb-2">Refresh interval</label>
              <select
                v-model="refreshInterval"
                class="w-full px-3 py-2 text-sm border border-default rounded bg-default text-default"
                data-testid="wizard-refresh-interval"
              >
                <option v-for="r in refreshOptions" :key="r.value" :value="r.value">{{ r.label }}</option>
              </select>
            </div>
            <div>
              <label class="block text-xs font-medium text-default mb-2">Visibility</label>
              <div class="grid grid-cols-3 gap-2">
                <button
                  type="button"
                  class="flex flex-col items-start gap-1 p-3 border rounded text-left"
                  :class="
                    visibility === 'private'
                      ? 'border-primary bg-primary/5'
                      : 'border-default hover:bg-muted'
                  "
                  data-testid="wizard-visibility-private"
                  @click="visibility = 'private'"
                >
                  <UIcon name="i-lucide-lock" class="size-4 text-primary" />
                  <div class="text-sm font-medium text-default">Private</div>
                  <div class="text-[11px] text-muted">Only you can edit; org can read.</div>
                </button>
                <button
                  type="button"
                  disabled
                  class="flex flex-col items-start gap-1 p-3 border border-default rounded opacity-50 cursor-not-allowed text-left"
                  data-testid="wizard-visibility-team"
                  :aria-label="'Team — Available on Enterprise'"
                  :title="'Available on Enterprise'"
                >
                  <span class="flex items-center gap-1.5 w-full">
                    <UIcon name="i-lucide-users" class="size-4 text-muted" />
                    <UEditionBadge edition="ee" class="ml-auto" />
                  </span>
                  <div class="text-sm font-medium text-muted">Team</div>
                  <div class="text-[11px] text-muted">Shared across orgs.</div>
                </button>
                <button
                  type="button"
                  disabled
                  class="flex flex-col items-start gap-1 p-3 border border-default rounded opacity-50 cursor-not-allowed text-left"
                  data-testid="wizard-visibility-public"
                  :aria-label="'Public — Available on Enterprise'"
                  :title="'Available on Enterprise'"
                >
                  <span class="flex items-center gap-1.5 w-full">
                    <UIcon name="i-lucide-globe" class="size-4 text-muted" />
                    <UEditionBadge edition="ee" class="ml-auto" />
                  </span>
                  <div class="text-sm font-medium text-muted">Public</div>
                  <div class="text-[11px] text-muted">Anonymous URL.</div>
                </button>
              </div>
            </div>
          </section>
        </div>

        <footer
          class="flex items-center justify-between px-5 py-3 border-t border-default"
        >
          <UButton
            v-if="step > 1"
            color="neutral"
            variant="ghost"
            size="sm"
            data-testid="wizard-back"
            @click="goBack"
          >
            Back
          </UButton>
          <span v-else />
          <div class="flex items-center gap-2">
            <UButton color="neutral" variant="ghost" size="sm" @click="tryClose">Cancel</UButton>
            <UButton
              v-if="step < 3"
              color="primary"
              size="sm"
              :disabled="(step === 1 && !canContinueStep1) || (step === 2 && !canContinueStep2)"
              data-testid="wizard-continue"
              @click="goNext"
            >
              Continue
            </UButton>
            <UButton
              v-else
              color="primary"
              size="sm"
              :loading="saving"
              data-testid="wizard-submit"
              @click="submit"
            >
              Create Dashboard ✨
            </UButton>
          </div>
        </footer>
      </div>
    </template>
  </UModal>
</template>
