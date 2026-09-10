package agentwire

import (
	"encoding/json"
	"errors"
	"math"
	"testing"
	"time"
)

func TestEncodeDecode_RoundTrip(t *testing.T) {
	in := Frame{
		OS:           "Ubuntu 24.04",
		AgentVersion: "0.1.0",
		CPUPct:       12.4,
		MemPct:       47.1,
		NetIn:        10432,
		NetOut:       88123,
		Disks:        []DiskUsage{{Mount: "/", UsedPct: 23.0}, {Mount: "/data", UsedPct: 71.2}},
	}
	b, err := Encode(in)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	out, err := Decode(b)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if out.SchemaVersion != SchemaVersion {
		t.Fatalf("SchemaVersion = %d, want %d (Encode should stamp it)", out.SchemaVersion, SchemaVersion)
	}
	if out.CPUPct != in.CPUPct || out.MemPct != in.MemPct || out.NetIn != in.NetIn || out.NetOut != in.NetOut {
		t.Fatalf("scalar mismatch: %+v vs %+v", out, in)
	}
	if len(out.Disks) != 2 || out.Disks[1].Mount != "/data" || out.Disks[1].UsedPct != 71.2 {
		t.Fatalf("disks mismatch: %+v", out.Disks)
	}
	if out.OS != in.OS || out.AgentVersion != in.AgentVersion {
		t.Fatalf("os/agent_version mismatch: %+v", out)
	}
}

func TestDecode_AbsentVersionIsV1(t *testing.T) {
	// A legacy frame (the exact shape spec 079 accepts) with NO schema_version.
	raw := []byte(`{"os":"Debian 12","cpu_pct":5,"mem_pct":10,"net_in":1,"net_out":2,"disks":[{"mount":"/","used_pct":3}]}`)
	f, err := Decode(raw)
	if err != nil {
		t.Fatalf("Decode legacy: %v", err)
	}
	if f.SchemaVersion != 1 {
		t.Fatalf("absent schema_version → %d, want 1", f.SchemaVersion)
	}
	if f.CPUPct != 5 || f.Disks[0].UsedPct != 3 {
		t.Fatalf("legacy fields not decoded: %+v", f)
	}
}

func TestDecode_UnknownVersionRejected(t *testing.T) {
	raw := []byte(`{"schema_version":999,"cpu_pct":1,"mem_pct":1,"net_in":0,"net_out":0,"disks":[]}`)
	_, err := Decode(raw)
	if !errors.Is(err, ErrUnsupportedVersion) {
		t.Fatalf("expected ErrUnsupportedVersion, got %v", err)
	}
}

func TestDecode_Malformed(t *testing.T) {
	if _, err := Decode([]byte("not json")); err == nil {
		t.Fatal("expected error decoding malformed JSON")
	}
}

func TestDecode_MissingRequiredField(t *testing.T) {
	// cpu_pct present, mem_pct absent → malformed.
	raw := []byte(`{"cpu_pct":5,"net_in":1,"net_out":2,"disks":[]}`)
	_, err := Decode(raw)
	if !errors.Is(err, ErrMissingField) {
		t.Fatalf("expected ErrMissingField, got %v", err)
	}
	// A real zero value is NOT missing.
	ok := []byte(`{"cpu_pct":0,"mem_pct":0,"net_in":0,"net_out":0,"disks":[]}`)
	if _, err := Decode(ok); err != nil {
		t.Fatalf("zero-valued frame should decode, got %v", err)
	}
}

func TestValidate(t *testing.T) {
	if err := (Frame{CPUPct: 12, MemPct: 30}).Validate(); err != nil {
		t.Fatalf("valid frame rejected: %v", err)
	}
	if err := (Frame{CPUPct: math.NaN()}).Validate(); err == nil {
		t.Fatal("NaN cpu_pct should be rejected")
	}
	if err := (Frame{CPUPct: math.Inf(1)}).Validate(); err == nil {
		t.Fatal("Inf cpu_pct should be rejected")
	}
	if err := (Frame{Disks: []DiskUsage{{Mount: "/", UsedPct: math.NaN()}}}).Validate(); err == nil {
		t.Fatal("NaN disk used_pct should be rejected")
	}
}

func TestEncode_OmitsEmptyOptionalStrings(t *testing.T) {
	b, err := Encode(Frame{CPUPct: 1, MemPct: 2})
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, ok := m["os"]; ok {
		t.Fatal("empty os should be omitted")
	}
	if m["schema_version"] != float64(SchemaVersion) {
		t.Fatalf("schema_version = %v, want %d", m["schema_version"], SchemaVersion)
	}
}

// --- Kernel events on the frame (spec 090) ---

// A frame from an agent predating kernel events: version 1, no events field. It
// must decode exactly as it always did. This compatibility is free — it is what
// "every frame is a metrics frame" already means, not a branch added on top.
func TestDecode_V1FrameWithoutEventsStillDecodes(t *testing.T) {
	raw := `{"schema_version":1,"cpu_pct":12.5,"mem_pct":40,"net_in":100,"net_out":200,"disks":[]}`

	f, err := Decode([]byte(raw))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if f.SchemaVersion != 1 {
		t.Errorf("schema_version = %d, want 1", f.SchemaVersion)
	}
	if len(f.Events) != 0 {
		t.Errorf("events = %d, want 0: no events field means no events, not a malformed frame", len(f.Events))
	}
}

func TestDecode_FrameWithoutVersionOrEvents(t *testing.T) {
	f, err := Decode([]byte(`{"cpu_pct":1,"mem_pct":2,"net_in":3,"net_out":4}`))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if f.SchemaVersion != 1 {
		t.Errorf("absent version = %d, want normalised to 1", f.SchemaVersion)
	}
	if len(f.Events) != 0 {
		t.Errorf("events = %d, want 0", len(f.Events))
	}
}

func TestDecode_V2FrameWithEvents(t *testing.T) {
	raw := `{"schema_version":2,"cpu_pct":97.4,"mem_pct":99.1,"net_in":1,"net_out":2,
	  "events":[{"kind":"oom_kill","occurred_at":"2026-09-09T14:02:47Z","source":"kmsg",
	             "occurrences":37,"process":"postgres","pid":4711,
	             "distinct_processes":["postgres","node"],"distinct_truncated":false}]}`

	f, err := Decode([]byte(raw))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if len(f.Events) != 1 {
		t.Fatalf("events = %d, want 1", len(f.Events))
	}
	e := f.Events[0]
	if e.Kind != KindOOMKill || e.Source != SourceKmsg {
		t.Errorf("kind/source = %q/%q", e.Kind, e.Source)
	}
	if e.Occurrences != 37 {
		t.Errorf("occurrences = %d, want 37", e.Occurrences)
	}
	if e.Process != "postgres" || e.PID != 4711 {
		t.Errorf("process/pid = %q/%d", e.Process, e.PID)
	}
	if len(e.DistinctProcesses) != 2 {
		t.Errorf("distinct = %v", e.DistinctProcesses)
	}
}

// A newer agent against an older backend still fails loudly. This gate is what
// makes adding a second frame type later free, so it must keep working exactly.
func TestDecode_NewerSchemaStillRefused(t *testing.T) {
	_, err := Decode([]byte(`{"schema_version":99,"cpu_pct":1,"mem_pct":2,"net_in":3,"net_out":4}`))
	if !errors.Is(err, ErrUnsupportedVersion) {
		t.Fatalf("err = %v, want ErrUnsupportedVersion: refused explicitly, never accepted in part", err)
	}
}

// Events do not exempt a frame from the required-field rule. Every frame is a
// metrics frame; one carrying events but missing cpu_pct is malformed, not a new
// shape that deserves different treatment.
func TestDecode_EventsDoNotBypassRequiredFields(t *testing.T) {
	raw := `{"schema_version":2,"mem_pct":2,"net_in":3,"net_out":4,
	  "events":[{"kind":"segfault","occurred_at":"2026-09-09T14:02:47Z","source":"kmsg","occurrences":1}]}`

	_, err := Decode([]byte(raw))
	if !errors.Is(err, ErrMissingField) {
		t.Fatalf("err = %v, want ErrMissingField", err)
	}
}

// An unknown kind is carried through rather than dropped: the set is open, and an
// older backend paired with a newer agent must lose nothing.
func TestDecode_UnknownEventKindIsPreserved(t *testing.T) {
	raw := `{"schema_version":2,"cpu_pct":1,"mem_pct":2,"net_in":3,"net_out":4,
	  "events":[{"kind":"something_new","occurred_at":"2026-09-09T14:02:47Z","source":"kmsg","occurrences":2}]}`

	f, err := Decode([]byte(raw))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if len(f.Events) != 1 || f.Events[0].Kind != "something_new" {
		t.Fatalf("unknown kind dropped: %+v", f.Events)
	}
}

func TestEncodeDecode_EventsRoundTrip(t *testing.T) {
	at := time.Date(2026, 9, 9, 14, 2, 47, 0, time.UTC)
	in := Frame{
		CPUPct: 1, MemPct: 2, NetIn: 3, NetOut: 4,
		Events: []KernelEvent{{
			Kind: KindOOMKill, OccurredAt: at, Source: SourceCgroup, Occurrences: 5,
			Process: "node", DistinctProcesses: []string{"node"}, DistinctTruncated: true,
		}},
	}

	b, err := Encode(in)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	out, err := Decode(b)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if out.SchemaVersion != SchemaVersion {
		t.Errorf("schema_version = %d, want %d", out.SchemaVersion, SchemaVersion)
	}
	if len(out.Events) != 1 {
		t.Fatalf("events = %d, want 1", len(out.Events))
	}
	if !out.Events[0].OccurredAt.Equal(at) || !out.Events[0].DistinctTruncated {
		t.Errorf("round trip lost fields: %+v", out.Events[0])
	}
}
