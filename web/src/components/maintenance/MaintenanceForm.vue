<script setup lang="ts">
// eslint-disable-next-line @typescript-eslint/ban-ts-comment
// @ts-nocheck — legacy AntDV file, migrated in later Slices.
import { computed, reactive, ref, watch } from 'vue'
import dayjs, { type Dayjs } from 'dayjs'

import type { CreateMaintenance, MaintenanceStrategy, UpdateMaintenance } from '@/types'
import CronGenerator from './CronGenerator.vue'

interface ResourceOption {
  label: string
  value: string
}

const props = defineProps<{
  mode: 'create' | 'edit'
  initialData?: Partial<CreateMaintenance>
  resourceOptions: ResourceOption[]
}>()

const emit = defineEmits<{
  (e: 'submit', payload: CreateMaintenance | UpdateMaintenance): void
  (e: 'cancel'): void
}>()

const formRef = ref()
const cronValid = ref(true)

const formState = reactive({
  title: '',
  description: '',
  strategy: 'one_time' as MaintenanceStrategy,
  start_at: null as Dayjs | null,
  end_at: null as Dayjs | null,
  cron_expr: '',
  window_minutes: null as number | null,
  timezone: '',
  effective_from: null as Dayjs | null,
  effective_until: null as Dayjs | null,
  resource_ids: [] as string[],
})

const isCron = computed(() => formState.strategy === 'cron')
const isOneTime = computed(() => formState.strategy === 'one_time')

const applyInitialData = () => {
  if (!props.initialData) return
  formState.title = props.initialData.title || ''
  formState.description = props.initialData.description || ''
  formState.strategy = (props.initialData.strategy as MaintenanceStrategy) || 'one_time'
  formState.start_at = props.initialData.start_at ? dayjs(props.initialData.start_at) : null
  formState.end_at = props.initialData.end_at ? dayjs(props.initialData.end_at) : null
  formState.cron_expr = props.initialData.cron_expr || ''
  formState.window_minutes = props.initialData.window_minutes ?? null
  formState.timezone = props.initialData.timezone || ''
  formState.effective_from = props.initialData.effective_from
    ? dayjs(props.initialData.effective_from)
    : null
  formState.effective_until = props.initialData.effective_until
    ? dayjs(props.initialData.effective_until)
    : null
  formState.resource_ids = props.initialData.resource_ids || []
}

watch(
  () => props.initialData,
  () => applyInitialData(),
  { immediate: true },
)

const resetNonStrategyFields = (strategy: MaintenanceStrategy) => {
  if (strategy === 'one_time') {
    formState.cron_expr = ''
    formState.window_minutes = null
    formState.timezone = ''
    formState.effective_from = null
    formState.effective_until = null
  } else {
    formState.start_at = null
    formState.end_at = null
  }
}

const handleStrategyChange = (value: MaintenanceStrategy) => {
  formState.strategy = value
  resetNonStrategyFields(value)
}

const buildPayload = (): CreateMaintenance | UpdateMaintenance => {
  const base: CreateMaintenance = {
    title: formState.title,
    description: formState.description || undefined,
    strategy: formState.strategy,
    resource_ids: formState.resource_ids,
  }

  if (formState.strategy === 'one_time') {
    if (formState.start_at) {
      base.start_at = formState.start_at.toISOString()
    }
    if (formState.end_at) {
      base.end_at = formState.end_at.toISOString()
    }
    base.cron_expr = undefined
    base.window_minutes = undefined
    base.timezone = undefined
  } else {
    base.start_at = undefined
    base.end_at = undefined
    base.cron_expr = formState.cron_expr || undefined
    base.window_minutes = formState.window_minutes ?? undefined
    base.timezone = formState.timezone || undefined
    base.effective_from = formState.effective_from
      ? formState.effective_from.toISOString()
      : undefined
    base.effective_until = formState.effective_until
      ? formState.effective_until.toISOString()
      : undefined
  }

  return base
}

const handleSubmit = async () => {
  await formRef.value?.validate()
  emit('submit', buildPayload())
}

function onStrategyChange(e: Event) {
  const target = e.target as HTMLInputElement
  handleStrategyChange(target.value as MaintenanceStrategy)
}
</script>

<template>
  <a-form ref="formRef" layout="vertical" :model="formState">
    <a-row :gutter="16">
      <a-col :span="12">
        <a-form-item
          name="title"
          label="Title"
          :rules="[{ required: true, message: 'Title is required' }]"
        >
          <a-input v-model:value="formState.title" placeholder="e.g., Database upgrade" />
        </a-form-item>
      </a-col>
      <a-col :span="12">
        <a-form-item name="strategy" label="Strategy" :rules="[{ required: true }]">
          <a-radio-group
            v-model:value="formState.strategy"
            :options="[
              { label: 'One-time', value: 'one_time' },
              { label: 'Cron', value: 'cron' },
            ]"
            @change="onStrategyChange"
          />
        </a-form-item>
      </a-col>
    </a-row>

    <a-form-item name="description" label="Description">
      <a-textarea
        v-model:value="formState.description"
        rows="3"
        placeholder="Short description of the maintenance"
      />
    </a-form-item>

    <a-row :gutter="16" v-if="isOneTime">
      <a-col :span="12">
        <a-form-item
          name="start_at"
          label="Start at"
          :rules="[{ required: true, message: 'Start time is required' }]"
        >
          <a-date-picker
            v-model:value="formState.start_at"
            show-time
            style="width: 100%"
            format="YYYY-MM-DD HH:mm"
            placeholder="Select start date and time"
          />
        </a-form-item>
      </a-col>
      <a-col :span="12">
        <a-form-item name="end_at" label="End at">
          <a-date-picker
            v-model:value="formState.end_at"
            show-time
            style="width: 100%"
            format="YYYY-MM-DD HH:mm"
            placeholder="Select end date and time"
          />
        </a-form-item>
      </a-col>
    </a-row>

    <template v-if="isCron">
      <!-- Cron schedule configuration -->
      <a-form-item name="cron_expr" :rules="[{ required: true, message: 'Schedule is required' }]">
        <CronGenerator
          v-model="formState.cron_expr"
          v-model:valid="cronValid"
          v-model:timezone="formState.timezone"
        />
      </a-form-item>

      <!-- Duration per occurrence -->
      <a-divider />
      <div style="font-weight: 600; margin-bottom: 8px">Duration (per occurrence)</div>
      <a-row :gutter="16">
        <a-col :span="8">
          <a-form-item
            name="window_minutes"
            label="Minutes"
            :rules="[{ required: true, message: 'Duration is required' }]"
          >
            <a-input-number
              v-model:value="formState.window_minutes"
              :min="1"
              :max="1440"
              style="width: 100%"
              placeholder="60"
            />
          </a-form-item>
        </a-col>
      </a-row>

      <!-- Advanced: Effective date range (optional) -->
      <a-collapse style="margin-top: 16px">
        <a-collapse-panel key="advanced" header="Advanced options (optional)">
          <div style="margin-bottom: 8px; color: #6b7280; font-size: 13px">
            Limits when this recurring schedule can run. Leave empty to run indefinitely.
          </div>
          <a-row :gutter="16">
            <a-col :span="12">
              <a-form-item name="effective_from" label="Effective from">
                <a-date-picker
                  v-model:value="formState.effective_from"
                  style="width: 100%"
                  placeholder="Start date (optional)"
                />
              </a-form-item>
            </a-col>
            <a-col :span="12">
              <a-form-item name="effective_until" label="Effective until">
                <a-date-picker
                  v-model:value="formState.effective_until"
                  style="width: 100%"
                  placeholder="End date (optional)"
                />
              </a-form-item>
            </a-col>
          </a-row>
        </a-collapse-panel>
      </a-collapse>
    </template>

    <a-form-item name="resource_ids" label="Resources">
      <a-select
        v-model:value="formState.resource_ids"
        mode="multiple"
        :options="resourceOptions"
        placeholder="Select impacted resources"
        allow-clear
        show-search
        option-filter-prop="label"
      />
    </a-form-item>

    <div style="display: flex; justify-content: flex-end; gap: 8px">
      <a-button @click="emit('cancel')">Cancel</a-button>
      <a-button type="primary" :disabled="isCron && !cronValid" @click="handleSubmit">
        {{ mode === 'create' ? 'Create' : 'Save changes' }}
      </a-button>
    </div>
  </a-form>
</template>

<style scoped></style>
