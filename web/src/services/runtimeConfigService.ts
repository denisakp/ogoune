import { http, request } from '@/core/http/client'

export interface RuntimeConfig {
  ssl_provider: 'letsencrypt' | 'external' | 'disabled'
  edition: 'community' | 'enterprise'
  version: string
}

const SKIP_TOASTS = { headers: { 'x-skip-success-toast': '1' } }

const runtimeConfigService = {
  async get(): Promise<RuntimeConfig> {
    return await request<RuntimeConfig>(http, 'config/runtime', SKIP_TOASTS)
  },
}

export default runtimeConfigService
