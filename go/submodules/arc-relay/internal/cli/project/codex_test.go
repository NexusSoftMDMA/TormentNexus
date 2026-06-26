package project

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
)

func TestCodexTargetName(t *testing.T) {
	target := &CodexTarget{}
	if target.Name() != "codex" {
		t.Errorf("Name() = %q, want %q", target.Name(), "codex")
	}
}

func TestCodexTargetConfigFileName(t *testing.T) {
	target := &CodexTarget{}
	if target.ConfigFileName() != ".codex/config.toml" {
		t.Errorf("ConfigFileName() = %q, want %q", target.ConfigFileName(), ".codex/config.toml")
	}
}

func TestCodexTargetDetect(t *testing.T) {
	target := &CodexTarget{}

	t.Run("exists", func(t *testing.T) {
		dir := t.TempDir()
		codexDir := filepath.Join(dir, ".codex")
		if err := os.MkdirAll(codexDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(codexDir, "config.toml"), []byte(""), 0644); err != nil {
			t.Fatal(err)
		}
		if !target.Detect(dir) {
			t.Error("expected Detect to return true when .codex/config.toml exists")
		}
	})

	t.Run("missing", func(t *testing.T) {
		dir := t.TempDir()
		if target.Detect(dir) {
			t.Error("expected Detect to return false when .codex/config.toml is missing")
		}
	})
}

func TestCodexTargetReadEmpty(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".codex"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".codex", "config.toml"), []byte("[mcp_servers]\n"), 0644); err != nil {
		t.Fatal(err)
	}

	target := &CodexTarget{}
	servers, err := target.Read(dir, testRelayURL)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(servers) != 0 {
		t.Errorf("expected 0 servers, got %d", len(servers))
	}
}

func TestCodexTargetReadNoFile(t *testing.T) {
	dir := t.TempDir()

	target := &CodexTarget{}
	servers, err := target.Read(dir, testRelayURL)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if servers != nil {
		t.Errorf("expected nil for missing file, got %v", servers)
	}
}

func TestCodexTargetReadRelayServers(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".codex"), 0755); err != nil {
		t.Fatal(err)
	}

	config := `
[mcp_servers.sentry]
url = "http://127.0.0.1:8080/mcp/sentry"
http_headers = {Authorization = "Bearer key"}

[mcp_servers.manual-server]
command = "node"
args = ["server.js"]

[mcp_servers.pfsense]
url = "http://127.0.0.1:8080/mcp/pfsense"
http_headers = {Authorization = "Bearer key"}
`
	if err := os.WriteFile(filepath.Join(dir, ".codex", "config.toml"), []byte(config), 0644); err != nil {
		t.Fatal(err)
	}

	target := &CodexTarget{}
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

func TestCodexTargetReadDifferentRelay(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".codex"), 0755); err != nil {
		t.Fatal(err)
	}

	config := `
[mcp_servers.sentry]
url = "http://127.0.0.1:8080/mcp/sentry"
http_headers = {Authorization = "Bearer key"}
`
	if err := os.WriteFile(filepath.Join(dir, ".codex", "config.toml"), []byte(config), 0644); err != nil {
		t.Fatal(err)
	}

	target := &CodexTarget{}
	servers, err := target.Read(dir, "http://other-relay:9090")
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if len(servers) != 0 {
		t.Errorf("expected 0 servers for different relay URL, got %d", len(servers))
	}
}

func TestCodexTargetWriteNewFile(t *testing.T) {
	dir := t.TempDir()
	target := &CodexTarget{}

	servers := []ManagedServer{
		{Name: "sentry", URL: "http://127.0.0.1:8080/mcp/sentry"},
	}

	if err := target.Write(dir, testRelayURL, "test-key", servers); err != nil {
		t.Fatalf("Write: %v", err)
	}

	path := filepath.Join(dir, ".codex", "config.toml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading config.toml: %v", err)
	}

	var raw map[string]any
	if _, err := toml.Decode(string(data), &raw); err != nil {
		t.Fatalf("parsing config.toml: %v", err)
	}

	mcpServers := raw["mcp_servers"].(map[string]any)
	entry := mcpServers["sentry"].(map[string]any)
	headers := entry["http_headers"].(map[string]any)

	if entry["url"] != "http://127.0.0.1:8080/mcp/sentry" {
		t.Errorf("url = %q, want http://127.0.0.1:8080/mcp/sentry", entry["url"])
	}
	if headers["Authorization"] != "Bearer test-key" {
		t.Errorf("Authorization = %q, want %q", headers["Authorization"], "Bearer test-key")
	}
	if !strings.Contains(string(data), "mcp_servers") {
		t.Error("expected mcp_servers table in config")
	}
}

func TestCodexTargetWritePreservesExisting(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".codex"), 0755); err != nil {
		t.Fatal(err)
	}

	existing := `
model = "gpt-5"

[profiles.default]
sandbox_mode = "workspace-write"

[mcp_servers.manual-server]
command = "node"
args = ["server.js"]
`
	if err := os.WriteFile(filepath.Join(dir, ".codex", "config.toml"), []byte(existing), 0644); err != nil {
		t.Fatal(err)
	}

	target := &CodexTarget{}
	servers := []ManagedServer{
		{Name: "sentry", URL: "http://127.0.0.1:8080/mcp/sentry"},
	}

	if err := target.Write(dir, testRelayURL, "key", servers); err != nil {
		t.Fatalf("Write: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".codex", "config.toml"))
	if err != nil {
		t.Fatalf("reading config.toml: %v", err)
	}

	var raw map[string]any
	if _, err := toml.Decode(string(data), &raw); err != nil {
		t.Fatalf("parsing config.toml: %v", err)
	}

	if raw["model"] != "gpt-5" {
		t.Errorf("model = %v, want %q", raw["model"], "gpt-5")
	}

	profiles := raw["profiles"].(map[string]any)
	defaultProfile := profiles["default"].(map[string]any)
	if defaultProfile["sandbox_mode"] != "workspace-write" {
		t.Errorf("sandbox_mode = %v, want %q", defaultProfile["sandbox_mode"], "workspace-write")
	}

	mcpServers := raw["mcp_servers"].(map[string]any)
	if _, ok := mcpServers["manual-server"]; !ok {
		t.Error("expected manual-server to be preserved")
	}
	if _, ok := mcpServers["sentry"]; !ok {
		t.Error("expected sentry to be added")
	}
}

func TestCodexTargetWriteMultipleServers(t *testing.T) {
	dir := t.TempDir()
	target := &CodexTarget{}

	servers := []ManagedServer{
		{Name: "sentry", URL: "http://127.0.0.1:8080/mcp/sentry"},
		{Name: "pfsense", URL: "http://127.0.0.1:8080/mcp/pfsense"},
		{Name: "shortcut", URL: "http://127.0.0.1:8080/mcp/shortcut"},
	}

	if err := target.Write(dir, testRelayURL, "key", servers); err != nil {
		t.Fatalf("Write: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".codex", "config.toml"))
	if err != nil {
		t.Fatalf("reading config.toml: %v", err)
	}

	var raw map[string]any
	if _, err := toml.Decode(string(data), &raw); err != nil {
		t.Fatalf("parsing config.toml: %v", err)
	}

	mcpServers := raw["mcp_servers"].(map[string]any)
	if len(mcpServers) != 3 {
		t.Errorf("expected 3 servers, got %d", len(mcpServers))
	}
}

func TestCodexTargetWriteIdempotent(t *testing.T) {
	dir := t.TempDir()
	target := &CodexTarget{}

	servers := []ManagedServer{
		{Name: "sentry", URL: "http://127.0.0.1:8080/mcp/sentry"},
	}

	if err := target.Write(dir, testRelayURL, "key", servers); err != nil {
		t.Fatal(err)
	}
	if err := target.Write(dir, testRelayURL, "key", servers); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".codex", "config.toml"))
	if err != nil {
		t.Fatalf("reading config.toml: %v", err)
	}

	var raw map[string]any
	if _, err := toml.Decode(string(data), &raw); err != nil {
		t.Fatalf("parsing config.toml: %v", err)
	}

	mcpServers := raw["mcp_servers"].(map[string]any)
	if len(mcpServers) != 1 {
		t.Errorf("expected 1 server after idempotent write, got %d", len(mcpServers))
	}
}

func TestCodexTargetRemove(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".codex"), 0755); err != nil {
		t.Fatal(err)
	}

	config := `
model = "gpt-5"

[mcp_servers.sentry]
url = "http://127.0.0.1:8080/mcp/sentry"
http_headers = {Authorization = "Bearer key"}

[mcp_servers.manual-server]
command = "node"
args = ["server.js"]
`
	if err := os.WriteFile(filepath.Join(dir, ".codex", "config.toml"), []byte(config), 0644); err != nil {
		t.Fatal(err)
	}

	target := &CodexTarget{}
	removed, err := target.Remove(dir, []string{"sentry"})
	if err != nil {
		t.Fatalf("Remove: %v", err)
	}

	if len(removed) != 1 || removed[0] != "sentry" {
		t.Fatalf("removed = %v, want [sentry]", removed)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".codex", "config.toml"))
	if err != nil {
		t.Fatalf("reading config.toml: %v", err)
	}

	var raw map[string]any
	if _, err := toml.Decode(string(data), &raw); err != nil {
		t.Fatalf("parsing config.toml: %v", err)
	}

	if raw["model"] != "gpt-5" {
		t.Errorf("model = %v, want %q", raw["model"], "gpt-5")
	}

	mcpServers := raw["mcp_servers"].(map[string]any)
	if _, ok := mcpServers["sentry"]; ok {
		t.Error("expected sentry to be removed")
	}
	if _, ok := mcpServers["manual-server"]; !ok {
		t.Error("expected manual-server to be preserved")
	}
}
