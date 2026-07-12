package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/zed-assistant/mcp/internal/api"
	"github.com/zed-assistant/mcp/internal/configuration"
	"github.com/zed-assistant/mcp/internal/logger"
)

func main() {
	args := getFlags()

	ctx := context.Background()

	appConfig, err := configuration.Load(args.configPath)
	if err != nil {
		log := logger.NewLogger(nil)
		log.ErrorContext(ctx, "Failed to load configuration", logger.LogError(err))
		os.Exit(1)
	}

	log := logger.NewLogger(appConfig)
	log.InfoContext(ctx, "Starting Zed Assistant MCP server")

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	deps, err := newServerDeps(appConfig, log)
	if err != nil {
		log.ErrorContext(ctx, "Failed to create server dependencies", logger.LogError(err))
		os.Exit(1) // nolint:gocritic
	}

	httpServer, err := api.NewHttpServer(appConfig, deps.authApi, log)
	if err != nil {
		log.ErrorContext(ctx, "Failed to create HTTP server", logger.LogError(err))
		os.Exit(1) // nolint:gocritic
	}

	go func() {
		err := httpServer.Start(ctx)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.ErrorContext(ctx, "Failed to start HTTP server", logger.LogError(err))
		} else if errors.Is(err, http.ErrServerClosed) {
			log.InfoContext(ctx, "HTTP server closed")
		}
	}()

	<-ctx.Done()

	log.InfoContext(ctx, "Received signal, shutting down...")
	if err := httpServer.Shutdown(ctx); err != nil {
		log.ErrorContext(ctx, "Failed to shut down HTTP server", logger.LogError(err))
	} else {
		log.InfoContext(ctx, "HTTP server shut down gracefully")
	}
}
