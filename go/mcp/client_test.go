package mcp

import (
	"os"
	"path/filepath"
	"testing"
)

func TestClientUsesAdapterToolHints(t *testing.T) {
	home := t.TempDir()
	tormentnexusDir := filepath.Join(home, ".tormentnexus")
	if err := os.MkdirAll(tormentnexusDir, 0o755); err != nil {
		t.Fatal(err)
	}
	config := `{"mcpServers":{"demo":{"command":"cmd","args":["/c","echo demo"]}}}`
	if err := os.WriteFile(filepath.Join(tormentnexusDir, "mcp.json"), []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}
	setMCPEnv(t, home)
	client := NewClient("")
	if err := client.Connect(); err != nil {
		t.Fatal(err)
	}
	tools, err := client.ListTools()
	if err != nil {
		t.Fatal(err)
	}
	if len(tools) == 0 {
		t.Fatal("expected MCP tool hints")
	}
	route, err := client.CallTool("demo", "list-tools", map[string]interface{}{"limit": 1})
	if err != nil {
		t.Fatal(err)
	}
	if route == "" {
		t.Fatal("expected MCP route")
	}
}

func setMCPEnv(t *testing.T, home string) {
	t.Helper()
	for _, key := range []string{"HOME", "USERPROFILE"} {
		old, had := os.LookupEnv(key)
		if err := os.Setenv(key, home); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if had {
				_ = os.Setenv(key, old)
			} else {
				_ = os.Unsetenv(key)
			}
		})
	}
}
