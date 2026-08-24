package matchmaking

import (
	"encoding/json"
	"net/http"

	"github.com/Abh19avM/recess/internal/httputil"
	"github.com/Abh19avM/recess/internal/middleware"
	"github.com/go-chi/chi/v5"
)

// Handler exposes matchmaking HTTP endpoints.
type Handler struct {
	service     Service
	authBarrier func(http.Handler) http.Handler
}

// NewHandler creates a new matchmaking HTTP handler.
func NewHandler(service Service, authBarrier func(http.Handler) http.Handler) *Handler {
	return &Handler{
		service:     service,
		authBarrier: authBarrier,
	}
}

// Routes mounts matchmaking endpoints on a Chi subrouter.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	// Protected endpoints (require auth or guest token)
	if h.authBarrier != nil {
		r.Use(h.authBarrier)
	}

	r.Post("/join", h.JoinQueue)
	r.Post("/leave", h.LeaveQueue)
	r.Get("/status", h.GetStatus)
	r.Get("/size", h.GetQueueSize)

	return r
}

// JoinQueue places the authenticated player into matchmaking.
func (h *Handler) JoinQueue(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetAuthenticatedUser(r.Context())
	if !ok || claims == nil {
		httputil.ErrorJSON(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	var req QueueJoinRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.ErrorJSON(w, r, http.StatusBadRequest, "INVALID_BODY", "failed to parse request payload")
		return
	}

	if req.GameType == "" {
		httputil.ValidationErrorJSON(w, r, "game_type is required", map[string]string{
			"game_type": "required",
		})
		return
	}
	if req.Mode == "" {
		req.Mode = ModeCasual
	}

	ticket := &MatchTicket{
		UserID:   claims.UserID,
		Username: claims.Username,
		GameType: req.GameType,
		Mode:     req.Mode,
		Rating:   1000,
	}

	match, err := h.service.Enqueue(r.Context(), ticket)
	if err != nil {
		httputil.ErrorJSON(w, r, http.StatusInternalServerError, "QUEUE_ERROR", err.Error())
		return
	}

	if match != nil {
		httputil.JSON(w, r, http.StatusOK, QueueJoinResponse{
			Status:   "matched",
			TicketID: ticket.TicketID,
			Match:    match,
		})
		return
	}

	httputil.JSON(w, r, http.StatusAccepted, QueueJoinResponse{
		Status:   "queued",
		TicketID: ticket.TicketID,
	})
}

// LeaveQueue removes the player from any active matchmaking queue.
func (h *Handler) LeaveQueue(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetAuthenticatedUser(r.Context())
	if !ok || claims == nil {
		httputil.ErrorJSON(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	var req struct {
		GameType string `json:"game_type"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	if err := h.service.Dequeue(r.Context(), claims.UserID, req.GameType); err != nil {
		httputil.ErrorJSON(w, r, http.StatusInternalServerError, "DEQUEUE_ERROR", err.Error())
		return
	}

	httputil.JSON(w, r, http.StatusOK, map[string]string{
		"status": "cancelled",
	})
}

// GetStatus checks whether a ticket has been successfully paired.
func (h *Handler) GetStatus(w http.ResponseWriter, r *http.Request) {
	ticketID := r.URL.Query().Get("ticket_id")
	if ticketID == "" {
		httputil.ErrorJSON(w, r, http.StatusBadRequest, "MISSING_TICKET", "ticket_id query parameter is required")
		return
	}

	match, matched, err := h.service.PollMatch(r.Context(), ticketID)
	if err != nil {
		httputil.ErrorJSON(w, r, http.StatusInternalServerError, "POLL_ERROR", err.Error())
		return
	}

	if matched && match != nil {
		httputil.JSON(w, r, http.StatusOK, QueueJoinResponse{
			Status:   "matched",
			TicketID: ticketID,
			Match:    match,
		})
		return
	}

	httputil.JSON(w, r, http.StatusOK, QueueJoinResponse{
		Status:   "queued",
		TicketID: ticketID,
	})
}

// GetQueueSize returns the active count in a queue.
func (h *Handler) GetQueueSize(w http.ResponseWriter, r *http.Request) {
	gameType := r.URL.Query().Get("game_type")
	mode := MatchMode(r.URL.Query().Get("mode"))
	if mode == "" {
		mode = ModeCasual
	}

	size := h.service.GetQueueSize(r.Context(), gameType, mode)
	httputil.JSON(w, r, http.StatusOK, map[string]int{
		"size": size,
	})
}
