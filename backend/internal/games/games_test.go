package games

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Abh19avM/recess/internal/httputil"
	"github.com/go-chi/chi/v5"
)

func TestGamesCatalogEndpoints(t *testing.T) {
	svc := NewService()
	handler := NewHandler(svc)

	r := chi.NewRouter()
	r.Mount("/api/v1/games", handler.Routes())

	// Test List Games
	req := httptest.NewRequest(http.MethodGet, "/api/v1/games", nil)
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

	// Test Get Specific Game
	req = httptest.NewRequest(http.MethodGet, "/api/v1/games/hand_cricket", nil)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	// Test Invalid Game Type
	req = httptest.NewRequest(http.MethodGet, "/api/v1/games/non_existent_game", nil)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}
}
