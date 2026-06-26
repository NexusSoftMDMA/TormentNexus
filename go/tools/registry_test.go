package tools

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func TestRegistryIncludesExactPiToolsAndRepomap(t *testing.T) {
	registry := NewRegistry()
	wanted := map[string]bool{
		"read":    false,
		"write":   false,
		"edit":    false,
		"bash":    false,
		"repomap": false,
	}
	for _, tool := range registry.Tools {
		if _, ok := wanted[tool.Name]; ok {
			wanted[tool.Name] = true
			if len(tool.Parameters) == 0 {
				t.Fatalf("tool %s missing parameter schema", tool.Name)
			}
		}
	}
	for name, found := range wanted {
		if !found {
			t.Fatalf("expected tool %s in registry", name)
		}
	}
}

func TestRegistryReadToolUsesFoundationBehavior(t *testing.T) {
	registry := NewRegistry()
	var readTool Tool
	for _, tool := range registry.Tools {
		if tool.Name == "read" {
			readTool = tool
			break
		}
	}
	if readTool.Name == "" {
		t.Fatal("missing read tool")
	}
	output, err := readTool.Execute(map[string]interface{}{"path": filepath.Join("..", "go.mod")})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output, "module ") {
		t.Fatalf("unexpected read output: %q", output)
	}
	var schema map[string]any
	if err := json.Unmarshal(readTool.Parameters, &schema); err != nil {
		t.Fatalf("invalid parameter schema: %v", err)
	}
}
