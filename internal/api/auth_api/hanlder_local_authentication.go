package authapi

import (
	"net/http"
	"time"

	"github.com/zed-assistant/mcp/internal/auth/idp"
	"github.com/zed-assistant/mcp/internal/logger"
)

func (a *AuthApi) localAuthentication(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	state := r.URL.Query().Get("state")
	if state == "" {
		a.log.WarnContext(ctx, "Missing state query parameter")
		a.writeHTMLResponse(ctx, w, 200, "<b>Missing state query parameter</b>")
		return
	}
	nonce := r.URL.Query().Get("nonce")
	if nonce == "" {
		a.log.WarnContext(ctx, "Missing nonce query parameter")
		a.writeHTMLResponse(ctx, w, 200, "<b>Missing nonce query parameter</b>")
		return
	}

	user, pass, ok := r.BasicAuth()
	if !ok {
		a.log.WarnContext(ctx, "Missing or invalid Authorization header")
		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	signedAssertion, err := a.localIDP.Authenticate(user, pass, state)
	if err != nil {
		a.log.WarnContext(ctx, "Authentication failed", logger.LogError(err))
		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     idp.IDPCallbackCookieName,
		Value:    signedAssertion,
		Path:     "/",
		MaxAge:   int((5 * time.Minute).Seconds()),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, a.appConfig.Server.ExternalUrl+"/auth/callback", http.StatusFound)
}
