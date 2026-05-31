package internaltest

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Synthetic paired-bench tests. We can't easily test RunPairedBench against
// a real *testing.B (the harness calls b.Fatalf on regression which would
// fail the surrounding test). Instead we test the building blocks: percentile
// math, BenchResult.String format, and runHalf timing capture.

func TestPercentile_BasicCases(t *testing.T) {
	samples := []time.Duration{
		10 * time.Millisecond,
		20 * time.Millisecond,
		30 * time.Millisecond,
		40 * time.Millisecond,
		50 * time.Millisecond,
		60 * time.Millisecond,
		70 * time.Millisecond,
		80 * time.Millisecond,
		90 * time.Millisecond,
		100 * time.Millisecond,
	}
	// p95 over 10 samples: idx = 9 (clamped from 10) → 100ms
	assert.Equal(t, 100*time.Millisecond, percentile(samples, 0.95))
	// p50 over 10 samples: idx = 5 → 60ms
	assert.Equal(t, 60*time.Millisecond, percentile(samples, 0.50))
	// Empty slice → 0
	assert.Equal(t, time.Duration(0), percentile(nil, 0.95))
}

func TestBenchResult_StringFormat(t *testing.T) {
	r := BenchResult{
		Name:       "BenchmarkFoo",
		GormP95:    12345 * time.Microsecond,
		SqlcP95:    12888 * time.Microsecond,
		Ratio:      1.044,
		Iterations: 1000,
		Warmup:     50,
	}
	s := r.String()
	assert.True(t, strings.HasPrefix(s, "paired_bench "))
	assert.Contains(t, s, "name=BenchmarkFoo")
	assert.Contains(t, s, "gorm_p95_us=12345")
	assert.Contains(t, s, "sqlc_p95_us=12888")
	assert.Contains(t, s, "ratio=1.044")
	assert.Contains(t, s, "iterations=1000")
	assert.Contains(t, s, "warmup=50")
}

func TestRunHalf_CapturesSamples(t *testing.T) {
	var calls int
	fn := func() {
		calls++
		time.Sleep(time.Microsecond)
	}
	samples := runHalf(fn)
	// Warmup (50) + sample (1000) calls.
	assert.Equal(t, defaultBenchWarmup+defaultBenchIterations, calls)
	assert.Len(t, samples, defaultBenchIterations)
	// Every sample should be > 0 (we slept at least a microsecond).
	for i, s := range samples {
		assert.Greater(t, s, time.Duration(0), "sample[%d] should be > 0", i)
	}
}
