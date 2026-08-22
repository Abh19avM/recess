package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Abh19avM/recess/internal/config"
	"github.com/Abh19avM/recess/internal/middleware"
)

func TestHealthEndpoint(t *testing.T) {
	cfg := config.Load()
	serverState := &Server{
		cfg:   cfg,
		db:    nil,
		redis: nil,
	}

	router := SetupRouter(serverState)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", rec.Code)
	}

	var resp HealthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response JSON: %v", err)
	}

	if resp.Status != "ok" {
		t.Errorf("expected status 'ok', got %q", resp.Status)
	}

	if resp.Environment != cfg.Environment {
		t.Errorf("expected environment %q, got %q", cfg.Environment, resp.Environment)
	}

	if resp.Timestamp == "" {
		t.Errorf("expected non-empty timestamp")
	}

	// Verify X-Request-ID is set
	if reqID := rec.Header().Get(middleware.RequestIDHeader); reqID == "" {
		t.Errorf("expected %s header to be present", middleware.RequestIDHeader)
	}
}

func TestReadyEndpointWhenDependenciesDown(t *testing.T) {
	cfg := config.Load()
	serverState := &Server{
		cfg:   cfg,
		db:    nil, // Simulating uninitialized database
		redis: nil, // Simulating uninitialized redis
	}

	router := SetupRouter(serverState)

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// Since dependencies are nil/down, /ready MUST return 503 Service Unavailable
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503 Service Unavailable, got %d", rec.Code)
	}

	var resp ReadyResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response JSON: %v", err)
	}

	if resp.Status != "unavailable" {
		t.Errorf("expected status 'unavailable', got %q", resp.Status)
	}

	if dbCheck, ok := resp.Checks["database"]; !ok || dbCheck.Status != "down" {
		t.Errorf("expected database check to be 'down', got %+v", dbCheck)
	}

	if redisCheck, ok := resp.Checks["redis"]; !ok || redisCheck.Status != "down" {
		t.Errorf("expected redis check to be 'down', got %+v", redisCheck)
	}
}

func TestRootEndpoint(t *testing.T) {
	cfg := config.Load()
	serverState := &Server{
		cfg: cfg,
	}

	router := SetupRouter(serverState)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}
