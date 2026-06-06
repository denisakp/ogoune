<script setup lang="ts">
/* eslint-disable @typescript-eslint/ban-ts-comment */
// @ts-nocheck — spec 059 polish debt: NuxtUI v4 template-emit types
interface Header {
  name: string
  value: string
}
interface WebhookConfig {
  url: string
  method: 'POST' | 'PUT'
  headers: Header[]
}

interface Props {
  modelValue: WebhookConfig
  fieldErrors?: Record<string, string>
}
const props = withDefaults(defineProps<Props>(), { fieldErrors: () => ({}) })
const emit = defineEmits<{ (e: 'update:modelValue', v: WebhookConfig): void }>()

function update<K extends keyof WebhookConfig>(key: K, value: WebhookConfig[K]) {
  emit('update:modelValue', { ...props.modelValue, [key]: value })
}

function addHeader() {
  update('headers', [...props.modelValue.headers, { name: '', value: '' }])
}

function removeHeader(i: number) {
  update(
    'headers',
    props.modelValue.headers.filter((_, idx) => idx !== i),
  )
}

function updateHeader(i: number, k: keyof Header, v: string) {
  update(
    'headers',
    props.modelValue.headers.map((h, idx) => (idx === i ? { ...h, [k]: v } : h)),
  )
}
</script>

<template>
  <div class="space-y-3">
    <UFormField label="URL" name="config.url" :error="fieldErrors['config.url']">
      <UInput
class="w-full"         :model-value="modelValue.url"
        placeholder="https://events.example.com/hook"
        @update:model-value="(v) => update('url', String(v))"
      />
    </UFormField>
    <UFormField label="Method" name="config.method" :error="fieldErrors['config.method']">
      <USelect
class="w-full"         :model-value="modelValue.method"
        :items="['POST', 'PUT']"
        @update:model-value="(v) => update('method', v as 'POST' | 'PUT')"
      />
    </UFormField>

    <div>
      <div class="flex items-center justify-between mb-2">
        <label class="text-sm font-medium text-default">Headers</label>
        <UButton size="xs" variant="ghost" icon="i-lucide-plus" @click="addHeader">Add</UButton>
      </div>
      <div v-if="modelValue.headers.length === 0" class="text-xs text-muted">
        No custom headers.
      </div>
      <ul v-else class="space-y-2">
        <li v-for="(h, i) in modelValue.headers" :key="i" class="flex items-center gap-2">
          <UInput
            class="flex-1"
            placeholder="Name"
            :model-value="h.name"
            @update:model-value="(v) => updateHeader(i, 'name', String(v))"
          />
          <UInput
            class="flex-1"
            placeholder="Value"
            :model-value="h.value"
            @update:model-value="(v) => updateHeader(i, 'value', String(v))"
          />
          <UButton
            size="xs"
            color="error"
            variant="ghost"
            icon="i-lucide-trash-2"
            @click="removeHeader(i)"
          />
        </li>
      </ul>
    </div>
  </div>
</template>
