package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Abh19avM/recess/internal/httputil"
	"github.com/Abh19avM/recess/internal/users"
	"github.com/go-chi/chi/v5"
)

func TestAuthRegisterAndLogin(t *testing.T) {
	repo := users.NewInMemoryRepository()
	svc := NewService(repo)
	handler := NewHandler(svc)

	r := chi.NewRouter()
	r.Mount("/api/v1/auth", handler.Routes())

	// Test Register
	regBody, _ := json.Marshal(RegisterRequest{
		Username: "TestPlayer",
		Email:    "test@recess.local",
		Password: "password123",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(regBody))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d", rec.Code)
	}

	var regResp httputil.ResponseEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &regResp); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}
	if !regResp.Success {
		t.Errorf("expected success true")
	}

	// Test Login
	loginBody, _ := json.Marshal(LoginRequest{
		Username: "TestPlayer",
		Password: "password123",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginBody))
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", rec.Code)
	}

	// Test Invalid Login
	badLoginBody, _ := json.Marshal(LoginRequest{
		Username: "NonExistent",
		Password: "wrong",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(badLoginBody))
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401 Unauthorized, got %d", rec.Code)
	}
}

func TestGuestLogin(t *testing.T) {
	repo := users.NewInMemoryRepository()
	svc := NewService(repo)
	handler := NewHandler(svc)

	r := chi.NewRouter()
	r.Mount("/api/v1/auth", handler.Routes())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/guest", bytes.NewReader([]byte(`{"nickname":"DeskRacer"}`)))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d", rec.Code)
	}
}
