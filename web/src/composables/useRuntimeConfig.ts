import runtimeConfigService, { type RuntimeConfig } from '@/services/runtimeConfigService'

let cached: RuntimeConfig | null = null
let inflight: Promise<RuntimeConfig> | null = null

const DEFAULT: RuntimeConfig = {
  ssl_provider: 'external',
  edition: 'community',
  version: 'unknown',
}

export async function loadRuntimeConfig(): Promise<RuntimeConfig> {
  if (cached) return cached
  if (inflight) return inflight
  inflight = runtimeConfigService
    .get()
    .then((cfg) => {
      cached = cfg
      return cfg
    })
    .catch(() => {
      cached = DEFAULT
      return DEFAULT
    })
    .finally(() => {
      inflight = null
    })
  return inflight
}

export function useRuntimeConfig(): RuntimeConfig {
  return cached ?? DEFAULT
}

export function _resetRuntimeConfigForTest() {
  cached = null
  inflight = null
}
