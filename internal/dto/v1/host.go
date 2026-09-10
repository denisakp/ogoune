package v1

// DiskUsageDTO is the per-mount disk utilisation entry shared by host snapshots
// and metric samples.
// @name DiskUsageDTO
type DiskUsageDTO struct {
	Mount   string  `json:"mount"`
	UsedPct float64 `json:"used_pct"`
}

// HostResponse is the v1 API representation of a monitored host, including its
// denormalized latest snapshot and derived online state.
// @name HostResponse
// HostEventResponse is one kernel event reported by a host's agent (spec 090).
// One per kind per collection interval, carrying how many kernel reports it
// aggregates -- an out-of-memory storm is one entry with occurrences=200, not two
// hundred entries.
type HostEventResponse struct {
	ID string `json:"id"`
	// Kind is an open set. Treat a value you do not recognise as displayable
	// rather than as an error: the set will grow.
	Kind string `json:"kind"`
	// OccurredAt is when the KERNEL reported it, not when it was stored.
	OccurredAt string `json:"occurred_at"`
	// Source is which reader saw it, so an operator investigating missing events
	// knows what was working.
	Source      string                   `json:"source"`
	Occurrences int                      `json:"occurrences"`
	Detail      *HostEventDetailResponse `json:"detail"`
}

// HostEventDetailResponse holds the classified fields of an event. Never the raw
// kernel line.
type HostEventDetailResponse struct {
	Process string `json:"process,omitempty"`
	PID     int    `json:"pid,omitempty"`
	Cgroup  string `json:"cgroup,omitempty"`
	// DistinctProcesses is bounded; DistinctTruncated says more were seen than it
	// holds, so a partial list is never read as complete.
	DistinctProcesses []string `json:"distinct_processes,omitempty"`
	DistinctTruncated bool     `json:"distinct_truncated,omitempty"`
}

type HostResponse struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	OS           *string        `json:"os"`
	AgentVersion *string        `json:"agent_version"`
	LastSeenAt   *string        `json:"last_seen_at"`
	Online       bool           `json:"online"`
	LastCPUPct   *float64       `json:"last_cpu_pct"`
	LastMemPct   *float64       `json:"last_mem_pct"`
	LastDiskPct  *float64       `json:"last_disk_pct"`
	LastNetIn    *int64         `json:"last_net_in"`
	LastNetOut   *int64         `json:"last_net_out"`
	LastDisks    []DiskUsageDTO `json:"last_disks"`
	CreatedAt    string         `json:"created_at"`
	UpdatedAt    string         `json:"updated_at"`
	// Events are the kernel events this host's agent reported, newest first
	// (spec 090). An empty array when there are none -- never null: a list that is
	// sometimes absent and sometimes empty is two shapes for one meaning, and every
	// consumer would have to handle both.
	Events []HostEventResponse `json:"events"`
}

// HostMetricSampleResponse is a single point-in-time host metric sample.
// @name HostMetricSampleResponse
type HostMetricSampleResponse struct {
	SampledAt string         `json:"sampled_at"`
	CPUPct    float64        `json:"cpu_pct"`
	MemPct    float64        `json:"mem_pct"`
	NetIn     int64          `json:"net_in"`
	NetOut    int64          `json:"net_out"`
	Disks     []DiskUsageDTO `json:"disks"`
}

// RegisterHostRequest is the request body for POST /api/v1/hosts.
// @name RegisterHostRequest
type RegisterHostRequest struct {
	Name string `json:"name"`
}

// CreateHostCredentialResponse carries a freshly-issued raw credential. The raw
// token is shown exactly once (registration or rotation).
// @name CreateHostCredentialResponse
type CreateHostCredentialResponse struct {
	Credential string `json:"credential"`
	Prefix     string `json:"prefix"`
}

// RegisterHostResponse is returned by POST /api/v1/hosts — the created host plus
// its one-time raw credential.
// @name RegisterHostResponse
type RegisterHostResponse struct {
	Host       HostResponse `json:"host"`
	Credential string       `json:"credential"`
	Prefix     string       `json:"prefix"`
}

// LinkHostRequest is the request body for POST /api/v1/monitors/{id}/host.
// @name LinkHostRequest
type LinkHostRequest struct {
	HostID string `json:"host_id"`
}
