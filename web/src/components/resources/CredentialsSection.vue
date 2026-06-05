<script setup lang="ts">
import { computed } from 'vue'

import type { CredentialCreatePayload } from '@/types'

const props = defineProps<{
  protocolType?: string
  modelValue: CredentialCreatePayload | null
  hasExistingCredential?: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: CredentialCreatePayload | null): void
  (e: 'clear'): void
}>()

const supported = computed(() => {
  return (
    props.protocolType === 'redis' ||
    props.protocolType === 'mysql' ||
    props.protocolType === 'postgres'
  )
})

const usernameLabel = computed(() => {
  switch (props.protocolType) {
    case 'redis':
      return 'Username (optional — Redis 6+ ACL)'
    case 'mysql':
    case 'postgres':
      return 'Username'
    default:
      return 'Username'
  }
})

const currentValue = computed<CredentialCreatePayload>(() => {
  return props.modelValue ?? { password: '' }
})

function update(patch: Partial<CredentialCreatePayload>) {
  emit('update:modelValue', { ...currentValue.value, ...patch })
}
</script>

<template>
  <div v-if="supported" data-testid="credentials-section" class="mb-4">
    <div class="flex items-center gap-2 my-3">
      <span class="text-sm font-medium">Authentication (optional)</span>
      <div class="flex-1 border-t border-slate-200 dark:border-slate-700"></div>
    </div>

    <UAlert
      v-if="hasExistingCredential"
      color="info"
      variant="soft"
      icon="i-lucide-info"
      class="mb-3"
      title="Credentials are configured for this monitor."
      description="Fill the fields to replace them, or click 'Remove credentials' to revert to the no-auth path."
    />

    <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
      <div>
        <label class="block text-sm font-medium mb-1">{{ usernameLabel }}</label>
        <UInput
          :model-value="currentValue.username ?? ''"
          placeholder="e.g. monitor"
          :maxlength="128"
          data-testid="credentials-username"
          class="w-full"
          @update:model-value="(v: string | number) => update({ username: String(v) })"
        />
      </div>
      <div>
        <label class="block text-sm font-medium mb-1">
          Password
          <span v-if="!hasExistingCredential" class="text-red-500">*</span>
        </label>
        <UInput
          :model-value="currentValue.password"
          type="password"
          :placeholder="hasExistingCredential ? 'Leave empty to keep current' : 'Required'"
          :maxlength="256"
          data-testid="credentials-password"
          class="w-full"
          @update:model-value="(v: string | number) => update({ password: String(v) })"
        />
      </div>
    </div>

    <UButton
      v-if="hasExistingCredential"
      color="error"
      variant="soft"
      size="xs"
      data-testid="credentials-clear"
      class="mt-2"
      @click="emit('clear')"
    >
      Remove credentials
    </UButton>
  </div>
</template>
