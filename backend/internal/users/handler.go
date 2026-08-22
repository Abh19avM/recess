package users

import (
	"errors"
	"net/http"

	"github.com/Abh19avM/recess/internal/httputil"
	"github.com/Abh19avM/recess/internal/middleware"
	"github.com/go-chi/chi/v5"
)

// Handler handles HTTP requests for user endpoints.
type Handler struct {
	service     Service
	authBarrier func(http.Handler) http.Handler
}

// NewHandler creates a new user HTTP handler with optional auth middleware.
func NewHandler(service Service, authBarrier func(http.Handler) http.Handler) *Handler {
	return &Handler{
		service:     service,
		authBarrier: authBarrier,
	}
}

// Routes returns the subrouter for /api/v1/users.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	// Protected current user endpoint
	if h.authBarrier != nil {
		r.With(h.authBarrier).Get("/me", h.GetMe)
	} else {
		r.Get("/me", h.GetMe)
	}

	// Public profile lookups
	r.Get("/{id}", h.GetUser)
	r.Get("/{id}/profile", h.GetProfile)

	return r
}

func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetAuthenticatedUserID(r.Context())
	if !ok || userID == "" {
		httputil.ErrorJSON(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	profile, err := h.service.GetUserProfile(r.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			httputil.ErrorJSON(w, r, http.StatusNotFound, "USER_NOT_FOUND", "user account not found")
			return
		}
		httputil.ErrorJSON(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to retrieve current profile")
		return
	}

	httputil.JSON(w, r, http.StatusOK, profile)
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		httputil.ErrorJSON(w, r, http.StatusBadRequest, "BAD_REQUEST", "user ID is required")
		return
	}

	user, err := h.service.GetUserByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			httputil.ErrorJSON(w, r, http.StatusNotFound, "NOT_FOUND", "user not found")
			return
		}
		httputil.ErrorJSON(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to retrieve user")
		return
	}

	httputil.JSON(w, r, http.StatusOK, user)
}

func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		httputil.ErrorJSON(w, r, http.StatusBadRequest, "BAD_REQUEST", "user ID is required")
		return
	}

	profile, err := h.service.GetUserProfile(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			httputil.ErrorJSON(w, r, http.StatusNotFound, "NOT_FOUND", "user profile not found")
			return
		}
		httputil.ErrorJSON(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to retrieve user profile")
		return
	}

	httputil.JSON(w, r, http.StatusOK, profile)
}
