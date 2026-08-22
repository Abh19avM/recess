package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestLoggerAssignsRequestID(t *testing.T) {
	handler := RequestLogger()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := GetRequestID(r.Context())
		if reqID == "" {
			t.Error("expected request ID in context, got empty")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	headerID := rec.Header().Get(RequestIDHeader)
	if headerID == "" {
		t.Errorf("expected %s header in response", RequestIDHeader)
	}
}

func TestRequestLoggerPreservesExistingRequestID(t *testing.T) {
	customID := "custom-req-id-12345"
	handler := RequestLogger()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := GetRequestID(r.Context())
		if reqID != customID {
			t.Errorf("expected request ID %s, got %s", customID, reqID)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(RequestIDHeader, customID)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Header().Get(RequestIDHeader) != customID {
		t.Errorf("expected header %s, got %s", customID, rec.Header().Get(RequestIDHeader))
	}
}

func TestRequestLoggerRecoversPanic(t *testing.T) {
	handler := RequestLogger()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("simulated panic in handler")
	}))

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()

	// Should not crash the process
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500 on recovered panic, got %d", rec.Code)
	}
}
