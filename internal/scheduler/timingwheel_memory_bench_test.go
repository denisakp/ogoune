//go:build linux

package scheduler

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestTimingWheelMemoryRSSDelta(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running memory benchmark in short mode")
	}
	if os.Getenv("OGOUNE_RUN_SC002") != "1" {
		t.Skip("set OGOUNE_RUN_SC002=1 to execute the 5 minute SC-002 RSS benchmark")
	}

	baseline, err := measureAverageRSS(nil)
	if err != nil {
		t.Fatalf("baseline RSS measurement failed: %v", err)
	}
	resources := make([]ScheduleItem, 0, 500)
	for index := 0; index < 500; index++ {
		resources = append(resources, ScheduleItem{
			ResourceID: fmt.Sprintf("resource-%03d", index),
			Interval:   5 * time.Minute,
		})
	}
	loaded, err := measureAverageRSS(resources)
	if err != nil {
		t.Fatalf("loaded RSS measurement failed: %v", err)
	}

	delta := loaded - baseline
	if delta >= 5*1024*1024 {
		t.Fatalf("SC-002 failed: RSS delta = %d bytes, want < 5242880 bytes", delta)
	}

	t.Logf("SC-002 passed: baseline=%d bytes loaded=%d bytes delta=%d bytes", baseline, loaded, delta)
}

func measureAverageRSS(resources []ScheduleItem) (int64, error) {
	cfg := &Config{
		Mode: ModeTimingWheel,
		TimingWheel: TimingWheelConfig{
			TickInterval:          time.Second,
			MaxWorkers:            1,
			ShutdownTimeout:       15 * time.Second,
			NotificationQueueSize: 100,
		},
	}
	runtime.GC()

	tw, err := NewTimingWheel(cfg)
	if err != nil {
		return 0, err
	}
	ctx := context.Background()
	if err := tw.Start(ctx, NewMockRepository(resources, nil)); err != nil {
		return 0, err
	}
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = tw.Stop(stopCtx)
	}()

	const totalDuration = 5 * time.Minute
	const sampleInterval = 5 * time.Second
	samples := int(totalDuration / sampleInterval)
	var total int64
	for sample := 0; sample < samples; sample++ {
		time.Sleep(sampleInterval)
		rss, err := currentRSSBytes()
		if err != nil {
			return 0, err
		}
		total += rss
	}

	return total / int64(samples), nil
}

func currentRSSBytes() (int64, error) {
	data, err := os.ReadFile("/proc/self/statm")
	if err != nil {
		return 0, err
	}
	fields := strings.Fields(string(data))
	if len(fields) < 2 {
		return 0, fmt.Errorf("unexpected /proc/self/statm format: %q", string(data))
	}
	pages, err := strconv.ParseInt(fields[1], 10, 64)
	if err != nil {
		return 0, err
	}
	return pages * int64(os.Getpagesize()), nil
}
