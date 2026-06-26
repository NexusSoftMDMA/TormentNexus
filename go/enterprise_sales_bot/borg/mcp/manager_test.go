package mcp

import (
	"os"
	"path/filepath"
	"testing"
)

func TestServerManagerUsesAdapterConfigPathAndToolHints(t *testing.T) {
	home := t.TempDir()
	tormentnexusDir := filepath.Join(home, ".tormentnexus")
	if err := os.MkdirAll(tormentnexusDir, 0o755); err != nil {
		t.Fatal(err)
	}
	config := `{"mcpServers":{"demo":{"command":"cmd","args":["/c","echo demo"]}}}`
	configPath := filepath.Join(tormentnexusDir, "mcp.json")
	if err := os.WriteFile(configPath, []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}
	setMCPEnv(t, home)
	manager := NewServerManager()
	manager.RegistryPath = configPath
	tools, err := manager.ListConfiguredTools()
	if err != nil {
		t.Fatal(err)
	}
	if len(tools) == 0 {
		t.Fatal("expected configured tool hints")
	}
	route, err := manager.RouteConfiguredToolCall("demo", "list-tools", map[string]interface{}{"limit": 2})
	if err != nil {
		t.Fatal(err)
	}
	if route == "" {
		t.Fatal("expected MCP route")
	}
	if _, err := manager.StartConfiguredServer("missing"); err == nil {
		t.Fatal("expected missing server error")
	}
}
