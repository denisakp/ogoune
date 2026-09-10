package v1_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	v1 "github.com/denisakp/ogoune/internal/api/handler/v1"
	"github.com/denisakp/ogoune/internal/correlation"
	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository/internaltest"
	"github.com/denisakp/ogoune/internal/repository/store"
	"github.com/denisakp/ogoune/internal/service"
)

// Incident detail p95, with and without a host attached (spec 089, SC-007/SC-007a;
// spec 091 SC-008).
//
// Unlike the list benchmarks, these wire the *real* IncidentService rather than a
// stub, because the whole point is to measure the enrichment that lives inside
// it -- the host context and, since spec 091, the causal narrative. The two
// benchmarks differ only in whether the monitor carries a host_id, so their delta
// is the enrichment's total cost and nothing else.
//
// The no-host case is the one that matters most: it is the majority of monitors,
// and both enrichments are supposed to return before issuing a single query for
// it. Its p95 is therefore the measured form of "this feature is free for
// operators who run no agent".

const (
	detailBenchIterations = 500
	detailBenchWarmup     = 50
	// One sample every 10s across the correlation window and a little either
	// side, which is what a real host at the default agent interval produces.
	detailBenchSamples = 60
)

func seedIncidentDetailFixture(b *testing.B, fx *internaltest.DialectFixture, withHost bool) (string, http.Handler) {
	b.Helper()
	ctx := context.Background()

	resourceRepo := store.NewResourceRepositorySQLC(fx.Runtime)
	incidentRepo := store.NewIncidentRepositorySQLC(fx.Runtime)
	eventStepRepo := store.NewIncidentEventStepRepositorySQLC(fx.Runtime)
	hostRepo := store.NewHostRepositorySQLC(fx.Runtime)
	hostMetricRepo := store.NewHostMetricRepositorySQLC(fx.Runtime)
	hostEventRepo := store.NewHostEventRepositorySQLC(fx.Runtime)

	suffix := "nohost"
	if withHost {
		suffix = "withhost"
	}

	startedAt := time.Now().Add(-10 * time.Minute)

	res := &domain.Resource{
		Base:     domain.Base{ID: "detail-bench-res-" + suffix, CreatedAt: startedAt},
		Name:     "detail-bench-" + suffix,
		Type:     domain.ResourceHTTP,
		Target:   "https://example.com",
		IsActive: true,
		Interval: 60,
		Timeout:  10,
	}

	if withHost {
		host := &domain.Host{Base: domain.Base{ID: "detail-bench-host"}, Name: "bench-host-01"}
		if err := hostRepo.Create(ctx, host); err != nil {
			b.Fatalf("seed host: %v", err)
		}
		hostID := host.ID
		res.HostID = &hostID

		// Samples spanning the window, each carrying two mounts so the disk
		// projection and its decoding are exercised too.
		for i := 0; i < detailBenchSamples; i++ {
			at := startedAt.Add(-5*time.Minute + time.Duration(i*10)*time.Second)
			if err := hostMetricRepo.Insert(ctx, &domain.HostMetricSample{
				HostID:    hostID,
				SampledAt: at,
				CPUPct:    float64(i % 100),
				MemPct:    float64((i * 3) % 100),
				Disks: []domain.DiskUsage{
					{Mount: "/", UsedPct: float64((i * 2) % 100)},
					{Mount: "/var", UsedPct: float64((i * 5) % 100)},
				},
			}); err != nil {
				b.Fatalf("seed sample %d: %v", i, err)
			}
		}
	}

	if _, err := resourceRepo.Create(ctx, res); err != nil {
		b.Fatalf("seed monitor: %v", err)
	}

	inc := &domain.Incident{
		Base:       domain.Base{ID: "detail-bench-inc-" + suffix, CreatedAt: startedAt, UpdatedAt: startedAt},
		ResourceID: res.ID,
		Cause:      "bench_failure",
		StartedAt:  startedAt,
	}
	if _, err := incidentRepo.Create(ctx, inc); err != nil {
		b.Fatalf("seed incident: %v", err)
	}

	svc := service.NewIncidentService(incidentRepo, eventStepRepo, hostMetricRepo, hostRepo, 7*24*time.Hour)
	// The correlator is attached because the shipped read path has it attached
	// (spec 091). Benchmarking the service without it would measure a
	// configuration nobody runs, and would leave the claim that a monitor with no
	// host pays nothing for correlation entirely unmeasured.
	svc = svc.WithCorrelator(correlation.New(hostEventRepo, hostRepo))
	handler := v1.NewIncidentHandler(svc)

	r := chi.NewRouter()
	r.Get("/api/v1/incidents/{id}", handler.Get)
	return inc.ID, r
}

func benchIncidentDetail(b *testing.B, name string, withHost bool) {
	fx := internaltest.SetupPostgres(b)
	if fx == nil {
		b.Skip("postgres backend unavailable")
		return
	}
	incidentID, router := seedIncidentDetailFixture(b, fx, withHost)
	req := func() *http.Request {
		return httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/incidents/%s", incidentID), nil)
	}

	for i := 0; i < detailBenchWarmup; i++ {
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req())
		if rr.Code != http.StatusOK {
			b.Fatalf("warm-up: unexpected status %d", rr.Code)
		}
	}

	durations := make([]time.Duration, 0, detailBenchIterations)
	b.ResetTimer()
	for i := 0; i < detailBenchIterations; i++ {
		start := time.Now()
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req())
		durations = append(durations, time.Since(start))
		if rr.Code != http.StatusOK {
			b.Fatalf("iter %d: unexpected status %d", i, rr.Code)
		}
	}
	b.StopTimer()

	reportP95(b, name, durations)
}

// SC-007 — an incident whose monitor has no host must not pay for the feature.
func BenchmarkAPI_IncidentDetail_NoHost_p95(b *testing.B) {
	benchIncidentDetail(b, "BenchmarkAPI_IncidentDetail_NoHost", false)
}

// SC-007a — with a host attached, the enrichment runs. Its cost against the
// no-host baseline is the number the 10% ceiling applies to.
func BenchmarkAPI_IncidentDetail_WithHost_p95(b *testing.B) {
	benchIncidentDetail(b, "BenchmarkAPI_IncidentDetail_WithHost", true)
}
