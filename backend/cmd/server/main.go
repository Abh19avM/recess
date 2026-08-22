package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Abh19avM/recess/internal/config"
	"github.com/Abh19avM/recess/internal/database"
	"github.com/Abh19avM/recess/internal/middleware"
	"github.com/Abh19avM/recess/internal/redis"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

var startTime = time.Now()

// HealthResponse represents the response payload for /health (Liveness)
type HealthResponse struct {
	Status        string  `json:"status"`
	Environment   string  `json:"environment"`
	UptimeSeconds float64 `json:"uptime_seconds"`
	Timestamp     string  `json:"timestamp"`
}

// ServiceCheck represents the status of an external dependency.
type ServiceCheck struct {
	Status    string  `json:"status"` // "up" or "down"
	LatencyMS float64 `json:"latency_ms,omitempty"`
	Error     string  `json:"error,omitempty"`
}

// ReadyResponse represents the response payload for /ready (Readiness)
type ReadyResponse struct {
	Status        string                  `json:"status"` // "ready" or "unavailable"
	Environment   string                  `json:"environment"`
	UptimeSeconds float64                 `json:"uptime_seconds"`
	Timestamp     string                  `json:"timestamp"`
	Checks        map[string]ServiceCheck `json:"checks"`
}

// Server holds dependencies for HTTP handlers.
type Server struct {
	cfg   *config.Config
	db    *database.Database
	redis *redis.Client
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	resp := HealthResponse{
		Status:        "ok",
		Environment:   s.cfg.Environment,
		UptimeSeconds: time.Since(startTime).Seconds(),
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) readyHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	checks := make(map[string]ServiceCheck)
	isReady := true

	// Check Database
	if s.db == nil {
		checks["database"] = ServiceCheck{
			Status: "down",
			Error:  "database client not initialized",
		}
		isReady = false
	} else {
		latency, err := s.db.Ping(ctx)
		if err != nil {
			checks["database"] = ServiceCheck{
				Status: "down",
				Error:  err.Error(),
			}
			isReady = false
		} else {
			checks["database"] = ServiceCheck{
				Status:    "up",
				LatencyMS: float64(latency.Microseconds()) / 1000.0,
			}
		}
	}

	// Check Redis
	if s.redis == nil {
		checks["redis"] = ServiceCheck{
			Status: "down",
			Error:  "redis client not initialized",
		}
		isReady = false
	} else {
		latency, err := s.redis.Ping(ctx)
		if err != nil {
			checks["redis"] = ServiceCheck{
				Status: "down",
				Error:  err.Error(),
			}
			isReady = false
		} else {
			checks["redis"] = ServiceCheck{
				Status:    "up",
				LatencyMS: float64(latency.Microseconds()) / 1000.0,
			}
		}
	}

	status := "ready"
	httpStatus := http.StatusOK
	if !isReady {
		status = "unavailable"
		httpStatus = http.StatusServiceUnavailable
	}

	resp := ReadyResponse{
		Status:        status,
		Environment:   s.cfg.Environment,
		UptimeSeconds: time.Since(startTime).Seconds(),
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		Checks:        checks,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	_ = json.NewEncoder(w).Encode(resp)
}

// SetupRouter creates and configures the Chi router with middlewares and routes.
func SetupRouter(s *Server) *chi.Mux {
	r := chi.NewRouter()

	// Base middlewares
	r.Use(middleware.RequestLogger())
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   s.cfg.CORSAllowed,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Request-ID"},
		ExposedHeaders:   []string{"Link", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Liveness & Readiness endpoints
	r.Get("/health", s.healthHandler)
	r.Get("/ready", s.readyHandler)

	// API Root info
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"name":"Recess API","status":"running","health":"/health","ready":"/ready"}`+"\n")
	})

	return r
}

func main() {
	cfg := config.Load()
	cfg.SetupLogger()

	slog.Info("starting Recess backend server",
		"environment", cfg.Environment,
		"port", cfg.Port,
		"log_level", cfg.LogLevel,
	)

	// Create root background context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize PostgreSQL connection pool
	db, err := database.New(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Warn("failed to initialize PostgreSQL pool on startup (service will start in degraded mode)", "error", err)
	}

	// Initialize Redis client
	rdb, err := redis.New(ctx, cfg.RedisURL)
	if err != nil {
		slog.Warn("failed to initialize Redis client on startup (service will start in degraded mode)", "error", err)
	}

	serverState := &Server{
		cfg:   cfg,
		db:    db,
		redis: rdb,
	}

	router := SetupRouter(serverState)

	httpServer := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	// Start server in background goroutine
	serverErr := make(chan error, 1)
	go func() {
		slog.Info(fmt.Sprintf("Recess server listening on http://0.0.0.0:%s", cfg.Port))
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	// Listen for shutdown signals
	stopSig := make(chan os.Signal, 1)
	signal.Notify(stopSig, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	select {
	case err := <-serverErr:
		slog.Error("server listener error", "error", err)
	case sig := <-stopSig:
		slog.Info("shutdown signal received", "signal", sig.String())
	}

	// Begin graceful shutdown
	slog.Info("initiating graceful shutdown", "timeout", cfg.ShutdownTimeout.String())

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer shutdownCancel()

	// Stop accepting new HTTP requests
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("error during HTTP server shutdown", "error", err)
	} else {
		slog.Info("HTTP server stopped gracefully")
	}

	// Close database pool
	if db != nil {
		db.Close()
	}

	// Close Redis client
	if rdb != nil {
		if err := rdb.Close(); err != nil {
			slog.Error("error closing Redis client", "error", err)
		}
	}

	slog.Info("Recess backend server shutdown complete")
}
