<script setup lang="ts">
/* eslint-disable @typescript-eslint/ban-ts-comment */
// @ts-nocheck — spec 059 polish debt: NuxtUI v4 template-emit types
/**
 * SMTP / email channel form. Spec 059 US3.
 */
interface SmtpConfig {
  host: string
  port: number
  username: string
  password: string
  sender: string
  recipient: string
}

interface Props {
  modelValue: SmtpConfig
  fieldErrors?: Record<string, string>
}
const props = withDefaults(defineProps<Props>(), { fieldErrors: () => ({}) })
const emit = defineEmits<{ (e: 'update:modelValue', v: SmtpConfig): void }>()

function update<K extends keyof SmtpConfig>(key: K, value: SmtpConfig[K]) {
  emit('update:modelValue', { ...props.modelValue, [key]: value })
}
</script>

<template>
  <div class="space-y-3">
    <UFormField label="SMTP host" name="config.host" :error="fieldErrors['config.host']">
      <UInput
        :model-value="modelValue.host"
        placeholder="smtp.gmail.com"
        @update:model-value="(v) => update('host', String(v))"
      />
    </UFormField>
    <UFormField label="Port" name="config.port" :error="fieldErrors['config.port']">
      <UInput
        type="number"
        :model-value="modelValue.port"
        @update:model-value="(v) => update('port', Number(v))"
      />
    </UFormField>
    <UFormField label="Username" name="config.username" :error="fieldErrors['config.username']">
      <UInput
        :model-value="modelValue.username"
        @update:model-value="(v) => update('username', String(v))"
      />
    </UFormField>
    <UFormField label="Password" name="config.password" :error="fieldErrors['config.password']">
      <UInput
        type="password"
        :model-value="modelValue.password"
        autocomplete="new-password"
        @update:model-value="(v) => update('password', String(v))"
      />
    </UFormField>
    <UFormField label="Sender" name="config.sender" :error="fieldErrors['config.sender']">
      <UInput
        type="email"
        :model-value="modelValue.sender"
        placeholder="noreply@example.com"
        @update:model-value="(v) => update('sender', String(v))"
      />
    </UFormField>
    <UFormField label="Recipient" name="config.recipient" :error="fieldErrors['config.recipient']">
      <UInput
        type="email"
        :model-value="modelValue.recipient"
        placeholder="ops@example.com"
        @update:model-value="(v) => update('recipient', String(v))"
      />
    </UFormField>
  </div>
</template>
