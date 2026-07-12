package authorization

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/ory/fosite"
)

type IntrospectionResult struct {
	Scopes     []string
	Expiration time.Time
	Email      string
	Sub        string
}

func (a *AuthorizationManager) IntrospectAccessToken(ctx context.Context, accessToken string) (*IntrospectionResult, error) {
	_, ar, err := a.oauthProvider.IntrospectToken(ctx, accessToken, fosite.AccessToken, &fosite.DefaultSession{})
	if err != nil {
		return nil, fmt.Errorf("failed to introspect access token: %w", err)
	}

	aud := []string(ar.GetGrantedAudience())
	if len(aud) == 0 {
		aud = []string(ar.GetRequestedAudience())
	}
	if !slices.Contains(aud, a.appConfig.Server.ExternalUrl+"/mcp") {
		return nil, fmt.Errorf("access token audience does not include the required audience: %s", a.appConfig.Server.ExternalUrl+"/mcp")
	}

	sess, _ := ar.GetSession().(*fosite.DefaultSession)
	if sess == nil || sess.Subject == "" || sess.Username == "" {
		return nil, fmt.Errorf("access token session is invalid or missing subject or username")
	}

	return &IntrospectionResult{
		Scopes:     []string(ar.GetGrantedScopes()),
		Expiration: sess.GetExpiresAt(fosite.AccessToken),
		Email:      sess.Username,
		Sub:        sess.Subject,
	}, nil
}
