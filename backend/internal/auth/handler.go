package auth

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Abh19avM/recess/internal/httputil"
	"github.com/Abh19avM/recess/internal/users"
	"github.com/go-chi/chi/v5"
)

// Handler handles HTTP requests for auth endpoints.
type Handler struct {
	service Service
}

// NewHandler creates a new auth HTTP handler.
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// Routes returns the subrouter for /api/v1/auth.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/register", h.Register)
	r.Post("/login", h.Login)
	r.Post("/refresh", h.Refresh)
	r.Post("/logout", h.Logout)
	r.Post("/guest", h.Guest)

	return r
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.ErrorJSON(w, r, http.StatusBadRequest, "INVALID_BODY", "failed to parse request body")
		return
	}

	if req.Username == "" || req.Password == "" {
		httputil.ValidationErrorJSON(w, r, "username and password are required", map[string]string{
			"username": "required",
			"password": "required",
		})
		return
	}

	resp, err := h.service.Register(r.Context(), req)
	if err != nil {
		if errors.Is(err, users.ErrUserAlreadyExists) {
			httputil.ErrorJSON(w, r, http.StatusConflict, "USER_EXISTS", "a student with this username or email already exists")
			return
		}
		httputil.ErrorJSON(w, r, http.StatusBadRequest, "REGISTRATION_FAILED", err.Error())
		return
	}

	httputil.JSON(w, r, http.StatusCreated, resp)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.ErrorJSON(w, r, http.StatusBadRequest, "INVALID_BODY", "failed to parse request body")
		return
	}

	resp, err := h.service.Login(r.Context(), req)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			httputil.ErrorJSON(w, r, http.StatusUnauthorized, "INVALID_CREDENTIALS", "invalid username or password")
			return
		}
		httputil.ErrorJSON(w, r, http.StatusInternalServerError, "LOGIN_FAILED", "failed to process login")
		return
	}

	httputil.JSON(w, r, http.StatusOK, resp)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.ErrorJSON(w, r, http.StatusBadRequest, "INVALID_BODY", "failed to parse request body")
		return
	}

	if req.RefreshToken == "" {
		httputil.ValidationErrorJSON(w, r, "refresh_token is required", map[string]string{
			"refresh_token": "required",
		})
		return
	}

	resp, err := h.service.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, ErrInvalidRefreshToken) {
			httputil.ErrorJSON(w, r, http.StatusUnauthorized, "INVALID_REFRESH_TOKEN", "refresh token is invalid, expired, or revoked")
			return
		}
		if errors.Is(err, ErrUserNotFound) {
			httputil.ErrorJSON(w, r, http.StatusUnauthorized, "USER_NOT_FOUND", "user account not found")
			return
		}
		httputil.ErrorJSON(w, r, http.StatusInternalServerError, "REFRESH_FAILED", "failed to refresh token")
		return
	}

	httputil.JSON(w, r, http.StatusOK, resp)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var req LogoutRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	if req.RefreshToken != "" {
		_ = h.service.Logout(r.Context(), req.RefreshToken)
	}

	httputil.JSON(w, r, http.StatusOK, map[string]string{
		"message": "logged out successfully",
	})
}

func (h *Handler) Guest(w http.ResponseWriter, r *http.Request) {
	var req GuestRequest
	_ = json.NewDecoder(r.Body).Decode(&req) // Optional body

	resp, err := h.service.CreateGuest(r.Context(), req)
	if err != nil {
		httputil.ErrorJSON(w, r, http.StatusInternalServerError, "GUEST_FAILED", "failed to create guest session")
		return
	}

	httputil.JSON(w, r, http.StatusCreated, resp)
}
