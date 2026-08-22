package httputil

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Abh19avM/recess/internal/middleware"
)

// ResponseEnvelope represents the standard API response structure.
type ResponseEnvelope struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Meta    *Meta  `json:"meta,omitempty"`
	Error   *Error `json:"error,omitempty"`
}

// Meta contains metadata for the response.
type Meta struct {
	RequestID string `json:"request_id,omitempty"`
	Timestamp string `json:"timestamp"`
}

// Error represents standard error details in API responses.
type Error struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
	Timestamp string `json:"timestamp"`
	Details   any    `json:"details,omitempty"`
}

// JSON sends a successful JSON response with metadata.
func JSON(w http.ResponseWriter, r *http.Request, status int, data any) {
	reqID := middleware.GetRequestID(r.Context())
	resp := ResponseEnvelope{
		Success: true,
		Data:    data,
		Meta: &Meta{
			RequestID: reqID,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(resp)
}

// ErrorJSON sends a standardized JSON error response.
func ErrorJSON(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	reqID := middleware.GetRequestID(r.Context())
	resp := ResponseEnvelope{
		Success: false,
		Error: &Error{
			Code:      code,
			Message:   message,
			RequestID: reqID,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(resp)
}

// ValidationErrorJSON sends a 422 Unprocessable Entity error response with field details.
func ValidationErrorJSON(w http.ResponseWriter, r *http.Request, message string, details any) {
	reqID := middleware.GetRequestID(r.Context())
	resp := ResponseEnvelope{
		Success: false,
		Error: &Error{
			Code:      "VALIDATION_ERROR",
			Message:   message,
			RequestID: reqID,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Details:   details,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnprocessableEntity)
	_ = json.NewEncoder(w).Encode(resp)
}
