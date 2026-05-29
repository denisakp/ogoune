package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/denisakp/ogoune/pkg/problemdetail"
	"golang.org/x/time/rate"
)

// PerUserRateLimit returns a middleware that limits the authenticated user to
// `requestsPerMinute` requests per minute (burst = requestsPerMinute). It is
// keyed off the `user_id` value injected by AuthMiddleware. Unauthenticated
// requests reach this middleware only when the route is misconfigured, in
// which case the middleware lets them through (auth enforcement is upstream).
//
// On overflow it returns HTTP 429 with an RFC 7807 problem and a Retry-After
// header denominated in seconds.
//
// The limiter map is unbounded; for the credential test endpoint the volume is
// negligible (operators run a handful of tests per session). If this is ever
// re-used on a higher-volume route, add a TTL sweeper.
func PerUserRateLimit(requestsPerMinute int) func(http.Handler) http.Handler {
	var (
		mu       sync.Mutex
		limiters = make(map[string]*rate.Limiter)
	)
	perRequest := time.Minute / time.Duration(requestsPerMinute)
	rateLimit := rate.Every(perRequest)
	burst := requestsPerMinute

	getLimiter := func(userID string) *rate.Limiter {
		mu.Lock()
		defer mu.Unlock()
		l, ok := limiters[userID]
		if !ok {
			l = rate.NewLimiter(rateLimit, burst)
			limiters[userID] = l
		}
		return l
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, _ := r.Context().Value("user_id").(string)
			if userID == "" {
				next.ServeHTTP(w, r)
				return
			}
			l := getLimiter(userID)
			if !l.Allow() {
				retryAfter := int(perRequest.Seconds())
				if retryAfter < 1 {
					retryAfter = 1
				}
				w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
				pd := problemdetail.New(
					"RATE_LIMIT_EXCEEDED",
					http.StatusText(http.StatusTooManyRequests),
					http.StatusTooManyRequests,
					fmt.Sprintf("Rate limit exceeded: max %d requests/minute. Retry after %d seconds.", requestsPerMinute, retryAfter),
				)
				problemdetail.Write(w, pd)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
