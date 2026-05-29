package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func withUser(req *http.Request, userID string) *http.Request {
	return req.WithContext(context.WithValue(req.Context(), "user_id", userID))
}

func TestPerUserRateLimit_AllowsUpToBurst(t *testing.T) {
	mw := PerUserRateLimit(10)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	for i := 0; i < 10; i++ {
		req := withUser(httptest.NewRequest(http.MethodPost, "/test", nil), "user-1")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusNoContent, rr.Code, "request %d should be allowed", i+1)
	}
}

func TestPerUserRateLimit_BlocksAfterBurst(t *testing.T) {
	mw := PerUserRateLimit(10)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	for i := 0; i < 10; i++ {
		req := withUser(httptest.NewRequest(http.MethodPost, "/test", nil), "user-1")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
	}
	// 11th request: rejected
	req := withUser(httptest.NewRequest(http.MethodPost, "/test", nil), "user-1")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusTooManyRequests, rr.Code)
	assert.NotEmpty(t, rr.Header().Get("Retry-After"))
}

func TestPerUserRateLimit_IndependentBucketsPerUser(t *testing.T) {
	mw := PerUserRateLimit(10)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	// Exhaust user-A's bucket
	for i := 0; i < 10; i++ {
		handler.ServeHTTP(httptest.NewRecorder(), withUser(httptest.NewRequest(http.MethodPost, "/t", nil), "user-A"))
	}
	// user-A: should now be blocked
	rrA := httptest.NewRecorder()
	handler.ServeHTTP(rrA, withUser(httptest.NewRequest(http.MethodPost, "/t", nil), "user-A"))
	assert.Equal(t, http.StatusTooManyRequests, rrA.Code)

	// user-B: independent bucket, should still pass
	rrB := httptest.NewRecorder()
	handler.ServeHTTP(rrB, withUser(httptest.NewRequest(http.MethodPost, "/t", nil), "user-B"))
	assert.Equal(t, http.StatusNoContent, rrB.Code)
}

func TestPerUserRateLimit_NoUserIDPassesThrough(t *testing.T) {
	mw := PerUserRateLimit(1)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	// 5 requests without user_id; none should be limited
	for i := 0; i < 5; i++ {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/t", nil))
		assert.Equal(t, http.StatusNoContent, rr.Code)
	}
}
