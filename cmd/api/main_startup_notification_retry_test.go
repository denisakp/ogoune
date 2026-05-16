package main

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/denisakp/ogoune/internal/service"
	"github.com/stretchr/testify/assert"
)

type startupRetryStub struct {
	called bool
	limit  int
	err    error
	sum    service.PendingNotificationRetrySummary
}

func (s *startupRetryStub) RetryPendingNotifications(ctx context.Context, limit int) (service.PendingNotificationRetrySummary, error) {
	s.called = true
	s.limit = limit
	if s.err != nil {
		return service.PendingNotificationRetrySummary{}, s.err
	}
	return s.sum, nil
}

func TestRunStartupPendingNotificationRetry_InvokesServiceWithStartupLimit(t *testing.T) {
	stub := &startupRetryStub{sum: service.PendingNotificationRetrySummary{ScannedCount: 3, RetriedCount: 2, ExpiredCount: 1}}

	runStartupPendingNotificationRetry(context.Background(), stub)

	assert.True(t, stub.called)
	assert.Equal(t, 1000, stub.limit)
}

func TestRunStartupPendingNotificationRetry_LogsWarningOnErrorAndDoesNotPanic(t *testing.T) {
	stub := &startupRetryStub{err: errors.New("redis unavailable")}

	var buf bytes.Buffer
	oldDefault := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))
	defer slog.SetDefault(oldDefault)

	runStartupPendingNotificationRetry(context.Background(), stub)

	assert.True(t, stub.called)
	assert.Contains(t, buf.String(), "pending notification retry failed")
	assert.Contains(t, buf.String(), "redis unavailable")
}

func TestRunStartupPendingNotificationRetry_LogsSummaryForNoPendingRows(t *testing.T) {
	stub := &startupRetryStub{sum: service.PendingNotificationRetrySummary{ScannedCount: 0, RetriedCount: 0, ExpiredCount: 0, FailedCount: 0, SkippedClaimedCount: 0}}

	var buf bytes.Buffer
	oldDefault := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))
	defer slog.SetDefault(oldDefault)

	runStartupPendingNotificationRetry(context.Background(), stub)

	assert.Contains(t, buf.String(), "scanned=0")
	assert.Contains(t, buf.String(), "retried=0")
	assert.Contains(t, buf.String(), "failed=0")
}

func TestRunStartupPendingNotificationRetry_NoServiceNoop(t *testing.T) {
	runStartupPendingNotificationRetry(context.Background(), nil)
}
