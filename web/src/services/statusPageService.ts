import { http, request } from '@/core/http/client'
import type { StatusPageData, PublicMonitorDetail } from '@/types'

/**
 * Public status page data. Unauthenticated.
 */
export const fetchStatusPageData = async (): Promise<StatusPageData> => {
  return await request<StatusPageData>(http, 'status')
}

/**
 * Public monitor detail. Unauthenticated.
 */
export const fetchStatusPageDataDetail = async (id: string): Promise<PublicMonitorDetail> => {
  return await request<PublicMonitorDetail>(http, `status/${id}`)
}
