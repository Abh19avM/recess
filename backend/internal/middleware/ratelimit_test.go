package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestIPRateLimiter(t *testing.T) {
	// Allow maximum 3 requests per 100ms
	limiter := NewIPRateLimiter(3, 100*time.Millisecond)

	dummyHandler := limiter.Handler()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.0.2.1:12345"
		rec := httptest.NewRecorder()
		dummyHandler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("request %d expected status 200, got %d", i+1, rec.Code)
		}
	}

	// 4th request from same IP should be blocked with 429
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.0.2.1:12345"
	rec := httptest.NewRecorder()
	dummyHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("4th request expected status 429 Too Many Requests, got %d", rec.Code)
	}

	if retryAfter := rec.Header().Get("Retry-After"); retryAfter == "" {
		t.Errorf("expected Retry-After header on 429 response")
	}

	// Request from another IP should still succeed
	reqOther := httptest.NewRequest(http.MethodGet, "/test", nil)
	reqOther.RemoteAddr = "198.51.100.1:54321"
	recOther := httptest.NewRecorder()
	dummyHandler.ServeHTTP(recOther, reqOther)

	if recOther.Code != http.StatusOK {
		t.Fatalf("request from different IP expected status 200, got %d", recOther.Code)
	}
}
