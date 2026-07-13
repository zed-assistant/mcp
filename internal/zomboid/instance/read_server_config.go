package instance

import (
	"context"

	"github.com/zed-assistant/mcp/internal/auth/authorization"
)

type ReadServerConfigInput struct {
	InstanceID string `json:"instanceId" jsonschema:"The ID of the Project Zomboid server instance to read the config for" validate:"required"`
}

func (m *ZomboidInstanceManager) ReadServerConfig(ctx context.Context, principal authorization.Principal, input ReadServerConfigInput) (map[string]any, error) {
	if err := m.instanceAuth.AuthorizeInstanceAccess(input.InstanceID, principal); err != nil {
		return nil, err
	}

	return map[string]any{
		"foo": "bar",
		"baz": 42,
		"qux": true,
	}, nil
}
