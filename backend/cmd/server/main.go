package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Abh19avM/recess/internal/app"
	"github.com/Abh19avM/recess/internal/config"
)

func main() {
	// 1. Load application configuration
	cfg := config.Load()
	cfg.SetupLogger()

	slog.Info("initializing Recess application",
		"environment", cfg.Environment,
		"port", cfg.Port,
		"log_level", cfg.LogLevel,
	)

	// 2. Initialize application with root context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	application, err := app.New(ctx, cfg)
	if err != nil {
		slog.Error("failed to bootstrap application", "error", err)
		os.Exit(1)
	}

	// 3. Start HTTP server
	serverErr := make(chan error, 1)
	go application.Start(serverErr)

	// 4. Await OS termination signals
	stopSig := make(chan os.Signal, 1)
	signal.Notify(stopSig, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	select {
	case err := <-serverErr:
		slog.Error("fatal HTTP server error", "error", err)
	case sig := <-stopSig:
		slog.Info("shutdown signal caught", "signal", sig.String())
	}

	// 5. Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer shutdownCancel()

	if err := application.Shutdown(shutdownCtx); err != nil {
		slog.Error("error encountered during shutdown", "error", err)
	}

	slog.Info("Recess server stopped successfully")
}
