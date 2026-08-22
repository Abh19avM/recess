package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Abh19avM/recess/internal/app"
	"github.com/Abh19avM/recess/internal/config"
	"github.com/Abh19avM/recess/internal/httputil"
	"github.com/Abh19avM/recess/internal/middleware"
)

func TestAppHealthEndpoint(t *testing.T) {
	cfg := config.Load()
	application, err := app.New(context.Background(), cfg)
	if err != nil {
		t.Fatalf("failed to initialize app: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	application.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", rec.Code)
	}

	var resp app.HealthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response JSON: %v", err)
	}

	if resp.Status != "ok" {
		t.Errorf("expected status 'ok', got %q", resp.Status)
	}

	if reqID := rec.Header().Get(middleware.RequestIDHeader); reqID == "" {
		t.Errorf("expected %s header to be present", middleware.RequestIDHeader)
	}
}

func TestAppReadyEndpoint(t *testing.T) {
	cfg := config.Load()
	application, err := app.New(context.Background(), cfg)
	if err != nil {
		t.Fatalf("failed to initialize app: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()

	application.Router().ServeHTTP(rec, req)

	// Since local Postgres/Redis aren't running in this unit test context, ready returns 503
	if rec.Code != http.StatusServiceUnavailable && rec.Code != http.StatusOK {
		t.Fatalf("expected status 503 or 200, got %d", rec.Code)
	}

	var resp app.ReadyResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response JSON: %v", err)
	}
}

func TestAPIV1RoutesMounted(t *testing.T) {
	cfg := config.Load()
	application, err := app.New(context.Background(), cfg)
	if err != nil {
		t.Fatalf("failed to initialize app: %v", err)
	}

	endpoints := []string{
		"/api/v1/games",
		"/api/v1/rooms",
		"/api/v1/users/usr_demo_headmaster",
	}

	for _, ep := range endpoints {
		req := httptest.NewRequest(http.MethodGet, ep, nil)
		rec := httptest.NewRecorder()

		application.Router().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected endpoint %s to return 200 OK, got %d", ep, rec.Code)
		}

		var resp httputil.ResponseEnvelope
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Errorf("endpoint %s failed to return valid ResponseEnvelope: %v", ep, err)
		}
		if !resp.Success {
			t.Errorf("endpoint %s expected success: true", ep)
		}
	}
}
