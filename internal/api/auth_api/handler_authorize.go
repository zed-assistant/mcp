package authapi

import (
	"fmt"
	"net/http"

	"github.com/ory/fosite"
	"github.com/zed-assistant/mcp/internal/logger"
)

func (a *AuthApi) authorize(w http.ResponseWriter, r *http.Request) {
	injectResourceAsAudience(r)

	ctx := a.oauthStore.WithLoopbackRedirect(r.Context(), r.URL.Query())

	authorizeRequest, err := a.oauthProvider.NewAuthorizeRequest(ctx, r)
	if err != nil {
		logError := oauthErrorToLoggerError(err)
		a.log.Warn("Unable to parse authorize request", logger.LogError(logError))
		a.oauthProvider.WriteAuthorizeError(ctx, w, authorizeRequest, err)
		return
	}

	if r.URL.Query().Get("code_challenge") == "" {
		err := fosite.ErrInvalidRequest.WithHint("Clients must include a code_challenge when performing the authorize code flow, but it is missing.")
		a.log.Warn("Missing code_challenge in authorize request", logger.LogError(err))
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

// injectResourceAsAudience copies the RFC 8707 "resource" parameter(s) sent by MCP
// clients into fosite's "audience" parameter, since fosite only understands "audience"
// and has no native support for the "resource" indicator used by the MCP auth spec.
func injectResourceAsAudience(r *http.Request) {
	query := r.URL.Query()
	resources := query["resource"]
	if len(resources) == 0 {
		return
	}
	query["audience"] = append(query["audience"], resources...)
	r.URL.RawQuery = query.Encode()
}

func (a *AuthApi) getAuthorizationURL(state string, nonce string) (string, error) {
	switch a.appConfig.OAuth2.IDP.Type {
	case "local":
		return a.localIDP.GetAuthorizationURL(state, nonce)
	default:
		return "", fmt.Errorf("IDP type %s is not supported", a.appConfig.OAuth2.IDP.Type)
	}
}
