package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/Abh19avM/recess/internal/auth"
	"github.com/Abh19avM/recess/internal/config"
	"github.com/Abh19avM/recess/internal/database"
	"github.com/Abh19avM/recess/internal/games"
	"github.com/Abh19avM/recess/internal/leaderboard"
	"github.com/Abh19avM/recess/internal/matchmaking"
	"github.com/Abh19avM/recess/internal/metrics"
	"github.com/Abh19avM/recess/internal/middleware"
	"github.com/Abh19avM/recess/internal/redis"
	"github.com/Abh19avM/recess/internal/rooms"
	"github.com/Abh19avM/recess/internal/users"
	"github.com/Abh19avM/recess/internal/websocket"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	goredis "github.com/redis/go-redis/v9"
)

var startTime = time.Now()

// App orchestrates application dependencies and the HTTP server lifecycle.
type App struct {
	cfg        *config.Config
	db         *database.Database
	redis      *redis.Client
	httpServer *http.Server
	wsHub      *websocket.Hub
}

// HealthResponse represents the response payload for /health (Liveness)
type HealthResponse struct {
	Status        string  `json:"status"`
	Environment   string  `json:"environment"`
	UptimeSeconds float64 `json:"uptime_seconds"`
	Timestamp     string  `json:"timestamp"`
}

// ServiceCheck represents the status of an external dependency.
type ServiceCheck struct {
	Status    string  `json:"status"`
	LatencyMS float64 `json:"latency_ms,omitempty"`
	Error     string  `json:"error,omitempty"`
}

// ReadyResponse represents the response payload for /ready (Readiness)
type ReadyResponse struct {
	Status        string                  `json:"status"`
	Environment   string                  `json:"environment"`
	UptimeSeconds float64                 `json:"uptime_seconds"`
	Timestamp     string                  `json:"timestamp"`
	Checks        map[string]ServiceCheck `json:"checks"`
}

// New initializes and wires all application layers and dependencies.
func New(ctx context.Context, cfg *config.Config) (*App, error) {
	// Initialize OpenTelemetry Tracing Provider
	_, _ = metrics.InitTracerProvider("recess-backend", cfg.Environment)

	// Set Default Redacting Logger to protect credentials & tokens
	cfg.SetupLogger()

	// 1. Initialize PostgreSQL Connection Pool
	db, err := database.New(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Warn("failed to connect to PostgreSQL on startup (starting in degraded/in-memory mode)", "error", err)
	}

	// 2. Initialize Redis Client
	rdb, err := redis.New(ctx, cfg.RedisURL)
	if err != nil {
		slog.Warn("failed to connect to Redis on startup (starting in degraded mode)", "error", err)
	}

	// 3. Initialize WebSocket Hub
	var wsHub *websocket.Hub
	if rdb != nil && rdb.RDB != nil {
		wsHub = websocket.NewHub(websocket.WithRedis(rdb.RDB))
	} else {
		wsHub = websocket.NewHub()
	}

	// 4. Initialize Repositories (with in-memory fallback for local dev / tests)
	userRepo := users.NewInMemoryRepository()
	roomRepo := rooms.NewInMemoryRepository()

	// 5. Initialize Auth Token Management & Session Store
	jwtManager := auth.NewJWTManager(cfg.JWTSecret)
	var tokenStore auth.TokenStore
	if rdb != nil && rdb.RDB != nil {
		pingCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
		if _, err := rdb.Ping(pingCtx); err == nil {
			tokenStore = auth.NewRedisTokenStore(rdb.RDB)
		}
		cancel()
	}
	if tokenStore == nil {
		tokenStore = auth.NewInMemoryTokenStore()
	}

	authBarrier := middleware.RequireAuth(jwtManager)
	authRateLimiter := middleware.NewIPRateLimiter(30, 1*time.Minute)

	// 6. Initialize Services
	userService := users.NewService(userRepo)
	authService := auth.NewService(userRepo, jwtManager, tokenStore)
	roomService := rooms.NewService(roomRepo)
	gameService := games.NewService()

	var rdbClient *goredis.Client
	if rdb != nil {
		rdbClient = rdb.RDB
	}
	matchmakingService := matchmaking.NewService(rdbClient)

	var dbPool *pgxpool.Pool
	if db != nil {
		dbPool = db.Pool
	}
	leaderboardService := leaderboard.NewService(dbPool, rdbClient)

	// 7. Initialize Handlers
	userHandler := users.NewHandler(userService, authBarrier)
	authHandler := auth.NewHandler(authService)
	roomHandler := rooms.NewHandler(roomService)
	gameHandler := games.NewHandler(gameService)
	matchmakingHandler := matchmaking.NewHandler(matchmakingService, authBarrier)
	leaderboardHandler := leaderboard.NewHandler(leaderboardService)
	wsHandler := websocket.NewHandler(wsHub, jwtManager)

	// 8. Configure Router
	r := chi.NewRouter()

	// Global Middlewares
	r.Use(metrics.HTTPMetricsMiddleware)
	r.Use(middleware.RequestLogger())
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)

	// CORS Configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORSAllowed,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Request-ID"},
		ExposedHeaders:   []string{"Link", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	app := &App{
		cfg:   cfg,
		db:    db,
		redis: rdb,
		wsHub: wsHub,
	}

	// Root Level Routes
	r.Get("/health", app.handleHealth)
	r.Get("/ready", app.handleReady)
	r.Handle("/metrics", promhttp.HandlerFor(metrics.Registry(), promhttp.HandlerOpts{}))
	r.Get("/ws", wsHandler.ServeWS)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"name":"Recess API","version":"v1","status":"running","health":"/health","ready":"/ready"}`+"\n")
	})

	// API v1 Subrouter
	r.Route("/api/v1", func(api chi.Router) {
		api.With(authRateLimiter.Handler()).Mount("/auth", authHandler.Routes())
		api.Mount("/users", userHandler.Routes())
		api.Mount("/rooms", roomHandler.Routes())
		api.Mount("/games", gameHandler.Routes())
		api.Mount("/matchmaking", matchmakingHandler.Routes())
		api.Mount("/leaderboards", leaderboardHandler.Routes())
		api.Mount("/progression", leaderboardHandler.UserRoutes())
	})

	app.httpServer = &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	return app, nil
}

// Router exposes the Chi router (useful for HTTP testing).
func (a *App) Router() http.Handler {
	return a.httpServer.Handler
}

// Start launches the HTTP server in a background listener.
func (a *App) Start(errChan chan<- error) {
	slog.Info(fmt.Sprintf("Recess server listening on http://0.0.0.0:%s", a.cfg.Port),
		"port", a.cfg.Port,
		"env", a.cfg.Environment,
	)

	if err := a.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		errChan <- err
	}
}

// Shutdown gracefully terminates the HTTP server and database/redis connections.
func (a *App) Shutdown(ctx context.Context) error {
	slog.Info("initiating graceful shutdown", "timeout", a.cfg.ShutdownTimeout.String())

	var errs []error

	if err := a.httpServer.Shutdown(ctx); err != nil {
		errs = append(errs, fmt.Errorf("http server shutdown: %w", err))
	} else {
		slog.Info("HTTP server stopped gracefully")
	}

	if a.db != nil {
		a.db.Close()
	}

	if a.redis != nil {
		if err := a.redis.Close(); err != nil {
			errs = append(errs, fmt.Errorf("redis close: %w", err))
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func (a *App) handleHealth(w http.ResponseWriter, r *http.Request) {
	resp := HealthResponse{
		Status:        "ok",
		Environment:   a.cfg.Environment,
		UptimeSeconds: time.Since(startTime).Seconds(),
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func (a *App) handleReady(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	defer cancel()

	checks := make(map[string]ServiceCheck)
	isReady := true

	// Check Database
	if a.db == nil {
		checks["database"] = ServiceCheck{
			Status: "down",
			Error:  "database client not initialized",
		}
		isReady = false
	} else {
		latency, err := a.db.Ping(ctx)
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
	if a.redis == nil {
		checks["redis"] = ServiceCheck{
			Status: "down",
			Error:  "redis client not initialized",
		}
		isReady = false
	} else {
		latency, err := a.redis.Ping(ctx)
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
		Environment:   a.cfg.Environment,
		UptimeSeconds: time.Since(startTime).Seconds(),
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		Checks:        checks,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	_ = json.NewEncoder(w).Encode(resp)
}
