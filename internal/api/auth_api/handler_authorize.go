package authapi

import (
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

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello Authorize!"))
}
