import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { CreateAPIKeyResponse } from '@/services/accountService'

/**
 * Transient store for API key creation.
 * `lastCreated` holds the full secret returned by `POST /api-keys`.
 * Spec 059 US4: shown ONCE in the banner; cleared on dismiss / reload.
 * Intentionally NOT persisted to localStorage.
 */
export const useApiKeyStore = defineStore('apiKeys', () => {
  const lastCreated = ref<CreateAPIKeyResponse | null>(null)

  function set(payload: CreateAPIKeyResponse) {
    lastCreated.value = payload
  }

  function clear() {
    lastCreated.value = null
  }

  return { lastCreated, set, clear }
})
