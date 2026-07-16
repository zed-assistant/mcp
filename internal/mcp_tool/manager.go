package mcptool

import (
	"context"
	"errors"
	"log/slog"
	"runtime/debug"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/zed-assistant/mcp/internal/auth/authorization"
	domainerror "github.com/zed-assistant/mcp/internal/domain_error"
	"github.com/zed-assistant/mcp/internal/logger"
	"github.com/zed-assistant/mcp/internal/zomboid/instance"
)

var validate = validator.New()
var enLocale = en.New()
var uni = ut.New(enLocale, enLocale)
var trans, _ = uni.GetTranslator("en")
var _ = en_translations.RegisterDefaultTranslations(validate, trans)

type McpToolManager struct {
	logger                 *slog.Logger
	zomboidInstanceManager *instance.ZomboidInstanceManager
}

func NewMcpToolManager(logger *slog.Logger, zomboidInstanceManager *instance.ZomboidInstanceManager) *McpToolManager {
	return &McpToolManager{
		logger:                 logger,
		zomboidInstanceManager: zomboidInstanceManager,
	}
}

func (m *McpToolManager) CollectTools() []Tool {
	return []Tool{
		m.ListZomboidInstances(),
		m.ReadZomboidServerConfig(),
		m.UpdateZomboidServerConfig(),
		m.ExecuteRawAdminCommand(),
	}
}

type Empty struct{}

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

func withUserRecoverNoInput[Out any](log *slog.Logger, h func(ctx context.Context, principal authorization.Principal) (Out, error)) mcp.ToolHandlerFor[Empty, Out] {
	return withUserRecover(log, noInput(h))
}

func withUserRecoverNoOutput[In any](log *slog.Logger, h func(ctx context.Context, principal authorization.Principal, in In) error) mcp.ToolHandlerFor[In, Empty] {
	return withUserRecover(log, noOutput(h))
}

func withUserRecoverNoInputNoOutput(log *slog.Logger, h func(ctx context.Context, principal authorization.Principal) error) mcp.ToolHandlerFor[Empty, Empty] {
	return withUserRecover(log, noInput(func(ctx context.Context, principal authorization.Principal) (Empty, error) {
		return Empty{}, h(ctx, principal)
	}))
}

type userToolFunc[In, Out any] func(ctx context.Context, principal authorization.Principal, in In) (Out, error)

// noInput adapts a domain function that takes no input into the userToolFunc[Empty, Out] shape.
func noInput[Out any](h func(ctx context.Context, principal authorization.Principal) (Out, error)) userToolFunc[Empty, Out] {
	return func(ctx context.Context, principal authorization.Principal, _ Empty) (Out, error) {
		return h(ctx, principal)
	}
}

// noOutput adapts a domain function that returns only an error into the userToolFunc[In, Empty] shape.
func noOutput[In any](h func(ctx context.Context, principal authorization.Principal, in In) error) userToolFunc[In, Empty] {
	return func(ctx context.Context, principal authorization.Principal, in In) (Empty, error) {
		return Empty{}, h(ctx, principal, in)
	}
}

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

		if err := validate.Struct(in); err != nil {
			log.WarnContext(ctx, "Tool handler called with invalid input", logger.LogError(err))
			translatedErrs := make([]string, 0)
			for _, e := range err.(validator.ValidationErrors) {
				translatedErrs = append(translatedErrs, e.Translate(trans))
			}
			return nil, zero, errors.New("Invalid input: " + strings.Join(translatedErrs, "; "))
		}

		out, err := h(ctx, principal, in)
		if err != nil {
			log.ErrorContext(ctx, "Tool handler returned an error", logger.LogError(err))
			if domainErr, ok := errors.AsType[*domainerror.DomainError](err); ok {
				return nil, zero, errors.New(domainErr.PublicMessage)
			}
			return nil, zero, errors.New("internal error")
		}
		return nil, out, nil
	}
}
