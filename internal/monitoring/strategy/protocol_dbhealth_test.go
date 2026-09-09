package strategy

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// T015 -- the allowance's arithmetic. That it is actually enforced in flight is
// T020's job; this covers how it is computed.
func TestDbHealthBudget(t *testing.T) {
	t.Run("is a quarter of the monitor's timeout", func(t *testing.T) {
		// 4s timeout, plenty of context left -> 1s, under the 2s cap.
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		assert.Equal(t, time.Second, dbHealthBudget(ctx, 4*time.Second))
	})

	t.Run("a generous timeout is capped in absolute terms", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
		defer cancel()
		// A quarter of 60s would be 15s; the cap holds it to 2s so introspection
		// can never become the dominant cost of a check.
		assert.Equal(t, dbHealthBudgetCap, dbHealthBudget(ctx, 60*time.Second))
	})

	t.Run("a short timeout is never overrun", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		got := dbHealthBudget(ctx, 2*time.Second)
		assert.Equal(t, 500*time.Millisecond, got)
		assert.Less(t, got, 2*time.Second, "collection never claims the whole check budget")
	})

	t.Run("never promises more than the context has left", func(t *testing.T) {
		// A generous configured timeout, but the liveness check already burned
		// almost all of it: the allowance follows what is actually left.
		ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
		defer cancel()
		got := dbHealthBudget(ctx, 30*time.Second)
		assert.LessOrEqual(t, got, 300*time.Millisecond)
		assert.Greater(t, got, time.Duration(0))
	})

	t.Run("an exhausted context yields no allowance", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
		defer cancel()
		time.Sleep(2 * time.Millisecond)
		assert.Zero(t, dbHealthBudget(ctx, 30*time.Second),
			"collection is skipped entirely rather than attempted with nothing left")
	})

	t.Run("a timeout too small to be worth trying yields nothing", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		// 40ms/4 = 10ms, below the floor: the round trip alone would consume it.
		assert.Zero(t, dbHealthBudget(ctx, 40*time.Millisecond))
	})

	t.Run("a zero or negative timeout yields nothing", func(t *testing.T) {
		assert.Zero(t, dbHealthBudget(context.Background(), 0))
		assert.Zero(t, dbHealthBudget(context.Background(), -time.Second))
	})

	t.Run("works without a context deadline", func(t *testing.T) {
		assert.Equal(t, time.Second, dbHealthBudget(context.Background(), 4*time.Second))
	})
}
