<script setup lang="ts">
/* eslint-disable @typescript-eslint/ban-ts-comment */
// @ts-nocheck — spec 059 polish debt: NuxtUI v4 template-emit types
interface SlackConfig {
  webhook_url: string
  channel: string
  display_name?: string
}
interface Props {
  modelValue: SlackConfig
  fieldErrors?: Record<string, string>
}
const props = withDefaults(defineProps<Props>(), { fieldErrors: () => ({}) })
const emit = defineEmits<{ (e: 'update:modelValue', v: SlackConfig): void }>()

function update<K extends keyof SlackConfig>(key: K, value: SlackConfig[K]) {
  emit('update:modelValue', { ...props.modelValue, [key]: value })
}
</script>

<template>
  <div class="space-y-3">
    <UFormField
      label="Slack incoming-webhook URL"
      name="config.webhook_url"
      :error="fieldErrors['config.webhook_url']"
    >
      <UInput
class="w-full"         :model-value="modelValue.webhook_url"
        placeholder="https://hooks.slack.com/services/T/B/X"
        @update:model-value="(v) => update('webhook_url', String(v))"
      />
    </UFormField>
    <UFormField label="Channel" name="config.channel" :error="fieldErrors['config.channel']">
      <UInput
class="w-full"         :model-value="modelValue.channel"
        placeholder="oncall"
        @update:model-value="(v) => update('channel', String(v))"
      />
    </UFormField>
    <UFormField
      label="Display name (optional)"
      name="config.display_name"
      :error="fieldErrors['config.display_name']"
    >
      <UInput
class="w-full"         :model-value="modelValue.display_name ?? ''"
        placeholder="Ogoune bot"
        @update:model-value="(v) => update('display_name', String(v))"
      />
    </UFormField>
  </div>
</template>
