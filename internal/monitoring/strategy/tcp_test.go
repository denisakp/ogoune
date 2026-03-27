package strategy

import (
	"context"
	"testing"
	"time"

	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTCPStrategy_SetsStructuredCauseOnConnectionFailure(t *testing.T) {
	strategy := NewTCPStrategy(2 * time.Second)
	resource := &domain.Resource{Target: "127.0.0.1:1", Timeout: 1}

	result, err := strategy.Execute(context.Background(), resource)
	require.NoError(t, err)
	require.NotNil(t, result.Cause)
	assert.Equal(t, string(domain.StatusDown), result.Status)
}
