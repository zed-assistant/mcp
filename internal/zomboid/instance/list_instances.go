package instance

import (
	"context"
	"slices"

	"github.com/zed-assistant/mcp/internal/auth/authorization"
)

type Instance struct {
	Name    string
	HomeDir string
}

func (m *ZomboidInstanceManager) ListInstances(ctx context.Context, principal authorization.Principal) ([]*Instance, error) {
	instances := make([]*Instance, 0)

	for _, instanceConfig := range m.appConfig.Zomboid.Instances {
		if (slices.Contains(instanceConfig.Users, principal.Email)) {
			instances = append(instances, &Instance{
				Name:    instanceConfig.Name,
				HomeDir: instanceConfig.HomeDir,
			})
		}
	}

	return instances, nil
}
