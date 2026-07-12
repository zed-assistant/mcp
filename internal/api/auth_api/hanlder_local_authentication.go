package authapi

import (
	"net/http"

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
	authenticatedUser, err := a.localIDP.Authenticate(user, pass)
	if err != nil {
		a.log.WarnContext(ctx, "Authentication failed", logger.LogError(err))
		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Local authentication successful " + authenticatedUser.Email))
}
