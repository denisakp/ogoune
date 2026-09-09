package v1

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/coder/websocket"

	"github.com/denisakp/ogoune/internal/api/middleware"
	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/port"
	"github.com/denisakp/ogoune/internal/service"
	"github.com/denisakp/ogoune/pkg/agentwire"
)

// AgentStreamHandler upgrades the agent ingestion route to a WebSocket and
// reads periodic metric frames, delegating validation + persistence to
// HostMetricsService.Ingest. The socket is kept off the unit-test path by
// keeping all logic in the service; this handler is a thin read-loop. Frames
// are decoded through the shared pkg/agentwire contract so the agent and the
// backend cannot drift.
// maxEventsPerFrame bounds how many kernel events one frame may contribute
// (spec 090, FR-018). The agent already bounds what it sends; this bounds what
// the server accepts, which is a different guarantee: the agent's cap protects
// the agent, and a server that trusts its clients to bound their own writes has
// no bound at all. An agent can be old, modified, or simply wrong.
const maxEventsPerFrame = 16

type AgentStreamHandler struct {
	metrics *service.HostMetricsService
	// events is optional: a nil repository means kernel events are discarded,
	// which keeps every existing construction of this handler working.
	events port.HostEventRepository
}

func NewAgentStreamHandler(metrics *service.HostMetricsService) *AgentStreamHandler {
	return &AgentStreamHandler{metrics: metrics}
}

// WithEvents attaches kernel event storage.
func (h *AgentStreamHandler) WithEvents(repo port.HostEventRepository) *AgentStreamHandler {
	h.events = repo
	return h
}

// ingestEvents stores the kernel events a frame carried.
//
// Best-effort throughout: a storage failure is logged and swallowed. Events are
// diagnostic context, and losing them must never drop an agent's connection or
// interrupt the metrics that share the frame.
func (h *AgentStreamHandler) ingestEvents(ctx context.Context, hostID string, frame agentwire.Frame) {
	if h.events == nil || len(frame.Events) == 0 {
		return
	}

	accepted := frame.Events
	if len(accepted) > maxEventsPerFrame {
		slog.Warn("agent stream: dropping events beyond the per-frame bound",
			"host_id", hostID, "received", len(accepted), "accepted", maxEventsPerFrame)
		accepted = accepted[:maxEventsPerFrame]
	}

	for _, e := range accepted {
		occurrences := e.Occurrences
		if occurrences < 1 {
			occurrences = 1
		}
		occurredAt := e.OccurredAt
		if occurredAt.IsZero() {
			occurredAt = time.Now().UTC()
		}

		stored := &domain.HostEvent{
			HostID:      hostID,
			OccurredAt:  occurredAt,
			Kind:        e.Kind,
			Source:      e.Source,
			Occurrences: occurrences,
		}
		// Detail is nil when the report named nothing -- a counter-based source
		// says how many, not which. Absent detail must stay absent rather than
		// becoming an empty object.
		if e.Process != "" || e.Cgroup != "" || len(e.DistinctProcesses) > 0 {
			stored.Detail = &domain.HostEventDetail{
				Process:           e.Process,
				PID:               e.PID,
				Cgroup:            e.Cgroup,
				DistinctProcesses: e.DistinctProcesses,
				DistinctTruncated: e.DistinctTruncated,
			}
		}

		if err := h.events.Create(ctx, stored); err != nil {
			slog.Debug("agent stream: store kernel event failed", "host_id", hostID, "error", err)
		}
	}
}

// Stream handles GET /api/v1/agent/stream (WebSocket upgrade, host-credential auth).
func (h *AgentStreamHandler) Stream(w http.ResponseWriter, r *http.Request) {
	hostID, ok := middleware.HostIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		// Auth is by bearer host credential, not Origin; agents are non-browser.
		InsecureSkipVerify: true,
	})
	if err != nil {
		return
	}
	defer conn.CloseNow()

	ctx := r.Context()
	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			// Client disconnected or context cancelled — end the read loop.
			break
		}
		frame, err := agentwire.Decode(data)
		if err != nil {
			// Malformed / missing-field / unsupported-version frame — do not
			// store, do not advance last_seen.
			slog.Debug("agent stream: rejected frame", "host_id", hostID, "error", err)
			continue
		}
		if err := h.metrics.Ingest(ctx, hostID, frameToSample(frame)); err != nil {
			slog.Warn("agent stream: ingest failed", "host_id", hostID, "error", err)
			// Keep the connection open; a transient bad frame must not drop the agent.
		}
		h.ingestEvents(ctx, hostID, frame)
	}
	conn.Close(websocket.StatusNormalClosure, "")
}

// frameToSample maps the shared wire frame to the ingestion sample. Optional
// string fields become nil pointers when empty so the host snapshot's
// COALESCE keeps prior values.
func frameToSample(f agentwire.Frame) service.IngestSample {
	s := service.IngestSample{
		CPUPct: f.CPUPct,
		MemPct: f.MemPct,
		NetIn:  f.NetIn,
		NetOut: f.NetOut,
	}
	if f.OS != "" {
		os := f.OS
		s.OS = &os
	}
	if f.AgentVersion != "" {
		av := f.AgentVersion
		s.AgentVersion = &av
	}
	for _, d := range f.Disks {
		s.Disks = append(s.Disks, domain.DiskUsage{Mount: d.Mount, UsedPct: d.UsedPct})
	}
	return s
}
