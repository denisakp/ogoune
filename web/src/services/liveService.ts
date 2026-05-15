import axiosHelper from '@/libs/axios.helper'
import type { LiveSnapshot } from '@/types'

export const fetchLiveSnapshot = async (resourceId: string): Promise<LiveSnapshot> => {
  const { data } = await axiosHelper.get<LiveSnapshot>(`/resources/${resourceId}/live`)
  return data
}
