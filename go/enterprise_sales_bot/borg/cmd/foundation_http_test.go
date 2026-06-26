package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	foundationpi "github.com/NexusSoftMDMA/TormentNexus/foundation/pi"
	foundationrepomap "github.com/NexusSoftMDMA/TormentNexus/foundation/repomap"
)

func TestExecuteFoundationToolAndSessions(t *testing.T) {
	cwd := t.TempDir()
	payload, err := executeFoundationTool(cwd, foundationExecRequest{Tool: "write", Input: mustJSON(t, foundationpi.WriteToolInput{Path: "note.txt", Content: "hello"})})
	if err != nil {
		t.Fatal(err)
	}
	if payload["result"] == nil {
		t.Fatalf("expected result payload: %#v", payload)
	}
	session, err := createFoundationSession(cwd, foundationSessionCreateRequest{Name: "alpha"})
	if err != nil {
		t.Fatal(err)
	}
	listed, err := listFoundationSessions(cwd)
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) == 0 {
		t.Fatal("expected session list")
	}
	loaded, err := getFoundationSession(cwd, session.Metadata.SessionID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Metadata.SessionID != session.Metadata.SessionID {
		t.Fatalf("unexpected session load: %#v", loaded)
	}
	forked, err := forkFoundationSession(cwd, session.Metadata.SessionID, foundationSessionForkRequest{Name: "beta"})
	if err != nil {
		t.Fatal(err)
	}
	if forked.Metadata.SessionID == session.Metadata.SessionID {
		t.Fatal("expected new forked session id")
	}
}

func TestFoundationAdaptersPayloadAndRepomap(t *testing.T) {
	cwd := t.TempDir()
	tormentnexusDir := filepath.Join(cwd, "..", "tormentnexus")
	if err := os.MkdirAll(tormentnexusDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tormentnexusDir, "README.md"), []byte("# TormentNexus"), 0o644); err != nil {
		t.Fatal(err)
	}
	payload := foundationAdaptersPayload(cwd)
	if payload["tormentnexus"] == nil || payload["mcp"] == nil {
		t.Fatalf("unexpected adapter payload: %#v", payload)
	}
	setMCPEnv(t, cwd)
	tormentnexusDir2 := filepath.Join(cwd, ".tormentnexus")
	if err := os.MkdirAll(tormentnexusDir2, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tormentnexusDir2, "mcp.json"), []byte(`{"mcpServers":{"demo":{"command":"cmd","args":["/c","echo demo"]}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	mcpTools, err := listFoundationMCPTools(cwd)
	if err != nil {
		t.Fatal(err)
	}
	if len(mcpTools) == 0 {
		t.Fatal("expected MCP tools")
	}
	call, err := callFoundationMCPTool(cwd, foundationMCPCallRequest{Server: "demo", Tool: "list-tools", Arguments: map[string]interface{}{"limit": 1}})
	if err != nil {
		t.Fatal(err)
	}
	if call.Route == "" {
		t.Fatal("expected MCP route")
	}
	providerStatus := providerStatusPayload()
	if providerStatus.CurrentProvider == "" {
		t.Fatal("expected provider status")
	}
	route := selectFoundationProviderRoute(foundationProviderRouteRequest{TaskType: "analysis", CostPreference: "budget"})
	if route.Provider == "" || route.Model == "" {
		t.Fatalf("unexpected provider route: %#v", route)
	}
	execution := prepareFoundationProviderExecution(foundationProviderPrepareRequest{Prompt: "Analyze this repo and explain the architecture.", CostPreference: "budget"})
	if execution.Route.Provider == "" || execution.ExecutionHint == "" {
		t.Fatalf("unexpected provider execution: %#v", execution)
	}
	if err := os.WriteFile(filepath.Join(cwd, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	plan, err := generateFoundationPlan(cwd, foundationPlanRequest{Prompt: "Analyze this repository and explain the architecture", IncludeRepo: true, MaxRepoFiles: 5})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Steps) == 0 {
		t.Fatal("expected plan steps")
	}
	result, err := generateFoundationRepomap(cwd, foundationrepomap.Options{MaxFiles: 5})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Entries) == 0 {
		t.Fatal("expected repomap entries")
	}
	content, err := encodeFoundationReadAsString(cwd, "main.go")
	if err != nil {
		t.Fatal(err)
	}
	if content == "" {
		t.Fatal("expected foundation-backed read content")
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

func mustJSON(t *testing.T, value any) json.RawMessage {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
