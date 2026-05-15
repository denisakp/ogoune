import { timeAgo, getTimeRangeCutoff } from '@/libs/date-time.helper'

export function useDateTime() {
  return {
    timeAgo,
    getTimeRangeCutoff,
  }
}
