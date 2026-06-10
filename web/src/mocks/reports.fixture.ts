import type { MonthlyReport, ReportHistoryEntry } from '@/types'

/**
 * Mocked reports feed — spec 070 / US1.
 * Replaced by a real backend in a future PRD; the UI imports through
 * `reportsService` so swapping the source is a one-file change (FR-030a).
 */

export function defaultMonthlyReport(): MonthlyReport {
  return {
    enabled: false,
    recipientEmail: 'admin@example.com',
    schedule: 'monthly-1st',
    scope: 'all-resources',
    lastSentAt: null,
  }
}

const SAMPLE_BREAKDOWN = [
  { name: 'api.example.com', uptimePct: 99.97, incidents: 1 },
  { name: 'web.example.com', uptimePct: 100, incidents: 0 },
  { name: 'db-primary', uptimePct: 99.82, incidents: 2 },
  { name: 'cdn-edge', uptimePct: 99.99, incidents: 0 },
]

export const REPORT_HISTORY_FIXTURE: ReportHistoryEntry[] = [
  {
    id: 'rpt-2026-05',
    period: 'May 2026',
    sentAt: '2026-06-01T08:00:00.000Z',
    status: 'delivered',
    uptimePct: 99.94,
    incidentCount: 3,
    downtimeSeconds: 1620,
    recipientEmail: 'admin@example.com',
    resourceBreakdown: SAMPLE_BREAKDOWN,
  },
  {
    id: 'rpt-2026-04',
    period: 'April 2026',
    sentAt: '2026-05-01T08:00:00.000Z',
    status: 'delivered',
    uptimePct: 99.89,
    incidentCount: 5,
    downtimeSeconds: 2849,
    recipientEmail: 'admin@example.com',
    resourceBreakdown: SAMPLE_BREAKDOWN,
  },
  {
    id: 'rpt-2026-03',
    period: 'March 2026',
    sentAt: '2026-04-01T08:00:00.000Z',
    status: 'delivered',
    uptimePct: 99.71,
    incidentCount: 8,
    downtimeSeconds: 7440,
    recipientEmail: 'admin@example.com',
    resourceBreakdown: SAMPLE_BREAKDOWN,
  },
]
