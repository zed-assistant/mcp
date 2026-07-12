package authapi

import (
	"net/http"

	"github.com/ory/fosite"
	"github.com/zed-assistant/mcp/internal/logger"
)

func (a *AuthApi) getToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := &fosite.DefaultSession{}

	accessReq, err := a.oauthProvider.NewAccessRequest(ctx, r, session)
	if err != nil {
		logError := oauthErrorToLoggerError(err)
		a.log.Error("Failed to create access request", logger.LogError(logError))
		a.oauthProvider.WriteAccessError(ctx, w, accessReq, err)
		return
	}

	resp, err := a.oauthProvider.NewAccessResponse(ctx, accessReq)
	if err != nil {
		logError := oauthErrorToLoggerError(err)
		a.log.Error("Failed to create access response", logger.LogError(logError))
		a.oauthProvider.WriteAccessError(ctx, w, accessReq, err)
		return
	}

	a.oauthProvider.WriteAccessResponse(ctx, w, accessReq, resp)
}
