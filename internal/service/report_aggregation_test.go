package service

import (
	"testing"

	domain "github.com/denisakp/ogoune/internal/domain"
)

func TestAggregatePeriod_KnownTotals(t *testing.T) {
	resources := []*domain.Resource{
		{Base: domain.Base{ID: "r1"}, Name: "API", Interval: 60},
		{Base: domain.Base{ID: "r2"}, Name: "DB", Interval: 30},
	}
	// r1: 90 up / 10 down over the period → 90% uptime, 10*60=600s down.
	// r2: 100 up / 0 down → 100% uptime, 0s down.
	aggs := []*domain.UptimeDailyAgg{
		{ResourceID: "r1", Up: 40, Down: 5},
		{ResourceID: "r1", Up: 50, Down: 5},
		{ResourceID: "r2", Up: 100},
	}
	incidents := map[string]int{"r1": 2}

	uptime, downtime, breakdown := aggregatePeriod(resources, aggs, incidents)

	// overall: up=190, counted=200 → 95%
	if uptime != 95.0 {
		t.Fatalf("overall uptime = %v, want 95.0", uptime)
	}
	if downtime != 600 {
		t.Fatalf("downtime = %d, want 600", downtime)
	}
	if len(breakdown) != 2 {
		t.Fatalf("breakdown len = %d, want 2", len(breakdown))
	}
	// sorted by name: API then DB
	if breakdown[0].Name != "API" || breakdown[0].UptimePct != 90.0 || breakdown[0].Incidents != 2 {
		t.Fatalf("API line = %+v", breakdown[0])
	}
	if breakdown[1].Name != "DB" || breakdown[1].UptimePct != 100.0 || breakdown[1].Incidents != 0 {
		t.Fatalf("DB line = %+v", breakdown[1])
	}
}

func TestAggregatePeriod_NoData(t *testing.T) {
	resources := []*domain.Resource{{Base: domain.Base{ID: "r1"}, Name: "API", Interval: 60}}
	uptime, downtime, breakdown := aggregatePeriod(resources, nil, nil)
	if uptime != 0 || downtime != 0 || len(breakdown) != 0 {
		t.Fatalf("empty period should be zeros; got uptime=%v downtime=%d breakdown=%d", uptime, downtime, len(breakdown))
	}
}
