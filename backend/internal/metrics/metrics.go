package metrics

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	registry = prometheus.NewRegistry()
	once     sync.Once

	// Active Connections & Presence
	ActivePlayers = promauto.With(registry).NewGauge(
		prometheus.GaugeOpts{
			Namespace: "recess",
			Name:      "active_players",
			Help:      "Total number of active online players.",
		},
	)

	ActiveMatches = promauto.With(registry).NewGauge(
		prometheus.GaugeOpts{
			Namespace: "recess",
			Name:      "active_matches",
			Help:      "Total number of active in-progress game matches.",
		},
	)

	WebSocketConnections = promauto.With(registry).NewGauge(
		prometheus.GaugeOpts{
			Namespace: "recess",
			Name:      "websocket_connections",
			Help:      "Total active WebSocket connections.",
		},
	)

	// WebSocket Messages
	WebSocketMessagesTotal = promauto.With(registry).NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "recess",
			Name:      "websocket_messages_total",
			Help:      "Total WebSocket event envelopes received or sent.",
		},
		[]string{"type", "direction"},
	)

	// Matchmaking
	MatchmakingQueueSize = promauto.With(registry).NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "recess",
			Name:      "matchmaking_queue_size",
			Help:      "Current size of the matchmaking queue by game type.",
		},
		[]string{"game_type"},
	)

	MatchmakingLatency = promauto.With(registry).NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "recess",
			Name:      "matchmaking_latency_seconds",
			Help:      "Time taken from queue entry to match creation in seconds.",
			Buckets:   []float64{0.1, 0.25, 0.5, 1.0, 2.0, 5.0, 10.0, 30.0},
		},
		[]string{"game_type"},
	)

	// Redis Operations
	RedisLatency = promauto.With(registry).NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "recess",
			Name:      "redis_latency_seconds",
			Help:      "Latency of Redis cache and pub/sub operations in seconds.",
			Buckets:   []float64{0.0005, 0.001, 0.002, 0.005, 0.01, 0.025, 0.05, 0.1},
		},
		[]string{"operation"},
	)

	// Database Queries
	DatabaseLatency = promauto.With(registry).NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "recess",
			Name:      "database_latency_seconds",
			Help:      "Latency of PostgreSQL database queries in seconds.",
			Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0},
		},
		[]string{"query", "table"},
	)

	// HTTP Requests
	HTTPRequestDuration = promauto.With(registry).NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "recess",
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request duration in seconds.",
			Buckets:   []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0},
		},
		[]string{"method", "path", "status"},
	)

	// Error Counter
	ErrorsTotal = promauto.With(registry).NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "recess",
			Name:      "errors_total",
			Help:      "Total error occurrences by component and classification.",
		},
		[]string{"component", "type"},
	)

	// Game Match Duration
	GameDuration = promauto.With(registry).NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "recess",
			Name:      "game_duration_seconds",
			Help:      "Total duration of completed matches in seconds from start to finish.",
			Buckets:   []float64{10, 30, 60, 120, 180, 300, 600, 1200},
		},
		[]string{"game_type"},
	)
)

// Registry returns the custom Prometheus registry used for Recess metrics.
func Registry() *prometheus.Registry {
	return registry
}

// ObserveRedis measures execution duration of a Redis operation.
func ObserveRedis(op string, start time.Time) {
	duration := time.Since(start).Seconds()
	RedisLatency.WithLabelValues(op).Observe(duration)
}

// ObserveDB measures execution duration of a Database query.
func ObserveDB(query, table string, start time.Time) {
	duration := time.Since(start).Seconds()
	DatabaseLatency.WithLabelValues(query, table).Observe(duration)
}

// RecordError increments the error counter for a component and error type.
func RecordError(component, errorType string) {
	ErrorsTotal.WithLabelValues(component, errorType).Inc()
}
