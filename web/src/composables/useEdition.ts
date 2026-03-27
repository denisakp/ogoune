import { computed, ref } from 'vue'

import axiosClient from '@/libs/axios.helper'
import type { CustomAxiosConfig } from '@/libs/axios.helper'

type Edition = 'community' | 'enterprise'

interface EditionResponse {
  edition: Edition
  version: string
}

const edition = ref<Edition>('community')
const version = ref<string>('1.0.0')
const isLoaded = ref(false)
const isLoading = ref(false)

export function useEdition() {
  const isEnterprise = computed(() => edition.value === 'enterprise')

  const load = async () => {
    if (isLoaded.value || isLoading.value) {
      return
    }

    isLoading.value = true
    try {
      const requestConfig: CustomAxiosConfig = {
        skipErrorToast: true,
      }

      const res = await axiosClient.get('/system/edition', requestConfig)
      const data = res.data as EditionResponse

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
