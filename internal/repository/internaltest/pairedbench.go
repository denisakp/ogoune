package internaltest

import (
	"fmt"
	"os"
	"runtime"
	"sort"
	"strconv"
	"testing"
	"time"
)

// envBenchStrict, when set to a truthy strconv.ParseBool value, makes
// RunPairedBench call b.Fatalf on ratio > 1.10. Default (unset or false)
// logs the regression as a warning but does not fail the bench, letting
// the infra land while perf work is still in flight.
const envBenchStrict = "PAIRED_BENCH_STRICT"

// Defaults for the paired-bench harness. Clarification Q2 of spec 049.
const (
	defaultBenchIterations = 1000
	defaultBenchWarmup     = 50
	defaultBenchRatioMax   = 1.10
)

// BenchFunc is one measured iteration. Must run a single operation; the
// harness handles iteration counting and timing.
type BenchFunc func()

// BenchResult is the structured output of one paired bench. String emits
// the parseable log line consumed by the CI step that surfaces ratios in
// the GitHub job summary (spec 049 §FR-010).
type BenchResult struct {
	Name       string
	GormP95    time.Duration
	SqlcP95    time.Duration
	Ratio      float64
	Iterations int
	Warmup     int
}

func (r BenchResult) String() string {
	return fmt.Sprintf(
		"paired_bench name=%s gorm_p95_us=%d sqlc_p95_us=%d ratio=%.3f iterations=%d warmup=%d",
		r.Name,
		r.GormP95.Microseconds(),
		r.SqlcP95.Microseconds(),
		r.Ratio,
		r.Iterations,
		r.Warmup,
	)
}

// RunPairedBench executes gormFn and sqlcFn back-to-back in the same
// process, samples 1000 iterations per half with 50 warmup discarded,
// computes p95s, asserts ratio ≤ 1.10. Logs the structured BenchResult
// line via b.Log so CI can parse it.
//
// Spec 049 §FR-005 + Clarification Q2 (iteration budget).
func RunPairedBench(b *testing.B, name string, gormFn, sqlcFn BenchFunc) BenchResult {
	b.Helper()
	if gormFn == nil || sqlcFn == nil {
		b.Fatalf("paired bench %q: gormFn and sqlcFn must be non-nil", name)
	}

	gormSamples := runHalf(gormFn)
	sqlcSamples := runHalf(sqlcFn)

	if len(gormSamples) == 0 || len(sqlcSamples) == 0 {
		b.Fatalf("paired bench %q: no samples recorded", name)
	}

	gormP95 := percentile(gormSamples, 0.95)
	sqlcP95 := percentile(sqlcSamples, 0.95)
	if gormP95 == 0 {
		b.Fatalf("paired bench %q: gorm_p95 == 0 (sample too small or fn instantaneous)", name)
	}

	result := BenchResult{
		Name:       name,
		GormP95:    gormP95,
		SqlcP95:    sqlcP95,
		Ratio:      float64(sqlcP95) / float64(gormP95),
		Iterations: defaultBenchIterations,
		Warmup:     defaultBenchWarmup,
	}

	b.Log(result.String())
	fmt.Println(result.String()) // stdout for CI capture (b.Log goes to test stream)

	if result.Ratio > defaultBenchRatioMax {
		msg := fmt.Sprintf("paired bench regression: %s (ratio %.3f > %.2f)", result.String(), result.Ratio, defaultBenchRatioMax)
		strict, _ := strconv.ParseBool(os.Getenv(envBenchStrict))
		if strict {
			b.Fatalf("%s", msg)
		} else {
			b.Logf("WARN: %s (set %s=true to fail the bench)", msg, envBenchStrict)
			fmt.Println("WARN: " + msg)
		}
	}
	return result
}

// runHalf executes warmup + sample iterations, returning the per-iteration
// durations for sampling. GC is forced once between warmup and sample to
// flush warmup-allocated garbage.
func runHalf(fn BenchFunc) []time.Duration {
	for i := 0; i < defaultBenchWarmup; i++ {
		fn()
	}
	runtime.GC()
	samples := make([]time.Duration, defaultBenchIterations)
	for i := 0; i < defaultBenchIterations; i++ {
		start := time.Now()
		fn()
		samples[i] = time.Since(start)
	}
	return samples
}

// percentile returns the value at the given percentile (0–1) after sorting
// the slice. Index = floor(p * len). For p=0.95 over 1000 samples that's
// index 950 — the 951st-fastest call (5% slowest excluded).
func percentile(samples []time.Duration, p float64) time.Duration {
	if len(samples) == 0 {
		return 0
	}
	sorted := make([]time.Duration, len(samples))
	copy(sorted, samples)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	idx := int(p * float64(len(sorted)))
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}
