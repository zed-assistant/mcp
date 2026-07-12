package oauth

import (
	"context"
	"errors"
	"strings"

	"github.com/ory/fosite"
	"github.com/ory/fosite/storage"
	"github.com/zed-assistant/mcp/internal/configuration"
)

type ClientIdMetadataDocumentResolver interface {
	Resolve(ctx context.Context, id string, allowedScopes []string, audience string) (fosite.Client, error)
}

type MemoryStore struct {
	*storage.MemoryStore

	cimdResolver ClientIdMetadataDocumentResolver
	appConfig    *configuration.AppConfig
}

func NewMemoryStore(cimdResolver ClientIdMetadataDocumentResolver, appConfig *configuration.AppConfig) *MemoryStore {
	return &MemoryStore{
		MemoryStore:  storage.NewMemoryStore(),
		cimdResolver: cimdResolver,
		appConfig:    appConfig,
	}
}

func (s *MemoryStore) GetClient(ctx context.Context, id string) (fosite.Client, error) {
	if c := clientOverrideFrom(ctx); c != nil && c.GetID() == id {
		return c, nil
	}
	if !strings.HasPrefix(id, "https://") {
		return nil, errors.New("invalid client ID: must start with 'https://' - only CIMD clients are supported")
	}
	return s.cimdResolver.Resolve(ctx, id, []string{"mcp:tools", "offline_access", "openid", "email", "profile"}, s.appConfig.Server.ExternalUrl+"/mcp")
}
