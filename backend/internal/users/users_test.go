package users

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Abh19avM/recess/internal/httputil"
	"github.com/go-chi/chi/v5"
)

func TestGetUserEndpoint(t *testing.T) {
	repo := NewInMemoryRepository()
	svc := NewService(repo)
	handler := NewHandler(svc, nil)

	r := chi.NewRouter()
	r.Mount("/api/v1/users", handler.Routes())

	// Test existing user
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/usr_demo_headmaster", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp httputil.ResponseEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}

	if !resp.Success {
		t.Errorf("expected success true")
	}

	// Test non-existing user
	req = httptest.NewRequest(http.MethodGet, "/api/v1/users/non_existing_user", nil)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}
}

func TestGetUserProfileEndpoint(t *testing.T) {
	repo := NewInMemoryRepository()
	svc := NewService(repo)
	handler := NewHandler(svc, nil)

	r := chi.NewRouter()
	r.Mount("/api/v1/users", handler.Routes())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/usr_demo_headmaster/profile", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}
