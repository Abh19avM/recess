package users

import (
	"errors"
	"net/http"

	"github.com/Abh19avM/recess/internal/httputil"
	"github.com/go-chi/chi/v5"
)

// Handler handles HTTP requests for user endpoints.
type Handler struct {
	service Service
}

// NewHandler creates a new user HTTP handler.
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// Routes returns the subrouter for /api/v1/users.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/{id}", h.GetUser)
	r.Get("/{id}/profile", h.GetProfile)

	return r
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
