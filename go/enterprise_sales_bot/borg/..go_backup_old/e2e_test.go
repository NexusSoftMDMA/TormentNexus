package tools

import (
	"context"
	"testing"
)

func TestE2E_RegistryAndTools(t *testing.T) {
	registry := NewRegistry()
	ctx := context.Background()

	tests := []struct {
		name string
		args map[string]interface{}
	}{
		{"ripgrep_search", map[string]interface{}{"pattern": "func", "path": "."}},
		{"anyquery", map[string]interface{}{"query": "SELECT * FROM files"}},
		{"codemod", map[string]interface{}{"command": "list"}},
		{"puppeteer_navigate", map[string]interface{}{"action": "goto", "url": "https://google.com"}},
		{"agentcortex_mcp", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !registry.HasTool(tt.name) {
				t.Fatalf("Registry missing tool: %s", tt.name)
			}

			resp, err := registry.Execute(ctx, tt.name, tt.args)
			if err != nil {
				t.Fatalf("Execution failed for %s: %v", tt.name, err)
			}

			if resp.IsError {
				t.Fatalf("Tool %s returned error: %s", tt.name, resp.Content[0].Text)
			}

			if len(resp.Content) == 0 || resp.Content[0].Text == "" {
				t.Fatalf("Tool %s returned empty response", tt.name)
			}
		})
	}
}
