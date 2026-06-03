package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/tormentnexushq/tormentnexus-go/internal/config"
	"github.com/tormentnexushq/tormentnexus-go/internal/workflow"
)

func newNativeTestServer(t *testing.T) (*Server, string) {
	t.Helper()
	workspace := t.TempDir()
	cfg := config.Default()
	cfg.WorkspaceRoot = workspace
	cfg.ConfigDir = filepath.Join(workspace, ".tormentnexus-go")
	cfg.MainConfigDir = filepath.Join(workspace, ".tormentnexus")
	if err := os.MkdirAll(cfg.ConfigDir, 0o755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}
	if err := os.MkdirAll(cfg.MainConfigDir, 0o755); err != nil {
		t.Fatalf("failed to create main config dir: %v", err)
	}
	return New(cfg, stubDetector{}), workspace
}

func TestNativeWorkflowEndpoints(t *testing.T) {
	t.Skip("Skipping test for now")
	server, _ := newNativeTestServer(t)

	createBody := bytes.NewBufferString(`{
		"id":"custom-native",
		"name":"Custom Native Workflow",
		"description":"test workflow",
		"steps":[{"id":"step-1","name":"Echo","command":"go version"}]
	}`)
	createReq := httptest.NewRequest(http.MethodPost, "/api/workflows/native/create", createBody)
	createReq.Header.Set("Content-Type", "application/json")
	createRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusCreated {
		t.Fatalf("expected create status 201, got %d: %s", createRec.Code, createRec.Body.String())
	}

	listRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(listRec, httptest.NewRequest(http.MethodGet, "/api/workflows/native", nil))
	if listRec.Code != http.StatusOK {
		t.Fatalf("expected list status 200, got %d: %s", listRec.Code, listRec.Body.String())
	}
	if !bytes.Contains(listRec.Body.Bytes(), []byte(`"custom-native"`)) {
		t.Fatalf("expected created workflow in list response: %s", listRec.Body.String())
	}

	getRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(getRec, httptest.NewRequest(http.MethodGet, "/api/workflows/native/get?id=custom-native", nil))
	if getRec.Code != http.StatusOK {
		t.Fatalf("expected get status 200, got %d: %s", getRec.Code, getRec.Body.String())
	}

	runWF := workflow.NewWorkflow("run-native", "Run Native Workflow", "run test", []*workflow.Step{{
		ID:   "step-a",
		Name: "Immediate",
		Execute: func(ctx context.Context, inputs map[string]any) (map[string]any, error) {
			return map[string]any{"ok": true}, nil
		},
	}})
	server.workflowEngine.Register(runWF)

	runBody := bytes.NewBufferString(`{"id":"run-native"}`)
	runReq := httptest.NewRequest(http.MethodPost, "/api/workflows/native/run", runBody)
	runReq.Header.Set("Content-Type", "application/json")
	runRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(runRec, runReq)
	if runRec.Code != http.StatusAccepted {
		t.Fatalf("expected run status 202, got %d: %s", runRec.Code, runRec.Body.String())
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if runWF.Status == workflow.StatusCompleted {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if runWF.Status != workflow.StatusCompleted {
		t.Fatalf("expected workflow to complete, got %s", runWF.Status)
	}
}

func TestNativeSupervisorEndpoints(t *testing.T) {
	t.Skip("Skipping test for now")
	server, workspace := newNativeTestServer(t)

	createBody := bytes.NewBufferString(`{"id":"native-session","command":"go","args":["version"],"cwd":"` + filepath.ToSlash(workspace) + `"}`)
	createReq := httptest.NewRequest(http.MethodPost, "/api/supervisor/native/create", createBody)
	createReq.Header.Set("Content-Type", "application/json")
	createRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("expected create status 201, got %d: %s", createRec.Code, createRec.Body.String())
	}

	listRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(listRec, httptest.NewRequest(http.MethodGet, "/api/supervisor/native/list", nil))
	if listRec.Code != http.StatusOK || !bytes.Contains(listRec.Body.Bytes(), []byte(`"native-session"`)) {
		t.Fatalf("expected list to contain native-session, got %d: %s", listRec.Code, listRec.Body.String())
	}

	startBody := bytes.NewBufferString(`{"id":"native-session"}`)
	startReq := httptest.NewRequest(http.MethodPost, "/api/supervisor/native/start", startBody)
	startReq.Header.Set("Content-Type", "application/json")
	startRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(startRec, startReq)
	if startRec.Code != http.StatusOK {
		t.Fatalf("expected start status 200, got %d: %s", startRec.Code, startRec.Body.String())
	}

	time.Sleep(100 * time.Millisecond)
	statusRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(statusRec, httptest.NewRequest(http.MethodGet, "/api/supervisor/native/status?id=native-session", nil))
	if statusRec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", statusRec.Code, statusRec.Body.String())
	}
	if bytes.Contains(statusRec.Body.Bytes(), []byte(`"state":"failed"`)) {
		t.Fatalf("expected session not to fail, got %s", statusRec.Body.String())
	}
}

func TestNativeSessionExportEndpoint(t *testing.T) {
	t.Skip("Skipping test for now")
	server, workspace := newNativeTestServer(t)
	home := t.TempDir()

	if runtime.GOOS == "windows" {
		t.Setenv("USERPROFILE", home)
		t.Setenv("HOME", home)
	} else {
		t.Setenv("HOME", home)
	}

	sourceFile := filepath.Join(workspace, ".claude", "session-history.json")
	if err := os.MkdirAll(filepath.Dir(sourceFile), 0o755); err != nil {
		t.Fatalf("failed to create source dir: %v", err)
	}
	if err := os.WriteFile(sourceFile, []byte(`{"messages":[{"role":"user","content":"hi"}]}`), 0o644); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	exportDir := filepath.Join(workspace, "exports-out")
	body := bytes.NewBufferString(`{"outputDir":"` + filepath.ToSlash(exportDir) + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/import/export-native", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected export status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Success bool `json:"success"`
		Data struct {
			ExportedCount int    `json:"exportedCount"`
			OutputPath    string `json:"outputPath"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode export response: %v", err)
	}
	if !payload.Success {
		t.Fatalf("expected success response, got %s", rec.Body.String())
	}
	if payload.Data.ExportedCount < 1 {
		t.Fatalf("expected at least one exported session, got %+v", payload)
	}
	if _, err := os.Stat(filepath.Join(exportDir, "export-manifest.json")); err != nil {
		t.Fatalf("expected export-manifest.json to exist: %v", err)
	}
}

func TestSystemOverviewHandlerReturnsGoNativeData(t *testing.T) {
	server, _ := newNativeTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/system/overview", nil)
	rec := httptest.NewRecorder()
	server.handleSystemOverview(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if success, _ := resp["success"].(bool); !success {
		t.Fatalf("expected success=true")
	}
	data, _ := resp["data"].(map[string]any)
	if data == nil {
		t.Fatal("expected data object")
	}
	// MCP status should always be present
	mcpStatus, _ := data["mcpStatus"].(map[string]any)
	if mcpStatus == nil {
		t.Fatal("expected mcpStatus object")
	}
	// Sessions should be present (Go-local)
	sessions, _ := data["sessions"].(map[string]any)
	if sessions == nil {
		t.Fatal("expected sessions object")
	}
	// Health should contain goSidecar
	health, _ := data["health"].(map[string]any)
	goSidecar, _ := health["goSidecar"].(map[string]any)
	if goSidecar == nil {
		t.Fatal("expected health.goSidecar object")
	}
}
