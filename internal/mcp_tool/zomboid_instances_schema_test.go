package mcptool

import (
	"log/slog"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestConfigToolsRegisterWithoutSchemaError guards against the map[string]any /
// nested-tree output types causing a JSON-schema reflection failure (e.g. a
// cycle) when the tools are registered with the real SDK server, since that
// would only surface at process startup otherwise.
func TestConfigToolsRegisterWithoutSchemaError(t *testing.T) {
	m := NewMcpToolManager(slog.Default(), nil)

	server := mcp.NewServer(&mcp.Implementation{
		Name:  "test-server",
		Title: "test",
	}, nil)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("registering config tools panicked: %v", r)
		}
	}()

	m.ReadZomboidConfig().Register(server)
	m.UpdateZomboidConfig().Register(server)
}
