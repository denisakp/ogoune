<script setup lang="ts">
import { ref } from 'vue'

import { retryAfterSeconds, testCredential } from '@/services/credentialService'
import type { CredentialCreatePayload, TestConnectionResponse } from '@/types'

/**
 * "Test connection" button used inside the credentials section of the
 * monitor edit form. Calls the rate-limited live-test endpoint
 * (10 req/min/user) and surfaces the outcome inline.
 *
 * The endpoint does NOT persist the supplied credential; this lets operators
 * verify they typed the right thing before clicking Save.
 */
const props = defineProps<{
  resourceId?: string
  payload: CredentialCreatePayload | null
}>()

const loading = ref(false)
const result = ref<TestConnectionResponse | null>(null)
const errorMessage = ref<string | null>(null)

const disabledReason = (): string | null => {
  if (!props.resourceId) {
    return 'Save the monitor first to enable live testing.'
  }
  if (!props.payload || !props.payload.password) {
    return 'Enter a password to enable testing.'
  }
  return null
}

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
  <div data-testid="test-connection" class="test-connection">
    <a-button
      :loading="loading"
      :disabled="!!disabledReason()"
      :title="disabledReason() ?? ''"
      data-testid="test-connection-button"
      @click="runTest"
    >
      Test connection
    </a-button>

    <a-alert
      v-if="result"
      :type="result.status === 'ok' ? 'success' : 'error'"
      show-icon
      style="margin-top: 8px"
      :message="
        result.status === 'ok'
          ? `Connection successful (${result.latency_ms} ms)`
          : `Connection failed: ${result.cause ?? 'unknown reason'}`
      "
      data-testid="test-connection-result"
    />

    <a-alert
      v-if="errorMessage"
      type="warning"
      show-icon
      style="margin-top: 8px"
      :message="errorMessage"
      data-testid="test-connection-error"
    />
  </div>
</template>

<style scoped>
.test-connection {
  margin-top: 8px;
}
</style>
