package mcpapi

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/zed-assistant/mcp/internal/auth/authorization"
	"github.com/zed-assistant/mcp/internal/configuration"
	"github.com/zed-assistant/mcp/internal/logger"
	mcptool "github.com/zed-assistant/mcp/internal/mcp_tool"
)

type AuthManager interface {
	IntrospectAccessToken(ctx context.Context, accessToken string) (*authorization.IntrospectionResult, error)
}

type ToolsManager interface {
	CollectTools() []mcptool.Tool
}

type McpApi struct {
	log          *slog.Logger
	authManager  AuthManager
	appConfig    *configuration.AppConfig
	toolsManager ToolsManager
}

func NewMcpApi(l *slog.Logger, authManager AuthManager, appConfig *configuration.AppConfig, toolsManager ToolsManager) *McpApi {
	return &McpApi{
		log:          l,
		authManager:  authManager,
		appConfig:    appConfig,
		toolsManager: toolsManager,
	}
}

func (a *McpApi) GetRouter() *chi.Mux {
	router := chi.NewRouter()

	server := mcp.NewServer(&mcp.Implementation{
		Name:  "zed-assistant-mcp",
		Title: "Zed Assistant MCP - AI tools for managing Project Zomboid server",
	}, &mcp.ServerOptions{
		Logger: a.log,
	})

	for _, tool := range a.toolsManager.CollectTools() {
		tool.Register(server)
	}

	stremable := mcp.NewStreamableHTTPHandler(
		func(r *http.Request) *mcp.Server {
			return server
		},
		nil,
	)

	requireAuth := auth.RequireBearerToken(a.VerifyToken, &auth.RequireBearerTokenOptions{
		ResourceMetadataURL: a.appConfig.Server.ExternalUrl + "/.well-known/oauth-protected-resource/mcp",
	})

	router.Use(requireAuth)

	router.Handle("/", stremable)

	return router
}

func (a *McpApi) VerifyToken(ctx context.Context, token string, _ *http.Request) (*auth.TokenInfo, error) {
	introspectionResult, err := a.authManager.IntrospectAccessToken(ctx, token)
	if err != nil {
		a.log.WarnContext(ctx, "Unable to introspect token", logger.LogError(err))
		return nil, auth.ErrInvalidToken
	}

	return &auth.TokenInfo{
		Scopes:     introspectionResult.Scopes,
		Expiration: introspectionResult.Expiration,
		UserID:     introspectionResult.Sub,
		Extra: map[string]any{
			"Email": introspectionResult.Email,
		},
	}, nil
}
