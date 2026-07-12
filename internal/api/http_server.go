package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	authapi "github.com/zed-assistant/mcp/internal/api/auth_api"
	httpmiddleware "github.com/zed-assistant/mcp/internal/api/http_middleware"
	mcpapi "github.com/zed-assistant/mcp/internal/api/mcp_api"
	wellknownapi "github.com/zed-assistant/mcp/internal/api/well_known_api"
	"github.com/zed-assistant/mcp/internal/configuration"
)

type HttpServer struct {
	appConfig  *configuration.AppConfig
	httpServer *http.Server
	logger     *slog.Logger
}

func NewHttpServer(appConfig *configuration.AppConfig, auth *authapi.AuthApi, wellKnown *wellknownapi.WellKnownApi, mcp *mcpapi.McpApi, logger *slog.Logger) (*HttpServer, error) {
	router := chi.NewRouter()

	router.Use(httpmiddleware.RequestIdMiddleware(logger))
	router.Use(httpmiddleware.RecoveryMiddleware(logger))

	router.Mount("/auth", auth.GetRouter())
	router.Mount("/.well-known", wellKnown.GetRouter())
	router.Mount("/mcp", mcp.GetRouter())

	httpServer := &http.Server{
		ReadHeaderTimeout: 10 * time.Second,
		Addr:              fmt.Sprintf(":%d", appConfig.Server.Port),
		Handler:           router,
	}

	return &HttpServer{
		appConfig:  appConfig,
		httpServer: httpServer,
		logger:     logger,
	}, nil
}

func (s *HttpServer) Start(ctx context.Context) error {
	s.logger.InfoContext(ctx, fmt.Sprintf("Starting HTTP server on port %d", s.appConfig.Server.Port))
	return s.httpServer.ListenAndServe()
}

func (s *HttpServer) Shutdown(ctx context.Context) error {
	s.logger.InfoContext(ctx, "Shutting down HTTP server")
	return s.httpServer.Shutdown(ctx)
}
