package handler

import (
	"sync"
	"time"
)

type pingWindowCounter struct {
	window int64
	count  int
}

// PingRateLimiter implements a simple in-memory fixed-window limiter per heartbeat slug.
type PingRateLimiter struct {
	mu       sync.Mutex
	max      int
	per      time.Duration
	counters map[string]pingWindowCounter
}

func NewPingRateLimiter(max int, per time.Duration) *PingRateLimiter {
	return &PingRateLimiter{
		max:      max,
		per:      per,
		counters: make(map[string]pingWindowCounter),
	}
}

func (l *PingRateLimiter) Allow(key string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	window := now.Unix() / int64(l.per.Seconds())
	counter := l.counters[key]
	if counter.window != window {
		counter.window = window
		counter.count = 0
	}
	if counter.count >= l.max {
		l.counters[key] = counter
		return false
	}
	counter.count++
	l.counters[key] = counter
	return true
}
