<script setup lang="ts">
import { reactive, watch } from 'vue'
import type { SlackConfig } from '@/types'

interface Props {
  modelValue: Partial<SlackConfig>
}

interface Emits {
  (e: 'update:modelValue', value: Partial<SlackConfig>): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const form = reactive<Partial<SlackConfig>>({
  webhook_url: '',
  channel: '',
  username: '',
  ...props.modelValue,
})

watch(
  form,
  (newValue) => {
    emit('update:modelValue', newValue)
  },
  { deep: true },
)
</script>

<template>
  <div>
    <!-- Webhook URL -->
    <a-form-item
      label="Webhook URL"
      :name="['config', 'webhook_url']"
      :rules="[{ required: true, message: 'webhook_url is required' }]"
    >
      <a-input
        v-model:value="form.webhook_url"
        type="password"
        placeholder="https://hooks.slack.com/services/YOUR/WEBHOOK/URL"
        :maxlength="500"
      />
    </a-form-item>

    <!-- Channel (Optional) -->
    <a-form-item label="Channel (Optional)" :name="['config', 'channel']">
      <a-input v-model:value="form.channel" placeholder="#alerts or @username" :maxlength="255" />
    </a-form-item>

    <!-- Username (Optional) -->
    <a-form-item label="Bot Username (Optional)" :name="['config', 'username']">
      <a-input v-model:value="form.username" placeholder="e.g., Ogoune Bot" :maxlength="255" />
    </a-form-item>
  </div>
</template>

<style scoped></style>
