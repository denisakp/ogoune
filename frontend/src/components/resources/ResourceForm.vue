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

const { addResource, updateResourceData } = useResources()

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
})

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
      }
    }
  },
  { immediate: true },
)

const handleSubmit = async () => {
  if (!form.value.name.trim()) {
    return
  }
  if (!form.value.target.trim()) {
    return
  }
  if (form.value.interval < 10) {
    return
  }
  if (form.value.timeout < 1) {
    return
  }
  if ((form.value.confirmation_checks ?? 0) <= 0) {
    return
  }
  if ((form.value.confirmation_interval ?? 0) <= 0) {
    return
  }
  if ((form.value.confirmation_interval ?? 0) >= form.value.interval) {
    return
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
})

const tagsOptions = computed(() => tags.value.map((tag) => ({ value: tag.name, label: tag.name })))

const componentOptions = computed(() => [
  { value: undefined, label: '⊘ No component (standalone resource)' },
  ...components.value.map((c) => ({ value: c.id, label: c.name })),
])
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
          </a-select>
        </a-form-item>
      </a-col>

      <a-col :xs="24" :sm="12">
        <a-form-item label="Target" required>
          <a-input
            v-model:value="form.target"
            :placeholder="
              form.type === 'dns'
                ? 'e.g., example.com or 8.8.8.8'
                : 'e.g., https://example.com or example.com:8080'
            "
          />
        </a-form-item>
      </a-col>
    </a-row>

    <!-- Interval -->
    <a-form-item label="Check Interval (seconds)">
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

    <!-- Timeout -->
    <a-form-item label="Timeout (seconds)">
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

    <a-row :gutter="16">
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

    <!-- Submit Buttons -->
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
