import type { KyInstance } from 'ky'
import { getAuthenticatedClient } from './client'

/**
 * Returns the singleton authenticated Ky instance.
 *
 * Setup-friendly sugar for components/composables. Services (plain TS
 * modules) can import `getAuthenticatedClient` directly from `./client`.
 *
 * @example
 * const http = useHttpClient()
 * const data = await request<Resource[]>(http, 'v1/resources')
 */
export function useHttpClient(): KyInstance {
  return getAuthenticatedClient()
}
