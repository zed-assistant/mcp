package authapi

import (
	"context"
	"net/http"
	"net/url"

	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/openid"
	"github.com/ory/fosite/token/jwt"
	"github.com/zed-assistant/mcp/internal/auth/idp"
	"github.com/zed-assistant/mcp/internal/auth/oauth"
	"github.com/zed-assistant/mcp/internal/logger"
)

func (a *AuthApi) authenticationCallback(w http.ResponseWriter, r *http.Request) {
	callbackCookie, err := r.Cookie(idp.IDPCallbackCookieName)
	ctx := r.Context()
	if err != nil {
		if err == http.ErrNoCookie {
			a.log.WarnContext(ctx, "Missing callback cookie")
			a.writeHTMLResponse(ctx, w, http.StatusOK, "<b>Authentication failed</b>")
			return
		}
		a.log.ErrorContext(ctx, "Error retrieving callback cookie", logger.LogError(err))
		a.writeHTMLResponse(ctx, w, http.StatusInternalServerError, "<b>Server Error</b>")
		return
	}

	authResult, err := a.idpManager.VerifyAssertion(callbackCookie.Value)
	if err != nil {
		a.log.ErrorContext(ctx, "Error verifying assertion", logger.LogError(err))
		a.writeHTMLResponse(ctx, w, http.StatusOK, "<b>Authentication failed</b>")
		return
	}
	pendingAuth, ok := a.pendingAuthStore.Get(authResult.PendingRequestID)
	if !ok || pendingAuth == nil {
		a.log.WarnContext(ctx, "Pending authentication request not found")
		a.writeHTMLResponse(ctx, w, http.StatusOK, "<b>Authentication failed</b>")
		return
	}
	defer a.pendingAuthStore.Delete(authResult.PendingRequestID)

	ctx, ar, err := a.rebuildAuthorizeRequest(ctx, pendingAuth)
	if err != nil {
		logError := oauthErrorToLoggerError(err)
		a.log.ErrorContext(ctx, "Error rebuilding authorize request", logger.LogError(logError))
		a.oauthProvider.WriteAuthorizeError(ctx, w, ar, err)
		return
	}

	for _, sc := range ar.GetRequestedScopes() {
		ar.GrantScope(sc)
	}
	for _, aud := range ar.GetRequestedAudience() {
		ar.GrantAudience(aud)
	}

	now := a.getCurrentTime()

	session := &openid.DefaultSession{
		Subject:  authResult.Sub,
		Username: authResult.Email,
		Claims: &jwt.IDTokenClaims{
			Subject:     authResult.Sub,
			Extra:       map[string]any{"email": authResult.Email},
			RequestedAt: now,
			IssuedAt:    now,
		},
		Headers: &jwt.Headers{},
	}

	resp, err := a.oauthProvider.NewAuthorizeResponse(ctx, ar, session)
	if err != nil {
		logError := oauthErrorToLoggerError(err)
		a.log.ErrorContext(ctx, "Error creating authorize response", logger.LogError(logError))
		a.oauthProvider.WriteAuthorizeError(ctx, w, ar, err)
		return
	}

	a.oauthProvider.WriteAuthorizeResponse(ctx, w, ar, resp)
}

func (a *AuthApi) rebuildAuthorizeRequest(ctx context.Context, pendingAuth *oauth.PendingAuth) (context.Context, fosite.AuthorizeRequester, error) {
	req := &http.Request{
		Method: http.MethodGet,
		URL:    &url.URL{Path: "/auth/authorize", RawQuery: pendingAuth.Query.Encode()},
	}
	ctx = a.oauthStore.WithLoopbackRedirect(ctx, pendingAuth.Query)
	ar, err := a.oauthProvider.NewAuthorizeRequest(ctx, req)
	return ctx, ar, err
}
