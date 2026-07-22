package main

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"time"

	authapi "github.com/zed-assistant/mcp/internal/api/auth_api"
	mcpapi "github.com/zed-assistant/mcp/internal/api/mcp_api"
	wellknownapi "github.com/zed-assistant/mcp/internal/api/well_known_api"
	"github.com/zed-assistant/mcp/internal/auth/authorization"
	instanceauth "github.com/zed-assistant/mcp/internal/auth/authorization/instance_auth"
	"github.com/zed-assistant/mcp/internal/auth/idp"
	localidp "github.com/zed-assistant/mcp/internal/auth/idp/local"
	"github.com/zed-assistant/mcp/internal/auth/oauth"
	"github.com/zed-assistant/mcp/internal/configuration"
	"github.com/zed-assistant/mcp/internal/jwt"
	mcptool "github.com/zed-assistant/mcp/internal/mcp_tool"
	"github.com/zed-assistant/mcp/internal/random"
	"github.com/zed-assistant/mcp/internal/request"
	admincommand "github.com/zed-assistant/mcp/internal/zomboid/admin_command"
	"github.com/zed-assistant/mcp/internal/zomboid/config"
	gamedatabase "github.com/zed-assistant/mcp/internal/zomboid/game_database"
	"github.com/zed-assistant/mcp/internal/zomboid/instance"
	"github.com/zed-assistant/mcp/internal/zomboid/whitelist"
)

func newServerDeps(appConfig *configuration.AppConfig, log *slog.Logger) (*serverDeps, error) {
	rand := random.NewRandom()

	ssrfSafeHttpTransport := newTransport(request.SSRFSafeDialContext)

	ssrfSafeHttpClient := &http.Client{
		Transport: ssrfSafeHttpTransport,
		Timeout:   15 * time.Second,
	}

	cimdResolver := oauth.NewCIMDResolver(ssrfSafeHttpClient, time.Now)
	oauthStore := oauth.NewMemoryStore(cimdResolver, appConfig)
	pendingAuthStore := oauth.NewPendingStore(appConfig, rand)
	oauthProvider, err := oauth.NewOAuth2Provider(appConfig, rand, oauthStore)
	if err != nil {
		return nil, err
	}
	idpManager, err := idp.NewIDPManger(appConfig, jwt.Sign, jwt.Verify, time.Now, rand)
	if err != nil {
		return nil, err
	}
	localIDP := localidp.NewLocalIDP(appConfig, idpManager)
	auth := authapi.NewAuthApi(appConfig, oauthProvider, oauthStore, pendingAuthStore, log, idpManager, localIDP, time.Now)

	wellKnown := wellknownapi.NewWellKnownApi(appConfig)

	instanceAuth := instanceauth.NewInstanceAuthorization(appConfig)
	authManger := authorization.NewAuthorizationManager(appConfig, oauthProvider)
	instanceLockManager := instance.NewInstanceLockManager()
	configManager := config.NewConfigManager()
	adminCommandManager := admincommand.NewAdminCommandRCON(configManager)
	gamedataManager := gamedatabase.NewGameDatabaseManager()
	whitelistManager := whitelist.NewWhitelistManager(gamedataManager)
	zomboidInstanceManager := instance.NewZomboidInstanceManager(appConfig, instanceAuth, instanceLockManager, configManager, log, adminCommandManager, whitelistManager)
	toolsManager := mcptool.NewMcpToolManager(log, zomboidInstanceManager)
	mcp := mcpapi.NewMcpApi(log, authManger, appConfig, toolsManager)

	return &serverDeps{
		authApi:      auth,
		wellKnownApi: wellKnown,
		mcpApi:       mcp,
	}, nil
}

func newTransport(dialContext func(ctx context.Context, network, addr string) (net.Conn, error)) *http.Transport {
	t := http.DefaultTransport.(*http.Transport).Clone()

	t.MaxIdleConns = 100
	t.MaxIdleConnsPerHost = 20
	t.IdleConnTimeout = 90 * time.Second

	if dialContext != nil {
		t.DialContext = dialContext
	}
	return t
}
