package instance

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/zed-assistant/mcp/internal/auth/authorization"
	domainerror "github.com/zed-assistant/mcp/internal/domain_error"
	filesystem "github.com/zed-assistant/mcp/internal/file_system"
)

const gameLogFileName = "server-console.txt"

type GetGameLogsInput struct {
	InstanceID string
	Lines      int
	Filter     string
}

func (m *ZomboidInstanceManager) GetGameLogs(ctx context.Context, principal authorization.Principal, input *GetGameLogsInput) ([]string, error) {
	if err := m.instanceAuth.AuthorizeInstanceAccess(input.InstanceID, principal); err != nil {
		return nil, err
	}

	m.instanceLockManager.RLock(input.InstanceID)
	defer m.instanceLockManager.RUnlock(input.InstanceID)

	instanceCfg := m.appConfig.Zomboid.Instances[input.InstanceID]
	logPath := filepath.Join(instanceCfg.HomeDir, gameLogFileName)

	logFileExists, err := filesystem.FileExists(logPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check game log file: %w", err)
	}
	if !logFileExists {
		return nil, &domainerror.DomainError{
			InternalMessage: fmt.Sprintf("game log file does not exist: %s", logPath),
			PublicMessage:   "Game log file was not found for this server instance",
			InternalCode:    domainerror.NotFound,
		}
	}

	var match func(line string) bool
	if input.Filter != "" {
		re := compileLineFilter(input.Filter)
		match = re.MatchString
	}

	lines, err := filesystem.ReadLinesFromEnd(logPath, input.Lines, match)
	if err != nil {
		return nil, fmt.Errorf("failed to read game log file: %w", err)
	}

	return lines, nil
}

// compileLineFilter builds a case-insensitive, unanchored matcher from pattern: plain text is matched as a
// substring anywhere in the line, and * matches any run of characters, e.g. "wa*n" matches "warn".
func compileLineFilter(pattern string) *regexp.Regexp {
	parts := strings.Split(pattern, "*")
	var sb strings.Builder
	sb.WriteString("(?i)")
	for i, part := range parts {
		if i > 0 {
			sb.WriteString(".*")
		}
		sb.WriteString(regexp.QuoteMeta(part))
	}
	return regexp.MustCompile(sb.String())
}
