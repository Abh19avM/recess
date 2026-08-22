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
			httputil.ErrorJSON(w, r, http.StatusConflict, "USER_EXISTS", "username is already taken")
			return
		}
		httputil.ErrorJSON(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "registration failed")
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
			httputil.ErrorJSON(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "invalid username or password")
			return
		}
		httputil.ErrorJSON(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "login failed")
		return
	}

	httputil.JSON(w, r, http.StatusOK, resp)
}

func (h *Handler) Guest(w http.ResponseWriter, r *http.Request) {
	var req GuestRequest
	_ = json.NewDecoder(r.Body).Decode(&req) // Optional body

	resp, err := h.service.CreateGuest(r.Context(), req)
	if err != nil {
		httputil.ErrorJSON(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create guest session")
		return
	}

	httputil.JSON(w, r, http.StatusCreated, resp)
}
