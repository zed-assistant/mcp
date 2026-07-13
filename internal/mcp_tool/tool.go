package mcptool

import "github.com/modelcontextprotocol/go-sdk/mcp"

type Tool interface {
	Register(s *mcp.Server)
}

type MCPTool[In, Out any] struct {
	Definition *mcp.Tool
	Handler    mcp.ToolHandlerFor[In, Out]
}

func (t *MCPTool[In, Out]) Register(s *mcp.Server) {
	mcp.AddTool(s, t.Definition, t.Handler)
}
