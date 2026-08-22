package rooms

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Abh19avM/recess/internal/httputil"
	"github.com/go-chi/chi/v5"
)

func TestRoomsEndpoints(t *testing.T) {
	repo := NewInMemoryRepository()
	svc := NewService(repo)
	handler := NewHandler(svc)

	r := chi.NewRouter()
	r.Mount("/api/v1/rooms", handler.Routes())

	// Test List Rooms
	req := httptest.NewRequest(http.MethodGet, "/api/v1/rooms", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	// Test Create Room
	createBody, _ := json.Marshal(CreateRoomRequest{
		GameType: "dots_boxes",
		Title:    "5x5 Quick Match",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/rooms", bytes.NewReader(createBody))
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d", rec.Code)
	}

	var resp httputil.ResponseEnvelope
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)

	// Test Get Seeded Room
	req = httptest.NewRequest(http.MethodGet, "/api/v1/rooms/RECESS-CRIC", nil)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	// Test Join Room
	req = httptest.NewRequest(http.MethodPost, "/api/v1/rooms/RECESS-CRIC/join", bytes.NewReader([]byte(`{}`)))
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}
