package project

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

const testRelayURL = "http://127.0.0.1:8080"

func TestClaudeCodeTargetName(t *testing.T) {
	target := &ClaudeCodeTarget{}
	if target.Name() != "claude-code" {
		t.Errorf("Name() = %q, want %q", target.Name(), "claude-code")
	}
}

func TestClaudeCodeTargetConfigFileName(t *testing.T) {
	target := &ClaudeCodeTarget{}
	if target.ConfigFileName() != ".mcp.json" {
		t.Errorf("ConfigFileName() = %q, want %q", target.ConfigFileName(), ".mcp.json")
	}
}

func TestClaudeCodeTargetDetect(t *testing.T) {
	target := &ClaudeCodeTarget{}

	t.Run("exists", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, ".mcp.json"), []byte("{}"), 0644); err != nil {
			t.Fatal(err)
		}
		if !target.Detect(dir) {
			t.Error("expected Detect to return true when .mcp.json exists")
		}
	})

	t.Run("missing", func(t *testing.T) {
		dir := t.TempDir()
		if target.Detect(dir) {
			t.Error("expected Detect to return false when .mcp.json missing")
		}
	})
}

func TestClaudeCodeTargetReadEmpty(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".mcp.json"), []byte(`{"mcpServers":{}}`), 0644); err != nil {
		t.Fatal(err)
	}

	target := &ClaudeCodeTarget{}
	servers, err := target.Read(dir, testRelayURL)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(servers) != 0 {
		t.Errorf("expected 0 servers, got %d", len(servers))
	}
}

func TestClaudeCodeTargetReadNoFile(t *testing.T) {
	dir := t.TempDir()

	target := &ClaudeCodeTarget{}
	servers, err := target.Read(dir, testRelayURL)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if servers != nil {
		t.Errorf("expected nil for missing file, got %v", servers)
	}
}

func TestClaudeCodeTargetReadRelayServers(t *testing.T) {
	dir := t.TempDir()
	mcpJSON := `{
  "mcpServers": {
    "sentry": {
      "type": "http",
      "url": "http://127.0.0.1:8080/mcp/sentry",
      "headers": {"Authorization": "Bearer key"}
    },
    "manual-server": {
      "type": "stdio",
      "command": "node",
      "args": ["server.js"]
    },
    "pfsense": {
      "type": "http",
      "url": "http://127.0.0.1:8080/mcp/pfsense",
      "headers": {"Authorization": "Bearer key"}
    }
  }
}`
	if err := os.WriteFile(filepath.Join(dir, ".mcp.json"), []byte(mcpJSON), 0644); err != nil {
		t.Fatal(err)
	}

	target := &ClaudeCodeTarget{}
	servers, err := target.Read(dir, testRelayURL)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if len(servers) != 2 {
		t.Fatalf("expected 2 relay servers, got %d", len(servers))
	}

	names := map[string]bool{}
	for _, s := range servers {
		names[s.Name] = true
	}
	if !names["sentry"] {
		t.Error("expected sentry in relay servers")
	}
	if !names["pfsense"] {
		t.Error("expected pfsense in relay servers")
	}
}

func TestClaudeCodeTargetWriteNewFile(t *testing.T) {
	dir := t.TempDir()
	target := &ClaudeCodeTarget{}

	servers := []ManagedServer{
		{Name: "sentry", URL: "http://127.0.0.1:8080/mcp/sentry"},
	}

	err := target.Write(dir, testRelayURL, "test-key", servers)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	// Verify the written file
	data, err := os.ReadFile(filepath.Join(dir, ".mcp.json"))
	if err != nil {
		t.Fatalf("reading .mcp.json: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("parsing .mcp.json: %v", err)
	}

	var mcpServers map[string]json.RawMessage
	if err := json.Unmarshal(raw["mcpServers"], &mcpServers); err != nil {
		t.Fatalf("parsing mcpServers: %v", err)
	}

	if _, ok := mcpServers["sentry"]; !ok {
		t.Fatal("expected sentry in mcpServers")
	}

	var entry mcpServerEntry
	if err := json.Unmarshal(mcpServers["sentry"], &entry); err != nil {
		t.Fatalf("parsing sentry entry: %v", err)
	}

	if entry.Type != "http" {
		t.Errorf("type = %q, want %q", entry.Type, "http")
	}
	if entry.URL != "http://127.0.0.1:8080/mcp/sentry" {
		t.Errorf("url = %q, want http://127.0.0.1:8080/mcp/sentry", entry.URL)
	}
	if entry.Headers["Authorization"] != "Bearer test-key" {
		t.Errorf("Authorization = %q, want %q", entry.Headers["Authorization"], "Bearer test-key")
	}
}

func TestClaudeCodeTargetWritePreservesExisting(t *testing.T) {
	dir := t.TempDir()
	existing := `{
  "mcpServers": {
    "manual-server": {
      "type": "stdio",
      "command": "node",
      "args": ["server.js"]
    }
  },
  "someOtherKey": "preserve-me"
}`
	if err := os.WriteFile(filepath.Join(dir, ".mcp.json"), []byte(existing), 0644); err != nil {
		t.Fatal(err)
	}

	target := &ClaudeCodeTarget{}
	servers := []ManagedServer{
		{Name: "sentry", URL: "http://127.0.0.1:8080/mcp/sentry"},
	}

	err := target.Write(dir, testRelayURL, "key", servers)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".mcp.json"))
	if err != nil {
		t.Fatalf("reading .mcp.json: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("parsing .mcp.json: %v", err)
	}

	// Check that someOtherKey is preserved
	if _, ok := raw["someOtherKey"]; !ok {
		t.Error("expected someOtherKey to be preserved")
	}

	// Check that manual-server is preserved
	var mcpServers map[string]json.RawMessage
	if err := json.Unmarshal(raw["mcpServers"], &mcpServers); err != nil {
		t.Fatalf("parsing mcpServers: %v", err)
	}

	if _, ok := mcpServers["manual-server"]; !ok {
		t.Error("expected manual-server to be preserved")
	}
	if _, ok := mcpServers["sentry"]; !ok {
		t.Error("expected sentry to be added")
	}
}

func TestClaudeCodeTargetWriteMultipleServers(t *testing.T) {
	dir := t.TempDir()
	target := &ClaudeCodeTarget{}

	servers := []ManagedServer{
		{Name: "sentry", URL: "http://127.0.0.1:8080/mcp/sentry"},
		{Name: "pfsense", URL: "http://127.0.0.1:8080/mcp/pfsense"},
		{Name: "shortcut", URL: "http://127.0.0.1:8080/mcp/shortcut"},
	}

	err := target.Write(dir, testRelayURL, "key", servers)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, ".mcp.json"))
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}

	var mcpServers map[string]json.RawMessage
	if err := json.Unmarshal(raw["mcpServers"], &mcpServers); err != nil {
		t.Fatal(err)
	}

	if len(mcpServers) != 3 {
		t.Errorf("expected 3 servers, got %d", len(mcpServers))
	}
}

func TestClaudeCodeTargetWriteIdempotent(t *testing.T) {
	dir := t.TempDir()
	target := &ClaudeCodeTarget{}

	servers := []ManagedServer{
		{Name: "sentry", URL: "http://127.0.0.1:8080/mcp/sentry"},
	}

	// Write twice
	if err := target.Write(dir, testRelayURL, "key", servers); err != nil {
		t.Fatal(err)
	}
	if err := target.Write(dir, testRelayURL, "key", servers); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, ".mcp.json"))
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}

	var mcpServers map[string]json.RawMessage
	if err := json.Unmarshal(raw["mcpServers"], &mcpServers); err != nil {
		t.Fatal(err)
	}

	if len(mcpServers) != 1 {
		t.Errorf("expected 1 server after idempotent write, got %d", len(mcpServers))
	}
}

func TestClaudeCodeTargetReadDifferentRelay(t *testing.T) {
	dir := t.TempDir()
	mcpJSON := `{
  "mcpServers": {
    "sentry": {
      "type": "http",
      "url": "http://127.0.0.1:8080/mcp/sentry",
      "headers": {"Authorization": "Bearer key"}
    }
  }
}`
	if err := os.WriteFile(filepath.Join(dir, ".mcp.json"), []byte(mcpJSON), 0644); err != nil {
		t.Fatal(err)
	}

	target := &ClaudeCodeTarget{}
	// Read with a different relay URL — should not match
	servers, err := target.Read(dir, "http://other-relay:9090")
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if len(servers) != 0 {
		t.Errorf("expected 0 servers for different relay URL, got %d", len(servers))
	}
}
