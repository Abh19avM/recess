package metrics

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel/attribute"
)

var idRegex = regexp.MustCompile(`/[0-9a-fA-F-]{8,}|/[0-9]+`)

// responseWriterInterceptor captures HTTP status code.
type responseWriterInterceptor struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriterInterceptor) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *responseWriterInterceptor) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hj.Hijack()
	}
	return nil, nil, fmt.Errorf("underlying ResponseWriter does not implement http.Hijacker")
}

func (w *responseWriterInterceptor) Flush() {
	if fl, ok := w.ResponseWriter.(http.Flusher); ok {
		fl.Flush()
	}
}

func (w *responseWriterInterceptor) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

// HTTPMetricsMiddleware tracks HTTP duration and OpenTelemetry spans for every API request.
func HTTPMetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriterInterceptor{ResponseWriter: w, statusCode: http.StatusOK}

		// Normalize path using Chi RouteContext or Regex
		path := r.URL.Path
		if routeCtx := chi.RouteContext(r.Context()); routeCtx != nil && routeCtx.RoutePattern() != "" {
			path = routeCtx.RoutePattern()
		} else {
			path = idRegex.ReplaceAllString(path, "/:id")
		}

		ctx, span := StartSpan(r.Context(), fmt.Sprintf("%s %s", r.Method, path))
		defer span.End()

		next.ServeHTTP(rw, r.WithContext(ctx))

		duration := time.Since(start).Seconds()
		statusStr := strconv.Itoa(rw.statusCode)

		HTTPRequestDuration.WithLabelValues(r.Method, path, statusStr).Observe(duration)

		span.SetAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.route", path),
			attribute.Int("http.status_code", rw.statusCode),
		)

		if rw.statusCode >= 500 {
			RecordError("http_handler", "5xx_server_error")
		} else if rw.statusCode >= 400 {
			RecordError("http_handler", "4xx_client_error")
		}
	})
}
