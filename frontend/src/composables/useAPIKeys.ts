import { ref } from 'vue'

import accountService, {
  type APIKey,
  type APIKeyScope,
  type CreateAPIKeyRequest,
} from '@/services/accountService'

export function useAPIKeys() {
  const keys = ref<APIKey[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const revealedKey = ref<string | null>(null)
  const revealKeyPrefix = ref<string | null>(null)

  const loadKeys = async () => {
    loading.value = true
    error.value = null
    try {
      keys.value = await accountService.listAPIKeys()
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load API keys'
      throw err
    } finally {
      loading.value = false
    }
  }

  const createKey = async (
    name: string,
    scope: APIKeyScope,
    expiresAt?: string,
  ): Promise<void> => {
    const payload: CreateAPIKeyRequest = {
      name,
      scope,
    }
    if (expiresAt) {
      payload.expires_at = expiresAt
    }

    loading.value = true
    error.value = null
    try {
      const created = await accountService.createAPIKey(payload)
      revealedKey.value = created.key
      revealKeyPrefix.value = created.key_prefix
      await loadKeys()
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to create API key'
      throw err
    } finally {
      loading.value = false
    }
  }

  const revokeKey = async (id: string) => {
    loading.value = true
    error.value = null
    try {
      await accountService.revokeAPIKey(id)
      await loadKeys()
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to revoke API key'
      throw err
    } finally {
      loading.value = false
    }
  }

  const clearRevealedKey = () => {
    revealedKey.value = null
    revealKeyPrefix.value = null
  }

  return {
    keys,
    loading,
    error,
    revealedKey,
    revealKeyPrefix,
    loadKeys,
    createKey,
    revokeKey,
    clearRevealedKey,
  }
}
