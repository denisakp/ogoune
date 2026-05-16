export const formatDate = (dateString: string): string => {
  const date = new Date(dateString)
  return date.toLocaleString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  })
}

export const formatDuration = (startDate: string, endDate?: string | null): string => {
  const start = new Date(startDate).getTime()
  const end = endDate ? new Date(endDate).getTime() : Date.now()
  const diff = end - start
  const hours = Math.floor(diff / (1000 * 60 * 60))
  const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60))
  if (hours > 0) return `${hours}h ${minutes}m`
  return `${minutes}m`
}

export const formatExpirationDate = (dateString: string): string => {
  const date = new Date(dateString)
  return date.toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' })
}

export const getDaysUntilExpiration = (dateString: string): number => {
  const diff = new Date(dateString).getTime() - Date.now()
  return Math.ceil(diff / (1000 * 60 * 60 * 24))
}

export interface ExpirationStatusResult {
  text: string
  color: string
  type: 'success' | 'warning' | 'danger'
}

export const getExpirationStatus = (dateString?: string): ExpirationStatusResult => {
  if (!dateString) return { text: 'Unknown', color: '#d9d9d9', type: 'success' }
  const days = getDaysUntilExpiration(dateString)
  if (days < 0) return { text: 'Expired', color: '#ff4d4f', type: 'danger' }
  if (days <= 7) return { text: `Expires in ${days} day${days !== 1 ? 's' : ''}`, color: '#ff4d4f', type: 'danger' }
  if (days <= 30) return { text: `Expires in ${days} days`, color: '#faad14', type: 'warning' }
  return { text: `Expires in ${days} days`, color: '#52c41a', type: 'success' }
}
