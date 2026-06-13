export interface MonthlyReport {
  enabled: boolean
  recipientEmail: string
  schedule: 'monthly-1st'
  scope: 'all-resources'
  lastSentAt: string | null
}

export type ReportStatus = 'delivered' | 'pending' | 'failed'

export interface ReportResourceBreakdown {
  name: string
  uptimePct: number
  incidents: number
}

export interface ReportHistoryEntry {
  id: string
  period: string
  sentAt: string
  status: ReportStatus
  uptimePct: number
  incidentCount: number
  downtimeSeconds: number
  recipientEmail: string
  resourceBreakdown: ReportResourceBreakdown[]
}
