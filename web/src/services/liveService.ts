import { getAuthenticatedClient, request } from '@/core/http/client'
import type { LiveSnapshot } from '@/types'

export const fetchLiveSnapshot = async (resourceId: string): Promise<LiveSnapshot> => {
  return await request<LiveSnapshot>(getAuthenticatedClient(), `resources/${resourceId}/live`)
}
