package instance

import (
	"context"
	"slices"

	"github.com/zed-assistant/mcp/internal/auth/authorization"
)

type Instance struct {
	Name    string `json:"name" jsonschema:"Name of the Project Zomboid server instance"`
	HomeDir string `json:"homeDir" jsonschema:"Absolute path to the instance's home directory on disk"`
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
