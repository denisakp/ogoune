package strategy

import (
	"context"
	"testing"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTCPStrategy_SetsStructuredCauseOnConnectionFailure(t *testing.T) {
	strategy := NewTCPStrategy(2 * time.Second)
	// Use a public IP with an unlikely port to test connection failure
	resource := &domain.Resource{Target: "93.184.216.34:1", Timeout: 1}

	result, err := strategy.Execute(context.Background(), resource)
	require.NoError(t, err)
	require.NotNil(t, result.Cause)
	assert.Equal(t, string(domain.StatusDown), result.Status)
}

func TestTCPStrategy_SSRFBlocksLoopback(t *testing.T) {
	strategy := NewTCPStrategy(2 * time.Second)
	resource := &domain.Resource{Target: "127.0.0.1:3306", Timeout: 1}

	result, err := strategy.Execute(context.Background(), resource)
	require.NoError(t, err)
	assert.Equal(t, string(domain.StatusDown), result.Status)
	assert.Contains(t, result.ResponseData, "blocked")
}

func TestTCPStrategy_SSRFBlocksPrivateIP(t *testing.T) {
	strategy := NewTCPStrategy(2 * time.Second)
	resource := &domain.Resource{Target: "10.0.0.1:5432", Timeout: 1}

	result, err := strategy.Execute(context.Background(), resource)
	require.NoError(t, err)
	assert.Equal(t, string(domain.StatusDown), result.Status)
	assert.Contains(t, result.ResponseData, "blocked")
}
