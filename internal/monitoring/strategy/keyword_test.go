package strategy

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func keywordResource(url, keyword, mode string) *domain.Resource {
	return &domain.Resource{Target: url, Timeout: 5, Keyword: &keyword, KeywordMode: &mode}
}

// TC1: contains mode — keyword present → UP
func TestKeywordStrategy_ContainsMode_KeywordPresent_UP(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "System status: operational and running fine")
	}))
	defer ts.Close()

	s := NewKeywordStrategy(5 * time.Second)
	result, err := s.Execute(context.Background(), keywordResource(ts.URL, "operational", "contains"))
	require.NoError(t, err)
	assert.Equal(t, string(domain.StatusUp), result.Status)
	assert.Nil(t, result.Cause)
	require.NotNil(t, result.KeywordContext)
	assert.True(t, result.KeywordContext.KeywordFound)
}

// TC2: contains mode — keyword absent → DOWN with keyword_not_found
func TestKeywordStrategy_ContainsMode_KeywordAbsent_DOWN(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "System status: degraded")
	}))
	defer ts.Close()

	s := NewKeywordStrategy(5 * time.Second)
	result, err := s.Execute(context.Background(), keywordResource(ts.URL, "operational", "contains"))
	require.NoError(t, err)
	assert.Equal(t, string(domain.StatusDown), result.Status)
	require.NotNil(t, result.Cause)
	assert.Equal(t, domain.KeywordNotFound, *result.Cause)
	require.NotNil(t, result.KeywordContext)
	assert.False(t, result.KeywordContext.KeywordFound)
}

// TC3: not_contains mode — keyword absent → UP
func TestKeywordStrategy_NotContainsMode_KeywordAbsent_UP(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Everything looks good")
	}))
	defer ts.Close()

	s := NewKeywordStrategy(5 * time.Second)
	result, err := s.Execute(context.Background(), keywordResource(ts.URL, "error", "not_contains"))
	require.NoError(t, err)
	assert.Equal(t, string(domain.StatusUp), result.Status)
	assert.Nil(t, result.Cause)
}

// TC4: not_contains mode — keyword present → DOWN with keyword_found
func TestKeywordStrategy_NotContainsMode_KeywordPresent_DOWN(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Internal server error occurred")
	}))
	defer ts.Close()

	s := NewKeywordStrategy(5 * time.Second)
	result, err := s.Execute(context.Background(), keywordResource(ts.URL, "error", "not_contains"))
	require.NoError(t, err)
	assert.Equal(t, string(domain.StatusDown), result.Status)
	require.NotNil(t, result.Cause)
	assert.Equal(t, domain.KeywordFound, *result.Cause)
}

// TC5: HTTP failure → DOWN with HTTP cause, keyword not evaluated
func TestKeywordStrategy_HTTPFailure_NoKeywordEvaluation(t *testing.T) {
	s := NewKeywordStrategy(1 * time.Second)
	result, err := s.Execute(context.Background(), keywordResource("http://127.0.0.1:19999", "operational", "contains"))
	require.NoError(t, err)
	assert.Equal(t, string(domain.StatusDown), result.Status)
	require.NotNil(t, result.Cause)
	assert.NotEqual(t, domain.KeywordNotFound, *result.Cause)
	assert.NotEqual(t, domain.KeywordFound, *result.Cause)
	assert.Nil(t, result.KeywordContext)
}

// TC6: body exactly 512 KB → not truncated
func TestKeywordStrategy_BodyExactly512KB_NotTruncated(t *testing.T) {
	body := strings.Repeat("a", 512*1024)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, body)
	}))
	defer ts.Close()

	s := NewKeywordStrategy(10 * time.Second)
	result, err := s.Execute(context.Background(), keywordResource(ts.URL, "zzz", "contains"))
	require.NoError(t, err)
	assert.False(t, result.BodyTruncated)
	assert.Equal(t, int64(512*1024), result.ReadBodySize)
}

// TC7: body exceeds 512 KB → BodyTruncated = true
func TestKeywordStrategy_BodyExceeds512KB_Truncated(t *testing.T) {
	body := strings.Repeat("a", 512*1024+1)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, body)
	}))
	defer ts.Close()

	s := NewKeywordStrategy(10 * time.Second)
	result, err := s.Execute(context.Background(), keywordResource(ts.URL, "zzz", "contains"))
	require.NoError(t, err)
	assert.True(t, result.BodyTruncated)
}

// TC8: empty body + contains mode → DOWN
func TestKeywordStrategy_EmptyBody_ContainsMode_DOWN(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// empty body
	}))
	defer ts.Close()

	s := NewKeywordStrategy(5 * time.Second)
	result, err := s.Execute(context.Background(), keywordResource(ts.URL, "operational", "contains"))
	require.NoError(t, err)
	assert.Equal(t, string(domain.StatusDown), result.Status)
	require.NotNil(t, result.Cause)
	assert.Equal(t, domain.KeywordNotFound, *result.Cause)
}

// TC9: case-sensitivity — keyword "Operational", body contains only "operational" → DOWN
func TestKeywordStrategy_CaseSensitive_MismatchedCase_DOWN(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "system is operational")
	}))
	defer ts.Close()

	s := NewKeywordStrategy(5 * time.Second)
	result, err := s.Execute(context.Background(), keywordResource(ts.URL, "Operational", "contains"))
	require.NoError(t, err)
	assert.Equal(t, string(domain.StatusDown), result.Status)
	require.NotNil(t, result.Cause)
	assert.Equal(t, domain.KeywordNotFound, *result.Cause)
}
