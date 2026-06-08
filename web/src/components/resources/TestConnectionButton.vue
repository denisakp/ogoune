<script setup lang="ts">
import { ref, computed } from 'vue'

import { retryAfterSeconds, testCredential } from '@/services/credentialService'
import type { CredentialCreatePayload, TestConnectionResponse } from '@/types'

const props = defineProps<{
  resourceId?: string
  payload: CredentialCreatePayload | null
}>()

const loading = ref(false)
const result = ref<TestConnectionResponse | null>(null)
const errorMessage = ref<string | null>(null)

const disabledReason = computed<string | null>(() => {
  if (!props.resourceId) {
    return 'Save the monitor first to enable live testing.'
  }
  if (!props.payload || !props.payload.password) {
    return 'Enter a password to enable testing.'
  }
  return null
})

async function runTest() {
  if (!props.resourceId || !props.payload || !props.payload.password) return
  loading.value = true
  result.value = null
  errorMessage.value = null
  try {
    result.value = await testCredential(props.resourceId, props.payload)
  } catch (err) {
    const retry = retryAfterSeconds(err)
    if (retry !== null) {
      errorMessage.value = `Too many test requests. Retry in ${retry} seconds.`
    } else {
      errorMessage.value = 'Test failed: unable to reach the server.'
    }
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div data-testid="test-connection" class="mt-2">
    <component :is="disabledReason ? 'UTooltip' : 'div'" :text="disabledReason ?? undefined">
      <UButton
        :loading="loading"
        :disabled="!!disabledReason"
        data-testid="test-connection-button"
        color="neutral"
        variant="soft"
        @click="runTest"
      >
        Test connection
      </UButton>
    </component>

    <UAlert
      v-if="result"
      :color="result.status === 'ok' ? 'success' : 'error'"
      variant="soft"
      :icon="result.status === 'ok' ? 'i-lucide-check-circle-2' : 'i-lucide-circle-alert'"
      class="mt-2"
      :title="
        result.status === 'ok'
          ? `Connection successful (${result.latency_ms} ms)`
          : `Connection failed: ${result.cause ?? 'unknown reason'}`
      "
      data-testid="test-connection-result"
    />

    <UAlert
      v-if="errorMessage"
      color="warning"
      variant="soft"
      icon="i-lucide-triangle-alert"
      class="mt-2"
      :title="errorMessage"
      data-testid="test-connection-error"
    />
  </div>
</template>
