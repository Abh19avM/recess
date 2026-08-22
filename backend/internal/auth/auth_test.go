package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Abh19avM/recess/internal/httputil"
	"github.com/Abh19avM/recess/internal/middleware"
	"github.com/Abh19avM/recess/internal/users"
	"github.com/go-chi/chi/v5"
)

func TestFullAuthFlow(t *testing.T) {
	userRepo := users.NewInMemoryRepository()
	jwtMgr := NewJWTManager("test_secret_key_1234567890")
	tokenStore := NewInMemoryTokenStore()
	svc := NewService(userRepo, jwtMgr, tokenStore)
	authHandler := NewHandler(svc)

	authBarrier := middleware.RequireAuth(jwtMgr)
	userHandler := users.NewHandler(users.NewService(userRepo), authBarrier)

	r := chi.NewRouter()
	r.Mount("/api/v1/auth", authHandler.Routes())
	r.Mount("/api/v1/users", userHandler.Routes())

	// 1. Registration
	regBody, _ := json.Marshal(RegisterRequest{
		Username: "ClassChampion",
		Email:    "champion@school.test",
		Password: "SecurePassword123!",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(regBody))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created on registration, got %d: %s", rec.Code, rec.Body.String())
	}

	var regEnvelope struct {
		Success bool         `json:"success"`
		Data    AuthResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &regEnvelope); err != nil {
		t.Fatalf("failed to decode register response: %v", err)
	}

	accessToken := regEnvelope.Data.Tokens.AccessToken
	refreshToken := regEnvelope.Data.Tokens.RefreshToken

	if accessToken == "" || refreshToken == "" {
		t.Fatal("expected non-empty access and refresh tokens")
	}

	// 2. Duplicate Registration (Should return 409 Conflict)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(regBody))
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 Conflict on duplicate registration, got %d", rec.Code)
	}

	// 3. Login with correct password
	loginBody, _ := json.Marshal(LoginRequest{
		Username: "ClassChampion",
		Password: "SecurePassword123!",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginBody))
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on valid login, got %d", rec.Code)
	}

	// 4. Login with invalid password (Should return 401 Unauthorized)
	badLoginBody, _ := json.Marshal(LoginRequest{
		Username: "ClassChampion",
		Password: "WrongPassword!",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(badLoginBody))
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized on bad credentials, got %d", rec.Code)
	}

	// 5. Protected Endpoint /api/v1/users/me without token (Should return 401 Unauthorized)
	req = httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized on protected route without token, got %d", rec.Code)
	}

	// 6. Protected Endpoint /api/v1/users/me WITH valid Bearer token
	req = httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on protected route with valid token, got %d: %s", rec.Code, rec.Body.String())
	}

	// 7. Token Refresh (Should rotate tokens)
	refBody, _ := json.Marshal(RefreshRequest{
		RefreshToken: refreshToken,
	})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(refBody))
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on token refresh, got %d: %s", rec.Code, rec.Body.String())
	}

	var refreshEnvelope struct {
		Success bool         `json:"success"`
		Data    AuthResponse `json:"data"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &refreshEnvelope)
	newRefreshToken := refreshEnvelope.Data.Tokens.RefreshToken

	// Old refresh token must now be invalid due to token rotation
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(refBody))
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized when reusing rotated refresh token, got %d", rec.Code)
	}

	// 8. Logout
	logoutBody, _ := json.Marshal(LogoutRequest{
		RefreshToken: newRefreshToken,
	})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", bytes.NewReader(logoutBody))
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on logout, got %d", rec.Code)
	}

	// Refresh after logout should fail
	refAfterLogoutBody, _ := json.Marshal(RefreshRequest{
		RefreshToken: newRefreshToken,
	})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(refAfterLogoutBody))
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized on refresh after logout, got %d", rec.Code)
	}
}

func TestGuestAuth(t *testing.T) {
	userRepo := users.NewInMemoryRepository()
	jwtMgr := NewJWTManager("secret_key_123")
	tokenStore := NewInMemoryTokenStore()
	svc := NewService(userRepo, jwtMgr, tokenStore)
	handler := NewHandler(svc)

	r := chi.NewRouter()
	r.Mount("/api/v1/auth", handler.Routes())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/guest", bytes.NewReader([]byte(`{"nickname":"QuickRunner"}`)))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created on guest login, got %d", rec.Code)
	}

	var resp httputil.ResponseEnvelope
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if !resp.Success {
		t.Errorf("expected success: true")
	}
}
