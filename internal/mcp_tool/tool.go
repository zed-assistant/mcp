package mcptool

import (
	"context"
	"reflect"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type Tool interface {
	Register(s *mcp.Server)
}

type MCPTool[In, Out any] struct {
	Definition *mcp.Tool
	Handler    mcp.ToolHandlerFor[In, Out]
}

type Envelope[T any] struct {
	Result T `json:"result"`
}

func isObjectShaped[T any]() bool {
	t := reflect.TypeFor[T]()
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t.Kind() == reflect.Struct || t.Kind() == reflect.Map
}

func (t *MCPTool[In, Out]) Register(s *mcp.Server) {
	if isObjectShaped[Out]() {
		mcp.AddTool(s, t.Definition, t.Handler)
		return
	}

	wrapped := func(ctx context.Context, req *mcp.CallToolRequest, in In) (*mcp.CallToolResult, Envelope[Out], error) {
		res, out, err := t.Handler(ctx, req, in)
		return res, Envelope[Out]{Result: out}, err
	}
	mcp.AddTool(s, t.Definition, wrapped)
}
