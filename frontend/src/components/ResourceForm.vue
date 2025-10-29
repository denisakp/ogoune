<script setup lang="ts">
import { ref, watch, onMounted, computed } from 'vue'
import { message } from 'ant-design-vue'

import type { Resource, CreateResource } from '@/types'
import { useResources } from '@/composables/useResources'
import { useTags } from '@/composables/useTags'

interface Props {
  resource?: Resource
}

const props = defineProps<Props>()
const emit = defineEmits<{ submit: [] }>()

const { addResource, updateResourceData } = useResources()

const form = ref<CreateResource>({
  name: '',
  type: 'http',
  target: '',
  interval: 300,
  timeout: 10,
  tags: [],
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
        tags: (resource.tags ?? []).map((t) => t.id),
      }
    } else {
      form.value = {
        name: '',
        type: 'http',
        target: '',
        interval: 300,
        timeout: 10,
        tags: [],
      }
    }
  },
  { immediate: true },
)

const handleSubmit = async () => {
  // Validation
  if (!form.value.name.trim()) {
    message.error('Name is required')
    return
  }
  if (!form.value.target.trim()) {
    message.error('Target is required')
    return
  }
  if (form.value.interval < 10) {
    message.error('Interval must be at least 10 seconds')
    return
  }
  if (form.value.timeout < 1) {
    message.error('Timeout must be at least 1 second')
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
        tags: form.value.tags,
      }
      await updateResourceData(props.resource.id, updateData)
    } else {
      await addResource(form.value)
    }
    emit('submit')
  } catch (err) {
    message.error(err instanceof Error ? err.message : 'An error occurred')
  } finally {
    loading.value = false
  }
}

const { tags, loadTags } = useTags()

onMounted(() => {
  loadTags()
})

const tagsOptions = computed(() => tags.value.map((tag) => ({ value: tag.id, label: tag.name })))
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
          </a-select>
        </a-form-item>
      </a-col>

      <a-col :xs="24" :sm="12">
        <a-form-item label="Target" required>
          <a-input
            v-model:value="form.target"
            placeholder="e.g., https://example.com or example.com:8080"
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
