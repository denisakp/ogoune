import type { MonthlyReport, ReportHistoryEntry } from '@/types'
import { defaultMonthlyReport, REPORT_HISTORY_FIXTURE } from '@/mocks/reports.fixture'

/**
 * Spec 070 — internal contract between the reports UI and its (mocked) feed.
 * Documented at `specs/070-reports-dashboards/contracts/reports-feed.contract.md`.
 *
 * UI components MUST import only the `ReportsFeed` interface and the
 * default export below. Swapping to a real backend in a future PRD is a
 * one-file change (FR-030a).
 *
 * Note: the zero-resources guard (FR-007) lives in `useReports` at the
 * composable layer so the mock stays pure and pinia-free at module
 * evaluation time. A real backend will enforce the same guard server-side.
 */
export interface ReportsFeed {
  fetchMonthly(): Promise<MonthlyReport>
  saveMonthly(next: MonthlyReport): Promise<MonthlyReport>
  fetchHistory(limit?: number): Promise<ReportHistoryEntry[]>
  fetchPendingPreview(): Promise<ReportHistoryEntry | null>
}

export function createMockReportsFeed(): ReportsFeed {
  let monthly: MonthlyReport = defaultMonthlyReport()
  const history: ReportHistoryEntry[] = [...REPORT_HISTORY_FIXTURE]

  return {
    async fetchMonthly() {
      return { ...monthly }
    },
    async saveMonthly(next: MonthlyReport) {
      monthly = { ...next }
      return { ...monthly }
    },
    async fetchHistory(limit = 6) {
      return history
        .slice()
        .sort((a, b) => new Date(b.sentAt).getTime() - new Date(a.sentAt).getTime())
        .slice(0, limit)
    },
    async fetchPendingPreview() {
      return null
    },
  }
}

export function createRemoteStub(): ReportsFeed {
  const notImplemented = (op: string) => async () => {
    throw new Error(`reportsService: 'remote' mode not implemented yet (${op})`)
  }
  return {
    fetchMonthly: notImplemented('fetchMonthly'),
    saveMonthly: notImplemented('saveMonthly') as ReportsFeed['saveMonthly'],
    fetchHistory: notImplemented('fetchHistory') as ReportsFeed['fetchHistory'],
    fetchPendingPreview: notImplemented('fetchPendingPreview') as ReportsFeed['fetchPendingPreview'],
  }
}

const mode = (import.meta.env.VITE_REPORTS_FEED_MODE as string | undefined) ?? 'mock'

let activeFeed: ReportsFeed = mode === 'remote' ? createRemoteStub() : createMockReportsFeed()

const reportsService: ReportsFeed = {
  fetchMonthly: () => activeFeed.fetchMonthly(),
  saveMonthly: (next) => activeFeed.saveMonthly(next),
  fetchHistory: (limit) => activeFeed.fetchHistory(limit),
  fetchPendingPreview: () => activeFeed.fetchPendingPreview(),
}

export default reportsService

// Test-only helpers.
export function __setReportsFeedForTests(feed: ReportsFeed): void {
  activeFeed = feed
}

export function __resetReportsForTests(): void {
  activeFeed = mode === 'remote' ? createRemoteStub() : createMockReportsFeed()
}

export function __createMockFeedForTests(): ReportsFeed {
  return createMockReportsFeed()
}

export function __createRemoteStubForTests(): ReportsFeed {
  return createRemoteStub()
}
