package mcptool

import (
	"context"
	"errors"
	"log/slog"
	"runtime/debug"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/zed-assistant/mcp/internal/auth/authorization"
	"github.com/zed-assistant/mcp/internal/zomboid/instance"
)

type McpToolManger struct {
	logger                 *slog.Logger
	zomboidInstanceManager *instance.ZomboidInstanceManager
}

func NewMcpToolManager(logger *slog.Logger, zomboidInstanceManager *instance.ZomboidInstanceManager) *McpToolManger {
	return &McpToolManger{
		logger:                 logger,
		zomboidInstanceManager: zomboidInstanceManager,
	}
}

func (m *McpToolManger) CollectTools() []Tool {
	return []Tool{
		m.ListZomboidInstances(),
	}
}

type EmptyInput struct{}

func withRecover[In, Out any](log *slog.Logger, h mcp.ToolHandlerFor[In, Out]) mcp.ToolHandlerFor[In, Out] {
	return func(ctx context.Context, req *mcp.CallToolRequest, in In) (result *mcp.CallToolResult, output Out, err error) {
		defer func() {
			if rvr := recover(); rvr != nil {
				log.ErrorContext(ctx, "Panic recovered in tool handler",
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

func withUserRecover[In, Out any](log *slog.Logger, h userToolFunc[In, Out]) mcp.ToolHandlerFor[In, Out] {
	return withRecover(log, withUser(log, h))
}

func withUserRecoverNoInput[Out any](log *slog.Logger, h func(ctx context.Context, principal authorization.Principal) (Out, error)) mcp.ToolHandlerFor[EmptyInput, Out] {
	return withUserRecover(log, func(ctx context.Context, principal authorization.Principal, _ EmptyInput) (Out, error) {
		return h(ctx, principal)
	})
}

type userToolFunc[In, Out any] func(ctx context.Context, principal authorization.Principal, in In) (Out, error)

func withUser[In, Out any](log *slog.Logger, h userToolFunc[In, Out],
) func(context.Context, *mcp.CallToolRequest, In) (*mcp.CallToolResult, Out, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in In) (*mcp.CallToolResult, Out, error) {
		var zero Out
		info := req.Extra.TokenInfo
		if info == nil {
			log.ErrorContext(ctx, "Tool handler called without token info")
			return nil, zero, errors.New("unauthenticated")
		}

		email, ok := info.Extra["Email"].(string)
		if !ok {
			log.ErrorContext(ctx, "Tool handler called without email in token info")
			return nil, zero, errors.New("unauthenticated")
		}

		principal := authorization.Principal{
			Subject: info.UserID,
			Email:   email,
		}

		out, err := h(ctx, principal, in)
		if err != nil {
			log.ErrorContext(ctx, "Tool handler returned an error",
				slog.Any("error", err))
			return nil, zero, errors.New("internal error")
		}
		return nil, out, nil
	}
}
