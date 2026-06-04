<script setup lang="ts">
/* eslint-disable @typescript-eslint/ban-ts-comment */
// @ts-nocheck — spec 059 polish debt: NuxtUI v4 template-emit types
/**
 * Maintenance create modal — design fidelity v2.
 *
 * Strategy = One-time → Start at / End at datetime inputs.
 * Strategy = Cron     → Time + Timezone, frequency pills + weekday checkboxes
 *                       (or month-days, or first/last day), generated cron
 *                       (editable), duration minutes, optional effective-from/until.
 * Resources = USelectMenu multiple from fetchResources().
 */
import { computed, ref, watch } from 'vue'
import cronstrue from 'cronstrue'
import { CronExpressionParser } from 'cron-parser'
import type { Resource, CreateMaintenance } from '@/types'
import { fetchResources } from '@/services/resourceService'

interface Props {
  open: boolean
}
defineProps<Props>()
const emit = defineEmits<{
  (e: 'update:open', v: boolean): void
  (e: 'submit', v: CreateMaintenance): void
}>()

// ---- shared fields ----
const title = ref<string>('')
const description = ref<string>('')
const strategy = ref<'one_time' | 'cron'>('one_time')
const resources = ref<Resource[]>([])
const selectedResourceIds = ref<string[]>([])

// ---- one-time ----
const startAt = ref<string>('')
const endAt = ref<string>('')

// ---- cron ----
const cronTime = ref<string>('02:00')
const cronTimezone = ref<string>('UTC')
const cronMode = ref<'weekdays' | 'monthdays' | 'first_last'>('weekdays')
const weekdays = ref<boolean[]>([false, false, false, false, false, false, false])
const monthDays = ref<number[]>([])
const firstLast = ref<'first' | 'last'>('first')
const editingCronManually = ref<boolean>(false)
const cronExpr = ref<string>('0 2 * * *')
const durationMinutes = ref<number>(60)
const showAdvanced = ref<boolean>(false)
const effectiveFrom = ref<string>('')
const effectiveUntil = ref<string>('')

const timezones = computed<string[]>(() => {
  type IntlMaybe = typeof Intl & { supportedValuesOf?: (k: string) => string[] }
  const I = Intl as IntlMaybe
  if (typeof I.supportedValuesOf === 'function') return I.supportedValuesOf('timeZone')
  return [
    'UTC',
    'Europe/Paris',
    'Europe/London',
    'America/New_York',
    'America/Los_Angeles',
    'Asia/Tokyo',
  ]
})

const weekdayLabels = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']

// Build cron expression from the friendly inputs.
function buildCron(): string {
  const [h, m] = cronTime.value.split(':').map((n) => Number(n) || 0)
  if (cronMode.value === 'weekdays') {
    const days = weekdays.value
      .map((checked, i) => (checked ? i : -1))
      .filter((i) => i >= 0)
      .join(',')
    return `${m} ${h} * * ${days || '*'}`
  }
  if (cronMode.value === 'monthdays') {
    return `${m} ${h} ${monthDays.value.length > 0 ? monthDays.value.join(',') : '*'} * *`
  }
  // first_last: a 5-field cron can express "1st of month" easily; "last day" needs
  // L which not all parsers support. We pick the 1st here; "last" surfaces 28-31
  // OR-ed and the server should accept it as best effort.
  if (firstLast.value === 'first') return `${m} ${h} 1 * *`
  return `${m} ${h} 28,29,30,31 * *`
}

watch(
  [cronTime, cronMode, weekdays, monthDays, firstLast],
  () => {
    if (!editingCronManually.value) cronExpr.value = buildCron()
  },
  { deep: true },
)

const cronHumanReadable = computed<string>(() => {
  try {
    return cronstrue.toString(cronExpr.value, { use24HourTimeFormat: true })
  } catch {
    return 'Invalid cron expression'
  }
})

const cronIsValid = computed<boolean>(() => {
  try {
    CronExpressionParser.parse(cronExpr.value)
    return true
  } catch {
    return false
  }
})

const canSubmit = computed<boolean>(() => {
  if (!title.value.trim()) return false
  if (strategy.value === 'one_time') {
    if (!startAt.value || !endAt.value) return false
    return new Date(startAt.value).getTime() < new Date(endAt.value).getTime()
  }
  return cronIsValid.value && durationMinutes.value >= 5
})

async function loadResources() {
  try {
    resources.value = await fetchResources()
  } catch {
    resources.value = []
  }
}

watch(
  () => undefined,
  () => loadResources(),
  { immediate: true },
)

function toggleWeekday(i: number) {
  weekdays.value = weekdays.value.map((v, idx) => (idx === i ? !v : v))
}

function reset() {
  title.value = ''
  description.value = ''
  strategy.value = 'one_time'
  selectedResourceIds.value = []
  startAt.value = ''
  endAt.value = ''
  cronTime.value = '02:00'
  cronTimezone.value = 'UTC'
  cronMode.value = 'weekdays'
  weekdays.value = [false, false, false, false, false, false, false]
  monthDays.value = []
  firstLast.value = 'first'
  editingCronManually.value = false
  cronExpr.value = '0 2 * * *'
  durationMinutes.value = 60
  showAdvanced.value = false
  effectiveFrom.value = ''
  effectiveUntil.value = ''
}

function close() {
  reset()
  emit('update:open', false)
}

function onSubmit() {
  if (!canSubmit.value) return
  if (strategy.value === 'one_time') {
    emit('submit', {
      title: title.value.trim(),
      description: description.value || undefined,
      strategy: 'one_time',
      start_at: new Date(startAt.value).toISOString(),
      end_at: new Date(endAt.value).toISOString(),
      resource_ids: selectedResourceIds.value,
    })
  } else {
    emit('submit', {
      title: title.value.trim(),
      description: description.value || undefined,
      strategy: 'cron',
      cron_expr: cronExpr.value,
      window_minutes: durationMinutes.value,
      timezone: cronTimezone.value,
      effective_from: effectiveFrom.value ? new Date(effectiveFrom.value).toISOString() : undefined,
      effective_until: effectiveUntil.value
        ? new Date(effectiveUntil.value).toISOString()
        : undefined,
      resource_ids: selectedResourceIds.value,
    })
  }
  reset()
}

defineExpose({
  title,
  description,
  strategy,
  startAt,
  endAt,
  cronExpr,
  cronTime,
  cronMode,
  weekdays,
  durationMinutes,
  canSubmit,
  cronIsValid,
  toggleWeekday,
  onSubmit,
})
</script>

<template>
  <UModal :open="open" title="New maintenance" @update:open="emit('update:open', $event)">
    <template #body>
      <div class="space-y-5">
        <!-- Title + Strategy row -->
        <div class="grid grid-cols-2 gap-5">
          <UFormField>
            <template #label> <span class="text-error">*</span> Title </template>
            <UInput v-model="title" placeholder="e.g., Database upgrade" />
          </UFormField>

          <UFormField>
            <template #label> <span class="text-error">*</span> Strategy </template>
            <div class="flex items-center gap-5">
              <label class="flex items-center gap-2 cursor-pointer text-sm">
                <input
                  type="radio"
                  value="one_time"
                  :checked="strategy === 'one_time'"
                  class="accent-primary size-4"
                  @change="strategy = 'one_time'"
                />
                One-time
              </label>
              <label class="flex items-center gap-2 cursor-pointer text-sm">
                <input
                  type="radio"
                  value="cron"
                  :checked="strategy === 'cron'"
                  class="accent-primary size-4"
                  @change="strategy = 'cron'"
                />
                Cron
              </label>
            </div>
          </UFormField>
        </div>

        <UFormField label="Description">
          <UTextarea
            v-model="description"
            placeholder="Short description of the maintenance"
            :rows="3"
          />
        </UFormField>

        <!-- ONE-TIME body -->
        <template v-if="strategy === 'one_time'">
          <div class="grid grid-cols-2 gap-5">
            <UFormField>
              <template #label> <span class="text-error">*</span> Start at </template>
              <UInput
                v-model="startAt"
                type="datetime-local"
                placeholder="Select start date and time"
              />
            </UFormField>
            <UFormField label="End at">
              <UInput
                v-model="endAt"
                type="datetime-local"
                placeholder="Select end date and time"
              />
            </UFormField>
          </div>
        </template>

        <!-- CRON body -->
        <template v-else>
          <section class="space-y-3">
            <h3 class="text-sm font-semibold text-default">Start time (time of day)</h3>
            <div class="grid grid-cols-2 gap-5">
              <UFormField>
                <template #label> <span class="text-error">*</span> Time </template>
                <UInput v-model="cronTime" type="time" />
              </UFormField>
              <UFormField label="Timezone">
                <USelectMenu v-model="cronTimezone" :items="timezones" searchable />
              </UFormField>
            </div>
          </section>

          <section class="space-y-2">
            <h3 class="text-sm font-semibold text-default">How often should it run?</h3>
            <div class="inline-flex rounded-md bg-elevated p-0.5">
              <button
                v-for="m in [
                  { value: 'weekdays', label: 'Days of week' },
                  { value: 'monthdays', label: 'Days of month' },
                  { value: 'first_last', label: 'First/Last day' },
                ]"
                :key="m.value"
                type="button"
                class="px-3 py-1.5 text-xs rounded-md transition-colors"
                :class="
                  cronMode === m.value
                    ? 'bg-primary text-inverted font-medium'
                    : 'text-muted hover:text-default'
                "
                @click="cronMode = m.value as 'weekdays' | 'monthdays' | 'first_last'"
              >
                {{ m.label }}
              </button>
            </div>

            <template v-if="cronMode === 'weekdays'">
              <p class="text-sm font-medium text-default">Weekdays</p>
              <div class="flex flex-wrap gap-3">
                <label
                  v-for="(d, i) in weekdayLabels"
                  :key="d"
                  class="flex items-center gap-1.5 text-sm cursor-pointer"
                >
                  <input
                    type="checkbox"
                    class="accent-primary size-4"
                    :checked="weekdays[i]"
                    @change="toggleWeekday(i)"
                  />
                  {{ d }}
                </label>
              </div>
            </template>

            <template v-else-if="cronMode === 'monthdays'">
              <p class="text-sm font-medium text-default">Days of month</p>
              <USelectMenu
                v-model="monthDays"
                multiple
                :items="Array.from({ length: 31 }, (_, i) => i + 1)"
              />
            </template>

            <template v-else>
              <p class="text-sm font-medium text-default">Pick</p>
              <div class="flex gap-3">
                <label class="flex items-center gap-2 text-sm cursor-pointer">
                  <input
                    type="radio"
                    class="accent-primary"
                    :checked="firstLast === 'first'"
                    @change="firstLast = 'first'"
                  />
                  First day of month
                </label>
                <label class="flex items-center gap-2 text-sm cursor-pointer">
                  <input
                    type="radio"
                    class="accent-primary"
                    :checked="firstLast === 'last'"
                    @change="firstLast = 'last'"
                  />
                  Last day of month
                </label>
              </div>
            </template>
          </section>

          <section class="space-y-2">
            <div class="flex items-center justify-between">
              <h3 class="text-sm font-semibold text-default">Generated cron expression</h3>
              <button
                type="button"
                class="text-xs text-primary hover:underline"
                @click="editingCronManually = !editingCronManually"
              >
                {{ editingCronManually ? 'Use builder' : 'Edit cron manually' }}
              </button>
            </div>
            <UInput
              v-model="cronExpr"
              placeholder="m h dom mon dow"
              :disabled="!editingCronManually"
            />
            <p class="text-xs text-muted">
              {{ cronIsValid ? cronHumanReadable : 'Invalid cron expression' }}
              <template v-if="editingCronManually">
                · Advanced users can edit the cron expression directly.</template
              >
            </p>
          </section>

          <section class="space-y-2">
            <h3 class="text-sm font-semibold text-default">Duration (per occurrence)</h3>
            <UFormField>
              <template #label> <span class="text-error">*</span> Minutes </template>
              <UInput
                v-model.number="durationMinutes"
                type="number"
                :min="5"
                :max="1440"
                class="max-w-xs"
              />
            </UFormField>
          </section>

          <section class="rounded-md border border-default/40 bg-elevated/50 p-3 space-y-2">
            <button
              type="button"
              class="flex items-center gap-1.5 text-sm font-medium text-default w-full"
              @click="showAdvanced = !showAdvanced"
            >
              <UIcon
                :name="showAdvanced ? 'i-lucide-chevron-down' : 'i-lucide-chevron-right'"
                class="size-4"
              />
              Advanced options (optional)
            </button>
            <template v-if="showAdvanced">
              <p class="text-xs text-muted">
                Limits when this recurring schedule can run. Leave empty to run indefinitely.
              </p>
              <div class="grid grid-cols-2 gap-5">
                <UFormField label="Effective from">
                  <UInput v-model="effectiveFrom" type="date" placeholder="Start date (optional)" />
                </UFormField>
                <UFormField label="Effective until">
                  <UInput v-model="effectiveUntil" type="date" placeholder="End date (optional)" />
                </UFormField>
              </div>
            </template>
          </section>
        </template>

        <UFormField label="Resources">
          <USelectMenu
            v-model="selectedResourceIds"
            multiple
            placeholder="Select impacted resources"
            :items="resources.map((r) => ({ label: r.name, value: r.id }))"
          />
        </UFormField>
      </div>
    </template>

    <template #footer>
      <div class="flex justify-end gap-2 w-full">
        <UButton variant="outline" color="neutral" @click="close">Cancel</UButton>
        <UButton color="primary" :disabled="!canSubmit" @click="onSubmit">Create</UButton>
      </div>
    </template>
  </UModal>
</template>
