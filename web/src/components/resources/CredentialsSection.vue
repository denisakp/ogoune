<script setup lang="ts">
import { computed } from 'vue'

import type { CredentialCreatePayload } from '@/types'

/**
 * Auth-credentials section for protocol-aware resources.
 *
 * Visible only when `protocolType ∈ {redis, mysql, postgres}`. For other types
 * it renders nothing — there is no concept of credentials at the byte/TCP layer.
 *
 * The plaintext password is held in `modelValue.password` only during the
 * lifetime of the form; it is sent once over HTTPS via `setCredential()` and
 * never returned by any subsequent read.
 */
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
  <div v-if="supported" data-testid="credentials-section" class="credentials-section">
    <a-divider orientation="left">Authentication (optional)</a-divider>

    <a-alert
      v-if="hasExistingCredential"
      type="info"
      show-icon
      style="margin-bottom: 12px"
      message="Credentials are configured for this monitor."
      description="Fill the fields to replace them, or click 'Remove credentials' to revert to the no-auth path."
    />

    <a-row :gutter="16">
      <a-col :xs="24" :sm="12">
        <a-form-item :label="usernameLabel">
          <a-input
            :value="currentValue.username ?? ''"
            placeholder="e.g. monitor"
            :maxlength="128"
            data-testid="credentials-username"
            @update:value="(v: string) => update({ username: v })"
          />
        </a-form-item>
      </a-col>
      <a-col :xs="24" :sm="12">
        <a-form-item label="Password" :required="!hasExistingCredential">
          <a-input-password
            :value="currentValue.password"
            :placeholder="hasExistingCredential ? 'Leave empty to keep current' : 'Required'"
            :maxlength="256"
            data-testid="credentials-password"
            @update:value="(v: string) => update({ password: v })"
          />
        </a-form-item>
      </a-col>
    </a-row>

    <a-button
      v-if="hasExistingCredential"
      danger
      size="small"
      data-testid="credentials-clear"
      @click="emit('clear')"
    >
      Remove credentials
    </a-button>
  </div>
</template>

<style scoped>
.credentials-section {
  margin-bottom: 16px;
}
</style>
