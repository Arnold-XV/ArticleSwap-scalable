package middleware

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"articleswap-scalable/services/internal/cache"
)

// RateLimit provides per-IP rate limiting using a Redis sliding window counter.
func RateLimit(rc *cache.RedisCache, maxPerMinute int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := extractIP(r)
			key := fmt.Sprintf("ratelimit:%s", ip)

			count, err := rc.Incr(r.Context(), key)
			if err != nil {
				// On Redis error, allow the request through (fail-open).
				next.ServeHTTP(w, r)
				return
			}

			// Set TTL on first request in the window.
			if count == 1 {
				_ = rc.Expire(r.Context(), key, time.Minute)
			}

			if count > int64(maxPerMinute) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", "60")
				w.WriteHeader(http.StatusTooManyRequests)
				fmt.Fprintf(w, `{"error":"rate limit exceeded, max %d requests per minute"}`, maxPerMinute)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func extractIP(r *http.Request) string {
	// Check X-Forwarded-For first (behind reverse proxy).
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
