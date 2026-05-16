<script setup lang="ts">
import { reactive, watch, computed, ref } from 'vue'
import dayjs from 'dayjs'

// v-model support
const props = defineProps<{
  modelValue: string
  valid?: boolean
  timezone?: string
}>()
const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
  (e: 'update:valid', value: boolean): void
  (e: 'update:timezone', value: string): void
}>()

// Internal state
type Mode = 'weekdays' | 'monthdays' | 'first_last'

const state = reactive({
  mode: 'weekdays' as Mode,
  hour: 2,
  minute: 0,
  // Weekdays: Monday=1 ... Sunday=0
  weekdays: [] as number[],
  // Days of month: 1..31
  monthdays: [] as number[],
  // First/Last day option
  firstOrLast: 'first' as 'first' | 'last',
  raw: props.modelValue || '',
  parseOk: true,
})

const timeValue = ref(dayjs().hour(state.hour).minute(state.minute))

const minute = () => state.minute
const hour = () => state.hour

function dowToLabelMap() {
  return [
    { label: 'Sunday', value: 0 },
    { label: 'Monday', value: 1 },
    { label: 'Tuesday', value: 2 },
    { label: 'Wednesday', value: 3 },
    { label: 'Thursday', value: 4 },
    { label: 'Friday', value: 5 },
    { label: 'Saturday', value: 6 },
  ]
}

const weekdayOptions = dowToLabelMap()
const monthdayOptions = Array.from({ length: 31 }, (_, i) => ({
  label: String(i + 1),
  value: i + 1,
}))
const timezoneOptions = [{ label: 'UTC', value: 'UTC' }]

// Generate cron expression from UI state (Unix 5-field: m h dom mon dow)
function generateCron(): string {
  const m = String(minute())
  const h = String(hour())
  if (state.mode === 'weekdays') {
    const dows = state.weekdays.sort((a, b) => a - b).join(',') || '*'
    return `${m} ${h} * * ${dows || '*'}`
  }
  if (state.mode === 'monthdays') {
    const doms = state.monthdays.sort((a, b) => a - b).join(',') || '*'
    return `${m} ${h} ${doms || '*'} * *`
  }
  // first_last mode
  if (state.firstOrLast === 'first') {
    return `${m} ${h} 1 * *`
  }
  // Note: 'L' (last day) is Quartz-specific; backend dialect may not support it.
  return `${m} ${h} L * *`
}

// Attempt to parse a 5-field cron into UI state
function parseCron(expr: string) {
  const parts = expr.trim().split(/\s+/)
  if (parts.length !== 5) return // unsupported; keep raw
  const [mStr, hStr, dom, , dow] = parts
  const m = Number(mStr)
  const h = Number(hStr)
  if (!Number.isNaN(m) && m >= 0 && m <= 59) state.minute = m
  if (!Number.isNaN(h) && h >= 0 && h <= 23) state.hour = h

  if (dom === '*' && dow && dow !== '*') {
    // Weekdays mode
    state.mode = 'weekdays'
    state.weekdays = dow
      .split(',')
      .map((d) => Number(d))
      .filter((n) => !Number.isNaN(n))
    state.parseOk = true
    return
  }
  if (dow === '*' && dom && dom !== '*' && dom !== 'L') {
    // Monthdays mode
    state.mode = 'monthdays'
    state.monthdays = dom
      .split(',')
      .map((d) => Number(d))
      .filter((n) => !Number.isNaN(n) && n >= 1 && n <= 31)
    state.parseOk = true
    return
  }
  if (dow === '*' && (dom === '1' || dom === 'L')) {
    // First/Last mode
    state.mode = 'first_last'
    state.firstOrLast = dom === 'L' ? 'last' : 'first'
    state.parseOk = true
    return
  }
  // Fallback: leave raw only
  state.parseOk = false
}

// Keep state in sync with external modelValue
watch(
  () => props.modelValue,
  (val) => {
    state.raw = val || ''
    if (val) parseCron(val)
  },
  { immediate: true },
)

// Update modelValue whenever UI changes
watch(
  () => [
    state.mode,
    state.hour,
    state.minute,
    state.weekdays.slice(),
    state.monthdays.slice(),
    state.firstOrLast,
  ],
  () => {
    const expr = generateCron()
    state.raw = expr
    emit('update:modelValue', expr)
  },
)

// Keep time picker in sync when hour/min change (e.g., via parse)
watch(
  () => [state.hour, state.minute],
  () => {
    timeValue.value = dayjs().hour(state.hour).minute(state.minute)
  },
)

// Update hour/min when time picker changes
watch(timeValue, (val) => {
  if (!val) return
  state.hour = val.hour()
  state.minute = val.minute()
})

// When user edits raw cron manually, try to parse, else keep as raw
function onRawInput(val: string) {
  state.raw = val
  emit('update:modelValue', val)
  parseCron(val)
}

// Validity for selected mode
const isValidForMode = computed(() => {
  const hourOk = state.hour >= 0 && state.hour <= 23
  const minuteOk = state.minute >= 0 && state.minute <= 59
  if (!hourOk || !minuteOk) return false
  switch (state.mode) {
    case 'weekdays':
      return state.weekdays.length > 0
    case 'monthdays':
      return state.monthdays.length > 0 && state.monthdays.every((n) => n >= 1 && n <= 31)
    case 'first_last':
      return state.firstOrLast === 'first' || state.firstOrLast === 'last'
    default:
      return false
  }
})

watch(isValidForMode, (val) => emit('update:valid', val), { immediate: true })

const warningMessage = computed(() => {
  if (!state.parseOk && state.raw) {
    return 'Unrecognized cron for selected mode. Keeping raw expression. Adjust selections or enter a supported pattern.'
  }
  return ''
})

// Manual edit toggle (read-only by default)
const manualEdit = reactive({ enabled: false })
</script>

<template>
  <div style="display: grid; gap: 16px">
    <!-- Section: Start time per occurrence -->
    <div style="font-weight: 600">Start time (time of day)</div>
    <a-row :gutter="12">
      <a-col :span="12">
        <a-form-item label="Time" required>
          <a-time-picker v-model:value="timeValue" format="HH:mm" style="width: 100%" />
        </a-form-item>
      </a-col>
      <a-col :span="12">
        <a-form-item label="Timezone">
          <a-select
            :value="props.timezone"
            @update:value="(val: string) => emit('update:timezone', val || '')"
            :options="timezoneOptions"
            placeholder="UTC"
            allow-clear
          />
        </a-form-item>
      </a-col>
    </a-row>

    <!-- Section: How often should it run? -->
    <div style="font-weight: 600">How often should it run?</div>
    <a-form-item>
      <a-radio-group v-model:value="state.mode">
        <a-radio-button value="weekdays">Days of week</a-radio-button>
        <a-radio-button value="monthdays">Days of month</a-radio-button>
        <a-radio-button value="first_last">First/Last day</a-radio-button>
      </a-radio-group>
    </a-form-item>

    <!-- Mode-specific configuration -->
    <div v-if="state.mode === 'weekdays'">
      <a-form-item label="Weekdays">
        <a-checkbox-group v-model:value="state.weekdays">
          <div style="display: flex; gap: 8px; flex-wrap: wrap">
            <a-checkbox v-for="opt in weekdayOptions" :key="opt.value" :value="opt.value">{{
              opt.label
            }}</a-checkbox>
          </div>
        </a-checkbox-group>
      </a-form-item>
    </div>

    <div v-else-if="state.mode === 'monthdays'">
      <a-form-item label="Days of month">
        <a-select
          v-model:value="state.monthdays"
          mode="multiple"
          :options="monthdayOptions"
          placeholder="Select days (1-31)"
          style="width: 100%"
        />
      </a-form-item>
    </div>

    <div v-else>
      <a-form-item label="Day of month">
        <a-radio-group v-model:value="state.firstOrLast">
          <a-radio value="first">First day of the month</a-radio>
          <a-radio value="last">Last day of the month</a-radio>
        </a-radio-group>
      </a-form-item>
      <div style="color: #6b7280; font-size: 12px">
        Note: "Last day" uses Quartz-style 'L'; backend parser may not support it.
      </div>
    </div>

    <div style="display: flex; align-items: center; gap: 8px">
      <div style="font-weight: 600">Generated cron expression</div>
      <a-button type="link" size="small" @click="manualEdit.enabled = !manualEdit.enabled">
        {{ manualEdit.enabled ? 'Stop editing' : 'Edit cron manually' }}
      </a-button>
    </div>
    <a-form-item>
      <a-input
        :readonly="!manualEdit.enabled"
        :value="state.raw"
        @update:value="onRawInput"
        placeholder="m h dom mon dow"
      />
      <div style="color: #6b7280; font-size: 12px; margin-top: 4px">
        Advanced users can edit the cron expression directly.
      </div>
    </a-form-item>

    <div v-if="warningMessage" style="color: #d97706; font-size: 12px">
      {{ warningMessage }}
    </div>
  </div>
</template>

<style scoped></style>
