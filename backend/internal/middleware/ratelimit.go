package middleware

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Abh19avM/recess/internal/httputil"
)

// IPRateLimiter provides sliding-window rate limiting per client IP.
type IPRateLimiter struct {
	mu      sync.Mutex
	limits  map[string][]time.Time
	limit   int
	window  time.Duration
	cleanup time.Duration
}

// NewIPRateLimiter creates a new rate limiter with the given maximum requests per window.
func NewIPRateLimiter(limit int, window time.Duration) *IPRateLimiter {
	limiter := &IPRateLimiter{
		limits:  make(map[string][]time.Time),
		limit:   limit,
		window:  window,
		cleanup: 5 * time.Minute,
	}

	go limiter.cleaner()
	return limiter
}

// Handler returns an HTTP middleware enforcing rate limits.
func (l *IPRateLimiter) Handler() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := extractIP(r)

			l.mu.Lock()
			now := time.Now()
			cutoff := now.Add(-l.window)

			// Clean expired timestamps for this IP
			var valid []time.Time
			for _, t := range l.limits[ip] {
				if t.After(cutoff) {
					valid = append(valid, t)
				}
			}

			if len(valid) >= l.limit {
				l.limits[ip] = valid
				l.mu.Unlock()

				retryAfterSec := int(l.window.Seconds())
				if retryAfterSec < 1 {
					retryAfterSec = 1
				}
				w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfterSec))
				httputil.ErrorJSON(w, r, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "too many requests, please slow down")
				return
			}

			l.limits[ip] = append(valid, now)
			l.mu.Unlock()

			next.ServeHTTP(w, r)
		})
	}
}

func (l *IPRateLimiter) cleaner() {
	ticker := time.NewTicker(l.cleanup)
	for range ticker.C {
		l.mu.Lock()
		now := time.Now()
		cutoff := now.Add(-l.window)
		for ip, timestamps := range l.limits {
			var valid []time.Time
			for _, t := range timestamps {
				if t.After(cutoff) {
					valid = append(valid, t)
				}
			}
			if len(valid) == 0 {
				delete(l.limits, ip)
			} else {
				l.limits[ip] = valid
			}
		}
		l.mu.Unlock()
	}
}

func extractIP(r *http.Request) string {
	// Check X-Forwarded-For
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			ip := strings.TrimSpace(parts[0])
			if ip != "" {
				return ip
			}
		}
	}

	// Check X-Real-IP
	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return strings.TrimSpace(xrip)
	}

	// Fallback to RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}
