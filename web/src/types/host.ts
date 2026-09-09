// Host view models. Adapted to camelCase from the snake_case v1 DTOs
// (@ogoune/api-types HostResponse / HostMetricSampleResponse / DiskUsageDTO).

export interface DiskUsage {
  mount: string
  usedPct: number
}

/** A host as shown in the UI. Derived fields (serviceCount, worstLoad) are added
 *  client-side and are optional on the raw mapping. */
export interface Host {
  id: string
  name: string
  os: string | null
  agentVersion: string | null
  online: boolean
  lastSeenAt: string | null
  lastCpuPct: number | null
  lastMemPct: number | null
  lastDiskPct: number | null
  lastNetIn: number | null
  lastNetOut: number | null
  lastDisks: DiskUsage[]
  createdAt: string
  updatedAt: string
  /** Count of monitors linked to this host — derived client-side. */
  serviceCount?: number
  /**
   * Kernel events this host's agent reported, newest first (spec 090).
   *
   * Undefined means "not loaded": the list endpoint does not fetch events, only
   * the detail endpoint does. An empty array means "loaded, and there are none".
   * The two are genuinely different and the type says so rather than flattening
   * them into one.
   */
  events?: HostEvent[]
}

/**
 * One kernel event. Aggregated per kind per collection interval, so an
 * out-of-memory storm is one entry with `occurrences: 200` rather than two
 * hundred entries.
 */
export interface HostEvent {
  id: string
  /** Open set — render an unrecognised kind rather than dropping it. */
  kind: string
  /** When the kernel reported it, not when it was stored. */
  occurredAt: string
  /** Which reader saw it; useful when investigating why events are missing. */
  source: string
  occurrences: number
  detail: HostEventDetail | null
}

/** Classified fields of a kernel event. Never the raw kernel line. */
export interface HostEventDetail {
  process: string | null
  pid: number | null
  cgroup: string | null
  /** Bounded; distinctTruncated says more were seen than this list holds. */
  distinctProcesses: string[]
  distinctTruncated: boolean
}

/** One point-in-time host metric sample (for detail graphs). */
export interface HostMetricSample {
  sampledAt: string
  cpuPct: number
  memPct: number
  netIn: number
  netOut: number
  disks: DiskUsage[]
}

/** Result of registering or rotating — the raw credential is shown exactly once. */
export interface HostCredentialResult {
  credential: string
  prefix: string
}

export interface RegisterHostResult {
  host: Host
  credential: string
  prefix: string
}

/** The detail-graph range presets. */
export type HostMetricRange = '1h' | '6h' | '24h' | '7d'

/** A thin monitor projection carrying its host link — sourced from the v1
 *  monitors endpoint (the only monitor read that includes host_id). Used for
 *  per-host service counts, the hosted-services list, and the monitor panel. */
export interface MonitorSummary {
  id: string
  name: string
  type: string
  status: string
  lastCheckedAt: string | null
  hostId: string | null
}
