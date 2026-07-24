package config

import (
	"fmt"
	"strings"

	domainerror "github.com/zed-assistant/mcp/internal/domain_error"
)

func NewInvalidConfigUpdateError(problems []string) *domainerror.DomainError {
	msg := strings.Join(problems, "; ")
	return &domainerror.DomainError{
		InternalMessage: "invalid config update: " + msg,
		PublicMessage:   "invalid config update: " + msg,
		InternalCode:    domainerror.InvalidInput,
	}
}

func jsonTypeName(v any) string {
	switch v.(type) {
	case nil:
		return "null"
	case bool:
		return "boolean"
	case float64:
		return "number"
	case string:
		return "string"
	case map[string]any:
		return "object"
	case []any:
		return "array"
	default:
		return fmt.Sprintf("%T", v)
	}
}
