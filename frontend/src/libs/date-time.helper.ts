export const timeAgo = (date: Date | string): string => {
  if (!date) return 'Never'

  const now = new Date()
  const past = new Date(date)

  const diffInSeconds = Math.floor((now.getTime() - past.getTime()) / 1000)

  // less thant 60 seconds
  if (diffInSeconds < 60) return 'Just now'

  // less then 60 minutes
  const diffInMinutes = Math.floor(diffInSeconds / 60)
  if (diffInMinutes < 60) return `${diffInMinutes} minute${diffInMinutes > 1 ? 's' : ''} ago`

  // less than 24 hours
  const diffInHours = Math.floor(diffInMinutes / 60)
  if (diffInHours < 24) return `${diffInHours} hour${diffInHours > 1 ? 's' : ''} ago`

  // 24 hours or more
  const diffInDays = Math.floor(diffInHours / 24)
  if (diffInDays > 30) return past.toLocaleDateString()

  return `${diffInDays} day${diffInDays > 1 ? 's' : ''} ago`
}

// Helper function to get time range cutoff date
export const getTimeRangeCutoff = (range: '24h' | '7d' | '30d' | '365d'): Date => {
  const now = new Date()
  const cutoff = new Date(now)

  switch (range) {
    case '24h':
      cutoff.setHours(now.getHours() - 24)
      break
    case '7d':
      cutoff.setDate(now.getDate() - 7)
      break
    case '30d':
      cutoff.setDate(now.getDate() - 30)
      break
    case '365d':
      cutoff.setFullYear(now.getFullYear() - 1)
      break
  }

  return cutoff
}
