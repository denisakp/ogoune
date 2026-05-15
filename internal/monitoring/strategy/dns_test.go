package strategy

import (
	"context"
	"testing"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDNSStrategy_SetsStructuredCauseOnLookupFailure(t *testing.T) {
	strategy := NewDNSStrategy(2 * time.Second)
	resource := &domain.Resource{Target: "definitely-not-a-real-host.invalid"}

	result, err := strategy.Execute(context.Background(), resource)
	require.NoError(t, err)
	require.NotNil(t, result.Cause)
	assert.Equal(t, domain.DNSResolutionFailed, *result.Cause)
	assert.Equal(t, string(domain.StatusDown), result.Status)
}
