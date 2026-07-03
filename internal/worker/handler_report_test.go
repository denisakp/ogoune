package worker

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/hibiken/asynq"
)

type stubReportGen struct {
	periods []string
	err     error
}

func (s *stubReportGen) GenerateAndDeliver(_ context.Context, period string) error {
	s.periods = append(s.periods, period)
	return s.err
}

func TestReportWorker_TargetsPreviousMonth(t *testing.T) {
	gen := &stubReportGen{}
	h := NewReportTaskHandler(gen)
	if err := h.ProcessTask(context.Background(), asynq.NewTask(TypeReportCheck, nil)); err != nil {
		t.Fatalf("ProcessTask returned error: %v", err)
	}
	if len(gen.periods) != 1 {
		t.Fatalf("want 1 generate call, got %d", len(gen.periods))
	}
	// The period must be the calendar month before now (UTC), matching the service key.
	want := previousMonthPeriod(time.Now().UTC())
	if gen.periods[0] != want {
		t.Fatalf("period = %q, want %q", gen.periods[0], want)
	}
}

func TestReportWorker_SwallowsErrors(t *testing.T) {
	gen := &stubReportGen{err: errors.New("boom")}
	h := NewReportTaskHandler(gen)
	if err := h.ProcessTask(context.Background(), asynq.NewTask(TypeReportCheck, nil)); err != nil {
		t.Fatalf("ProcessTask must not propagate errors, got %v", err)
	}
}

func TestPreviousMonthPeriod(t *testing.T) {
	cases := map[string]string{
		"2026-07-01": "2026-06",
		"2026-01-15": "2025-12",
		"2026-03-31": "2026-02",
	}
	for in, want := range cases {
		now, _ := time.Parse("2006-01-02", in)
		if got := previousMonthPeriod(now.UTC()); got != want {
			t.Errorf("previousMonthPeriod(%s) = %s, want %s", in, got, want)
		}
	}
}
