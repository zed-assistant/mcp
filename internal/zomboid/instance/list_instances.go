package instance

import (
	"context"
	"slices"

	"github.com/zed-assistant/mcp/internal/auth/authorization"
)

type Instance struct {
	ID      string `json:"id" jsonschema:"Unique identifier of the Project Zomboid server instance"`
	Name    string `json:"name" jsonschema:"Name of the Project Zomboid server instance"`
	HomeDir string `json:"homeDir" jsonschema:"Absolute path to the instance's home directory on disk"`
}

func (m *ZomboidInstanceManager) ListInstances(ctx context.Context, principal authorization.Principal) ([]*Instance, error) {
	instances := make([]*Instance, 0)

	for id, instanceConfig := range m.appConfig.Zomboid.Instances {
		if slices.Contains(instanceConfig.Users, principal.Email) {
			instances = append(instances, &Instance{
				ID:      id,
				Name:    instanceConfig.Name,
				HomeDir: instanceConfig.HomeDir,
			})
		}
	}

	return instances, nil
}
