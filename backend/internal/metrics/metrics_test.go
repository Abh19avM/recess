package metrics

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func TestMetrics_GaugesAndCounters(t *testing.T) {
	ActivePlayers.Set(42)
	ActiveMatches.Set(12)
	WebSocketConnections.Set(84)
	MatchmakingQueueSize.WithLabelValues("hand_cricket").Set(3)
	MatchmakingLatency.WithLabelValues("hand_cricket").Observe(1.4)
	WebSocketMessagesTotal.WithLabelValues("game.move", "inbound").Inc()
	GameDuration.WithLabelValues("xo").Observe(45.2)
	RecordError("websocket", "client_timeout")
	ObserveRedis("ping", time.Now().Add(-10*time.Millisecond))
	ObserveDB("select", "users", time.Now().Add(-25*time.Millisecond))

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()

	handler := promhttp.HandlerFor(Registry(), promhttp.HandlerOpts{})
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 from /metrics, got %d", rec.Code)
	}

	body := rec.Body.String()

	expectedMetrics := []string{
		"recess_active_players 42",
		"recess_active_matches 12",
		"recess_websocket_connections 84",
		"recess_matchmaking_queue_size",
		"recess_matchmaking_latency_seconds",
		"recess_websocket_messages_total",
		"recess_game_duration_seconds",
		"recess_errors_total",
		"recess_redis_latency_seconds",
		"recess_database_latency_seconds",
	}

	for _, metricName := range expectedMetrics {
		if !strings.Contains(body, metricName) {
			t.Errorf("expected metrics endpoint to contain '%s'", metricName)
		}
	}
}

func TestMetrics_HTTPMiddleware(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/error" {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("internal error"))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	wrapped := HTTPMetricsMiddleware(testHandler)

	// 1. Successful request
	req1 := httptest.NewRequest(http.MethodGet, "/api/v1/games", nil)
	rec1 := httptest.NewRecorder()
	wrapped.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec1.Code)
	}

	// 2. Error request
	req2 := httptest.NewRequest(http.MethodGet, "/error", nil)
	rec2 := httptest.NewRecorder()
	wrapped.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec2.Code)
	}

	// Check metrics output
	recMetrics := httptest.NewRecorder()
	promhttp.HandlerFor(Registry(), promhttp.HandlerOpts{}).ServeHTTP(recMetrics, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	metricsBody := recMetrics.Body.String()
	if !strings.Contains(metricsBody, "recess_http_request_duration_seconds") {
		t.Errorf("expected http_request_duration_seconds metric in /metrics output")
	}
}

func TestMetrics_RedactingLogger(t *testing.T) {
	var buf bytes.Buffer
	baseHandler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	redactingHandler := NewRedactingHandler(baseHandler)
	logger := slog.New(redactingHandler)

	logger.Info("user login attempt",
		"user_id", "usr_12345",
		"username", "StudentLeader",
		"password", "UltraSecretPassword123!",
		"access_token", "jwt.token.payload.signature",
		"secret_code", "exam_cheat_sheet",
	)

	logOutput := buf.String()

	if strings.Contains(logOutput, "UltraSecretPassword123!") {
		t.Errorf("password was not redacted in logs: %s", logOutput)
	}
	if strings.Contains(logOutput, "jwt.token.payload.signature") {
		t.Errorf("access_token was not redacted in logs: %s", logOutput)
	}
	if strings.Contains(logOutput, "exam_cheat_sheet") {
		t.Errorf("secret_code was not redacted in logs: %s", logOutput)
	}

	if !strings.Contains(logOutput, "[REDACTED]") {
		t.Errorf("expected [REDACTED] in logs, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "StudentLeader") {
		t.Errorf("expected username to remain unredacted, got: %s", logOutput)
	}
}

func TestMetrics_OpenTelemetryTracer(t *testing.T) {
	tp, err := InitTracerProvider("recess-test", "test")
	if err != nil {
		t.Fatalf("failed to initialize tracer provider: %v", err)
	}
	defer func() { _ = tp.Shutdown(context.Background()) }()

	ctx, span := StartSpan(context.Background(), "test_span")
	if span == nil {
		t.Fatalf("expected non-nil span")
	}
	time.Sleep(2 * time.Millisecond)
	span.End()

	_ = ctx
}

func init() {
	// Silence standard logger output during tests
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
}
