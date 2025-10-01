package monitoring

import (
	"context"
	"time"

	"github.com/denisakp/pulseguard/internal/domain"
)

type Result struct {
	Status       string
	ResponseTime time.Duration
	ResponseData string
}

type Strategy interface {
	Execute(ctx context.Context, resource *domain.Resource) (Result, error)
}
