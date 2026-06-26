package sync

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/comma-compliance/arc-relay/internal/cli/config"
)

type testServer struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	ServerType  string `json:"server_type"`
	Status      string `json:"status"`
}

func setupRelayMock(t *testing.T, servers []testServer, token string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/servers" {
			http.NotFound(w, r)
			return
		}
		if token != "" {
			auth := r.Header.Get("Authorization")
			if auth != "Bearer "+token {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(servers)
	}))
}

func setupTest(t *testing.T, servers []testServer, mcpJSON string) (configDir, projectDir string, relayURL string) {
	t.Helper()

	token := "test-key"
	ts := setupRelayMock(t, servers, token)
	t.Cleanup(ts.Close)

	configDir = t.TempDir()
	cfg := &config.Config{RelayURL: ts.URL, APIKey: token}
	if err := config.SaveConfig(configDir, cfg); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	projectDir = t.TempDir()
	if mcpJSON != "" {
		if err := os.WriteFile(filepath.Join(projectDir, ".mcp.json"), []byte(mcpJSON), 0644); err != nil {
			t.Fatalf("writing .mcp.json: %v", err)
		}
	}

	return configDir, projectDir, ts.URL
}

func TestSyncNonInteractiveAddsAll(t *testing.T) {
	servers := []testServer{
		{ID: "1", Name: "sentry", DisplayName: "Sentry", Status: "running"},
		{ID: "2", Name: "pfsense", DisplayName: "pfSense", Status: "running"},
	}

	configDir, projectDir, _ := setupTest(t, servers, "")

	var output bytes.Buffer
	result, err := Run(Options{
		ConfigDir:      configDir,
		ProjectDir:     projectDir,
		NonInteractive: true,
		Output:         &output,
		Input:          strings.NewReader(""),
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if len(result.Added) != 2 {
		t.Errorf("expected 2 added, got %d", len(result.Added))
	}

	// Verify .mcp.json was written
	data, err := os.ReadFile(filepath.Join(projectDir, ".mcp.json"))
	if err != nil {
		t.Fatalf("reading .mcp.json: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshaling .mcp.json: %v", err)
	}
	var mcpServers map[string]json.RawMessage
	if err := json.Unmarshal(raw["mcpServers"], &mcpServers); err != nil {
		t.Fatalf("unmarshaling mcpServers: %v", err)
	}

	if len(mcpServers) != 2 {
		t.Errorf("expected 2 servers in .mcp.json, got %d", len(mcpServers))
	}
}

func TestSyncSkipsAlreadyConfigured(t *testing.T) {
	servers := []testServer{
		{ID: "1", Name: "sentry", Status: "running"},
		{ID: "2", Name: "pfsense", Status: "running"},
	}

	// Pre-configure sentry
	configDir, projectDir, relayURL := setupTest(t, servers, "")
	existing := fmt.Sprintf(`{"mcpServers":{"sentry":{"type":"http","url":"%s/mcp/sentry","headers":{"Authorization":"Bearer test-key"}}}}`, relayURL)
	if err := os.WriteFile(filepath.Join(projectDir, ".mcp.json"), []byte(existing), 0644); err != nil {
		t.Fatalf("writing .mcp.json: %v", err)
	}

	var output bytes.Buffer
	result, err := Run(Options{
		ConfigDir:      configDir,
		ProjectDir:     projectDir,
		NonInteractive: true,
		Output:         &output,
		Input:          strings.NewReader(""),
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if len(result.Added) != 1 {
		t.Errorf("expected 1 added (pfsense only), got %d: %v", len(result.Added), result.Added)
	}
	if len(result.Existed) != 1 {
		t.Errorf("expected 1 existing, got %d", len(result.Existed))
	}
}

func TestSyncSkipsStoppedServers(t *testing.T) {
	servers := []testServer{
		{ID: "1", Name: "sentry", Status: "running"},
		{ID: "2", Name: "broken", Status: "stopped"},
	}

	configDir, projectDir, _ := setupTest(t, servers, "")

	var output bytes.Buffer
	result, err := Run(Options{
		ConfigDir:      configDir,
		ProjectDir:     projectDir,
		NonInteractive: true,
		Output:         &output,
		Input:          strings.NewReader(""),
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if len(result.Added) != 1 {
		t.Errorf("expected 1 added (only running), got %d", len(result.Added))
	}
}

func TestSyncInteractiveYesNo(t *testing.T) {
	servers := []testServer{
		{ID: "1", Name: "sentry", Status: "running"},
		{ID: "2", Name: "pfsense", Status: "running"},
	}

	configDir, projectDir, _ := setupTest(t, servers, "")

	// User says yes to sentry, no to pfsense
	input := "y\nn\n"
	var output bytes.Buffer
	result, err := Run(Options{
		ConfigDir:  configDir,
		ProjectDir: projectDir,
		Output:     &output,
		Input:      strings.NewReader(input),
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if len(result.Added) != 1 {
		t.Errorf("expected 1 added, got %d", len(result.Added))
	}
	// "n" now skips permanently
	if len(result.Skipped) != 1 {
		t.Errorf("expected 1 skipped, got %d", len(result.Skipped))
	}

	// Verify pfsense is in skip list and won't be prompted again
	state, _ := config.LoadState(configDir)
	if !state.IsSkipped(projectDir, "pfsense") {
		t.Error("expected pfsense to be in skip list after saying no")
	}
}

func TestSyncInteractiveSkip(t *testing.T) {
	servers := []testServer{
		{ID: "1", Name: "sentry", Status: "running"},
	}

	configDir, projectDir, _ := setupTest(t, servers, "")

	input := "s\n"
	var output bytes.Buffer
	result, err := Run(Options{
		ConfigDir:  configDir,
		ProjectDir: projectDir,
		Output:     &output,
		Input:      strings.NewReader(input),
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if len(result.Skipped) != 1 {
		t.Errorf("expected 1 skipped, got %d", len(result.Skipped))
	}

	// Verify state was saved
	state, _ := config.LoadState(configDir)
	if !state.IsSkipped(projectDir, "sentry") {
		t.Error("expected sentry to be in skip list")
	}

	// Run again — should not prompt for sentry
	var output2 bytes.Buffer
	result2, err := Run(Options{
		ConfigDir:  configDir,
		ProjectDir: projectDir,
		Output:     &output2,
		Input:      strings.NewReader(""),
	})
	if err != nil {
		t.Fatalf("Run (2nd): %v", err)
	}

	if len(result2.Added) != 0 {
		t.Errorf("expected 0 added on second run, got %d", len(result2.Added))
	}
}

func TestSyncDryRun(t *testing.T) {
	servers := []testServer{
		{ID: "1", Name: "sentry", Status: "running"},
	}

	configDir, projectDir, _ := setupTest(t, servers, "")

	var output bytes.Buffer
	_, err := Run(Options{
		ConfigDir:      configDir,
		ProjectDir:     projectDir,
		NonInteractive: true,
		DryRun:         true,
		Output:         &output,
		Input:          strings.NewReader(""),
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if !strings.Contains(output.String(), "DRY RUN") {
		t.Error("expected DRY RUN in output")
	}

	// Verify no .mcp.json was written
	if _, err := os.Stat(filepath.Join(projectDir, ".mcp.json")); !os.IsNotExist(err) {
		t.Error("expected .mcp.json to NOT be created during dry run")
	}
}

func TestSyncPreservesManualEntries(t *testing.T) {
	servers := []testServer{
		{ID: "1", Name: "sentry", Status: "running"},
	}

	existing := `{
  "mcpServers": {
    "my-local-server": {
      "type": "stdio",
      "command": "node",
      "args": ["server.js"]
    }
  }
}`

	configDir, projectDir, _ := setupTest(t, servers, existing)

	var output bytes.Buffer
	_, err := Run(Options{
		ConfigDir:      configDir,
		ProjectDir:     projectDir,
		NonInteractive: true,
		Output:         &output,
		Input:          strings.NewReader(""),
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(projectDir, ".mcp.json"))
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshaling .mcp.json: %v", err)
	}
	var mcpServers map[string]json.RawMessage
	if err := json.Unmarshal(raw["mcpServers"], &mcpServers); err != nil {
		t.Fatalf("unmarshaling mcpServers: %v", err)
	}

	if _, ok := mcpServers["my-local-server"]; !ok {
		t.Error("expected my-local-server to be preserved")
	}
	if _, ok := mcpServers["sentry"]; !ok {
		t.Error("expected sentry to be added")
	}
}

func TestSyncNoRunningServers(t *testing.T) {
	servers := []testServer{
		{ID: "1", Name: "broken", Status: "stopped"},
	}

	configDir, projectDir, _ := setupTest(t, servers, "")

	var output bytes.Buffer
	result, err := Run(Options{
		ConfigDir:      configDir,
		ProjectDir:     projectDir,
		NonInteractive: true,
		Output:         &output,
		Input:          strings.NewReader(""),
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if len(result.Added) != 0 {
		t.Errorf("expected 0 added, got %d", len(result.Added))
	}
	if !strings.Contains(output.String(), "No running servers") {
		t.Error("expected 'No running servers' message")
	}
}

func TestSyncWritesAllDetectedTargets(t *testing.T) {
	servers := []testServer{
		{ID: "1", Name: "sentry", Status: "running"},
	}

	configDir, projectDir, _ := setupTest(t, servers, `{"mcpServers":{}}`)
	if err := os.MkdirAll(filepath.Join(projectDir, ".codex"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, ".codex", "config.toml"), []byte("[mcp_servers]\n"), 0644); err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer
	result, err := Run(Options{
		ConfigDir:      configDir,
		ProjectDir:     projectDir,
		NonInteractive: true,
		Output:         &output,
		Input:          strings.NewReader(""),
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(result.Added) != 1 || result.Added[0] != "sentry" {
		t.Fatalf("expected sentry to be added, got %v", result.Added)
	}

	data, err := os.ReadFile(filepath.Join(projectDir, ".mcp.json"))
	if err != nil {
		t.Fatalf("reading .mcp.json: %v", err)
	}
	var rawJSON map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawJSON); err != nil {
		t.Fatalf("parsing .mcp.json: %v", err)
	}
	var mcpServers map[string]json.RawMessage
	if err := json.Unmarshal(rawJSON["mcpServers"], &mcpServers); err != nil {
		t.Fatalf("parsing mcpServers: %v", err)
	}
	if _, ok := mcpServers["sentry"]; !ok {
		t.Fatal("expected sentry in .mcp.json")
	}

	tomlData, err := os.ReadFile(filepath.Join(projectDir, ".codex", "config.toml"))
	if err != nil {
		t.Fatalf("reading config.toml: %v", err)
	}
	var rawTOML map[string]any
	if _, err := toml.Decode(string(tomlData), &rawTOML); err != nil {
		t.Fatalf("parsing config.toml: %v", err)
	}
	codexServers := rawTOML["mcp_servers"].(map[string]any)
	if _, ok := codexServers["sentry"]; !ok {
		t.Fatal("expected sentry in .codex/config.toml")
	}
}

func TestSyncConfiguredIfPresentInAnyTarget(t *testing.T) {
	servers := []testServer{
		{ID: "1", Name: "sentry", Status: "running"},
		{ID: "2", Name: "pfsense", Status: "running"},
	}

	configDir, projectDir, relayURL := setupTest(t, servers, `{"mcpServers":{}}`)
	if err := os.MkdirAll(filepath.Join(projectDir, ".codex"), 0755); err != nil {
		t.Fatal(err)
	}
	codexConfig := fmt.Sprintf(`
[mcp_servers.sentry]
url = "%s/mcp/sentry"
http_headers = {Authorization = "Bearer test-key"}
`, relayURL)
	if err := os.WriteFile(filepath.Join(projectDir, ".codex", "config.toml"), []byte(codexConfig), 0644); err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer
	result, err := Run(Options{
		ConfigDir:      configDir,
		ProjectDir:     projectDir,
		NonInteractive: true,
		Output:         &output,
		Input:          strings.NewReader(""),
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if len(result.Added) != 1 || result.Added[0] != "pfsense" {
		t.Fatalf("expected only pfsense to be added, got %v", result.Added)
	}
	if len(result.Existed) != 1 || result.Existed[0] != "sentry" {
		t.Fatalf("expected sentry to be treated as existing, got %v", result.Existed)
	}

	jsonData, err := os.ReadFile(filepath.Join(projectDir, ".mcp.json"))
	if err != nil {
		t.Fatalf("reading .mcp.json: %v", err)
	}
	var rawJSON map[string]json.RawMessage
	if err := json.Unmarshal(jsonData, &rawJSON); err != nil {
		t.Fatalf("parsing .mcp.json: %v", err)
	}
	var jsonServers map[string]json.RawMessage
	if err := json.Unmarshal(rawJSON["mcpServers"], &jsonServers); err != nil {
		t.Fatalf("parsing mcpServers: %v", err)
	}
	if _, ok := jsonServers["sentry"]; ok {
		t.Fatal("did not expect sentry to be backfilled into .mcp.json")
	}
	if _, ok := jsonServers["pfsense"]; !ok {
		t.Fatal("expected pfsense in .mcp.json")
	}

	tomlData, err := os.ReadFile(filepath.Join(projectDir, ".codex", "config.toml"))
	if err != nil {
		t.Fatalf("reading config.toml: %v", err)
	}
	var rawTOML map[string]any
	if _, err := toml.Decode(string(tomlData), &rawTOML); err != nil {
		t.Fatalf("parsing config.toml: %v", err)
	}
	codexServers := rawTOML["mcp_servers"].(map[string]any)
	if _, ok := codexServers["sentry"]; !ok {
		t.Fatal("expected sentry to remain in .codex/config.toml")
	}
	if _, ok := codexServers["pfsense"]; !ok {
		t.Fatal("expected pfsense in .codex/config.toml")
	}
}
