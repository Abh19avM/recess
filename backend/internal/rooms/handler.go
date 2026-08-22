package rooms

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Abh19avM/recess/internal/httputil"
	"github.com/go-chi/chi/v5"
)

// Handler handles HTTP requests for room management.
type Handler struct {
	service Service
}

// NewHandler creates a new room HTTP handler.
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// Routes returns the subrouter for /api/v1/rooms.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.ListRooms)
	r.Post("/", h.CreateRoom)
	r.Get("/{code}", h.GetRoom)
	r.Post("/{code}/join", h.JoinRoom)

	return r
}

func (h *Handler) ListRooms(w http.ResponseWriter, r *http.Request) {
	rooms, err := h.service.ListActiveRooms(r.Context())
	if err != nil {
		httputil.ErrorJSON(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list active rooms")
		return
	}

	httputil.JSON(w, r, http.StatusOK, map[string]any{
		"rooms": rooms,
		"count": len(rooms),
	})
}

func (h *Handler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	var req CreateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.ErrorJSON(w, r, http.StatusBadRequest, "INVALID_BODY", "failed to parse request body")
		return
	}

	if req.GameType == "" {
		httputil.ValidationErrorJSON(w, r, "game_type is required", map[string]string{
			"game_type": "required",
		})
		return
	}

	// Scaffolding host identity (in Phase 5 this comes from JWT context)
	hostID := "usr_host_player"
	hostUsername := "RoomHost"
	hostAvatar := "pencil_sketch_1"
	hostRating := 1200

	room, err := h.service.CreateRoom(r.Context(), hostID, hostUsername, hostAvatar, hostRating, req)
	if err != nil {
		httputil.ErrorJSON(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create room")
		return
	}

	httputil.JSON(w, r, http.StatusCreated, room)
}

func (h *Handler) GetRoom(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" {
		httputil.ErrorJSON(w, r, http.StatusBadRequest, "BAD_REQUEST", "room code is required")
		return
	}

	room, err := h.service.GetRoomByCode(r.Context(), code)
	if err != nil {
		if errors.Is(err, ErrRoomNotFound) {
			httputil.ErrorJSON(w, r, http.StatusNotFound, "NOT_FOUND", "room not found")
			return
		}
		httputil.ErrorJSON(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to retrieve room")
		return
	}

	httputil.JSON(w, r, http.StatusOK, room)
}

func (h *Handler) JoinRoom(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" {
		httputil.ErrorJSON(w, r, http.StatusBadRequest, "BAD_REQUEST", "room code is required")
		return
	}

	var req JoinRoomRequest
	_ = json.NewDecoder(r.Body).Decode(&req) // Optional body for passcode

	userID := "usr_joining_player"
	username := "Challenger"
	avatar := "pencil_sketch_2"
	rating := 1250

	room, err := h.service.JoinRoom(r.Context(), code, userID, username, avatar, rating, req.Passcode)
	if err != nil {
		if errors.Is(err, ErrRoomNotFound) {
			httputil.ErrorJSON(w, r, http.StatusNotFound, "NOT_FOUND", "room not found")
			return
		}
		if errors.Is(err, ErrRoomFull) {
			httputil.ErrorJSON(w, r, http.StatusConflict, "ROOM_FULL", "room is already full")
			return
		}
		if errors.Is(err, ErrInvalidPasscode) {
			httputil.ErrorJSON(w, r, http.StatusForbidden, "INVALID_PASSCODE", "invalid passcode for private room")
			return
		}
		httputil.ErrorJSON(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to join room")
		return
	}

	httputil.JSON(w, r, http.StatusOK, room)
}
