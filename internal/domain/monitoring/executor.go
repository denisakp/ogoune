package monitoring

import (
	"context"
	"fmt"

	"github.com/denisakp/pulseguard/internal/domain"
)

type Executor struct {
	strategies map[domain.ResourceType]Strategy
}

func NewExecutor(strategies map[domain.ResourceType]Strategy) *Executor {
	return &Executor{strategies: strategies}
}

func (e *Executor) ExecuteCheck(resource *domain.Resource) (Result, error) {
	strategy, ok := e.strategies[resource.Type]
	if !ok {
		return Result{}, fmt.Errorf("no strategy found for resource type: %s", resource.Type)
	}

	ctx := context.Background()

	return strategy.Execute(ctx, resource)
}