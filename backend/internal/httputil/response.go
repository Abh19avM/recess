package httputil

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type contextKey string

const (
	// RequestIDHeader is the HTTP header for propagating the unique request ID.
	RequestIDHeader = "X-Request-ID"
	// RequestIDKey is the context key for the unique request ID.
	RequestIDKey contextKey = "request_id"
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
	reqID := GetRequestID(r.Context())
	if reqID == "" {
		reqID = r.Header.Get(RequestIDHeader)
	}
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
	reqID := GetRequestID(r.Context())
	if reqID == "" {
		reqID = r.Header.Get(RequestIDHeader)
	}
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
	reqID := GetRequestID(r.Context())
	if reqID == "" {
		reqID = r.Header.Get(RequestIDHeader)
	}
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

// GetRequestID extracts the request ID from context if set.
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	return ""
}
