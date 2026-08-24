package middleware

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"
)

type contextKey string

const (
	// RequestIDKey is the context key for the unique request ID.
	RequestIDKey contextKey = "request_id"
	// RequestIDHeader is the HTTP header for propagating the request ID.
	RequestIDHeader = "X-Request-ID"
)

// responseWriter is a wrapper around http.ResponseWriter that captures status code and bytes written.
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int64
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += int64(n)
	return n, err
}

func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := rw.ResponseWriter.(http.Hijacker); ok {
		return hj.Hijack()
	}
	return nil, nil, fmt.Errorf("underlying ResponseWriter does not implement http.Hijacker")
}

func (rw *responseWriter) Flush() {
	if fl, ok := rw.ResponseWriter.(http.Flusher); ok {
		fl.Flush()
	}
}

func (rw *responseWriter) Unwrap() http.ResponseWriter {
	return rw.ResponseWriter
}

// RequestLogger returns a middleware that logs HTTP requests with structured slog fields and assigns a unique Request ID.
func RequestLogger() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			reqID := r.Header.Get(RequestIDHeader)
			if reqID == "" {
				b := make([]byte, 8)
				_, _ = rand.Read(b)
				reqID = hex.EncodeToString(b)
			}

			w.Header().Set(RequestIDHeader, reqID)
			ctx := context.WithValue(r.Context(), RequestIDKey, reqID)

			rw := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			defer func() {
				if rec := recover(); rec != nil {
					rw.statusCode = http.StatusInternalServerError
					slog.Error("panic recovered in HTTP handler",
						"request_id", reqID,
						"method", r.Method,
						"path", r.URL.Path,
						"panic", rec,
					)
					http.Error(rw, `{"error":"internal server error"}`, http.StatusInternalServerError)
				}

				duration := time.Since(start)
				// Don't spam debug logs for frequent /health checks if desired, but keep structured log
				logLevel := slog.LevelInfo
				if rw.statusCode >= 500 {
					logLevel = slog.LevelError
				} else if rw.statusCode >= 400 {
					logLevel = slog.LevelWarn
				}

				slog.Log(ctx, logLevel, "http request",
					"request_id", reqID,
					"method", r.Method,
					"path", r.URL.Path,
					"status", rw.statusCode,
					"duration_ms", float64(duration.Microseconds())/1000.0,
					"bytes", rw.bytesWritten,
					"ip", r.RemoteAddr,
					"user_agent", r.UserAgent(),
				)
			}()

			next.ServeHTTP(rw, r.WithContext(ctx))
		})
	}
}

// GetRequestID extracts the request ID from the context if available.
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	return ""
}
