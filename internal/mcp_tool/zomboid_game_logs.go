package mcptool

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/zed-assistant/mcp/internal/auth/authorization"
	"github.com/zed-assistant/mcp/internal/zomboid/instance"
)

const defaultGameLogLines = 100

type GetGameLogsInput struct {
	InstanceID string `json:"instanceId" jsonschema:"The ID of the Project Zomboid server instance" validate:"required"`
	Lines      *int   `json:"lines,omitempty" jsonschema:"Number of most recent matching log lines to return. Defaults to 100, maximum 1000." validate:"omitempty,min=1,max=1000"`
	Filter     string `json:"filter,omitempty" jsonschema:"Optional filter for log lines. Matches any line containing this text as a substring, case-insensitive - no wildcard needed for a plain search. You can also use * to match any run of characters, e.g. 'wa*n' matches lines containing 'warn'. When set, the 'lines' limit applies to matching lines only - the file is scanned backward from the end until enough matches are found or the start of the file is reached."`
}

func (m *McpToolManager) GetGameLogs() Tool {
	return &MCPTool[GetGameLogsInput, []string]{
		Definition: &mcp.Tool{
			Name:        "get-zomboid-game-logs",
			Description: "Returns the most recent lines from the Project Zomboid server console log, read from the end of the file backward for efficiency. Lines are ordered most recent first. Optionally filter lines with a case-insensitive substring match. Log lines do not carry real timestamps (only a per-boot relative counter), so time-window filtering ('since') is not supported.",
			Title:       "Get Project Zomboid game logs",
			Annotations: &mcp.ToolAnnotations{
				DestructiveHint: new(false),
				IdempotentHint:  true,
				OpenWorldHint:   new(false),
				ReadOnlyHint:    true,
				Title:           "Get Project Zomboid game logs",
			},
		},
		Handler: withUserRecover(m.logger, func(ctx context.Context, principal authorization.Principal, input GetGameLogsInput) ([]string, error) {
			lines := defaultGameLogLines
			if input.Lines != nil {
				lines = *input.Lines
			}

			return m.zomboidInstanceManager.GetGameLogs(ctx, principal, &instance.GetGameLogsInput{
				InstanceID: input.InstanceID,
				Lines:      lines,
				Filter:     input.Filter,
			})
		}),
	}
}
