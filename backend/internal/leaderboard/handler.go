package leaderboard

import (
	"net/http"
	"strconv"

	"github.com/Abh19avM/recess/internal/httputil"
	"github.com/go-chi/chi/v5"
)

// Handler handles HTTP requests for leaderboards and player progression.
type Handler struct {
	service Service
}

// NewHandler creates a new leaderboard HTTP handler.
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// Routes mounts leaderboard endpoints on a Chi router.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.GetGlobalLeaderboard)
	r.Get("/{game_type}", h.GetGameLeaderboard)

	return r
}

// UserRoutes mounts user history and achievement sub-endpoints.
func (h *Handler) UserRoutes() chi.Router {
	r := chi.NewRouter()

	r.Get("/{id}/history", h.GetMatchHistory)
	r.Get("/{id}/achievements", h.GetUserAchievements)

	return r
}

func (h *Handler) GetGlobalLeaderboard(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 20
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}

	ranks, err := h.service.GetLeaderboard(r.Context(), "global", limit)
	if err != nil {
		httputil.ErrorJSON(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to fetch global leaderboard")
		return
	}

	httputil.JSON(w, r, http.StatusOK, map[string]any{
		"game_type":    "global",
		"leaderboard": ranks,
		"total":       len(ranks),
	})
}

func (h *Handler) GetGameLeaderboard(w http.ResponseWriter, r *http.Request) {
	gameType := chi.URLParam(r, "game_type")
	limitStr := r.URL.Query().Get("limit")
	limit := 20
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}

	ranks, err := h.service.GetLeaderboard(r.Context(), gameType, limit)
	if err != nil {
		httputil.ErrorJSON(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to fetch game leaderboard")
		return
	}

	httputil.JSON(w, r, http.StatusOK, map[string]any{
		"game_type":    gameType,
		"leaderboard": ranks,
		"total":       len(ranks),
	})
}

func (h *Handler) GetMatchHistory(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	if userID == "" {
		httputil.ErrorJSON(w, r, http.StatusBadRequest, "BAD_REQUEST", "user ID is required")
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}

	history, err := h.service.GetMatchHistory(r.Context(), userID, limit)
	if err != nil {
		httputil.ErrorJSON(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to fetch match history")
		return
	}

	httputil.JSON(w, r, http.StatusOK, map[string]any{
		"user_id": userID,
		"matches": history,
		"total":   len(history),
	})
}

func (h *Handler) GetUserAchievements(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	if userID == "" {
		httputil.ErrorJSON(w, r, http.StatusBadRequest, "BAD_REQUEST", "user ID is required")
		return
	}

	achievements, err := h.service.GetUserAchievements(r.Context(), userID)
	if err != nil {
		httputil.ErrorJSON(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to fetch achievements")
		return
	}

	unlockedCount := 0
	for _, a := range achievements {
		if a.IsUnlocked {
			unlockedCount++
		}
	}

	httputil.JSON(w, r, http.StatusOK, map[string]any{
		"user_id":      userID,
		"achievements": achievements,
		"unlocked":     unlockedCount,
		"total":        len(achievements),
	})
}
