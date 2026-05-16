package strategy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPStrategy_SSRFBlocksDirectLoopback(t *testing.T) {
	// Directly target a loopback URL — SafeTransport blocks the connection
	strategy := NewHTTPStrategy(2 * time.Second)
	resource := &domain.Resource{Target: "http://127.0.0.1:1/path", Timeout: 2}

	result, err := strategy.Execute(context.Background(), resource)
	require.NoError(t, err)
	assert.Equal(t, string(domain.StatusDown), result.Status)
	assert.Contains(t, result.ErrorMessage, "blocked")
}

func TestHTTPStrategy_SetsStructuredCauseOnHTTPFailure(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer ts.Close()

	strategy := NewHTTPStrategyWithClient(&http.Client{Timeout: 3 * time.Second})
	resource := &domain.Resource{Target: ts.URL, Timeout: 3}

	result, err := strategy.Execute(context.Background(), resource)
	require.NoError(t, err)
	require.NotNil(t, result.Cause)
	assert.Equal(t, domain.HTTPInvalidStatusCode, *result.Cause)
	assert.Equal(t, string(domain.StatusDown), result.Status)
}
