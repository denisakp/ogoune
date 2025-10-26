const timeAgo = (date: Date | string): string => {
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

export default timeAgo