package admincommand

import (
	"fmt"
	"strconv"

	"github.com/gorcon/rcon"
	"github.com/zed-assistant/mcp/internal/configuration"
	"github.com/zed-assistant/mcp/internal/zomboid/config"
)

type AdminCommandRCON struct {
	cfgManager *config.ConfigManager
}

func NewAdminCommandRCON(c *config.ConfigManager) *AdminCommandRCON {
	return &AdminCommandRCON{
		cfgManager: c,
	}
}

func (s *AdminCommandRCON) getRCONConfig(instanceConfig *configuration.ZomboidInstanceConfig) (int, string, error) {
	cfg, err := s.cfgManager.ReadServerConfig(instanceConfig, []string{"rcon*"})
	if err != nil {
		return 0, "", err
	}

	passwordEntry, ok := cfg["RCONPassword"]
	if !ok {
		return 0, "", fmt.Errorf("RCON password not found in server config")
	}
	if passwordEntry.Value == "" {
		return 0, "", fmt.Errorf("RCON password is empty in server config")
	}

	portEntry, ok := cfg["RCONPort"]
	if !ok {
		return 0, "", fmt.Errorf("RCON port not found in server config")
	}
	if portEntry.Value == "" {
		return 0, "", fmt.Errorf("RCON port is empty in server config")
	}
	port, err := strconv.Atoi(portEntry.Value)
	if err != nil {
		return 0, "", fmt.Errorf("invalid RCON port in server config: %w", err)
	}

	return port, passwordEntry.Value, nil
}

func (s *AdminCommandRCON) ExecuteCommand(instanceConfig *configuration.ZomboidInstanceConfig, command string) (string, error) {
	port, password, err := s.getRCONConfig(instanceConfig)
	if err != nil {
		return "", err
	}

	rconClient, err := rcon.Dial(fmt.Sprintf("%s:%d", instanceConfig.RCONHost, port), password)
	if err != nil {
		return "", fmt.Errorf("failed to connect to RCON server: %w", err)
	}
	defer func() { _ = rconClient.Close() }()

	response, err := rconClient.Execute(command)
	if err != nil {
		return "", fmt.Errorf("failed to execute RCON command: %w", err)
	}
	return response, nil
}

func (s *AdminCommandRCON) ExecuteMany(instanceConfig *configuration.ZomboidInstanceConfig, commands []string) ([]string, error) {
	port, password, err := s.getRCONConfig(instanceConfig)
	if err != nil {
		return nil, err
	}
	rconClient, err := rcon.Dial(fmt.Sprintf("%s:%d", instanceConfig.RCONHost, port), password)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RCON server: %w", err)
	}
	defer func() { _ = rconClient.Close() }()

	responses := make([]string, 0, len(commands))
	for _, command := range commands {
		response, err := rconClient.Execute(command)
		if err != nil {
			return responses, fmt.Errorf("failed to execute RCON command: %w", err)
		}
		responses = append(responses, response)
	}
	return responses, nil
}

