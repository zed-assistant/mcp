package mcptool

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type McpToolManger struct {
	logger *slog.Logger
}

func NewMcpToolManager(logger *slog.Logger) *McpToolManger {
	return &McpToolManger{logger: logger}
}

func (m *McpToolManger) CollectTools() []Tool {
	return []Tool{
		&MCPTool[EmptyOutput, StringOutput]{
			Definition: &mcp.Tool{
				Name:        "whoami",
				Description: "Returns information about the authenticated user",
			},
			Handler: withRecover(m.logger, withUser(m.logger, whoami)),
		},
	}
}

type EmptyOutput struct{}

type StringOutput struct {
	Text string
}

func whoami(_ context.Context, email string, _ EmptyOutput) (StringOutput, error) {
	return StringOutput{
		Text: fmt.Sprintf("You are authenticated as email: %s", email),
	}, nil
}

func withRecover[In, Out any](logger *slog.Logger, h mcp.ToolHandlerFor[In, Out]) mcp.ToolHandlerFor[In, Out] {
	return func(ctx context.Context, req *mcp.CallToolRequest, in In) (result *mcp.CallToolResult, output Out, err error) {
		defer func() {
			if rvr := recover(); rvr != nil {
				logger.ErrorContext(ctx, "Panic recovered in tool handler",
					slog.Any("error", rvr),
					slog.String("stack", string(debug.Stack())))

				var zero Out
				result = nil
				output = zero
				err = errors.New("internal error")
			}
		}()

		return h(ctx, req, in)
	}
}

type userToolFunc[In, Out any] func(ctx context.Context, email string, in In) (Out, error)

func withUser[In, Out any](logger *slog.Logger, h userToolFunc[In, Out],
) func(context.Context, *mcp.CallToolRequest, In) (*mcp.CallToolResult, Out, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in In) (*mcp.CallToolResult, Out, error) {
		var zero Out
		info := req.Extra.TokenInfo
		if info == nil {
			logger.ErrorContext(ctx, "Tool handler called without token info")
			return nil, zero, errors.New("unauthenticated")
		}

		email, ok := info.Extra["Email"].(string)
		if !ok {
			logger.ErrorContext(ctx, "Tool handler called without email in token info")
			return nil, zero, errors.New("unauthenticated")
		}

		out, err := h(ctx, email, in)
		if err != nil {
			logger.ErrorContext(ctx, "Tool handler returned an error",
				slog.Any("error", err))
			return nil, zero, errors.New("internal error")
		}
		return nil, out, nil
	}
}
