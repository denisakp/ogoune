package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository/fake"
	"github.com/denisakp/ogoune/internal/service"
)

func newToolboxSvc(t *testing.T, targets ...string) *service.ToolboxService {
	t.Helper()
	repo := fake.NewResourceFake()
	for i, tg := range targets {
		_, err := repo.Create(context.Background(), &domain.Resource{
			Name:   "r" + string(rune('a'+i)),
			Type:   domain.ResourceTCP,
			Target: tg,
		})
		if err != nil {
			t.Fatalf("seed resource: %v", err)
		}
	}
	return service.NewToolboxService(repo, 3*time.Second)
}

func TestToolboxDNS_Validation(t *testing.T) {
	svc := newToolboxSvc(t)
	ctx := context.Background()

	cases := []struct {
		name string
		q    service.ToolboxDNSQuery
		want error
	}{
		{"empty domain", service.ToolboxDNSQuery{Domain: ""}, service.ErrToolboxValidation},
		{"bad record type", service.ToolboxDNSQuery{Domain: "example.com", RecordTypes: []string{"WAT"}}, service.ErrToolboxValidation},
		{"unknown resolver", service.ToolboxDNSQuery{Domain: "example.com", Resolver: "nope"}, service.ErrToolboxValidation},
		{"custom resolver not ip", service.ToolboxDNSQuery{Domain: "example.com", Resolver: "custom", CustomResolver: "not-an-ip"}, service.ErrToolboxValidation},
		{"custom resolver private", service.ToolboxDNSQuery{Domain: "example.com", Resolver: "custom", CustomResolver: "10.0.0.1"}, service.ErrToolboxTargetBlocked},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.DNS(ctx, tc.q)
			if !errors.Is(err, tc.want) {
				t.Fatalf("got %v, want %v", err, tc.want)
			}
		})
	}
}

func TestToolboxPortScan_Validation(t *testing.T) {
	svc := newToolboxSvc(t)
	ctx := context.Background()

	tooMany := make([]int, 101)
	for i := range tooMany {
		tooMany[i] = i + 1
	}

	cases := []struct {
		name string
		q    service.ToolboxPortScanQuery
	}{
		{"empty target", service.ToolboxPortScanQuery{Ports: []int{80}}},
		{"no ports", service.ToolboxPortScanQuery{Target: "example.com"}},
		{"too many ports", service.ToolboxPortScanQuery{Target: "example.com", Ports: tooMany}},
		{"port out of range", service.ToolboxPortScanQuery{Target: "example.com", Ports: []int{70000}}},
		{"timeout too low", service.ToolboxPortScanQuery{Target: "example.com", Ports: []int{80}, TimeoutMs: 10}},
		{"timeout too high", service.ToolboxPortScanQuery{Target: "example.com", Ports: []int{80}, TimeoutMs: 9000}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.PortScan(ctx, tc.q)
			if !errors.Is(err, service.ErrToolboxValidation) {
				t.Fatalf("got %v, want validation error", err)
			}
		})
	}
}

func TestToolboxPortScan_TargetGating(t *testing.T) {
	ctx := context.Background()

	t.Run("unregistered refused before dial", func(t *testing.T) {
		svc := newToolboxSvc(t, "https://known.example.com")
		_, err := svc.PortScan(ctx, service.ToolboxPortScanQuery{
			Target: "unknown.example.org", Ports: []int{80}, TimeoutMs: 500,
		})
		if !errors.Is(err, service.ErrToolboxTargetNotRegistered) {
			t.Fatalf("got %v, want ErrToolboxTargetNotRegistered", err)
		}
	})

	t.Run("registered host matched via URL target normalization", func(t *testing.T) {
		// Seed a monitor whose Target is a full URL; the scanner must match by hostname.
		svc := newToolboxSvc(t, "https://known.example.com/health")
		_, err := svc.PortScan(ctx, service.ToolboxPortScanQuery{
			Target: "known.example.com", Ports: []int{1}, TimeoutMs: 100,
		})
		// Must NOT be refused as unregistered (it may fail to connect — that's fine).
		if errors.Is(err, service.ErrToolboxTargetNotRegistered) {
			t.Fatalf("registered host was wrongly refused: %v", err)
		}
	})
}

func TestToolboxSSL_Validation(t *testing.T) {
	svc := newToolboxSvc(t)
	ctx := context.Background()

	if _, err := svc.SSL(ctx, service.ToolboxSSLQuery{Domain: ""}); !errors.Is(err, service.ErrToolboxValidation) {
		t.Fatalf("empty domain: got %v", err)
	}
	if _, err := svc.SSL(ctx, service.ToolboxSSLQuery{Domain: "example.com", Port: 70000}); !errors.Is(err, service.ErrToolboxValidation) {
		t.Fatalf("bad port: got %v", err)
	}
}

func TestToolboxWHOIS_Validation(t *testing.T) {
	svc := newToolboxSvc(t)
	if _, err := svc.WHOIS(context.Background(), "  "); !errors.Is(err, service.ErrToolboxValidation) {
		t.Fatalf("empty domain: got %v", err)
	}
}
