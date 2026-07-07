import { computed, ref } from 'vue'

import { getAuthenticatedClient, request } from '@/core/http/client'

type Edition = 'community' | 'enterprise'

interface EditionResponse {
  edition: Edition
  version: string
}

const edition = ref<Edition>('community')
const version = ref<string>('1.0.0-beta')
const isLoaded = ref(false)
const isLoading = ref(false)

/**
 * Reports the running edition (`community` | `enterprise`) and version.
 * Source of truth for every EE-gating affordance in the UI.
 */
export function useLicence() {
  const isEnterprise = computed(() => edition.value === 'enterprise')

  const load = async () => {
    if (isLoaded.value || isLoading.value) {
      return
    }

    isLoading.value = true
    try {
      const data = await request<EditionResponse>(getAuthenticatedClient(), 'system/edition', {
        headers: { 'x-skip-error-toast': '1' },
      })
      edition.value = data.edition
      version.value = data.version
      isLoaded.value = true
    } finally {
      isLoading.value = false
    }
  }

  return {
    edition,
    version,
    isEnterprise,
    isLoaded,
    load,
  }
}
