package main

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"time"

	authapi "github.com/zed-assistant/mcp/internal/api/auth_api"
	localidp "github.com/zed-assistant/mcp/internal/auth/idp/local"
	"github.com/zed-assistant/mcp/internal/auth/oauth"
	"github.com/zed-assistant/mcp/internal/configuration"
	"github.com/zed-assistant/mcp/internal/random"
	"github.com/zed-assistant/mcp/internal/request"
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
	localIDP := localidp.NewLocalIDP(appConfig)
	if err != nil {
		return nil, err
	}
	auth := authapi.NewAuthApi(appConfig, oauthProvider, oauthStore, pendingAuthStore, log, localIDP)

	return &serverDeps{
		authApi: auth,
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
