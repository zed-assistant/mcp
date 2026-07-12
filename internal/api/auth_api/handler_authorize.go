package authapi

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/zed-assistant/mcp/internal/logger"
)

func (a *AuthApi) authorize(w http.ResponseWriter, r *http.Request) {
	ctx := a.oauthStore.WithLoopbackRedirect(r.Context(), r.URL.Query())

	authorizeRequest, err := a.oauthProvider.NewAuthorizeRequest(ctx, r)
	if err != nil {
		logError := oauthErrorToLoggerError(err)
		a.log.Warn("Unable to parse authorize request", logger.LogError(logError))
		a.oauthProvider.WriteAuthorizeError(ctx, w, authorizeRequest, err)
		return
	}

	pending, err := a.pendingAuthStore.StorePendingAuth(r.URL.Query())
	if err != nil {
		a.log.Error("Unable to store pending auth request", logger.LogError(err))
		a.oauthProvider.WriteAuthorizeError(ctx, w, authorizeRequest, err)
		return
	}

	redirectURL, err := a.getAuthorizationURL(pending.ID, pending.Nonce)
	if err != nil {
		a.log.Error("Unable to get authorization URL", logger.LogError(err))
		a.oauthProvider.WriteAuthorizeError(ctx, w, authorizeRequest, err)
		return
	}

	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func (a *AuthApi) getAuthorizationURL(state string, nonce string) (string, error) {
	switch a.appConfig.OAuth2.IDP.Type {
	case "local":
		return a.localIDP.GetAuthorizationURL(state, nonce)
	default:
		return "", errors.New(fmt.Sprintf("IDP type %s is not supported", a.appConfig.OAuth2.IDP.Type))
	}
}
