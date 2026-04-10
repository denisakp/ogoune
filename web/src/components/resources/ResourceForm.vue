<script setup lang="ts">
import { ref, watch, onMounted, computed } from 'vue'
import { FolderOutlined } from '@ant-design/icons-vue'

import type { Resource, CreateResource } from '@/types'
import { useResources } from '@/composables/useResources.ts'
import { useTags } from '@/composables/useTags.ts'
import { useComponents } from '@/composables/useComponents.ts'

interface Props {
  resource?: Resource
}

const props = defineProps<Props>()
const emit = defineEmits<{ submit: [] }>()

const { addResource, updateResourceData, capabilities, loadCapabilities } = useResources()

const form = ref<CreateResource & { component_id?: string }>({
  name: '',
  type: 'http',
  target: '',
  interval: 300,
  timeout: 10,
  confirmation_checks: 2,
  confirmation_interval: 30,
  tags: [],
  component_id: undefined,
  expiry_alert_thresholds: undefined,
  flap_detection_enabled: undefined,
  flap_threshold: undefined,
  flap_window_seconds: undefined,
  flap_max_duration_minutes: undefined,
  reminder_interval_minutes: undefined,
  heartbeat_interval: 300,
  heartbeat_grace: 60,
  keyword: undefined,
  keyword_mode: 'contains' as 'contains' | 'not_contains',
})

const isHeartbeat = computed(() => form.value.type === 'heartbeat')
const isKeyword = computed(() => form.value.type === 'keyword')

const loading = ref(false)

watch(
  () => props.resource,
  (resource) => {
    if (resource) {
      form.value = {
        name: resource.name,
        type: resource.type,
        target: resource.target,
        interval: resource.interval,
        timeout: resource.timeout,
        confirmation_checks: resource.confirmation_checks,
        confirmation_interval: resource.confirmation_interval,
        tags: (resource.tags ?? []).map((t) => t.name),
        component_id: resource.component_id || undefined,
        expiry_alert_thresholds: resource.expiry_alert_thresholds ?? undefined,
        flap_detection_enabled: resource.flap_detection_enabled,
        flap_threshold: resource.flap_threshold,
        flap_window_seconds: resource.flap_window_seconds,
        flap_max_duration_minutes: resource.flap_max_duration_minutes,
        reminder_interval_minutes: resource.reminder_interval_minutes,
        heartbeat_interval: resource.heartbeat_interval ?? 300,
        heartbeat_grace: resource.heartbeat_grace ?? 60,
        keyword: resource.keyword ?? undefined,
        keyword_mode: (resource.keyword_mode as 'contains' | 'not_contains') ?? 'contains',
      }
    } else {
      form.value = {
        name: '',
        type: 'http',
        target: '',
        interval: 300,
        timeout: 10,
        confirmation_checks: 2,
        confirmation_interval: 30,
        tags: [],
        component_id: undefined,
        expiry_alert_thresholds: undefined,
        flap_detection_enabled: undefined,
        flap_threshold: undefined,
        flap_window_seconds: undefined,
        flap_max_duration_minutes: undefined,
        reminder_interval_minutes: undefined,
        heartbeat_interval: 300,
        heartbeat_grace: 60,
        keyword: undefined,
        keyword_mode: 'contains',
      }
    }
  },
  { immediate: true },
)

const handleSubmit = async () => {
  if (!form.value.name.trim()) {
    return
  }
  if (isHeartbeat.value) {
    const hbInterval = form.value.heartbeat_interval ?? 0
    const hbGrace = form.value.heartbeat_grace ?? 0
    if (hbInterval < 60 || hbInterval > 86400) return
    if (hbGrace < 60 || hbGrace > 3600) return
  } else {
    if (!form.value.target?.trim()) {
      return
    }
    if ((form.value.interval ?? 0) < 10) {
      return
    }
    if ((form.value.timeout ?? 0) < 1) {
      return
    }
    if ((form.value.confirmation_checks ?? 0) <= 0) {
      return
    }
    if ((form.value.confirmation_interval ?? 0) <= 0) {
      return
    }
    if ((form.value.confirmation_interval ?? 0) >= (form.value.interval ?? 0)) {
      return
    }
  }
  if (isKeyword.value) {
    if (!form.value.keyword?.trim()) return
    if ((form.value.keyword?.length ?? 0) > 500) return
  }
  const thresholds = form.value.expiry_alert_thresholds?.trim()
  if (thresholds) {
    const parts = thresholds.split(',').map((s) => s.trim())
    const invalid = parts.some((p) => !/^\d+$/.test(p) || +p < 1 || +p > 365)
    if (invalid) {
      return
    }
  }

  loading.value = true

  try {
    if (props.resource) {
      const updateData = {
        name: form.value.name,
        type: form.value.type,
        target: form.value.target,
        interval: form.value.interval,
        timeout: form.value.timeout,
        confirmation_checks: form.value.confirmation_checks,
        confirmation_interval: form.value.confirmation_interval,
        tags: form.value.tags,
        component_id: form.value.component_id,
        expiry_alert_thresholds: form.value.expiry_alert_thresholds || undefined,
        flap_detection_enabled: form.value.flap_detection_enabled,
        flap_threshold: form.value.flap_threshold,
        flap_window_seconds: form.value.flap_window_seconds,
        flap_max_duration_minutes: form.value.flap_max_duration_minutes,
        reminder_interval_minutes: form.value.reminder_interval_minutes,
        heartbeat_interval: isHeartbeat.value ? form.value.heartbeat_interval : undefined,
        heartbeat_grace: isHeartbeat.value ? form.value.heartbeat_grace : undefined,
        keyword: isKeyword.value ? form.value.keyword : undefined,
        keyword_mode: isKeyword.value ? form.value.keyword_mode : undefined,
      }
      await updateResourceData(props.resource.id, updateData)
    } else {
      await addResource(form.value)
    }
    emit('submit')
  } catch {
  } finally {
    loading.value = false
  }
}

const { tags, loadTags } = useTags()
const { components, loadComponents } = useComponents()

onMounted(() => {
  loadTags()
  loadComponents()
  loadCapabilities()
})

const tagsOptions = computed(() => tags.value.map((tag) => ({ value: tag.name, label: tag.name })))

const componentOptions = computed(() => [
  { value: undefined, label: '⊘ No component (standalone resource)' },
  ...components.value.map((c) => ({ value: c.id, label: c.name })),
])

const targetPlaceholder = computed(() => {
  switch (form.value.type) {
    case 'dns':
      return 'e.g., example.com or 8.8.8.8'
    case 'icmp':
      return 'e.g., hostname or IP address (e.g., 8.8.8.8)'
    default:
      return 'e.g., https://example.com or example.com:8080'
  }
})

const icmpWarning = computed(() => {
  if (form.value.type !== 'icmp') return null
  if (!capabilities.value) return null
  const icmp = capabilities.value.icmp
  if (!icmp.enabled) {
    return {
      type: 'error' as const,
      message: 'ICMP monitoring is disabled (set ENABLE_ICMP=true to enable)',
    }
  }
  if (!icmp.capability_available) {
    return {
      type: 'warning' as const,
      message: `ICMP capability unavailable: ${icmp.reason || 'requires elevated privileges (root or CAP_NET_RAW)'}`,
    }
  }
  return null
})
</script>

<template>
  <a-form layout="vertical">
    <!-- Name -->
    <a-form-item label="Monitor Name" required>
      <a-input
        v-model:value="form.name"
        placeholder="e.g., My Website"
        @keyup.enter="handleSubmit"
      />
    </a-form-item>

    <!-- Type & Target Row -->
    <a-row :gutter="16">
      <a-col :xs="24" :sm="12">
        <a-form-item label="Type" required>
          <a-select v-model:value="form.type">
            <a-select-option value="http">HTTP/HTTPS</a-select-option>
            <a-select-option value="tcp">TCP</a-select-option>
            <a-select-option value="dns">DNS</a-select-option>
            <a-select-option value="icmp">ICMP (Ping)</a-select-option>
            <a-select-option value="heartbeat">Heartbeat / Push</a-select-option>
            <a-select-option value="keyword">Keyword / Content Check</a-select-option>
          </a-select>
        </a-form-item>
      </a-col>

      <a-col v-if="!isHeartbeat" :xs="24" :sm="12">
        <a-form-item label="Target" required>
          <a-input
            v-model:value="form.target"
            :placeholder="targetPlaceholder"
          />
        </a-form-item>
      </a-col>
    </a-row>

    <!-- Heartbeat fields -->
    <template v-if="isHeartbeat">
      <a-row :gutter="16">
        <a-col :xs="24" :sm="12">
          <a-form-item label="Ping Interval (seconds)" required>
            <a-input-number
              v-model:value="form.heartbeat_interval"
              :min="60"
              :max="86400"
              style="width: 100%"
              placeholder="e.g. 300 (5 min)"
              data-testid="heartbeat-interval"
            />
          </a-form-item>
        </a-col>
        <a-col :xs="24" :sm="12">
          <a-form-item label="Grace Period (seconds)" required>
            <a-input-number
              v-model:value="form.heartbeat_grace"
              :min="60"
              :max="3600"
              style="width: 100%"
              placeholder="e.g. 60 (1 min)"
              data-testid="heartbeat-grace"
            />
          </a-form-item>
        </a-col>
      </a-row>
    </template>

    <!-- Keyword fields -->
    <template v-if="isKeyword">
      <a-row :gutter="16">
        <a-col :xs="24" :sm="16">
          <a-form-item label="Keyword" required>
            <a-input
              v-model:value="form.keyword"
              placeholder="e.g., operational"
              :maxlength="500"
              data-testid="keyword-input"
            />
          </a-form-item>
        </a-col>
        <a-col :xs="24" :sm="8">
          <a-form-item label="Match Mode">
            <a-select v-model:value="form.keyword_mode" data-testid="keyword-mode-select">
              <a-select-option value="contains">Contains</a-select-option>
              <a-select-option value="not_contains">Does Not Contain</a-select-option>
            </a-select>
          </a-form-item>
        </a-col>
      </a-row>
    </template>

    <!-- ICMP availability warning -->
    <a-alert
      v-if="icmpWarning"
      style="margin-bottom: 16px"
      :type="icmpWarning.type"
      show-icon
      :message="icmpWarning.message"
    />

    <!-- Interval (non-heartbeat only) -->
    <a-form-item v-if="!isHeartbeat" label="Check Interval (seconds)">
      <div style="display: flex; align-items: center; gap: 16px">
        <a-slider
          v-model:value="form.interval"
          :min="10"
          :max="3600"
          :step="10"
          style="flex: 1"
          :marks="{ 10: '10s', 300: '5m', 600: '10m', 3600: '1h' }"
        />
        <div style="min-width: 60px; text-align: right; font-weight: bold">
          {{ form.interval }}s
        </div>
      </div>
    </a-form-item>

    <!-- Timeout (non-heartbeat only) -->
    <a-form-item v-if="!isHeartbeat" label="Timeout (seconds)">
      <div style="display: flex; align-items: center; gap: 16px">
        <a-slider
          v-model:value="form.timeout"
          :min="1"
          :max="60"
          :step="1"
          style="flex: 1"
          :marks="{ 1: '1s', 10: '10s', 30: '30s', 60: '60s' }"
        />
        <div style="min-width: 60px; text-align: right; font-weight: bold">{{ form.timeout }}s</div>
      </div>
    </a-form-item>

    <a-row v-if="!isHeartbeat" :gutter="16">
      <a-col :xs="24" :sm="12">
        <a-form-item label="Confirmation Checks">
          <a-input-number
            v-model:value="form.confirmation_checks"
            :min="1"
            :max="20"
            style="width: 100%"
          />
        </a-form-item>
      </a-col>
      <a-col :xs="24" :sm="12">
        <a-form-item label="Confirmation Interval (seconds)">
          <a-input-number
            v-model:value="form.confirmation_interval"
            :min="1"
            :max="3600"
            style="width: 100%"
          />
        </a-form-item>
      </a-col>
    </a-row>
    <a-alert
      v-if="!isHeartbeat"
      style="margin-bottom: 16px"
      type="info"
      show-icon
      :message="
        form.confirmation_checks === 1
          ? 'Immediate mode enabled: incident is created on first failure.'
          : 'Confirmation mode: incidents trigger only after consecutive failures reach the threshold.'
      "
      description="confirmation_interval must be lower than the regular check interval."
    />

    <!-- Tags selection -->
    <a-form-item label="Tags">
      <a-select
        v-model:value="form.tags"
        mode="tags"
        style="width: 100%"
        placeholder="Tags"
        :options="tagsOptions"
      >
      </a-select>
    </a-form-item>

    <!-- Component Assignment -->
    <a-form-item label="Component (Optional)">
      <a-select
        v-model:value="form.component_id"
        allow-clear
        style="width: 100%"
        placeholder="Assign to a component group"
        :options="componentOptions"
      >
        <template #suffixIcon>
          <FolderOutlined />
        </template>
      </a-select>
      <div style="margin-top: 8px; font-size: 12px; color: rgba(0, 0, 0, 0.45)">
        💡 Group related resources together (e.g., "Frontend Services", "API Servers"). A resource
        can belong to only one component.
      </div>
    </a-form-item>

    <!-- Expiry Alert Thresholds (HTTP only) -->
    <a-form-item v-if="form.type === 'http'" label="Custom Expiry Alert Thresholds (days)">
      <a-input
        v-model:value="form.expiry_alert_thresholds"
        placeholder="e.g. 30,14,7,1"
        allow-clear
      />
      <div style="margin-top: 8px; font-size: 12px; color: rgba(0, 0, 0, 0.45)">
        🔒 Comma-separated days before SSL/domain expiry to send alerts (each value 1–365). Leave
        blank to use global defaults.
      </div>
    </a-form-item>

    <!-- Submit Buttons -->
    <!-- Smart Alerting -->
    <a-collapse style="margin-bottom: 24px" :bordered="false">
      <a-collapse-panel key="smart-alerting" header="⚡ Smart Alerting (Optional)">
        <a-form-item label="Flap Detection">
          <a-switch v-model:checked="form.flap_detection_enabled" />
          <span style="margin-left: 8px; font-size: 13px; color: rgba(0, 0, 0, 0.45)">
            Suppress repeated alerts when the service oscillates between UP and DOWN
          </span>
        </a-form-item>

        <template v-if="form.flap_detection_enabled">
          <a-row :gutter="16">
            <a-col :xs="24" :sm="8">
              <a-form-item label="Flap Threshold (transitions)">
                <a-input-number
                  v-model:value="form.flap_threshold"
                  :min="2"
                  :max="20"
                  style="width: 100%"
                  placeholder="e.g. 3"
                />
              </a-form-item>
            </a-col>
            <a-col :xs="24" :sm="8">
              <a-form-item label="Flap Window (seconds)">
                <a-input-number
                  v-model:value="form.flap_window_seconds"
                  :min="60"
                  :max="3600"
                  style="width: 100%"
                  placeholder="e.g. 300"
                />
              </a-form-item>
            </a-col>
            <a-col :xs="24" :sm="8">
              <a-form-item label="Max Flap Duration (minutes, 0=unlimited)">
                <a-input-number
                  v-model:value="form.flap_max_duration_minutes"
                  :min="0"
                  style="width: 100%"
                  placeholder="0"
                />
              </a-form-item>
            </a-col>
          </a-row>
        </template>

        <a-form-item label="Reminder Interval (minutes, 0=disabled)">
          <a-input-number
            v-model:value="form.reminder_interval_minutes"
            :min="0"
            style="width: 100%"
            placeholder="0"
          />
          <div style="margin-top: 4px; font-size: 12px; color: rgba(0, 0, 0, 0.45)">
            Send a reminder notification if an incident remains open for this many minutes.
          </div>
        </a-form-item>
      </a-collapse-panel>
    </a-collapse>

    <a-form-item style="margin-top: 24px">
      <a-space>
        <a-button @click="() => emit('submit')">Cancel</a-button>
        <a-button type="primary" :loading="loading" @click="handleSubmit">
          {{ props.resource ? 'Update Monitor' : 'Create Monitor' }}
        </a-button>
      </a-space>
    </a-form-item>
  </a-form>
</template>

<style scoped></style>
