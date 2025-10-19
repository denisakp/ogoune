package domain

import (
	"context"
	"time"
)

// CheckResult represents the result of a health check execution.
type CheckResult struct {
	Status       string
	ResponseTime time.Duration
	ResponseData string
}

// CheckStrategy defines the interface for executing health checks on resources.
// Different resource types (HTTP, TCP, etc.) implement this interface.
type CheckStrategy interface {
	Execute(ctx context.Context, resource *Resource) (CheckResult, error)
}

// CheckExecutor executes health checks using the appropriate strategy for each resource type.
type CheckExecutor struct {
	strategies map[ResourceType]CheckStrategy
}

// NewCheckExecutor creates a new CheckExecutor with the given strategies.
func NewCheckExecutor(strategies map[ResourceType]CheckStrategy) *CheckExecutor {
	return &CheckExecutor{
		strategies: strategies,
	}
}

// ExecuteCheck executes a health check for the given resource using the appropriate strategy.
func (e *CheckExecutor) ExecuteCheck(resource *Resource) (CheckResult, error) {
	strategy, exists := e.strategies[resource.Type]
	if !exists {
		return CheckResult{
			Status:       string(StatusError),
			ResponseTime: 0,
			ResponseData: "unsupported resource type",
		}, nil
	}

	ctx := context.Background()
	return strategy.Execute(ctx, resource)
}
