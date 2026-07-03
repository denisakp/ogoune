package service

import (
	"context"
	"errors"
	"testing"
	"time"

	domain "github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository/fake"
)

func newTestReportService() (*ReportService, *fake.ReportSettingsFake, *fake.ReportHistoryFake, *fake.ResourceFake, *fake.UptimeDailyAggRepository, *fake.IncidentFake, *fake.NotificationChannelFake) {
	sf := fake.NewReportSettingsFake()
	hf := fake.NewReportHistoryFake()
	rf := fake.NewResourceFake()
	uf := fake.NewUptimeDailyAggRepository()
	inf := fake.NewIncidentFake()
	cf := fake.NewNotificationChannelFake()
	svc := NewReportService(sf, hf, rf, uf, inf, cf)
	return svc, sf, hf, rf, uf, inf, cf
}

// ── US1: config ──

func TestReport_GetSettings_DefaultWhenUnsaved(t *testing.T) {
	svc, _, _, _, _, _, _ := newTestReportService()
	got, err := svc.GetSettings(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got.Enabled || got.Schedule != domain.ReportScheduleMonthly1st || got.Scope != domain.ReportScopeAllResources {
		t.Fatalf("unexpected default: %+v", got)
	}
}

func TestReport_SaveSettings_PersistsAndValidates(t *testing.T) {
	svc, _, _, _, _, _, _ := newTestReportService()
	ctx := context.Background()

	if _, err := svc.SaveSettings(ctx, &domain.ReportSettings{Enabled: true, RecipientEmail: "ops@example.com"}); err != nil {
		t.Fatalf("save valid: %v", err)
	}
	got, _ := svc.GetSettings(ctx)
	if !got.Enabled || got.RecipientEmail != "ops@example.com" {
		t.Fatalf("not persisted: %+v", got)
	}

	// enabled + empty recipient → validation error
	if _, err := svc.SaveSettings(ctx, &domain.ReportSettings{Enabled: true, RecipientEmail: ""}); !errors.Is(err, ErrReportValidation) {
		t.Fatalf("want ErrReportValidation, got %v", err)
	}
	// disabled + empty recipient → allowed
	if _, err := svc.SaveSettings(ctx, &domain.ReportSettings{Enabled: false, RecipientEmail: ""}); err != nil {
		t.Fatalf("disabled empty recipient should be allowed: %v", err)
	}
}

// ── US2: generate + deliver ──

func seedPeriodData(t *testing.T, rf *fake.ResourceFake, uf *fake.UptimeDailyAggRepository, inf *fake.IncidentFake) {
	t.Helper()
	ctx := context.Background()
	if _, err := rf.Create(ctx, &domain.Resource{Base: domain.Base{ID: "r1"}, Name: "API", Interval: 60}); err != nil {
		t.Fatal(err)
	}
	day := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
	if err := uf.Upsert(ctx, &domain.UptimeDailyAgg{ResourceID: "r1", Day: day, Samples: 100, Up: 95, Down: 5}); err != nil {
		t.Fatal(err)
	}
	if _, err := inf.Create(ctx, &domain.Incident{Base: domain.Base{ID: "i1"}, ResourceID: "r1", StartedAt: day}); err != nil {
		t.Fatal(err)
	}
}

func TestReport_GenerateAndDeliver_HappyPath(t *testing.T) {
	svc, _, hf, rf, uf, inf, _ := newTestReportService()
	ctx := context.Background()
	if _, err := svc.SaveSettings(ctx, &domain.ReportSettings{Enabled: true, RecipientEmail: "ops@example.com"}); err != nil {
		t.Fatal(err)
	}
	seedPeriodData(t, rf, uf, inf)

	var calls int
	svc.deliver = func(_ context.Context, recipient, period string, _ float64, _ int, _ int64, _ []domain.ReportBreakdownLine) error {
		calls++
		return nil
	}

	if err := svc.GenerateAndDeliver(ctx, "2026-06"); err != nil {
		t.Fatalf("generate: %v", err)
	}
	rows, _ := hf.ListRecent(ctx, 10)
	if len(rows) != 1 || rows[0].Status != domain.ReportStatusDelivered {
		t.Fatalf("want 1 delivered row, got %+v", rows)
	}
	if rows[0].IncidentCount != 1 || rows[0].DowntimeSeconds != 300 { // 5 down * 60s
		t.Fatalf("totals wrong: %+v", rows[0])
	}
	got, _ := svc.GetSettings(ctx)
	if got.LastSentAt == nil {
		t.Fatal("lastSentAt not advanced")
	}

	// idempotent: second run adds nothing, no second delivery
	if err := svc.GenerateAndDeliver(ctx, "2026-06"); err != nil {
		t.Fatal(err)
	}
	rows, _ = hf.ListRecent(ctx, 10)
	if len(rows) != 1 || calls != 1 {
		t.Fatalf("idempotency broken: rows=%d calls=%d", len(rows), calls)
	}
}

func TestReport_GenerateAndDeliver_DeliveryFailureRecordedFailed(t *testing.T) {
	svc, _, hf, rf, uf, inf, _ := newTestReportService()
	ctx := context.Background()
	_, _ = svc.SaveSettings(ctx, &domain.ReportSettings{Enabled: true, RecipientEmail: "ops@example.com"})
	seedPeriodData(t, rf, uf, inf)

	svc.deliver = func(_ context.Context, _, _ string, _ float64, _ int, _ int64, _ []domain.ReportBreakdownLine) error {
		return errors.New("smtp down")
	}
	if err := svc.GenerateAndDeliver(ctx, "2026-06"); err != nil {
		t.Fatalf("delivery failure must not abort: %v", err)
	}
	rows, _ := hf.ListRecent(ctx, 10)
	if len(rows) != 1 || rows[0].Status != domain.ReportStatusFailed {
		t.Fatalf("want 1 failed row, got %+v", rows)
	}
	got, _ := svc.GetSettings(ctx)
	if got.LastSentAt != nil {
		t.Fatal("lastSentAt must not advance on failure")
	}
}

func TestReport_GenerateAndDeliver_NoSMTPChannel_Failed(t *testing.T) {
	// Uses the real smtpDeliver with an empty channel fake → no transport → failed.
	svc, _, hf, rf, uf, inf, _ := newTestReportService()
	ctx := context.Background()
	_, _ = svc.SaveSettings(ctx, &domain.ReportSettings{Enabled: true, RecipientEmail: "ops@example.com"})
	seedPeriodData(t, rf, uf, inf)

	if err := svc.GenerateAndDeliver(ctx, "2026-06"); err != nil {
		t.Fatalf("must not abort: %v", err)
	}
	rows, _ := hf.ListRecent(ctx, 10)
	if len(rows) != 1 || rows[0].Status != domain.ReportStatusFailed {
		t.Fatalf("want 1 failed row (no smtp channel), got %+v", rows)
	}
}

func TestReport_GenerateAndDeliver_DisabledIsNoop(t *testing.T) {
	svc, _, hf, _, _, _, _ := newTestReportService()
	ctx := context.Background()
	_, _ = svc.SaveSettings(ctx, &domain.ReportSettings{Enabled: false, RecipientEmail: ""})
	if err := svc.GenerateAndDeliver(ctx, "2026-06"); err != nil {
		t.Fatal(err)
	}
	rows, _ := hf.ListRecent(ctx, 10)
	if len(rows) != 0 {
		t.Fatalf("disabled must generate nothing, got %d", len(rows))
	}
}

// ── US3: history + preview ──

func TestReport_ListHistory_ClampAndOrder(t *testing.T) {
	svc, _, hf, _, _, _, _ := newTestReportService()
	ctx := context.Background()
	base := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	for i, p := range []string{"2026-01", "2026-02", "2026-03"} {
		_, err := hf.Create(ctx, &domain.ReportHistory{Period: p, SentAt: base.AddDate(0, i, 0), Status: domain.ReportStatusDelivered})
		if err != nil {
			t.Fatal(err)
		}
	}
	rows, err := svc.ListHistory(ctx, 0) // default
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 3 || rows[0].Period != "2026-03" {
		t.Fatalf("want 3 newest-first, got %+v", rows)
	}
}

func TestReport_GeneratePreview_NotPersisted(t *testing.T) {
	svc, _, hf, rf, uf, inf, _ := newTestReportService()
	ctx := context.Background()
	_, _ = svc.SaveSettings(ctx, &domain.ReportSettings{Enabled: true, RecipientEmail: "ops@example.com"})
	seedPeriodData(t, rf, uf, inf)

	pv, err := svc.GeneratePreview(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if pv.Status != domain.ReportStatusPending {
		t.Fatalf("preview status = %s, want pending", pv.Status)
	}
	rows, _ := hf.ListRecent(ctx, 10)
	if len(rows) != 0 {
		t.Fatalf("preview must not persist, got %d rows", len(rows))
	}
}
