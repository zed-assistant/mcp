package oauth

import (
	"context"
	"net/url"

	"github.com/ory/fosite"
)

type ctxKeyClientOverride struct{}

func withClientOverride(ctx context.Context, c fosite.Client) context.Context {
	return context.WithValue(ctx, ctxKeyClientOverride{}, c)
}

func clientOverrideFrom(ctx context.Context) fosite.Client {
	c, _ := ctx.Value(ctxKeyClientOverride{}).(fosite.Client)
	return c
}

func (s *MemoryStore) WithLoopbackRedirect(ctx context.Context, q url.Values) context.Context {
	clientID := q.Get("client_id")
	redirect := q.Get("redirect_uri")
	if clientID == "" || redirect == "" {
		return ctx
	}
	u, err := url.Parse(redirect)
	if err != nil || u.Scheme != "http" || !isLoopbackHost(u.Hostname()) {
		return ctx // only plain-http loopback gets the relaxed matching
	}
	base, err := s.GetClient(ctx, clientID)
	if err != nil {
		return ctx // let fosite produce the proper error
	}
	for _, reg := range base.GetRedirectURIs() {
		ru, err := url.Parse(reg)
		if err != nil {
			continue
		}
		if ru.Scheme == "http" && isLoopbackHost(ru.Hostname()) && ru.Path == u.Path {
			return withClientOverride(ctx, cloneWithRedirect(base, redirect))
		}
	}
	return ctx
}

func cloneWithRedirect(c fosite.Client, redirect string) fosite.Client {
	return &fosite.DefaultClient{
		ID:            c.GetID(),
		RedirectURIs:  append(append([]string{}, c.GetRedirectURIs()...), redirect),
		GrantTypes:    []string(c.GetGrantTypes()),
		ResponseTypes: []string(c.GetResponseTypes()),
		Scopes:        []string(c.GetScopes()),
		Audience:      []string(c.GetAudience()),
		Public:        c.IsPublic(),
	}
}
