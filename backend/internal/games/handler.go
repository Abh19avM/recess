package games

import (
	"errors"
	"net/http"

	"github.com/Abh19avM/recess/internal/httputil"
	"github.com/go-chi/chi/v5"
)

// Handler handles HTTP requests for game catalog and metadata.
type Handler struct {
	service Service
}

// NewHandler creates a new games HTTP handler.
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// Routes returns the subrouter for /api/v1/games.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.ListGames)
	r.Get("/{type}", h.GetGame)

	return r
}

func (h *Handler) ListGames(w http.ResponseWriter, r *http.Request) {
	gamesList, err := h.service.ListGames(r.Context())
	if err != nil {
		httputil.ErrorJSON(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list games")
		return
	}

	httputil.JSON(w, r, http.StatusOK, map[string]any{
		"games": gamesList,
		"count": len(gamesList),
	})
}

func (h *Handler) GetGame(w http.ResponseWriter, r *http.Request) {
	gType := GameType(chi.URLParam(r, "type"))
	if gType == "" {
		httputil.ErrorJSON(w, r, http.StatusBadRequest, "BAD_REQUEST", "game type is required")
		return
	}

	gameInfo, err := h.service.GetGameByType(r.Context(), gType)
	if err != nil {
		if errors.Is(err, ErrGameNotFound) {
			httputil.ErrorJSON(w, r, http.StatusNotFound, "NOT_FOUND", "game type not found")
			return
		}
		httputil.ErrorJSON(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to retrieve game info")
		return
	}

	httputil.JSON(w, r, http.StatusOK, gameInfo)
}
