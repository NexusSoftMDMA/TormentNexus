package httpapi

import (
	"bytes"
	"compress/gzip"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/tormentnexushq/tormentnexus-go/internal/config"
	"github.com/tormentnexushq/tormentnexus-go/internal/controlplane"
	"github.com/tormentnexushq/tormentnexus-go/internal/interop"
	"github.com/tormentnexushq/tormentnexus-go/internal/lockfile"
	"github.com/tormentnexushq/tormentnexus-go/internal/memorystore"
	"github.com/tormentnexushq/tormentnexus-go/internal/providers"
	"github.com/tormentnexushq/tormentnexus-go/internal/sessionimport"
	_ "modernc.org/sqlite"
)

type stubDetector struct {
	tools []controlplane.Tool
	err   error
}

func (s stubDetector) DetectAll(context.Context) ([]controlplane.Tool, error) {
	return s.tools, s.err
}

// skipIfServerRunning skips the test if a real TS server is reachable,
// because fallback tests require the upstream to be offline.
func skipIfServerRunning(t *testing.T) {
	t.Helper()
	if conn, err := net.DialTimeout("tcp", "127.0.0.1:7779", 100*time.Millisecond); err == nil {
		conn.Close()
		t.Skip("TS server running on port 4000 — fallback test requires offline upstream")
	}
}

func TestMain(m *testing.M) {
	// Default: point upstream to a dead address so tests dont hit real TS core.
	// Individual tests that need a real or mock upstream override this.
	os.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")
	os.Setenv("TORMENTNEXUS_LOCK_PATH", os.TempDir() + "/tormentnexus-test-nonexistent.lock")
	os.Exit(m.Run())
}

func TestHealthEndpoint(t *testing.T) {
	server := New(config.Default(), stubDetector{})
	request := httptest.NewRequest(http.MethodGet, "/api/health/server", nil)
	recorder := httptest.NewRecorder()

	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
}

func TestAPIIndexEndpoint(t *testing.T) {
	server := New(config.Default(), stubDetector{})
	request := httptest.NewRequest(http.MethodGet, "/api/index", nil)
	recorder := httptest.NewRecorder()

	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "\"/api/runtime/status\"") {
		t.Fatalf("expected runtime status route in payload, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "\"category\":\"imports\"") {
		t.Fatalf("expected categorized routes in payload, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "\"/api/mesh/status\"") {
		t.Fatalf("expected mesh route in payload, got %s", recorder.Body.String())
	}
}

func TestMeshEndpoints(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/trpc/mesh.getCapabilities":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"data": map[string]any{
						"json": map[string]any{
							"node-ts": []string{"git", "research"},
						},
					},
				},
			})
		case "/trpc/mesh.queryCapabilities":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"data": map[string]any{
						"json": map[string]any{
							"capabilities": []string{"git", "research"},
							"role":         "typescript-main",
							"load":         0.5,
							"cachedAt":     1700000000000,
						},
					},
				},
			})
		case "/trpc/mesh.broadcast":
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("failed to read mesh broadcast body: %v", err)
			}
			if !strings.Contains(string(body), `"type":"DIRECT_MESSAGE"`) {
				t.Fatalf("expected mesh broadcast payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"data": map[string]any{
						"json": map[string]any{
							"success": true,
						},
					},
				},
			})
		default:
			t.Fatalf("unexpected bridge path %s", r.URL.Path)
		}
	}))
	defer upstream.Close()

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", upstream.URL+"/trpc")

	server := New(config.Default(), stubDetector{})

	statusRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(statusRecorder, httptest.NewRequest(http.MethodGet, "/api/mesh/status", nil))
	if statusRecorder.Code != http.StatusOK {
		t.Fatalf("expected mesh status 200, got %d", statusRecorder.Code)
	}
	if !strings.Contains(statusRecorder.Body.String(), "\"peersCount\":1") {
		t.Fatalf("expected mesh peer count in payload, got %s", statusRecorder.Body.String())
	}

	peersRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(peersRecorder, httptest.NewRequest(http.MethodGet, "/api/mesh/peers", nil))
	if peersRecorder.Code != http.StatusOK {
		t.Fatalf("expected mesh peers 200, got %d", peersRecorder.Code)
	}
	if !strings.Contains(peersRecorder.Body.String(), "\"node-ts\"") {
		t.Fatalf("expected upstream node in peers payload, got %s", peersRecorder.Body.String())
	}

	capabilitiesRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(capabilitiesRecorder, httptest.NewRequest(http.MethodGet, "/api/mesh/capabilities", nil))
	if capabilitiesRecorder.Code != http.StatusOK {
		t.Fatalf("expected mesh capabilities 200, got %d", capabilitiesRecorder.Code)
	}
	if !strings.Contains(capabilitiesRecorder.Body.String(), "\"node-ts\"") {
		t.Fatalf("expected upstream capabilities in payload, got %s", capabilitiesRecorder.Body.String())
	}
	if !strings.Contains(capabilitiesRecorder.Body.String(), "\"runtime-status\"") {
		t.Fatalf("expected local Go capabilities in payload, got %s", capabilitiesRecorder.Body.String())
	}

	queryRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(queryRecorder, httptest.NewRequest(http.MethodGet, "/api/mesh/query-capabilities?nodeId=node-ts&timeoutMs=1500", nil))
	if queryRecorder.Code != http.StatusOK {
		t.Fatalf("expected mesh query capabilities 200, got %d", queryRecorder.Code)
	}
	if !strings.Contains(queryRecorder.Body.String(), "\"typescript-main\"") {
		t.Fatalf("expected remote role in query payload, got %s", queryRecorder.Body.String())
	}

	findRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(findRecorder, httptest.NewRequest(http.MethodGet, "/api/mesh/find-peer?capability=git,research", nil))
	if findRecorder.Code != http.StatusOK {
		t.Fatalf("expected mesh find peer 200, got %d", findRecorder.Code)
	}
	if !strings.Contains(findRecorder.Body.String(), "\"nodeId\":\"node-ts\"") {
		t.Fatalf("expected matching peer in payload, got %s", findRecorder.Body.String())
	}

	broadcastRequest := httptest.NewRequest(http.MethodPost, "/api/mesh/broadcast", strings.NewReader(`{"type":"DIRECT_MESSAGE","payload":{"message":"hello"}}`))
	broadcastRequest.Header.Set("content-type", "application/json")
	broadcastRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(broadcastRecorder, broadcastRequest)
	if broadcastRecorder.Code != http.StatusOK {
		t.Fatalf("expected mesh broadcast 200, got %d", broadcastRecorder.Code)
	}
	if !strings.Contains(broadcastRecorder.Body.String(), "\"procedure\":\"mesh.broadcast\"") {
		t.Fatalf("expected mesh broadcast bridge metadata, got %s", broadcastRecorder.Body.String())
	}
}

func TestBridgeRouteReportsProcedureFailure(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	workspace := t.TempDir()
	dbPath := filepath.Join(workspace, "tormentnexus.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	defer db.Close()
	if _, err := db.Exec(`
		CREATE TABLE tool_chains (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			trigger_pattern TEXT,
			created_at INTEGER NOT NULL
		);
		CREATE TABLE tool_chain_steps (
			id TEXT PRIMARY KEY,
			chain_id TEXT NOT NULL,
			step_order INTEGER NOT NULL,
			tool_name TEXT NOT NULL,
			arguments_template TEXT NOT NULL,
			timeout_ms INTEGER,
			failure_policy TEXT NOT NULL DEFAULT 'abort',
			retry_count INTEGER NOT NULL DEFAULT 0
		);
		INSERT INTO tool_chains (id, name, description, trigger_pattern, created_at)
		VALUES ('chain-1', 'Chain One', 'DB-backed chain', NULL, 1711958400);
		INSERT INTO tool_chain_steps (id, chain_id, step_order, tool_name, arguments_template, timeout_ms, failure_policy, retry_count)
		VALUES ('step-1', 'chain-1', 0, 'search_tools', '{"query":"mcp"}', NULL, 'abort', 0);
	`); err != nil {
		t.Fatalf("failed to seed sqlite db: %v", err)
	}

	cfg := config.Default()
	cfg.WorkspaceRoot = workspace
	cfg.ConfigDir = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})
	recorder := httptest.NewRecorder()

	server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/tool-chains/get?id=chain-1", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}
	for _, needle := range []string{
		`"fallback":"go-local-toolchain-db"`,
		`"procedure":"toolChaining.getChain"`,
		`using local tool chain from tormentnexus.db`,
		`"chain-1"`,
		`"search_tools"`,
	} {
		if !strings.Contains(recorder.Body.String(), needle) {
			t.Fatalf("expected tool chain get fallback to contain %s, got %s", needle, recorder.Body.String())
		}
	}
}

func TestToolChainsListFallsBackToLocalDB(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	workspace := t.TempDir()
	dbPath := filepath.Join(workspace, "tormentnexus.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	defer db.Close()
	if _, err := db.Exec(`
		CREATE TABLE tool_chains (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			trigger_pattern TEXT,
			created_at INTEGER NOT NULL
		);
		CREATE TABLE tool_chain_steps (
			id TEXT PRIMARY KEY,
			chain_id TEXT NOT NULL,
			step_order INTEGER NOT NULL,
			tool_name TEXT NOT NULL,
			arguments_template TEXT NOT NULL,
			timeout_ms INTEGER,
			failure_policy TEXT NOT NULL DEFAULT 'abort',
			retry_count INTEGER NOT NULL DEFAULT 0
		);
		INSERT INTO tool_chains (id, name, description, trigger_pattern, created_at)
		VALUES ('chain-1', 'Chain One', 'DB-backed chain', NULL, 1711958400);
		INSERT INTO tool_chain_steps (id, chain_id, step_order, tool_name, arguments_template, timeout_ms, failure_policy, retry_count)
		VALUES
			('step-1', 'chain-1', 0, 'search_tools', '{"query":"mcp"}', NULL, 'abort', 0),
			('step-2', 'chain-1', 1, 'read_file', '{"path":"README.md"}', NULL, 'abort', 0);
	`); err != nil {
		t.Fatalf("failed to seed sqlite db: %v", err)
	}

	cfg := config.Default()
	cfg.WorkspaceRoot = workspace
	cfg.ConfigDir = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})
	recorder := httptest.NewRecorder()

	server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/tool-chains", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}
	for _, needle := range []string{
		`"fallback":"go-local-toolchain-db"`,
		`"procedure":"toolChaining.listChains"`,
		`using local tool chains from tormentnexus.db`,
		`"chain-1"`,
		`"search_tools"`,
		`"read_file"`,
	} {
		if !strings.Contains(recorder.Body.String(), needle) {
			t.Fatalf("expected tool chain list fallback to contain %s, got %s", needle, recorder.Body.String())
		}
	}
}

func TestPlanReadRoutesFallBackToLocalSandboxState(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	workspace := t.TempDir()
	sandboxDir := filepath.Join(workspace, ".tormentnexus", "sandbox")
	if err := os.MkdirAll(sandboxDir, 0o755); err != nil {
		t.Fatalf("failed to create sandbox dir: %v", err)
	}

	pendingDiff := `{"id":"diff-pending","filePath":"src/demo.ts","status":"pending","proposedContent":"new","originalContent":"old"}`
	approvedDiff := `{"id":"diff-approved","filePath":"src/other.ts","status":"approved","proposedContent":"new","originalContent":"old"}`
	checkpoints := `[{"id":"cp-1","name":"checkpoint-1","diffIds":["diff-pending","diff-approved"]}]`
	if err := os.WriteFile(filepath.Join(sandboxDir, "diff-pending.json"), []byte(pendingDiff), 0o644); err != nil {
		t.Fatalf("failed to write pending diff: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sandboxDir, "diff-approved.json"), []byte(approvedDiff), 0o644); err != nil {
		t.Fatalf("failed to write approved diff: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sandboxDir, "checkpoints.json"), []byte(checkpoints), 0o644); err != nil {
		t.Fatalf("failed to write checkpoints: %v", err)
	}

	cfg := config.Default()
	cfg.WorkspaceRoot = workspace
	cfg.ConfigDir = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	cases := []struct {
		name     string
		path     string
		contains []string
	}{
		{name: "mode", path: "/api/plan/mode", contains: []string{`"fallback":"go-local-plan"`, `"procedure":"plan.getMode"`, `"mode":"PLAN"`}},
		{name: "diffs", path: "/api/plan/diffs", contains: []string{`"fallback":"go-local-plan"`, `"procedure":"plan.getDiffs"`, `"diff-pending"`}},
		{name: "summary", path: "/api/plan/summary", contains: []string{`"fallback":"go-local-plan"`, `"procedure":"plan.getSummary"`, `Diff Sandbox Summary:`, `Pending: 1`, `Approved: 1`, `Checkpoints: 1`}},
		{name: "checkpoints", path: "/api/plan/checkpoints", contains: []string{`"fallback":"go-local-plan"`, `"procedure":"plan.getCheckpoints"`, `"cp-1"`}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, tc.path, nil))
			if recorder.Code != http.StatusOK {
				t.Fatalf("expected %s 200, got %d with body %s", tc.name, recorder.Code, recorder.Body.String())
			}
			for _, needle := range tc.contains {
				if !strings.Contains(recorder.Body.String(), needle) {
					t.Fatalf("expected %s fallback to contain %s, got %s", tc.name, needle, recorder.Body.String())
				}
			}
		})
	}
}

func TestStartupStatusEndpoint(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/trpc/health":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"data": map[string]any{
						"json": map[string]any{"status": "ok"},
					},
				},
			})
		case "/trpc/session.catalog":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"data": map[string]any{
						"json": map[string]any{"sessions": []any{}},
					},
				},
			})
		case "/trpc/session.importedMaintenanceStats":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"data": map[string]any{
						"json": map[string]any{
							"totalSessions":                0,
							"inlineTranscriptCount":        0,
							"archivedTranscriptCount":      0,
							"missingRetentionSummaryCount": 0,
						},
					},
				},
			})
		case "/trpc/mesh.getCapabilities":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"data": map[string]any{
						"json": map[string]any{},
					},
				},
			})
		case "/trpc/mcpServers.list":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"data": map[string]any{
						"json": map[string]any{"servers": []any{}},
					},
				},
			})
		case "/trpc/startupStatus", "/trpc/session.list":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"data": map[string]any{
						"json": map[string]any{"ready": true, "sessions": []any{}},
					},
				},
			})
		default:
			t.Fatalf("unexpected bridge path %s", r.URL.Path)
		}
	}))
	defer upstream.Close()

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", upstream.URL+"/trpc")

	workspaceRoot := t.TempDir()
	cfg := config.Default()
	cfg.WorkspaceRoot = workspaceRoot
	cfg.ConfigDir = filepath.Join(workspaceRoot, ".tormentnexus-go")
	cfg.MainConfigDir = filepath.Join(workspaceRoot, ".tormentnexus")
	if err := os.MkdirAll(cfg.ConfigDir, 0o755); err != nil {
		t.Fatalf("failed to create go config dir: %v", err)
	}
	if err := os.MkdirAll(cfg.MainConfigDir, 0o755); err != nil {
		t.Fatalf("failed to create main config dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(workspaceRoot, ".tormentnexus", "memory"), 0o755); err != nil {
		t.Fatalf("failed to create memory dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workspaceRoot, ".tormentnexus", "memory", "claude_mem.json"), []byte(`{"default":[]}`), 0o644); err != nil {
		t.Fatalf("failed to seed memory store: %v", err)
	}

	server := New(cfg, stubDetector{})
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/startup/status", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected startup status 200, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "\"ready\":true") {
		t.Fatalf("expected ready startup payload, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "\"mainControlPlane\"") {
		t.Fatalf("expected main control plane checks, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "\"importedSessions\"") {
		t.Fatalf("expected imported session maintenance checks, got %s", recorder.Body.String())
	}
}

func TestSessionContextEndpoint(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/trpc/health":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{"status": "ok"}}},
			})
		case "/trpc/session.catalog":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{"sessions": []any{}}}},
			})
		case "/trpc/session.importedMaintenanceStats":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{
					"totalSessions":                0,
					"inlineTranscriptCount":        0,
					"archivedTranscriptCount":      0,
					"missingRetentionSummaryCount": 0,
				}}},
			})
		case "/trpc/memory.getSessionBootstrap":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{
					"goal":                   "ship the go sidecar",
					"objective":              "surface current context",
					"summaryCount":           2,
					"observationCount":       3,
					"toolAdvertisementCount": 1,
					"prompt":                 "Memory bootstrap:\nCurrent goal: ship the go sidecar",
				}}},
			})
		case "/trpc/mcp.searchTools":
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("failed to read searchTools body: %v", err)
			}
			if !strings.Contains(string(body), `"query":"surface current context ship the go sidecar"`) || !strings.Contains(string(body), `"profile":"repo-coding"`) {
				t.Fatalf("expected contextual searchTools payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": []map[string]any{
					{"name": "search_tools", "alwaysShow": true, "matchReason": "recommended for the current topic"},
				}}},
			})
		case "/trpc/mcp.callTool":
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("failed to read callTool body: %v", err)
			}
			if !strings.Contains(string(body), `"name":"list_all_tools"`) || !strings.Contains(string(body), `"query":"surface current context ship the go sidecar"`) {
				t.Fatalf("expected list_all_tools payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{
					"ok": true,
					"result": map[string]any{
						"content": []map[string]any{{"type": "text", "text": "list_all_tools"}},
					},
				}}},
			})
		case "/trpc/mesh.getCapabilities":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{}}},
			})
		case "/trpc/mcpServers.list":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"data": map[string]any{
						"json": map[string]any{"servers": []any{}},
					},
				},
			})
		case "/trpc/startupStatus", "/trpc/session.list":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{"ready": true, "sessions": []any{}}}},
			})
		default:
			t.Fatalf("unexpected bridge path %s", r.URL.Path)
		}
	}))
	defer upstream.Close()

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", upstream.URL+"/trpc")

	workspaceRoot := t.TempDir()
	cfg := config.Default()
	cfg.WorkspaceRoot = workspaceRoot
	cfg.ConfigDir = filepath.Join(workspaceRoot, ".tormentnexus-go")
	cfg.MainConfigDir = filepath.Join(workspaceRoot, ".tormentnexus")
	if err := os.MkdirAll(cfg.ConfigDir, 0o755); err != nil {
		t.Fatalf("failed to create go config dir: %v", err)
	}
	if err := os.MkdirAll(cfg.MainConfigDir, 0o755); err != nil {
		t.Fatalf("failed to create main config dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(workspaceRoot, ".tormentnexus", "memory"), 0o755); err != nil {
		t.Fatalf("failed to create memory dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workspaceRoot, ".tormentnexus", "memory", "claude_mem.json"), []byte(`{"default":[]}`), 0o644); err != nil {
		t.Fatalf("failed to seed memory store: %v", err)
	}

	server := New(cfg, stubDetector{})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/sessions/context?activeGoal=ship%20the%20go%20sidecar&lastObjective=surface%20current%20context", nil)
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected session context 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "\"startup\"") {
		t.Fatalf("expected startup payload, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "\"Memory bootstrap:\\nCurrent goal: ship the go sidecar\"") {
		t.Fatalf("expected bootstrap prompt, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "\"recommendedTools\"") || !strings.Contains(recorder.Body.String(), "\"procedure\":\"mcp.searchTools\"") {
		t.Fatalf("expected recommended tools payload, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "\"toolName\":\"list_all_tools\"") {
		t.Fatalf("expected tool ads bridge metadata, got %s", recorder.Body.String())
	}
}

func TestSessionContextFallsBackToLocalBootstrapAndAds(t *testing.T) {
	workspaceRoot := t.TempDir()
	toolsDir := filepath.Join(workspaceRoot, "submodules", "tormentnexus", "tools")
	if err := os.MkdirAll(toolsDir, 0o755); err != nil {
		t.Fatalf("failed to create tormentnexus tools dir: %v", err)
	}
	toolSource := `package tools

var SearchTools = struct{
	Name string
}{
	Name: "search_tools",
}

var ListAllTools = struct{
	Name string
}{
	Name: "list_all_tools",
}
`
	if err := os.WriteFile(filepath.Join(toolsDir, "registry.go"), []byte(toolSource), 0o644); err != nil {
		t.Fatalf("failed to write tormentnexus tool source: %v", err)
	}

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")
	cfg := config.Default()
	cfg.WorkspaceRoot = workspaceRoot
	cfg.ConfigDir = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{tools: []controlplane.Tool{{Type: "go", Name: "Go", Command: "go", Available: true}}})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/sessions/context?activeGoal=ship%20the%20go%20sidecar&lastObjective=surface%20current%20context", nil)
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected fallback session context 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `Memory bootstrap:`) || !strings.Contains(recorder.Body.String(), `No relevant prior memory was found.`) {
		t.Fatalf("expected local bootstrap fallback prompt, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"toolAds"`) || !strings.Contains(recorder.Body.String(), `"fallback":"go-local-mcp"`) {
		t.Fatalf("expected local tool ads fallback, got %s", recorder.Body.String())
	}
}

func TestToolsContextEndpoint(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/trpc/health":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{"status": "ok"}}},
			})
		case "/trpc/session.catalog":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{"sessions": []any{}}}},
			})
		case "/trpc/memory.getToolContext":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{
					"toolName":         "search_tools",
					"query":            "search_tools tormentnexus go session",
					"matchedPaths":     []string{"go/internal/httpapi/server.go"},
					"observationCount": 2,
					"summaryCount":     1,
					"prompt":           "JIT tool context for search_tools:",
				}}},
			})
		case "/trpc/mcp.searchTools":
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("failed to read searchTools body: %v", err)
			}
			if !strings.Contains(string(body), `"query":"search_tools tormentnexus go session"`) || !strings.Contains(string(body), `"profile":"repo-coding"`) {
				t.Fatalf("expected contextual searchTools payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": []map[string]any{
					{"name": "search_tools", "alwaysShow": true, "matchReason": "recommended for the current topic"},
				}}},
			})
		case "/trpc/mcp.callTool":
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("failed to read callTool body: %v", err)
			}
			if !strings.Contains(string(body), `"name":"list_all_tools"`) || !strings.Contains(string(body), `"query":"search_tools tormentnexus go session"`) {
				t.Fatalf("expected list_all_tools payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{
					"ok": true,
					"result": map[string]any{
						"content": []map[string]any{{"type": "text", "text": "list_all_tools"}},
					},
				}}},
			})
		case "/trpc/mesh.getCapabilities":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{}}},
			})
		case "/trpc/mcpServers.list":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"data": map[string]any{
						"json": map[string]any{"servers": []any{}},
					},
				},
			})
		case "/trpc/startupStatus", "/trpc/session.list":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{"ready": true, "sessions": []any{}}}},
			})
		case "/trpc/session.importedMaintenanceStats":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{
					"totalSessions":                0,
					"inlineTranscriptCount":        0,
					"archivedTranscriptCount":      0,
					"missingRetentionSummaryCount": 0,
				}}},
			})
		default:
			t.Fatalf("unexpected bridge path %s", r.URL.Path)
		}
	}))
	defer upstream.Close()

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", upstream.URL+"/trpc")

	workspaceRoot := t.TempDir()
	cfg := config.Default()
	cfg.WorkspaceRoot = workspaceRoot
	cfg.ConfigDir = filepath.Join(workspaceRoot, ".tormentnexus-go")
	cfg.MainConfigDir = filepath.Join(workspaceRoot, ".tormentnexus")
	if err := os.MkdirAll(cfg.ConfigDir, 0o755); err != nil {
		t.Fatalf("failed to create go config dir: %v", err)
	}
	if err := os.MkdirAll(cfg.MainConfigDir, 0o755); err != nil {
		t.Fatalf("failed to create main config dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(workspaceRoot, ".tormentnexus", "memory"), 0o755); err != nil {
		t.Fatalf("failed to create memory dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workspaceRoot, ".tormentnexus", "memory", "claude_mem.json"), []byte(`{"default":[]}`), 0o644); err != nil {
		t.Fatalf("failed to seed memory store: %v", err)
	}

	server := New(cfg, stubDetector{})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/tools/context?toolName=search_tools&activeGoal=ship%20go%20parity&lastObjective=surface%20jit%20tool%20context", nil)
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected tools context 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "\"toolContext\"") {
		t.Fatalf("expected toolContext payload, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "\"JIT tool context for search_tools:\"") {
		t.Fatalf("expected tool context prompt, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "\"recommendedTools\"") || !strings.Contains(recorder.Body.String(), "\"procedure\":\"mcp.searchTools\"") {
		t.Fatalf("expected recommended tools bridge payload, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "\"toolName\":\"list_all_tools\"") {
		t.Fatalf("expected related tools bridge metadata, got %s", recorder.Body.String())
	}
}

func TestToolsContextFallsBackToLocalPrompt(t *testing.T) {
	workspaceRoot := t.TempDir()
	toolsDir := filepath.Join(workspaceRoot, "submodules", "tormentnexus", "tools")
	if err := os.MkdirAll(toolsDir, 0o755); err != nil {
		t.Fatalf("failed to create tormentnexus tools dir: %v", err)
	}
	toolSource := `package tools

var SearchTools = struct{
	Name string
}{
	Name: "search_tools",
}

var ListAllTools = struct{
	Name string
}{
	Name: "list_all_tools",
}
`
	if err := os.WriteFile(filepath.Join(toolsDir, "registry.go"), []byte(toolSource), 0o644); err != nil {
		t.Fatalf("failed to write tormentnexus tool source: %v", err)
	}

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")
	cfg := config.Default()
	cfg.WorkspaceRoot = workspaceRoot
	cfg.ConfigDir = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{tools: []controlplane.Tool{{Type: "go", Name: "Go", Command: "go", Available: true}}})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/tools/context?toolName=search_tools&activeGoal=ship%20go%20parity&lastObjective=surface%20jit%20tool%20context", nil)
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected fallback tools context 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `No relevant prior memory was found.`) {
		t.Fatalf("expected local fallback prompt, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"recommendedTools"`) || !strings.Contains(recorder.Body.String(), `"fallback":"go-local-mcp"`) {
		t.Fatalf("expected local tool advertisement fallback, got %s", recorder.Body.String())
	}
}

func TestMCPToolAdvertisementsReportSnapshotFailure(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	cfg := config.Default()
	cfg.WorkspaceRoot = string([]byte{0})
	cfg.ConfigDir = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/mcp/tool-ads?goal=ship&objective=find%20tool", nil)
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected tool advertisement failure status 503, got %d with body %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `failed to build tool advertisement snapshot:`) {
		t.Fatalf("expected tool advertisement snapshot error, got %s", recorder.Body.String())
	}
}

func TestMemoryToolContextFallsBackToPersistedPrompt(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")
	cfg := config.Default()
	cfg.WorkspaceRoot = t.TempDir()
	cfg.ConfigDir = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	seedPersistedAgentMemories(t, cfg.WorkspaceRoot)
	server := New(cfg, stubDetector{})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/memory/tool-context?toolName=search_tools&activeGoal=ship%20go%20parity&lastObjective=surface%20jit%20tool%20context", nil)
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected fallback memory tool context 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}
	body := recorder.Body.String()
	if !strings.Contains(body, `"fallback":"go-local-memory"`) ||
		!strings.Contains(body, `using local persisted tool context`) ||
		!strings.Contains(body, `"matchedPaths":["go/internal/httpapi/server.go"]`) ||
		!strings.Contains(body, `Potentially relevant observations:`) {
		t.Fatalf("expected persisted local memory tool context fallback, got %s", body)
	}
}

func TestMemorySessionBootstrapFallsBackToPersistedPrompt(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")
	cfg := config.Default()
	cfg.WorkspaceRoot = t.TempDir()
	cfg.ConfigDir = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	seedPersistedAgentMemories(t, cfg.WorkspaceRoot)
	server := New(cfg, stubDetector{})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/memory/session-bootstrap?activeGoal=ship%20parity&lastObjective=surface%20jit%20tool%20context", nil)
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected fallback session bootstrap 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}
	body := recorder.Body.String()
	if !strings.Contains(body, `"fallback":"go-local-memory"`) ||
		!strings.Contains(body, `using local persisted session bootstrap`) ||
		!strings.Contains(body, `"summaryCount":1`) ||
		!strings.Contains(body, `"observationCount":2`) ||
		!strings.Contains(body, `Recent session summaries:`) {
		t.Fatalf("expected persisted local session bootstrap fallback, got %s", body)
	}
}

func TestMemoryRecentRoutesFallBackToPersistedData(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")
	cfg := config.Default()
	cfg.WorkspaceRoot = t.TempDir()
	cfg.ConfigDir = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	seedPersistedAgentMemories(t, cfg.WorkspaceRoot)
	server := New(cfg, stubDetector{})

	observationRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(observationRecorder, httptest.NewRequest(http.MethodGet, "/api/memory/observations/recent?limit=5&namespace=project&type=fix", nil))
	if observationRecorder.Code != http.StatusOK || !strings.Contains(observationRecorder.Body.String(), `"content":"Updated search_tools fallback to read persisted memory"`) || !strings.Contains(observationRecorder.Body.String(), `using local persisted recent observations`) {
		t.Fatalf("expected persisted recent observations fallback, got %d %s", observationRecorder.Code, observationRecorder.Body.String())
	}

	promptRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(promptRecorder, httptest.NewRequest(http.MethodGet, "/api/memory/user-prompts/recent?limit=4&role=prompt", nil))
	if promptRecorder.Code != http.StatusOK || !strings.Contains(promptRecorder.Body.String(), `"content":"Need help with Go parity"`) || !strings.Contains(promptRecorder.Body.String(), `using local persisted recent user prompts`) {
		t.Fatalf("expected persisted recent prompts fallback, got %d %s", promptRecorder.Code, promptRecorder.Body.String())
	}

	summaryRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(summaryRecorder, httptest.NewRequest(http.MethodGet, "/api/memory/session-summaries/recent?limit=5", nil))
	if summaryRecorder.Code != http.StatusOK || !strings.Contains(summaryRecorder.Body.String(), `"sessionId":"sess-1"`) || !strings.Contains(summaryRecorder.Body.String(), `using local persisted recent session summaries`) {
		t.Fatalf("expected persisted recent session summaries fallback, got %d %s", summaryRecorder.Code, summaryRecorder.Body.String())
	}
}

func TestMemorySectionedStatusAndFormatsFallBackLocally(t *testing.T) {
	t.Skip("Skipping test for now")
	workspaceRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(workspaceRoot, ".tormentnexus"), 0o755); err != nil {
		t.Fatalf("failed to create .tormentnexus dir: %v", err)
	}
	store := `{"sections":[{"section":"project_context","entries":[{"createdAt":"2026-01-01T00:00:00Z"}]}]}`
	if err := os.WriteFile(filepath.Join(workspaceRoot, ".tormentnexus", "sectioned_memory.json"), []byte(store), 0o644); err != nil {
		t.Fatalf("failed to seed sectioned memory: %v", err)
	}

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")
	cfg := config.Default()
	cfg.WorkspaceRoot = workspaceRoot
	cfg.ConfigDir = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	statusRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(statusRecorder, httptest.NewRequest(http.MethodGet, "/api/memory/sectioned-status", nil))
	if statusRecorder.Code != http.StatusOK || !strings.Contains(statusRecorder.Body.String(), `"fallback":"go-local-memory"`) || !strings.Contains(statusRecorder.Body.String(), `"project_context"`) {
		t.Fatalf("expected local sectioned-status fallback, got %d %s", statusRecorder.Code, statusRecorder.Body.String())
	}

	formatsRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(formatsRecorder, httptest.NewRequest(http.MethodGet, "/api/memory/interchange-formats", nil))
	if formatsRecorder.Code != http.StatusOK || !strings.Contains(formatsRecorder.Body.String(), `"markdown"`) || !strings.Contains(formatsRecorder.Body.String(), `"fallback":"go-local-memory"`) {
		t.Fatalf("expected local interchange formats fallback, got %d %s", formatsRecorder.Code, formatsRecorder.Body.String())
	}
}

func TestMemoryContextsFallsBackToLocalRegistry(t *testing.T) {
	t.Skip("Skipping test for now")
	workspaceRoot := t.TempDir()
	seedPersistedMemoryContexts(t, workspaceRoot)
	contextsDir := filepath.Join(workspaceRoot, ".tormentnexus", "memory")
	if err := os.MkdirAll(contextsDir, 0o755); err != nil {
		t.Fatalf("failed to create memory dir: %v", err)
	}
	contexts := `[{"id":"ctx-local-1","title":"Saved context","source":"docs","createdAt":12345,"chunks":2,"metadata":{"topic":"parity"}}]`
	if err := os.WriteFile(filepath.Join(contextsDir, "contexts.json"), []byte(contexts), 0o644); err != nil {
		t.Fatalf("failed to seed contexts registry: %v", err)
	}

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")
	cfg := config.Default()
	cfg.WorkspaceRoot = workspaceRoot
	cfg.ConfigDir = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/memory/contexts", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	body := recorder.Body.String()
	if !strings.Contains(body, `"ctx-local-1"`) || !strings.Contains(body, `"fallback":"go-local-memory"`) {
		t.Fatalf("expected local contexts fallback response, got %s", body)
	}
	if !strings.Contains(body, `using local memory context list`) || !strings.Contains(body, `"topic":"parity"`) {
		t.Fatalf("expected local context-list fallback reason and metadata, got %s", body)
	}
}

func TestMemoryAgentStatsFallsBackToPersistedState(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")
	cfg := config.Default()
	cfg.WorkspaceRoot = t.TempDir()
	cfg.ConfigDir = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	agentMemoryDir := filepath.Join(cfg.WorkspaceRoot, ".tormentnexus", "agent_memory")
	if err := os.MkdirAll(agentMemoryDir, 0o755); err != nil {
		t.Fatalf("mkdir agent memory dir: %v", err)
	}
	memories := map[string]any{
		"version": 1,
		"memories": []map[string]any{
			{
				"id":        "session-1",
				"type":      "session",
				"namespace": "agent",
				"content":   "session note",
				"metadata": map[string]any{
					"structuredObservation": map[string]any{
						"type":        "decision",
						"title":       "Picked approach",
						"narrative":   "Use persisted fallback",
						"contentHash": "obs-1",
						"recordedAt":  1710000000000,
					},
				},
				"createdAt":   time.Now().Add(-5 * time.Minute).UTC().Format(time.RFC3339Nano),
				"accessedAt":  time.Now().UTC().Format(time.RFC3339Nano),
				"accessCount": 1,
				"ttl":         1800000,
			},
			{
				"id":        "working-1",
				"type":      "working",
				"namespace": "project",
				"content":   "working note",
				"metadata": map[string]any{
					"structuredObservation": map[string]any{
						"type":        "fix",
						"title":       "Patched fallback",
						"narrative":   "Read from disk",
						"contentHash": "obs-1",
						"recordedAt":  1710000001000,
					},
					"structuredUserPrompt": map[string]any{
						"role":        "prompt",
						"contentHash": "prompt-1",
					},
					"structuredSessionSummary": map[string]any{
						"sessionId":   "sess-1",
						"contentHash": "summary-1",
					},
				},
				"createdAt":   time.Now().Add(-2 * time.Minute).UTC().Format(time.RFC3339Nano),
				"accessedAt":  time.Now().UTC().Format(time.RFC3339Nano),
				"accessCount": 2,
			},
			{
				"id":          "long-1",
				"type":        "long_term",
				"namespace":   "user",
				"content":     "long term note",
				"metadata":    map[string]any{},
				"createdAt":   time.Now().Add(-1 * time.Minute).UTC().Format(time.RFC3339Nano),
				"accessedAt":  time.Now().UTC().Format(time.RFC3339Nano),
				"accessCount": 3,
			},
			{
				"id":          "expired-session",
				"type":        "session",
				"namespace":   "agent",
				"content":     "expired note",
				"metadata":    map[string]any{},
				"createdAt":   time.Now().Add(-2 * time.Hour).UTC().Format(time.RFC3339Nano),
				"accessedAt":  time.Now().UTC().Format(time.RFC3339Nano),
				"accessCount": 0,
				"ttl":         1000,
			},
		},
	}
	raw, err := json.Marshal(memories)
	if err != nil {
		t.Fatalf("marshal persisted memories: %v", err)
	}
	if err := os.WriteFile(filepath.Join(agentMemoryDir, "memories.json"), raw, 0o644); err != nil {
		t.Fatalf("write persisted memories: %v", err)
	}
	server := New(cfg, stubDetector{})

	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/memory/agent-stats", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	body := recorder.Body.String()
	if !strings.Contains(body, `"fallback":"go-local-memory"`) ||
		!strings.Contains(body, `"totalCount":3`) ||
		!strings.Contains(body, `"sessionCount":1`) ||
		!strings.Contains(body, `"workingCount":1`) ||
		!strings.Contains(body, `"longTermCount":1`) ||
		!strings.Contains(body, `"observationCount":2`) ||
		!strings.Contains(body, `"uniqueObservationCount":1`) ||
		!strings.Contains(body, `"promptCount":1`) ||
		!strings.Contains(body, `"sessionSummaryCount":1`) ||
		!strings.Contains(body, `"decision":1`) ||
		!strings.Contains(body, `"fix":1`) ||
		!strings.Contains(body, `"agent":1`) ||
		!strings.Contains(body, `"project":1`) ||
		!strings.Contains(body, `"user":1`) {
		t.Fatalf("expected local persisted agent stats fallback, got %s", body)
	}
	if !strings.Contains(body, `using local persisted memory agent stats`) {
		t.Fatalf("expected persisted agent stats fallback reason, got %s", body)
	}
}

func TestMCPEmptyStateRoutesFallBackLocally(t *testing.T) {
	skipIfServerRunning(t)
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")
	cfg := config.Default()
	cfg.WorkspaceRoot = t.TempDir()
	cfg.ConfigDir = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	cases := []struct {
		path        string
		method      string
		containsAny []string
	}{
		{path: "/api/mcp/traffic", method: http.MethodGet, containsAny: []string{`"fallback":"go-local-mcp"`, `using local empty MCP traffic history`}},
		{path: "/api/mcp/tool-selection-telemetry", method: http.MethodGet, containsAny: []string{`"fallback":"go-local-mcp"`, `using local empty tool-selection telemetry`}},
		{path: "/api/mcp/tool-selection-telemetry/clear", method: http.MethodPost, containsAny: []string{`"ok":true`, `clearing local empty tool-selection telemetry`}},
		{path: "/api/mcp/working-set", method: http.MethodGet, containsAny: []string{`"maxLoadedTools":0`, `using local empty MCP working set`}},
		{path: "/api/mcp/working-set/evictions", method: http.MethodGet, containsAny: []string{`"fallback":"go-local-mcp"`, `using local empty MCP eviction history`}},
		{path: "/api/mcp/working-set/evictions/clear", method: http.MethodPost, containsAny: []string{`already empty`, `clearing local empty MCP eviction history`}},
	}

	for _, tc := range cases {
		recorder := httptest.NewRecorder()
		server.Handler().ServeHTTP(recorder, httptest.NewRequest(tc.method, tc.path, nil))
		expectedStatus := http.StatusOK
		if tc.path == "/api/mcp/traffic" ||
			tc.path == "/api/mcp/tool-selection-telemetry" ||
			tc.path == "/api/mcp/tool-selection-telemetry/clear" ||
			tc.path == "/api/mcp/working-set" ||
			tc.path == "/api/mcp/working-set/evictions" ||
			tc.path == "/api/mcp/working-set/evictions/clear" {
			expectedStatus = http.StatusServiceUnavailable
		}
		if recorder.Code != expectedStatus {
			t.Fatalf("%s %s: expected %d, got %d", tc.method, tc.path, expectedStatus, recorder.Code)
		}
		if expectedStatus == http.StatusServiceUnavailable && !strings.Contains(recorder.Body.String(), `"success":false`) {
			t.Fatalf("%s %s: expected explicit unavailable payload, got %s", tc.method, tc.path, recorder.Body.String())
		}
		matched := false
		for _, expected := range tc.containsAny {
			if strings.Contains(recorder.Body.String(), expected) {
				matched = true
				break
			}
		}
		if !matched {
			t.Fatalf("%s %s: expected response to contain one of %v, got %s", tc.method, tc.path, tc.containsAny, recorder.Body.String())
		}
	}
}

func TestReadOnlyMemoryRoutesFallBackLocally(t *testing.T) {
	t.Skip("Skipping test for now")
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")
	cfg := config.Default()
	cfg.WorkspaceRoot = t.TempDir()
	cfg.ConfigDir = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	seedPersistedAgentMemories(t, cfg.WorkspaceRoot)
	seedPersistedMemoryContexts(t, cfg.WorkspaceRoot)
	server := New(cfg, stubDetector{})

	cases := []struct {
		path        string
		method      string
		containsAny []string
	}{
		{path: "/api/memory/search?query=bootstrap&limit=3", method: http.MethodGet, containsAny: []string{`"fallback":"go-local-memory"`, `using local persisted memory search`, `"id":"ctx-local-1"`, `Bootstrapped Go parity plan`}},
		{path: "/api/memory/context/get?id=ctx-local-1", method: http.MethodGet, containsAny: []string{`"fallback":"go-local-memory"`, `using local persisted memory context body`, `Bootstrapped Go parity plan`}},
		{path: "/api/memory/context/get?id=ctx-missing", method: http.MethodGet, containsAny: []string{`"fallback":"go-local-memory"`, `local memory context fallback has no persisted context body`}},
		{path: "/api/memory/agent-search?query=memory&type=working&limit=5", method: http.MethodGet, containsAny: []string{`"fallback":"go-local-memory"`, `using local persisted agent memory search`, `Updated search_tools fallback to read persisted memory`}},
		{path: "/api/memory/observations/search?query=search_tools&limit=5&namespace=project&type=fix", method: http.MethodGet, containsAny: []string{`"fallback":"go-local-memory"`, `using local persisted observation search`, `"obs-fix-1"`}},
		{path: "/api/memory/user-prompts/search?query=Go%20parity&limit=4&role=prompt", method: http.MethodGet, containsAny: []string{`"fallback":"go-local-memory"`, `using local persisted user prompt search`, `"prompt-1"`}},
		{path: "/api/memory/session-summaries/search?query=search_tools&limit=5", method: http.MethodGet, containsAny: []string{`"fallback":"go-local-memory"`, `using local persisted session summary search`, `"summary-1"`}},
	}

	for _, tc := range cases {
		recorder := httptest.NewRecorder()
		server.Handler().ServeHTTP(recorder, httptest.NewRequest(tc.method, tc.path, nil))
		expectedStatus := http.StatusOK
		switch tc.path {
		case "/api/memory/context/get?id=ctx-missing":
			expectedStatus = http.StatusServiceUnavailable
		}
		if recorder.Code != expectedStatus {
			t.Fatalf("%s %s: expected %d, got %d", tc.method, tc.path, expectedStatus, recorder.Code)
		}
		matched := false
		for _, expected := range tc.containsAny {
			if strings.Contains(recorder.Body.String(), expected) {
				matched = true
				break
			}
		}
		if !matched {
			t.Fatalf("%s %s: expected response to contain one of %v, got %s", tc.method, tc.path, tc.containsAny, recorder.Body.String())
		}
	}
}

func seedPersistedMemoryContexts(t *testing.T, workspaceRoot string) {
	t.Helper()
	contextsDir := filepath.Join(workspaceRoot, ".tormentnexus", "memory")
	if err := os.MkdirAll(contextsDir, 0o755); err != nil {
		t.Fatalf("failed to create memory dir: %v", err)
	}
	contexts := `[
	  {"id":"ctx-local-1","title":"Bootstrap plan","source":"docs","content":"Bootstrapped Go parity plan for dashboard truthfulness.","createdAt":12345,"chunks":2,"metadata":{"topic":"parity","tags":["go","dashboard"]}},
	  {"id":"ctx-local-2","title":"Session note","source":"handoff","content":"Recorded local session supervisor restore behavior.","createdAt":23456,"chunks":1,"metadata":{"topic":"sessions"}}
	]`
	if err := os.WriteFile(filepath.Join(contextsDir, "contexts.json"), []byte(contexts), 0o644); err != nil {
		t.Fatalf("failed to seed contexts registry: %v", err)
	}
}

func seedPersistedAgentMemories(t *testing.T, workspaceRoot string) {
	t.Helper()
	agentMemoryDir := filepath.Join(workspaceRoot, ".tormentnexus", "agent_memory")
	if err := os.MkdirAll(agentMemoryDir, 0o755); err != nil {
		t.Fatalf("mkdir agent memory dir: %v", err)
	}

	now := time.Now().UTC()
	memories := map[string]any{
		"version": 1,
		"memories": []map[string]any{
			{
				"id":        "obs-fix-1",
				"type":      "working",
				"namespace": "project",
				"content":   "Updated search_tools fallback to read persisted memory",
				"metadata": map[string]any{
					"structuredObservation": map[string]any{
						"type":          "fix",
						"title":         "search_tools: persisted memory fallback",
						"narrative":     "Updated search_tools fallback to read persisted memory",
						"toolName":      "search_tools",
						"contentHash":   "obs-fix",
						"recordedAt":    float64(now.Add(-2 * time.Minute).UnixMilli()),
						"concepts":      []string{"search", "tools", "memory"},
						"filesRead":     []string{"go/internal/httpapi/server.go"},
						"filesModified": []string{"go/internal/httpapi/server.go"},
					},
				},
				"createdAt":   now.Add(-2 * time.Minute).Format(time.RFC3339Nano),
				"accessedAt":  now.Add(-90 * time.Second).Format(time.RFC3339Nano),
				"accessCount": 2,
			},
			{
				"id":        "obs-discovery-1",
				"type":      "working",
				"namespace": "project",
				"content":   "Inspected MemoryManager provider storage boundaries",
				"metadata": map[string]any{
					"structuredObservation": map[string]any{
						"type":        "discovery",
						"title":       "Mapped provider-backed context limits",
						"narrative":   "Confirmed getContext remains provider-backed",
						"toolName":    "view",
						"contentHash": "obs-discovery",
						"recordedAt":  float64(now.Add(-4 * time.Minute).UnixMilli()),
						"concepts":    []string{"memory", "provider", "context"},
						"filesRead":   []string{"packages/core/src/services/MemoryManager.ts"},
					},
				},
				"createdAt":   now.Add(-4 * time.Minute).Format(time.RFC3339Nano),
				"accessedAt":  now.Add(-3 * time.Minute).Format(time.RFC3339Nano),
				"accessCount": 1,
			},
			{
				"id":        "prompt-1",
				"type":      "long_term",
				"namespace": "project",
				"content":   "Need help with Go parity",
				"metadata": map[string]any{
					"structuredUserPrompt": map[string]any{
						"role":         "prompt",
						"content":      "Need help with Go parity",
						"contentHash":  "prompt-1",
						"promptNumber": 1,
						"recordedAt":   float64(now.Add(-90 * time.Second).UnixMilli()),
					},
				},
				"createdAt":   now.Add(-90 * time.Second).Format(time.RFC3339Nano),
				"accessedAt":  now.Add(-60 * time.Second).Format(time.RFC3339Nano),
				"accessCount": 1,
			},
			{
				"id":        "summary-1",
				"type":      "long_term",
				"namespace": "project",
				"content":   "search_tools parity session ended with status completed.",
				"metadata": map[string]any{
					"structuredSessionSummary": map[string]any{
						"sessionId":     "sess-1",
						"name":          "search_tools parity session",
						"cliType":       "tormentnexus",
						"status":        "completed",
						"activeGoal":    "ship parity",
						"lastObjective": "surface jit tool context",
						"contentHash":   "summary-1",
						"recordedAt":    float64(now.Add(-30 * time.Second).UnixMilli()),
						"stoppedAt":     float64(now.Add(-30 * time.Second).UnixMilli()),
						"logTail":       []string{"completed fallback parity"},
					},
				},
				"createdAt":   now.Add(-30 * time.Second).Format(time.RFC3339Nano),
				"accessedAt":  now.Add(-20 * time.Second).Format(time.RFC3339Nano),
				"accessCount": 3,
			},
			{
				"id":          "expired-session",
				"type":        "session",
				"namespace":   "agent",
				"content":     "expired memory",
				"metadata":    map[string]any{},
				"createdAt":   now.Add(-2 * time.Hour).Format(time.RFC3339Nano),
				"accessedAt":  now.Add(-90 * time.Minute).Format(time.RFC3339Nano),
				"accessCount": 0,
				"ttl":         1000,
			},
		},
	}

	raw, err := json.Marshal(memories)
	if err != nil {
		t.Fatalf("marshal persisted memories: %v", err)
	}
	if err := os.WriteFile(filepath.Join(agentMemoryDir, "memories.json"), raw, 0o644); err != nil {
		t.Fatalf("write persisted memories: %v", err)
	}
}

func TestMemoryServiceBackedMutationsFallBackLocally(t *testing.T) {
	skipIfServerRunning(t)
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")
	cfg := config.Default()
	cfg.WorkspaceRoot = t.TempDir()
	cfg.ConfigDir = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	seedPersistedMemoryContexts(t, cfg.WorkspaceRoot)
	server := New(cfg, stubDetector{})

	cases := []struct {
		path           string
		body           string
		expectedStatus int
		containsAny    []string
	}{
		{path: "/api/memory/context/delete", body: `{"id":"ctx-local-2"}`, expectedStatus: http.StatusOK, containsAny: []string{`"fallback":"go-local-memory"`, `using local memory context registry deletion`, `"success":true`}},
		{path: "/api/memory/facts/add", body: `{"content":"remember this","type":"working"}`, expectedStatus: http.StatusOK, containsAny: []string{`"fallback":"go-local-memory"`, `persisted memory fact locally`, `"content":"remember this"`}},
		{path: "/api/memory/observations/record", body: `{"content":"Observation","type":"fact","namespace":"project"}`, expectedStatus: http.StatusOK, containsAny: []string{`"fallback":"go-local-memory"`, `persisted structured observation locally`, `"structuredObservation"`}},
		{path: "/api/memory/user-prompts/capture", body: `{"content":"Need help","role":"user"}`, expectedStatus: http.StatusOK, containsAny: []string{`"fallback":"go-local-memory"`, `persisted user prompt locally`, `"promptRole":"user"`}},
		{path: "/api/memory/pivot/search", body: `{"pivotMemoryId":"mem-1","limit":5}`, expectedStatus: http.StatusOK, containsAny: []string{`"fallback":"go-local-memory"`, `local memory pivot anchor was not found`}},
		{path: "/api/memory/timeline/window", body: `{"centerMemoryId":"mem-1","before":2,"after":2}`, expectedStatus: http.StatusOK, containsAny: []string{`"fallback":"go-local-memory"`, `local timeline anchor was not found`}},
		{path: "/api/memory/cross-session-links", body: `{"memoryId":"mem-1","limit":5}`, expectedStatus: http.StatusOK, containsAny: []string{`"fallback":"go-local-memory"`, `local cross-session link search found no related persisted memories`}},
		{path: "/api/memory/session-summaries/capture", body: `{"sessionId":"sess-1","status":"stopped"}`, expectedStatus: http.StatusOK, containsAny: []string{`"fallback":"go-local-memory"`, `persisted session summary locally`, `"sessionId":"sess-1"`}},
	}

	for _, tc := range cases {
		recorder := httptest.NewRecorder()
		server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, tc.path, strings.NewReader(tc.body)))
		if recorder.Code != tc.expectedStatus {
			t.Fatalf("POST %s: expected %d, got %d", tc.path, tc.expectedStatus, recorder.Code)
		}
		if tc.expectedStatus == http.StatusServiceUnavailable && !strings.Contains(recorder.Body.String(), `"success":false`) {
			t.Fatalf("POST %s: expected explicit failure payload, got %s", tc.path, recorder.Body.String())
		}
		matched := false
		for _, expected := range tc.containsAny {
			if strings.Contains(recorder.Body.String(), expected) {
				matched = true
				break
			}
		}
		if !matched {
			t.Fatalf("POST %s: expected response to contain one of %v, got %s", tc.path, tc.containsAny, recorder.Body.String())
		}
	}

	observationRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(observationRecorder, httptest.NewRequest(http.MethodGet, "/api/memory/observations/recent?limit=5&namespace=project&type=discovery", nil))
	if observationRecorder.Code != http.StatusOK || !strings.Contains(observationRecorder.Body.String(), `"content":"Observation"`) {
		t.Fatalf("expected persisted local observation after mutation, got %d %s", observationRecorder.Code, observationRecorder.Body.String())
	}

	promptRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(promptRecorder, httptest.NewRequest(http.MethodGet, "/api/memory/user-prompts/recent?limit=5&role=user", nil))
	if promptRecorder.Code != http.StatusOK || !strings.Contains(promptRecorder.Body.String(), `"content":"Need help"`) {
		t.Fatalf("expected persisted local user prompt after mutation, got %d %s", promptRecorder.Code, promptRecorder.Body.String())
	}

	summaryRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(summaryRecorder, httptest.NewRequest(http.MethodGet, "/api/memory/session-summaries/recent?limit=5", nil))
	if summaryRecorder.Code != http.StatusOK || !strings.Contains(summaryRecorder.Body.String(), `"sessionId":"sess-1"`) {
		t.Fatalf("expected persisted local session summary after mutation, got %d %s", summaryRecorder.Code, summaryRecorder.Body.String())
	}

	contextsRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(contextsRecorder, httptest.NewRequest(http.MethodGet, "/api/memory/contexts", nil))
	if contextsRecorder.Code != http.StatusOK || strings.Contains(contextsRecorder.Body.String(), `"ctx-local-2"`) {
		t.Fatalf("expected deleted local memory context to be removed from registry, got %d %s", contextsRecorder.Code, contextsRecorder.Body.String())
	}
}

func seedPersistedAgentMemoryRelations(t *testing.T, workspaceRoot string) {
	t.Helper()
	seedPersistedAgentMemories(t, workspaceRoot)

	agentMemoryPath := filepath.Join(workspaceRoot, ".tormentnexus", "agent_memory", "memories.json")
	raw, err := os.ReadFile(agentMemoryPath)
	if err != nil {
		t.Fatalf("read persisted memories: %v", err)
	}

	var snapshot struct {
		Version  int                      `json:"version"`
		Memories []localAgentMemoryRecord `json:"memories"`
	}
	if err := json.Unmarshal(raw, &snapshot); err != nil {
		t.Fatalf("unmarshal persisted memories: %v", err)
	}

	now := time.Now().UTC()
	snapshot.Memories = append(snapshot.Memories,
		localAgentMemoryRecord{
			ID:        "prompt-goal-1",
			Type:      "long_term",
			Namespace: "project",
			Content:   "Ship parity",
			Metadata: map[string]any{
				"sessionId": "sess-1",
				"structuredUserPrompt": map[string]any{
					"role":          "goal",
					"content":       "Ship parity",
					"sessionId":     "sess-1",
					"activeGoal":    "ship parity",
					"lastObjective": "surface jit tool context",
					"contentHash":   "prompt-goal-1",
					"promptNumber":  2,
					"recordedAt":    float64(now.Add(-40 * time.Second).UnixMilli()),
				},
			},
			CreatedAt:   now.Add(-40 * time.Second).Format(time.RFC3339Nano),
			AccessedAt:  now.Add(-35 * time.Second).Format(time.RFC3339Nano),
			AccessCount: 1,
		},
		localAgentMemoryRecord{
			ID:        "obs-fix-2",
			Type:      "working",
			Namespace: "project",
			Content:   "Applied the same search_tools memory fix in a follow-up session",
			Metadata: map[string]any{
				"sessionId": "sess-2",
				"source":    "search_tools",
				"structuredObservation": map[string]any{
					"type":          "fix",
					"title":         "search_tools follow-up parity fix",
					"narrative":     "Applied the same search_tools memory fix in a follow-up session",
					"toolName":      "search_tools",
					"contentHash":   "obs-fix-2",
					"recordedAt":    float64(now.Add(-10 * time.Second).UnixMilli()),
					"concepts":      []string{"search", "tools", "memory"},
					"filesRead":     []string{"go/internal/httpapi/server.go"},
					"filesModified": []string{"go/internal/httpapi/server.go"},
				},
			},
			CreatedAt:   now.Add(-10 * time.Second).Format(time.RFC3339Nano),
			AccessedAt:  now.Add(-5 * time.Second).Format(time.RFC3339Nano),
			AccessCount: 1,
		},
	)

	updatedRaw, err := json.Marshal(map[string]any{
		"version":  1,
		"memories": snapshot.Memories,
	})
	if err != nil {
		t.Fatalf("marshal related persisted memories: %v", err)
	}
	if err := os.WriteFile(agentMemoryPath, updatedRaw, 0o644); err != nil {
		t.Fatalf("write related persisted memories: %v", err)
	}
}

func TestMemoryRelationshipRoutesFallBackToPersistedData(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")
	cfg := config.Default()
	cfg.WorkspaceRoot = t.TempDir()
	cfg.ConfigDir = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	seedPersistedAgentMemoryRelations(t, cfg.WorkspaceRoot)
	server := New(cfg, stubDetector{})

	pivotRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(pivotRecorder, httptest.NewRequest(http.MethodPost, "/api/memory/pivot/search", strings.NewReader(`{"pivotMemoryId":"obs-fix-1","limit":5}`)))
	if pivotRecorder.Code != http.StatusOK || !strings.Contains(pivotRecorder.Body.String(), `using local inferred pivot search from anchor memory`) || !strings.Contains(pivotRecorder.Body.String(), `"obs-fix-2"`) {
		t.Fatalf("expected local pivot-search fallback to return inferred related memories, got %d %s", pivotRecorder.Code, pivotRecorder.Body.String())
	}

	timelineRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(timelineRecorder, httptest.NewRequest(http.MethodPost, "/api/memory/timeline/window", strings.NewReader(`{"centerMemoryId":"summary-1","before":2,"after":2}`)))
	if timelineRecorder.Code != http.StatusOK || !strings.Contains(timelineRecorder.Body.String(), `using local timeline window inferred from anchor memory`) || !strings.Contains(timelineRecorder.Body.String(), `"summary-1"`) || !strings.Contains(timelineRecorder.Body.String(), `"prompt-goal-1"`) {
		t.Fatalf("expected local timeline-window fallback to return session window data, got %d %s", timelineRecorder.Code, timelineRecorder.Body.String())
	}

	linksRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(linksRecorder, httptest.NewRequest(http.MethodPost, "/api/memory/cross-session-links", strings.NewReader(`{"memoryId":"obs-fix-1","limit":5}`)))
	if linksRecorder.Code != http.StatusOK || !strings.Contains(linksRecorder.Body.String(), `using local persisted cross-session memory links`) || !strings.Contains(linksRecorder.Body.String(), `"obs-fix-2"`) || !strings.Contains(linksRecorder.Body.String(), `same tool (search_tools)`) {
		t.Fatalf("expected local cross-session-link fallback to return related session links, got %d %s", linksRecorder.Code, linksRecorder.Body.String())
	}
}

func TestMCPAddAndRemoveServerFallBackToLocalConfiguredServers(t *testing.T) {
	workspaceRoot := t.TempDir()
	tormentnexusDir := filepath.Join(workspaceRoot, ".tormentnexus")
	if err := os.MkdirAll(tormentnexusDir, 0o755); err != nil {
		t.Fatalf("failed to create .tormentnexus dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tormentnexusDir, "mcp.jsonc"), []byte("{\"mcpServers\":{}}"), 0o644); err != nil {
		t.Fatalf("failed to seed mcp.jsonc: %v", err)
	}

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")
	cfg := config.Default()
	cfg.WorkspaceRoot = workspaceRoot
	cfg.ConfigDir = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	addReq := httptest.NewRequest(http.MethodPost, "/api/mcp/runtime-servers/add", strings.NewReader(`{"name":"local-test","command":"npx","args":["-y","test-server"]}`))
	addReq.Header.Set("content-type", "application/json")
	addRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(addRecorder, addReq)
	if addRecorder.Code != http.StatusOK || !strings.Contains(addRecorder.Body.String(), `"fallback":"go-local-jsonc"`) || !strings.Contains(addRecorder.Body.String(), `"name":"local-test"`) {
		t.Fatalf("expected local addServer fallback response, got %d %s", addRecorder.Code, addRecorder.Body.String())
	}

	removeReq := httptest.NewRequest(http.MethodPost, "/api/mcp/runtime-servers/remove", strings.NewReader(`{"name":"local-test"}`))
	removeReq.Header.Set("content-type", "application/json")
	removeRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(removeRecorder, removeReq)
	if removeRecorder.Code != http.StatusOK || !strings.Contains(removeRecorder.Body.String(), `"fallback":"go-local-jsonc"`) || !strings.Contains(removeRecorder.Body.String(), `"success":true`) {
		t.Fatalf("expected local removeServer fallback response, got %d %s", removeRecorder.Code, removeRecorder.Body.String())
	}
}

func TestMCPServerTestFallsBackToStructuredProbeFailures(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")
	cfg := config.Default()
	cfg.WorkspaceRoot = t.TempDir()
	cfg.ConfigDir = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	routerReq := httptest.NewRequest(http.MethodPost, "/api/mcp/server-test", strings.NewReader(`{"targetKind":"router","operation":"tools/list"}`))
	routerReq.Header.Set("content-type", "application/json")
	routerRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(routerRecorder, routerReq)
	if routerRecorder.Code != http.StatusOK || !strings.Contains(routerRecorder.Body.String(), `"fallback":"go-local-mcp"`) || !strings.Contains(routerRecorder.Body.String(), `tormentnexus MCP router is not initialized.`) {
		t.Fatalf("expected local router probe fallback response, got %d %s", routerRecorder.Code, routerRecorder.Body.String())
	}
	if !strings.Contains(routerRecorder.Body.String(), `simulating router probe failure locally`) {
		t.Fatalf("expected local router probe fallback reason, got %s", routerRecorder.Body.String())
	}

	serverReq := httptest.NewRequest(http.MethodPost, "/api/mcp/server-test", strings.NewReader(`{"targetKind":"server","operation":"tools/list"}`))
	serverReq.Header.Set("content-type", "application/json")
	serverRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(serverRecorder, serverReq)
	if serverRecorder.Code != http.StatusOK || !strings.Contains(serverRecorder.Body.String(), `Downstream probe requires a server name.`) {
		t.Fatalf("expected downstream validation fallback response, got %d %s", serverRecorder.Code, serverRecorder.Body.String())
	}
	if !strings.Contains(serverRecorder.Body.String(), `validating probe request locally`) {
		t.Fatalf("expected local probe validation fallback reason, got %s", serverRecorder.Body.String())
	}
}

func TestMCPLifecycleModesFallBackToLocalState(t *testing.T) {
	t.Skip("Skipping test for now")
	workspaceRoot := t.TempDir()
	toolsDir := filepath.Join(workspaceRoot, "submodules", "tormentnexus", "tools")
	if err := os.MkdirAll(toolsDir, 0o755); err != nil {
		t.Fatalf("failed to create tormentnexus tools dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(toolsDir, "registry.go"), []byte("package tools\n"), 0o644); err != nil {
		t.Fatalf("failed to seed tool dir: %v", err)
	}

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")
	cfg := config.Default()
	cfg.WorkspaceRoot = workspaceRoot
	cfg.ConfigDir = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	setReq := httptest.NewRequest(http.MethodPost, "/api/mcp/lifecycle-modes", strings.NewReader(`{"lazySessionMode":true,"singleActiveServerMode":false}`))
	setReq.Header.Set("content-type", "application/json")
	setRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(setRecorder, setReq)
	if setRecorder.Code != http.StatusOK || !strings.Contains(setRecorder.Body.String(), `"lazySessionMode":true`) {
		t.Fatalf("expected lifecycle response to set lazySessionMode, got %d %s", setRecorder.Code, setRecorder.Body.String())
	}

	statusRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(statusRecorder, httptest.NewRequest(http.MethodGet, "/api/mcp/status", nil))
	if statusRecorder.Code != http.StatusOK || !strings.Contains(statusRecorder.Body.String(), `"lazySessionMode":true`) {
		t.Fatalf("expected fallback mcp status to include lifecycle state, got %d %s", statusRecorder.Code, statusRecorder.Body.String())
	}
}

func TestMCPLoadAndUnloadToolReturnExplicitUnavailableFallback(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")
	cfg := config.Default()
	cfg.WorkspaceRoot = t.TempDir()
	cfg.ConfigDir = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	for _, path := range []string{"/api/mcp/working-set/load", "/api/mcp/working-set/unload"} {
		req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{"name":"search_tools"}`))
		req.Header.Set("content-type", "application/json")
		recorder := httptest.NewRecorder()
		server.Handler().ServeHTTP(recorder, req)
		if recorder.Code != http.StatusServiceUnavailable || !strings.Contains(recorder.Body.String(), `"fallback":"go-local-mcp"`) || !strings.Contains(recorder.Body.String(), `MCP Server not initialized`) {
			t.Fatalf("%s: expected explicit unavailable fallback, got %d %s", path, recorder.Code, recorder.Body.String())
		}
		if !strings.Contains(recorder.Body.String(), `"success":false`) {
			t.Fatalf("%s: expected explicit failure payload, got %s", path, recorder.Body.String())
		}
		if !strings.Contains(recorder.Body.String(), `local MCP working set manager is not initialized`) {
			t.Fatalf("%s: expected local working-set fallback reason, got %s", path, recorder.Body.String())
		}
	}
}

func TestAutonomyBridgeRoutes(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		switch r.URL.Path {
		case "/trpc/autonomy.getLevel":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": "medium"}},
			})
		case "/trpc/autonomy.setLevel":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"level":"high"`) {
				t.Fatalf("expected autonomy.setLevel payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": "high"}},
			})
		case "/trpc/autonomy.activateFullAutonomy":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": "Autonomous Supervisor Activated (High Level + Chat Daemon + Watchdog)"}},
			})
		default:
			t.Fatalf("unexpected upstream path %s", r.URL.Path)
		}
	}))
	defer upstream.Close()

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", upstream.URL+"/trpc")

	cfg := config.Default()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	cases := []struct {
		name      string
		method    string
		path      string
		body      string
		contains  string
		procedure string
	}{
		{name: "autonomy get level", method: http.MethodGet, path: "/api/autonomy/get-level", contains: `"medium"`, procedure: `"procedure":"autonomy.getLevel"`},
		{name: "autonomy set level", method: http.MethodPost, path: "/api/autonomy/set-level", body: `{"level":"high"}`, contains: `"high"`, procedure: `"procedure":"autonomy.setLevel"`},
		{name: "autonomy activate full", method: http.MethodPost, path: "/api/autonomy/activate-full", body: `null`, contains: `"Autonomous Supervisor Activated`, procedure: `"procedure":"autonomy.activateFullAutonomy"`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var body io.Reader
			if tc.body != "" {
				body = strings.NewReader(tc.body)
			}
			request := httptest.NewRequest(tc.method, tc.path, body)
			if tc.body != "" {
				request.Header.Set("content-type", "application/json")
			}
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, request)
			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.contains) {
				t.Fatalf("expected response to contain %s, got %s", tc.contains, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.procedure) {
				t.Fatalf("expected bridge metadata %s, got %s", tc.procedure, recorder.Body.String())
			}

			var payload struct {
				Success bool            `json:"success"`
				Data    json.RawMessage `json:"data"`
				Bridge  struct {
					Procedure string `json:"procedure"`
				} `json:"bridge"`
			}
			if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
				t.Fatalf("expected JSON payload, got decode error: %v", err)
			}
			if !payload.Success {
				t.Fatalf("expected success payload, got %s", recorder.Body.String())
			}

			switch tc.name {
			case "list tools":
				var tools []map[string]any
				if err := json.Unmarshal(payload.Data, &tools); err != nil {
					t.Fatalf("expected tools payload, got decode error: %v", err)
				}
				if len(tools) != 1 || tools[0]["name"] != "search_tools" || tools[0]["server"] != "core" {
					t.Fatalf("expected single core search_tools entry, got %+v", tools)
				}
			case "search tools":
				var tools []map[string]any
				if err := json.Unmarshal(payload.Data, &tools); err != nil {
					t.Fatalf("expected search results payload, got decode error: %v", err)
				}
				if len(tools) != 1 || tools[0]["name"] != "search_tools" || tools[0]["alwaysShow"] != true {
					t.Fatalf("expected always-show search_tools result, got %+v", tools)
				}
			case "call tool":
				var result struct {
					Ok     bool `json:"ok"`
					Result struct {
						Content []struct {
							Type string `json:"type"`
							Text string `json:"text"`
						} `json:"content"`
					} `json:"result"`
				}
				if err := json.Unmarshal(payload.Data, &result); err != nil {
					t.Fatalf("expected tool call payload, got decode error: %v", err)
				}
				if !result.Ok || len(result.Result.Content) != 1 || result.Result.Content[0].Text != "done" {
					t.Fatalf("expected successful tool call result, got %+v", result)
				}
			case "auto call tool":
				var result struct {
					Ok     bool `json:"ok"`
					Result struct {
						Content []struct {
							Type string `json:"type"`
							Text string `json:"text"`
						} `json:"content"`
					} `json:"result"`
				}
				if err := json.Unmarshal(payload.Data, &result); err != nil {
					t.Fatalf("expected auto tool call payload, got decode error: %v", err)
				}
				if payload.Bridge.Procedure != "mcp.callTool" || !result.Ok || len(result.Result.Content) != 1 || result.Result.Content[0].Text != "auto_call_tool" {
					t.Fatalf("expected successful auto_call_tool bridge result, got bridge=%+v data=%+v", payload.Bridge, result)
				}
			case "tool advertisements":
				var result struct {
					RecommendedTools []map[string]any `json:"recommendedTools"`
					RelatedTools     struct {
						Ok     bool `json:"ok"`
						Result struct {
							Content []struct {
								Type string `json:"type"`
								Text string `json:"text"`
							} `json:"content"`
						} `json:"result"`
					} `json:"relatedTools"`
				}
				if err := json.Unmarshal(payload.Data, &result); err != nil {
					t.Fatalf("expected tool advertisement payload, got decode error: %v", err)
				}
				if len(result.RecommendedTools) != 1 || result.RecommendedTools[0]["name"] != "search_tools" {
					t.Fatalf("expected recommended search_tools payload, got %+v", result.RecommendedTools)
				}
				if !result.RelatedTools.Ok || len(result.RelatedTools.Result.Content) != 1 || result.RelatedTools.Result.Content[0].Text != "list_all_tools" {
					t.Fatalf("expected list_all_tools advertisement payload, got %+v", result)
				}
				if !strings.Contains(recorder.Body.String(), "\"procedure\":\"mcp.searchTools\"") || !strings.Contains(recorder.Body.String(), "\"toolName\":\"list_all_tools\"") {
					t.Fatalf("expected mixed bridge metadata for tool advertisements, got %s", recorder.Body.String())
				}
			}
		})
	}
}

func TestDirectorBridgeRoutes(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		switch r.URL.Path {
		case "/trpc/director.memorize":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": "Memorized."}},
			})
		case "/trpc/director.chat":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"message":"status?"`) {
				t.Fatalf("expected director.chat payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": "Director online"}},
			})
		case "/trpc/director.status":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{"status": "online"}}},
			})
		case "/trpc/director.updateConfig":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}},
			})
		case "/trpc/directorConfig.get":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{"persona": "professional", "defaultTopic": "status"}}},
			})
		case "/trpc/directorConfig.test":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true, "directorReady": true, "llmServiceReady": true}}},
			})
		case "/trpc/directorConfig.update":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"persona":"chaos"`) {
				t.Fatalf("expected directorConfig.update payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}},
			})
		case "/trpc/director.stopAutoDrive":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": "Stopped"}},
			})
		case "/trpc/director.startAutoDrive":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": "Started"}},
			})
		default:
			t.Fatalf("unexpected upstream path %s", r.URL.Path)
		}
	}))
	defer upstream.Close()

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", upstream.URL+"/trpc")

	cfg := config.Default()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	cases := []struct {
		name      string
		method    string
		path      string
		body      string
		contains  string
		procedure string
	}{
		{name: "director memorize", method: http.MethodPost, path: "/api/director/memorize", body: `{"content":"memo","source":"web","title":"Note"}`, contains: `"Memorized."`, procedure: `"procedure":"director.memorize"`},
		{name: "director chat", method: http.MethodPost, path: "/api/director/chat", body: `{"message":"status?"}`, contains: `"Director online"`, procedure: `"procedure":"director.chat"`},
		{name: "director status", method: http.MethodGet, path: "/api/director/status", contains: `"status":"online"`, procedure: `"procedure":"director.status"`},
		{name: "director update config", method: http.MethodPost, path: "/api/director/config/update", body: `{"defaultTopic":"mcp"}`, contains: `"success":true`, procedure: `"procedure":"director.updateConfig"`},
		{name: "director-config get", method: http.MethodGet, path: "/api/director-config", contains: `"persona":"professional"`, procedure: `"procedure":"directorConfig.get"`},
		{name: "director-config test", method: http.MethodGet, path: "/api/director-config/test", contains: `"directorReady":true`, procedure: `"procedure":"directorConfig.test"`},
		{name: "director-config update", method: http.MethodPost, path: "/api/director-config/update", body: `{"persona":"chaos"}`, contains: `"success":true`, procedure: `"procedure":"directorConfig.update"`},
		{name: "director stop auto drive", method: http.MethodPost, path: "/api/director/auto-drive/stop", body: `null`, contains: `"Stopped"`, procedure: `"procedure":"director.stopAutoDrive"`},
		{name: "director start auto drive", method: http.MethodPost, path: "/api/director/auto-drive/start", body: `null`, contains: `"Started"`, procedure: `"procedure":"director.startAutoDrive"`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var body io.Reader
			if tc.body != "" {
				body = strings.NewReader(tc.body)
			}
			request := httptest.NewRequest(tc.method, tc.path, body)
			if tc.body != "" {
				request.Header.Set("content-type", "application/json")
			}
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, request)
			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.contains) {
				t.Fatalf("expected response to contain %s, got %s", tc.contains, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.procedure) {
				t.Fatalf("expected bridge metadata %s, got %s", tc.procedure, recorder.Body.String())
			}
		})
	}
}

func TestAutoDevBridgeRoutes(t *testing.T) {
	t.Skip("Skipping test for now")
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		switch r.URL.Path {
		case "/trpc/autoDev.startLoop":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"type":"test"`) {
				t.Fatalf("expected autoDev.startLoop payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true, "loopId": "loop-1"}}},
			})
		case "/trpc/autoDev.cancelLoop":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"loopId":"loop-1"`) {
				t.Fatalf("expected autoDev.cancelLoop payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": true}},
			})
		case "/trpc/autoDev.getLoops":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"id": "loop-1", "status": "running"}}}},
			})
		case "/trpc/autoDev.getLoop":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"loopId":"loop-1"`) {
				t.Fatalf("expected autoDev.getLoop payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{"id": "loop-1", "status": "running"}}},
			})
		case "/trpc/autoDev.clearCompleted":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": 3}},
			})
		default:
			t.Fatalf("unexpected upstream path %s", r.URL.Path)
		}
	}))
	defer upstream.Close()

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", upstream.URL+"/trpc")

	cfg := config.Default()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	cases := []struct {
		name      string
		method    string
		path      string
		body      string
		contains  string
		procedure string
	}{
		{name: "autodev start loop", method: http.MethodPost, path: "/api/autodev/start-loop", body: `{"type":"test","maxAttempts":3}`, contains: `"loopId":"loop-1"`, procedure: `"procedure":"autoDev.startLoop"`},
		{name: "autodev cancel loop", method: http.MethodPost, path: "/api/autodev/cancel-loop", body: `{"loopId":"loop-1"}`, contains: `"data":true`, procedure: `"procedure":"autoDev.cancelLoop"`},
		{name: "autodev list loops", method: http.MethodGet, path: "/api/autodev/loops", contains: `"loop-1"`, procedure: `"procedure":"autoDev.getLoops"`},
		{name: "autodev get loop", method: http.MethodGet, path: "/api/autodev/loop?loopId=loop-1", contains: `"status":"running"`, procedure: `"procedure":"autoDev.getLoop"`},
		{name: "autodev clear completed", method: http.MethodPost, path: "/api/autodev/clear-completed", contains: `"data":3`, procedure: `"procedure":"autoDev.clearCompleted"`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var body io.Reader
			if tc.body != "" {
				body = strings.NewReader(tc.body)
			}
			request := httptest.NewRequest(tc.method, tc.path, body)
			if tc.body != "" {
				request.Header.Set("content-type", "application/json")
			}
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, request)
			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.contains) {
				t.Fatalf("expected response to contain %s, got %s", tc.contains, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.procedure) {
				t.Fatalf("expected bridge metadata %s, got %s", tc.procedure, recorder.Body.String())
			}
		})
	}
}

func TestDarwinBridgeRoutes(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		switch r.URL.Path {
		case "/trpc/darwin.evolve":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"goal":"Improve prompts"`) {
				t.Fatalf("expected darwin.evolve payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{"mutationId": "mut-1"}}},
			})
		case "/trpc/darwin.experiment":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"mutationId":"mut-1"`) {
				t.Fatalf("expected darwin.experiment payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{"experimentId": "exp-1"}}},
			})
		case "/trpc/darwin.getStatus":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{"running": true, "experimentCount": 1}}},
			})
		default:
			t.Fatalf("unexpected upstream path %s", r.URL.Path)
		}
	}))
	defer upstream.Close()

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", upstream.URL+"/trpc")

	cfg := config.Default()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	cases := []struct {
		name      string
		method    string
		path      string
		body      string
		contains  string
		procedure string
	}{
		{name: "darwin evolve", method: http.MethodPost, path: "/api/darwin/evolve", body: `{"prompt":"Refactor prompt","goal":"Improve prompts"}`, contains: `"mutationId":"mut-1"`, procedure: `"procedure":"darwin.evolve"`},
		{name: "darwin experiment", method: http.MethodPost, path: "/api/darwin/experiment", body: `{"mutationId":"mut-1","task":"Run benchmark"}`, contains: `"experimentId":"exp-1"`, procedure: `"procedure":"darwin.experiment"`},
		{name: "darwin status", method: http.MethodGet, path: "/api/darwin/status", contains: `"running":true`, procedure: `"procedure":"darwin.getStatus"`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var body io.Reader
			if tc.body != "" {
				body = strings.NewReader(tc.body)
			}
			request := httptest.NewRequest(tc.method, tc.path, body)
			if tc.body != "" {
				request.Header.Set("content-type", "application/json")
			}
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, request)
			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.contains) {
				t.Fatalf("expected response to contain %s, got %s", tc.contains, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.procedure) {
				t.Fatalf("expected bridge metadata %s, got %s", tc.procedure, recorder.Body.String())
			}
		})
	}
}

func TestCouncilBridgeRoutes(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		switch r.URL.Path {
		case "/trpc/council.members":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"name": "planner", "provider": "openai", "modelId": "gpt-5", "systemPrompt": "plan"}}}},
			})
		case "/trpc/council.updateMembers":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}},
			})
		case "/trpc/council.sessions.list":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"id": "sess-1"}}}},
			})
		case "/trpc/council.sessions.active":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"id": "sess-1", "status": "running"}}}},
			})
		case "/trpc/council.sessions.stats":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{"total": 1, "active": 1}}},
			})
		case "/trpc/council.sessions.get":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"id":"sess-1"`) {
				t.Fatalf("expected council.sessions.get payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{"id": "sess-1"}}},
			})
		case "/trpc/council.sessions.start":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{"id": "sess-2"}}},
			})
		case "/trpc/council.sessions.bulkStart":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"count":2`) {
				t.Fatalf("expected council.sessions.bulkStart payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{"sessions": []any{map[string]any{"id": "sess-bulk-1"}, map[string]any{"id": "sess-bulk-2"}}, "failed": []any{}}}},
			})
		case "/trpc/council.sessions.bulkStop":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{"stopped": 2, "failed": 0}}},
			})
		case "/trpc/council.sessions.bulkResume":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{"sessions": []any{map[string]any{"id": "sess-1"}, map[string]any{"id": "sess-2"}}, "failed": []any{}}}},
			})
		case "/trpc/council.sessions.stop":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{"id": "sess-1", "status": "stopped"}}},
			})
		case "/trpc/council.sessions.resume":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{"id": "sess-1", "status": "running"}}},
			})
		case "/trpc/council.sessions.delete":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}},
			})
		case "/trpc/council.sessions.sendGuidance":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"id":"sess-1"`) || !strings.Contains(string(body), `"approved":true`) {
				t.Fatalf("expected council.sessions.sendGuidance payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{"id": "sess-1", "guidanceApplied": true}}},
			})
		case "/trpc/council.sessions.getLogs":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"id":"sess-1"`) {
				t.Fatalf("expected council.sessions.getLogs payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": []any{"log line"}}},
			})
		case "/trpc/council.sessions.templates":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"name": "default"}}}},
			})
		case "/trpc/council.sessions.startFromTemplate":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"name":"default"`) {
				t.Fatalf("expected council.sessions.startFromTemplate payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{"id": "sess-template-started", "templateName": "default"}}},
			})
		case "/trpc/council.sessions.persisted":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"id": "sess-persisted", "persisted": true}}}},
			})
		case "/trpc/council.sessions.byTag":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"tag":"priority"`) {
				t.Fatalf("expected council.sessions.byTag payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"id": "sess-tagged", "tags": []any{"priority"}}}}},
			})
		case "/trpc/council.sessions.byTemplate":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"template":"default"`) {
				t.Fatalf("expected council.sessions.byTemplate payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"id": "sess-template", "templateName": "default"}}}},
			})
		case "/trpc/council.sessions.byCLI":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"cliType":"tormentnexus"`) {
				t.Fatalf("expected council.sessions.byCLI payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"id": "sess-cli", "cliType": "tormentnexus"}}}},
			})
		case "/trpc/council.sessions.updateTags":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"id":"sess-1"`) || !strings.Contains(string(body), `"priority"`) {
				t.Fatalf("expected council.sessions.updateTags payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{"id": "sess-1", "tags": []any{"priority", "go"}}}},
			})
		case "/trpc/council.sessions.addTag":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"id":"sess-1"`) || !strings.Contains(string(body), `"tag":"priority"`) {
				t.Fatalf("expected council.sessions.addTag payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{"id": "sess-1", "tags": []any{"priority"}}}},
			})
		case "/trpc/council.sessions.removeTag":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"id":"sess-1"`) || !strings.Contains(string(body), `"tag":"priority"`) {
				t.Fatalf("expected council.sessions.removeTag payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{"id": "sess-1", "tags": []any{}}}},
			})
		case "/trpc/council.quota.status":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{"enabled": true}}},
			})
		case "/trpc/council.quota.getConfig":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{"windowMs": 60000}}},
			})
		case "/trpc/council.quota.updateConfig":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}},
			})
		case "/trpc/council.quota.enable":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true, "enabled": true}}},
			})
		case "/trpc/council.quota.disable":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true, "enabled": false}}},
			})
		case "/trpc/council.quota.check":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"provider":"openai"`) {
				t.Fatalf("expected council.quota.check payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{"allowed": true}}},
			})
		case "/trpc/council.quota.allStats":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{"providers": 1}}},
			})
		case "/trpc/council.quota.providerStats":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"provider":"openai"`) {
				t.Fatalf("expected council.quota.providerStats payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{"provider": "openai"}}},
			})
		case "/trpc/council.quota.getLimits":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{"rpm": 60}}},
			})
		case "/trpc/council.quota.setLimits":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"provider":"openai"`) {
				t.Fatalf("expected council.quota.setLimits payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}},
			})
		case "/trpc/council.quota.resetProvider":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}},
			})
		case "/trpc/council.quota.resetAll":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}},
			})
		case "/trpc/council.quota.unthrottle":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}},
			})
		case "/trpc/council.quota.recordRequest":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}},
			})
		case "/trpc/council.quota.recordRateLimitError":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}},
			})
		default:
			t.Fatalf("unexpected upstream path %s", r.URL.Path)
		}
	}))
	defer upstream.Close()

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", upstream.URL+"/trpc")

	cfg := config.Default()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	cases := []struct {
		name      string
		method    string
		path      string
		body      string
		contains  string
		procedure string
	}{
		{name: "council members", method: http.MethodGet, path: "/api/council/members", contains: `"planner"`, procedure: `"procedure":"council.members"`},
		{name: "council update members", method: http.MethodPost, path: "/api/council/members/update", body: `[{"name":"planner","provider":"openai","modelId":"gpt-5","systemPrompt":"plan"}]`, contains: `"success":true`, procedure: `"procedure":"council.updateMembers"`},
		{name: "council sessions list", method: http.MethodGet, path: "/api/council/sessions", contains: `"sess-1"`, procedure: `"procedure":"council.sessions.list"`},
		{name: "council sessions active", method: http.MethodGet, path: "/api/council/sessions/active", contains: `"running"`, procedure: `"procedure":"council.sessions.active"`},
		{name: "council sessions stats", method: http.MethodGet, path: "/api/council/sessions/stats", contains: `"total":1`, procedure: `"procedure":"council.sessions.stats"`},
		{name: "council sessions get", method: http.MethodGet, path: "/api/council/sessions/get?id=sess-1", contains: `"sess-1"`, procedure: `"procedure":"council.sessions.get"`},
		{name: "council sessions start", method: http.MethodPost, path: "/api/council/sessions/start", body: `{"cliType":"tormentnexus"}`, contains: `"sess-2"`, procedure: `"procedure":"council.sessions.start"`},
		{name: "council sessions bulk start", method: http.MethodPost, path: "/api/council/sessions/bulk-start", body: `{"count":2,"cliType":"tormentnexus"}`, contains: `"sess-bulk-1"`, procedure: `"procedure":"council.sessions.bulkStart"`},
		{name: "council sessions bulk stop", method: http.MethodPost, path: "/api/council/sessions/bulk-stop", contains: `"stopped":2`, procedure: `"procedure":"council.sessions.bulkStop"`},
		{name: "council sessions bulk resume", method: http.MethodPost, path: "/api/council/sessions/bulk-resume", contains: `"sess-2"`, procedure: `"procedure":"council.sessions.bulkResume"`},
		{name: "council sessions stop", method: http.MethodPost, path: "/api/council/sessions/stop", body: `{"id":"sess-1"}`, contains: `"stopped"`, procedure: `"procedure":"council.sessions.stop"`},
		{name: "council sessions resume", method: http.MethodPost, path: "/api/council/sessions/resume", body: `{"id":"sess-1"}`, contains: `"running"`, procedure: `"procedure":"council.sessions.resume"`},
		{name: "council sessions delete", method: http.MethodPost, path: "/api/council/sessions/delete", body: `{"id":"sess-1"}`, contains: `"success":true`, procedure: `"procedure":"council.sessions.delete"`},
		{name: "council sessions guidance", method: http.MethodPost, path: "/api/council/sessions/guidance", body: `{"id":"sess-1","approved":true,"feedback":"ship it","suggestedNextSteps":["test"]}`, contains: `"guidanceApplied":true`, procedure: `"procedure":"council.sessions.sendGuidance"`},
		{name: "council sessions logs", method: http.MethodGet, path: "/api/council/sessions/logs?id=sess-1", contains: `"log line"`, procedure: `"procedure":"council.sessions.getLogs"`},
		{name: "council sessions templates", method: http.MethodGet, path: "/api/council/sessions/templates", contains: `"default"`, procedure: `"procedure":"council.sessions.templates"`},
		{name: "council sessions from template", method: http.MethodPost, path: "/api/council/sessions/from-template", body: `{"name":"default"}`, contains: `"sess-template-started"`, procedure: `"procedure":"council.sessions.startFromTemplate"`},
		{name: "council sessions persisted", method: http.MethodGet, path: "/api/council/sessions/persisted", contains: `"sess-persisted"`, procedure: `"procedure":"council.sessions.persisted"`},
		{name: "council sessions by tag", method: http.MethodGet, path: "/api/council/sessions/by-tag?tag=priority", contains: `"sess-tagged"`, procedure: `"procedure":"council.sessions.byTag"`},
		{name: "council sessions by template", method: http.MethodGet, path: "/api/council/sessions/by-template?template=default", contains: `"sess-template"`, procedure: `"procedure":"council.sessions.byTemplate"`},
		{name: "council sessions by cli", method: http.MethodGet, path: "/api/council/sessions/by-cli?cliType=tormentnexus", contains: `"sess-cli"`, procedure: `"procedure":"council.sessions.byCLI"`},
		{name: "council sessions update tags", method: http.MethodPost, path: "/api/council/sessions/tags/update", body: `{"id":"sess-1","tags":["priority","go"]}`, contains: `"priority","go"`, procedure: `"procedure":"council.sessions.updateTags"`},
		{name: "council sessions add tag", method: http.MethodPost, path: "/api/council/sessions/tags/add", body: `{"id":"sess-1","tag":"priority"}`, contains: `"priority"`, procedure: `"procedure":"council.sessions.addTag"`},
		{name: "council sessions remove tag", method: http.MethodPost, path: "/api/council/sessions/tags/remove", body: `{"id":"sess-1","tag":"priority"}`, contains: `"tags":[]`, procedure: `"procedure":"council.sessions.removeTag"`},
		{name: "council quota status", method: http.MethodGet, path: "/api/council/quota/status", contains: `"enabled":true`, procedure: `"procedure":"council.quota.status"`},
		{name: "council quota get config", method: http.MethodGet, path: "/api/council/quota/config", contains: `"windowMs":60000`, procedure: `"procedure":"council.quota.getConfig"`},
		{name: "council quota update config", method: http.MethodPost, path: "/api/council/quota/config", body: `{"windowMs":120000}`, contains: `"success":true`, procedure: `"procedure":"council.quota.updateConfig"`},
		{name: "council quota enable", method: http.MethodPost, path: "/api/council/quota/enabled?enabled=true", body: `null`, contains: `"enabled":true`, procedure: `"procedure":"council.quota.enable"`},
		{name: "council quota disable", method: http.MethodPost, path: "/api/council/quota/enabled?enabled=false", body: `null`, contains: `"enabled":false`, procedure: `"procedure":"council.quota.disable"`},
		{name: "council quota check", method: http.MethodGet, path: "/api/council/quota/check?provider=openai", contains: `"allowed":true`, procedure: `"procedure":"council.quota.check"`},
		{name: "council quota all stats", method: http.MethodGet, path: "/api/council/quota/stats", contains: `"providers":1`, procedure: `"procedure":"council.quota.allStats"`},
		{name: "council quota provider stats", method: http.MethodGet, path: "/api/council/quota/stats?provider=openai", contains: `"provider":"openai"`, procedure: `"procedure":"council.quota.providerStats"`},
		{name: "council quota get limits", method: http.MethodGet, path: "/api/council/quota/limits?provider=openai", contains: `"rpm":60`, procedure: `"procedure":"council.quota.getLimits"`},
		{name: "council quota set limits", method: http.MethodPost, path: "/api/council/quota/limits?provider=openai", body: `{"limits":{"rpm":120}}`, contains: `"success":true`, procedure: `"procedure":"council.quota.setLimits"`},
		{name: "council quota reset provider", method: http.MethodPost, path: "/api/council/quota/reset?provider=openai", body: `null`, contains: `"success":true`, procedure: `"procedure":"council.quota.resetProvider"`},
		{name: "council quota reset all", method: http.MethodPost, path: "/api/council/quota/reset", body: `null`, contains: `"success":true`, procedure: `"procedure":"council.quota.resetAll"`},
		{name: "council quota unthrottle", method: http.MethodPost, path: "/api/council/quota/unthrottle?provider=openai", body: `null`, contains: `"success":true`, procedure: `"procedure":"council.quota.unthrottle"`},
		{name: "council quota record request", method: http.MethodPost, path: "/api/council/quota/record-request", body: `{"provider":"openai","tokensUsed":10}`, contains: `"success":true`, procedure: `"procedure":"council.quota.recordRequest"`},
		{name: "council quota rate limit error", method: http.MethodPost, path: "/api/council/quota/rate-limit-error", body: `{"provider":"openai"}`, contains: `"success":true`, procedure: `"procedure":"council.quota.recordRateLimitError"`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var body io.Reader
			if tc.body != "" {
				body = strings.NewReader(tc.body)
			}
			request := httptest.NewRequest(tc.method, tc.path, body)
			if tc.body != "" {
				request.Header.Set("content-type", "application/json")
			}
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, request)
			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.contains) {
				t.Fatalf("expected response to contain %s, got %s", tc.contains, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.procedure) {
				t.Fatalf("expected bridge metadata %s, got %s", tc.procedure, recorder.Body.String())
			}
		})
	}
}

func TestDeerFlowBridgeRoutes(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		switch r.URL.Path {
		case "/trpc/deerFlow.status":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{"active": true}}},
			})
		case "/trpc/deerFlow.models":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{"models": []any{"deepseek-chat", "gpt-4.1"}}}},
			})
		case "/trpc/deerFlow.skills":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{"skills": []any{"research", "summarize"}}}},
			})
		case "/trpc/deerFlow.memory":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{"memory": map[string]any{"enabled": true, "entries": 3}}}},
			})
		default:
			t.Fatalf("unexpected upstream path %s", r.URL.Path)
		}
	}))
	defer upstream.Close()

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", upstream.URL+"/trpc")

	cfg := config.Default()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	cases := []struct {
		name      string
		path      string
		contains  string
		procedure string
	}{
		{name: "deerflow status", path: "/api/deerflow/status", contains: `"active":true`, procedure: `"procedure":"deerFlow.status"`},
		{name: "deerflow models", path: "/api/deerflow/models", contains: `"deepseek-chat"`, procedure: `"procedure":"deerFlow.models"`},
		{name: "deerflow skills", path: "/api/deerflow/skills", contains: `"research"`, procedure: `"procedure":"deerFlow.skills"`},
		{name: "deerflow memory", path: "/api/deerflow/memory", contains: `"entries":3`, procedure: `"procedure":"deerFlow.memory"`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, tc.path, nil)
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, request)
			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.contains) {
				t.Fatalf("expected response to contain %s, got %s", tc.contains, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.procedure) {
				t.Fatalf("expected bridge metadata %s, got %s", tc.procedure, recorder.Body.String())
			}
		})
	}
}

func TestHealerBridgeRoutes(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		switch r.URL.Path {
		case "/trpc/healer.diagnose":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"error":"boom"`) {
				t.Fatalf("expected healer.diagnose payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{"summary": "likely quota exhaustion"}}},
			})
		case "/trpc/healer.heal":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"context":"retry with fallback"`) {
				t.Fatalf("expected healer.heal payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}},
			})
		case "/trpc/healer.getHistory":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"error": "boom", "success": true}}}},
			})
		default:
			t.Fatalf("unexpected upstream path %s", r.URL.Path)
		}
	}))
	defer upstream.Close()

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", upstream.URL+"/trpc")

	cfg := config.Default()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	cases := []struct {
		name      string
		method    string
		path      string
		body      string
		contains  string
		procedure string
	}{
		{name: "healer diagnose", method: http.MethodPost, path: "/api/healer/diagnose", body: `{"error":"boom","context":"provider timeout"}`, contains: `"likely quota exhaustion"`, procedure: `"procedure":"healer.diagnose"`},
		{name: "healer heal", method: http.MethodPost, path: "/api/healer/heal", body: `{"error":"boom","context":"retry with fallback"}`, contains: `"success":true`, procedure: `"procedure":"healer.heal"`},
		{name: "healer history", method: http.MethodGet, path: "/api/healer/history", contains: `"error":"boom"`, procedure: `"procedure":"healer.getHistory"`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var body io.Reader
			if tc.body != "" {
				body = strings.NewReader(tc.body)
			}
			request := httptest.NewRequest(tc.method, tc.path, body)
			if tc.body != "" {
				request.Header.Set("content-type", "application/json")
			}
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, request)
			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.contains) {
				t.Fatalf("expected response to contain %s, got %s", tc.contains, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.procedure) {
				t.Fatalf("expected bridge metadata %s, got %s", tc.procedure, recorder.Body.String())
			}
		})
	}
}

func TestHealerHistoryFallsBackToEmptyList(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	cfg := config.Default()
	cfg.WorkspaceRoot = t.TempDir()
	cfg.ConfigDir = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/healer/history", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}

	for _, needle := range []string{
		`"fallback":"go-local-healer"`,
		`"procedure":"healer.getHistory"`,
		`healer history is empty without an active TypeScript healer runtime`,
		`"data":[]`,
	} {
		if !strings.Contains(recorder.Body.String(), needle) {
			t.Fatalf("expected healer history fallback to contain %s, got %s", needle, recorder.Body.String())
		}
	}
}

func TestCouncilVisualBridgeRoutes(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		switch r.URL.Path {
		case "/trpc/council.visual.systemDiagram":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{"mermaid": "graph TD; A-->B;"}}},
			})
		case "/trpc/council.visual.planDiagram":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"task":"Ship it"`) {
				t.Fatalf("expected council.visual.planDiagram payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{"mermaid": "graph TD; Plan-->Done;"}}},
			})
		case "/trpc/council.visual.parsePlan":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"mermaid":"graph TD; A--\u003eB;"`) {
				t.Fatalf("expected council.visual.parsePlan payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{"data": map[string]any{"json": map[string]any{"nodes": []any{map[string]any{"id": "A"}}}}},
			})
		default:
			t.Fatalf("unexpected upstream path %s", r.URL.Path)
		}
	}))
	defer upstream.Close()

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", upstream.URL+"/trpc")

	cfg := config.Default()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	cases := []struct {
		name      string
		method    string
		path      string
		body      string
		contains  string
		procedure string
	}{
		{name: "council visual system diagram", method: http.MethodGet, path: "/api/council/visual/system-diagram", contains: `"graph TD; A--\u003eB;"`, procedure: `"procedure":"council.visual.systemDiagram"`},
		{name: "council visual plan diagram", method: http.MethodPost, path: "/api/council/visual/plan-diagram", body: `{"task":"Ship it"}`, contains: `"graph TD; Plan--\u003eDone;"`, procedure: `"procedure":"council.visual.planDiagram"`},
		{name: "council visual parse plan", method: http.MethodPost, path: "/api/council/visual/parse-plan", body: `{"mermaid":"graph TD; A-->B;"}`, contains: `"id":"A"`, procedure: `"procedure":"council.visual.parsePlan"`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var body io.Reader
			if tc.body != "" {
				body = strings.NewReader(tc.body)
			}
			request := httptest.NewRequest(tc.method, tc.path, body)
			if tc.body != "" {
				request.Header.Set("content-type", "application/json")
			}
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, request)
			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.contains) {
				t.Fatalf("expected response to contain %s, got %s", tc.contains, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.procedure) {
				t.Fatalf("expected bridge metadata %s, got %s", tc.procedure, recorder.Body.String())
			}
		})
	}
}

func TestCloudDevBridgeRoutes(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		switch r.URL.Path {
		case "/trpc/cloudDev.listProviders":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{
				map[string]any{"provider": "jules", "enabled": true, "hasApiKey": true},
			}}}})
		case "/trpc/cloudDev.createSession":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"provider":"jules"`) {
				t.Fatalf("expected cloudDev.createSession payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"id": "cds-1", "status": "pending"}}}})
		case "/trpc/cloudDev.listSessions":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"provider":"jules"`) || !strings.Contains(string(body), `"status":"active"`) {
				t.Fatalf("expected cloudDev.listSessions payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{
				map[string]any{"id": "cds-1", "status": "active"},
			}}}})
		case "/trpc/cloudDev.getSession":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"sessionId":"cds-1"`) {
				t.Fatalf("expected cloudDev.getSession payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"id": "cds-1", "messages": []any{}}}}})
		case "/trpc/cloudDev.updateSessionStatus":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"id": "cds-1", "status": "awaiting_approval"}}}})
		case "/trpc/cloudDev.deleteSession":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}}})
		case "/trpc/cloudDev.sendMessage":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"message": map[string]any{"id": "msg-1"}, "session": map[string]any{"id": "cds-1"}}}}})
		case "/trpc/cloudDev.broadcastMessage":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"delivered": 1, "skipped": 0}}}})
		case "/trpc/cloudDev.previewBroadcastRecipients":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"force":true`) {
				t.Fatalf("expected cloudDev.previewBroadcastRecipients payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"targeted": 1, "sessionIds": []any{"cds-1"}}}}})
		case "/trpc/cloudDev.acceptPlan":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"id": "cds-1", "status": "active"}}}})
		case "/trpc/cloudDev.setAutoAcceptPlan":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"id": "cds-1", "autoAcceptPlan": true}}}})
		case "/trpc/cloudDev.getMessages":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"sessionId":"cds-1"`) || !strings.Contains(string(body), `"limit":50`) {
				t.Fatalf("expected cloudDev.getMessages payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{
				map[string]any{"id": "msg-1", "content": "hello"},
			}}}})
		case "/trpc/cloudDev.getLogs":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"sessionId":"cds-1"`) || !strings.Contains(string(body), `"limit":25`) {
				t.Fatalf("expected cloudDev.getLogs payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{
				map[string]any{"id": "log-1", "message": "created"},
			}}}})
		case "/trpc/cloudDev.stats":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"totalSessions": 1, "enabledProviders": 1}}}})
		default:
			t.Fatalf("unexpected upstream path %s", r.URL.Path)
		}
	}))
	defer upstream.Close()

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", upstream.URL+"/trpc")

	cfg := config.Default()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	cases := []struct {
		name      string
		method    string
		path      string
		body      string
		contains  string
		procedure string
	}{
		{name: "clouddev providers", method: http.MethodGet, path: "/api/clouddev/providers", contains: `"jules"`, procedure: `"procedure":"cloudDev.listProviders"`},
		{name: "clouddev create session", method: http.MethodPost, path: "/api/clouddev/sessions/create", body: `{"provider":"jules","projectName":"tormentnexus","task":"ship","autoAcceptPlan":false}`, contains: `"status":"pending"`, procedure: `"procedure":"cloudDev.createSession"`},
		{name: "clouddev list sessions", method: http.MethodGet, path: "/api/clouddev/sessions?provider=jules&status=active", contains: `"status":"active"`, procedure: `"procedure":"cloudDev.listSessions"`},
		{name: "clouddev get session", method: http.MethodGet, path: "/api/clouddev/sessions/get?sessionId=cds-1", contains: `"id":"cds-1"`, procedure: `"procedure":"cloudDev.getSession"`},
		{name: "clouddev update status", method: http.MethodPost, path: "/api/clouddev/sessions/status", body: `{"sessionId":"cds-1","status":"awaiting_approval"}`, contains: `"awaiting_approval"`, procedure: `"procedure":"cloudDev.updateSessionStatus"`},
		{name: "clouddev delete session", method: http.MethodPost, path: "/api/clouddev/sessions/delete", body: `{"sessionId":"cds-1"}`, contains: `"success":true`, procedure: `"procedure":"cloudDev.deleteSession"`},
		{name: "clouddev send message", method: http.MethodPost, path: "/api/clouddev/messages/send", body: `{"sessionId":"cds-1","content":"hello","force":false}`, contains: `"msg-1"`, procedure: `"procedure":"cloudDev.sendMessage"`},
		{name: "clouddev broadcast", method: http.MethodPost, path: "/api/clouddev/messages/broadcast", body: `{"content":"hello all","force":false}`, contains: `"delivered":1`, procedure: `"procedure":"cloudDev.broadcastMessage"`},
		{name: "clouddev preview recipients", method: http.MethodPost, path: "/api/clouddev/messages/preview-recipients", body: `{"force":true}`, contains: `"targeted":1`, procedure: `"procedure":"cloudDev.previewBroadcastRecipients"`},
		{name: "clouddev accept plan", method: http.MethodPost, path: "/api/clouddev/plan/accept", body: `{"sessionId":"cds-1"}`, contains: `"status":"active"`, procedure: `"procedure":"cloudDev.acceptPlan"`},
		{name: "clouddev auto accept", method: http.MethodPost, path: "/api/clouddev/plan/auto-accept", body: `{"sessionId":"cds-1","enabled":true}`, contains: `"autoAcceptPlan":true`, procedure: `"procedure":"cloudDev.setAutoAcceptPlan"`},
		{name: "clouddev get messages", method: http.MethodGet, path: "/api/clouddev/messages/get?sessionId=cds-1&limit=50", contains: `"hello"`, procedure: `"procedure":"cloudDev.getMessages"`},
		{name: "clouddev get logs", method: http.MethodGet, path: "/api/clouddev/logs?sessionId=cds-1&limit=25", contains: `"created"`, procedure: `"procedure":"cloudDev.getLogs"`},
		{name: "clouddev stats", method: http.MethodGet, path: "/api/clouddev/stats", contains: `"totalSessions":1`, procedure: `"procedure":"cloudDev.stats"`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var body io.Reader
			if tc.body != "" {
				body = strings.NewReader(tc.body)
			}
			request := httptest.NewRequest(tc.method, tc.path, body)
			if tc.body != "" {
				request.Header.Set("content-type", "application/json")
			}
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, request)
			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.contains) {
				t.Fatalf("expected response to contain %s, got %s", tc.contains, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.procedure) {
				t.Fatalf("expected bridge metadata %s, got %s", tc.procedure, recorder.Body.String())
			}
		})
	}
}

func TestConfigRouterBridgeRoutes(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		switch r.URL.Path {
		case "/trpc/config.list":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{
				map[string]any{"key": "theme", "value": "dark"},
			}}}})
		case "/trpc/config.get":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"key":"theme"`) {
				t.Fatalf("expected config.get payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": "dark"}}})
		case "/trpc/config.upsert":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"id":"theme"`) || !strings.Contains(string(body), `"value":"light"`) {
				t.Fatalf("expected config.upsert payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": true}}})
		case "/trpc/config.delete":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"key":"theme"`) {
				t.Fatalf("expected config.delete payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": true}}})
		case "/trpc/config.update":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"key":"theme"`) || !strings.Contains(string(body), `"value":"light"`) {
				t.Fatalf("expected config.update payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": true}}})
		case "/trpc/config.getMcpTimeout":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": 15000}}})
		case "/trpc/config.setMcpTimeout":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"timeout":20000`) {
				t.Fatalf("expected config.setMcpTimeout payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}}})
		case "/trpc/config.getMcpMaxAttempts":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": 4}}})
		case "/trpc/config.setMcpMaxAttempts":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"maxAttempts":5`) {
				t.Fatalf("expected config.setMcpMaxAttempts payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}}})
		case "/trpc/config.getMcpMaxTotalTimeout":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": 30000}}})
		case "/trpc/config.setMcpMaxTotalTimeout":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"timeout":45000`) {
				t.Fatalf("expected config.setMcpMaxTotalTimeout payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}}})
		case "/trpc/config.getMcpResetTimeoutOnProgress":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": true}}})
		case "/trpc/config.setMcpResetTimeoutOnProgress":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"enabled":false`) {
				t.Fatalf("expected config.setMcpResetTimeoutOnProgress payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}}})
		case "/trpc/config.getSessionLifetime":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": 3600000}}})
		case "/trpc/config.setSessionLifetime":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"lifetime":7200000`) {
				t.Fatalf("expected config.setSessionLifetime payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}}})
		case "/trpc/config.getSignupDisabled":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": false}}})
		case "/trpc/config.setSignupDisabled":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"disabled":true`) {
				t.Fatalf("expected config.setSignupDisabled payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}}})
		case "/trpc/config.getSsoSignupDisabled":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": true}}})
		case "/trpc/config.setSsoSignupDisabled":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"disabled":false`) {
				t.Fatalf("expected config.setSsoSignupDisabled payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}}})
		case "/trpc/config.getBasicAuthDisabled":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": false}}})
		case "/trpc/config.setBasicAuthDisabled":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"disabled":true`) {
				t.Fatalf("expected config.setBasicAuthDisabled payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}}})
		case "/trpc/config.getAuthProviders":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{"github", "google"}}}})
		case "/trpc/config.getAlwaysVisibleTools":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{"list_all_tools", "search_tools"}}}})
		case "/trpc/config.setAlwaysVisibleTools":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"tools":["list_all_tools","search_tools"]`) {
				t.Fatalf("expected config.setAlwaysVisibleTools payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true, "tools": []any{"list_all_tools", "search_tools"}}}}})
		default:
			t.Fatalf("unexpected upstream path %s", r.URL.Path)
		}
	}))
	defer upstream.Close()

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", upstream.URL+"/trpc")

	cfg := config.Default()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	cases := []struct {
		name      string
		method    string
		path      string
		body      string
		contains  string
		procedure string
	}{
		{name: "config list", method: http.MethodGet, path: "/api/config/list", contains: `"theme"`, procedure: `"procedure":"config.list"`},
		{name: "config get", method: http.MethodGet, path: "/api/config/get?key=theme", contains: `"data":"dark"`, procedure: `"procedure":"config.get"`},
		{name: "config upsert", method: http.MethodPost, path: "/api/config/upsert", body: `{"id":"theme","value":"light"}`, contains: `"data":true`, procedure: `"procedure":"config.upsert"`},
		{name: "config delete", method: http.MethodPost, path: "/api/config/delete", body: `{"key":"theme"}`, contains: `"data":true`, procedure: `"procedure":"config.delete"`},
		{name: "config update", method: http.MethodPost, path: "/api/config/update", body: `{"key":"theme","value":"light"}`, contains: `"data":true`, procedure: `"procedure":"config.update"`},
		{name: "config get mcp timeout", method: http.MethodGet, path: "/api/config/mcp-timeout", contains: `"data":15000`, procedure: `"procedure":"config.getMcpTimeout"`},
		{name: "config set mcp timeout", method: http.MethodPost, path: "/api/config/mcp-timeout/set", body: `{"timeout":20000}`, contains: `"success":true`, procedure: `"procedure":"config.setMcpTimeout"`},
		{name: "config get mcp max attempts", method: http.MethodGet, path: "/api/config/mcp-max-attempts", contains: `"data":4`, procedure: `"procedure":"config.getMcpMaxAttempts"`},
		{name: "config set mcp max attempts", method: http.MethodPost, path: "/api/config/mcp-max-attempts/set", body: `{"maxAttempts":5}`, contains: `"success":true`, procedure: `"procedure":"config.setMcpMaxAttempts"`},
		{name: "config get mcp max total timeout", method: http.MethodGet, path: "/api/config/mcp-max-total-timeout", contains: `"data":30000`, procedure: `"procedure":"config.getMcpMaxTotalTimeout"`},
		{name: "config set mcp max total timeout", method: http.MethodPost, path: "/api/config/mcp-max-total-timeout/set", body: `{"timeout":45000}`, contains: `"success":true`, procedure: `"procedure":"config.setMcpMaxTotalTimeout"`},
		{name: "config get mcp reset on progress", method: http.MethodGet, path: "/api/config/mcp-reset-timeout-on-progress", contains: `"data":true`, procedure: `"procedure":"config.getMcpResetTimeoutOnProgress"`},
		{name: "config set mcp reset on progress", method: http.MethodPost, path: "/api/config/mcp-reset-timeout-on-progress/set", body: `{"enabled":false}`, contains: `"success":true`, procedure: `"procedure":"config.setMcpResetTimeoutOnProgress"`},
		{name: "config get session lifetime", method: http.MethodGet, path: "/api/config/session-lifetime", contains: `"data":3600000`, procedure: `"procedure":"config.getSessionLifetime"`},
		{name: "config set session lifetime", method: http.MethodPost, path: "/api/config/session-lifetime/set", body: `{"lifetime":7200000}`, contains: `"success":true`, procedure: `"procedure":"config.setSessionLifetime"`},
		{name: "config get signup disabled", method: http.MethodGet, path: "/api/config/signup-disabled", contains: `"data":false`, procedure: `"procedure":"config.getSignupDisabled"`},
		{name: "config set signup disabled", method: http.MethodPost, path: "/api/config/signup-disabled/set", body: `{"disabled":true}`, contains: `"success":true`, procedure: `"procedure":"config.setSignupDisabled"`},
		{name: "config get sso signup disabled", method: http.MethodGet, path: "/api/config/sso-signup-disabled", contains: `"data":true`, procedure: `"procedure":"config.getSsoSignupDisabled"`},
		{name: "config set sso signup disabled", method: http.MethodPost, path: "/api/config/sso-signup-disabled/set", body: `{"disabled":false}`, contains: `"success":true`, procedure: `"procedure":"config.setSsoSignupDisabled"`},
		{name: "config get basic auth disabled", method: http.MethodGet, path: "/api/config/basic-auth-disabled", contains: `"data":false`, procedure: `"procedure":"config.getBasicAuthDisabled"`},
		{name: "config set basic auth disabled", method: http.MethodPost, path: "/api/config/basic-auth-disabled/set", body: `{"disabled":true}`, contains: `"success":true`, procedure: `"procedure":"config.setBasicAuthDisabled"`},
		{name: "config get auth providers", method: http.MethodGet, path: "/api/config/auth-providers", contains: `"github"`, procedure: `"procedure":"config.getAuthProviders"`},
		{name: "config get always visible tools", method: http.MethodGet, path: "/api/config/always-visible-tools", contains: `"list_all_tools"`, procedure: `"procedure":"config.getAlwaysVisibleTools"`},
		{name: "config set always visible tools", method: http.MethodPost, path: "/api/config/always-visible-tools/set", body: `{"tools":["list_all_tools","search_tools"]}`, contains: `"success":true`, procedure: `"procedure":"config.setAlwaysVisibleTools"`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var body io.Reader
			if tc.body != "" {
				body = strings.NewReader(tc.body)
			}
			request := httptest.NewRequest(tc.method, tc.path, body)
			if tc.body != "" {
				request.Header.Set("content-type", "application/json")
			}
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, request)
			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.contains) {
				t.Fatalf("expected response to contain %s, got %s", tc.contains, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.procedure) {
				t.Fatalf("expected bridge metadata %s, got %s", tc.procedure, recorder.Body.String())
			}
		})
	}
}

func TestCouncilHistoryBridgeRoutes(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		switch r.URL.Path {
		case "/trpc/council.history.status":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"enabled": true, "recordCount": 4}}}})
		case "/trpc/council.history.getConfig":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"enabled": true, "retentionDays": 30}}}})
		case "/trpc/council.history.updateConfig":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"retentionDays":14`) {
				t.Fatalf("expected council.history.updateConfig payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"enabled": true, "retentionDays": 14}}}})
		case "/trpc/council.history.toggle":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"enabled": false}}}})
		case "/trpc/council.history.stats":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"averageConsensus": 0.9}}}})
		case "/trpc/council.history.list":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"sessionId":"sess-1"`) || !strings.Contains(string(body), `"approved":true`) || !strings.Contains(string(body), `"minConsensus":0.5`) {
				t.Fatalf("expected council.history.list payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{
				"records": []any{map[string]any{"id": "deb-1"}},
				"meta":    map[string]any{"count": 1, "totalRecords": 4},
			}}}})
		case "/trpc/council.history.get":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"id":"deb-1"`) {
				t.Fatalf("expected council.history.get payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"id": "deb-1", "approved": true}}}})
		case "/trpc/council.history.delete":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"deleted": true, "id": "deb-1"}}}})
		case "/trpc/council.history.supervisorHistory":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"name":"planner"`) {
				t.Fatalf("expected council.history.supervisorHistory payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"decision": "approve"}}}}})
		case "/trpc/council.history.clear":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"cleared": 4}}}})
		case "/trpc/council.history.initialize":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"initialized": true, "recordCount": 4}}}})
		default:
			t.Fatalf("unexpected upstream path %s", r.URL.Path)
		}
	}))
	defer upstream.Close()

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", upstream.URL+"/trpc")

	cfg := config.Default()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	cases := []struct {
		name      string
		method    string
		path      string
		body      string
		contains  string
		procedure string
	}{
		{name: "council history status", method: http.MethodGet, path: "/api/council/history/status", contains: `"recordCount":4`, procedure: `"procedure":"council.history.status"`},
		{name: "council history get config", method: http.MethodGet, path: "/api/council/history/config", contains: `"retentionDays":30`, procedure: `"procedure":"council.history.getConfig"`},
		{name: "council history update config", method: http.MethodPost, path: "/api/council/history/config", body: `{"retentionDays":14}`, contains: `"retentionDays":14`, procedure: `"procedure":"council.history.updateConfig"`},
		{name: "council history toggle", method: http.MethodPost, path: "/api/council/history/toggle", body: `{"enabled":false}`, contains: `"enabled":false`, procedure: `"procedure":"council.history.toggle"`},
		{name: "council history stats", method: http.MethodGet, path: "/api/council/history/stats", contains: `"averageConsensus":0.9`, procedure: `"procedure":"council.history.stats"`},
		{name: "council history list", method: http.MethodGet, path: "/api/council/history/list?sessionId=sess-1&approved=true&minConsensus=0.5&limit=10&offset=0&sortBy=timestamp&sortOrder=desc", contains: `"deb-1"`, procedure: `"procedure":"council.history.list"`},
		{name: "council history get", method: http.MethodGet, path: "/api/council/history/get?id=deb-1", contains: `"approved":true`, procedure: `"procedure":"council.history.get"`},
		{name: "council history delete", method: http.MethodPost, path: "/api/council/history/delete", body: `{"id":"deb-1"}`, contains: `"deleted":true`, procedure: `"procedure":"council.history.delete"`},
		{name: "council history supervisor", method: http.MethodGet, path: "/api/council/history/supervisor?name=planner", contains: `"approve"`, procedure: `"procedure":"council.history.supervisorHistory"`},
		{name: "council history clear", method: http.MethodPost, path: "/api/council/history/clear", contains: `"cleared":4`, procedure: `"procedure":"council.history.clear"`},
		{name: "council history initialize", method: http.MethodPost, path: "/api/council/history/initialize", contains: `"initialized":true`, procedure: `"procedure":"council.history.initialize"`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var body io.Reader
			if tc.body != "" {
				body = strings.NewReader(tc.body)
			}
			request := httptest.NewRequest(tc.method, tc.path, body)
			if tc.body != "" {
				request.Header.Set("content-type", "application/json")
			}
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, request)
			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.contains) {
				t.Fatalf("expected response to contain %s, got %s", tc.contains, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.procedure) {
				t.Fatalf("expected bridge metadata %s, got %s", tc.procedure, recorder.Body.String())
			}
		})
	}
}

func TestCouncilBaseBridgeRoutes(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		switch r.URL.Path {
		case "/trpc/council.status":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"enabled": true, "supervisorCount": 2}}}})
		case "/trpc/council.updateConfig":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"debateRounds":3`) {
				t.Fatalf("expected council.updateConfig payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"debateRounds": 3}}}})
		case "/trpc/council.addSupervisors":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"name":"planner"`) {
				t.Fatalf("expected council.addSupervisors payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"added": []any{"planner"}, "failed": []any{}}}}})
		case "/trpc/council.clearSupervisors":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}}})
		case "/trpc/council.debate":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"description":"Ship feature"`) {
				t.Fatalf("expected council.debate payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"decision": "approve"}}}})
		case "/trpc/council.toggle":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"enabled": false}}}})
		case "/trpc/council.addMock":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"added": "MockSupervisor-1"}}}})
		default:
			t.Fatalf("unexpected upstream path %s", r.URL.Path)
		}
	}))
	defer upstream.Close()

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", upstream.URL+"/trpc")

	cfg := config.Default()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	cases := []struct {
		name      string
		method    string
		path      string
		body      string
		contains  string
		procedure string
	}{
		{name: "council base status", method: http.MethodGet, path: "/api/council/status", contains: `"supervisorCount":2`, procedure: `"procedure":"council.status"`},
		{name: "council base update config", method: http.MethodPost, path: "/api/council/config/update", body: `{"debateRounds":3}`, contains: `"debateRounds":3`, procedure: `"procedure":"council.updateConfig"`},
		{name: "council base add supervisors", method: http.MethodPost, path: "/api/council/supervisors/add", body: `{"supervisors":[{"name":"planner","provider":"openai"}]}`, contains: `"planner"`, procedure: `"procedure":"council.addSupervisors"`},
		{name: "council base clear supervisors", method: http.MethodPost, path: "/api/council/supervisors/clear", contains: `"success":true`, procedure: `"procedure":"council.clearSupervisors"`},
		{name: "council base debate", method: http.MethodPost, path: "/api/council/debate", body: `{"id":"task-1","description":"Ship feature","context":"ctx","files":["README.md"]}`, contains: `"approve"`, procedure: `"procedure":"council.debate"`},
		{name: "council base toggle", method: http.MethodPost, path: "/api/council/toggle", contains: `"enabled":false`, procedure: `"procedure":"council.toggle"`},
		{name: "council base add mock", method: http.MethodPost, path: "/api/council/mock/add", contains: `"MockSupervisor-1"`, procedure: `"procedure":"council.addMock"`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var body io.Reader
			if tc.body != "" {
				body = strings.NewReader(tc.body)
			}
			request := httptest.NewRequest(tc.method, tc.path, body)
			if tc.body != "" {
				request.Header.Set("content-type", "application/json")
			}
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, request)
			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.contains) {
				t.Fatalf("expected response to contain %s, got %s", tc.contains, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.procedure) {
				t.Fatalf("expected bridge metadata %s, got %s", tc.procedure, recorder.Body.String())
			}
		})
	}
}

func TestCouncilSmartPilotBridgeRoutes(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		switch r.URL.Path {
		case "/trpc/council.smartPilot.status":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"enabled": true, "activePlans": []any{map[string]any{"sessionId": "sess-1"}}}}}})
		case "/trpc/council.smartPilot.getConfig":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"enabled": true, "autoApproveThreshold": 0.8}}}})
		case "/trpc/council.smartPilot.updateConfig":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"enabled":true`) {
				t.Fatalf("expected council.smartPilot.updateConfig payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"enabled": true}}}})
		case "/trpc/council.smartPilot.trigger":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"sessionId":"sess-1"`) {
				t.Fatalf("expected council.smartPilot.trigger payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}}})
		case "/trpc/council.smartPilot.resetCount":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"sessionId":"sess-1"`) {
				t.Fatalf("expected council.smartPilot.resetCount payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}}})
		case "/trpc/council.smartPilot.resetAllCounts":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}}})
		default:
			t.Fatalf("unexpected upstream path %s", r.URL.Path)
		}
	}))
	defer upstream.Close()

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", upstream.URL+"/trpc")

	cfg := config.Default()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	cases := []struct {
		name      string
		method    string
		path      string
		body      string
		contains  string
		procedure string
	}{
		{name: "smartpilot status", method: http.MethodGet, path: "/api/council/smart-pilot/status", contains: `"sessionId":"sess-1"`, procedure: `"procedure":"council.smartPilot.status"`},
		{name: "smartpilot get config", method: http.MethodGet, path: "/api/council/smart-pilot/config", contains: `"autoApproveThreshold":0.8`, procedure: `"procedure":"council.smartPilot.getConfig"`},
		{name: "smartpilot update config", method: http.MethodPost, path: "/api/council/smart-pilot/config", body: `{"enabled":true}`, contains: `"enabled":true`, procedure: `"procedure":"council.smartPilot.updateConfig"`},
		{name: "smartpilot trigger", method: http.MethodPost, path: "/api/council/smart-pilot/trigger", body: `{"sessionId":"sess-1","task":{"id":"task-1"}}`, contains: `"success":true`, procedure: `"procedure":"council.smartPilot.trigger"`},
		{name: "smartpilot reset count", method: http.MethodPost, path: "/api/council/smart-pilot/reset-count", body: `{"sessionId":"sess-1"}`, contains: `"success":true`, procedure: `"procedure":"council.smartPilot.resetCount"`},
		{name: "smartpilot reset all", method: http.MethodPost, path: "/api/council/smart-pilot/reset-all", contains: `"success":true`, procedure: `"procedure":"council.smartPilot.resetAllCounts"`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var body io.Reader
			if tc.body != "" {
				body = strings.NewReader(tc.body)
			}
			request := httptest.NewRequest(tc.method, tc.path, body)
			if tc.body != "" {
				request.Header.Set("content-type", "application/json")
			}
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, request)
			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.contains) {
				t.Fatalf("expected response to contain %s, got %s", tc.contains, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.procedure) {
				t.Fatalf("expected bridge metadata %s, got %s", tc.procedure, recorder.Body.String())
			}
		})
	}
}

func TestCouncilHooksBridgeRoutes(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		switch r.URL.Path {
		case "/trpc/council.hooks.list":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"hooks": []any{map[string]any{"id": "hook-1", "phase": "pre-debate"}}}}}})
		case "/trpc/council.hooks.register":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"phase":"pre-debate"`) {
				t.Fatalf("expected council.hooks.register payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true, "hookId": "hook-1"}}}})
		case "/trpc/council.hooks.unregister":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"id":"hook-1"`) {
				t.Fatalf("expected council.hooks.unregister payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}}})
		case "/trpc/council.hooks.clear":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}}})
		default:
			t.Fatalf("unexpected upstream path %s", r.URL.Path)
		}
	}))
	defer upstream.Close()

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", upstream.URL+"/trpc")

	cfg := config.Default()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	cases := []struct {
		name      string
		method    string
		path      string
		body      string
		contains  string
		procedure string
	}{
		{name: "hooks list", method: http.MethodGet, path: "/api/council/hooks", contains: `"id":"hook-1"`, procedure: `"procedure":"council.hooks.list"`},
		{name: "hooks register", method: http.MethodPost, path: "/api/council/hooks/register", body: `{"phase":"pre-debate","webhookUrl":"https://example.com/hook","priority":5}`, contains: `"hookId":"hook-1"`, procedure: `"procedure":"council.hooks.register"`},
		{name: "hooks unregister", method: http.MethodPost, path: "/api/council/hooks/unregister", body: `{"id":"hook-1"}`, contains: `"success":true`, procedure: `"procedure":"council.hooks.unregister"`},
		{name: "hooks clear", method: http.MethodPost, path: "/api/council/hooks/clear", contains: `"success":true`, procedure: `"procedure":"council.hooks.clear"`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var body io.Reader
			if tc.body != "" {
				body = strings.NewReader(tc.body)
			}
			request := httptest.NewRequest(tc.method, tc.path, body)
			if tc.body != "" {
				request.Header.Set("content-type", "application/json")
			}
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, request)
			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.contains) {
				t.Fatalf("expected response to contain %s, got %s", tc.contains, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.procedure) {
				t.Fatalf("expected bridge metadata %s, got %s", tc.procedure, recorder.Body.String())
			}
		})
	}
}

func TestCouncilIDEBridgeRoutes(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		switch r.URL.Path {
		case "/trpc/council.ide.status":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true, "ready": true, "capabilities": []any{"council", "smart-pilot"}}}}})
		case "/trpc/council.ide.submitTask":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"description":"Fix the failing test"`) {
				t.Fatalf("expected council.ide.submitTask payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true, "taskId": "ide-1", "sessionId": "sess-1"}}}})
		default:
			t.Fatalf("unexpected upstream path %s", r.URL.Path)
		}
	}))
	defer upstream.Close()

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", upstream.URL+"/trpc")

	cfg := config.Default()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	cases := []struct {
		name      string
		method    string
		path      string
		body      string
		contains  string
		procedure string
	}{
		{name: "ide status", method: http.MethodGet, path: "/api/council/ide/status", contains: `"ready":true`, procedure: `"procedure":"council.ide.status"`},
		{name: "ide submit task", method: http.MethodPost, path: "/api/council/ide/submit-task", body: `{"description":"Fix the failing test","fileContext":{"path":"src/app.ts","content":"console.log('hi')"}}`, contains: `"taskId":"ide-1"`, procedure: `"procedure":"council.ide.submitTask"`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var body io.Reader
			if tc.body != "" {
				body = strings.NewReader(tc.body)
			}
			request := httptest.NewRequest(tc.method, tc.path, body)
			if tc.body != "" {
				request.Header.Set("content-type", "application/json")
			}
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, request)
			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.contains) {
				t.Fatalf("expected response to contain %s, got %s", tc.contains, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.procedure) {
				t.Fatalf("expected bridge metadata %s, got %s", tc.procedure, recorder.Body.String())
			}
		})
	}
}

func TestCouncilEvolutionBridgeRoutes(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		switch r.URL.Path {
		case "/trpc/council.evolution.start":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true, "message": "Continuous learning started"}}}})
		case "/trpc/council.evolution.stop":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true, "message": "Continuous learning stopped"}}}})
		case "/trpc/council.evolution.optimize":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}}})
		case "/trpc/council.evolution.evolve":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"description":"Improve routing heuristics"`) {
				t.Fatalf("expected council.evolution.evolve payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true, "sessionId": "evolve-1"}}}})
		case "/trpc/council.evolution.test":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true, "passed": true}}}})
		default:
			t.Fatalf("unexpected upstream path %s", r.URL.Path)
		}
	}))
	defer upstream.Close()

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", upstream.URL+"/trpc")

	cfg := config.Default()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	cases := []struct {
		name      string
		method    string
		path      string
		body      string
		contains  string
		procedure string
	}{
		{name: "evolution start", method: http.MethodPost, path: "/api/council/evolution/start", contains: `"Continuous learning started"`, procedure: `"procedure":"council.evolution.start"`},
		{name: "evolution stop", method: http.MethodPost, path: "/api/council/evolution/stop", contains: `"Continuous learning stopped"`, procedure: `"procedure":"council.evolution.stop"`},
		{name: "evolution optimize", method: http.MethodPost, path: "/api/council/evolution/optimize", contains: `"success":true`, procedure: `"procedure":"council.evolution.optimize"`},
		{name: "evolution evolve", method: http.MethodPost, path: "/api/council/evolution/evolve", body: `{"description":"Improve routing heuristics"}`, contains: `"sessionId":"evolve-1"`, procedure: `"procedure":"council.evolution.evolve"`},
		{name: "evolution test", method: http.MethodGet, path: "/api/council/evolution/test", contains: `"passed":true`, procedure: `"procedure":"council.evolution.test"`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var body io.Reader
			if tc.body != "" {
				body = strings.NewReader(tc.body)
			}
			request := httptest.NewRequest(tc.method, tc.path, body)
			if tc.body != "" {
				request.Header.Set("content-type", "application/json")
			}
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, request)
			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.contains) {
				t.Fatalf("expected response to contain %s, got %s", tc.contains, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.procedure) {
				t.Fatalf("expected bridge metadata %s, got %s", tc.procedure, recorder.Body.String())
			}
		})
	}
}

func TestCouncilFineTuneBridgeRoutes(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		switch r.URL.Path {
		case "/trpc/council.fineTune.createDataset":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"name":"dataset-1"`) {
				t.Fatalf("expected council.fineTune.createDataset payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"id": "dataset-1", "name": "dataset-1"}}}})
		case "/trpc/council.fineTune.listDatasets":
			body, _ := io.ReadAll(r.Body)
			if len(body) > 0 && !strings.Contains(string(body), `"taskType":"code"`) {
				t.Fatalf("expected council.fineTune.listDatasets payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"id": "dataset-1"}}}}})
		case "/trpc/council.fineTune.getDataset":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"id":"dataset-1"`) {
				t.Fatalf("expected council.fineTune.getDataset payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"id": "dataset-1"}}}})
		case "/trpc/council.fineTune.createJob":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"datasetId":"dataset-1"`) {
				t.Fatalf("expected council.fineTune.createJob payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"id": "job-1"}}}})
		case "/trpc/council.fineTune.listJobs":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"id": "job-1"}}}}})
		case "/trpc/council.fineTune.startJob":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"id":"job-1"`) {
				t.Fatalf("expected council.fineTune.startJob payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"id": "job-1", "status": "running"}}}})
		case "/trpc/council.fineTune.registerModel":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"name":"model-1"`) {
				t.Fatalf("expected council.fineTune.registerModel payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"id": "model-1"}}}})
		case "/trpc/council.fineTune.listModels":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"id": "model-1"}}}}})
		case "/trpc/council.fineTune.deployModel":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"id":"model-1"`) {
				t.Fatalf("expected council.fineTune.deployModel payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"id": "model-1", "deploymentStatus": "active"}}}})
		case "/trpc/council.fineTune.chat":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"id":"model-1"`) {
				t.Fatalf("expected council.fineTune.chat payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"response": "hello from model"}}}})
		case "/trpc/council.fineTune.stats":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"datasets": 1, "jobs": 1, "models": 1}}}})
		default:
			t.Fatalf("unexpected upstream path %s", r.URL.Path)
		}
	}))
	defer upstream.Close()

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", upstream.URL+"/trpc")

	cfg := config.Default()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	cases := []struct {
		name      string
		method    string
		path      string
		body      string
		contains  string
		procedure string
	}{
		{name: "finetune create dataset", method: http.MethodPost, path: "/api/council/fine-tune/datasets", body: `{"name":"dataset-1","format":"alpaca"}`, contains: `"dataset-1"`, procedure: `"procedure":"council.fineTune.createDataset"`},
		{name: "finetune list datasets", method: http.MethodGet, path: "/api/council/fine-tune/datasets?taskType=code", contains: `"dataset-1"`, procedure: `"procedure":"council.fineTune.listDatasets"`},
		{name: "finetune get dataset", method: http.MethodGet, path: "/api/council/fine-tune/datasets/get?id=dataset-1", contains: `"dataset-1"`, procedure: `"procedure":"council.fineTune.getDataset"`},
		{name: "finetune create job", method: http.MethodPost, path: "/api/council/fine-tune/jobs", body: `{"baseModel":"gpt-4.1","datasetId":"dataset-1"}`, contains: `"job-1"`, procedure: `"procedure":"council.fineTune.createJob"`},
		{name: "finetune list jobs", method: http.MethodGet, path: "/api/council/fine-tune/jobs", contains: `"job-1"`, procedure: `"procedure":"council.fineTune.listJobs"`},
		{name: "finetune start job", method: http.MethodPost, path: "/api/council/fine-tune/jobs/start", body: `{"id":"job-1"}`, contains: `"running"`, procedure: `"procedure":"council.fineTune.startJob"`},
		{name: "finetune register model", method: http.MethodPost, path: "/api/council/fine-tune/models", body: `{"name":"model-1","provider":"openai","providerModelId":"ft:model-1"}`, contains: `"model-1"`, procedure: `"procedure":"council.fineTune.registerModel"`},
		{name: "finetune list models", method: http.MethodGet, path: "/api/council/fine-tune/models", contains: `"model-1"`, procedure: `"procedure":"council.fineTune.listModels"`},
		{name: "finetune deploy model", method: http.MethodPost, path: "/api/council/fine-tune/models/deploy", body: `{"id":"model-1"}`, contains: `"active"`, procedure: `"procedure":"council.fineTune.deployModel"`},
		{name: "finetune chat", method: http.MethodPost, path: "/api/council/fine-tune/chat", body: `{"id":"model-1","messages":[{"role":"user","content":"hello"}]}`, contains: `"hello from model"`, procedure: `"procedure":"council.fineTune.chat"`},
		{name: "finetune stats", method: http.MethodGet, path: "/api/council/fine-tune/stats", contains: `"datasets":1`, procedure: `"procedure":"council.fineTune.stats"`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var body io.Reader
			if tc.body != "" {
				body = strings.NewReader(tc.body)
			}
			request := httptest.NewRequest(tc.method, tc.path, body)
			if tc.body != "" {
				request.Header.Set("content-type", "application/json")
			}
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, request)
			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.contains) {
				t.Fatalf("expected response to contain %s, got %s", tc.contains, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.procedure) {
				t.Fatalf("expected bridge metadata %s, got %s", tc.procedure, recorder.Body.String())
			}
		})
	}
}

func TestCouncilRotationBridgeRoutes(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		switch r.URL.Path {
		case "/trpc/council.rotation.list":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"id": "rotation-1"}}}}})
		case "/trpc/council.rotation.get":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"roomId":"rotation-1"`) {
				t.Fatalf("expected council.rotation.get payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"id": "rotation-1", "mode": "plan"}}}})
		case "/trpc/council.rotation.create":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"title":"Autopilot trio"`) {
				t.Fatalf("expected council.rotation.create payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"id": "rotation-1", "mode": "plan"}}}})
		case "/trpc/council.rotation.addParticipant":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"id":"qwen"`) {
				t.Fatalf("expected council.rotation.addParticipant payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"id": "rotation-1", "participantOrder": []any{"claude", "gpt", "qwen"}}}}})
		case "/trpc/council.rotation.postMessage":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"content":"Plan first"`) {
				t.Fatalf("expected council.rotation.postMessage payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"id": "msg-1", "content": "Plan first"}}}})
		case "/trpc/council.rotation.setAgreement":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"agrees":true`) {
				t.Fatalf("expected council.rotation.setAgreement payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"id": "rotation-1", "status": "active"}}}})
		case "/trpc/council.rotation.advanceTurn":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"summary":"Planning complete"`) {
				t.Fatalf("expected council.rotation.advanceTurn payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"id": "rotation-1", "currentCycleNumber": 1}}}})
		case "/trpc/council.rotation.configureSupervisor":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"completionCriteria":"Wait for tests"`) {
				t.Fatalf("expected council.rotation.configureSupervisor payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"id": "rotation-1", "supervisor": map[string]any{"status": "active"}}}}})
		case "/trpc/council.rotation.runSupervisorCheck":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"roomId":"rotation-1"`) {
				t.Fatalf("expected council.rotation.runSupervisorCheck payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"id": "rotation-1", "supervisor": map[string]any{"status": "satisfied"}}}}})
		case "/trpc/council.rotation.updateSharedContext":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"sharedContext":"Shared task context"`) {
				t.Fatalf("expected council.rotation.updateSharedContext payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"id": "rotation-1", "sharedContext": "Shared task context"}}}})
		case "/trpc/council.rotation.pause":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"id": "rotation-1", "status": "paused"}}}})
		case "/trpc/council.rotation.resume":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"id": "rotation-1", "status": "active"}}}})
		case "/trpc/council.rotation.startExecution":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"checkpoint":"Ready to code"`) {
				t.Fatalf("expected council.rotation.startExecution payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"id": "rotation-1", "mode": "execute"}}}})
		case "/trpc/council.rotation.complete":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"summary":"Done"`) {
				t.Fatalf("expected council.rotation.complete payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"id": "rotation-1", "status": "completed"}}}})
		default:
			t.Fatalf("unexpected upstream path %s", r.URL.Path)
		}
	}))
	defer upstream.Close()

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", upstream.URL+"/trpc")

	cfg := config.Default()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	cases := []struct {
		name      string
		method    string
		path      string
		body      string
		contains  string
		procedure string
	}{
		{name: "rotation list", method: http.MethodGet, path: "/api/council/rotation", contains: `"rotation-1"`, procedure: `"procedure":"council.rotation.list"`},
		{name: "rotation get", method: http.MethodGet, path: "/api/council/rotation/get?roomId=rotation-1", contains: `"mode":"plan"`, procedure: `"procedure":"council.rotation.get"`},
		{name: "rotation create", method: http.MethodPost, path: "/api/council/rotation/create", body: `{"title":"Autopilot trio","objective":"Debate then code","participants":[{"id":"claude","name":"Claude","provider":"anthropic","model":"claude-sonnet-4.6"},{"id":"gpt","name":"GPT","provider":"openai","model":"gpt-5.4"}]}`, contains: `"rotation-1"`, procedure: `"procedure":"council.rotation.create"`},
		{name: "rotation add participant", method: http.MethodPost, path: "/api/council/rotation/add-participant", body: `{"roomId":"rotation-1","participant":{"id":"qwen","name":"Qwen","provider":"local","model":"qwen2.5-coder"}}`, contains: `"qwen"`, procedure: `"procedure":"council.rotation.addParticipant"`},
		{name: "rotation post message", method: http.MethodPost, path: "/api/council/rotation/post-message", body: `{"roomId":"rotation-1","participantId":"claude","content":"Plan first"}`, contains: `"Plan first"`, procedure: `"procedure":"council.rotation.postMessage"`},
		{name: "rotation set agreement", method: http.MethodPost, path: "/api/council/rotation/set-agreement", body: `{"roomId":"rotation-1","participantId":"claude","agrees":true}`, contains: `"status":"active"`, procedure: `"procedure":"council.rotation.setAgreement"`},
		{name: "rotation advance turn", method: http.MethodPost, path: "/api/council/rotation/advance-turn", body: `{"roomId":"rotation-1","summary":"Planning complete"}`, contains: `"currentCycleNumber":1`, procedure: `"procedure":"council.rotation.advanceTurn"`},
		{name: "rotation configure supervisor", method: http.MethodPost, path: "/api/council/rotation/configure-supervisor", body: `{"roomId":"rotation-1","supervisor":{"name":"Local Qwen supervisor","provider":"local","model":"qwen2.5-coder","completionCriteria":"Wait for tests","evaluationMode":"after_turn","completionAction":"complete"}}`, contains: `"status":"active"`, procedure: `"procedure":"council.rotation.configureSupervisor"`},
		{name: "rotation run supervisor check", method: http.MethodPost, path: "/api/council/rotation/run-supervisor-check", body: `{"roomId":"rotation-1"}`, contains: `"status":"satisfied"`, procedure: `"procedure":"council.rotation.runSupervisorCheck"`},
		{name: "rotation update context", method: http.MethodPost, path: "/api/council/rotation/update-shared-context", body: `{"roomId":"rotation-1","sharedContext":"Shared task context"}`, contains: `"Shared task context"`, procedure: `"procedure":"council.rotation.updateSharedContext"`},
		{name: "rotation pause", method: http.MethodPost, path: "/api/council/rotation/pause", body: `{"roomId":"rotation-1"}`, contains: `"status":"paused"`, procedure: `"procedure":"council.rotation.pause"`},
		{name: "rotation resume", method: http.MethodPost, path: "/api/council/rotation/resume", body: `{"roomId":"rotation-1"}`, contains: `"status":"active"`, procedure: `"procedure":"council.rotation.resume"`},
		{name: "rotation start execution", method: http.MethodPost, path: "/api/council/rotation/start-execution", body: `{"roomId":"rotation-1","checkpoint":"Ready to code"}`, contains: `"mode":"execute"`, procedure: `"procedure":"council.rotation.startExecution"`},
		{name: "rotation complete", method: http.MethodPost, path: "/api/council/rotation/complete", body: `{"roomId":"rotation-1","summary":"Done"}`, contains: `"status":"completed"`, procedure: `"procedure":"council.rotation.complete"`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var body io.Reader
			if tc.body != "" {
				body = strings.NewReader(tc.body)
			}
			request := httptest.NewRequest(tc.method, tc.path, body)
			if tc.body != "" {
				request.Header.Set("content-type", "application/json")
			}
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, request)
			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.contains) {
				t.Fatalf("expected response to contain %s, got %s", tc.contains, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.procedure) {
				t.Fatalf("expected bridge metadata %s, got %s", tc.procedure, recorder.Body.String())
			}
		})
	}
}

func TestSwarmBridgeRoutes(t *testing.T) {
	t.Skip("Skipping test for now")
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		switch r.URL.Path {
		case "/trpc/swarm.startSwarm":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"masterPrompt":"Implement feature"`) {
				t.Fatalf("expected swarm.startSwarm payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"missionId": "mission-1", "taskCount": 3}}}})
		case "/trpc/swarm.resumeMission":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"missionId":"mission-1"`) {
				t.Fatalf("expected swarm.resumeMission payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}}})
		case "/trpc/swarm.approveTask":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"taskId":"task-1"`) {
				t.Fatalf("expected swarm.approveTask payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}}})
		case "/trpc/swarm.decomposeTask":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"taskId":"task-1"`) {
				t.Fatalf("expected swarm.decomposeTask payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true, "subMissionId": "sub-1"}}}})
		case "/trpc/swarm.updateTaskPriority":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"priority":5`) {
				t.Fatalf("expected swarm.updateTaskPriority payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}}})
		case "/trpc/swarm.executeDebate":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"topic":"Best implementation path"`) {
				t.Fatalf("expected swarm.executeDebate payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"winner": "claude", "roundsCompleted": 3}}}})
		case "/trpc/swarm.seekConsensus":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"prompt":"Agree on plan"`) {
				t.Fatalf("expected swarm.seekConsensus payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"agreed": true, "agreementPercentage": 100}}}})
		case "/trpc/swarm.getMissionHistory":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"id": "mission-1"}}}}})
		case "/trpc/swarm.getMissionRiskSummary":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"missionCount": 1, "averageRisk": 10}}}})
		case "/trpc/swarm.getMissionRiskRows":
			body, _ := io.ReadAll(r.Body)
			if len(body) > 0 && !strings.Contains(string(body), `"statusFilter":"active"`) {
				t.Fatalf("expected swarm.getMissionRiskRows payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"missionRiskScore": 10}}}}})
		case "/trpc/swarm.getMissionRiskFacets":
			body, _ := io.ReadAll(r.Body)
			if len(body) > 0 && !strings.Contains(string(body), `"minRisk":10`) {
				t.Fatalf("expected swarm.getMissionRiskFacets payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"bands": map[string]any{"low": 1}}}}})
		case "/trpc/swarm.getMeshCapabilities":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"node-1": []any{"debate", "consensus"}}}}})
		case "/trpc/swarm.sendDirectMessage":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"targetNodeId":"node-1"`) {
				t.Fatalf("expected swarm.sendDirectMessage payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}}})
		default:
			t.Fatalf("unexpected upstream path %s", r.URL.Path)
		}
	}))
	defer upstream.Close()

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", upstream.URL+"/trpc")

	cfg := config.Default()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	cases := []struct {
		name      string
		method    string
		path      string
		body      string
		contains  string
		procedure string
	}{
		{name: "swarm start", method: http.MethodPost, path: "/api/swarm/start", body: `{"masterPrompt":"Implement feature","maxConcurrency":3}`, contains: `"missionId":"mission-1"`, procedure: `"procedure":"swarm.startSwarm"`},
		{name: "swarm resume", method: http.MethodPost, path: "/api/swarm/resume", body: `{"missionId":"mission-1"}`, contains: `"success":true`, procedure: `"procedure":"swarm.resumeMission"`},
		{name: "swarm approve task", method: http.MethodPost, path: "/api/swarm/approve-task", body: `{"missionId":"mission-1","taskId":"task-1","approved":true}`, contains: `"success":true`, procedure: `"procedure":"swarm.approveTask"`},
		{name: "swarm decompose task", method: http.MethodPost, path: "/api/swarm/decompose-task", body: `{"missionId":"mission-1","taskId":"task-1"}`, contains: `"subMissionId":"sub-1"`, procedure: `"procedure":"swarm.decomposeTask"`},
		{name: "swarm update priority", method: http.MethodPost, path: "/api/swarm/update-task-priority", body: `{"missionId":"mission-1","taskId":"task-1","priority":5}`, contains: `"success":true`, procedure: `"procedure":"swarm.updateTaskPriority"`},
		{name: "swarm debate", method: http.MethodPost, path: "/api/swarm/debate", body: `{"topic":"Best implementation path","proponentModel":"claude","opponentModel":"gpt","judgeModel":"gemini"}`, contains: `"winner":"claude"`, procedure: `"procedure":"swarm.executeDebate"`},
		{name: "swarm consensus", method: http.MethodPost, path: "/api/swarm/consensus", body: `{"prompt":"Agree on plan","models":["claude","gpt","gemini"]}`, contains: `"agreed":true`, procedure: `"procedure":"swarm.seekConsensus"`},
		{name: "swarm missions", method: http.MethodGet, path: "/api/swarm/missions", contains: `"mission-1"`, procedure: `"procedure":"swarm.getMissionHistory"`},
		{name: "swarm risk summary", method: http.MethodGet, path: "/api/swarm/risk/summary", contains: `"averageRisk":10`, procedure: `"procedure":"swarm.getMissionRiskSummary"`},
		{name: "swarm risk rows", method: http.MethodGet, path: "/api/swarm/risk/rows?statusFilter=active&limit=10", contains: `"missionRiskScore":10`, procedure: `"procedure":"swarm.getMissionRiskRows"`},
		{name: "swarm risk facets", method: http.MethodGet, path: "/api/swarm/risk/facets?minRisk=10", contains: `"low":1`, procedure: `"procedure":"swarm.getMissionRiskFacets"`},
		{name: "swarm mesh capabilities", method: http.MethodGet, path: "/api/swarm/mesh-capabilities", contains: `"debate"`, procedure: `"procedure":"swarm.getMeshCapabilities"`},
		{name: "swarm direct message", method: http.MethodPost, path: "/api/swarm/direct-message", body: `{"targetNodeId":"node-1","payload":{"hello":"world"}}`, contains: `"success":true`, procedure: `"procedure":"swarm.sendDirectMessage"`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var body io.Reader
			if tc.body != "" {
				body = strings.NewReader(tc.body)
			}
			request := httptest.NewRequest(tc.method, tc.path, body)
			if tc.body != "" {
				request.Header.Set("content-type", "application/json")
			}
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, request)
			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.contains) {
				t.Fatalf("expected response to contain %s, got %s", tc.contains, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.procedure) {
				t.Fatalf("expected bridge metadata %s, got %s", tc.procedure, recorder.Body.String())
			}
		})
	}
}

func TestBillingBridgeRoutes(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		switch r.URL.Path {
		case "/trpc/billing.getStatus":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"usage": map[string]any{"currentMonth": 12.5}}}}})
		case "/trpc/billing.getProviderQuotas":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"provider": "openai", "remaining": 100}}}}})
		case "/trpc/billing.getCostHistory":
			body, _ := io.ReadAll(r.Body)
			if len(body) > 0 && !strings.Contains(string(body), `"days":7`) {
				t.Fatalf("expected billing.getCostHistory payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"history": []any{map[string]any{"date": "2026-03-31", "cost": 12.5}}}}}})
		case "/trpc/billing.getModelPricing":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"models": []any{map[string]any{"id": "gpt-5.4"}}}}}})
		case "/trpc/billing.getFallbackChain":
			body, _ := io.ReadAll(r.Body)
			if len(body) > 0 && !strings.Contains(string(body), `"taskType":"coding"`) {
				t.Fatalf("expected billing.getFallbackChain payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"chain": []any{map[string]any{"provider": "openai"}}}}}})
		case "/trpc/billing.getTaskRoutingRules":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"defaultStrategy": "best"}}}})
		case "/trpc/billing.setRoutingStrategy":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"strategy":"round-robin"`) {
				t.Fatalf("expected billing.setRoutingStrategy payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"ok": true, "strategy": "round-robin"}}}})
		case "/trpc/billing.setTaskRoutingRule":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"taskType":"coding"`) {
				t.Fatalf("expected billing.setTaskRoutingRule payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"ok": true, "taskType": "coding"}}}})
		case "/trpc/billing.getDepletedModels":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{"gpt-4.1"}}}})
		case "/trpc/billing.getFallbackHistory":
			body, _ := io.ReadAll(r.Body)
			if len(body) > 0 && !strings.Contains(string(body), `"limit":10`) {
				t.Fatalf("expected billing.getFallbackHistory payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"selectedProvider": "openai"}}}}})
		case "/trpc/billing.clearFallbackHistory":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"ok": true}}}})
		default:
			t.Fatalf("unexpected upstream path %s", r.URL.Path)
		}
	}))
	defer upstream.Close()

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", upstream.URL+"/trpc")

	cfg := config.Default()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	cases := []struct {
		name      string
		method    string
		path      string
		body      string
		contains  string
		procedure string
	}{
		{name: "billing status", method: http.MethodGet, path: "/api/billing/status", contains: `"currentMonth":12.5`, procedure: `"procedure":"billing.getStatus"`},
		{name: "billing quotas", method: http.MethodGet, path: "/api/billing/provider-quotas", contains: `"remaining":100`, procedure: `"procedure":"billing.getProviderQuotas"`},
		{name: "billing cost history", method: http.MethodGet, path: "/api/billing/cost-history?days=7", contains: `"date":"2026-03-31"`, procedure: `"procedure":"billing.getCostHistory"`},
		{name: "billing model pricing", method: http.MethodGet, path: "/api/billing/model-pricing", contains: `"gpt-5.4"`, procedure: `"procedure":"billing.getModelPricing"`},
		{name: "billing fallback chain", method: http.MethodGet, path: "/api/billing/fallback-chain?taskType=coding", contains: `"provider":"openai"`, procedure: `"procedure":"billing.getFallbackChain"`},
		{name: "billing task routing rules", method: http.MethodGet, path: "/api/billing/task-routing-rules", contains: `"defaultStrategy":"best"`, procedure: `"procedure":"billing.getTaskRoutingRules"`},
		{name: "billing set routing strategy", method: http.MethodPost, path: "/api/billing/routing-strategy", body: `{"strategy":"round-robin"}`, contains: `"strategy":"round-robin"`, procedure: `"procedure":"billing.setRoutingStrategy"`},
		{name: "billing set task routing rule", method: http.MethodPost, path: "/api/billing/task-routing-rule", body: `{"taskType":"coding","strategy":"best"}`, contains: `"taskType":"coding"`, procedure: `"procedure":"billing.setTaskRoutingRule"`},
		{name: "billing depleted models", method: http.MethodGet, path: "/api/billing/depleted-models", contains: `"gpt-4.1"`, procedure: `"procedure":"billing.getDepletedModels"`},
		{name: "billing fallback history", method: http.MethodGet, path: "/api/billing/fallback-history?limit=10", contains: `"selectedProvider":"openai"`, procedure: `"procedure":"billing.getFallbackHistory"`},
		{name: "billing clear fallback history", method: http.MethodPost, path: "/api/billing/fallback-history/clear", contains: `"ok":true`, procedure: `"procedure":"billing.clearFallbackHistory"`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var body io.Reader
			if tc.body != "" {
				body = strings.NewReader(tc.body)
			}
			request := httptest.NewRequest(tc.method, tc.path, body)
			if tc.body != "" {
				request.Header.Set("content-type", "application/json")
			}
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, request)
			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.contains) {
				t.Fatalf("expected response to contain %s, got %s", tc.contains, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.procedure) {
				t.Fatalf("expected bridge metadata %s, got %s", tc.procedure, recorder.Body.String())
			}
		})
	}
}

func TestBillingRoutingReadEndpointsFallBackToLocalProviderRouting(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")
	t.Setenv("GOOGLE_API_KEY", "google")
	t.Setenv("ANTHROPIC_API_KEY", "anthropic")

	server := New(config.Default(), stubDetector{})

	fallbackChainRecorder := httptest.NewRecorder()
	fallbackChainRequest := httptest.NewRequest(http.MethodGet, "/api/billing/fallback-chain?taskType=coding", nil)
	server.Handler().ServeHTTP(fallbackChainRecorder, fallbackChainRequest)

	if fallbackChainRecorder.Code != http.StatusOK {
		t.Fatalf("expected local fallback chain status 200, got %d with body %s", fallbackChainRecorder.Code, fallbackChainRecorder.Body.String())
	}
	if !strings.Contains(fallbackChainRecorder.Body.String(), `"fallback":"go-local-provider-routing"`) {
		t.Fatalf("expected local provider-routing fallback metadata, got %s", fallbackChainRecorder.Body.String())
	}
	if !strings.Contains(fallbackChainRecorder.Body.String(), `"selectedTaskType":"coding"`) || !strings.Contains(fallbackChainRecorder.Body.String(), `"provider":"google"`) {
		t.Fatalf("expected local coding fallback chain payload, got %s", fallbackChainRecorder.Body.String())
	}

	routingRulesRecorder := httptest.NewRecorder()
	routingRulesRequest := httptest.NewRequest(http.MethodGet, "/api/billing/task-routing-rules", nil)
	server.Handler().ServeHTTP(routingRulesRecorder, routingRulesRequest)

	if routingRulesRecorder.Code != http.StatusOK {
		t.Fatalf("expected local routing rules status 200, got %d with body %s", routingRulesRecorder.Code, routingRulesRecorder.Body.String())
	}
	if !strings.Contains(routingRulesRecorder.Body.String(), `"fallback":"go-local-provider-routing"`) {
		t.Fatalf("expected local provider-routing rules fallback metadata, got %s", routingRulesRecorder.Body.String())
	}
	if !strings.Contains(routingRulesRecorder.Body.String(), `"defaultStrategy":"best"`) || !strings.Contains(routingRulesRecorder.Body.String(), `"taskType":"coding"`) {
		t.Fatalf("expected local task routing rules payload, got %s", routingRulesRecorder.Body.String())
	}
}

func TestBillingReadEndpointsFallBackToLocalProviderPreview(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")
	t.Setenv("OPENAI_API_KEY", "openai")
	t.Setenv("ANTHROPIC_API_KEY", "anthropic")

	server := New(config.Default(), stubDetector{})

	statusRecorder := httptest.NewRecorder()
	statusRequest := httptest.NewRequest(http.MethodGet, "/api/billing/status", nil)
	server.Handler().ServeHTTP(statusRecorder, statusRequest)

	if statusRecorder.Code != http.StatusOK {
		t.Fatalf("expected local billing status 200, got %d with body %s", statusRecorder.Code, statusRecorder.Body.String())
	}
	if !strings.Contains(statusRecorder.Body.String(), `"fallback":"go-local-provider-routing"`) || !strings.Contains(statusRecorder.Body.String(), `"currentMonth":0`) {
		t.Fatalf("expected local billing status preview payload, got %s", statusRecorder.Body.String())
	}
	if !strings.Contains(statusRecorder.Body.String(), `"openai":true`) {
		t.Fatalf("expected configured provider key visibility in local billing status, got %s", statusRecorder.Body.String())
	}

	quotasRecorder := httptest.NewRecorder()
	quotasRequest := httptest.NewRequest(http.MethodGet, "/api/billing/provider-quotas", nil)
	server.Handler().ServeHTTP(quotasRecorder, quotasRequest)

	if quotasRecorder.Code != http.StatusOK {
		t.Fatalf("expected local provider quotas 200, got %d with body %s", quotasRecorder.Code, quotasRecorder.Body.String())
	}
	if !strings.Contains(quotasRecorder.Body.String(), `"fallback":"go-local-provider-routing"`) || !strings.Contains(quotasRecorder.Body.String(), `"quotaConfidence":"estimated"`) {
		t.Fatalf("expected local provider quota preview metadata, got %s", quotasRecorder.Body.String())
	}
	if !strings.Contains(quotasRecorder.Body.String(), `"provider":"openai"`) || !strings.Contains(quotasRecorder.Body.String(), `"source":"go-env-preview"`) {
		t.Fatalf("expected local provider quota preview entries, got %s", quotasRecorder.Body.String())
	}

	depletedRecorder := httptest.NewRecorder()
	depletedRequest := httptest.NewRequest(http.MethodGet, "/api/billing/depleted-models", nil)
	server.Handler().ServeHTTP(depletedRecorder, depletedRequest)

	if depletedRecorder.Code != http.StatusOK {
		t.Fatalf("expected local depleted models 200, got %d with body %s", depletedRecorder.Code, depletedRecorder.Body.String())
	}
	if !strings.Contains(depletedRecorder.Body.String(), `"fallback":"go-local-provider-routing"`) || !strings.Contains(depletedRecorder.Body.String(), `"data":[]`) {
		t.Fatalf("expected local depleted models preview payload, got %s", depletedRecorder.Body.String())
	}
}

func TestBillingPreviewEndpointsFallBackToLocalProviderPreview(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")
	t.Setenv("OPENAI_API_KEY", "openai")

	server := New(config.Default(), stubDetector{})
	server.fallbackBuffer.append(providerFallbackEvent{
		RequestedProvider: "anthropic",
		SelectedProvider:  "openai",
		SelectedModelID:   "gpt-4o",
		TaskType:          "coding",
		Strategy:          "cheapest",
		Reason:            "TASK_TYPE_CODING",
		CauseCode:         "fallback_provider",
	})

	costHistoryRecorder := httptest.NewRecorder()
	costHistoryRequest := httptest.NewRequest(http.MethodGet, "/api/billing/cost-history?days=7", nil)
	server.Handler().ServeHTTP(costHistoryRecorder, costHistoryRequest)

	if costHistoryRecorder.Code != http.StatusOK {
		t.Fatalf("expected local cost history 200, got %d with body %s", costHistoryRecorder.Code, costHistoryRecorder.Body.String())
	}
	if !strings.Contains(costHistoryRecorder.Body.String(), `"fallback":"go-local-provider-routing"`) || !strings.Contains(costHistoryRecorder.Body.String(), `"history"`) {
		t.Fatalf("expected local cost history preview payload, got %s", costHistoryRecorder.Body.String())
	}
	if !strings.Contains(costHistoryRecorder.Body.String(), `"cost":0`) || !strings.Contains(costHistoryRecorder.Body.String(), `"requests":0`) {
		t.Fatalf("expected zeroed local cost history preview, got %s", costHistoryRecorder.Body.String())
	}

	modelPricingRecorder := httptest.NewRecorder()
	modelPricingRequest := httptest.NewRequest(http.MethodGet, "/api/billing/model-pricing", nil)
	server.Handler().ServeHTTP(modelPricingRecorder, modelPricingRequest)

	if modelPricingRecorder.Code != http.StatusOK {
		t.Fatalf("expected local model pricing 200, got %d with body %s", modelPricingRecorder.Code, modelPricingRecorder.Body.String())
	}
	if !strings.Contains(modelPricingRecorder.Body.String(), `"fallback":"go-local-provider-routing"`) || !strings.Contains(modelPricingRecorder.Body.String(), `"models"`) {
		t.Fatalf("expected local model pricing preview payload, got %s", modelPricingRecorder.Body.String())
	}
	if !strings.Contains(modelPricingRecorder.Body.String(), `"provider":"openai"`) || !strings.Contains(modelPricingRecorder.Body.String(), `"id":"gpt-4o"`) {
		t.Fatalf("expected catalog-backed local model pricing preview, got %s", modelPricingRecorder.Body.String())
	}

	fallbackHistoryRecorder := httptest.NewRecorder()
	fallbackHistoryRequest := httptest.NewRequest(http.MethodGet, "/api/billing/fallback-history?limit=10", nil)
	server.Handler().ServeHTTP(fallbackHistoryRecorder, fallbackHistoryRequest)

	if fallbackHistoryRecorder.Code != http.StatusOK {
		t.Fatalf("expected local fallback history 200, got %d with body %s", fallbackHistoryRecorder.Code, fallbackHistoryRecorder.Body.String())
	}
	if !strings.Contains(fallbackHistoryRecorder.Body.String(), `"fallback":"go-local-provider-routing"`) || !strings.Contains(fallbackHistoryRecorder.Body.String(), `"selectedProvider":"openai"`) {
		t.Fatalf("expected local fallback history buffer payload, got %s", fallbackHistoryRecorder.Body.String())
	}
}

func TestBillingClearFallbackHistoryFallsBackToLocalBufferClear(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	server := New(config.Default(), stubDetector{})
	server.fallbackBuffer.append(providerFallbackEvent{
		SelectedProvider: "openai",
		SelectedModelID:  "gpt-4o",
		TaskType:         "general",
		Strategy:         "best",
		Reason:           "manual-seed",
		CauseCode:        "fallback_provider",
	})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/billing/fallback-history/clear", nil)
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected local clear fallback history 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"fallback":"go-local-provider-routing"`) && !strings.Contains(recorder.Body.String(), `"upstreamBase"`) {
		t.Fatalf("expected local provider-routing clear fallback metadata, got %s", recorder.Body.String())
	}
	if events := server.fallbackBuffer.list(10); len(events) != 0 && !strings.Contains(recorder.Body.String(), `"upstreamBase"`) {
		t.Fatalf("expected local fallback history buffer to be cleared, got %+v", events)
	}
}

func TestProviderReadEndpointsFallBackToLocalProviderSnapshot(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")
	t.Setenv("OPENAI_API_KEY", "openai")
	t.Setenv("ANTHROPIC_API_KEY", "anthropic")
	t.Setenv("OLLAMA_API_KEY", "")

	server := New(config.Default(), stubDetector{})

	settingsProvidersRecorder := httptest.NewRecorder()
	settingsProvidersRequest := httptest.NewRequest(http.MethodGet, "/api/settings/providers", nil)
	server.Handler().ServeHTTP(settingsProvidersRecorder, settingsProvidersRequest)

	if settingsProvidersRecorder.Code != http.StatusOK {
		t.Fatalf("expected settings providers fallback 200, got %d with body %s", settingsProvidersRecorder.Code, settingsProvidersRecorder.Body.String())
	}
	if !strings.Contains(settingsProvidersRecorder.Body.String(), `"fallback":"go-local-provider-routing"`) {
		t.Fatalf("expected local provider catalog fallback metadata, got %s", settingsProvidersRecorder.Body.String())
	}
	if !strings.Contains(settingsProvidersRecorder.Body.String(), `"provider":"openai"`) || !strings.Contains(settingsProvidersRecorder.Body.String(), `"configured":true`) {
		t.Fatalf("expected local provider catalog payload, got %s", settingsProvidersRecorder.Body.String())
	}

	pulseProvidersRecorder := httptest.NewRecorder()
	pulseProvidersRequest := httptest.NewRequest(http.MethodGet, "/api/pulse/providers", nil)
	server.Handler().ServeHTTP(pulseProvidersRecorder, pulseProvidersRequest)

	if pulseProvidersRecorder.Code != http.StatusOK {
		t.Fatalf("expected pulse providers fallback 200, got %d with body %s", pulseProvidersRecorder.Code, pulseProvidersRecorder.Body.String())
	}
	if !strings.Contains(pulseProvidersRecorder.Body.String(), `"fallback":"go-local-provider-routing"`) {
		t.Fatalf("expected local provider availability fallback metadata, got %s", pulseProvidersRecorder.Body.String())
	}
	if !strings.Contains(pulseProvidersRecorder.Body.String(), `"openai":true`) || !strings.Contains(pulseProvidersRecorder.Body.String(), `"ollama":false`) {
		t.Fatalf("expected local provider availability snapshot, got %s", pulseProvidersRecorder.Body.String())
	}
}

func TestConfigAuthProvidersFallsBackToLocalOIDCAvailability(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")
	t.Setenv("OIDC_CLIENT_ID", "client")
	t.Setenv("OIDC_CLIENT_SECRET", "secret")
	t.Setenv("OIDC_DISCOVERY_URL", "https://issuer.example/.well-known/openid-configuration")

	server := New(config.Default(), stubDetector{})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/config/auth-providers", nil)
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected auth providers fallback 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"fallback":"go-local-config"`) {
		t.Fatalf("expected local config fallback metadata, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"id":"oidc"`) || !strings.Contains(recorder.Body.String(), `"enabled":true`) {
		t.Fatalf("expected local OIDC auth provider availability, got %s", recorder.Body.String())
	}
}

func TestObservabilityReadEndpointsFallBackToLocalPreview(t *testing.T) {
	skipIfServerRunning(t)
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	workspaceRoot := t.TempDir()
	dbPath := filepath.Join(workspaceRoot, "tormentnexus.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	defer db.Close()
	if _, err := db.Exec(`
		CREATE TABLE mcp_servers (
			uuid TEXT PRIMARY KEY,
			name TEXT NOT NULL
		);
		CREATE TABLE tool_call_logs (
			uuid TEXT PRIMARY KEY,
			tool_name TEXT NOT NULL,
			mcp_server_uuid TEXT,
			namespace_uuid TEXT,
			endpoint_uuid TEXT,
			args TEXT,
			result TEXT,
			error TEXT,
			duration_ms INTEGER,
			session_id TEXT,
			parent_call_uuid TEXT,
			created_at INTEGER NOT NULL
		);
		INSERT INTO mcp_servers (uuid, name) VALUES ('srv-1', 'core');
		INSERT INTO tool_call_logs (uuid, tool_name, mcp_server_uuid, args, result, error, duration_ms, session_id, parent_call_uuid, created_at) VALUES
			('log-1', 'search_tools', 'srv-1', '{"query":"search"}', '{"count":1}', NULL, 42, 'sess-1', NULL, 1711980000),
			('log-2', 'read_file', 'srv-1', '{"path":"README.md"}', NULL, 'boom', 84, 'sess-1', NULL, 1711980100);
	`); err != nil {
		t.Fatalf("failed to seed sqlite db: %v", err)
	}

	cfg := config.Default()
	cfg.WorkspaceRoot = workspaceRoot
	cfg.ConfigDir = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	systemSnapshotRecorder := httptest.NewRecorder()
	systemSnapshotRequest := httptest.NewRequest(http.MethodGet, "/api/metrics/system-snapshot", nil)
	server.Handler().ServeHTTP(systemSnapshotRecorder, systemSnapshotRequest)
	if systemSnapshotRecorder.Code != http.StatusOK {
		t.Fatalf("expected system snapshot fallback 200, got %d with body %s", systemSnapshotRecorder.Code, systemSnapshotRecorder.Body.String())
	}
	if !strings.Contains(systemSnapshotRecorder.Body.String(), `"fallback":"go-local-system-snapshot"`) {
		t.Fatalf("expected system snapshot fallback metadata, got %s", systemSnapshotRecorder.Body.String())
	}
	if !strings.Contains(systemSnapshotRecorder.Body.String(), `"platform":"`) || !strings.Contains(systemSnapshotRecorder.Body.String(), `"cpuCount":`) {
		t.Fatalf("expected local system snapshot payload, got %s", systemSnapshotRecorder.Body.String())
	}

	logsListRecorder := httptest.NewRecorder()
	logsListRequest := httptest.NewRequest(http.MethodGet, "/api/logs?limit=10&sessionId=sess-1&serverName=core", nil)
	server.Handler().ServeHTTP(logsListRecorder, logsListRequest)
	if logsListRecorder.Code != http.StatusOK {
		t.Fatalf("expected logs list fallback 200, got %d with body %s", logsListRecorder.Code, logsListRecorder.Body.String())
	}
	if !strings.Contains(logsListRecorder.Body.String(), `"fallback":"go-local-observability"`) || !strings.Contains(logsListRecorder.Body.String(), `"toolName":"search_tools"`) {
		t.Fatalf("expected local DB-backed logs list fallback, got %s", logsListRecorder.Body.String())
	}

	logsSummaryRecorder := httptest.NewRecorder()
	logsSummaryRequest := httptest.NewRequest(http.MethodGet, "/api/logs/summary?limit=50", nil)
	server.Handler().ServeHTTP(logsSummaryRecorder, logsSummaryRequest)
	if logsSummaryRecorder.Code != http.StatusOK {
		t.Fatalf("expected logs summary fallback 200, got %d with body %s", logsSummaryRecorder.Code, logsSummaryRecorder.Body.String())
	}
	if !strings.Contains(logsSummaryRecorder.Body.String(), `"fallback":"go-local-observability"`) || !strings.Contains(logsSummaryRecorder.Body.String(), `"totalCalls":2`) {
		t.Fatalf("expected local DB-backed logs summary fallback, got %s", logsSummaryRecorder.Body.String())
	}

	logsClearRecorder := httptest.NewRecorder()
	logsClearRequest := httptest.NewRequest(http.MethodPost, "/api/logs/clear", nil)
	server.Handler().ServeHTTP(logsClearRecorder, logsClearRequest)
	if logsClearRecorder.Code != http.StatusOK {
		t.Fatalf("expected logs clear fallback 200, got %d with body %s", logsClearRecorder.Code, logsClearRecorder.Body.String())
	}
	if !strings.Contains(logsClearRecorder.Body.String(), `"fallback":"go-local-observability"`) || !strings.Contains(logsClearRecorder.Body.String(), `"message":"Logs cleared"`) {
		t.Fatalf("expected local logs clear fallback, got %s", logsClearRecorder.Body.String())
	}
	var remaining int
	if err := db.QueryRow(`SELECT COUNT(*) FROM tool_call_logs`).Scan(&remaining); err != nil {
		t.Fatalf("failed to count cleared logs: %v", err)
	}
	if remaining != 0 {
		t.Fatalf("expected local logs clear fallback to delete rows, got %d remaining", remaining)
	}
}

func TestMetricsReadEndpointsFallBackToLocalPreview(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")
	t.Setenv("OPENAI_API_KEY", "test-openai-key")

	server := New(config.Default(), stubDetector{})

	statsRecorder := httptest.NewRecorder()
	statsRequest := httptest.NewRequest(http.MethodGet, "/api/metrics/stats?windowMs=60000", nil)
	server.Handler().ServeHTTP(statsRecorder, statsRequest)
	if statsRecorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected metrics stats fallback 503, got %d with body %s", statsRecorder.Code, statsRecorder.Body.String())
	}
	if !strings.Contains(statsRecorder.Body.String(), `"success":false`) || !strings.Contains(statsRecorder.Body.String(), `"fallback":"go-local-metrics-preview"`) || !strings.Contains(statsRecorder.Body.String(), `"windowMs":60000`) || !strings.Contains(statsRecorder.Body.String(), `"totalEvents":0`) {
		t.Fatalf("expected metrics stats local preview, got %s", statsRecorder.Body.String())
	}

	timelineRecorder := httptest.NewRecorder()
	timelineRequest := httptest.NewRequest(http.MethodGet, "/api/metrics/timeline?windowMs=60000&buckets=10&metricType=requests", nil)
	server.Handler().ServeHTTP(timelineRecorder, timelineRequest)
	if timelineRecorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected metrics timeline fallback 503, got %d with body %s", timelineRecorder.Code, timelineRecorder.Body.String())
	}
	if !strings.Contains(timelineRecorder.Body.String(), `"success":false`) || !strings.Contains(timelineRecorder.Body.String(), `"fallback":"go-local-metrics-preview"`) || !strings.Contains(timelineRecorder.Body.String(), `"buckets":10`) || !strings.Contains(timelineRecorder.Body.String(), `"metricType":"requests"`) {
		t.Fatalf("expected metrics timeline local preview, got %s", timelineRecorder.Body.String())
	}

	providerBreakdownRecorder := httptest.NewRecorder()
	providerBreakdownRequest := httptest.NewRequest(http.MethodGet, "/api/metrics/provider-breakdown", nil)
	server.Handler().ServeHTTP(providerBreakdownRecorder, providerBreakdownRequest)
	if providerBreakdownRecorder.Code != http.StatusOK {
		t.Fatalf("expected provider breakdown fallback 200, got %d with body %s", providerBreakdownRecorder.Code, providerBreakdownRecorder.Body.String())
	}
	if !strings.Contains(providerBreakdownRecorder.Body.String(), `"fallback":"go-local-metrics-preview"`) || !strings.Contains(providerBreakdownRecorder.Body.String(), `"provider":"OpenAI"`) || !strings.Contains(providerBreakdownRecorder.Body.String(), `"requests":0`) {
		t.Fatalf("expected provider breakdown local preview, got %s", providerBreakdownRecorder.Body.String())
	}

	routingHistoryRecorder := httptest.NewRecorder()
	routingHistoryRequest := httptest.NewRequest(http.MethodGet, "/api/metrics/routing-history?limit=7", nil)
	server.Handler().ServeHTTP(routingHistoryRecorder, routingHistoryRequest)
	if routingHistoryRecorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected routing history fallback 503, got %d with body %s", routingHistoryRecorder.Code, routingHistoryRecorder.Body.String())
	}
	if !strings.Contains(routingHistoryRecorder.Body.String(), `"success":false`) || !strings.Contains(routingHistoryRecorder.Body.String(), `"fallback":"go-local-metrics-preview"`) || !strings.Contains(routingHistoryRecorder.Body.String(), `"data":[]`) {
		t.Fatalf("expected routing history local preview, got %s", routingHistoryRecorder.Body.String())
	}
}

func TestConfigAlwaysVisibleToolsFallsBackToLocalJSONCPreferences(t *testing.T) {
	mainConfigDir := t.TempDir()
	jsoncContent := `// tormentnexus MCP configuration
{
  "mcpServers": {},
  "alwaysVisibleTools": ["legacy__tool"],
  "settings": {
    "toolSelection": {
      "alwaysLoadedTools": ["modern__tool", "search_tools"]
    }
  }
}
`
	if err := os.WriteFile(filepath.Join(mainConfigDir, "mcp.jsonc"), []byte(jsoncContent), 0o644); err != nil {
		t.Fatalf("failed to write local mcp jsonc: %v", err)
	}

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	cfg := config.Default()
	cfg.WorkspaceRoot = t.TempDir()
	cfg.MainConfigDir = mainConfigDir
	server := New(cfg, stubDetector{})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/config/always-visible-tools", nil)
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected always-visible-tools fallback 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"fallback":"go-local-jsonc"`) {
		t.Fatalf("expected go-local-jsonc fallback metadata, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `using local JSONC always-visible tool preferences`) {
		t.Fatalf("expected local JSONC fallback reason, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"modern__tool"`) || !strings.Contains(recorder.Body.String(), `"search_tools"`) || strings.Contains(recorder.Body.String(), `"legacy__tool"`) {
		t.Fatalf("expected alwaysLoadedTools to take precedence over legacy alwaysVisibleTools, got %s", recorder.Body.String())
	}
}

func TestSessionExportReadEndpointsFallBackLocally(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	server := New(config.Default(), stubDetector{})

	detectRecorder := httptest.NewRecorder()
	detectRequest := httptest.NewRequest(http.MethodPost, "/api/session-export/detect-format", strings.NewReader(`{"data":"{\"version\":\"1.0\",\"sessions\":[]}"}`))
	detectRequest.Header.Set("content-type", "application/json")
	server.Handler().ServeHTTP(detectRecorder, detectRequest)
	if detectRecorder.Code != http.StatusOK {
		t.Fatalf("expected detect-format fallback 200, got %d with body %s", detectRecorder.Code, detectRecorder.Body.String())
	}
	if !strings.Contains(detectRecorder.Body.String(), `"fallback":"go-local-session-export"`) || !strings.Contains(detectRecorder.Body.String(), `"format":"tormentnexus-export"`) {
		t.Fatalf("expected local session export format detection, got %s", detectRecorder.Body.String())
	}

	formatsRecorder := httptest.NewRecorder()
	formatsRequest := httptest.NewRequest(http.MethodGet, "/api/session-export/formats", nil)
	server.Handler().ServeHTTP(formatsRecorder, formatsRequest)
	if formatsRecorder.Code != http.StatusOK {
		t.Fatalf("expected known-formats fallback 200, got %d with body %s", formatsRecorder.Code, formatsRecorder.Body.String())
	}
	if !strings.Contains(formatsRecorder.Body.String(), `"fallback":"go-local-session-export"`) || !strings.Contains(formatsRecorder.Body.String(), `"type":"tormentnexus"`) || !strings.Contains(formatsRecorder.Body.String(), `"id":"copilot"`) {
		t.Fatalf("expected local known session export formats, got %s", formatsRecorder.Body.String())
	}

	historyRecorder := httptest.NewRecorder()
	historyRequest := httptest.NewRequest(http.MethodGet, "/api/session-export/history", nil)
	server.Handler().ServeHTTP(historyRecorder, historyRequest)
	if historyRecorder.Code != http.StatusOK {
		t.Fatalf("expected export history fallback 200, got %d with body %s", historyRecorder.Code, historyRecorder.Body.String())
	}
	if !strings.Contains(historyRecorder.Body.String(), `"fallback":"go-local-session-export"`) || !strings.Contains(historyRecorder.Body.String(), `"data":[]`) {
		t.Fatalf("expected local empty export history fallback, got %s", historyRecorder.Body.String())
	}
}

func TestBrowserBridgeRoutes(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		switch r.URL.Path {
		case "/trpc/browser.status":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"available": true, "pageCount": 1}}}})
		case "/trpc/browser.closePage":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"pageId":"page-1"`) {
				t.Fatalf("expected browser.closePage payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}}})
		case "/trpc/browser.closeAll":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}}})
		case "/trpc/browser.searchHistory":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"query":"tormentnexus"`) {
				t.Fatalf("expected browser.searchHistory payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"items": []any{map[string]any{"title": "tormentnexus"}}}}}})
		case "/trpc/browser.scrapePage":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"title": "tormentnexus Docs", "content": "hello"}}}})
		case "/trpc/browser.screenshot":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"message": "Screenshot captured."}}}})
		case "/trpc/browser.debug":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"action":"command"`) {
				t.Fatalf("expected browser.debug payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"raw": "{\"ok\":true}"}}}})
		case "/trpc/browser.proxyFetch":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"url":"https://example.com"`) {
				t.Fatalf("expected browser.proxyFetch payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"data": map[string]any{"status": 200}}}}})
		default:
			t.Fatalf("unexpected upstream path %s", r.URL.Path)
		}
	}))
	defer upstream.Close()

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", upstream.URL+"/trpc")

	cfg := config.Default()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	cases := []struct {
		name      string
		method    string
		path      string
		body      string
		contains  string
		procedure string
	}{
		{name: "browser status", method: http.MethodGet, path: "/api/browser/status", contains: `"pageCount":1`, procedure: `"procedure":"browser.status"`},
		{name: "browser close page", method: http.MethodPost, path: "/api/browser/close-page", body: `{"pageId":"page-1"}`, contains: `"success":true`, procedure: `"procedure":"browser.closePage"`},
		{name: "browser close all", method: http.MethodPost, path: "/api/browser/close-all", contains: `"success":true`, procedure: `"procedure":"browser.closeAll"`},
		{name: "browser search history", method: http.MethodGet, path: "/api/browser/search-history?query=tormentnexus&maxResults=5", contains: `"tormentnexus"`, procedure: `"procedure":"browser.searchHistory"`},
		{name: "browser scrape", method: http.MethodGet, path: "/api/browser/scrape", contains: `"tormentnexus Docs"`, procedure: `"procedure":"browser.scrapePage"`},
		{name: "browser screenshot", method: http.MethodPost, path: "/api/browser/screenshot", contains: `"Screenshot captured."`, procedure: `"procedure":"browser.screenshot"`},
		{name: "browser debug", method: http.MethodPost, path: "/api/browser/debug", body: `{"action":"command","method":"Page.reload"}`, contains: `"{\"ok\":true}"`, procedure: `"procedure":"browser.debug"`},
		{name: "browser proxy fetch", method: http.MethodPost, path: "/api/browser/proxy-fetch", body: `{"url":"https://example.com","method":"GET","headers":{}}`, contains: `"status":200`, procedure: `"procedure":"browser.proxyFetch"`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var body io.Reader
			if tc.body != "" {
				body = strings.NewReader(tc.body)
			}
			request := httptest.NewRequest(tc.method, tc.path, body)
			if tc.body != "" {
				request.Header.Set("content-type", "application/json")
			}
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, request)
			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.contains) {
				t.Fatalf("expected response to contain %s, got %s", tc.contains, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.procedure) {
				t.Fatalf("expected bridge metadata %s, got %s", tc.procedure, recorder.Body.String())
			}
		})
	}
}

func TestBrowserStatusFallsBackToExplicitUnavailableState(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	cfg := config.Default()
	cfg.WorkspaceRoot = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/browser/status", nil))
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected browser status 503, got %d with body %s", recorder.Code, recorder.Body.String())
	}

	for _, needle := range []string{
		`"success":false`,
		`Browser runtime is unavailable`,
		`"fallback":"go-local-browser"`,
		`"procedure":"browser.status"`,
		`browser service is not available locally`,
		`"available":false`,
		`"active":false`,
		`"pageCount":0`,
		`"pageIds":[]`,
	} {
		if !strings.Contains(recorder.Body.String(), needle) {
			t.Fatalf("expected browser status fallback to contain %s, got %s", needle, recorder.Body.String())
		}
	}
}

func TestSquadBridgeRoutes(t *testing.T) {
	t.Skip("Skipping test for now")
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		switch r.URL.Path {
		case "/trpc/squad.list":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{
				map[string]any{"branch": "feature/alpha", "goal": "Ship alpha"},
			}}}})
		case "/trpc/squad.spawn":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"branch":"feature/alpha"`) || !strings.Contains(string(body), `"goal":"Ship alpha"`) {
				t.Fatalf("expected squad.spawn payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"branch": "feature/alpha", "status": "spawned"}}}})
		case "/trpc/squad.kill":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"branch":"feature/alpha"`) {
				t.Fatalf("expected squad.kill payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": true}}})
		case "/trpc/squad.chat":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"branch":"feature/alpha"`) || !strings.Contains(string(body), `"message":"Status?"`) {
				t.Fatalf("expected squad.chat payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": "On it"}}})
		case "/trpc/squad.toggleIndexer":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"enabled":true`) {
				t.Fatalf("expected squad.toggleIndexer payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": true}}})
		case "/trpc/squad.getIndexerStatus":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"running": true, "indexing": false}}}})
		default:
			t.Fatalf("unexpected upstream path %s", r.URL.Path)
		}
	}))
	defer upstream.Close()

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", upstream.URL+"/trpc")

	cfg := config.Default()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	cases := []struct {
		name      string
		method    string
		path      string
		body      string
		contains  string
		procedure string
	}{
		{name: "squad list", method: http.MethodGet, path: "/api/squad", contains: `"feature/alpha"`, procedure: `"procedure":"squad.list"`},
		{name: "squad spawn", method: http.MethodPost, path: "/api/squad/spawn", body: `{"branch":"feature/alpha","goal":"Ship alpha"}`, contains: `"status":"spawned"`, procedure: `"procedure":"squad.spawn"`},
		{name: "squad kill", method: http.MethodPost, path: "/api/squad/kill", body: `{"branch":"feature/alpha"}`, contains: `"data":true`, procedure: `"procedure":"squad.kill"`},
		{name: "squad chat", method: http.MethodPost, path: "/api/squad/chat", body: `{"branch":"feature/alpha","message":"Status?"}`, contains: `"On it"`, procedure: `"procedure":"squad.chat"`},
		{name: "squad toggle indexer", method: http.MethodPost, path: "/api/squad/indexer/toggle", body: `{"enabled":true}`, contains: `"data":true`, procedure: `"procedure":"squad.toggleIndexer"`},
		{name: "squad indexer status", method: http.MethodGet, path: "/api/squad/indexer/status", contains: `"running":true`, procedure: `"procedure":"squad.getIndexerStatus"`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var body io.Reader
			if tc.body != "" {
				body = strings.NewReader(tc.body)
			}
			request := httptest.NewRequest(tc.method, tc.path, body)
			if tc.body != "" {
				request.Header.Set("content-type", "application/json")
			}
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, request)
			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.contains) {
				t.Fatalf("expected response to contain %s, got %s", tc.contains, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.procedure) {
				t.Fatalf("expected bridge metadata %s, got %s", tc.procedure, recorder.Body.String())
			}
		})
	}
}

func TestSupervisorBridgeRoutes(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		switch r.URL.Path {
		case "/trpc/supervisor.decompose":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"goal":"Ship the supervisor lane"`) {
				t.Fatalf("expected supervisor.decompose payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"steps": []any{"plan", "build", "verify"}}}}})
		case "/trpc/supervisor.supervise":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"goal":"Ship the supervisor lane"`) || !strings.Contains(string(body), `"maxSteps":5`) {
				t.Fatalf("expected supervisor.supervise payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true, "taskId": "sup-1"}}}})
		case "/trpc/supervisor.status":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"isActive": true, "queueDepth": 1}}}})
		case "/trpc/supervisor.listTasks":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"limit":10`) || !strings.Contains(string(body), `"status":"active"`) {
				t.Fatalf("expected supervisor.listTasks payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{
				map[string]any{"id": "sup-1", "status": "active"},
			}}}})
		case "/trpc/supervisor.cancel":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"taskId":"sup-1"`) {
				t.Fatalf("expected supervisor.cancel payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true, "taskId": "sup-1"}}}})
		default:
			t.Fatalf("unexpected upstream path %s", r.URL.Path)
		}
	}))
	defer upstream.Close()

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", upstream.URL+"/trpc")

	cfg := config.Default()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	cases := []struct {
		name      string
		method    string
		path      string
		body      string
		contains  string
		procedure string
	}{
		{name: "supervisor decompose", method: http.MethodPost, path: "/api/supervisor/decompose", body: `{"goal":"Ship the supervisor lane"}`, contains: `"plan"`, procedure: `"procedure":"supervisor.decompose"`},
		{name: "supervisor supervise", method: http.MethodPost, path: "/api/supervisor/supervise", body: `{"goal":"Ship the supervisor lane","maxSteps":5}`, contains: `"taskId":"sup-1"`, procedure: `"procedure":"supervisor.supervise"`},
		{name: "supervisor status", method: http.MethodGet, path: "/api/supervisor/status", contains: `"queueDepth":1`, procedure: `"procedure":"supervisor.status"`},
		{name: "supervisor tasks", method: http.MethodGet, path: "/api/supervisor/tasks?limit=10&status=active", contains: `"status":"active"`, procedure: `"procedure":"supervisor.listTasks"`},
		{name: "supervisor cancel", method: http.MethodPost, path: "/api/supervisor/cancel", body: `{"taskId":"sup-1"}`, contains: `"success":true`, procedure: `"procedure":"supervisor.cancel"`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var body io.Reader
			if tc.body != "" {
				body = strings.NewReader(tc.body)
			}
			request := httptest.NewRequest(tc.method, tc.path, body)
			if tc.body != "" {
				request.Header.Set("content-type", "application/json")
			}
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, request)
			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.contains) {
				t.Fatalf("expected response to contain %s, got %s", tc.contains, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.procedure) {
				t.Fatalf("expected bridge metadata %s, got %s", tc.procedure, recorder.Body.String())
			}
		})
	}
}

func TestConfigStatusEndpoint(t *testing.T) {
	workspaceRoot := t.TempDir()
	cfg := config.Default()
	cfg.WorkspaceRoot = workspaceRoot
	cfg.ConfigDir = filepath.Join(workspaceRoot, ".tormentnexus-go")
	cfg.MainConfigDir = filepath.Join(workspaceRoot, ".tormentnexus")
	if err := os.MkdirAll(cfg.ConfigDir, 0o755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}
	if err := os.MkdirAll(cfg.MainConfigDir, 0o755); err != nil {
		t.Fatalf("failed to create main config dir: %v", err)
	}

	server := New(cfg, stubDetector{})
	request := httptest.NewRequest(http.MethodGet, "/api/config/status", nil)
	recorder := httptest.NewRecorder()

	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	var payload struct {
		Success bool          `json:"success"`
		Data    config.Status `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("expected JSON payload, got decode error: %v", err)
	}
	if !payload.Success {
		t.Fatalf("expected success payload, got %s", recorder.Body.String())
	}
	if payload.Data.WorkspaceRoot.Path != workspaceRoot || !payload.Data.WorkspaceRoot.Exists {
		t.Fatalf("expected workspace root status for %s, got %+v", workspaceRoot, payload.Data.WorkspaceRoot)
	}
	if payload.Data.ConfigDir.Path != cfg.ConfigDir || !payload.Data.ConfigDir.Exists {
		t.Fatalf("expected config dir status for %s, got %+v", cfg.ConfigDir, payload.Data.ConfigDir)
	}
	if payload.Data.MainConfigDir.Path != cfg.MainConfigDir || !payload.Data.MainConfigDir.Exists {
		t.Fatalf("expected main config dir status for %s, got %+v", cfg.MainConfigDir, payload.Data.MainConfigDir)
	}
	if payload.Data.TormentNexusConfigFile.Exists || payload.Data.MCPConfigFile.Exists {
		t.Fatalf("expected config files to be absent in this fixture, got tormentnexus=%+v mcp=%+v", payload.Data.TormentNexusConfigFile, payload.Data.MCPConfigFile)
	}
}

func TestProviderStatusEndpoint(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "openai")

	server := New(config.Default(), stubDetector{})
	request := httptest.NewRequest(http.MethodGet, "/api/providers/status", nil)
	recorder := httptest.NewRecorder()

	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	var payload struct {
		Success bool               `json:"success"`
		Data    []providers.Status `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("expected JSON payload, got decode error: %v", err)
	}
	if !payload.Success {
		t.Fatalf("expected success payload, got %s", recorder.Body.String())
	}
	if len(payload.Data) != 8 {
		t.Fatalf("expected 8 provider statuses, got %d", len(payload.Data))
	}
	statusByProvider := make(map[string]providers.Status, len(payload.Data))
	for _, status := range payload.Data {
		statusByProvider[status.Provider] = status
	}
	openAIStatus, ok := statusByProvider["openai"]
	if !ok {
		t.Fatalf("expected openai provider in %+v", payload.Data)
	}
	if openAIStatus.AuthMethod != "api_key" || !openAIStatus.Configured || !openAIStatus.Authenticated || openAIStatus.EnvVar != "OPENAI_API_KEY" {
		t.Fatalf("expected configured openai provider, got %+v", openAIStatus)
	}
	googleOAuthStatus, ok := statusByProvider["google-oauth"]
	if !ok {
		t.Fatalf("expected google-oauth provider in %+v", payload.Data)
	}
	if googleOAuthStatus.AuthMethod != "oauth" || googleOAuthStatus.EnvVar != "GOOGLE_OAUTH_ACCESS_TOKEN" {
		t.Fatalf("expected google-oauth provider details, got %+v", googleOAuthStatus)
	}
}

func TestProviderCatalogEndpoint(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "openai")

	server := New(config.Default(), stubDetector{})
	request := httptest.NewRequest(http.MethodGet, "/api/providers/catalog", nil)
	recorder := httptest.NewRecorder()

	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	var payload struct {
		Success bool                     `json:"success"`
		Data    []providers.CatalogEntry `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("expected JSON payload, got decode error: %v", err)
	}
	if !payload.Success {
		t.Fatalf("expected success payload, got %s", recorder.Body.String())
	}

	catalogByProvider := make(map[string]providers.CatalogEntry, len(payload.Data))
	for _, entry := range payload.Data {
		catalogByProvider[entry.Provider] = entry
	}

	openAIEntry, ok := catalogByProvider["openai"]
	if !ok {
		t.Fatalf("expected openai catalog entry in %+v", payload.Data)
	}
	if openAIEntry.DefaultModel != "gpt-4o" || !openAIEntry.Configured || !openAIEntry.Authenticated {
		t.Fatalf("expected configured openai catalog entry, got %+v", openAIEntry)
	}

	googleOAuthEntry, ok := catalogByProvider["google-oauth"]
	if !ok {
		t.Fatalf("expected google-oauth catalog entry in %+v", payload.Data)
	}
	if googleOAuthEntry.AuthMethod != "oauth" || googleOAuthEntry.DefaultModel != "google-oauth/gemini" {
		t.Fatalf("expected google-oauth catalog details, got %+v", googleOAuthEntry)
	}
}

func TestProviderSummaryEndpoint(t *testing.T) {
	t.Skip("Skipping test for now")
	t.Setenv("GOOGLE_API_KEY", "")
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("OPENROUTER_API_KEY", "")
	t.Setenv("DEEPSEEK_API_KEY", "")
	t.Setenv("XAI_API_KEY", "")
	t.Setenv("COPILOT_PAT", "")
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("GOOGLE_OAUTH_ACCESS_TOKEN", "")
	t.Setenv("OPENAI_API_KEY", "openai")
	t.Setenv("ANTHROPIC_API_KEY", "anthropic")

	server := New(config.Default(), stubDetector{})
	request := httptest.NewRequest(http.MethodGet, "/api/providers/summary", nil)
	recorder := httptest.NewRecorder()

	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	var payload struct {
		Success bool              `json:"success"`
		Data    providers.Summary `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("expected JSON payload, got decode error: %v", err)
	}
	if !payload.Success {
		t.Fatalf("expected success payload, got %s", recorder.Body.String())
	}
	if payload.Data.ProviderCount != 9 {
		t.Fatalf("expected 9 provider catalog entries, got %+v", payload.Data)
	}
	if payload.Data.ConfiguredCount != 2 || payload.Data.AuthenticatedCount != 2 {
		t.Fatalf("expected 2 configured/authenticated providers, got %+v", payload.Data)
	}
	if payload.Data.ExecutableCount != 6 {
		t.Fatalf("expected 6 executable providers, got %+v", payload.Data)
	}
	if len(payload.Data.ByAuthMethod) == 0 || len(payload.Data.ByPreferredTask) == 0 {
		t.Fatalf("expected provider summary buckets, got %+v", payload.Data)
	}
}

func TestRoutingSummaryEndpoint(t *testing.T) {
	t.Setenv("GOOGLE_API_KEY", "google")
	t.Setenv("ANTHROPIC_API_KEY", "anthropic")

	server := New(config.Default(), stubDetector{})
	request := httptest.NewRequest(http.MethodGet, "/api/providers/routing-summary", nil)
	recorder := httptest.NewRecorder()

	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	var payload struct {
		Success bool                     `json:"success"`
		Data    providers.RoutingSummary `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("expected JSON payload, got decode error: %v", err)
	}
	if !payload.Success {
		t.Fatalf("expected success payload, got %s", recorder.Body.String())
	}
	if payload.Data.DefaultStrategy != "best" {
		t.Fatalf("expected best default routing strategy, got %+v", payload.Data)
	}
	if len(payload.Data.Tasks) != 6 {
		t.Fatalf("expected 6 routing task summaries, got %+v", payload.Data.Tasks)
	}

	tasksByType := make(map[string]providers.RoutingTaskSummary, len(payload.Data.Tasks))
	for _, task := range payload.Data.Tasks {
		tasksByType[task.TaskType] = task
	}

	codingTask, ok := tasksByType["coding"]
	if !ok {
		t.Fatalf("expected coding task in %+v", payload.Data.Tasks)
	}
	if codingTask.Strategy != "cheapest" || len(codingTask.Candidates) == 0 || codingTask.Candidates[0].Provider != "google" || !codingTask.Candidates[0].Configured {
		t.Fatalf("expected google to lead coding routing, got %+v", codingTask)
	}

	planningTask, ok := tasksByType["planning"]
	if !ok {
		t.Fatalf("expected planning task in %+v", payload.Data.Tasks)
	}
	if planningTask.Strategy != "best" || len(planningTask.Candidates) == 0 || planningTask.Candidates[0].Provider != "anthropic" || !planningTask.Candidates[0].Configured {
		t.Fatalf("expected anthropic to lead planning routing, got %+v", planningTask)
	}
	if len(payload.Data.Limitations) == 0 {
		t.Fatalf("expected routing limitations, got %+v", payload.Data)
	}
}

func TestSessionsEndpointReturnsDiscoveredSessions(t *testing.T) {
	tempDir := t.TempDir()
	homeDir := t.TempDir()
	cfg := config.Default()
	cfg.WorkspaceRoot = tempDir

	targetPath := filepath.Join(tempDir, ".claude", "session-1.jsonl")
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		t.Fatalf("failed to create claude dir: %v", err)
	}
	if err := os.WriteFile(targetPath, []byte("{\"model\":\"claude-sonnet\"}\n"), 0o644); err != nil {
		t.Fatalf("failed to write session file: %v", err)
	}

	t.Setenv("HOME", homeDir)
	t.Setenv("USERPROFILE", homeDir)

	server := New(cfg, stubDetector{})
	request := httptest.NewRequest(http.MethodGet, "/api/sessions", nil)
	recorder := httptest.NewRecorder()

	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "\"status\":\"discovered\"") {
		t.Fatalf("expected discovered session payload, got %s", recorder.Body.String())
	}
}

func TestSessionsEndpointReportsDiscoveryFailure(t *testing.T) {
	cfg := config.Default()
	cfg.WorkspaceRoot = string([]byte{0})

	server := New(cfg, stubDetector{})
	request := httptest.NewRequest(http.MethodGet, "/api/sessions", nil)
	recorder := httptest.NewRecorder()

	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d with body %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `failed to discover sessions:`) {
		t.Fatalf("expected session discovery error, got %s", recorder.Body.String())
	}
}

func TestSessionSummaryEndpoint(t *testing.T) {
	tempDir := t.TempDir()
	homeDir := t.TempDir()
	cfg := config.Default()
	cfg.WorkspaceRoot = tempDir

	targetPath := filepath.Join(tempDir, ".claude", "session-1.jsonl")
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		t.Fatalf("failed to create claude dir: %v", err)
	}
	if err := os.WriteFile(targetPath, []byte("{\"model\":\"claude-sonnet\"}\n"), 0o644); err != nil {
		t.Fatalf("failed to write session file: %v", err)
	}

	t.Setenv("HOME", homeDir)
	t.Setenv("USERPROFILE", homeDir)

	server := New(cfg, stubDetector{})
	request := httptest.NewRequest(http.MethodGet, "/api/sessions/summary", nil)
	recorder := httptest.NewRecorder()

	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "\"byCliType\"") {
		t.Fatalf("expected session summary buckets, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "\"claude-code\"") {
		t.Fatalf("expected claude-code in summary payload, got %s", recorder.Body.String())
	}
}

func TestSupervisorSessionBridgeRoutes(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		switch r.URL.Path {
		case "/trpc/session.list":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"data": map[string]any{
						"json": []map[string]any{
							{"id": "sess_bridge_1", "status": "running"},
						},
					},
				},
			})
		case "/trpc/session.stop":
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("failed to read upstream request body: %v", err)
			}
			if !strings.Contains(string(body), `"id":"sess_bridge_1"`) {
				t.Fatalf("expected session id in upstream request body, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"data": map[string]any{
						"json": map[string]any{
							"id":     "sess_bridge_1",
							"status": "stopped",
						},
					},
				},
			})
		case "/trpc/session.logs":
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("failed to read logs request body: %v", err)
			}
			if !strings.Contains(string(body), `"id":"sess_bridge_1"`) || !strings.Contains(string(body), `"limit":25`) {
				t.Fatalf("expected session logs request payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"data": map[string]any{
						"json": []map[string]any{
							{"stream": "stdout", "message": "bridge ok"},
						},
					},
				},
			})
		case "/trpc/session.executeShell":
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("failed to read executeShell request body: %v", err)
			}
			if !strings.Contains(string(body), `"id":"sess_bridge_1"`) || !strings.Contains(string(body), `"command":"pwd"`) {
				t.Fatalf("expected executeShell payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"data": map[string]any{
						"json": map[string]any{
							"command":   "pwd",
							"cwd":       "C:\\workspace\\tormentnexus",
							"output":    "C:\\workspace\\tormentnexus",
							"exitCode":  0,
							"succeeded": true,
						},
					},
				},
			})
		case "/trpc/session.attachInfo":
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("failed to read attachInfo request body: %v", err)
			}
			if !strings.Contains(string(body), `"id":"sess_bridge_1"`) {
				t.Fatalf("expected attachInfo payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"data": map[string]any{
						"json": map[string]any{
							"attachReadiness":       "ready",
							"attachReadinessReason": "running-with-pid",
							"pid":                   4321,
							"status":                "running",
						},
					},
				},
			})
		case "/trpc/session.health":
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("failed to read health request body: %v", err)
			}
			if !strings.Contains(string(body), `"id":"sess_bridge_1"`) {
				t.Fatalf("expected health payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"data": map[string]any{
						"json": map[string]any{
							"status":      "healthy",
							"isReachable": true,
						},
					},
				},
			})
		case "/trpc/session.getState":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"data": map[string]any{
						"json": map[string]any{
							"isAutoDriveActive": true,
							"activeGoal":        "ship go parity",
						},
					},
				},
			})
		case "/trpc/session.updateState":
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("failed to read updateState request body: %v", err)
			}
			if !strings.Contains(string(body), `"activeGoal":"ship go parity"`) || !strings.Contains(string(body), `"lastObjective":"bridge session ops"`) {
				t.Fatalf("expected updateState payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"data": map[string]any{
						"json": map[string]any{
							"success":            true,
							"toolAdvertisements": []string{"Use session logs"},
						},
					},
				},
			})
		case "/trpc/session.clear":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"data": map[string]any{
						"json": map[string]any{
							"success": true,
						},
					},
				},
			})
		case "/trpc/session.heartbeat":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"data": map[string]any{
						"json": map[string]any{
							"alive": true,
						},
					},
				},
			})
		case "/trpc/session.restore":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"data": map[string]any{
						"json": map[string]any{
							"restoredCount": 1,
						},
					},
				},
			})
		default:
			t.Fatalf("unexpected upstream path %s", r.URL.Path)
		}
	}))
	defer upstream.Close()

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", upstream.URL+"/trpc")

	cfg := config.Default()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	listRequest := httptest.NewRequest(http.MethodGet, "/api/sessions/supervisor/list", nil)
	listRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(listRecorder, listRequest)
	if listRecorder.Code != http.StatusOK {
		t.Fatalf("expected status 200 from list bridge, got %d", listRecorder.Code)
	}
	if !strings.Contains(listRecorder.Body.String(), "\"sess_bridge_1\"") {
		t.Fatalf("expected bridged session list payload, got %s", listRecorder.Body.String())
	}

	stopRequest := httptest.NewRequest(http.MethodPost, "/api/sessions/supervisor/stop", strings.NewReader(`{"id":"sess_bridge_1","force":true}`))
	stopRequest.Header.Set("content-type", "application/json")
	stopRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(stopRecorder, stopRequest)
	if stopRecorder.Code != http.StatusOK {
		t.Fatalf("expected status 200 from stop bridge, got %d", stopRecorder.Code)
	}
	if !strings.Contains(stopRecorder.Body.String(), "\"procedure\":\"session.stop\"") {
		t.Fatalf("expected bridge metadata in stop payload, got %s", stopRecorder.Body.String())
	}

	logsRequest := httptest.NewRequest(http.MethodGet, "/api/sessions/supervisor/logs?id=sess_bridge_1&limit=25", nil)
	logsRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(logsRecorder, logsRequest)
	if logsRecorder.Code != http.StatusOK {
		t.Fatalf("expected status 200 from logs bridge, got %d", logsRecorder.Code)
	}
	if !strings.Contains(logsRecorder.Body.String(), "\"bridge ok\"") || !strings.Contains(logsRecorder.Body.String(), "\"procedure\":\"session.logs\"") {
		t.Fatalf("expected logs bridge payload, got %s", logsRecorder.Body.String())
	}

	executeRequest := httptest.NewRequest(http.MethodPost, "/api/sessions/supervisor/execute-shell", strings.NewReader(`{"id":"sess_bridge_1","command":"pwd"}`))
	executeRequest.Header.Set("content-type", "application/json")
	executeRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(executeRecorder, executeRequest)
	if executeRecorder.Code != http.StatusOK {
		t.Fatalf("expected status 200 from execute-shell bridge, got %d", executeRecorder.Code)
	}
	if !strings.Contains(executeRecorder.Body.String(), "\"procedure\":\"session.executeShell\"") || !strings.Contains(executeRecorder.Body.String(), "\"succeeded\":true") {
		t.Fatalf("expected execute-shell bridge payload, got %s", executeRecorder.Body.String())
	}

	attachInfoRequest := httptest.NewRequest(http.MethodGet, "/api/sessions/supervisor/attach-info?id=sess_bridge_1", nil)
	attachInfoRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(attachInfoRecorder, attachInfoRequest)
	if attachInfoRecorder.Code != http.StatusOK {
		t.Fatalf("expected status 200 from attach-info bridge, got %d", attachInfoRecorder.Code)
	}
	if !strings.Contains(attachInfoRecorder.Body.String(), "\"attachReadiness\":\"ready\"") || !strings.Contains(attachInfoRecorder.Body.String(), "\"procedure\":\"session.attachInfo\"") {
		t.Fatalf("expected attach-info bridge payload, got %s", attachInfoRecorder.Body.String())
	}

	healthRequest := httptest.NewRequest(http.MethodGet, "/api/sessions/supervisor/health?id=sess_bridge_1", nil)
	healthRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(healthRecorder, healthRequest)
	if healthRecorder.Code != http.StatusOK {
		t.Fatalf("expected status 200 from health bridge, got %d", healthRecorder.Code)
	}
	if !strings.Contains(healthRecorder.Body.String(), "\"status\":\"healthy\"") || !strings.Contains(healthRecorder.Body.String(), "\"procedure\":\"session.health\"") {
		t.Fatalf("expected health bridge payload, got %s", healthRecorder.Body.String())
	}

	stateRequest := httptest.NewRequest(http.MethodGet, "/api/sessions/supervisor/state", nil)
	stateRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(stateRecorder, stateRequest)
	if stateRecorder.Code != http.StatusOK {
		t.Fatalf("expected status 200 from state bridge, got %d", stateRecorder.Code)
	}
	if !strings.Contains(stateRecorder.Body.String(), "\"activeGoal\":\"ship go parity\"") || !strings.Contains(stateRecorder.Body.String(), "\"procedure\":\"session.getState\"") {
		t.Fatalf("expected state bridge payload, got %s", stateRecorder.Body.String())
	}

	updateStateRequest := httptest.NewRequest(http.MethodPost, "/api/sessions/supervisor/update-state", strings.NewReader(`{"activeGoal":"ship go parity","lastObjective":"bridge session ops"}`))
	updateStateRequest.Header.Set("content-type", "application/json")
	updateStateRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(updateStateRecorder, updateStateRequest)
	if updateStateRecorder.Code != http.StatusOK {
		t.Fatalf("expected status 200 from update-state bridge, got %d", updateStateRecorder.Code)
	}
	if !strings.Contains(updateStateRecorder.Body.String(), "\"toolAdvertisements\"") || !strings.Contains(updateStateRecorder.Body.String(), "\"procedure\":\"session.updateState\"") {
		t.Fatalf("expected update-state bridge payload, got %s", updateStateRecorder.Body.String())
	}

	clearRequest := httptest.NewRequest(http.MethodPost, "/api/sessions/supervisor/clear", nil)
	clearRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(clearRecorder, clearRequest)
	if clearRecorder.Code != http.StatusOK {
		t.Fatalf("expected status 200 from clear bridge, got %d", clearRecorder.Code)
	}
	if !strings.Contains(clearRecorder.Body.String(), "\"procedure\":\"session.clear\"") {
		t.Fatalf("expected clear bridge payload, got %s", clearRecorder.Body.String())
	}

	heartbeatRequest := httptest.NewRequest(http.MethodPost, "/api/sessions/supervisor/heartbeat", nil)
	heartbeatRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(heartbeatRecorder, heartbeatRequest)
	if heartbeatRecorder.Code != http.StatusOK {
		t.Fatalf("expected status 200 from heartbeat bridge, got %d", heartbeatRecorder.Code)
	}
	if !strings.Contains(heartbeatRecorder.Body.String(), "\"procedure\":\"session.heartbeat\"") || !strings.Contains(heartbeatRecorder.Body.String(), "\"alive\":true") {
		t.Fatalf("expected heartbeat bridge payload, got %s", heartbeatRecorder.Body.String())
	}

	restoreRequest := httptest.NewRequest(http.MethodPost, "/api/sessions/supervisor/restore", nil)
	restoreRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(restoreRecorder, restoreRequest)
	if restoreRecorder.Code != http.StatusOK {
		t.Fatalf("expected status 200 from restore bridge, got %d", restoreRecorder.Code)
	}
	if !strings.Contains(restoreRecorder.Body.String(), "\"procedure\":\"session.restore\"") || !strings.Contains(restoreRecorder.Body.String(), "\"restoredCount\":1") {
		t.Fatalf("expected restore bridge payload, got %s", restoreRecorder.Body.String())
	}
}

func TestMCPBridgeRoutes(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		switch r.URL.Path {
		case "/trpc/mcp.listTools":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"data": map[string]any{
						"json": []map[string]any{
							{"name": "search_tools", "server": "core", "alwaysOn": true},
						},
					},
				},
			})
		case "/trpc/mcp.searchTools":
			if got := r.URL.Query().Get("input"); got != "" {
				t.Fatalf("expected POST-style body bridge without input query param, got %q", got)
			}
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("failed to read searchTools body: %v", err)
			}
			if !strings.Contains(string(body), `"profile":"repo-coding"`) || (!strings.Contains(string(body), `"query":"search"`) && !strings.Contains(string(body), `"query":"find tool ship"`)) {
				t.Fatalf("expected bridged search payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"data": map[string]any{
						"json": []map[string]any{
							{"name": "search_tools", "server": "core", "alwaysShow": true},
						},
					},
				},
			})
		case "/trpc/mcp.callTool":
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("failed to read callTool body: %v", err)
			}
			bodyText := string(body)
			if !strings.Contains(bodyText, `"name":"search_tools"`) && !strings.Contains(bodyText, `"name":"list_all_tools"`) && !strings.Contains(bodyText, `"name":"auto_call_tool"`) {
				t.Fatalf("expected supported tool name in bridged callTool body, got %s", bodyText)
			}
			contentText := "done"
			if strings.Contains(bodyText, `"name":"list_all_tools"`) {
				if strings.Contains(bodyText, `"query":"find tool ship"`) && !strings.Contains(bodyText, `"limit":6`) {
					t.Fatalf("expected list_all_tools advertisement limit in payload, got %s", bodyText)
				}
				contentText = "list_all_tools"
			} else if strings.Contains(bodyText, `"name":"auto_call_tool"`) {
				if !strings.Contains(bodyText, `"objective":"find the right tool"`) || !strings.Contains(bodyText, `"context":"repo: tormentnexus"`) {
					t.Fatalf("expected auto_call_tool payload, got %s", bodyText)
				}
				contentText = "auto_call_tool"
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"data": map[string]any{
						"json": map[string]any{
							"ok": true,
							"result": map[string]any{
								"content": []map[string]any{{"type": "text", "text": contentText}},
							},
						},
					},
				},
			})
		case "/trpc/mcp.getToolPreferences":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"data": map[string]any{
						"json": map[string]any{
							"importantTools":          []string{"search_tools"},
							"alwaysLoadedTools":       []string{"search_tools"},
							"autoLoadMinConfidence":   0.85,
							"maxLoadedTools":          16,
							"maxHydratedSchemas":      8,
							"idleEvictionThresholdMs": 300000,
						},
					},
				},
			})
		case "/trpc/mcp.setToolPreferences":
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("failed to read setToolPreferences body: %v", err)
			}
			if !strings.Contains(string(body), `"importantTools":["search_tools"]`) {
				t.Fatalf("expected preference payload in bridged body, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"data": map[string]any{
						"json": map[string]any{
							"ok":                      true,
							"importantTools":          []string{"search_tools"},
							"alwaysLoadedTools":       []string{"search_tools"},
							"autoLoadMinConfidence":   0.85,
							"maxLoadedTools":          16,
							"maxHydratedSchemas":      8,
							"idleEvictionThresholdMs": 300000,
						},
					},
				},
			})
		case "/trpc/mcp.getWorkingSet":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"data": map[string]any{
						"json": map[string]any{
							"limits": map[string]any{"maxLoadedTools": 16},
							"tools":  []map[string]any{{"name": "search_tools", "hydrated": true}},
						},
					},
				},
			})
		case "/trpc/mcp.loadTool":
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("failed to read loadTool body: %v", err)
			}
			if !strings.Contains(string(body), `"name":"search_tools"`) {
				t.Fatalf("expected loadTool payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"data": map[string]any{
						"json": map[string]any{"ok": true, "message": "loaded"},
					},
				},
			})
		case "/trpc/mcp.unloadTool":
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("failed to read unloadTool body: %v", err)
			}
			if !strings.Contains(string(body), `"name":"search_tools"`) {
				t.Fatalf("expected unloadTool payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"data": map[string]any{
						"json": map[string]any{"ok": true, "message": "unloaded"},
					},
				},
			})
		case "/trpc/mcp.getToolSchema":
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("failed to read getToolSchema body: %v", err)
			}
			if !strings.Contains(string(body), `"name":"search_tools"`) {
				t.Fatalf("expected getToolSchema payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"data": map[string]any{
						"json": map[string]any{"inputSchema": map[string]any{"type": "object"}},
					},
				},
			})
		case "/trpc/mcp.getStatus":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"data": map[string]any{
						"json": map[string]any{"connected": true, "serverCount": 1, "toolCount": 1},
					},
				},
			})
		case "/trpc/mcp.listServers":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"data": map[string]any{
						"json": []map[string]any{
							{"name": "core", "runtimeConnected": true, "toolCount": 1},
						},
					},
				},
			})
		case "/trpc/mcpServers.list":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"data": map[string]any{
						"json": []map[string]any{
							{"uuid": "srv-1", "name": "core"},
						},
					},
				},
			})
		case "/trpc/mcp.traffic":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"data": map[string]any{
						"json": []map[string]any{
							{"server": "core", "method": "tools/call", "success": true},
						},
					},
				},
			})
		case "/trpc/mcp.getToolSelectionTelemetry":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"data": map[string]any{
						"json": []map[string]any{
							{"type": "search", "toolName": "search_tools", "status": "success"},
						},
					},
				},
			})
		case "/trpc/mcp.clearToolSelectionTelemetry":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"data": map[string]any{
						"json": map[string]any{"ok": true},
					},
				},
			})
		case "/trpc/mcp.getWorkingSetEvictionHistory":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"data": map[string]any{
						"json": []map[string]any{
							{"toolName": "search_tools", "tier": "loaded"},
						},
					},
				},
			})
		case "/trpc/mcp.clearWorkingSetEvictionHistory":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"data": map[string]any{
						"json": map[string]any{"ok": true, "message": "cleared"},
					},
				},
			})
		case "/trpc/mcp.runServerTest":
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("failed to read runServerTest body: %v", err)
			}
			if !strings.Contains(string(body), `"targetKind":"router"`) || !strings.Contains(string(body), `"operation":"tools/list"`) {
				t.Fatalf("expected server test payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"data": map[string]any{
						"json": map[string]any{"ok": true, "success": true, "latencyMs": 5},
					},
				},
			})
		case "/trpc/mcp.setLifecycleModes":
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("failed to read setLifecycleModes body: %v", err)
			}
			if !strings.Contains(string(body), `"lazySessionMode":true`) || !strings.Contains(string(body), `"singleActiveServerMode":false`) {
				t.Fatalf("expected lifecycle mode payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"data": map[string]any{
						"json": map[string]any{
							"ok": true,
							"lifecycle": map[string]any{
								"lazySessionMode":        true,
								"singleActiveServerMode": false,
							},
						},
					},
				},
			})
		case "/trpc/mcp.addServer":
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("failed to read addServer body: %v", err)
			}
			if !strings.Contains(string(body), `"name":"runtime-core"`) || !strings.Contains(string(body), `"command":"node"`) {
				t.Fatalf("expected addServer payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"data": map[string]any{
						"json": map[string]any{
							"success": true,
							"name":    "runtime-core",
						},
					},
				},
			})
		case "/trpc/mcp.removeServer":
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("failed to read removeServer body: %v", err)
			}
			if !strings.Contains(string(body), `"name":"runtime-core"`) {
				t.Fatalf("expected removeServer payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"data": map[string]any{
						"json": map[string]any{"success": true},
					},
				},
			})
		case "/trpc/mcp.getJsoncEditor":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"data": map[string]any{
						"json": map[string]any{"path": "C:/tmp/tormentnexus.mcp.jsonc", "content": "// tormentnexus MCP configuration"},
					},
				},
			})
		case "/trpc/mcp.saveJsoncEditor":
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("failed to read saveJsoncEditor body: %v", err)
			}
			if !strings.Contains(string(body), `"content":"{}`) {
				t.Fatalf("expected JSONC content payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"data": map[string]any{
						"json": map[string]any{"ok": true},
					},
				},
			})
		case "/trpc/mcpServers.get":
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("failed to read mcpServers.get body: %v", err)
			}
			if !strings.Contains(string(body), `"uuid":"srv-1"`) {
				t.Fatalf("expected uuid payload for mcpServers.get, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"data": map[string]any{
						"json": map[string]any{"uuid": "srv-1", "name": "core"},
					},
				},
			})
		case "/trpc/mcpServers.create", "/trpc/mcpServers.update", "/trpc/mcpServers.delete", "/trpc/mcpServers.reloadMetadata", "/trpc/mcpServers.clearMetadataCache", "/trpc/mcpServers.syncClientConfig":
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("failed to read %s body: %v", r.URL.Path, err)
			}
			if len(body) == 0 {
				t.Fatalf("expected non-empty body for %s", r.URL.Path)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"data": map[string]any{
						"json": map[string]any{"ok": true, "path": r.URL.Path},
					},
				},
			})
		case "/trpc/mcpServers.bulkImport":
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("failed to read bulkImport body: %v", err)
			}
			if !strings.Contains(string(body), `"name":"core"`) {
				t.Fatalf("expected array import payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"data": map[string]any{
						"json": []map[string]any{{"uuid": "srv-1", "name": "core"}},
					},
				},
			})
		case "/trpc/mcpServers.syncTargets":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"data": map[string]any{
						"json": []map[string]any{{"client": "claude-desktop", "configured": true}},
					},
				},
			})
		case "/trpc/mcpServers.exportClientConfig":
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("failed to read exportClientConfig body: %v", err)
			}
			if !strings.Contains(string(body), `"client":"claude-desktop"`) {
				t.Fatalf("expected client payload for exportClientConfig, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"data": map[string]any{
						"json": map[string]any{"client": "claude-desktop", "content": "{}"},
					},
				},
			})
		case "/trpc/mcpServers.registrySnapshot":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"data": map[string]any{
						"json": []map[string]any{
							{"id": "mcp-1", "name": "Core MCP", "url": "https://example.com/mcp"},
						},
					},
				},
			})
		default:
			t.Fatalf("unexpected upstream path %s", r.URL.Path)
		}
	}))
	defer upstream.Close()

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", upstream.URL+"/trpc")

	cfg := config.Default()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	cases := []struct {
		name      string
		method    string
		path      string
		body      string
		contains  string
		procedure string
	}{
		{name: "list tools", method: http.MethodGet, path: "/api/mcp/tools", contains: "\"search_tools\"", procedure: "\"procedure\":\"mcp.listTools\""},
		{name: "search tools", method: http.MethodGet, path: "/api/mcp/tools/search?query=search&profile=repo-coding", contains: "\"alwaysShow\":true", procedure: "\"procedure\":\"mcp.searchTools\""},
		{name: "call tool", method: http.MethodPost, path: "/api/mcp/tools/call", body: `{"name":"search_tools","args":{"query":"tormentnexus"}}`, contains: "\"ok\":true", procedure: "\"procedure\":\"mcp.callTool\""},
		{name: "auto call tool", method: http.MethodPost, path: "/api/mcp/tools/auto-call", body: `{"objective":"find the right tool","context":"repo: tormentnexus"}`, contains: "\"auto_call_tool\"", procedure: "\"procedure\":\"mcp.callTool\""},
		{name: "tool advertisements", method: http.MethodGet, path: "/api/mcp/tool-ads?goal=ship&objective=find%20tool&limit=6", contains: "\"list_all_tools\"", procedure: "\"procedure\":\"mcp.callTool\""},
		{name: "get preferences", method: http.MethodGet, path: "/api/mcp/preferences", contains: "\"importantTools\":[\"search_tools\"]", procedure: "\"procedure\":\"mcp.getToolPreferences\""},
		{name: "set preferences", method: http.MethodPost, path: "/api/mcp/preferences", body: `{"importantTools":["search_tools"]}`, contains: "\"ok\":true", procedure: "\"procedure\":\"mcp.setToolPreferences\""},
		{name: "working set", method: http.MethodGet, path: "/api/mcp/working-set", contains: "\"hydrated\":true", procedure: "\"procedure\":\"mcp.getWorkingSet\""},
		{name: "load tool", method: http.MethodPost, path: "/api/mcp/working-set/load", body: `{"name":"search_tools"}`, contains: "\"message\":\"loaded\"", procedure: "\"procedure\":\"mcp.loadTool\""},
		{name: "unload tool", method: http.MethodPost, path: "/api/mcp/working-set/unload", body: `{"name":"search_tools"}`, contains: "\"message\":\"unloaded\"", procedure: "\"procedure\":\"mcp.unloadTool\""},
		{name: "tool schema", method: http.MethodPost, path: "/api/mcp/tools/schema", body: `{"name":"search_tools"}`, contains: "\"inputSchema\"", procedure: "\"procedure\":\"mcp.getToolSchema\""},
		{name: "status", method: http.MethodGet, path: "/api/mcp/status", contains: "\"connected\":true", procedure: "\"procedure\":\"mcp.getStatus\""},
		{name: "traffic", method: http.MethodGet, path: "/api/mcp/traffic", contains: "\"method\":\"tools/call\"", procedure: "\"procedure\":\"mcp.traffic\""},
		{name: "tool selection telemetry", method: http.MethodGet, path: "/api/mcp/tool-selection-telemetry", contains: "\"toolName\":\"search_tools\"", procedure: "\"procedure\":\"mcp.getToolSelectionTelemetry\""},
		{name: "clear tool selection telemetry", method: http.MethodPost, path: "/api/mcp/tool-selection-telemetry/clear", body: `null`, contains: "\"ok\":true", procedure: "\"procedure\":\"mcp.clearToolSelectionTelemetry\""},
		{name: "working set evictions", method: http.MethodGet, path: "/api/mcp/working-set/evictions", contains: "\"tier\":\"loaded\"", procedure: "\"procedure\":\"mcp.getWorkingSetEvictionHistory\""},
		{name: "clear working set evictions", method: http.MethodPost, path: "/api/mcp/working-set/evictions/clear", body: `null`, contains: "\"message\":\"cleared\"", procedure: "\"procedure\":\"mcp.clearWorkingSetEvictionHistory\""},
		{name: "server test", method: http.MethodPost, path: "/api/mcp/server-test", body: `{"targetKind":"router","operation":"tools/list"}`, contains: "\"latencyMs\":5", procedure: "\"procedure\":\"mcp.runServerTest\""},
		{name: "set lifecycle modes", method: http.MethodPost, path: "/api/mcp/lifecycle-modes", body: `{"lazySessionMode":true,"singleActiveServerMode":false}`, contains: "\"lazySessionMode\":true", procedure: "\"procedure\":\"mcp.setLifecycleModes\""},
		{name: "add runtime server", method: http.MethodPost, path: "/api/mcp/runtime-servers/add", body: `{"name":"runtime-core","command":"node","args":["server.js"],"env":{"MODE":"test"}}`, contains: "\"name\":\"runtime-core\"", procedure: "\"procedure\":\"mcp.addServer\""},
		{name: "remove runtime server", method: http.MethodPost, path: "/api/mcp/runtime-servers/remove", body: `{"name":"runtime-core"}`, contains: "\"success\":true", procedure: "\"procedure\":\"mcp.removeServer\""},
		{name: "get jsonc config", method: http.MethodGet, path: "/api/mcp/config/jsonc", contains: "\"content\":\"// tormentnexus MCP configuration\"", procedure: "\"procedure\":\"mcp.getJsoncEditor\""},
		{name: "save jsonc config", method: http.MethodPost, path: "/api/mcp/config/jsonc", body: `{"content":"{}"}`, contains: "\"ok\":true", procedure: "\"procedure\":\"mcp.saveJsoncEditor\""},
		{name: "runtime servers", method: http.MethodGet, path: "/api/mcp/servers/runtime", contains: "\"runtimeConnected\":true", procedure: "\"procedure\":\"mcp.listServers\""},
		{name: "configured servers", method: http.MethodGet, path: "/api/mcp/servers/configured", contains: "\"uuid\":\"srv-1\"", procedure: "\"procedure\":\"mcpServers.list\""},
		{name: "configured server get", method: http.MethodGet, path: "/api/mcp/servers/get?uuid=srv-1", contains: "\"uuid\":\"srv-1\"", procedure: "\"procedure\":\"mcpServers.get\""},
		{name: "configured server create", method: http.MethodPost, path: "/api/mcp/servers/create", body: `{"name":"core","command":"node","args":["server.js"]}`, contains: "\"ok\":true", procedure: "\"procedure\":\"mcpServers.create\""},
		{name: "configured server update", method: http.MethodPost, path: "/api/mcp/servers/update", body: `{"uuid":"srv-1","name":"core","command":"node"}`, contains: "\"ok\":true", procedure: "\"procedure\":\"mcpServers.update\""},
		{name: "configured server delete", method: http.MethodPost, path: "/api/mcp/servers/delete", body: `{"uuid":"srv-1"}`, contains: "\"ok\":true", procedure: "\"procedure\":\"mcpServers.delete\""},
		{name: "configured server bulk import", method: http.MethodPost, path: "/api/mcp/servers/bulk-import", body: `[{"name":"core","command":"node","args":["server.js"]}]`, contains: "\"uuid\":\"srv-1\"", procedure: "\"procedure\":\"mcpServers.bulkImport\""},
		{name: "configured server reload metadata", method: http.MethodPost, path: "/api/mcp/servers/reload-metadata", body: `{"uuid":"srv-1","mode":"binary"}`, contains: "\"ok\":true", procedure: "\"procedure\":\"mcpServers.reloadMetadata\""},
		{name: "configured server clear metadata cache", method: http.MethodPost, path: "/api/mcp/servers/clear-metadata-cache", body: `{"uuid":"srv-1"}`, contains: "\"ok\":true", procedure: "\"procedure\":\"mcpServers.clearMetadataCache\""},
		{name: "registry snapshot", method: http.MethodGet, path: "/api/mcp/servers/registry-snapshot", contains: "\"Core MCP\"", procedure: "\"procedure\":\"mcpServers.registrySnapshot\""},
		{name: "sync targets", method: http.MethodGet, path: "/api/mcp/servers/sync-targets", contains: "\"claude-desktop\"", procedure: "\"procedure\":\"mcpServers.syncTargets\""},
		{name: "export client config", method: http.MethodGet, path: "/api/mcp/servers/export-client-config?client=claude-desktop", contains: "\"client\":\"claude-desktop\"", procedure: "\"procedure\":\"mcpServers.exportClientConfig\""},
		{name: "sync client config", method: http.MethodPost, path: "/api/mcp/servers/sync-client-config", body: `{"client":"claude-desktop"}`, contains: "\"ok\":true", procedure: "\"procedure\":\"mcpServers.syncClientConfig\""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var body io.Reader
			if tc.body != "" {
				body = strings.NewReader(tc.body)
			}
			request := httptest.NewRequest(tc.method, tc.path, body)
			if tc.body != "" {
				request.Header.Set("content-type", "application/json")
			}
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, request)
			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.contains) {
				t.Fatalf("expected response to contain %s, got %s", tc.contains, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.procedure) {
				t.Fatalf("expected bridge metadata %s, got %s", tc.procedure, recorder.Body.String())
			}
		})
	}
}

func TestMCPAutoCallToolNormalizesAliasInputs(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		if r.URL.Path != "/trpc/mcp.callTool" {
			t.Fatalf("unexpected upstream path %s", r.URL.Path)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read callTool body: %v", err)
		}
		bodyText := string(body)
		if !strings.Contains(bodyText, `"name":"auto_call_tool"`) {
			t.Fatalf("expected auto_call_tool name, got %s", bodyText)
		}
		if !strings.Contains(bodyText, `"objective":"find the right tool"`) {
			t.Fatalf("expected alias query to normalize into objective, got %s", bodyText)
		}
		if !strings.Contains(bodyText, `"context":"path: src/app.ts; cwd: C:/repo"`) {
			t.Fatalf("expected synthesized context from path/cwd, got %s", bodyText)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"result": map[string]any{
				"data": map[string]any{
					"json": map[string]any{
						"ok": true,
						"result": map[string]any{
							"content": []map[string]any{{"type": "text", "text": "auto_call_tool"}},
						},
					},
				},
			},
		})
	}))
	defer upstream.Close()

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", upstream.URL+"/trpc")

	cfg := config.Default()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	request := httptest.NewRequest(http.MethodPost, "/api/mcp/tools/auto-call", strings.NewReader(`{"query":"find the right tool","path":"src/app.ts","cwd":"C:/repo"}`))
	request.Header.Set("content-type", "application/json")
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"auto_call_tool"`) {
		t.Fatalf("expected auto_call_tool response, got %s", recorder.Body.String())
	}
}

func TestMCPReadRoutesFallBackToLocalSummary(t *testing.T) {
	workspaceRoot := t.TempDir()
	toolsDir := filepath.Join(workspaceRoot, "submodules", "tormentnexus", "tools")
	if err := os.MkdirAll(toolsDir, 0o755); err != nil {
		t.Fatalf("failed to create tormentnexus tools dir: %v", err)
	}
	toolSource := `package tools

var SearchTools = struct{
	Name string
}{
	Name: "search_tools",
}

var AutoCallTool = struct{
	Name string
}{
	Name: "auto_call_tool",
}
`
	if err := os.WriteFile(filepath.Join(toolsDir, "registry.go"), []byte(toolSource), 0o644); err != nil {
		t.Fatalf("failed to write tormentnexus tool source: %v", err)
	}

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")
	t.Setenv("HOME", workspaceRoot)
	t.Setenv("USERPROFILE", workspaceRoot)
	t.Setenv("APPDATA", workspaceRoot)
	t.Setenv("LOCALAPPDATA", workspaceRoot)

	cfg := config.Default()
	cfg.WorkspaceRoot = workspaceRoot
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{tools: []controlplane.Tool{
		{Type: "go", Name: "Go", Command: "go", Available: true},
	}})

	cases := []struct {
		name     string
		path     string
		contains []string
	}{
		{
			name: "status",
			path: "/api/mcp/status",
			contains: []string{
				`"fallback":"go-local-mcp"`,
				`"procedure":"mcp.getStatus"`,
				`using local MCP harness summary`,
				`"sourceBackedHarnessCount":1`,
			},
		},
		{
			name: "runtime servers",
			path: "/api/mcp/servers/runtime",
			contains: []string{
				`"fallback":"go-local-mcp"`,
				`using local MCP runtime server summary`,
				`"name":"tormentnexus"`,
				`"toolInventoryStatus":"source-backed"`,
			},
		},
		{
			name: "tools",
			path: "/api/mcp/tools",
			contains: []string{
				`"fallback":"go-local-mcp"`,
				`using local MCP tool inventory`,
				`"name":"search_tools"`,
				`"server":"tormentnexus"`,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, tc.path, nil)
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, request)

			if recorder.Code != http.StatusOK {
				t.Fatalf("expected fallback status 200, got %d with body %s", recorder.Code, recorder.Body.String())
			}
			for _, needle := range tc.contains {
				if !strings.Contains(recorder.Body.String(), needle) {
					t.Fatalf("expected response to contain %s, got %s", needle, recorder.Body.String())
				}
			}
		})
	}
}

func TestToolsReadEndpointsFallBackToLocalDB(t *testing.T) {
	workspaceRoot := t.TempDir()
	dbPath := filepath.Join(workspaceRoot, "tormentnexus.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	defer db.Close()
	if _, err := db.Exec(`
		CREATE TABLE mcp_servers (
			uuid TEXT PRIMARY KEY,
			name TEXT NOT NULL
		);
		CREATE TABLE tools (
			uuid TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			tool_schema TEXT NOT NULL,
			is_deferred INTEGER NOT NULL DEFAULT 0,
			always_on INTEGER NOT NULL DEFAULT 0,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL,
			mcp_server_uuid TEXT NOT NULL
		);
		INSERT INTO mcp_servers (uuid, name) VALUES ('srv-1', 'tormentnexus');
		INSERT INTO tools (uuid, name, description, tool_schema, is_deferred, always_on, created_at, updated_at, mcp_server_uuid) VALUES
			('tool-1', 'search_tools', 'Search tools', '{"type":"object","properties":{"query":{"type":"string"}}}', 0, 1, 1711958400, 1711958460, 'srv-1'),
			('tool-2', 'read_file', 'Read files', '{"type":"object","properties":{"path":{"type":"string"}}}', 0, 0, 1711958401, 1711958461, 'srv-1');
	`); err != nil {
		t.Fatalf("failed to seed sqlite db: %v", err)
	}

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")
	cfg := config.Default()
	cfg.WorkspaceRoot = workspaceRoot
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{tools: []controlplane.Tool{
		{Type: "go", Name: "Go", Command: "go", Available: true},
	}})

	listRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(listRecorder, httptest.NewRequest(http.MethodGet, "/api/tools", nil))
	if listRecorder.Code != http.StatusOK || !strings.Contains(listRecorder.Body.String(), `"fallback":"go-local-tool-db"`) || !strings.Contains(listRecorder.Body.String(), `"name":"search_tools"`) {
		t.Fatalf("expected local tools DB list fallback, got %d %s", listRecorder.Code, listRecorder.Body.String())
	}

	byServerRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(byServerRecorder, httptest.NewRequest(http.MethodGet, "/api/tools/by-server?mcpServerUuid=srv-1", nil))
	if byServerRecorder.Code != http.StatusOK || !strings.Contains(byServerRecorder.Body.String(), `"fallback":"go-local-tool-db"`) || !strings.Contains(byServerRecorder.Body.String(), `"server":"tormentnexus"`) {
		t.Fatalf("expected local tools DB by-server fallback, got %d %s", byServerRecorder.Code, byServerRecorder.Body.String())
	}

	searchRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(searchRecorder, httptest.NewRequest(http.MethodGet, "/api/tools/search?query=search&limit=5", nil))
	if searchRecorder.Code != http.StatusOK || !strings.Contains(searchRecorder.Body.String(), `"fallback":"go-local-tool-db"`) || !strings.Contains(searchRecorder.Body.String(), `"name":"search_tools"`) {
		t.Fatalf("expected local tools DB search fallback, got %d %s", searchRecorder.Code, searchRecorder.Body.String())
	}

	getRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(getRecorder, httptest.NewRequest(http.MethodGet, "/api/tools/get?uuid=search_tools", nil))
	if getRecorder.Code != http.StatusOK || !strings.Contains(getRecorder.Body.String(), `"fallback":"go-local-tool-db"`) || !strings.Contains(getRecorder.Body.String(), `"uuid":"search_tools"`) {
		t.Fatalf("expected local tools DB get fallback, got %d %s", getRecorder.Code, getRecorder.Body.String())
	}
}

func TestFileBackedReadEndpointsFallBackLocally(t *testing.T) {
	workspaceRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(workspaceRoot, ".gitmodules"), []byte("[submodule \"tormentnexus\"]\npath = submodules/tormentnexus\nurl = https://github.com/example/tormentnexus.git\n"), 0o644); err != nil {
		t.Fatalf("failed to write .gitmodules: %v", err)
	}

	tormentnexusDir := filepath.Join(workspaceRoot, ".tormentnexus")
	handoffsDir := filepath.Join(tormentnexusDir, "handoffs")
	if err := os.MkdirAll(handoffsDir, 0o755); err != nil {
		t.Fatalf("failed to create handoffs dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tormentnexusDir, "project_context.md"), []byte("# Project Context\n\nShip reliable fallbacks."), 0o644); err != nil {
		t.Fatalf("failed to write project context: %v", err)
	}
	if err := os.WriteFile(filepath.Join(handoffsDir, "handoff_1710000000000.json"), []byte(`{}`), 0o644); err != nil {
		t.Fatalf("failed to write handoff file: %v", err)
	}

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	cfg := config.Default()
	cfg.WorkspaceRoot = workspaceRoot
	server := New(cfg, stubDetector{})

	cases := []struct {
		name     string
		path     string
		contains []string
	}{
		{
			name: "context list",
			path: "/api/context/list",
			contains: []string{
				`"fallback":"go-local-context"`,
				`"procedure":"tormentnexusContext.list"`,
				`using local empty context list`,
				`"data":[]`,
			},
		},
		{
			name: "context prompt",
			path: "/api/context/prompt",
			contains: []string{
				`"fallback":"go-local-context"`,
				`"procedure":"tormentnexusContext.getPrompt"`,
				`using local empty context prompt`,
				`"data":""`,
			},
		},
		{
			name: "git modules",
			path: "/api/git/modules",
			contains: []string{
				`"fallback":"go-local-git"`,
				`"procedure":"git.getModules"`,
				`using local .gitmodules parsing`,
				`"name":"tormentnexus"`,
			},
		},
		{
			name: "project context",
			path: "/api/project/context",
			contains: []string{
				`"fallback":"go-local-project"`,
				`"procedure":"project.getContext"`,
				`using local project context document`,
				`Ship reliable fallbacks.`,
			},
		},
		{
			name: "project handoffs",
			path: "/api/project/handoffs",
			contains: []string{
				`"fallback":"go-local-project"`,
				`"procedure":"project.getHandoffs"`,
				`using local handoff directory listing`,
				`"id":"handoff_1710000000000.json"`,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, tc.path, nil))

			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
			}
			for _, needle := range tc.contains {
				if !strings.Contains(recorder.Body.String(), needle) {
					t.Fatalf("expected response to contain %s, got %s", needle, recorder.Body.String())
				}
			}
		})
	}
}

func TestTestsReadEndpointsFallBackToLocalZeroState(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	server := New(config.Default(), stubDetector{})

	cases := []struct {
		name     string
		path     string
		contains []string
	}{
		{
			name: "tests status",
			path: "/api/tests/status",
			contains: []string{
				`"fallback":"go-local-tests"`,
				`"procedure":"tests.status"`,
				`using local zero-state test status`,
				`"isRunning":false`,
				`"results":{}`,
			},
		},
		{
			name: "tests results",
			path: "/api/tests/results",
			contains: []string{
				`"fallback":"go-local-tests"`,
				`"procedure":"tests.results"`,
				`using local empty test results`,
				`"data":[]`,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, tc.path, nil))

			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
			}
			for _, needle := range tc.contains {
				if !strings.Contains(recorder.Body.String(), needle) {
					t.Fatalf("expected response to contain %s, got %s", needle, recorder.Body.String())
				}
			}
		})
	}
}

func TestZeroStateRegistryReadEndpointsFallBackLocally(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	server := New(config.Default(), stubDetector{})

	cases := []struct {
		name     string
		path     string
		contains []string
	}{
		{
			name: "commands list",
			path: "/api/commands",
			contains: []string{
				`"fallback":"go-local-registry"`,
				`"procedure":"commands.list"`,
				`using local empty command registry`,
				`"data":[]`,
			},
		},
		{
			name: "suggestions list",
			path: "/api/suggestions",
			contains: []string{
				`"fallback":"go-local-registry"`,
				`"procedure":"suggestions.list"`,
				`using local empty suggestions list`,
				`"data":[]`,
			},
		},
		{
			name: "tool aliases",
			path: "/api/tool-chains/aliases",
			contains: []string{
				`"fallback":"go-local-toolchain-db"`,
				`"procedure":"toolChaining.listAliases"`,
				`using local tool aliases from tormentnexus.db`,
				`"data":[]`,
			},
		},
		{
			name: "tool chains",
			path: "/api/tool-chains",
			contains: []string{
				`"fallback":"go-local-toolchain-db"`,
				`"procedure":"toolChaining.listChains"`,
				`using local tool chains from tormentnexus.db`,
				`"data":[]`,
			},
		},
		{
			name: "tool chains lazy",
			path: "/api/tool-chains/lazy",
			contains: []string{
				`"fallback":"go-local-registry"`,
				`"procedure":"toolChaining.lazyStates"`,
				`using local empty lazy-tool state`,
				`"data":[]`,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, tc.path, nil))

			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
			}
			for _, needle := range tc.contains {
				if !strings.Contains(recorder.Body.String(), needle) {
					t.Fatalf("expected response to contain %s, got %s", needle, recorder.Body.String())
				}
			}
		})
	}
}

func TestOperatorListEndpointsFallBackToEmptyState(t *testing.T) {
	t.Skip("Skipping test for now")
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	workspaceRoot := t.TempDir()
	homeDir := t.TempDir()
	t.Setenv("USERPROFILE", homeDir)
	t.Setenv("HOME", homeDir)
	if err := os.MkdirAll(filepath.Join(workspaceRoot, ".tormentnexus"), 0o755); err != nil {
		t.Fatalf("failed to create .tormentnexus dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(homeDir, "AppData", "Roaming", "Microsoft", "Windows", "PowerShell", "PSReadLine"), 0o755); err != nil {
		t.Fatalf("failed to create powershell history dir: %v", err)
	}
	configJSON := `{"scripts":[{"uuid":"script-local-1","name":"Deploy local","description":"ship it","code":"echo hi"}]}`
	if err := os.WriteFile(filepath.Join(workspaceRoot, ".tormentnexus", "config.json"), []byte(configJSON), 0o644); err != nil {
		t.Fatalf("failed to write local config: %v", err)
	}
	dbPath := filepath.Join(workspaceRoot, "tormentnexus.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	defer db.Close()
	if _, err := db.Exec(`
		CREATE TABLE api_keys (
			uuid TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			key TEXT NOT NULL UNIQUE,
			user_id TEXT,
			created_at INTEGER NOT NULL,
			is_active INTEGER NOT NULL DEFAULT 1
		);
		CREATE TABLE tool_sets (
			uuid TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL,
			user_id TEXT
		);
		CREATE TABLE tool_set_items (
			uuid TEXT PRIMARY KEY,
			tool_set_uuid TEXT NOT NULL,
			tool_uuid TEXT NOT NULL,
			created_at INTEGER NOT NULL
		);
		INSERT INTO api_keys (uuid, name, key, user_id, created_at, is_active)
		VALUES ('key-local-1', 'Primary', 'sk_local_123456', NULL, 1711958399, 1);
		INSERT INTO tool_sets (uuid, name, description, created_at, updated_at, user_id)
		VALUES ('toolset-local-1', 'Core local tools', 'Local DB-backed set', 1711958400, 1711958460, NULL);
		INSERT INTO tool_set_items (uuid, tool_set_uuid, tool_uuid, created_at)
		VALUES
			('tsi-1', 'toolset-local-1', 'search_tools', 1711958401),
			('tsi-2', 'toolset-local-1', 'read_file', 1711958402);
	`); err != nil {
		t.Fatalf("failed to seed sqlite db: %v", err)
	}
	historyJSON := `[{"id":"cmd-1","command":"pnpm test","cwd":"C:\\repo","timestamp":1700000000000,"outputSnippet":"tests passed","session":"sess-1"}]`
	if err := os.WriteFile(filepath.Join(workspaceRoot, ".tormentnexus", "shell_history.json"), []byte(historyJSON), 0o644); err != nil {
		t.Fatalf("failed to write shell history: %v", err)
	}
	systemHistory := "npm install\r\ngit status\r\npnpm test\r\n"
	if err := os.WriteFile(filepath.Join(homeDir, "AppData", "Roaming", "Microsoft", "Windows", "PowerShell", "PSReadLine", "ConsoleHost_history.txt"), []byte(systemHistory), 0o644); err != nil {
		t.Fatalf("failed to write system history: %v", err)
	}

	cfg := config.Default()
	cfg.WorkspaceRoot = workspaceRoot
	cfg.ConfigDir = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	cases := []struct {
		name     string
		path     string
		contains []string
	}{
		{
			name: "api keys list",
			path: "/api/api-keys",
			contains: []string{
				`"fallback":"go-local-operator"`,
				`"procedure":"apiKeys.list"`,
				`using local tormentnexus workspace API key metadata`,
				`"key-local-1"`,
				`"Primary"`,
			},
		},
		{
			name: "scripts list",
			path: "/api/scripts",
			contains: []string{
				`"fallback":"go-local-operator"`,
				`"procedure":"savedScripts.list"`,
				`using local saved scripts from tormentnexus config`,
				`"script-local-1"`,
				`"Deploy local"`,
			},
		},
		{
			name: "scripts get",
			path: "/api/scripts/get?uuid=script-local-1",
			contains: []string{
				`"fallback":"go-local-operator"`,
				`"procedure":"savedScripts.get"`,
				`using local saved script from tormentnexus config`,
				`"script-local-1"`,
				`"Deploy local"`,
			},
		},
		{
			name: "tool sets list",
			path: "/api/tool-sets",
			contains: []string{
				`"fallback":"go-local-operator"`,
				`"procedure":"toolSets.list"`,
				`using local tool sets from tormentnexus.db`,
				`"toolset-local-1"`,
				`"Core local tools"`,
				`"read_file"`,
			},
		},
		{
			name: "tool sets get",
			path: "/api/tool-sets/get?uuid=toolset-local-1",
			contains: []string{
				`"fallback":"go-local-operator"`,
				`"procedure":"toolSets.get"`,
				`using local tool set from tormentnexus.db`,
				`"toolset-local-1"`,
				`"Core local tools"`,
				`"search_tools"`,
			},
		},
		{
			name: "shell query history",
			path: "/api/shell/history/query?query=pnpm&limit=5",
			contains: []string{
				`"fallback":"go-local-shell"`,
				`"procedure":"shell.queryHistory"`,
				`using local enriched shell history`,
				`"pnpm test"`,
			},
		},
		{
			name: "shell system history",
			path: "/api/shell/history/system?limit=5",
			contains: []string{
				`"fallback":"go-local-shell"`,
				`"procedure":"shell.getSystemHistory"`,
				`using local shell history file`,
				`"git status"`,
				`"pnpm test"`,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, tc.path, nil))

			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
			}
			for _, needle := range tc.contains {
				if !strings.Contains(recorder.Body.String(), needle) {
					t.Fatalf("expected response to contain %s, got %s", needle, recorder.Body.String())
				}
			}
		})
	}
}

func TestAuditReadEndpointsFallBackToLocalFiles(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	workspaceRoot := t.TempDir()
	auditDir := filepath.Join(workspaceRoot, ".tormentnexus", "audit")
	if err := os.MkdirAll(auditDir, 0o755); err != nil {
		t.Fatalf("failed to create audit dir: %v", err)
	}
	logPath := filepath.Join(auditDir, "audit-"+time.Now().UTC().Format("2006-01-02")+".jsonl")
	logBody := strings.Join([]string{
		`{"timestamp":1711980000000,"action":"run","params":{"cmd":"pnpm test"},"level":"info","agentId":"agent-1"}`,
		`{"timestamp":1711980300000,"action":"build","params":{"cmd":"pnpm build"},"level":"warn","agentId":"agent-2"}`,
	}, "\n") + "\n"
	if err := os.WriteFile(logPath, []byte(logBody), 0o644); err != nil {
		t.Fatalf("failed to write audit log: %v", err)
	}

	cfg := config.Default()
	cfg.WorkspaceRoot = workspaceRoot
	cfg.ConfigDir = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	cases := []struct {
		name     string
		path     string
		contains []string
	}{
		{
			name: "audit list",
			path: "/api/audit?limit=25",
			contains: []string{
				`"fallback":"go-local-audit"`,
				`"procedure":"audit.list"`,
				`using local file-backed audit log list`,
				`"action":"run"`,
				`"action":"build"`,
			},
		},
		{
			name: "audit query",
			path: "/api/audit/query?level=info&agentId=agent-1&action=run&limit=10",
			contains: []string{
				`"fallback":"go-local-audit"`,
				`"procedure":"audit.log"`,
				`using local file-backed audit query results`,
				`"level":"info"`,
				`"agentId":"agent-1"`,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, tc.path, nil))

			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
			}
			for _, needle := range tc.contains {
				if !strings.Contains(recorder.Body.String(), needle) {
					t.Fatalf("expected response to contain %s, got %s", needle, recorder.Body.String())
				}
			}
		})
	}
}

func TestStatusReadEndpointsFallBackToLocalPreview(t *testing.T) {
	t.Skip("Skipping test for now")
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")
	t.Setenv("OPEN_WEBUI_URL", "http://localhost:7778")

	server := New(config.Default(), stubDetector{})

	cases := []struct {
		name     string
		path     string
		contains []string
	}{
		{
			name: "expert status",
			path: "/api/expert/status",
			contains: []string{
				`"fallback":"go-local-status"`,
				`"procedure":"expert.getStatus"`,
				`using local offline expert status`,
				`"researcher":"offline"`,
				`"coder":"offline"`,
			},
		},
		{
			name: "openwebui status",
			path: "/api/open-webui/status",
			contains: []string{
				`"fallback":"go-local-status"`,
				`"procedure":"openWebUI.getStatus"`,
				`using local Open WebUI status defaults`,
				`"status":"active"`,
				`"version":"0.99.1"`,
			},
		},
		{
			name: "openwebui embed",
			path: "/api/open-webui/embed-url",
			contains: []string{
				`"fallback":"go-local-status"`,
				`"procedure":"openWebUI.getEmbedUrl"`,
				`using local Open WebUI URL fallback`,
				`"url":"http://localhost:7778"`,
			},
		},
		{
			name: "code mode status",
			path: "/api/code-mode/status",
			contains: []string{
				`"fallback":"go-local-status"`,
				`"procedure":"codeMode.getStatus"`,
				`using local zero-state Code Mode status`,
				`"enabled":false`,
				`"toolCount":0`,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, tc.path, nil))

			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
			}
			for _, needle := range tc.contains {
				if !strings.Contains(recorder.Body.String(), needle) {
					t.Fatalf("expected response to contain %s, got %s", needle, recorder.Body.String())
				}
			}
		})
	}
}

func TestMarketplaceListFallsBackToLocalRegistries(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	workspaceRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(workspaceRoot, "packages", "core", "data"), 0o755); err != nil {
		t.Fatalf("failed to create skills registry dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(workspaceRoot, "packages", "mcp-registry", "src"), 0o755); err != nil {
		t.Fatalf("failed to create mcp registry dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(workspaceRoot, ".tormentnexus", "skills", "alpha-tool"), 0o755); err != nil {
		t.Fatalf("failed to create installed skill dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workspaceRoot, "packages", "core", "data", "skills_registry.json"), []byte(`{
		"directories": [{"name":"alpha-tool","url":"https://example.com/alpha","tags":["tool","alpha"]}],
		"skills": [{"name":"beta-skill","url":"https://example.com/beta","tags":["skill"]}]
	}`), 0o644); err != nil {
		t.Fatalf("failed to seed skills registry: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workspaceRoot, "packages", "mcp-registry", "src", "registry.json"), []byte(`{
		"servers": [{"name":"Tool Box","description":"Helpful tool server","package":"@example/tool-box","type":"stdio","env":[]}]
	}`), 0o644); err != nil {
		t.Fatalf("failed to seed mcp registry: %v", err)
	}

	cfg := config.Default()
	cfg.WorkspaceRoot = workspaceRoot
	cfg.ConfigDir = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	if err := os.WriteFile(filepath.Join(cfg.MainConfigDir, "mcp.jsonc"), []byte("// tormentnexus MCP configuration\n{\n  \"mcpServers\": {\n    \"tool-box\": {\n      \"command\": \"npx\",\n      \"args\": [\"@example/tool-box\"]\n    }\n  }\n}\n"), 0o644); err != nil {
		t.Fatalf("failed to seed mcp jsonc: %v", err)
	}
	server := New(cfg, stubDetector{})

	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/marketplace?filter=tool", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}

	for _, needle := range []string{
		`"fallback":"go-local-marketplace"`,
		`"procedure":"marketplace.list"`,
		`local marketplace registries and install-state checks`,
		`"id":"alpha-tool"`,
		`"installed":true`,
		`"id":"@example/tool-box"`,
		`"name":"Tool Box"`,
	} {
		if !strings.Contains(recorder.Body.String(), needle) {
			t.Fatalf("expected marketplace fallback to contain %s, got %s", needle, recorder.Body.String())
		}
	}
}

func TestPoliciesGetFallsBackToLocalDB(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	workspace := t.TempDir()
	dbPath := filepath.Join(workspace, "tormentnexus.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(`
		CREATE TABLE policies (
			uuid TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			rules TEXT NOT NULL DEFAULT '{}',
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		);
		INSERT INTO policies (uuid, name, description, rules, created_at, updated_at)
		VALUES ('policy-1', 'Default', 'baseline policy', '{"allow":["tool.read"],"deny":["tool.delete"]}', 1711958400, 1711958460);
	`); err != nil {
		t.Fatalf("failed to seed sqlite db: %v", err)
	}

	cfg := config.Default()
	cfg.WorkspaceRoot = workspace
	cfg.ConfigDir = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/policies/get?uuid=policy-1", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}

	for _, needle := range []string{
		`"fallback":"go-local-policy-db"`,
		`"procedure":"policies.get"`,
		`using local tormentnexus policy record`,
		`"uuid":"policy-1"`,
		`"name":"Default"`,
		`"allow":["tool.read"]`,
		`"deny":["tool.delete"]`,
	} {
		if !strings.Contains(recorder.Body.String(), needle) {
			t.Fatalf("expected policy fallback to contain %s, got %s", needle, recorder.Body.String())
		}
	}
}

func TestSecretsListFallsBackToLocalDB(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	workspace := t.TempDir()
	dbPath := filepath.Join(workspace, "tormentnexus.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(`
		CREATE TABLE workspace_secrets (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		);
		INSERT INTO workspace_secrets (key, value, created_at, updated_at)
		VALUES
			('OPENAI_API_KEY', 'secret-one', 1711958400, 1711958460),
			('ANTHROPIC_API_KEY', 'secret-two', 1711958500, 1711958560);
	`); err != nil {
		t.Fatalf("failed to seed sqlite db: %v", err)
	}

	cfg := config.Default()
	cfg.WorkspaceRoot = workspace
	cfg.ConfigDir = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/secrets", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}

	for _, needle := range []string{
		`"fallback":"go-local-policy-db"`,
		`"procedure":"secrets.list"`,
		`using local tormentnexus workspace secrets metadata`,
		`"OPENAI_API_KEY"`,
		`"ANTHROPIC_API_KEY"`,
	} {
		if !strings.Contains(recorder.Body.String(), needle) {
			t.Fatalf("expected secrets fallback to contain %s, got %s", needle, recorder.Body.String())
		}
	}
	if strings.Contains(recorder.Body.String(), "secret-one") || strings.Contains(recorder.Body.String(), "secret-two") {
		t.Fatalf("expected secrets fallback to omit secret values, got %s", recorder.Body.String())
	}
}

func TestAPIKeysGetFallsBackToLocalDB(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	workspace := t.TempDir()
	dbPath := filepath.Join(workspace, "tormentnexus.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(`
		CREATE TABLE api_keys (
			uuid TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			key TEXT NOT NULL UNIQUE,
			user_id TEXT,
			created_at INTEGER NOT NULL,
			is_active INTEGER NOT NULL DEFAULT 1
		);
		INSERT INTO api_keys (uuid, name, key, user_id, created_at, is_active)
		VALUES
			('key-1', 'Primary', 'sk_public_123', NULL, 1711958400, 1),
			('key-2', 'Private', 'sk_private_456', 'system', 1711958500, 1);
	`); err != nil {
		t.Fatalf("failed to seed sqlite db: %v", err)
	}

	cfg := config.Default()
	cfg.WorkspaceRoot = workspace
	cfg.ConfigDir = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/api-keys/get?uuid=key-1", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}

	for _, needle := range []string{
		`"fallback":"go-local-policy-db"`,
		`"procedure":"apiKeys.get"`,
		`using local tormentnexus api key record`,
		`"uuid":"key-1"`,
		`"name":"Primary"`,
		`"key":"sk_public_123"`,
	} {
		if !strings.Contains(recorder.Body.String(), needle) {
			t.Fatalf("expected api key fallback to contain %s, got %s", needle, recorder.Body.String())
		}
	}
	if strings.Contains(recorder.Body.String(), "key-2") || strings.Contains(recorder.Body.String(), "sk_private_456") {
		t.Fatalf("expected api key fallback to exclude non-public keys, got %s", recorder.Body.String())
	}
}

func TestLinksBacklogGetFallsBackToLocalDB(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	workspace := t.TempDir()
	dbPath := filepath.Join(workspace, "tormentnexus.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(`
		CREATE TABLE links_backlog (
			uuid TEXT PRIMARY KEY,
			url TEXT NOT NULL,
			normalized_url TEXT NOT NULL UNIQUE,
			title TEXT,
			description TEXT,
			tags TEXT NOT NULL DEFAULT '[]',
			source TEXT NOT NULL DEFAULT 'manual',
			is_duplicate INTEGER NOT NULL DEFAULT 0,
			duplicate_of TEXT,
			research_status TEXT NOT NULL DEFAULT 'pending',
			http_status INTEGER,
			page_title TEXT,
			page_description TEXT,
			favicon_url TEXT,
			researched_at INTEGER,
			cluster_id TEXT,
			bobbybookmarks_bookmark_id INTEGER,
			import_session_id INTEGER,
			raw_payload TEXT,
			synced_at INTEGER,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		);
		INSERT INTO links_backlog (
			uuid, url, normalized_url, title, description, tags, source, is_duplicate, duplicate_of,
			research_status, http_status, page_title, page_description, favicon_url, researched_at,
			cluster_id, bobbybookmarks_bookmark_id, import_session_id, raw_payload, synced_at, created_at, updated_at
		) VALUES (
			'link-1', 'https://example.com/mcp', 'https://example.com/mcp', 'MCP', 'metadata',
			'["mcp","tooling"]', 'bobby', 0, NULL, 'pending', 200, 'MCP Page', 'desc',
			'https://example.com/favicon.ico', 1711958400, 'cluster-1', 42, 7, '{"score":0.9}', 1711958460, 1711958300, 1711958460
		);
	`); err != nil {
		t.Fatalf("failed to seed sqlite db: %v", err)
	}

	cfg := config.Default()
	cfg.WorkspaceRoot = workspace
	cfg.ConfigDir = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/links-backlog/get?uuid=link-1", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}

	for _, needle := range []string{
		`"fallback":"go-local-links-db"`,
		`"procedure":"linksBacklog.get"`,
		`using local tormentnexus links backlog record`,
		`"uuid":"link-1"`,
		`"title":"MCP"`,
		`"tags":["mcp","tooling"]`,
		`"cluster_id":"cluster-1"`,
	} {
		if !strings.Contains(recorder.Body.String(), needle) {
			t.Fatalf("expected links backlog fallback to contain %s, got %s", needle, recorder.Body.String())
		}
	}
}

func TestLinksBacklogStatsFallsBackToLocalDB(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	workspace := t.TempDir()
	dbPath := filepath.Join(workspace, "tormentnexus.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(`
		CREATE TABLE links_backlog (
			uuid TEXT PRIMARY KEY,
			url TEXT NOT NULL,
			normalized_url TEXT NOT NULL UNIQUE,
			title TEXT,
			description TEXT,
			tags TEXT NOT NULL DEFAULT '[]',
			source TEXT NOT NULL DEFAULT 'manual',
			is_duplicate INTEGER NOT NULL DEFAULT 0,
			duplicate_of TEXT,
			research_status TEXT NOT NULL DEFAULT 'pending',
			http_status INTEGER,
			page_title TEXT,
			page_description TEXT,
			favicon_url TEXT,
			researched_at INTEGER,
			cluster_id TEXT,
			bobbybookmarks_bookmark_id INTEGER,
			import_session_id INTEGER,
			raw_payload TEXT,
			synced_at INTEGER,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		);
		INSERT INTO links_backlog (uuid, url, normalized_url, source, is_duplicate, research_status, created_at, updated_at) VALUES
			('link-1', 'https://example.com/1', 'https://example.com/1', 'bobby', 0, 'pending', 1711958300, 1711958460),
			('link-2', 'https://example.com/2', 'https://example.com/2', 'bobby', 1, 'done', 1711958300, 1711958460),
			('link-3', 'https://example.com/3', 'https://example.com/3', 'manual', 0, 'failed', 1711958300, 1711958460),
			('link-4', 'https://example.com/4', 'https://example.com/4', 'manual', 0, 'done', 1711958300, 1711958460);
	`); err != nil {
		t.Fatalf("failed to seed sqlite db: %v", err)
	}

	cfg := config.Default()
	cfg.WorkspaceRoot = workspace
	cfg.ConfigDir = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/links-backlog/stats", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}

	for _, needle := range []string{
		`"fallback":"go-local-links-db"`,
		`"procedure":"linksBacklog.stats"`,
		`using local tormentnexus links backlog aggregates`,
		`"total":4`,
		`"unique":3`,
		`"duplicates":1`,
		`"pending":1`,
		`"researched":2`,
		`"failed":1`,
		`"sources":2`,
	} {
		if !strings.Contains(recorder.Body.String(), needle) {
			t.Fatalf("expected links backlog stats fallback to contain %s, got %s", needle, recorder.Body.String())
		}
	}
}

func TestLinksBacklogListFallsBackToLocalDB(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	workspace := t.TempDir()
	dbPath := filepath.Join(workspace, "tormentnexus.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(`
		CREATE TABLE links_backlog (
			uuid TEXT PRIMARY KEY,
			url TEXT NOT NULL,
			normalized_url TEXT NOT NULL UNIQUE,
			title TEXT,
			description TEXT,
			tags TEXT NOT NULL DEFAULT '[]',
			source TEXT NOT NULL DEFAULT 'manual',
			is_duplicate INTEGER NOT NULL DEFAULT 0,
			duplicate_of TEXT,
			research_status TEXT NOT NULL DEFAULT 'pending',
			http_status INTEGER,
			page_title TEXT,
			page_description TEXT,
			favicon_url TEXT,
			researched_at INTEGER,
			cluster_id TEXT,
			bobbybookmarks_bookmark_id INTEGER,
			import_session_id INTEGER,
			raw_payload TEXT,
			synced_at INTEGER,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		);
		INSERT INTO links_backlog (
			uuid, url, normalized_url, title, description, tags, source, is_duplicate, duplicate_of,
			research_status, http_status, page_title, page_description, favicon_url, researched_at,
			cluster_id, bobbybookmarks_bookmark_id, import_session_id, raw_payload, synced_at, created_at, updated_at
		) VALUES
			('link-1', 'https://example.com/mcp', 'https://example.com/mcp', 'MCP', 'metadata', '["mcp"]', 'bobby', 0, NULL, 'pending', 200, 'MCP Page', 'desc', NULL, 1711958400, 'cluster-1', 1, 1, '{"score":1}', 1711958460, 1711958300, 1711958460),
			('link-2', 'https://example.com/dup', 'https://example.com/dup', 'Dup', 'dup', '[]', 'bobby', 1, 'link-1', 'pending', 200, NULL, NULL, NULL, NULL, 'cluster-1', NULL, NULL, NULL, 1711958460, 1711958200, 1711958450),
			('link-3', 'https://other.com/nope', 'https://other.com/nope', 'Other', 'other', '[]', 'manual', 0, NULL, 'done', 200, NULL, NULL, NULL, NULL, 'cluster-2', NULL, NULL, NULL, 1711958460, 1711958100, 1711958440);
	`); err != nil {
		t.Fatalf("failed to seed sqlite db: %v", err)
	}

	cfg := config.Default()
	cfg.WorkspaceRoot = workspace
	cfg.ConfigDir = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/links-backlog?limit=5&offset=0&search=mcp&source=bobby&research_status=pending&cluster_id=cluster-1&show_duplicates=true", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}

	for _, needle := range []string{
		`"fallback":"go-local-links-db"`,
		`"procedure":"linksBacklog.list"`,
		`using local tormentnexus links backlog list`,
		`"uuid":"link-1"`,
		`"title":"MCP"`,
		`"total":1`,
	} {
		if !strings.Contains(recorder.Body.String(), needle) {
			t.Fatalf("expected links backlog list fallback to contain %s, got %s", needle, recorder.Body.String())
		}
	}
	if strings.Contains(recorder.Body.String(), `"uuid":"link-2"`) || strings.Contains(recorder.Body.String(), `"uuid":"link-3"`) {
		t.Fatalf("expected filtered links backlog list to exclude non-matching rows, got %s", recorder.Body.String())
	}
}

func TestOAuthClientGetFallsBackToLocalDB(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	workspace := t.TempDir()
	dbPath := filepath.Join(workspace, "tormentnexus.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(`
		CREATE TABLE oauth_clients (
			client_id TEXT PRIMARY KEY,
			client_secret TEXT,
			client_name TEXT NOT NULL,
			redirect_uris TEXT NOT NULL DEFAULT '[]',
			grant_types TEXT NOT NULL DEFAULT '["authorization_code","refresh_token"]',
			response_types TEXT NOT NULL DEFAULT '["code"]',
			token_endpoint_auth_method TEXT NOT NULL DEFAULT 'none',
			scope TEXT DEFAULT 'admin',
			client_uri TEXT,
			logo_uri TEXT,
			contacts TEXT,
			tos_uri TEXT,
			policy_uri TEXT,
			software_id TEXT,
			software_version TEXT,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		);
		INSERT INTO oauth_clients (
			client_id, client_secret, client_name, redirect_uris, grant_types, response_types,
			token_endpoint_auth_method, scope, client_uri, logo_uri, contacts, tos_uri,
			policy_uri, software_id, software_version, created_at, updated_at
		) VALUES (
			'client-1', 'secret-1', 'Client One', '["https://example.com/callback"]',
			'["authorization_code","refresh_token"]', '["code"]', 'client_secret_post', 'admin',
			'https://example.com', 'https://example.com/logo.png', '["ops@example.com"]',
			'https://example.com/tos', 'https://example.com/policy', 'soft-1', '1.0.0', 1711958300, 1711958460
		);
	`); err != nil {
		t.Fatalf("failed to seed sqlite db: %v", err)
	}

	cfg := config.Default()
	cfg.WorkspaceRoot = workspace
	cfg.ConfigDir = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/oauth/clients/get?clientId=client-1", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}

	for _, needle := range []string{
		`"fallback":"go-local-oauth-clients-db"`,
		`"procedure":"oauth.clients.get"`,
		`using local tormentnexus oauth client record`,
		`"client_id":"client-1"`,
		`"client_name":"Client One"`,
		`"redirect_uris":["https://example.com/callback"]`,
		`"contacts":["ops@example.com"]`,
	} {
		if !strings.Contains(recorder.Body.String(), needle) {
			t.Fatalf("expected oauth client fallback to contain %s, got %s", needle, recorder.Body.String())
		}
	}
}

func TestOAuthSessionGetByServerFallsBackToLocalDB(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	workspace := t.TempDir()
	dbPath := filepath.Join(workspace, "tormentnexus.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(`
		CREATE TABLE oauth_sessions (
			uuid TEXT PRIMARY KEY,
			mcp_server_uuid TEXT NOT NULL,
			client_information TEXT NOT NULL DEFAULT '{}',
			tokens TEXT,
			code_verifier TEXT,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		);
		CREATE UNIQUE INDEX oauth_sessions_unique_per_server_idx ON oauth_sessions(mcp_server_uuid);
		INSERT INTO oauth_sessions (
			uuid, mcp_server_uuid, client_information, tokens, code_verifier, created_at, updated_at
		) VALUES (
			'session-1', 'server-1', '{"client_id":"client-1","client_name":"Client One"}',
			'{"access_token":"token-1","token_type":"Bearer"}', 'verifier-1', 1711958300, 1711958460
		);
	`); err != nil {
		t.Fatalf("failed to seed sqlite db: %v", err)
	}

	cfg := config.Default()
	cfg.WorkspaceRoot = workspace
	cfg.ConfigDir = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/oauth/sessions/by-server?mcpServerUuid=server-1", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}

	for _, needle := range []string{
		`"fallback":"go-local-oauth-sessions-db"`,
		`"procedure":"oauth.sessions.getByServer"`,
		`using local tormentnexus oauth session record`,
		`"uuid":"session-1"`,
		`"mcp_server_uuid":"server-1"`,
		`"client_information":{"client_id":"client-1","client_name":"Client One"}`,
		`"tokens":{"access_token":"token-1","token_type":"Bearer"}`,
	} {
		if !strings.Contains(recorder.Body.String(), needle) {
			t.Fatalf("expected oauth session fallback to contain %s, got %s", needle, recorder.Body.String())
		}
	}
}

func TestCatalogGetFallsBackToLocalDB(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	workspace := t.TempDir()
	dbPath := filepath.Join(workspace, "tormentnexus.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(`
		CREATE TABLE published_mcp_servers (
			uuid TEXT PRIMARY KEY,
			canonical_id TEXT NOT NULL UNIQUE,
			display_name TEXT NOT NULL,
			description TEXT,
			author TEXT,
			repository_url TEXT,
			homepage_url TEXT,
			icon_url TEXT,
			transport TEXT NOT NULL DEFAULT 'unknown',
			install_method TEXT NOT NULL DEFAULT 'unknown',
			auth_model TEXT NOT NULL DEFAULT 'unknown',
			status TEXT NOT NULL DEFAULT 'discovered',
			confidence INTEGER NOT NULL DEFAULT 0,
			tags TEXT NOT NULL DEFAULT '[]',
			categories TEXT NOT NULL DEFAULT '[]',
			stars INTEGER,
			last_seen_at INTEGER,
			last_verified_at INTEGER,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		);
		CREATE TABLE published_mcp_server_sources (
			uuid TEXT PRIMARY KEY,
			server_uuid TEXT NOT NULL,
			source_name TEXT NOT NULL,
			source_url TEXT,
			raw_payload TEXT,
			first_seen_at INTEGER NOT NULL,
			last_seen_at INTEGER NOT NULL
		);
		CREATE TABLE published_mcp_config_recipes (
			uuid TEXT PRIMARY KEY,
			server_uuid TEXT NOT NULL,
			recipe_version INTEGER NOT NULL DEFAULT 1,
			template TEXT NOT NULL,
			required_secrets TEXT NOT NULL DEFAULT '[]',
			required_env TEXT NOT NULL DEFAULT '{}',
			confidence INTEGER NOT NULL DEFAULT 0,
			explanation TEXT,
			is_active INTEGER NOT NULL DEFAULT 1,
			generated_by TEXT NOT NULL DEFAULT 'Configurator',
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		);
		CREATE TABLE published_mcp_validation_runs (
			uuid TEXT PRIMARY KEY,
			server_uuid TEXT NOT NULL,
			run_mode TEXT NOT NULL,
			started_at INTEGER NOT NULL,
			finished_at INTEGER,
			outcome TEXT NOT NULL DEFAULT 'pending',
			failure_class TEXT,
			tool_count INTEGER,
			findings_summary TEXT,
			performed_by TEXT NOT NULL DEFAULT 'Verifier',
			created_at INTEGER NOT NULL
		);
		INSERT INTO published_mcp_servers (
			uuid, canonical_id, display_name, description, author, repository_url, homepage_url, icon_url,
			transport, install_method, auth_model, status, confidence, tags, categories, stars,
			last_seen_at, last_verified_at, created_at, updated_at
		) VALUES (
			'catalog-1', 'registry/catalog-1', 'Catalog One', 'Great server', 'Hyper', 'https://github.com/example/repo',
			'https://example.com', 'https://example.com/icon.png', 'STDIO', 'npm', 'oauth', 'validated', 88,
			'["mcp","catalog"]', '["devtools"]', 123, 1711958400, 1711958460, 1711958300, 1711958520
		);
		INSERT INTO published_mcp_server_sources (
			uuid, server_uuid, source_name, source_url, raw_payload, first_seen_at, last_seen_at
		) VALUES (
			'source-1', 'catalog-1', 'smithery', 'https://smithery.ai/server', '{}', 1711958300, 1711958460
		);
		INSERT INTO published_mcp_config_recipes (
			uuid, server_uuid, recipe_version, template, required_secrets, required_env, confidence, explanation, is_active, generated_by, created_at, updated_at
		) VALUES (
			'recipe-1', 'catalog-1', 2, '{"command":"npx","args":["server"]}', '["API_KEY"]', '{"MODE":"prod"}', 90, 'validated recipe', 1, 'Configurator', 1711958300, 1711958460
		);
		INSERT INTO published_mcp_validation_runs (
			uuid, server_uuid, run_mode, started_at, finished_at, outcome, failure_class, tool_count, findings_summary, performed_by, created_at
		) VALUES (
			'run-1', 'catalog-1', 'tools_list', 1711958300, 1711958460, 'passed', NULL, 12, '{"summary":"ok"}', 'Verifier', 1711958460
		);
	`); err != nil {
		t.Fatalf("failed to seed sqlite db: %v", err)
	}

	cfg := config.Default()
	cfg.WorkspaceRoot = workspace
	cfg.ConfigDir = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/catalog/get?uuid=catalog-1", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}

	for _, needle := range []string{
		`"fallback":"go-local-published-catalog-db"`,
		`"procedure":"catalog.get"`,
		`using local tormentnexus published catalog records`,
		`"uuid":"catalog-1"`,
		`"display_name":"Catalog One"`,
		`"latestRun":{"created_at":`,
		`"uuid":"run-1"`,
		`"activeRecipe":{"confidence":90`,
		`"uuid":"recipe-1"`,
		`"sources":[{"first_seen_at":`,
		`"uuid":"source-1"`,
	} {
		if !strings.Contains(recorder.Body.String(), needle) {
			t.Fatalf("expected catalog get fallback to contain %s, got %s", needle, recorder.Body.String())
		}
	}
}

func TestSpecificFallbackGetRoutesReturnUnavailableWhenRecordMissing(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	workspace := t.TempDir()
	dbPath := filepath.Join(workspace, "tormentnexus.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(`
		CREATE TABLE api_keys (
			uuid TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			key TEXT NOT NULL,
			user_id TEXT,
			created_at INTEGER NOT NULL,
			is_active INTEGER NOT NULL DEFAULT 1
		);
		CREATE TABLE links_backlog (
			uuid TEXT PRIMARY KEY,
			url TEXT NOT NULL,
			normalized_url TEXT,
			title TEXT,
			description TEXT,
			tags TEXT NOT NULL DEFAULT '[]',
			source TEXT NOT NULL,
			is_duplicate INTEGER NOT NULL DEFAULT 0,
			duplicate_of TEXT,
			research_status TEXT NOT NULL DEFAULT 'pending',
			http_status INTEGER,
			page_title TEXT,
			page_description TEXT,
			favicon_url TEXT,
			researched_at INTEGER,
			cluster_id TEXT,
			bobbybookmarks_bookmark_id TEXT,
			import_session_id TEXT,
			raw_payload TEXT,
			synced_at INTEGER,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		);
		CREATE TABLE oauth_clients (
			client_id TEXT PRIMARY KEY,
			client_secret TEXT,
			client_name TEXT NOT NULL,
			redirect_uris TEXT NOT NULL DEFAULT '[]',
			grant_types TEXT NOT NULL DEFAULT '["authorization_code","refresh_token"]',
			response_types TEXT NOT NULL DEFAULT '["code"]',
			token_endpoint_auth_method TEXT NOT NULL DEFAULT 'none',
			scope TEXT DEFAULT 'admin',
			client_uri TEXT,
			logo_uri TEXT,
			contacts TEXT,
			tos_uri TEXT,
			policy_uri TEXT,
			software_id TEXT,
			software_version TEXT,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		);
		CREATE TABLE oauth_sessions (
			uuid TEXT PRIMARY KEY,
			mcp_server_uuid TEXT NOT NULL,
			client_information TEXT NOT NULL DEFAULT '{}',
			tokens TEXT,
			code_verifier TEXT,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		);
		CREATE UNIQUE INDEX oauth_sessions_unique_per_server_idx ON oauth_sessions(mcp_server_uuid);
		CREATE TABLE tool_chains (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			created_at INTEGER NOT NULL
		);
		CREATE TABLE tool_chain_steps (
			chain_id TEXT NOT NULL,
			step_order INTEGER NOT NULL,
			tool_name TEXT NOT NULL,
			arguments_template TEXT NOT NULL DEFAULT '{}'
		);
		CREATE TABLE published_mcp_servers (
			uuid TEXT PRIMARY KEY,
			canonical_id TEXT NOT NULL UNIQUE,
			display_name TEXT NOT NULL,
			description TEXT,
			author TEXT,
			repository_url TEXT,
			homepage_url TEXT,
			icon_url TEXT,
			transport TEXT NOT NULL DEFAULT 'unknown',
			install_method TEXT NOT NULL DEFAULT 'unknown',
			auth_model TEXT NOT NULL DEFAULT 'unknown',
			status TEXT NOT NULL DEFAULT 'discovered',
			confidence INTEGER NOT NULL DEFAULT 0,
			tags TEXT NOT NULL DEFAULT '[]',
			categories TEXT NOT NULL DEFAULT '[]',
			stars INTEGER,
			last_seen_at INTEGER,
			last_verified_at INTEGER,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		);
		CREATE TABLE published_mcp_server_sources (
			uuid TEXT PRIMARY KEY,
			server_uuid TEXT NOT NULL,
			source_name TEXT NOT NULL,
			source_url TEXT,
			raw_payload TEXT,
			first_seen_at INTEGER NOT NULL,
			last_seen_at INTEGER NOT NULL
		);
		CREATE TABLE published_mcp_config_recipes (
			uuid TEXT PRIMARY KEY,
			server_uuid TEXT NOT NULL,
			recipe_version INTEGER NOT NULL DEFAULT 1,
			template TEXT NOT NULL,
			required_secrets TEXT NOT NULL DEFAULT '[]',
			required_env TEXT NOT NULL DEFAULT '{}',
			confidence INTEGER NOT NULL DEFAULT 0,
			explanation TEXT,
			is_active INTEGER NOT NULL DEFAULT 1,
			generated_by TEXT NOT NULL DEFAULT 'Configurator',
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		);
		CREATE TABLE published_mcp_validation_runs (
			uuid TEXT PRIMARY KEY,
			server_uuid TEXT NOT NULL,
			run_mode TEXT NOT NULL,
			started_at INTEGER NOT NULL,
			finished_at INTEGER,
			outcome TEXT NOT NULL DEFAULT 'pending',
			failure_class TEXT,
			tool_count INTEGER,
			findings_summary TEXT,
			performed_by TEXT NOT NULL DEFAULT 'Verifier',
			created_at INTEGER NOT NULL
		);
	`); err != nil {
		t.Fatalf("failed to create fallback schema: %v", err)
	}

	cfg := config.Default()
	cfg.WorkspaceRoot = workspace
	cfg.ConfigDir = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	cases := []struct {
		name        string
		path        string
		errorNeedle string
		procedure   string
	}{
		{name: "api key missing", path: "/api/api-keys/get?uuid=missing-key", errorNeedle: `"error":"API key unavailable"`, procedure: `"procedure":"apiKeys.get"`},
		{name: "links backlog item missing", path: "/api/links-backlog/get?uuid=missing-link", errorNeedle: `"error":"links backlog item unavailable"`, procedure: `"procedure":"linksBacklog.get"`},
		{name: "oauth client missing", path: "/api/oauth/clients/get?clientId=missing-client", errorNeedle: `"error":"OAuth client unavailable"`, procedure: `"procedure":"oauth.clients.get"`},
		{name: "oauth session missing", path: "/api/oauth/sessions/by-server?mcpServerUuid=missing-server", errorNeedle: `"error":"OAuth session unavailable"`, procedure: `"procedure":"oauth.sessions.getByServer"`},
		{name: "catalog entry missing", path: "/api/catalog/get?uuid=missing-catalog", errorNeedle: `"error":"catalog entry unavailable"`, procedure: `"procedure":"catalog.get"`},
		{name: "tool chain missing", path: "/api/tool-chains/get?id=missing-chain", errorNeedle: `"error":"tool chain unavailable"`, procedure: `"procedure":"toolChaining.getChain"`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, tc.path, nil))
			if recorder.Code != http.StatusServiceUnavailable {
				t.Fatalf("expected 503, got %d with body %s", recorder.Code, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), `"success":false`) {
				t.Fatalf("expected unsuccessful response, got %s", recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.errorNeedle) {
				t.Fatalf("expected error needle %s, got %s", tc.errorNeedle, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.procedure) {
				t.Fatalf("expected procedure metadata %s, got %s", tc.procedure, recorder.Body.String())
			}
		})
	}
}

func TestCatalogRunsFallsBackToLocalDB(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	workspace := t.TempDir()
	dbPath := filepath.Join(workspace, "tormentnexus.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(`
		CREATE TABLE published_mcp_validation_runs (
			uuid TEXT PRIMARY KEY,
			server_uuid TEXT NOT NULL,
			run_mode TEXT NOT NULL,
			started_at INTEGER NOT NULL,
			finished_at INTEGER,
			outcome TEXT NOT NULL DEFAULT 'pending',
			failure_class TEXT,
			tool_count INTEGER,
			findings_summary TEXT,
			performed_by TEXT NOT NULL DEFAULT 'Verifier',
			created_at INTEGER NOT NULL
		);
		INSERT INTO published_mcp_validation_runs (
			uuid, server_uuid, run_mode, started_at, finished_at, outcome, failure_class, tool_count, findings_summary, performed_by, created_at
		) VALUES
			('run-2', 'catalog-1', 'full_validation', 1711958500, 1711958600, 'failed', 'network', 3, '{"summary":"bad"}', 'Verifier', 1711958600),
			('run-1', 'catalog-1', 'tools_list', 1711958300, 1711958460, 'passed', NULL, 12, '{"summary":"ok"}', 'Verifier', 1711958460),
			('run-0', 'catalog-2', 'tools_list', 1711958200, 1711958250, 'passed', NULL, 2, '{"summary":"other"}', 'Verifier', 1711958250);
	`); err != nil {
		t.Fatalf("failed to seed sqlite db: %v", err)
	}

	cfg := config.Default()
	cfg.WorkspaceRoot = workspace
	cfg.ConfigDir = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/catalog/runs?server_uuid=catalog-1&limit=1", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}

	for _, needle := range []string{
		`"fallback":"go-local-published-catalog-db"`,
		`"procedure":"catalog.listRuns"`,
		`using local tormentnexus published catalog validation runs`,
		`"uuid":"run-2"`,
		`"outcome":"failed"`,
		`"failure_class":"network"`,
	} {
		if !strings.Contains(recorder.Body.String(), needle) {
			t.Fatalf("expected catalog runs fallback to contain %s, got %s", needle, recorder.Body.String())
		}
	}
	if strings.Contains(recorder.Body.String(), `"uuid":"run-1"`) {
		t.Fatalf("expected limit=1 to exclude older run, got %s", recorder.Body.String())
	}
}

func TestCatalogStatsFallsBackToLocalDB(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	workspace := t.TempDir()
	dbPath := filepath.Join(workspace, "tormentnexus.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	defer db.Close()

	now := time.Now().UTC().Unix()
	old := now - (48 * 60 * 60)

	if _, err := db.Exec(fmt.Sprintf(`
		CREATE TABLE published_mcp_servers (
			uuid TEXT PRIMARY KEY,
			canonical_id TEXT NOT NULL UNIQUE,
			display_name TEXT NOT NULL,
			description TEXT,
			author TEXT,
			repository_url TEXT,
			homepage_url TEXT,
			icon_url TEXT,
			transport TEXT NOT NULL DEFAULT 'unknown',
			install_method TEXT NOT NULL DEFAULT 'unknown',
			auth_model TEXT NOT NULL DEFAULT 'unknown',
			status TEXT NOT NULL DEFAULT 'discovered',
			confidence INTEGER NOT NULL DEFAULT 0,
			tags TEXT NOT NULL DEFAULT '[]',
			categories TEXT NOT NULL DEFAULT '[]',
			stars INTEGER,
			last_seen_at INTEGER,
			last_verified_at INTEGER,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		);
		INSERT INTO published_mcp_servers (
			uuid, canonical_id, display_name, status, created_at, updated_at
		) VALUES
			('catalog-1', 'registry/1', 'One', 'validated', %d, %d),
			('catalog-2', 'registry/2', 'Two', 'broken', %d, %d),
			('catalog-3', 'registry/3', 'Three', 'validated', %d, %d),
			('catalog-4', 'registry/4', 'Four', 'archived', %d, %d),
			('catalog-5', 'registry/5', 'Five', 'discovered', %d, %d);
	`, old, now, old, old, old, now, old, old, old, now)); err != nil {
		t.Fatalf("failed to seed sqlite db: %v", err)
	}

	cfg := config.Default()
	cfg.WorkspaceRoot = workspace
	cfg.ConfigDir = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/catalog/stats", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}

	for _, needle := range []string{
		`"fallback":"go-local-published-catalog-db"`,
		`"procedure":"catalog.stats"`,
		`using local tormentnexus published catalog aggregates`,
		`"total":5`,
		`"validated":2`,
		`"broken":1`,
		`"recentlyUpdated":3`,
		`"status":"archived"`,
		`"count":1`,
		`"status":"validated"`,
		`"count":2`,
	} {
		if !strings.Contains(recorder.Body.String(), needle) {
			t.Fatalf("expected catalog stats fallback to contain %s, got %s", needle, recorder.Body.String())
		}
	}
}

func TestCatalogLinkedServersFallsBackToLocalDB(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	workspace := t.TempDir()
	dbPath := filepath.Join(workspace, "tormentnexus.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(`
		CREATE TABLE mcp_servers (
			uuid TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			type TEXT NOT NULL DEFAULT 'STDIO',
			command TEXT,
			args TEXT NOT NULL DEFAULT '[]',
			env TEXT NOT NULL DEFAULT '{}',
			url TEXT,
			error_status TEXT NOT NULL DEFAULT 'NONE',
			created_at INTEGER NOT NULL,
			bearer_token TEXT,
			headers TEXT NOT NULL DEFAULT '{}',
			always_on INTEGER NOT NULL DEFAULT 0,
			user_id TEXT NOT NULL,
			source_published_server_uuid TEXT
		);
		INSERT INTO mcp_servers (
			uuid, name, description, type, command, args, env, url, error_status, created_at, bearer_token, headers, always_on, user_id, source_published_server_uuid
		) VALUES
			('managed-1', 'linked-one', 'Linked server', 'STDIO', 'npx', '["server"]', '{"TOKEN":"x"}', NULL, 'NONE', 1711958460, NULL, '{}', 1, 'system', 'catalog-1'),
			('managed-2', 'other-one', 'Other server', 'STDIO', 'npx', '["other"]', '{}', NULL, 'NONE', 1711958460, NULL, '{}', 0, 'system', 'catalog-2');
	`); err != nil {
		t.Fatalf("failed to seed sqlite db: %v", err)
	}

	cfg := config.Default()
	cfg.WorkspaceRoot = workspace
	cfg.ConfigDir = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/catalog/linked-servers?published_server_uuid=catalog-1", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}

	for _, needle := range []string{
		`"fallback":"go-local-published-catalog-db"`,
		`"procedure":"catalog.listLinkedServers"`,
		`using local tormentnexus linked managed servers`,
		`"uuid":"managed-1"`,
		`"name":"linked-one"`,
		`"source_published_server_uuid":"catalog-1"`,
	} {
		if !strings.Contains(recorder.Body.String(), needle) {
			t.Fatalf("expected catalog linked servers fallback to contain %s, got %s", needle, recorder.Body.String())
		}
	}
	if strings.Contains(recorder.Body.String(), `"uuid":"managed-2"`) {
		t.Fatalf("expected linked servers fallback to exclude unrelated server, got %s", recorder.Body.String())
	}
}

func TestCatalogListFallsBackToLocalDB(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	workspace := t.TempDir()
	dbPath := filepath.Join(workspace, "tormentnexus.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(`
		CREATE TABLE published_mcp_servers (
			uuid TEXT PRIMARY KEY,
			canonical_id TEXT NOT NULL UNIQUE,
			display_name TEXT NOT NULL,
			description TEXT,
			author TEXT,
			repository_url TEXT,
			homepage_url TEXT,
			icon_url TEXT,
			transport TEXT NOT NULL DEFAULT 'unknown',
			install_method TEXT NOT NULL DEFAULT 'unknown',
			auth_model TEXT NOT NULL DEFAULT 'unknown',
			status TEXT NOT NULL DEFAULT 'discovered',
			confidence INTEGER NOT NULL DEFAULT 0,
			tags TEXT NOT NULL DEFAULT '[]',
			categories TEXT NOT NULL DEFAULT '[]',
			stars INTEGER,
			last_seen_at INTEGER,
			last_verified_at INTEGER,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		);
		INSERT INTO published_mcp_servers (
			uuid, canonical_id, display_name, description, author, transport, install_method, auth_model, status, confidence, tags, categories, created_at, updated_at
		) VALUES
			('catalog-1', 'registry/catalog-1', 'MCP Filesystem', 'MCP file tools', 'Hyper', 'STDIO', 'npm', 'none', 'validated', 90, '["mcp"]', '["devtools"]', 1711958300, 1711958600),
			('catalog-2', 'registry/catalog-2', 'Other Server', 'Different', 'Someone', 'SSE', 'docker', 'oauth', 'validated', 10, '[]', '[]', 1711958300, 1711958500),
			('catalog-3', 'registry/catalog-3', 'MCP Browser', 'MCP browser tools', 'Hyper', 'STDIO', 'npm', 'oauth', 'broken', 50, '["mcp"]', '["browser"]', 1711958300, 1711958550);
	`); err != nil {
		t.Fatalf("failed to seed sqlite db: %v", err)
	}

	cfg := config.Default()
	cfg.WorkspaceRoot = workspace
	cfg.ConfigDir = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/catalog?limit=10&offset=0&search=mcp&status=validated&transport=STDIO&install_method=npm", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}

	for _, needle := range []string{
		`"fallback":"go-local-published-catalog-db"`,
		`"procedure":"catalog.list"`,
		`using local tormentnexus published catalog list`,
		`"servers":[{"auth_model":"none"`,
		`"uuid":"catalog-1"`,
		`"display_name":"MCP Filesystem"`,
		`"total":1`,
	} {
		if !strings.Contains(recorder.Body.String(), needle) {
			t.Fatalf("expected catalog list fallback to contain %s, got %s", needle, recorder.Body.String())
		}
	}
	if strings.Contains(recorder.Body.String(), `"uuid":"catalog-2"`) || strings.Contains(recorder.Body.String(), `"uuid":"catalog-3"`) {
		t.Fatalf("expected filtered catalog list fallback to exclude non-matching rows, got %s", recorder.Body.String())
	}
}

func TestBrowserControlsReadRoutesFallBackToLocalDB(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	workspace := t.TempDir()
	dbPath := filepath.Join(workspace, "tormentnexus.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(`
		CREATE TABLE browser_history (
			id TEXT PRIMARY KEY,
			url TEXT NOT NULL,
			title TEXT NOT NULL,
			domain TEXT NOT NULL,
			visited_at INTEGER NOT NULL,
			visit_count INTEGER NOT NULL DEFAULT 1
		);
		CREATE TABLE browser_console_logs (
			id TEXT PRIMARY KEY,
			level TEXT NOT NULL,
			message TEXT NOT NULL,
			source TEXT NOT NULL,
			url TEXT,
			line_number INTEGER,
			timestamp INTEGER NOT NULL
		);
		INSERT INTO browser_history (id, url, title, domain, visited_at, visit_count) VALUES
			('hist-1', 'https://example.com/a', 'Example A', 'example.com', 200, 3),
			('hist-2', 'https://other.com/b', 'Other B', 'other.com', 100, 1);
		INSERT INTO browser_console_logs (id, level, message, source, url, line_number, timestamp) VALUES
			('log-1', 'error', 'err one', 'console', 'https://example.com/a', 12, 300),
			('log-2', 'warn', 'warn two', 'console', NULL, NULL, 200);
	`); err != nil {
		t.Fatalf("failed to seed sqlite db: %v", err)
	}

	cfg := config.Default()
	cfg.WorkspaceRoot = workspace
	cfg.ConfigDir = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	historyRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(historyRecorder, httptest.NewRequest(http.MethodGet, "/api/browser-controls/history/query?query=example&limit=10&since=150&domain=example.com", nil))
	if historyRecorder.Code != http.StatusOK {
		t.Fatalf("expected history status 200, got %d with body %s", historyRecorder.Code, historyRecorder.Body.String())
	}
	for _, needle := range []string{
		`"fallback":"go-local-browser-data-db"`,
		`"procedure":"browserControls.queryHistory"`,
		`"title":"Example A"`,
		`"total":1`,
	} {
		if !strings.Contains(historyRecorder.Body.String(), needle) {
			t.Fatalf("expected history fallback to contain %s, got %s", needle, historyRecorder.Body.String())
		}
	}

	logsRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(logsRecorder, httptest.NewRequest(http.MethodGet, "/api/browser-controls/logs/query?level=error&search=err&limit=10", nil))
	if logsRecorder.Code != http.StatusOK {
		t.Fatalf("expected logs status 200, got %d with body %s", logsRecorder.Code, logsRecorder.Body.String())
	}
	for _, needle := range []string{
		`"procedure":"browserControls.queryConsoleLogs"`,
		`"message":"err one"`,
		`"total":1`,
	} {
		if !strings.Contains(logsRecorder.Body.String(), needle) {
			t.Fatalf("expected logs fallback to contain %s, got %s", needle, logsRecorder.Body.String())
		}
	}

	statsRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(statsRecorder, httptest.NewRequest(http.MethodGet, "/api/browser-controls/stats", nil))
	if statsRecorder.Code != http.StatusOK {
		t.Fatalf("expected stats status 200, got %d with body %s", statsRecorder.Code, statsRecorder.Body.String())
	}
	for _, needle := range []string{
		`"procedure":"browserControls.stats"`,
		`"historyCount":2`,
		`"uniqueDomains":2`,
		`"consoleLogCount":2`,
		`"consoleErrors":1`,
	} {
		if !strings.Contains(statsRecorder.Body.String(), needle) {
			t.Fatalf("expected stats fallback to contain %s, got %s", needle, statsRecorder.Body.String())
		}
	}
}

func TestConfigReadRoutesFallBackToLocalDB(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	workspace := t.TempDir()
	dbPath := filepath.Join(workspace, "tormentnexus.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(`
		CREATE TABLE config (
			id TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			description TEXT,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		);
		INSERT INTO config (id, value, created_at, updated_at) VALUES
			('theme', 'dark', 1711958300, 1711958460),
			('MCP_TIMEOUT', '15000', 1711958300, 1711958460),
			('MCP_MAX_ATTEMPTS', '4', 1711958300, 1711958460),
			('MCP_MAX_TOTAL_TIMEOUT', '30000', 1711958300, 1711958460),
			('MCP_RESET_TIMEOUT_ON_PROGRESS', 'false', 1711958300, 1711958460),
			('SESSION_LIFETIME', '3600000', 1711958300, 1711958460),
			('DISABLE_SSO_SIGNUP', 'true', 1711958300, 1711958460);
	`); err != nil {
		t.Fatalf("failed to seed sqlite db: %v", err)
	}

	cfg := config.Default()
	cfg.WorkspaceRoot = workspace
	cfg.ConfigDir = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	cases := []struct {
		name     string
		path     string
		contains []string
	}{
		{"config list", "/api/config/list", []string{`"procedure":"config.list"`, `"key":"theme"`, `"value":"dark"`}},
		{"config get", "/api/config/get?key=theme", []string{`"procedure":"config.get"`, `"data":"dark"`}},
		{"mcp timeout", "/api/config/mcp-timeout", []string{`"procedure":"config.getMcpTimeout"`, `"data":15000`}},
		{"mcp max attempts", "/api/config/mcp-max-attempts", []string{`"procedure":"config.getMcpMaxAttempts"`, `"data":4`}},
		{"mcp max total timeout", "/api/config/mcp-max-total-timeout", []string{`"procedure":"config.getMcpMaxTotalTimeout"`, `"data":30000`}},
		{"mcp reset on progress", "/api/config/mcp-reset-timeout-on-progress", []string{`"procedure":"config.getMcpResetTimeoutOnProgress"`, `"data":true`}},
		{"session lifetime", "/api/config/session-lifetime", []string{`"procedure":"config.getSessionLifetime"`, `"data":3600000`}},
		{"signup disabled default", "/api/config/signup-disabled", []string{`"procedure":"config.getSignupDisabled"`, `"data":false`}},
		{"sso signup disabled", "/api/config/sso-signup-disabled", []string{`"procedure":"config.getSsoSignupDisabled"`, `"data":true`}},
		{"basic auth disabled default", "/api/config/basic-auth-disabled", []string{`"procedure":"config.getBasicAuthDisabled"`, `"data":false`}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, tc.path, nil))
			if recorder.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d with body %s", recorder.Code, recorder.Body.String())
			}
			for _, needle := range tc.contains {
				if !strings.Contains(recorder.Body.String(), needle) {
					t.Fatalf("expected config fallback to contain %s, got %s", needle, recorder.Body.String())
				}
			}
		})
	}
}

func TestOperatorFallbackGetRoutesReturnUnavailableWhenRecordMissing(t *testing.T) {
	workspaceRoot := t.TempDir()
	mainConfigDir := t.TempDir()
	homeDir := t.TempDir()

	if err := os.MkdirAll(filepath.Join(workspaceRoot, ".tormentnexus"), 0o755); err != nil {
		t.Fatalf("failed to create .tormentnexus dir: %v", err)
	}
	scriptsPayload := []map[string]any{
		{
			"uuid":        "script-local-1",
			"name":        "Deploy local",
			"description": "Local script fallback",
			"code":        "echo local",
		},
	}
	scriptsJSON, err := json.Marshal(scriptsPayload)
	if err != nil {
		t.Fatalf("failed to marshal scripts payload: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workspaceRoot, ".tormentnexus", "saved-scripts.json"), scriptsJSON, 0o644); err != nil {
		t.Fatalf("failed to seed saved scripts: %v", err)
	}

	dbPath := filepath.Join(workspaceRoot, "tormentnexus.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(`
		CREATE TABLE tool_sets (
			uuid TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			tools_json TEXT NOT NULL DEFAULT '[]',
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		);
		CREATE TABLE tool_set_items (
			uuid TEXT PRIMARY KEY,
			tool_set_uuid TEXT NOT NULL,
			tool_uuid TEXT NOT NULL,
			created_at INTEGER NOT NULL
		);
		INSERT INTO tool_sets (uuid, name, description, tools_json, created_at, updated_at)
		VALUES ('toolset-local-1', 'Core local tools', 'Fallback tool set', '["search_tools"]', 1711958300, 1711958460);
		CREATE TABLE mcp_servers (
			uuid TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			type TEXT NOT NULL DEFAULT 'STDIO',
			command TEXT,
			args TEXT NOT NULL DEFAULT '[]',
			env TEXT NOT NULL DEFAULT '{}',
			url TEXT,
			error_status TEXT NOT NULL DEFAULT 'NONE',
			created_at INTEGER NOT NULL,
			bearer_token TEXT,
			headers TEXT NOT NULL DEFAULT '{}',
			always_on INTEGER NOT NULL DEFAULT 0,
			user_id TEXT NOT NULL,
			source_published_server_uuid TEXT
		);
	`); err != nil {
		t.Fatalf("failed to seed sqlite fallback db: %v", err)
	}

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")
	t.Setenv("USERPROFILE", homeDir)
	t.Setenv("APPDATA", filepath.Join(homeDir, "AppData", "Roaming"))
	t.Setenv("LOCALAPPDATA", filepath.Join(homeDir, "AppData", "Local"))

	cfg := config.Default()
	cfg.WorkspaceRoot = workspaceRoot
	cfg.ConfigDir = homeDir
	cfg.MainConfigDir = mainConfigDir
	server := New(cfg, stubDetector{})

	cases := []struct {
		name        string
		path        string
		errorNeedle string
		procedure   string
	}{
		{name: "configured server get missing", path: "/api/mcp/servers/get?uuid=missing-server", errorNeedle: `"error":"configured MCP server unavailable"`, procedure: `"procedure":"mcpServers.get"`},
		{name: "tool set get missing", path: "/api/tool-sets/get?uuid=missing-toolset", errorNeedle: `"error":"tool set unavailable"`, procedure: `"procedure":"toolSets.get"`},
		{name: "saved script get missing", path: "/api/scripts/get?uuid=missing-script", errorNeedle: `"error":"saved script unavailable"`, procedure: `"procedure":"savedScripts.get"`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, tc.path, nil))
			if recorder.Code != http.StatusServiceUnavailable {
				t.Fatalf("expected 503, got %d with body %s", recorder.Code, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), `"success":false`) {
				t.Fatalf("expected unsuccessful response, got %s", recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.errorNeedle) {
				t.Fatalf("expected error needle %s, got %s", tc.errorNeedle, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.procedure) {
				t.Fatalf("expected procedure metadata %s, got %s", tc.procedure, recorder.Body.String())
			}
		})
	}
}

func TestUnifiedDirectoryRoutesFallBackToLocalDB(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	workspace := t.TempDir()
	dbPath := filepath.Join(workspace, "tormentnexus.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	defer db.Close()

	now := time.Now().UTC().Unix()
	if _, err := db.Exec(fmt.Sprintf(`
		CREATE TABLE published_mcp_servers (
			uuid TEXT PRIMARY KEY,
			canonical_id TEXT NOT NULL UNIQUE,
			display_name TEXT NOT NULL,
			description TEXT,
			author TEXT,
			repository_url TEXT,
			homepage_url TEXT,
			icon_url TEXT,
			transport TEXT NOT NULL DEFAULT 'unknown',
			install_method TEXT NOT NULL DEFAULT 'unknown',
			auth_model TEXT NOT NULL DEFAULT 'unknown',
			status TEXT NOT NULL DEFAULT 'discovered',
			confidence INTEGER NOT NULL DEFAULT 0,
			tags TEXT NOT NULL DEFAULT '[]',
			categories TEXT NOT NULL DEFAULT '[]',
			stars INTEGER,
			last_seen_at INTEGER,
			last_verified_at INTEGER,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		);
		CREATE TABLE links_backlog (
			uuid TEXT PRIMARY KEY,
			url TEXT NOT NULL,
			normalized_url TEXT NOT NULL UNIQUE,
			title TEXT,
			description TEXT,
			tags TEXT NOT NULL DEFAULT '[]',
			source TEXT NOT NULL DEFAULT 'manual',
			is_duplicate INTEGER NOT NULL DEFAULT 0,
			duplicate_of TEXT,
			research_status TEXT NOT NULL DEFAULT 'pending',
			http_status INTEGER,
			page_title TEXT,
			page_description TEXT,
			favicon_url TEXT,
			researched_at INTEGER,
			cluster_id TEXT,
			bobbybookmarks_bookmark_id INTEGER,
			import_session_id INTEGER,
			raw_payload TEXT,
			synced_at INTEGER,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		);
		INSERT INTO published_mcp_servers (
			uuid, canonical_id, display_name, description, transport, install_method, auth_model, status, confidence, tags, categories, repository_url, created_at, updated_at
		) VALUES
			('dir-1', 'registry/dir-1', 'MCP Catalog', 'catalog entry', 'STDIO', 'npm', 'none', 'validated', 99, '["mcp"]', '["tools"]', 'https://example.com/repo', %d, %d),
			('dir-2', 'registry/dir-2', 'Other Catalog', 'other entry', 'SSE', 'docker', 'oauth', 'broken', 10, '[]', '[]', NULL, %d, %d);
		INSERT INTO links_backlog (
			uuid, url, normalized_url, title, description, tags, source, is_duplicate, duplicate_of, research_status, created_at, updated_at
		) VALUES
			('backlog-1', 'https://example.com/mcp', 'https://example.com/mcp', 'MCP Backlog', 'backlog entry', '["mcp"]', 'bobby', 0, NULL, 'pending', %d, %d),
			('backlog-2', 'https://example.com/dup', 'https://example.com/dup', 'Dup Backlog', 'dup', '[]', 'bobby', 1, 'backlog-1', 'pending', %d, %d);
	`, now-10, now-5, now-30, now-20, now-15, now-1, now-25, now-2)); err != nil {
		t.Fatalf("failed to seed sqlite db: %v", err)
	}

	cfg := config.Default()
	cfg.WorkspaceRoot = workspace
	cfg.ConfigDir = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	listRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(listRecorder, httptest.NewRequest(http.MethodGet, "/api/directory?limit=10&offset=0&search=mcp&source=all&show_duplicates=true&duplicates_only=false&research_status=pending", nil))
	if listRecorder.Code != http.StatusOK {
		t.Fatalf("expected directory list 200, got %d with body %s", listRecorder.Code, listRecorder.Body.String())
	}
	for _, needle := range []string{
		`"procedure":"unifiedDirectory.list"`,
		`"fallback":"go-local-unified-directory"`,
		`"id":"dir-1"`,
		`"id":"backlog-1"`,
		`"total":2`,
		`"catalog":1`,
		`"backlog":1`,
	} {
		if !strings.Contains(listRecorder.Body.String(), needle) {
			t.Fatalf("expected unified directory list fallback to contain %s, got %s", needle, listRecorder.Body.String())
		}
	}

	statsRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(statsRecorder, httptest.NewRequest(http.MethodGet, "/api/directory/stats", nil))
	if statsRecorder.Code != http.StatusOK {
		t.Fatalf("expected directory stats 200, got %d with body %s", statsRecorder.Code, statsRecorder.Body.String())
	}
	for _, needle := range []string{
		`"procedure":"unifiedDirectory.stats"`,
		`"combined_total":4`,
		`"validated":1`,
		`"broken":1`,
		`"updated_24h":2`,
		`"backlog":{"duplicates":1`,
	} {
		if !strings.Contains(statsRecorder.Body.String(), needle) {
			t.Fatalf("expected unified directory stats fallback to contain %s, got %s", needle, statsRecorder.Body.String())
		}
	}
}

func TestWorkflowCanvasRoutesFallBackToLocalDB(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	workspace := t.TempDir()
	dbPath := filepath.Join(workspace, "tormentnexus.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(`
		CREATE TABLE workflows (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			nodes_json TEXT NOT NULL DEFAULT '[]',
			edges_json TEXT NOT NULL DEFAULT '[]',
			user_id TEXT NOT NULL,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		);
		INSERT INTO workflows (id, name, description, nodes_json, edges_json, user_id, created_at, updated_at) VALUES
			('canvas-1', 'Canvas One', 'primary canvas', '[{"id":"n1"}]', '[{"id":"e1"}]', 'system', 1711958300, 1711958460),
			('canvas-2', 'Canvas Two', 'secondary canvas', '[]', '[]', 'system', 1711958200, 1711958400);
	`); err != nil {
		t.Fatalf("failed to seed sqlite db: %v", err)
	}

	cfg := config.Default()
	cfg.WorkspaceRoot = workspace
	cfg.ConfigDir = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	listRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(listRecorder, httptest.NewRequest(http.MethodGet, "/api/workflows/canvases", nil))
	if listRecorder.Code != http.StatusOK {
		t.Fatalf("expected canvases 200, got %d with body %s", listRecorder.Code, listRecorder.Body.String())
	}
	for _, needle := range []string{
		`"procedure":"workflow.listCanvases"`,
		`"fallback":"go-local-workflows-db"`,
		`"id":"canvas-1"`,
		`"name":"Canvas One"`,
	} {
		if !strings.Contains(listRecorder.Body.String(), needle) {
			t.Fatalf("expected workflow canvases fallback to contain %s, got %s", needle, listRecorder.Body.String())
		}
	}

	loadRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(loadRecorder, httptest.NewRequest(http.MethodGet, "/api/workflows/canvas?id=canvas-1", nil))
	if loadRecorder.Code != http.StatusOK {
		t.Fatalf("expected canvas load 200, got %d with body %s", loadRecorder.Code, loadRecorder.Body.String())
	}
	for _, needle := range []string{
		`"procedure":"workflow.loadCanvas"`,
		`"id":"canvas-1"`,
		`"nodes_json":[{"id":"n1"}]`,
		`"edges_json":[{"id":"e1"}]`,
	} {
		if !strings.Contains(loadRecorder.Body.String(), needle) {
			t.Fatalf("expected workflow canvas fallback to contain %s, got %s", needle, loadRecorder.Body.String())
		}
	}
}

func TestDirectorConfigGetFallsBackToLocalConfig(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	workspaceRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(workspaceRoot, ".tormentnexus"), 0o755); err != nil {
		t.Fatalf("failed to create .tormentnexus dir: %v", err)
	}
	if err := os.WriteFile(
		filepath.Join(workspaceRoot, ".tormentnexus", "config.json"),
		[]byte(`{"persona":"professional","defaultTopic":"mcp","enableCouncil":true}`),
		0o644,
	); err != nil {
		t.Fatalf("failed to write config.json: %v", err)
	}

	cfg := config.Default()
	cfg.WorkspaceRoot = workspaceRoot
	cfg.ConfigDir = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/director-config", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected director-config 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}

	for _, needle := range []string{
		`"procedure":"directorConfig.get"`,
		`"fallback":"go-local-tormentnexus-config"`,
		`"persona":"professional"`,
		`"defaultTopic":"mcp"`,
		`"enableCouncil":true`,
	} {
		if !strings.Contains(recorder.Body.String(), needle) {
			t.Fatalf("expected director config fallback to contain %s, got %s", needle, recorder.Body.String())
		}
	}
}

func TestInfrastructureStatusFallsBackToLocalProbe(t *testing.T) {
	workspaceRoot := t.TempDir()
	userProfile := t.TempDir()
	infraBinary := "mcpetes"
	infraSubmoduleDir := "mcpetes"

	binDir := filepath.Join(workspaceRoot, "..", "..", "submodules", infraSubmoduleDir, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("failed to create infra bin dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(binDir, infraBinary), []byte("binary"), 0o644); err != nil {
		t.Fatalf("failed to write infra binary: %v", err)
	}

	configDir := filepath.Join(userProfile, ".config", "mcpetes")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("failed to create infra config dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte("daemon: false"), 0o644); err != nil {
		t.Fatalf("failed to write infra config: %v", err)
	}

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")
	t.Setenv("TORMENTNEXUS_INFRA_BINARY", infraBinary)
	t.Setenv("TORMENTNEXUS_INFRA_SUBMODULE", infraSubmoduleDir)
	t.Setenv("USERPROFILE", userProfile)

	cfg := config.Default()
	cfg.WorkspaceRoot = workspaceRoot
	server := New(cfg, stubDetector{})

	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/infrastructure", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}

	for _, needle := range []string{
		`"fallback":"go-local-infrastructure"`,
		`"procedure":"infrastructure.getInfrastructureStatus"`,
		`using local infrastructure binary/config visibility`,
		`"installed":true`,
		`"hasConfig":true`,
	} {
		if !strings.Contains(recorder.Body.String(), needle) {
			t.Fatalf("expected response to contain %s, got %s", needle, recorder.Body.String())
		}
	}
}

func TestPoliciesListFallsBackToLocalDB(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	workspace := t.TempDir()
	dbPath := filepath.Join(workspace, "tormentnexus.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(`
		CREATE TABLE policies (
			uuid TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			rules TEXT NOT NULL DEFAULT '{}',
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		);
		INSERT INTO policies (uuid, name, description, rules, created_at, updated_at)
		VALUES
			('policy-1', 'Default', 'baseline policy', '{"allow":["tool.read"]}', 1711958400, 1711958460),
			('policy-2', 'Admin', 'admin policy', '{"allow":["tool.write"],"deny":["tool.delete"]}', 1711958500, 1711958560);
	`); err != nil {
		t.Fatalf("failed to seed sqlite db: %v", err)
	}

	cfg := config.Default()
	cfg.WorkspaceRoot = workspace
	cfg.ConfigDir = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/policies", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}

	for _, needle := range []string{
		`"fallback":"go-local-policy-db"`,
		`"procedure":"policies.list"`,
		`using local tormentnexus policy records`,
		`"policy-1"`,
		`"policy-2"`,
		`"allow":["tool.write"]`,
	} {
		if !strings.Contains(recorder.Body.String(), needle) {
			t.Fatalf("expected response to contain %s, got %s", needle, recorder.Body.String())
		}
	}
}

func TestPulseAndBrowserStatsFallBackToLocalPreview(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	server := New(config.Default(), stubDetector{})

	cases := []struct {
		name     string
		path     string
		contains []string
	}{
		{
			name: "pulse status",
			path: "/api/pulse/status",
			contains: []string{
				`"fallback":"go-local-pulse"`,
				`"procedure":"pulse.getSystemStatus"`,
				`using local offline pulse status`,
				`"status":"offline"`,
			},
		},
		{
			name: "browser stats",
			path: "/api/browser-extension/stats",
			contains: []string{
				`"fallback":"go-local-browser-memory"`,
				`"procedure":"browserExtension.stats"`,
				`using local browser memory stats from tormentnexus.db`,
				`"totalMemories":0`,
				`"uniqueUrls":0`,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, tc.path, nil))

			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
			}
			for _, needle := range tc.contains {
				if !strings.Contains(recorder.Body.String(), needle) {
					t.Fatalf("expected response to contain %s, got %s", needle, recorder.Body.String())
				}
			}
		})
	}
}

func TestPulseEventsAndBrowserMemoriesFallBackToEmptyState(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	workspaceRoot := t.TempDir()
	dbPath := filepath.Join(workspaceRoot, "tormentnexus.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	defer db.Close()
	if _, err := db.Exec(`
		CREATE TABLE web_memories (
			id TEXT PRIMARY KEY,
			url TEXT NOT NULL,
			normalized_url TEXT NOT NULL,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			selected_text TEXT,
			tags TEXT NOT NULL DEFAULT '[]',
			favicon TEXT,
			source TEXT NOT NULL,
			content_hash TEXT NOT NULL,
			saved_at INTEGER NOT NULL
		);
		INSERT INTO web_memories (id, url, normalized_url, title, content, selected_text, tags, favicon, source, content_hash, saved_at)
		VALUES
			('mem-local-1', 'https://example.com/mcp', 'https://example.com/mcp', 'MCP article', 'Model context and tools', 'context', '["tool","mcp"]', NULL, 'browser-extension', 'hash-1', 1711958400),
			('mem-local-2', 'https://example.com/agent', 'https://example.com/agent', 'Agent notes', 'Agent orchestration and tools', NULL, '["agent","tool"]', NULL, 'browser-extension', 'hash-2', 1711958460);
	`); err != nil {
		t.Fatalf("failed to seed sqlite db: %v", err)
	}

	cfg := config.Default()
	cfg.WorkspaceRoot = workspaceRoot
	server := New(cfg, stubDetector{})

	cases := []struct {
		name     string
		path     string
		contains []string
	}{
		{
			name: "pulse events",
			path: "/api/pulse/events?limit=5&afterTimestamp=100",
			contains: []string{
				`"fallback":"go-local-pulse"`,
				`"procedure":"pulse.getLatestEvents"`,
				`using local empty pulse event history`,
				`"data":[]`,
			},
		},
		{
			name: "browser memories",
			path: "/api/browser-extension/memories?search=tools&tag=tool&limit=5&offset=0",
			contains: []string{
				`"fallback":"go-local-browser-memory"`,
				`"procedure":"browserExtension.listMemories"`,
				`using local browser memories from tormentnexus.db`,
				`"mem-local-2"`,
				`"total":2`,
			},
		},
		{
			name: "browser memory stats",
			path: "/api/browser-extension/stats",
			contains: []string{
				`"fallback":"go-local-browser-memory"`,
				`"procedure":"browserExtension.stats"`,
				`using local browser memory stats from tormentnexus.db`,
				`"totalMemories":2`,
				`"uniqueUrls":2`,
				`"tag":"tool"`,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, tc.path, nil))

			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
			}
			for _, needle := range tc.contains {
				if !strings.Contains(recorder.Body.String(), needle) {
					t.Fatalf("expected response to contain %s, got %s", needle, recorder.Body.String())
				}
			}
		})
	}
}

func TestResearchQueriesFallsBackToTopicQuery(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	server := New(config.Default(), stubDetector{})
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/research/queries?topic=mcp%20bridges", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}

	for _, needle := range []string{
		`"fallback":"go-local-research"`,
		`"procedure":"research.generateQueries"`,
		`using topic-as-query research fallback`,
		`"queries":["mcp bridges"]`,
	} {
		if !strings.Contains(recorder.Body.String(), needle) {
			t.Fatalf("expected response to contain %s, got %s", needle, recorder.Body.String())
		}
	}
}

func TestResearchQueueFallsBackToLocalFiles(t *testing.T) {
	workspaceRoot := t.TempDir()
	scriptsDir := filepath.Join(workspaceRoot, "scripts")
	if err := os.MkdirAll(scriptsDir, 0o755); err != nil {
		t.Fatalf("failed to create scripts dir: %v", err)
	}

	statusContent := `{
  "processed": ["https://example.com/processed"],
  "pending": ["https://example.com/pending", "https://example.com/pending-2"],
  "failed": [
    {
      "url": "https://example.com/failed",
      "error": "timeout",
      "source": "fetcher",
      "attempts": 3,
      "last_attempt_at": "2026-04-01T00:00:00.000Z"
    }
  ]
}`
	if err := os.WriteFile(filepath.Join(scriptsDir, "ingestion-status.json"), []byte(statusContent), 0o644); err != nil {
		t.Fatalf("failed to write ingestion status: %v", err)
	}

	indexContent := `{
  "categories": {
    "research": [
      {"id":"processed-1","name":"Processed Link","url":"https://example.com/processed"},
      {"id":"pending-1","name":"Pending Link","url":"https://example.com/pending"},
      {"id":"failed-1","name":"Failed Link","url":"https://example.com/failed"}
    ]
  }
}`
	if err := os.WriteFile(filepath.Join(workspaceRoot, "TORMENTNEXUS_MASTER_INDEX.jsonc"), []byte(indexContent), 0o644); err != nil {
		t.Fatalf("failed to write master index: %v", err)
	}

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	cfg := config.Default()
	cfg.WorkspaceRoot = workspaceRoot
	server := New(cfg, stubDetector{})

	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/research/queue", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}

	for _, needle := range []string{
		`"fallback":"go-local-research"`,
		`"procedure":"research.ingestionQueue"`,
		`using local research queue files`,
		`"processed":1`,
		`"pending":2`,
		`"failed":1`,
		`"name":"Processed Link"`,
		`"name":"Failed Link"`,
	} {
		if !strings.Contains(recorder.Body.String(), needle) {
			t.Fatalf("expected response to contain %s, got %s", needle, recorder.Body.String())
		}
	}
}

func TestSettingsReadEndpointsFallBackLocally(t *testing.T) {
	workspaceRoot := t.TempDir()
	mainConfigDir := filepath.Join(workspaceRoot, ".tormentnexus")
	if err := os.MkdirAll(mainConfigDir, 0o755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(mainConfigDir, "config.json"), []byte(`{"theme":"local-dark","nested":{"enabled":true}}`), 0o644); err != nil {
		t.Fatalf("failed to write local settings config: %v", err)
	}

	configBody := `{
  "mcpServers": {
    "core": {
      "command": "node",
      "args": ["server.js"],
      "description": "Core MCP server",
      "always_on": true
    }
  }
}`
	if err := os.WriteFile(filepath.Join(mainConfigDir, "mcp.jsonc"), []byte(configBody), 0o644); err != nil {
		t.Fatalf("failed to write mcp config: %v", err)
	}

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")
	t.Setenv("NODE_ENV", "test")
	t.Setenv("PORT", "4310")

	cfg := config.Default()
	cfg.WorkspaceRoot = workspaceRoot
	cfg.MainConfigDir = mainConfigDir
	server := New(cfg, stubDetector{})

	settingsRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(settingsRecorder, httptest.NewRequest(http.MethodGet, "/api/settings", nil))
	if settingsRecorder.Code != http.StatusOK {
		t.Fatalf("expected settings status 200, got %d with body %s", settingsRecorder.Code, settingsRecorder.Body.String())
	}
	for _, needle := range []string{
		`"fallback":"go-local-settings"`,
		`"procedure":"settings.get"`,
		`using local .tormentnexus config fallback`,
		`"theme":"local-dark"`,
		`"enabled":true`,
	} {
		if !strings.Contains(settingsRecorder.Body.String(), needle) {
			t.Fatalf("expected settings response to contain %s, got %s", needle, settingsRecorder.Body.String())
		}
	}

	environmentRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(environmentRecorder, httptest.NewRequest(http.MethodGet, "/api/settings/environment", nil))
	if environmentRecorder.Code != http.StatusOK {
		t.Fatalf("expected environment status 200, got %d with body %s", environmentRecorder.Code, environmentRecorder.Body.String())
	}
	for _, needle := range []string{
		`"fallback":"go-local-settings"`,
		`"procedure":"settings.getEnvironment"`,
		`using local Go runtime environment diagnostics`,
		`"platform":"` + runtime.GOOS + `"`,
		`"arch":"` + runtime.GOARCH + `"`,
		`"NODE_ENV":"test"`,
		`"PORT":"4310"`,
	} {
		if !strings.Contains(environmentRecorder.Body.String(), needle) {
			t.Fatalf("expected environment response to contain %s, got %s", needle, environmentRecorder.Body.String())
		}
	}

	serversRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(serversRecorder, httptest.NewRequest(http.MethodGet, "/api/settings/mcp-servers", nil))
	if serversRecorder.Code != http.StatusOK {
		t.Fatalf("expected mcp servers status 200, got %d with body %s", serversRecorder.Code, serversRecorder.Body.String())
	}
	for _, needle := range []string{
		`"fallback":"go-local-settings"`,
		`"procedure":"settings.getMcpServers"`,
		`using local MCP config servers`,
		`"name":"core"`,
		`"description":"Core MCP server"`,
		`"always_on":true`,
		`"command":"node"`,
	} {
		if !strings.Contains(serversRecorder.Body.String(), needle) {
			t.Fatalf("expected mcp servers response to contain %s, got %s", needle, serversRecorder.Body.String())
		}
	}
}

func TestServerHealthFallsBackToCachedMCPMetadata(t *testing.T) {
	workspaceRoot := t.TempDir()
	mainConfigDir := filepath.Join(workspaceRoot, ".tormentnexus")
	if err := os.MkdirAll(mainConfigDir, 0o755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}
	configBody := `{
  "mcpServers": {
    "core": {
      "command": "node",
      "args": ["server.js"],
      "_meta": {
        "uuid": "srv-local-1",
        "status": "failed",
        "crashCount": 4,
        "maxAttempts": 7
      }
    }
  }
}`
	if err := os.WriteFile(filepath.Join(mainConfigDir, "mcp.jsonc"), []byte(configBody), 0o644); err != nil {
		t.Fatalf("failed to write mcp config: %v", err)
	}

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	cfg := config.Default()
	cfg.WorkspaceRoot = workspaceRoot
	cfg.MainConfigDir = mainConfigDir
	server := New(cfg, stubDetector{})

	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/server-health/check?serverUuid=srv-local-1", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected server health 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}

	for _, needle := range []string{
		`"fallback":"go-local-server-health"`,
		`"procedure":"serverHealth.check"`,
		`using cached local mcp.jsonc server metadata`,
		`"status":"ERROR"`,
		`"crashCount":4`,
		`"maxAttempts":7`,
	} {
		if !strings.Contains(recorder.Body.String(), needle) {
			t.Fatalf("expected server health response to contain %s, got %s", needle, recorder.Body.String())
		}
	}
}

func TestSymbolsReadRoutesFallBackToEmptyResults(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	cfg := config.Default()
	cfg.WorkspaceRoot = t.TempDir()
	cfg.ConfigDir = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	cases := []struct {
		name      string
		path      string
		procedure string
		reason    string
	}{
		{name: "list", path: "/api/symbols", procedure: "symbols.list", reason: "symbol pins are not initialized"},
		{name: "find", path: "/api/symbols/find?query=Run&limit=5", procedure: "symbols.find", reason: "symbol search is not initialized"},
		{name: "for_file", path: "/api/symbols/file?filePath=src%2Fapp.ts", procedure: "symbols.forFile", reason: "symbol pins are not initialized"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, tc.path, nil))
			if recorder.Code != http.StatusServiceUnavailable {
				t.Fatalf("expected %s 503, got %d with body %s", tc.name, recorder.Code, recorder.Body.String())
			}

			for _, needle := range []string{
				`"success":false`,
				`"fallback":"go-local-symbols"`,
				`"procedure":"` + tc.procedure + `"`,
				tc.reason,
				`"data":[]`,
			} {
				if !strings.Contains(recorder.Body.String(), needle) {
					t.Fatalf("expected %s fallback to contain %s, got %s", tc.name, needle, recorder.Body.String())
				}
			}
		})
	}
}

func TestGraphSymbolsFallsBackToEmptyGraph(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	cfg := config.Default()
	cfg.WorkspaceRoot = t.TempDir()
	cfg.ConfigDir = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/graph/symbols", nil))
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected graph symbols 503, got %d with body %s", recorder.Code, recorder.Body.String())
	}

	for _, needle := range []string{
		`"success":false`,
		`"fallback":"go-local-graph"`,
		`"procedure":"graph.getSymbolsGraph"`,
		`symbol graph data is not initialized`,
		`"nodes":[]`,
		`"links":[]`,
	} {
		if !strings.Contains(recorder.Body.String(), needle) {
			t.Fatalf("expected graph symbols fallback to contain %s, got %s", needle, recorder.Body.String())
		}
	}
}

func TestGraphFileReadsFallBackToEmptyLists(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	cfg := config.Default()
	cfg.WorkspaceRoot = t.TempDir()
	cfg.ConfigDir = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	cases := []struct {
		name      string
		path      string
		procedure string
	}{
		{name: "consumers", path: "/api/graph/consumers?filePath=src/app.ts", procedure: "graph.getConsumers"},
		{name: "dependencies", path: "/api/graph/dependencies?filePath=src/app.ts", procedure: "graph.getDependencies"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, tc.path, nil))
			if recorder.Code != http.StatusServiceUnavailable {
				t.Fatalf("expected %s 503, got %d with body %s", tc.name, recorder.Code, recorder.Body.String())
			}

			for _, needle := range []string{
				`"success":false`,
				`"fallback":"go-local-graph"`,
				`"procedure":"` + tc.procedure + `"`,
				`repo graph is not initialized`,
				`"data":[]`,
			} {
				if !strings.Contains(recorder.Body.String(), needle) {
					t.Fatalf("expected %s fallback to contain %s, got %s", tc.name, needle, recorder.Body.String())
				}
			}
		})
	}
}

func TestWorkflowReadRoutesFallBackToEngineZeroState(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	cfg := config.Default()
	cfg.WorkspaceRoot = t.TempDir()
	cfg.ConfigDir = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	cases := []struct {
		name      string
		path      string
		procedure string
		contains  []string
	}{
		{name: "list", path: "/api/workflows", procedure: "workflow.list", contains: []string{`"data":[]`}},
		{name: "graph", path: "/api/workflows/graph?workflowId=wf-1", procedure: "workflow.getGraph", contains: []string{`"nodes":[]`, `"edges":[]`}},
		{name: "executions", path: "/api/workflows/executions", procedure: "workflow.listExecutions", contains: []string{`"data":[]`}},
		{name: "execution", path: "/api/workflows/execution?executionId=exec-1", procedure: "workflow.getExecution", contains: []string{`"data":null`}},
		{name: "history", path: "/api/workflows/history?executionId=exec-1", procedure: "workflow.getHistory", contains: []string{`"data":[]`}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, tc.path, nil))
			if recorder.Code != http.StatusServiceUnavailable {
				t.Fatalf("expected %s 503, got %d with body %s", tc.name, recorder.Code, recorder.Body.String())
			}

			baseNeedles := []string{
				`"success":false`,
				`"fallback":"go-local-workflow"`,
				`"procedure":"` + tc.procedure + `"`,
				`workflow engine is not initialized`,
			}
			for _, needle := range append(baseNeedles, tc.contains...) {
				if !strings.Contains(recorder.Body.String(), needle) {
					t.Fatalf("expected %s fallback to contain %s, got %s", tc.name, needle, recorder.Body.String())
				}
			}
		})
	}
}

func TestToolAliasResolveFallsBackToUnresolvedState(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	cfg := config.Default()
	cfg.WorkspaceRoot = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/tool-chains/aliases/resolve?name=search", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected tool alias resolve 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}

	for _, needle := range []string{
		`"fallback":"go-local-toolchain-db"`,
		`"procedure":"toolChaining.resolveAlias"`,
		`using local tool alias from tormentnexus.db`,
		`"resolved":false`,
	} {
		if !strings.Contains(recorder.Body.String(), needle) {
			t.Fatalf("expected tool alias resolve fallback to contain %s, got %s", needle, recorder.Body.String())
		}
	}
}

func TestToolAliasesListFallsBackToLocalDB(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	workspace := t.TempDir()
	dbPath := filepath.Join(workspace, "tormentnexus.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	defer db.Close()
	if _, err := db.Exec(`
		CREATE TABLE tool_aliases (
			alias TEXT PRIMARY KEY,
			target_tool TEXT NOT NULL,
			description TEXT,
			default_arguments TEXT NOT NULL DEFAULT '{}',
			created_at INTEGER NOT NULL
		);
		INSERT INTO tool_aliases (alias, target_tool, description, default_arguments, created_at)
		VALUES ('search', 'search_tools', 'Short alias', '{}', 1711958400);
	`); err != nil {
		t.Fatalf("failed to seed sqlite db: %v", err)
	}

	cfg := config.Default()
	cfg.WorkspaceRoot = workspace
	cfg.ConfigDir = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})
	recorder := httptest.NewRecorder()

	server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/tool-chains/aliases", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected tool alias list 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}

	for _, needle := range []string{
		`"fallback":"go-local-toolchain-db"`,
		`"procedure":"toolChaining.listAliases"`,
		`using local tool aliases from tormentnexus.db`,
		`"alias":"search"`,
		`"originalName":"search_tools"`,
	} {
		if !strings.Contains(recorder.Body.String(), needle) {
			t.Fatalf("expected tool alias list fallback to contain %s, got %s", needle, recorder.Body.String())
		}
	}
}

func TestMemoryExportFallsBackToLocalSnapshotAndRegistry(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	t.Run("json_provider_snapshot", func(t *testing.T) {
		workspaceRoot := t.TempDir()
		snapshot := `[
  {
    "uuid": "mem-1",
    "content": "Saved memory",
    "metadata": {
      "section": "user_facts"
    },
    "userId": "default",
    "createdAt": "2026-01-01T00:00:00Z"
  }
]`
		if err := os.WriteFile(filepath.Join(workspaceRoot, "memory.json"), []byte(snapshot), 0o644); err != nil {
			t.Fatalf("failed to seed memory.json: %v", err)
		}

		cfg := config.Default()
		cfg.WorkspaceRoot = workspaceRoot
		cfg.ConfigDir = t.TempDir()
		cfg.MainConfigDir = t.TempDir()
		server := New(cfg, stubDetector{})

		recorder := httptest.NewRecorder()
		server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/memory/export?userId=default&format=json-provider", nil))

		if recorder.Code != http.StatusOK {
			t.Fatalf("expected snapshot export 200, got %d with body %s", recorder.Code, recorder.Body.String())
		}

		body := recorder.Body.String()
		for _, needle := range []string{
			`"fallback":"go-local-memory"`,
			`"procedure":"memory.exportMemories"`,
			`using local memory export sources`,
			`"format":"json-provider"`,
			`\n  {\n    \"uuid\": \"mem-1\"`,
		} {
			if !strings.Contains(body, needle) {
				t.Fatalf("expected snapshot export fallback to contain %s, got %s", needle, body)
			}
		}
	})

	t.Run("registry_formats", func(t *testing.T) {
		workspaceRoot := t.TempDir()
		contextsDir := filepath.Join(workspaceRoot, ".tormentnexus", "memory")
		if err := os.MkdirAll(contextsDir, 0o755); err != nil {
			t.Fatalf("failed to create memory dir: %v", err)
		}
		contexts := `[
  {
    "id": "ctx-local-1",
    "title": "Saved context",
    "source": "docs",
    "createdAt": 1704067200000,
    "metadata": {
      "section": "project_context",
      "tags": ["go", "parity"],
      "source": "docs"
    }
  }
]`
		if err := os.WriteFile(filepath.Join(contextsDir, "contexts.json"), []byte(contexts), 0o644); err != nil {
			t.Fatalf("failed to seed contexts registry: %v", err)
		}

		cfg := config.Default()
		cfg.WorkspaceRoot = workspaceRoot
		cfg.ConfigDir = t.TempDir()
		cfg.MainConfigDir = t.TempDir()
		server := New(cfg, stubDetector{})

		checks := []struct {
			path     string
			contains []string
		}{
			{path: "/api/memory/export?userId=default&format=json", contains: []string{`"format":"json"`, `\"uuid\": \"ctx-local-1\"`}},
			{path: "/api/memory/export?userId=default&format=jsonl", contains: []string{`"format":"jsonl"`, `\"uuid\":\"ctx-local-1\"`}},
			{path: "/api/memory/export?userId=default&format=csv", contains: []string{`"format":"csv"`, `uuid,content,userId,agentId,createdAt,metadata`, `ctx-local-1`}},
			{path: "/api/memory/export?userId=default&format=sectioned-memory-store", contains: []string{`"format":"sectioned-memory-store"`, `\"section\": \"project_context\"`, `\"uuid\": \"ctx-local-1\"`}},
		}

		for _, tc := range checks {
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, tc.path, nil))

			if recorder.Code != http.StatusOK {
				t.Fatalf("expected registry export 200 for %s, got %d with body %s", tc.path, recorder.Code, recorder.Body.String())
			}

			body := recorder.Body.String()
			if !strings.Contains(body, `"fallback":"go-local-memory"`) || !strings.Contains(body, `"procedure":"memory.exportMemories"`) {
				t.Fatalf("expected local export fallback metadata for %s, got %s", tc.path, body)
			}

			for _, needle := range tc.contains {
				if !strings.Contains(body, needle) {
					t.Fatalf("expected export body for %s to contain %s, got %s", tc.path, needle, body)
				}
			}
		}
	})
}

func TestAgentMemoryStatsFallsBackToPersistedState(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	cfg := config.Default()
	cfg.WorkspaceRoot = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	seedPersistedAgentMemories(t, cfg.WorkspaceRoot)
	server := New(cfg, stubDetector{})

	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/agent-memory/stats", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected agent memory stats 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}

	for _, needle := range []string{
		`"success":true`,
		`"fallback":"go-local-agent-memory"`,
		`"procedure":"agentMemory.stats"`,
		`using local persisted agent memory stats`,
		`"session":0`,
		`"working":2`,
		`"longTerm":2`,
		`"total":4`,
	} {
		if !strings.Contains(recorder.Body.String(), needle) {
			t.Fatalf("expected agent memory stats fallback to contain %s, got %s", needle, recorder.Body.String())
		}
	}
}

func TestAgentMemoryExportFallsBackToPersistedSnapshot(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	cfg := config.Default()
	cfg.WorkspaceRoot = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	seedPersistedAgentMemories(t, cfg.WorkspaceRoot)
	server := New(cfg, stubDetector{})

	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/agent-memory/export", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected agent memory export 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}

	for _, needle := range []string{
		`"success":true`,
		`"fallback":"go-local-agent-memory"`,
		`"procedure":"agentMemory.export"`,
		`exporting persisted agent memories locally`,
		`obs-fix-1`,
		`summary-1`,
	} {
		if !strings.Contains(recorder.Body.String(), needle) {
			t.Fatalf("expected agent memory export fallback to contain %s, got %s", needle, recorder.Body.String())
		}
	}
}

func TestAgentMemoryReadRoutesFallBackToPersistedResults(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	cfg := config.Default()
	cfg.WorkspaceRoot = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	seedPersistedAgentMemories(t, cfg.WorkspaceRoot)
	server := New(cfg, stubDetector{})

	cases := []struct {
		name      string
		path      string
		procedure string
		needles   []string
	}{
		{name: "search", path: "/api/agent-memory/search?query=memory&type=working&namespace=project&limit=5", procedure: "agentMemory.search", needles: []string{`"obs-fix-1"`, `using local persisted agent memory search`}},
		{name: "recent", path: "/api/agent-memory/recent?limit=5", procedure: "agentMemory.getRecent", needles: []string{`"summary-1"`, `using local persisted recent agent memories`}},
		{name: "by_type", path: "/api/agent-memory/by-type?type=working", procedure: "agentMemory.getByType", needles: []string{`"obs-discovery-1"`, `filtered by type`}},
		{name: "by_namespace", path: "/api/agent-memory/by-namespace?namespace=project", procedure: "agentMemory.getByNamespace", needles: []string{`"prompt-1"`, `filtered by namespace`}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, tc.path, nil))
			if recorder.Code != http.StatusOK {
				t.Fatalf("expected %s 200, got %d with body %s", tc.name, recorder.Code, recorder.Body.String())
			}

			for _, needle := range append([]string{
				`"success":true`,
				`"fallback":"go-local-agent-memory"`,
				`"procedure":"` + tc.procedure + `"`,
			}, tc.needles...) {
				if !strings.Contains(recorder.Body.String(), needle) {
					t.Fatalf("expected %s fallback to contain %s, got %s", tc.name, needle, recorder.Body.String())
				}
			}
		})
	}
}

func TestAgentMemoryMutationRoutesFallBackToLocalPersistence(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	cfg := config.Default()
	cfg.WorkspaceRoot = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	seedPersistedAgentMemories(t, cfg.WorkspaceRoot)
	server := New(cfg, stubDetector{})

	addRecorder := httptest.NewRecorder()
	addReq := httptest.NewRequest(http.MethodPost, "/api/agent-memory/add", strings.NewReader(`{"content":"remember local agent parity","type":"working","namespace":"project","tags":["parity"]}`))
	server.Handler().ServeHTTP(addRecorder, addReq)
	if addRecorder.Code != http.StatusOK || !strings.Contains(addRecorder.Body.String(), `persisted agent memory locally`) || !strings.Contains(addRecorder.Body.String(), `remember local agent parity`) {
		t.Fatalf("expected local agent-memory add fallback, got %d %s", addRecorder.Code, addRecorder.Body.String())
	}

	deleteRecorder := httptest.NewRecorder()
	deleteReq := httptest.NewRequest(http.MethodPost, "/api/agent-memory/delete", strings.NewReader(`{"id":"prompt-1"}`))
	server.Handler().ServeHTTP(deleteRecorder, deleteReq)
	if deleteRecorder.Code != http.StatusOK || !strings.Contains(deleteRecorder.Body.String(), `deleting persisted agent memory locally`) || !strings.Contains(deleteRecorder.Body.String(), `"data":true`) {
		t.Fatalf("expected local agent-memory delete fallback, got %d %s", deleteRecorder.Code, deleteRecorder.Body.String())
	}

	addSessionRecorder := httptest.NewRecorder()
	addSessionReq := httptest.NewRequest(http.MethodPost, "/api/agent-memory/add", strings.NewReader(`{"content":"temporary session context","type":"session","namespace":"agent"}`))
	server.Handler().ServeHTTP(addSessionRecorder, addSessionReq)
	if addSessionRecorder.Code != http.StatusOK {
		t.Fatalf("expected local agent-memory session add fallback, got %d %s", addSessionRecorder.Code, addSessionRecorder.Body.String())
	}

	clearRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(clearRecorder, httptest.NewRequest(http.MethodPost, "/api/agent-memory/clear-session", nil))
	if (clearRecorder.Code != http.StatusOK || !strings.Contains(clearRecorder.Body.String(), `cleared persisted session-tier agent memories locally`) || !strings.Contains(clearRecorder.Body.String(), `"cleared":1`)) && !strings.Contains(clearRecorder.Body.String(), `"upstreamBase"`) {
		t.Fatalf("expected local agent-memory clear-session fallback, got %d %s", clearRecorder.Code, clearRecorder.Body.String())
	}

	recentRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recentRecorder, httptest.NewRequest(http.MethodGet, "/api/agent-memory/recent?limit=10", nil))
	if (recentRecorder.Code != http.StatusOK || strings.Contains(recentRecorder.Body.String(), `"prompt-1"`) || !strings.Contains(recentRecorder.Body.String(), `remember local agent parity`)) && !strings.Contains(recentRecorder.Body.String(), `"upstreamBase"`) {
		t.Fatalf("expected persisted agent-memory mutation effects, got %d %s", recentRecorder.Code, recentRecorder.Body.String())
	}
}

func TestAgentMemoryHandoffAndPickupFallBackToLocalPersistence(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	cfg := config.Default()
	cfg.WorkspaceRoot = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	seedPersistedAgentMemoryRelations(t, cfg.WorkspaceRoot)
	server := New(cfg, stubDetector{})

	addSessionRecorder := httptest.NewRecorder()
	addSessionReq := httptest.NewRequest(http.MethodPost, "/api/agent-memory/add", strings.NewReader(`{"content":"volatile handoff context","type":"session","namespace":"agent"}`))
	server.Handler().ServeHTTP(addSessionRecorder, addSessionReq)
	if addSessionRecorder.Code != http.StatusOK {
		t.Fatalf("expected local session memory before handoff, got %d %s", addSessionRecorder.Code, addSessionRecorder.Body.String())
	}

	handoffRecorder := httptest.NewRecorder()
	handoffReq := httptest.NewRequest(http.MethodPost, "/api/agent-memory/handoff", strings.NewReader(`{"notes":"bridge handoff"}`))
	server.Handler().ServeHTTP(handoffRecorder, handoffReq)
	if (handoffRecorder.Code != http.StatusOK || !strings.Contains(handoffRecorder.Body.String(), `generated local agent-memory handoff artifact`)) && !strings.Contains(handoffRecorder.Body.String(), `"upstreamBase"`) {
		t.Fatalf("expected local handoff fallback, got %d %s", handoffRecorder.Code, handoffRecorder.Body.String())
	}

	var handoffResponse struct {
		Success bool   `json:"success"`
		Data    string `json:"data"`
	}
	if err := json.Unmarshal(handoffRecorder.Body.Bytes(), &handoffResponse); err != nil {
		t.Fatalf("unmarshal handoff response: %v", err)
	}
	if (!handoffResponse.Success || !strings.Contains(handoffResponse.Data, `"notes": "bridge handoff"`) || !strings.Contains(handoffResponse.Data, `volatile handoff context`)) && !strings.Contains(handoffRecorder.Body.String(), `"upstreamBase"`) {
		t.Fatalf("expected handoff artifact payload, got %s", handoffRecorder.Body.String())
	}

	clearRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(clearRecorder, httptest.NewRequest(http.MethodPost, "/api/agent-memory/clear-session", nil))
	if clearRecorder.Code != http.StatusOK {
		t.Fatalf("expected clear-session fallback before pickup, got %d %s", clearRecorder.Code, clearRecorder.Body.String())
	}

	pickupRecorder := httptest.NewRecorder()
	pickupReq := httptest.NewRequest(http.MethodPost, "/api/agent-memory/pickup", strings.NewReader(`{"artifact":`+strconv.Quote(handoffResponse.Data)+`}`))
	server.Handler().ServeHTTP(pickupRecorder, pickupReq)
	if (pickupRecorder.Code != http.StatusOK || !strings.Contains(pickupRecorder.Body.String(), `restored local agent-memory handoff artifact`) || (!strings.Contains(pickupRecorder.Body.String(), `"count":1`) && !strings.Contains(pickupRecorder.Body.String(), `"count":0`))) && !strings.Contains(pickupRecorder.Body.String(), `"upstreamBase"`) {
		t.Fatalf("expected local pickup fallback, got %d %s", pickupRecorder.Code, pickupRecorder.Body.String())
	}

	recentRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recentRecorder, httptest.NewRequest(http.MethodGet, "/api/agent-memory/recent?type=session&limit=10", nil))
	if recentRecorder.Code != http.StatusOK || !strings.Contains(recentRecorder.Body.String(), `volatile handoff context`) {
		t.Fatalf("expected pickup to restore session memory, got %d %s", recentRecorder.Code, recentRecorder.Body.String())
	}
}

func TestToolsRuntimeDetectionFallsBackLocally(t *testing.T) {
	workspaceRoot := t.TempDir()
	mainConfigDir := filepath.Join(workspaceRoot, ".tormentnexus")
	if err := os.MkdirAll(mainConfigDir, 0o755); err != nil {
		t.Fatalf("failed to create main config dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(workspaceRoot, "apps", "tormentnexus-extension", "dist-chromium"), 0o755); err != nil {
		t.Fatalf("failed to create chromium dist: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(workspaceRoot, "packages", "vscode", "dist"), 0o755); err != nil {
		t.Fatalf("failed to create vscode dist: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workspaceRoot, "packages", "vscode", "dist", "extension.js"), []byte("console.log('ok')"), 0o644); err != nil {
		t.Fatalf("failed to write vscode artifact: %v", err)
	}
	if err := os.WriteFile(filepath.Join(mainConfigDir, "mcp.jsonc"), []byte(`{"mcpServers":{"core":{"command":"node"}}}`), 0o644); err != nil {
		t.Fatalf("failed to write mcp config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workspaceRoot, "apps", "tormentnexus-extension", "package.json"), []byte(`{"version":"1.2.3"}`), 0o644); err != nil {
		t.Fatalf("failed to write extension package json: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(workspaceRoot, "apps", "tormentnexus-extension"), 0o755); err != nil {
		t.Fatalf("failed to create extension dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workspaceRoot, "packages", "vscode", "package.json"), []byte(`{"version":"0.9.0"}`), 0o644); err != nil {
		t.Fatalf("failed to write vscode package json: %v", err)
	}

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	detector := stubDetector{tools: []controlplane.Tool{
		{Type: "codex", Name: "Codex CLI", Command: "codex", Available: true, Version: "1.0.0", Path: "C:\\tools\\codex.exe", Capabilities: []string{"chat", "code"}},
		{Type: "gemini", Name: "Gemini CLI", Command: "gemini", Available: true, Version: "2.0.0", Path: "C:\\tools\\gemini.exe", Capabilities: []string{"chat", "multimodal"}},
	}}

	cfg := config.Default()
	cfg.WorkspaceRoot = workspaceRoot
	cfg.MainConfigDir = mainConfigDir
	server := New(cfg, detector)

	cliRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(cliRecorder, httptest.NewRequest(http.MethodGet, "/api/tools/detect-cli-harnesses", nil))
	if cliRecorder.Code != http.StatusOK {
		t.Fatalf("expected cli harness status 200, got %d with body %s", cliRecorder.Code, cliRecorder.Body.String())
	}
	for _, needle := range []string{
		`"fallback":"go-local-tools"`,
		`"procedure":"tools.detectCliHarnesses"`,
		`using local Go harness detection`,
		`"id":"codex"`,
		`"resolvedPath":"C:\\tools\\codex.exe"`,
	} {
		if !strings.Contains(cliRecorder.Body.String(), needle) {
			t.Fatalf("expected cli harness response to contain %s, got %s", needle, cliRecorder.Body.String())
		}
	}

	environmentRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(environmentRecorder, httptest.NewRequest(http.MethodGet, "/api/tools/detect-execution-environment", nil))
	if environmentRecorder.Code != http.StatusOK {
		t.Fatalf("expected execution environment status 200, got %d with body %s", environmentRecorder.Code, environmentRecorder.Body.String())
	}
	for _, needle := range []string{
		`"fallback":"go-local-tools"`,
		`"procedure":"tools.detectExecutionEnvironment"`,
		`using local Go execution environment detection`,
		`"harnesses":[`,
		`"tools":[`,
		`"os":"` + runtime.GOOS + `"`,
	} {
		if !strings.Contains(environmentRecorder.Body.String(), needle) {
			t.Fatalf("expected execution environment response to contain %s, got %s", needle, environmentRecorder.Body.String())
		}
	}

	installRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(installRecorder, httptest.NewRequest(http.MethodGet, "/api/tools/detect-install-surfaces", nil))
	if installRecorder.Code != http.StatusOK {
		t.Fatalf("expected install surfaces status 200, got %d with body %s", installRecorder.Code, installRecorder.Body.String())
	}
	for _, needle := range []string{
		`"fallback":"go-local-tools"`,
		`"procedure":"tools.detectInstallSurfaces"`,
		`using local Go install-surface detection`,
		`"id":"browser-extension-chromium"`,
		`"id":"vscode-extension"`,
		`"status":"ready"`,
		`"status":"partial"`,
	} {
		if !strings.Contains(installRecorder.Body.String(), needle) {
			t.Fatalf("expected install surfaces response to contain %s, got %s", needle, installRecorder.Body.String())
		}
	}
}

func TestGitStatusFallsBackLocally(t *testing.T) {
	workspaceRoot := t.TempDir()
	initRepo := exec.Command("git", "init", "-b", "main")
	initRepo.Dir = workspaceRoot
	if output, err := initRepo.CombinedOutput(); err != nil {
		t.Fatalf("failed to init git repo: %v (%s)", err, string(output))
	}
	configName := exec.Command("git", "config", "user.name", "tormentnexus Test")
	configName.Dir = workspaceRoot
	if output, err := configName.CombinedOutput(); err != nil {
		t.Fatalf("failed to configure git user.name: %v (%s)", err, string(output))
	}
	configEmail := exec.Command("git", "config", "user.email", "test@tormentnexus.local")
	configEmail.Dir = workspaceRoot
	if output, err := configEmail.CombinedOutput(); err != nil {
		t.Fatalf("failed to configure git user.email: %v (%s)", err, string(output))
	}

	trackedPath := filepath.Join(workspaceRoot, "tracked.txt")
	if err := os.WriteFile(trackedPath, []byte("hello\n"), 0o644); err != nil {
		t.Fatalf("failed to write tracked file: %v", err)
	}
	addTracked := exec.Command("git", "add", "tracked.txt")
	addTracked.Dir = workspaceRoot
	if output, err := addTracked.CombinedOutput(); err != nil {
		t.Fatalf("failed to stage tracked file: %v (%s)", err, string(output))
	}
	commitTracked := exec.Command("git", "commit", "-m", "seed tracked file")
	commitTracked.Dir = workspaceRoot
	if output, err := commitTracked.CombinedOutput(); err != nil {
		t.Fatalf("failed to commit tracked file: %v (%s)", err, string(output))
	}

	modifiedPath := filepath.Join(workspaceRoot, "modified.txt")
	if err := os.WriteFile(modifiedPath, []byte("first\n"), 0o644); err != nil {
		t.Fatalf("failed to write modified file: %v", err)
	}
	addModified := exec.Command("git", "add", "modified.txt")
	addModified.Dir = workspaceRoot
	if output, err := addModified.CombinedOutput(); err != nil {
		t.Fatalf("failed to add modified file: %v (%s)", err, string(output))
	}
	if err := os.WriteFile(modifiedPath, []byte("first\nsecond\n"), 0o644); err != nil {
		t.Fatalf("failed to modify tracked file: %v", err)
	}

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	cfg := config.Default()
	cfg.WorkspaceRoot = workspaceRoot
	server := New(cfg, stubDetector{})

	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/git/status", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected git status 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}

	for _, needle := range []string{
		`"fallback":"go-local-git"`,
		`"procedure":"git.getStatus"`,
		`using local git status fallback`,
		`"branch":"main"`,
		`"clean":false`,
		`"modified.txt"`,
	} {
		if !strings.Contains(recorder.Body.String(), needle) {
			t.Fatalf("expected git status response to contain %s, got %s", needle, recorder.Body.String())
		}
	}
}

func TestGitLogFallsBackLocally(t *testing.T) {
	workspaceRoot := t.TempDir()
	initRepo := exec.Command("git", "init", "-b", "main")
	initRepo.Dir = workspaceRoot
	if output, err := initRepo.CombinedOutput(); err != nil {
		t.Fatalf("failed to init git repo: %v (%s)", err, string(output))
	}
	for _, args := range [][]string{
		{"config", "user.name", "tormentnexus Test"},
		{"config", "user.email", "test@tormentnexus.local"},
	} {
		command := exec.Command("git", args...)
		command.Dir = workspaceRoot
		if output, err := command.CombinedOutput(); err != nil {
			t.Fatalf("failed to run git %v: %v (%s)", args, err, string(output))
		}
	}

	for _, commit := range []struct {
		file    string
		content string
		message string
	}{
		{file: "first.txt", content: "one\n", message: "first commit"},
		{file: "second.txt", content: "two\n", message: "second commit"},
	} {
		if err := os.WriteFile(filepath.Join(workspaceRoot, commit.file), []byte(commit.content), 0o644); err != nil {
			t.Fatalf("failed to write %s: %v", commit.file, err)
		}
		add := exec.Command("git", "add", commit.file)
		add.Dir = workspaceRoot
		if output, err := add.CombinedOutput(); err != nil {
			t.Fatalf("failed to add %s: %v (%s)", commit.file, err, string(output))
		}
		commitCmd := exec.Command("git", "commit", "-m", commit.message)
		commitCmd.Dir = workspaceRoot
		if output, err := commitCmd.CombinedOutput(); err != nil {
			t.Fatalf("failed to commit %s: %v (%s)", commit.file, err, string(output))
		}
	}

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	cfg := config.Default()
	cfg.WorkspaceRoot = workspaceRoot
	server := New(cfg, stubDetector{})

	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/git/log?limit=1", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected git log 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}

	for _, needle := range []string{
		`"fallback":"go-local-git"`,
		`"procedure":"git.getLog"`,
		`using local git log fallback`,
		`"message":"second commit"`,
		`"author":"tormentnexus Test"`,
	} {
		if !strings.Contains(recorder.Body.String(), needle) {
			t.Fatalf("expected git log response to contain %s, got %s", needle, recorder.Body.String())
		}
	}
	if strings.Contains(recorder.Body.String(), `"message":"first commit"`) {
		t.Fatalf("expected limit=1 to exclude first commit, got %s", recorder.Body.String())
	}
}

func TestSubmoduleReadEndpointsFallBackLocally(t *testing.T) {
	workspaceRoot := t.TempDir()
	gitmodules := `[submodule "submodules/tormentnexus"]
	path = submodules/tormentnexus
	url = https://github.com/example/tormentnexus.git
`
	if err := os.WriteFile(filepath.Join(workspaceRoot, ".gitmodules"), []byte(gitmodules), 0o644); err != nil {
		t.Fatalf("failed to write .gitmodules: %v", err)
	}

	tormentnexusPath := filepath.Join(workspaceRoot, "submodules", "tormentnexus")
	if err := os.MkdirAll(filepath.Join(tormentnexusPath, "dist"), 0o755); err != nil {
		t.Fatalf("failed to create submodule dist dir: %v", err)
	}
	packageJSON := `{
  "keywords": ["mcp-server"],
  "dependencies": {"@modelcontextprotocol/sdk": "^1.0.0"},
  "scripts": {"build": "tsc", "start": "node index.js"}
}`
	if err := os.WriteFile(filepath.Join(tormentnexusPath, "package.json"), []byte(packageJSON), 0o644); err != nil {
		t.Fatalf("failed to write package.json: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(tormentnexusPath, "node_modules"), 0o755); err != nil {
		t.Fatalf("failed to create node_modules dir: %v", err)
	}

	pythonPath := filepath.Join(workspaceRoot, "submodules", "python-tool")
	if err := os.MkdirAll(filepath.Join(pythonPath, ".venv"), 0o755); err != nil {
		t.Fatalf("failed to create python venv dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pythonPath, "requirements.txt"), []byte("requests\n"), 0o644); err != nil {
		t.Fatalf("failed to write requirements.txt: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pythonPath, "main.py"), []byte("print('ok')\n"), 0o644); err != nil {
		t.Fatalf("failed to write main.py: %v", err)
	}

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	cfg := config.Default()
	cfg.WorkspaceRoot = workspaceRoot
	server := New(cfg, stubDetector{})

	listRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(listRecorder, httptest.NewRequest(http.MethodGet, "/api/submodules", nil))
	if listRecorder.Code != http.StatusOK {
		t.Fatalf("expected submodule list 200, got %d with body %s", listRecorder.Code, listRecorder.Body.String())
	}

	for _, needle := range []string{
		`"fallback":"go-local-submodules"`,
		`"procedure":"submodule.list"`,
		`using local .gitmodules submodule fallback`,
		`"path":"submodules/tormentnexus"`,
		`"status":"clean"`,
		`"capabilities":["mcp-server","mcp-sdk","build"]`,
		`"isInstalled":true`,
		`"isBuilt":true`,
		`"startCommand":"npm start"`,
	} {
		if !strings.Contains(listRecorder.Body.String(), needle) {
			t.Fatalf("expected submodule list response to contain %s, got %s", needle, listRecorder.Body.String())
		}
	}

	capabilitiesRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(capabilitiesRecorder, httptest.NewRequest(http.MethodGet, "/api/submodules/capabilities?path=submodules%2Fpython-tool", nil))
	if capabilitiesRecorder.Code != http.StatusOK {
		t.Fatalf("expected submodule capabilities 200, got %d with body %s", capabilitiesRecorder.Code, capabilitiesRecorder.Body.String())
	}

	for _, needle := range []string{
		`"fallback":"go-local-submodules"`,
		`"procedure":"submodule.detectCapabilities"`,
		`using local submodule capability fallback`,
		`"caps":["python"]`,
		`"startCommand":"python main.py"`,
	} {
		if !strings.Contains(capabilitiesRecorder.Body.String(), needle) {
			t.Fatalf("expected submodule capabilities response to contain %s, got %s", needle, capabilitiesRecorder.Body.String())
		}
	}
}

func TestKnowledgeReadEndpointsFallBackLocally(t *testing.T) {
	workspaceRoot := t.TempDir()
	memoryDir := filepath.Join(workspaceRoot, ".tormentnexus", "memory")
	if err := os.MkdirAll(memoryDir, 0o755); err != nil {
		t.Fatalf("failed to create memory dir: %v", err)
	}
	contexts := `[
  {"id":"ctx-1"},
  {"id":"ctx-2"},
  {"id":"ctx-3"}
]`
	if err := os.WriteFile(filepath.Join(memoryDir, "contexts.json"), []byte(contexts), 0o644); err != nil {
		t.Fatalf("failed to write contexts.json: %v", err)
	}

	knowledgeDir := filepath.Join(workspaceRoot, "knowledge")
	if err := os.MkdirAll(knowledgeDir, 0o755); err != nil {
		t.Fatalf("failed to create knowledge dir: %v", err)
	}
	resources := `{
  "lastUpdated": "2026-04-01",
  "categories": [{"name":"MCP","count":2}]
}`
	if err := os.WriteFile(filepath.Join(knowledgeDir, "resources.json"), []byte(resources), 0o644); err != nil {
		t.Fatalf("failed to write resources.json: %v", err)
	}

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	cfg := config.Default()
	cfg.WorkspaceRoot = workspaceRoot
	server := New(cfg, stubDetector{})

	statsRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(statsRecorder, httptest.NewRequest(http.MethodGet, "/api/knowledge/stats", nil))
	if statsRecorder.Code != http.StatusOK {
		t.Fatalf("expected knowledge stats 200, got %d with body %s", statsRecorder.Code, statsRecorder.Body.String())
	}

	for _, needle := range []string{
		`"fallback":"go-local-knowledge"`,
		`"procedure":"knowledge.getStats"`,
		`using local memory context count for knowledge stats`,
		`"count":3`,
	} {
		if !strings.Contains(statsRecorder.Body.String(), needle) {
			t.Fatalf("expected knowledge stats response to contain %s, got %s", needle, statsRecorder.Body.String())
		}
	}

	resourcesRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(resourcesRecorder, httptest.NewRequest(http.MethodGet, "/api/knowledge/resources", nil))
	if resourcesRecorder.Code != http.StatusOK {
		t.Fatalf("expected knowledge resources 200, got %d with body %s", resourcesRecorder.Code, resourcesRecorder.Body.String())
	}

	for _, needle := range []string{
		`"fallback":"go-local-knowledge"`,
		`"procedure":"knowledge.getResources"`,
		`using local knowledge resources file`,
		`"lastUpdated":"2026-04-01"`,
		`"name":"MCP"`,
	} {
		if !strings.Contains(resourcesRecorder.Body.String(), needle) {
			t.Fatalf("expected knowledge resources response to contain %s, got %s", needle, resourcesRecorder.Body.String())
		}
	}
}

func TestMCPSearchToolsFallsBackToLocalInventory(t *testing.T) {
	workspaceRoot := t.TempDir()
	toolsDir := filepath.Join(workspaceRoot, "submodules", "tormentnexus", "tools")
	if err := os.MkdirAll(toolsDir, 0o755); err != nil {
		t.Fatalf("failed to create tormentnexus tools dir: %v", err)
	}
	toolSource := `package tools

var SearchTools = struct{
	Name string
}{
	Name: "search_tools",
}
`
	if err := os.WriteFile(filepath.Join(toolsDir, "search.go"), []byte(toolSource), 0o644); err != nil {
		t.Fatalf("failed to write tormentnexus tool source: %v", err)
	}

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	cfg := config.Default()
	cfg.WorkspaceRoot = workspaceRoot
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{tools: []controlplane.Tool{
		{Type: "go", Name: "Go", Command: "go", Available: true},
	}})

	request := httptest.NewRequest(http.MethodGet, "/api/mcp/tools/search?query=search", nil)
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected fallback status 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"fallback":"go-local-mcp"`) {
		t.Fatalf("expected go-local-mcp fallback metadata, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"procedure":"mcp.searchTools"`) {
		t.Fatalf("expected searchTools procedure metadata, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `using local MCP inventory cache`) {
		t.Fatalf("expected local search fallback reason, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"name":"start_search"`) && !strings.Contains(recorder.Body.String(), `"name":"search_tools"`) {
		t.Fatalf("expected local source-backed search result, got %s", recorder.Body.String())
	}
}

func TestMCPCallToolFallsBackToLocalMetaTools(t *testing.T) {
	workspaceRoot := t.TempDir()
	toolsDir := filepath.Join(workspaceRoot, "submodules", "tormentnexus", "tools")
	if err := os.MkdirAll(toolsDir, 0o755); err != nil {
		t.Fatalf("failed to create tormentnexus tools dir: %v", err)
	}
	toolSource := `package tools

var SearchTools = struct{
	Name string
}{
	Name: "search_tools",
}

var ListAllTools = struct{
	Name string
}{
	Name: "list_all_tools",
}
`
	if err := os.WriteFile(filepath.Join(toolsDir, "search.go"), []byte(toolSource), 0o644); err != nil {
		t.Fatalf("failed to write tormentnexus tool source: %v", err)
	}

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	cfg := config.Default()
	cfg.WorkspaceRoot = workspaceRoot
	cfg.ConfigDir = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{tools: []controlplane.Tool{
		{Type: "go", Name: "Go", Command: "go", Available: true},
	}})

	callRequest := httptest.NewRequest(http.MethodPost, "/api/mcp/tools/call", strings.NewReader(`{"name":"search_tools","args":{"query":"search"}}`))
	callRequest.Header.Set("content-type", "application/json")
	callRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(callRecorder, callRequest)
	if callRecorder.Code != http.StatusOK || !strings.Contains(callRecorder.Body.String(), `"fallback":"go-local-mcp"`) || !strings.Contains(callRecorder.Body.String(), `tormentnexus`) || !strings.Contains(callRecorder.Body.String(), `search_tools`) {
		t.Fatalf("expected local callTool fallback response, got %d %s", callRecorder.Code, callRecorder.Body.String())
	}
	if !strings.Contains(callRecorder.Body.String(), `using local MCP meta-tool execution`) {
		t.Fatalf("expected local callTool fallback reason, got %s", callRecorder.Body.String())
	}

	autoRequest := httptest.NewRequest(http.MethodPost, "/api/mcp/tools/auto-call", strings.NewReader(`{"objective":"find the right tool","context":"repo: tormentnexus"}`))
	autoRequest.Header.Set("content-type", "application/json")
	autoRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(autoRecorder, autoRequest)
	if autoRecorder.Code != http.StatusOK || !strings.Contains(autoRecorder.Body.String(), `"fallback":"go-local-mcp"`) || !strings.Contains(autoRecorder.Body.String(), `Auto-Execution Logic`) {
		t.Fatalf("expected local auto_call_tool fallback response, got %d %s", autoRecorder.Code, autoRecorder.Body.String())
	}
	if !strings.Contains(autoRecorder.Body.String(), `using local auto-call meta-tool execution`) {
		t.Fatalf("expected local auto-call fallback reason, got %s", autoRecorder.Body.String())
	}
}

func TestMCPToolSchemaFallsBackToLocalMetaSchemas(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	cfg := config.Default()
	cfg.WorkspaceRoot = t.TempDir()
	cfg.ConfigDir = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	request := httptest.NewRequest(http.MethodPost, "/api/mcp/tools/schema", strings.NewReader(`{"name":"search_tools"}`))
	request.Header.Set("content-type", "application/json")
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected fallback schema status 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"fallback":"go-local-mcp"`) || !strings.Contains(recorder.Body.String(), `"procedure":"mcp.getToolSchema"`) {
		t.Fatalf("expected local schema fallback metadata, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `using local MCP tool schema fallback`) {
		t.Fatalf("expected local schema fallback reason, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"query"`) || !strings.Contains(recorder.Body.String(), `"inputSchema"`) {
		t.Fatalf("expected local search_tools schema payload, got %s", recorder.Body.String())
	}
}

func TestMCPRegistrySnapshotFallsBackToMasterIndex(t *testing.T) {
	workspaceRoot := t.TempDir()
	indexContent := `{
  // comment
  "categories": {
    "mcpServers": [
      {
        "id": "mcp-1",
        "name": "Core MCP",
        "url": "https://example.com/mcp",
        "description": "Core MCP server",
        "tags": ["mcp", "tools"]
      }
    ],
    "other": [
      {
        "id": "other-1",
        "name": "Not MCP",
        "url": "https://example.com/other",
        "description": "Ignore me"
      }
    ]
  }
}`
	if err := os.WriteFile(filepath.Join(workspaceRoot, "TORMENTNEXUS_MASTER_INDEX.jsonc"), []byte(indexContent), 0o644); err != nil {
		t.Fatalf("failed to write master index fixture: %v", err)
	}

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	cfg := config.Default()
	cfg.WorkspaceRoot = workspaceRoot
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	request := httptest.NewRequest(http.MethodGet, "/api/mcp/servers/registry-snapshot", nil)
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected fallback status 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"fallback":"go-master-index"`) {
		t.Fatalf("expected go-master-index fallback metadata, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `using local master index registry snapshot`) {
		t.Fatalf("expected master index registry fallback reason, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"name":"Core MCP"`) {
		t.Fatalf("expected MCP-like registry entry, got %s", recorder.Body.String())
	}
	if strings.Contains(recorder.Body.String(), `"name":"Not MCP"`) {
		t.Fatalf("expected non-MCP entry to be filtered, got %s", recorder.Body.String())
	}
}

func TestMCPJsoncEditorFallsBackToLocalFile(t *testing.T) {
	mainConfigDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(mainConfigDir, "mcp.jsonc"), []byte("// tormentnexus MCP configuration\n{\n  \"mcpServers\": {}\n}\n"), 0o644); err != nil {
		t.Fatalf("failed to write local mcp jsonc: %v", err)
	}

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	cfg := config.Default()
	cfg.WorkspaceRoot = t.TempDir()
	cfg.ConfigDir = t.TempDir()
	cfg.MainConfigDir = mainConfigDir
	server := New(cfg, stubDetector{})

	request := httptest.NewRequest(http.MethodGet, "/api/mcp/config/jsonc", nil)
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected fallback status 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"fallback":"go-local-jsonc"`) {
		t.Fatalf("expected go-local-jsonc fallback metadata, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `using local MCP JSONC editor payload`) {
		t.Fatalf("expected local MCP JSONC editor fallback reason, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"path":"`) || !strings.Contains(recorder.Body.String(), `"content":"// tormentnexus MCP configuration`) {
		t.Fatalf("expected local editor payload, got %s", recorder.Body.String())
	}
}

func TestMCPJsoncEditorSaveFallsBackToLocalWrite(t *testing.T) {
	mainConfigDir := t.TempDir()
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	cfg := config.Default()
	cfg.WorkspaceRoot = t.TempDir()
	cfg.ConfigDir = t.TempDir()
	cfg.MainConfigDir = mainConfigDir
	server := New(cfg, stubDetector{})

	request := httptest.NewRequest(http.MethodPost, "/api/mcp/config/jsonc", strings.NewReader(`{"content":"// comment\n{\"mcpServers\":{\"core\":{\"command\":\"node\",\"_meta\":{\"toolCount\":1}}},\"settings\":{\"x\":1}}"}`))
	request.Header.Set("content-type", "application/json")
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected fallback status 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"fallback":"go-local-jsonc"`) {
		t.Fatalf("expected go-local-jsonc fallback metadata, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `saving MCP JSONC through local compatibility writer`) {
		t.Fatalf("expected local MCP JSONC save fallback reason, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"ok":true`) {
		t.Fatalf("expected ok response, got %s", recorder.Body.String())
	}

	jsoncContent, err := os.ReadFile(filepath.Join(mainConfigDir, "mcp.jsonc"))
	if err != nil {
		t.Fatalf("expected local mcp.jsonc to be written: %v", err)
	}
	if !strings.Contains(string(jsoncContent), `"settings"`) || !strings.Contains(string(jsoncContent), `"_meta"`) {
		t.Fatalf("expected jsonc file to preserve settings and _meta, got %s", string(jsoncContent))
	}

	jsonContent, err := os.ReadFile(filepath.Join(mainConfigDir, "mcp.json"))
	if err != nil {
		t.Fatalf("expected compatibility mcp.json to be written: %v", err)
	}
	if strings.Contains(string(jsonContent), `"_meta"`) || strings.Contains(string(jsonContent), `"settings"`) {
		t.Fatalf("expected compatibility json to strip _meta and settings, got %s", string(jsonContent))
	}
}

func TestMCPConfiguredServersFallBackToLocalJsonc(t *testing.T) {
	mainConfigDir := t.TempDir()
	workspaceRoot := t.TempDir()
	jsoncContent := `// tormentnexus MCP configuration
{
  "mcpServers": {
    "core": {
      "_meta": {
        "toolCount": 1
      }
    }
  }
}
`
	if err := os.WriteFile(filepath.Join(mainConfigDir, "mcp.jsonc"), []byte(jsoncContent), 0o644); err != nil {
		t.Fatalf("failed to write local mcp jsonc: %v", err)
	}

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	cfg := config.Default()
	cfg.WorkspaceRoot = workspaceRoot
	cfg.MainConfigDir = mainConfigDir

	db, err := sql.Open("sqlite", filepath.Join(workspaceRoot, "tormentnexus.db"))
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	defer db.Close()
	if _, err := db.Exec(`
		CREATE TABLE mcp_servers (
			uuid TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			type TEXT NOT NULL DEFAULT 'STDIO',
			command TEXT,
			args TEXT NOT NULL DEFAULT '[]',
			env TEXT NOT NULL DEFAULT '{}',
			url TEXT,
			error_status TEXT NOT NULL DEFAULT 'NONE',
			created_at INTEGER NOT NULL,
			bearer_token TEXT,
			headers TEXT NOT NULL DEFAULT '{}',
			always_on INTEGER NOT NULL DEFAULT 0,
			user_id TEXT NOT NULL,
			source_published_server_uuid TEXT
		);
		CREATE TABLE workspace_secrets (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		);
		INSERT INTO workspace_secrets (key, value, created_at, updated_at)
		VALUES ('OPENAI_API_KEY', 'secret-one', 1711958400, 1711958460);
		INSERT INTO mcp_servers (
			uuid, name, description, type, command, args, env, url, error_status, created_at, bearer_token, headers, always_on, user_id, source_published_server_uuid
		) VALUES (
			'srv-db-1', 'core', 'Core server from db', 'STDIO', 'node', '["server.js"]', '{"LOCAL_ONLY":"1"}', NULL, 'NONE', 1711958460, NULL, '{"x-test":"1"}', 1, 'system', NULL
		);
	`); err != nil {
		t.Fatalf("failed to seed sqlite db: %v", err)
	}

	server := New(cfg, stubDetector{})

	listRequest := httptest.NewRequest(http.MethodGet, "/api/mcp/servers/configured", nil)
	listRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(listRecorder, listRequest)

	if listRecorder.Code != http.StatusOK {
		t.Fatalf("expected fallback list status 200, got %d with body %s", listRecorder.Code, listRecorder.Body.String())
	}
	if !strings.Contains(listRecorder.Body.String(), `"fallback":"go-local-mcp-db"`) {
		t.Fatalf("expected go-local-mcp-db fallback metadata, got %s", listRecorder.Body.String())
	}
	if !strings.Contains(listRecorder.Body.String(), `using local MCP server definitions from tormentnexus.db with JSONC metadata overlay`) {
		t.Fatalf("expected configured server list fallback reason, got %s", listRecorder.Body.String())
	}
	if !strings.Contains(listRecorder.Body.String(), `"uuid":"srv-db-1"`) || !strings.Contains(listRecorder.Body.String(), `"command":"node"`) {
		t.Fatalf("expected configured server from local db, got %s", listRecorder.Body.String())
	}
	if !strings.Contains(listRecorder.Body.String(), `"LOCAL_ONLY":"1"`) || !strings.Contains(listRecorder.Body.String(), `"OPENAI_API_KEY":"secret-one"`) {
		t.Fatalf("expected db env plus workspace secret injection, got %s", listRecorder.Body.String())
	}
	if !strings.Contains(listRecorder.Body.String(), `"toolCount":1`) {
		t.Fatalf("expected jsonc _meta overlay, got %s", listRecorder.Body.String())
	}

	expectedUUID := "srv-db-1"
	getRequest := httptest.NewRequest(http.MethodGet, "/api/mcp/servers/get?uuid="+expectedUUID, nil)
	getRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(getRecorder, getRequest)

	if getRecorder.Code != http.StatusOK {
		t.Fatalf("expected fallback get status 200, got %d with body %s", getRecorder.Code, getRecorder.Body.String())
	}
	if !strings.Contains(getRecorder.Body.String(), `using local MCP server definition from tormentnexus.db with JSONC metadata overlay`) {
		t.Fatalf("expected configured server get fallback reason, got %s", getRecorder.Body.String())
	}
	if !strings.Contains(getRecorder.Body.String(), `"uuid":"`+expectedUUID+`"`) {
		t.Fatalf("expected db uuid %s, got %s", expectedUUID, getRecorder.Body.String())
	}
	if !strings.Contains(getRecorder.Body.String(), `"name":"core"`) {
		t.Fatalf("expected configured server get payload, got %s", getRecorder.Body.String())
	}
}

func TestMCPSyncTargetsAndExportFallBackToLocalJsonc(t *testing.T) {
	mainConfigDir := t.TempDir()
	jsoncContent := `// tormentnexus MCP configuration
{
  "mcpServers": {
    "core": {
      "command": "node",
      "args": ["server.js"]
    }
  }
}
`
	if err := os.WriteFile(filepath.Join(mainConfigDir, "mcp.jsonc"), []byte(jsoncContent), 0o644); err != nil {
		t.Fatalf("failed to write local mcp jsonc: %v", err)
	}

	appData := t.TempDir()
	t.Setenv("APPDATA", appData)
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	cfg := config.Default()
	cfg.WorkspaceRoot = t.TempDir()
	cfg.MainConfigDir = mainConfigDir
	server := New(cfg, stubDetector{})

	syncRequest := httptest.NewRequest(http.MethodGet, "/api/mcp/servers/sync-targets", nil)
	syncRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(syncRecorder, syncRequest)

	if syncRecorder.Code != http.StatusOK {
		t.Fatalf("expected sync-targets fallback status 200, got %d with body %s", syncRecorder.Code, syncRecorder.Body.String())
	}
	if !strings.Contains(syncRecorder.Body.String(), `"fallback":"go-local-jsonc"`) {
		t.Fatalf("expected go-local-jsonc fallback metadata, got %s", syncRecorder.Body.String())
	}
	if !strings.Contains(syncRecorder.Body.String(), `using local JSONC sync targets`) {
		t.Fatalf("expected sync-target fallback reason, got %s", syncRecorder.Body.String())
	}
	if !strings.Contains(syncRecorder.Body.String(), `"client":"claude-desktop"`) {
		t.Fatalf("expected claude-desktop sync target, got %s", syncRecorder.Body.String())
	}

	exportRequest := httptest.NewRequest(http.MethodGet, "/api/mcp/servers/export-client-config?client=claude-desktop", nil)
	exportRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(exportRecorder, exportRequest)

	if exportRecorder.Code != http.StatusOK {
		t.Fatalf("expected export fallback status 200, got %d with body %s", exportRecorder.Code, exportRecorder.Body.String())
	}
	if !strings.Contains(exportRecorder.Body.String(), `"procedure":"mcpServers.exportClientConfig"`) {
		t.Fatalf("expected export procedure metadata, got %s", exportRecorder.Body.String())
	}
	if !strings.Contains(exportRecorder.Body.String(), `using local JSONC client config export preview`) {
		t.Fatalf("expected export preview fallback reason, got %s", exportRecorder.Body.String())
	}
	if !strings.Contains(exportRecorder.Body.String(), `"client":"claude-desktop"`) || !strings.Contains(exportRecorder.Body.String(), `"core"`) || !strings.Contains(exportRecorder.Body.String(), `"command":"node"`) {
		t.Fatalf("expected local client config preview, got %s", exportRecorder.Body.String())
	}
}

func TestMCPToolPreferencesFallBackToLocalJsonc(t *testing.T) {
	skipIfServerRunning(t)
	mainConfigDir := t.TempDir()
	jsoncContent := `// tormentnexus MCP configuration
{
  "mcpServers": {},
  "settings": {
    "toolSelection": {
      "importantTools": ["search_tools"],
      "alwaysLoadedTools": ["search_tools"],
      "autoLoadMinConfidence": 0.9,
      "maxLoadedTools": 12,
      "maxHydratedSchemas": 6,
      "idleEvictionThresholdMs": 120000
    }
  }
}
`
	if err := os.WriteFile(filepath.Join(mainConfigDir, "mcp.jsonc"), []byte(jsoncContent), 0o644); err != nil {
		t.Fatalf("failed to write local mcp jsonc: %v", err)
	}

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	cfg := config.Default()
	cfg.WorkspaceRoot = t.TempDir()
	cfg.MainConfigDir = mainConfigDir
	server := New(cfg, stubDetector{})

	getRequest := httptest.NewRequest(http.MethodGet, "/api/mcp/preferences", nil)
	getRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(getRecorder, getRequest)

	if getRecorder.Code != http.StatusOK {
		t.Fatalf("expected get fallback status 200, got %d with body %s", getRecorder.Code, getRecorder.Body.String())
	}
	if !strings.Contains(getRecorder.Body.String(), `"fallback":"go-local-jsonc"`) {
		t.Fatalf("expected go-local-jsonc fallback metadata, got %s", getRecorder.Body.String())
	}
	if !strings.Contains(getRecorder.Body.String(), `using local JSONC tool preferences`) {
		t.Fatalf("expected local preferences fallback reason, got %s", getRecorder.Body.String())
	}
	if !strings.Contains(getRecorder.Body.String(), `"importantTools":["search_tools"]`) || !strings.Contains(getRecorder.Body.String(), `"maxLoadedTools":12`) {
		t.Fatalf("expected local tool preferences payload, got %s", getRecorder.Body.String())
	}

	postRequest := httptest.NewRequest(http.MethodPost, "/api/mcp/preferences", strings.NewReader(`{"importantTools":["search_tools","auto_call_tool"],"maxLoadedTools":20}`))
	postRequest.Header.Set("content-type", "application/json")
	postRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(postRecorder, postRequest)

	if postRecorder.Code != http.StatusOK {
		t.Fatalf("expected post fallback status 200, got %d with body %s", postRecorder.Code, postRecorder.Body.String())
	}
	if !strings.Contains(postRecorder.Body.String(), `saving tool preferences to local JSONC`) {
		t.Fatalf("expected local preference save fallback reason, got %s", postRecorder.Body.String())
	}
	if !strings.Contains(postRecorder.Body.String(), `"ok":true`) {
		t.Fatalf("expected updated tool preferences response, got %s", postRecorder.Body.String())
	}

	jsoncWritten, err := os.ReadFile(filepath.Join(mainConfigDir, "mcp.jsonc"))
	if err != nil {
		t.Fatalf("expected updated jsonc preferences file: %v", err)
	}
	if strings.Contains(postRecorder.Body.String(), `"fallback":"go-local-jsonc"`) && (!strings.Contains(string(jsoncWritten), `"auto_call_tool"`) || !strings.Contains(string(jsoncWritten), `"maxLoadedTools": 20`)) {
		t.Fatalf("expected persisted updated preferences, got %s", string(jsoncWritten))
	}
}

func TestMCPConfiguredServerMutationsFallBackToLocalJsonc(t *testing.T) {
	mainConfigDir := t.TempDir()
	initialConfig := `// tormentnexus MCP configuration
{
  "mcpServers": {
    "core": {
      "command": "node",
      "args": ["server.js"]
    }
  }
}
`
	if err := os.WriteFile(filepath.Join(mainConfigDir, "mcp.jsonc"), []byte(initialConfig), 0o644); err != nil {
		t.Fatalf("failed to seed local mcp jsonc: %v", err)
	}

	appData := t.TempDir()
	t.Setenv("APPDATA", appData)
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	cfg := config.Default()
	cfg.WorkspaceRoot = t.TempDir()
	cfg.MainConfigDir = mainConfigDir
	server := New(cfg, stubDetector{})

	createRequest := httptest.NewRequest(http.MethodPost, "/api/mcp/servers/create", strings.NewReader(`{"name":"alpha","type":"STDIO","command":"python","args":["srv.py"]}`))
	createRequest.Header.Set("content-type", "application/json")
	createRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(createRecorder, createRequest)
	if createRecorder.Code != http.StatusOK || !strings.Contains(createRecorder.Body.String(), `"fallback":"go-local-jsonc"`) || !strings.Contains(createRecorder.Body.String(), `"name":"alpha"`) {
		t.Fatalf("expected create fallback response, got %d %s", createRecorder.Code, createRecorder.Body.String())
	}

	alphaUUID := syntheticServerUUID("alpha")
	updateRequest := httptest.NewRequest(http.MethodPost, "/api/mcp/servers/update", strings.NewReader(`{"uuid":"`+alphaUUID+`","name":"alpha-renamed","command":"python3"}`))
	updateRequest.Header.Set("content-type", "application/json")
	updateRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(updateRecorder, updateRequest)
	if updateRecorder.Code != http.StatusOK || !strings.Contains(updateRecorder.Body.String(), `"name":"alpha-renamed"`) {
		t.Fatalf("expected update fallback response, got %d %s", updateRecorder.Code, updateRecorder.Body.String())
	}

	bulkRequest := httptest.NewRequest(http.MethodPost, "/api/mcp/servers/bulk-import", strings.NewReader(`[{"name":"beta","type":"STDIO","command":"node","args":["beta.js"]}]`))
	bulkRequest.Header.Set("content-type", "application/json")
	bulkRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(bulkRecorder, bulkRequest)
	if bulkRecorder.Code != http.StatusOK || !strings.Contains(bulkRecorder.Body.String(), `"name":"beta"`) {
		t.Fatalf("expected bulk import fallback response, got %d %s", bulkRecorder.Code, bulkRecorder.Body.String())
	}
	if !strings.Contains(bulkRecorder.Body.String(), `using local JSONC bulk import`) {
		t.Fatalf("expected bulk import fallback reason, got %s", bulkRecorder.Body.String())
	}

	syncRequest := httptest.NewRequest(http.MethodPost, "/api/mcp/servers/sync-client-config", strings.NewReader(`{"client":"claude-desktop"}`))
	syncRequest.Header.Set("content-type", "application/json")
	syncRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(syncRecorder, syncRequest)
	if syncRecorder.Code != http.StatusOK || !strings.Contains(syncRecorder.Body.String(), `"written":true`) {
		t.Fatalf("expected sync-client-config fallback response, got %d %s", syncRecorder.Code, syncRecorder.Body.String())
	}

	deleteRequest := httptest.NewRequest(http.MethodPost, "/api/mcp/servers/delete", strings.NewReader(`{"uuid":"`+syntheticServerUUID("beta")+`"}`))
	deleteRequest.Header.Set("content-type", "application/json")
	deleteRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(deleteRecorder, deleteRequest)
	if deleteRecorder.Code != http.StatusOK || !strings.Contains(deleteRecorder.Body.String(), `"ok":true`) {
		t.Fatalf("expected delete fallback response, got %d %s", deleteRecorder.Code, deleteRecorder.Body.String())
	}

	jsoncWritten, err := os.ReadFile(filepath.Join(mainConfigDir, "mcp.jsonc"))
	if err != nil {
		t.Fatalf("expected jsonc file to exist: %v", err)
	}
	if !strings.Contains(string(jsoncWritten), `"alpha-renamed"`) || strings.Contains(string(jsoncWritten), `"beta"`) {
		t.Fatalf("expected persisted create/update/delete state, got %s", string(jsoncWritten))
	}
}

func TestMCPConfiguredServerMetadataMutationsFallBackToLocalJsonc(t *testing.T) {
	mainConfigDir := t.TempDir()
	jsoncContent := `// tormentnexus MCP configuration
{
  "mcpServers": {
    "core": {
      "command": "node",
      "args": ["server.js"],
      "_meta": {
        "status": "ready",
        "toolCount": 2,
        "tools": [{"name":"search"}]
      }
    }
  }
}
`
	if err := os.WriteFile(filepath.Join(mainConfigDir, "mcp.jsonc"), []byte(jsoncContent), 0o644); err != nil {
		t.Fatalf("failed to seed local mcp jsonc: %v", err)
	}

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")
	cfg := config.Default()
	cfg.WorkspaceRoot = t.TempDir()
	cfg.ConfigDir = t.TempDir()
	cfg.MainConfigDir = mainConfigDir
	server := New(cfg, stubDetector{})

	uuid := syntheticServerUUID("core")
	reloadRequest := httptest.NewRequest(http.MethodPost, "/api/mcp/servers/reload-metadata", strings.NewReader(`{"uuid":"`+uuid+`","mode":"binary"}`))
	reloadRequest.Header.Set("content-type", "application/json")
	reloadRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(reloadRecorder, reloadRequest)
	if reloadRecorder.Code != http.StatusOK || !strings.Contains(reloadRecorder.Body.String(), `"reloadDecision":"go-local-placeholder"`) || !strings.Contains(reloadRecorder.Body.String(), `"fallback":"go-local-jsonc"`) {
		t.Fatalf("expected reload metadata fallback response, got %d %s", reloadRecorder.Code, reloadRecorder.Body.String())
	}
	if !strings.Contains(reloadRecorder.Body.String(), `applying local JSONC metadata placeholder fallback`) {
		t.Fatalf("expected reload metadata fallback reason, got %s", reloadRecorder.Body.String())
	}

	clearRequest := httptest.NewRequest(http.MethodPost, "/api/mcp/servers/clear-metadata-cache", strings.NewReader(`{"uuid":"`+uuid+`"}`))
	clearRequest.Header.Set("content-type", "application/json")
	clearRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(clearRecorder, clearRequest)
	if clearRecorder.Code != http.StatusOK || !strings.Contains(clearRecorder.Body.String(), `"toolCount":0`) || !strings.Contains(clearRecorder.Body.String(), `"ok":true`) {
		t.Fatalf("expected clear metadata fallback response, got %d %s", clearRecorder.Code, clearRecorder.Body.String())
	}
	if !strings.Contains(clearRecorder.Body.String(), `applying local JSONC metadata placeholder fallback`) {
		t.Fatalf("expected clear metadata fallback reason, got %s", clearRecorder.Body.String())
	}

	written, err := os.ReadFile(filepath.Join(mainConfigDir, "mcp.jsonc"))
	if err != nil {
		t.Fatalf("expected updated jsonc file: %v", err)
	}
	if !strings.Contains(string(written), `"status": "pending"`) || !strings.Contains(string(written), `"toolCount": 0`) || !strings.Contains(string(written), `Cache cleared at`) {
		t.Fatalf("expected cleared metadata cache state, got %s", string(written))
	}
}

func TestSkillsFallBackToLocalSkillRegistry(t *testing.T) {
	workspaceRoot := t.TempDir()
	skillDir := filepath.Join(workspaceRoot, ".tormentnexus", "skills", "debug")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("failed to create skill dir: %v", err)
	}
	initialSkill := "---\nname: debug\ndescription: Debug help\n---\n\n# Debug\n\nSkill content\n"
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(initialSkill), 0o644); err != nil {
		t.Fatalf("failed to seed skill: %v", err)
	}

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")
	cfg := config.Default()
	cfg.WorkspaceRoot = workspaceRoot
	cfg.ConfigDir = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	listRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(listRecorder, httptest.NewRequest(http.MethodGet, "/api/skills", nil))
	if listRecorder.Code != http.StatusOK || !strings.Contains(listRecorder.Body.String(), `"fallback":"go-local-skills"`) || !strings.Contains(listRecorder.Body.String(), `"name":"debug"`) {
		t.Fatalf("expected local skills list fallback, got %d %s", listRecorder.Code, listRecorder.Body.String())
	}
	if !strings.Contains(listRecorder.Body.String(), `using local skills metadata`) {
		t.Fatalf("expected local skills list fallback reason, got %s", listRecorder.Body.String())
	}

	summaryRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(summaryRecorder, httptest.NewRequest(http.MethodGet, "/api/skills/summary?query=deb", nil))
	if summaryRecorder.Code != http.StatusOK || !strings.Contains(summaryRecorder.Body.String(), `"folder":"debug"`) {
		t.Fatalf("expected local skills summary fallback, got %d %s", summaryRecorder.Code, summaryRecorder.Body.String())
	}
	if !strings.Contains(summaryRecorder.Body.String(), `using local skill folder summaries`) {
		t.Fatalf("expected local skills summary fallback reason, got %s", summaryRecorder.Body.String())
	}

	readRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(readRecorder, httptest.NewRequest(http.MethodGet, "/api/skills/read?name=debug", nil))
	if readRecorder.Code != http.StatusOK || !strings.Contains(readRecorder.Body.String(), `Skill content`) {
		t.Fatalf("expected local skills read fallback, got %d %s", readRecorder.Code, readRecorder.Body.String())
	}
	if !strings.Contains(readRecorder.Body.String(), `using local skill document`) {
		t.Fatalf("expected local skills read fallback reason, got %s", readRecorder.Body.String())
	}

	createRequest := httptest.NewRequest(http.MethodPost, "/api/skills/create", strings.NewReader(`{"id":"trace","name":"trace","description":"Trace help"}`))
	createRequest.Header.Set("content-type", "application/json")
	createRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(createRecorder, createRequest)
	if createRecorder.Code != http.StatusOK || !strings.Contains(createRecorder.Body.String(), `Created skill 'trace'`) {
		t.Fatalf("expected local skills create fallback, got %d %s", createRecorder.Code, createRecorder.Body.String())
	}
	if !strings.Contains(createRecorder.Body.String(), `applying local skill mutation`) {
		t.Fatalf("expected local skills create fallback reason, got %s", createRecorder.Body.String())
	}

	saveRequest := httptest.NewRequest(http.MethodPost, "/api/skills/save", strings.NewReader(`{"id":"trace","content":"Updated content"}`))
	saveRequest.Header.Set("content-type", "application/json")
	saveRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(saveRecorder, saveRequest)
	if saveRecorder.Code != http.StatusOK || !strings.Contains(saveRecorder.Body.String(), `"Saved skill 'trace'."`) {
		t.Fatalf("expected local skills save fallback, got %d %s", saveRecorder.Code, saveRecorder.Body.String())
	}
	if !strings.Contains(saveRecorder.Body.String(), `applying local skill mutation`) {
		t.Fatalf("expected local skills save fallback reason, got %s", saveRecorder.Body.String())
	}

	writtenSkill, err := os.ReadFile(filepath.Join(workspaceRoot, ".tormentnexus", "skills", "trace", "SKILL.md"))
	if err != nil {
		t.Fatalf("expected created skill file: %v", err)
	}
	if string(writtenSkill) != "Updated content" {
		t.Fatalf("expected saved skill content, got %s", string(writtenSkill))
	}
}

func TestImportedSessionBridgeRoutes(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		switch r.URL.Path {
		case "/trpc/session.importedList":
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("failed to read importedList body: %v", err)
			}
			if !strings.Contains(string(body), `"limit":25`) {
				t.Fatalf("expected importedList limit payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"data": map[string]any{
						"json": []map[string]any{
							{"id": "imp-1", "sourceTool": "claude-code", "parsedMemories": []any{}},
						},
					},
				},
			})
		case "/trpc/session.importedGet":
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("failed to read importedGet body: %v", err)
			}
			if !strings.Contains(string(body), `"id":"imp-1"`) {
				t.Fatalf("expected importedGet id payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"data": map[string]any{
						"json": map[string]any{"id": "imp-1", "sourceTool": "claude-code", "parsedMemories": []any{}},
					},
				},
			})
		case "/trpc/session.importedScan":
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("failed to read importedScan body: %v", err)
			}
			if !strings.Contains(string(body), `"force":true`) {
				t.Fatalf("expected importedScan force payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"data": map[string]any{
						"json": map[string]any{
							"discoveredCount":    3,
							"importedCount":      1,
							"skippedCount":       2,
							"storedMemoryCount":  4,
							"instructionDocPath": "C:/tmp/auto-imported-agent-instructions.md",
							"tools":              []string{"claude-code"},
						},
					},
				},
			})
		case "/trpc/session.importedInstructionDocs":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"data": map[string]any{
						"json": []map[string]any{
							{"path": "C:/tmp/auto-imported-agent-instructions.md", "name": "auto-imported-agent-instructions.md"},
						},
					},
				},
			})
		default:
			t.Fatalf("unexpected upstream path %s", r.URL.Path)
		}
	}))
	defer upstream.Close()

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", upstream.URL+"/trpc")

	cfg := config.Default()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	cases := []struct {
		name      string
		method    string
		path      string
		body      string
		contains  string
		procedure string
	}{
		{name: "imported list", method: http.MethodGet, path: "/api/sessions/imported/list?limit=25", contains: "\"sourceTool\":\"claude-code\"", procedure: "\"procedure\":\"session.importedList\""},
		{name: "imported get", method: http.MethodGet, path: "/api/sessions/imported/get?id=imp-1", contains: "\"id\":\"imp-1\"", procedure: "\"procedure\":\"session.importedGet\""},
		{name: "imported scan", method: http.MethodPost, path: "/api/sessions/imported/scan", body: `{"force":true}`, contains: "\"storedMemoryCount\":4", procedure: "\"procedure\":\"session.importedScan\""},
		{name: "imported docs", method: http.MethodGet, path: "/api/sessions/imported/instruction-docs", contains: "\"auto-imported-agent-instructions.md\"", procedure: "\"procedure\":\"session.importedInstructionDocs\""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var body io.Reader
			if tc.body != "" {
				body = strings.NewReader(tc.body)
			}
			request := httptest.NewRequest(tc.method, tc.path, body)
			if tc.body != "" {
				request.Header.Set("content-type", "application/json")
			}
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, request)
			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.contains) {
				t.Fatalf("expected response to contain %s, got %s", tc.contains, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.procedure) {
				t.Fatalf("expected bridge metadata %s, got %s", tc.procedure, recorder.Body.String())
			}
		})
	}
}

func TestImportedSessionScanFallsBackToGoScanner(t *testing.T) {
	workspaceRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(workspaceRoot, ".claude"), 0o755); err != nil {
		t.Fatalf("failed to create claude root: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workspaceRoot, ".claude", "session.jsonl"), []byte("{\"model\":\"claude\"}\n"), 0o644); err != nil {
		t.Fatalf("failed to seed claude session: %v", err)
	}

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")
	t.Setenv("HOME", workspaceRoot)
	t.Setenv("USERPROFILE", workspaceRoot)
	t.Setenv("APPDATA", workspaceRoot)
	t.Setenv("LOCALAPPDATA", workspaceRoot)

	cfg := config.Default()
	cfg.WorkspaceRoot = workspaceRoot
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	request := httptest.NewRequest(http.MethodPost, "/api/sessions/imported/scan", strings.NewReader(`{"force":true}`))
	request.Header.Set("content-type", "application/json")
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected fallback status 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"fallback":"go-sessionimport"`) {
		t.Fatalf("expected go-sessionimport fallback metadata, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `using scan-only imported session discovery summary`) {
		t.Fatalf("expected scan-only scan fallback reason, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"instructionDocPath":null`) {
		t.Fatalf("expected no instruction doc path in raw scan fallback, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"discoveredCount":`) {
		t.Fatalf("expected discoveredCount in fallback summary, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"claude-code"`) {
		t.Fatalf("expected claude-code tool in fallback summary, got %s", recorder.Body.String())
	}
}

func TestImportedSessionScanFallsBackToArchivedRecords(t *testing.T) {
	workspaceRoot := t.TempDir()
	seedArchivedImportedSession(t, workspaceRoot)
	if err := os.MkdirAll(filepath.Join(workspaceRoot, ".copilot", "session-state"), 0o755); err != nil {
		t.Fatalf("failed to create copilot session dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workspaceRoot, ".copilot", "session-state", "checkpoint-history.json"), []byte(`{"history":[]}`), 0o644); err != nil {
		t.Fatalf("failed to seed extra discovered session: %v", err)
	}

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")
	t.Setenv("HOME", workspaceRoot)
	t.Setenv("USERPROFILE", workspaceRoot)
	t.Setenv("APPDATA", workspaceRoot)
	t.Setenv("LOCALAPPDATA", workspaceRoot)

	cfg := config.Default()
	cfg.WorkspaceRoot = workspaceRoot
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	request := httptest.NewRequest(http.MethodPost, "/api/sessions/imported/scan", strings.NewReader(`{"force":true}`))
	request.Header.Set("content-type", "application/json")
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected archived fallback status 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"procedure":"session.importedScan"`) {
		t.Fatalf("expected imported scan fallback metadata, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `merged persisted imported sessions with validated scan candidates`) {
		t.Fatalf("expected merged scan fallback reason, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"importedCount":1`) {
		t.Fatalf("expected archived importedCount=1, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"discoveredCount":2`) {
		t.Fatalf("expected merged discoveredCount=2, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"storedMemoryCount":2`) {
		t.Fatalf("expected archived storedMemoryCount=2, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"copilot-cli"`) {
		t.Fatalf("expected merged tool list to include discovery candidates, got %s", recorder.Body.String())
	}
}

func TestImportedSessionScanFallsBackToWorkspaceInstructionDoc(t *testing.T) {
	workspaceRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(workspaceRoot, ".claude"), 0o755); err != nil {
		t.Fatalf("failed to create claude root: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workspaceRoot, ".claude", "session.jsonl"), []byte("{\"model\":\"claude\"}\n"), 0o644); err != nil {
		t.Fatalf("failed to seed claude session: %v", err)
	}

	cfg := config.Default()
	cfg.WorkspaceRoot = workspaceRoot
	cfg.MainConfigDir = t.TempDir()
	docPath := cfg.ImportedInstructionsPath()
	if err := os.MkdirAll(filepath.Dir(docPath), 0o755); err != nil {
		t.Fatalf("failed to create imported instructions directory: %v", err)
	}
	if err := os.WriteFile(docPath, []byte("# Auto-imported Agent Instructions\n"), 0o644); err != nil {
		t.Fatalf("failed to write imported instructions doc: %v", err)
	}

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")
	t.Setenv("HOME", workspaceRoot)
	t.Setenv("USERPROFILE", workspaceRoot)
	t.Setenv("APPDATA", workspaceRoot)
	t.Setenv("LOCALAPPDATA", workspaceRoot)
	server := New(cfg, stubDetector{})

	request := httptest.NewRequest(http.MethodPost, "/api/sessions/imported/scan", strings.NewReader(`{"force":true}`))
	request.Header.Set("content-type", "application/json")
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected fallback status 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"instructionDocPath":"`) {
		t.Fatalf("expected workspace instruction doc path in fallback summary, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `auto-imported-agent-instructions.md`) {
		t.Fatalf("expected workspace instruction doc filename in fallback summary, got %s", recorder.Body.String())
	}
}

func TestImportedSessionListFallsBackToGoScanner(t *testing.T) {
	workspaceRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(workspaceRoot, ".claude"), 0o755); err != nil {
		t.Fatalf("failed to create claude root: %v", err)
	}
	sessionPath := filepath.Join(workspaceRoot, ".claude", "session.jsonl")
	if err := os.WriteFile(sessionPath, []byte("{\"model\":\"claude\"}\n"), 0o644); err != nil {
		t.Fatalf("failed to seed claude session: %v", err)
	}

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	cfg := config.Default()
	cfg.WorkspaceRoot = workspaceRoot
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	request := httptest.NewRequest(http.MethodGet, "/api/sessions/imported/list?limit=5", nil)
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected fallback status 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"fallback":"go-sessionimport"`) {
		t.Fatalf("expected go-sessionimport fallback metadata, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `using scan-only imported session records`) {
		t.Fatalf("expected scan-only list fallback reason, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"sourceTool":"`) {
		t.Fatalf("expected at least one fallback sourceTool entry, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"parsedMemories":[]`) {
		t.Fatalf("expected empty parsed memories in fallback entry, got %s", recorder.Body.String())
	}
}

func writeGzipFile(t *testing.T, path string, contents []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("failed to create gzip parent dir: %v", err)
	}
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)
	if _, err := writer.Write(contents); err != nil {
		t.Fatalf("failed to write gzip contents: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close gzip writer: %v", err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatalf("failed to persist gzip file: %v", err)
	}
}

func seedArchivedImportedSession(t *testing.T, workspaceRoot string) string {
	t.Helper()
	archiveDir := filepath.Join(workspaceRoot, ".tormentnexus", "imported_sessions", "archive", "ab", "cd")
	transcriptHash := "abcd1234ef567890abcd1234ef567890"
	sessionID := "imp-archived-1"
	metadataPath := filepath.Join(archiveDir, transcriptHash+".meta.json.gz")
	transcriptPath := filepath.Join(archiveDir, transcriptHash+".txt.gz")

	writeGzipFile(t, transcriptPath, []byte("Archived imported transcript\nimportant context"))
	sidecar := map[string]any{
		"sessionId":               sessionID,
		"sourceTool":              "claude-code",
		"sourcePath":              filepath.Join(workspaceRoot, ".claude", "session.jsonl"),
		"sessionFormat":           "jsonl",
		"transcriptHash":          transcriptHash,
		"title":                   "Archived Session",
		"workingDirectory":        workspaceRoot,
		"transcriptLength":        43,
		"excerpt":                 "Archived imported transcript",
		"durableMemoryCount":      2,
		"durableInstructionCount": 1,
		"memoryTags":              []string{"go", "archive"},
		"retentionSummary": map[string]any{
			"archiveDisposition": "archive_only",
		},
		"archivedAt": int64(1711933200000),
	}
	payload, err := json.Marshal(sidecar)
	if err != nil {
		t.Fatalf("failed to marshal sidecar: %v", err)
	}
	writeGzipFile(t, metadataPath, payload)
	return sessionID
}

func TestImportedSessionListFallsBackToArchivedRecords(t *testing.T) {
	workspaceRoot := t.TempDir()
	sessionID := seedArchivedImportedSession(t, workspaceRoot)

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	cfg := config.Default()
	cfg.WorkspaceRoot = workspaceRoot
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	request := httptest.NewRequest(http.MethodGet, "/api/sessions/imported/list?limit=5", nil)
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected fallback status 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `using archived imported session records`) {
		t.Fatalf("expected archived list fallback reason, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"id":"`+sessionID+`"`) {
		t.Fatalf("expected archived imported session id %s, got %s", sessionID, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"archiveFormat":"gzip-text-v1"`) {
		t.Fatalf("expected archived metadata marker, got %s", recorder.Body.String())
	}
}

func TestImportedSessionGetFallsBackToGoScanner(t *testing.T) {
	workspaceRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(workspaceRoot, ".claude"), 0o755); err != nil {
		t.Fatalf("failed to create claude root: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workspaceRoot, ".claude", "session.jsonl"), []byte("{\"model\":\"claude\"}\n"), 0o644); err != nil {
		t.Fatalf("failed to seed claude session: %v", err)
	}

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	cfg := config.Default()
	cfg.WorkspaceRoot = workspaceRoot
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	candidates, err := server.scanValidatedImportSources()
	if err != nil {
		t.Fatalf("failed to scan validated import sources: %v", err)
	}
	records := server.importedSessionFallbackRecords(candidates)
	if len(records) == 0 {
		t.Fatalf("expected at least one fallback imported session record")
	}

	request := httptest.NewRequest(http.MethodGet, "/api/sessions/imported/get?id="+records[0].ID, nil)
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected fallback status 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"fallback":"go-sessionimport"`) {
		t.Fatalf("expected go-sessionimport fallback metadata, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `using scan-only imported session record`) {
		t.Fatalf("expected scan-only get fallback reason, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"id":"`+records[0].ID+`"`) {
		t.Fatalf("expected fallback imported session id %s, got %s", records[0].ID, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"sourceTool":"`+records[0].SourceTool+`"`) {
		t.Fatalf("expected sourceTool %s in fallback entry, got %s", records[0].SourceTool, recorder.Body.String())
	}
}

func TestImportedSessionGetFallsBackToArchivedRecords(t *testing.T) {
	workspaceRoot := t.TempDir()
	sessionID := seedArchivedImportedSession(t, workspaceRoot)

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	cfg := config.Default()
	cfg.WorkspaceRoot = workspaceRoot
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	request := httptest.NewRequest(http.MethodGet, "/api/sessions/imported/get?id="+sessionID, nil)
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected fallback status 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `using archived imported session record`) {
		t.Fatalf("expected archived get fallback reason, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"id":"`+sessionID+`"`) {
		t.Fatalf("expected archived imported session id %s, got %s", sessionID, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `Archived imported transcript`) {
		t.Fatalf("expected archived transcript contents, got %s", recorder.Body.String())
	}
}

func TestImportedSessionGetReturnsUnavailableWhenFallbackRecordMissing(t *testing.T) {
	workspaceRoot := t.TempDir()

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	cfg := config.Default()
	cfg.WorkspaceRoot = workspaceRoot
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	request := httptest.NewRequest(http.MethodGet, "/api/sessions/imported/get?id=missing-imported-session", nil)
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected unavailable status 503, got %d with body %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"success":false`) {
		t.Fatalf("expected unsuccessful response body, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"error":"imported session unavailable"`) {
		t.Fatalf("expected unavailable error message, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"procedure":"session.importedGet"`) {
		t.Fatalf("expected importedGet procedure metadata, got %s", recorder.Body.String())
	}
}

func TestImportedInstructionDocsFallsBackToEmptyList(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	cfg := config.Default()
	cfg.WorkspaceRoot = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	request := httptest.NewRequest(http.MethodGet, "/api/sessions/imported/instruction-docs", nil)
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected fallback status 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"fallback":"go-sessionimport"`) {
		t.Fatalf("expected go-sessionimport fallback metadata, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `using workspace imported instruction documents`) {
		t.Fatalf("expected workspace docs fallback reason, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"data":[]`) {
		t.Fatalf("expected empty instruction doc fallback list, got %s", recorder.Body.String())
	}
}

func TestImportedInstructionDocsFallsBackToWorkspaceDoc(t *testing.T) {
	workspaceRoot := t.TempDir()
	cfg := config.Default()
	cfg.WorkspaceRoot = workspaceRoot
	cfg.MainConfigDir = t.TempDir()

	docPath := cfg.ImportedInstructionsPath()
	if err := os.MkdirAll(filepath.Dir(docPath), 0o755); err != nil {
		t.Fatalf("failed to create imported instructions directory: %v", err)
	}
	if err := os.WriteFile(docPath, []byte("# Auto-imported Agent Instructions\n"), 0o644); err != nil {
		t.Fatalf("failed to write imported instructions doc: %v", err)
	}

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")
	server := New(cfg, stubDetector{})

	request := httptest.NewRequest(http.MethodGet, "/api/sessions/imported/instruction-docs", nil)
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected fallback status 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `using workspace imported instruction documents`) {
		t.Fatalf("expected workspace docs fallback reason, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"auto-imported-agent-instructions.md"`) {
		t.Fatalf("expected workspace imported instructions doc, got %s", recorder.Body.String())
	}
}

func TestImportedSessionMaintenanceStatsFallsBackToGoScanner(t *testing.T) {
	workspaceRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(workspaceRoot, ".claude"), 0o755); err != nil {
		t.Fatalf("failed to create claude root: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workspaceRoot, ".claude", "session.jsonl"), []byte("{\"model\":\"claude\"}\n"), 0o644); err != nil {
		t.Fatalf("failed to seed claude session: %v", err)
	}

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")
	t.Setenv("HOME", workspaceRoot)
	t.Setenv("USERPROFILE", workspaceRoot)
	t.Setenv("APPDATA", workspaceRoot)
	t.Setenv("LOCALAPPDATA", workspaceRoot)

	cfg := config.Default()
	cfg.WorkspaceRoot = workspaceRoot
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	request := httptest.NewRequest(http.MethodGet, "/api/sessions/imported/maintenance-stats", nil)
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected fallback status 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"fallback":"go-sessionimport"`) {
		t.Fatalf("expected go-sessionimport fallback metadata, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `using scan-only imported session maintenance stats`) {
		t.Fatalf("expected scan-only maintenance fallback reason, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"inlineTranscriptCount":0`) {
		t.Fatalf("expected scan-only fallback to report zero inline transcripts, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"missingRetentionSummaryCount":0`) {
		t.Fatalf("expected scan-only fallback to report zero missing retention summaries, got %s", recorder.Body.String())
	}
}

func TestStartupImportedSessionMaintenanceStatsUsesScanOnlySemantics(t *testing.T) {
	workspaceRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(workspaceRoot, ".claude"), 0o755); err != nil {
		t.Fatalf("failed to create claude root: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workspaceRoot, ".claude", "session.jsonl"), []byte("{\"model\":\"claude\"}\n"), 0o644); err != nil {
		t.Fatalf("failed to seed claude session: %v", err)
	}

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")
	t.Setenv("HOME", workspaceRoot)
	t.Setenv("USERPROFILE", workspaceRoot)
	t.Setenv("APPDATA", workspaceRoot)
	t.Setenv("LOCALAPPDATA", workspaceRoot)

	cfg := config.Default()
	cfg.WorkspaceRoot = workspaceRoot
	cfg.ConfigDir = filepath.Join(workspaceRoot, ".tormentnexus-go")
	cfg.MainConfigDir = filepath.Join(workspaceRoot, ".tormentnexus")
	if err := os.MkdirAll(cfg.ConfigDir, 0o755); err != nil {
		t.Fatalf("failed to create go config dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(workspaceRoot, ".tormentnexus", "memory"), 0o755); err != nil {
		t.Fatalf("failed to create memory dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workspaceRoot, ".tormentnexus", "memory", "claude_mem.json"), []byte(`{"default":[]}`), 0o644); err != nil {
		t.Fatalf("failed to seed memory store: %v", err)
	}

	server := New(cfg, stubDetector{})
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/startup/status", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected startup status 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"inlineTranscriptCount":0`) {
		t.Fatalf("expected startup imported sessions to report zero inline transcripts for scan-only fallback, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"missingRetentionSummaryCount":0`) {
		t.Fatalf("expected startup imported sessions to report zero missing retention summaries for scan-only fallback, got %s", recorder.Body.String())
	}
}

func TestImportedSessionMaintenanceStatsFallsBackToArchivedRecords(t *testing.T) {
	workspaceRoot := t.TempDir()
	seedArchivedImportedSession(t, workspaceRoot)

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	cfg := config.Default()
	cfg.WorkspaceRoot = workspaceRoot
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	request := httptest.NewRequest(http.MethodGet, "/api/sessions/imported/maintenance-stats", nil)
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected fallback status 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `using archived imported session maintenance stats`) {
		t.Fatalf("expected archived maintenance fallback reason, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"archivedTranscriptCount":1`) {
		t.Fatalf("expected archived transcript fallback count, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"inlineTranscriptCount":0`) {
		t.Fatalf("expected zero inline transcript fallback count, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"missingRetentionSummaryCount":0`) {
		t.Fatalf("expected retention summary to be present, got %s", recorder.Body.String())
	}
}

func TestMemoryBridgeRoutes(t *testing.T) {
	t.Skip("Skipping test for now")
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		switch r.URL.Path {
		case "/trpc/memory.query":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"query":"bootstrap"`) {
				t.Fatalf("expected memory.query payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []map[string]any{{"id": "ctx-1", "title": "Bootstrap Context"}}}}})
		case "/trpc/memory.listContexts":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []map[string]any{{"id": "ctx-1", "title": "Bootstrap Context"}}}}})
		case "/trpc/memory.getContext":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"id":"ctx-1"`) {
				t.Fatalf("expected memory.getContext id payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"id": "ctx-1", "content": "hello"}}}})
		case "/trpc/memory.deleteContext":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}}})
		case "/trpc/memory.getAgentStats":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"session": 1, "working": 2, "longTerm": 3, "total": 6}}}})
		case "/trpc/memory.searchAgentMemory":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"query":"memory"`) {
				t.Fatalf("expected memory.searchAgentMemory payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []map[string]any{{"id": "mem-1", "content": "memory result"}}}}})
		case "/trpc/memory.saveContext":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"source":"docs"`) || !strings.Contains(string(body), `"url":"https://example.com"`) {
				t.Fatalf("expected memory.saveContext payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true, "id": "ctx-2"}}}})
		case "/trpc/memory.addFact":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"content":"remember this"`) {
				t.Fatalf("expected memory.addFact payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}}})
		case "/trpc/memory.recordObservation":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"content":"Observation"`) {
				t.Fatalf("expected memory.recordObservation payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true, "memory": map[string]any{"id": "obs-1"}}}}})
		case "/trpc/memory.captureUserPrompt":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"content":"Need help"`) {
				t.Fatalf("expected memory.captureUserPrompt payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true, "memory": map[string]any{"id": "prompt-1"}}}}})
		case "/trpc/memory.getRecentObservations":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"limit":5`) {
				t.Fatalf("expected memory.getRecentObservations payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []map[string]any{{"id": "obs-1", "content": "Observation"}}}}})
		case "/trpc/memory.searchObservations":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"query":"signal"`) {
				t.Fatalf("expected memory.searchObservations payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []map[string]any{{"id": "obs-2", "content": "signal"}}}}})
		case "/trpc/memory.getRecentUserPrompts":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"limit":4`) {
				t.Fatalf("expected memory.getRecentUserPrompts payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []map[string]any{{"id": "prompt-1", "content": "Need help"}}}}})
		case "/trpc/memory.searchUserPrompts":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"query":"help"`) {
				t.Fatalf("expected memory.searchUserPrompts payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []map[string]any{{"id": "prompt-2", "content": "help"}}}}})
		case "/trpc/memory.searchMemoryPivot":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"pivotMemoryId":"mem-1"`) {
				t.Fatalf("expected memory.searchMemoryPivot payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []map[string]any{{"id": "pivot-1", "score": 0.9}}}}})
		case "/trpc/memory.getMemoryTimelineWindow":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"centerMemoryId":"mem-1"`) {
				t.Fatalf("expected memory.getMemoryTimelineWindow payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []map[string]any{{"id": "timeline-1"}}}}})
		case "/trpc/memory.getCrossSessionMemoryLinks":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"memoryId":"mem-1"`) {
				t.Fatalf("expected memory.getCrossSessionMemoryLinks payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []map[string]any{{"sessionId": "sess-2", "memoryId": "mem-9"}}}}})
		case "/trpc/memory.captureSessionSummary":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"sessionId":"sess-1"`) {
				t.Fatalf("expected memory.captureSessionSummary payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true, "memory": map[string]any{"id": "summary-1"}}}}})
		case "/trpc/memory.getSessionBootstrap":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"activeGoal": "ship parity", "memories": []any{}}}}})
		case "/trpc/memory.getToolContext":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"toolName":"search_tools"`) {
				t.Fatalf("expected memory.getToolContext tool payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"toolName": "search_tools", "context": []any{}}}}})
		case "/trpc/memory.getRecentSessionSummaries":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []map[string]any{{"sessionId": "sess-1", "summary": "recent"}}}}})
		case "/trpc/memory.searchSessionSummaries":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"query":"recent"`) {
				t.Fatalf("expected memory.searchSessionSummaries payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []map[string]any{{"sessionId": "sess-1", "summary": "recent"}}}}})
		case "/trpc/memory.listInterchangeFormats":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []string{"json", "markdown"}}}})
		case "/trpc/memory.exportMemories":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"data": "{}", "format": "json"}}}})
		case "/trpc/memory.importMemories":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"imported": 2}}}})
		case "/trpc/memory.convertMemories":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"data": "---", "toFormat": "markdown"}}}})
		case "/trpc/memory.getSectionedMemoryStatus":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"available": true, "sections": []any{}}}}})
		default:
			t.Fatalf("unexpected upstream path %s", r.URL.Path)
		}
	}))
	defer upstream.Close()

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", upstream.URL+"/trpc")

	cfg := config.Default()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	cases := []struct {
		name      string
		method    string
		path      string
		body      string
		contains  string
		procedure string
	}{
		{name: "memory search", method: http.MethodGet, path: "/api/memory/search?query=bootstrap&limit=3", contains: "\"Bootstrap Context\"", procedure: "\"procedure\":\"memory.query\""},
		{name: "memory contexts", method: http.MethodGet, path: "/api/memory/contexts", contains: "\"ctx-1\"", procedure: "\"procedure\":\"memory.listContexts\""},
		{name: "memory save context", method: http.MethodPost, path: "/api/memory/context/save", body: `{"source":"docs","url":"https://example.com","content":"hello world"}`, contains: "\"id\":\"ctx-2\"", procedure: "\"procedure\":\"memory.saveContext\""},
		{name: "memory get context", method: http.MethodGet, path: "/api/memory/context/get?id=ctx-1", contains: "\"content\":\"hello\"", procedure: "\"procedure\":\"memory.getContext\""},
		{name: "memory delete context", method: http.MethodPost, path: "/api/memory/context/delete", body: `{"id":"ctx-1"}`, contains: "\"success\":true", procedure: "\"procedure\":\"memory.deleteContext\""},
		{name: "memory agent stats", method: http.MethodGet, path: "/api/memory/agent-stats", contains: "\"total\":6", procedure: "\"procedure\":\"memory.getAgentStats\""},
		{name: "memory agent search", method: http.MethodGet, path: "/api/memory/agent-search?query=memory&type=working&limit=5", contains: "\"memory result\"", procedure: "\"procedure\":\"memory.searchAgentMemory\""},
		{name: "memory add fact", method: http.MethodPost, path: "/api/memory/facts/add", body: `{"content":"remember this","type":"working"}`, contains: "\"success\":true", procedure: "\"procedure\":\"memory.addFact\""},
		{name: "memory record observation", method: http.MethodPost, path: "/api/memory/observations/record", body: `{"content":"Observation","type":"fact","namespace":"ops"}`, contains: "\"obs-1\"", procedure: "\"procedure\":\"memory.recordObservation\""},
		{name: "memory recent observations", method: http.MethodGet, path: "/api/memory/observations/recent?limit=5&namespace=ops&type=fact", contains: "\"Observation\"", procedure: "\"procedure\":\"memory.getRecentObservations\""},
		{name: "memory search observations", method: http.MethodGet, path: "/api/memory/observations/search?query=signal&limit=5&namespace=ops&type=fact", contains: "\"signal\"", procedure: "\"procedure\":\"memory.searchObservations\""},
		{name: "memory capture prompt", method: http.MethodPost, path: "/api/memory/user-prompts/capture", body: `{"content":"Need help","role":"user"}`, contains: "\"prompt-1\"", procedure: "\"procedure\":\"memory.captureUserPrompt\""},
		{name: "memory recent prompts", method: http.MethodGet, path: "/api/memory/user-prompts/recent?limit=4&role=user", contains: "\"Need help\"", procedure: "\"procedure\":\"memory.getRecentUserPrompts\""},
		{name: "memory search prompts", method: http.MethodGet, path: "/api/memory/user-prompts/search?query=help&limit=4&role=user", contains: "\"help\"", procedure: "\"procedure\":\"memory.searchUserPrompts\""},
		{name: "memory pivot search", method: http.MethodPost, path: "/api/memory/pivot/search", body: `{"pivotMemoryId":"mem-1","limit":5}`, contains: "\"pivot-1\"", procedure: "\"procedure\":\"memory.searchMemoryPivot\""},
		{name: "memory timeline window", method: http.MethodPost, path: "/api/memory/timeline/window", body: `{"centerMemoryId":"mem-1","before":2,"after":2}`, contains: "\"timeline-1\"", procedure: "\"procedure\":\"memory.getMemoryTimelineWindow\""},
		{name: "memory cross session links", method: http.MethodPost, path: "/api/memory/cross-session-links", body: `{"memoryId":"mem-1","limit":5}`, contains: "\"sess-2\"", procedure: "\"procedure\":\"memory.getCrossSessionMemoryLinks\""},
		{name: "memory session bootstrap", method: http.MethodGet, path: "/api/memory/session-bootstrap?activeGoal=ship%20parity", contains: "\"activeGoal\":\"ship parity\"", procedure: "\"procedure\":\"memory.getSessionBootstrap\""},
		{name: "memory tool context", method: http.MethodGet, path: "/api/memory/tool-context?toolName=search_tools", contains: "\"toolName\":\"search_tools\"", procedure: "\"procedure\":\"memory.getToolContext\""},
		{name: "capture session summary", method: http.MethodPost, path: "/api/memory/session-summaries/capture", body: `{"sessionId":"sess-1","status":"stopped"}`, contains: "\"summary-1\"", procedure: "\"procedure\":\"memory.captureSessionSummary\""},
		{name: "recent session summaries", method: http.MethodGet, path: "/api/memory/session-summaries/recent?limit=5", contains: "\"sessionId\":\"sess-1\"", procedure: "\"procedure\":\"memory.getRecentSessionSummaries\""},
		{name: "search session summaries", method: http.MethodGet, path: "/api/memory/session-summaries/search?query=recent&limit=5", contains: "\"summary\":\"recent\"", procedure: "\"procedure\":\"memory.searchSessionSummaries\""},
		{name: "memory sectioned status", method: http.MethodGet, path: "/api/memory/sectioned-status", contains: "\"available\":true", procedure: "\"procedure\":\"memory.getSectionedMemoryStatus\""},
		{name: "memory interchange formats", method: http.MethodGet, path: "/api/memory/interchange-formats", contains: "\"markdown\"", procedure: "\"procedure\":\"memory.listInterchangeFormats\""},
		{name: "memory export", method: http.MethodGet, path: "/api/memory/export?userId=default&format=json", contains: "\"format\":\"json\"", procedure: "\"procedure\":\"memory.exportMemories\""},
		{name: "memory import", method: http.MethodPost, path: "/api/memory/import", body: `{"userId":"default","format":"json","data":"{}"}`, contains: "\"imported\":2", procedure: "\"procedure\":\"memory.importMemories\""},
		{name: "memory convert", method: http.MethodPost, path: "/api/memory/convert", body: `{"userId":"default","fromFormat":"json","toFormat":"markdown","data":"{}"}`, contains: "\"toFormat\":\"markdown\"", procedure: "\"procedure\":\"memory.convertMemories\""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var body io.Reader
			if tc.body != "" {
				body = strings.NewReader(tc.body)
			}
			request := httptest.NewRequest(tc.method, tc.path, body)
			if tc.body != "" {
				request.Header.Set("content-type", "application/json")
			}
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, request)
			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.contains) {
				t.Fatalf("expected response to contain %s, got %s", tc.contains, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.procedure) {
				t.Fatalf("expected bridge metadata %s, got %s", tc.procedure, recorder.Body.String())
			}
		})
	}
}

func TestAgentMemoryBridgeRoutes(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		switch r.URL.Path {
		case "/trpc/agentMemory.search":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"query":"bridge"`) {
				t.Fatalf("expected agentMemory.search payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []map[string]any{{"id": "am-1", "content": "bridge result"}}}}})
		case "/trpc/agentMemory.add":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"id": "am-2", "content": "added memory"}}}})
		case "/trpc/agentMemory.getRecent":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []map[string]any{{"id": "am-3", "type": "working"}}}}})
		case "/trpc/agentMemory.getByType":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"type":"working"`) {
				t.Fatalf("expected agentMemory.getByType payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []map[string]any{{"id": "am-4", "type": "working"}}}}})
		case "/trpc/agentMemory.getByNamespace":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"namespace":"project"`) {
				t.Fatalf("expected agentMemory.getByNamespace payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []map[string]any{{"id": "am-5", "namespace": "project"}}}}})
		case "/trpc/agentMemory.delete":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": true}}})
		case "/trpc/agentMemory.clearSession":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}}})
		case "/trpc/agentMemory.export":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"session": []any{}, "working": []any{}, "long_term": []any{}}}}})
		case "/trpc/agentMemory.handoff":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"artifact": "handoff.md"}}}})
		case "/trpc/agentMemory.pickup":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"restored": 3}}}})
		case "/trpc/agentMemory.stats":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"session": 2, "working": 3, "longTerm": 4, "total": 9}}}})
		default:
			t.Fatalf("unexpected upstream path %s", r.URL.Path)
		}
	}))
	defer upstream.Close()

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", upstream.URL+"/trpc")

	cfg := config.Default()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	cases := []struct {
		name      string
		method    string
		path      string
		body      string
		contains  string
		procedure string
	}{
		{name: "agent memory search", method: http.MethodGet, path: "/api/agent-memory/search?query=bridge&namespace=project&type=working&limit=5", contains: "\"bridge result\"", procedure: "\"procedure\":\"agentMemory.search\""},
		{name: "agent memory add", method: http.MethodPost, path: "/api/agent-memory/add", body: `{"content":"remember this","type":"working","namespace":"project","tags":["bridge"]}`, contains: "\"added memory\"", procedure: "\"procedure\":\"agentMemory.add\""},
		{name: "agent memory recent", method: http.MethodGet, path: "/api/agent-memory/recent?type=working&limit=3", contains: "\"type\":\"working\"", procedure: "\"procedure\":\"agentMemory.getRecent\""},
		{name: "agent memory by type", method: http.MethodGet, path: "/api/agent-memory/by-type?type=working", contains: "\"am-4\"", procedure: "\"procedure\":\"agentMemory.getByType\""},
		{name: "agent memory by namespace", method: http.MethodGet, path: "/api/agent-memory/by-namespace?namespace=project", contains: "\"namespace\":\"project\"", procedure: "\"procedure\":\"agentMemory.getByNamespace\""},
		{name: "agent memory delete", method: http.MethodPost, path: "/api/agent-memory/delete", body: `{"id":"am-1"}`, contains: "\"data\":true", procedure: "\"procedure\":\"agentMemory.delete\""},
		{name: "agent memory clear session", method: http.MethodPost, path: "/api/agent-memory/clear-session", contains: "\"success\":true", procedure: "\"procedure\":\"agentMemory.clearSession\""},
		{name: "agent memory export", method: http.MethodGet, path: "/api/agent-memory/export", contains: "\"long_term\":[]", procedure: "\"procedure\":\"agentMemory.export\""},
		{name: "agent memory handoff", method: http.MethodPost, path: "/api/agent-memory/handoff", body: `{"notes":"bridge handoff"}`, contains: "\"artifact\":\"handoff.md\"", procedure: "\"procedure\":\"agentMemory.handoff\""},
		{name: "agent memory pickup", method: http.MethodPost, path: "/api/agent-memory/pickup", body: `{"artifact":"handoff.md"}`, contains: "\"restored\":3", procedure: "\"procedure\":\"agentMemory.pickup\""},
		{name: "agent memory stats", method: http.MethodGet, path: "/api/agent-memory/stats", contains: "\"total\":9", procedure: "\"procedure\":\"agentMemory.stats\""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var body io.Reader
			if tc.body != "" {
				body = strings.NewReader(tc.body)
			}
			request := httptest.NewRequest(tc.method, tc.path, body)
			if tc.body != "" {
				request.Header.Set("content-type", "application/json")
			}
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, request)
			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.contains) {
				t.Fatalf("expected response to contain %s, got %s", tc.contains, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.procedure) {
				t.Fatalf("expected bridge metadata %s, got %s", tc.procedure, recorder.Body.String())
			}
		})
	}
}

func TestCodeBridgeRoutes(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		switch r.URL.Path {
		case "/trpc/graph.get":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"nodes": []map[string]any{{"id": "file-a"}}, "links": []any{}, "dependencies": map[string]any{"file-a": []string{"file-b"}}}}}})
		case "/trpc/graph.rebuild":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true, "nodes": []map[string]any{{"id": "file-b"}}}}}})
		case "/trpc/graph.getConsumers":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"filePath":"src/app.ts"`) {
				t.Fatalf("expected graph.getConsumers payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []string{"src/consumer.ts"}}}})
		case "/trpc/graph.getDependencies":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"filePath":"src/app.ts"`) {
				t.Fatalf("expected graph.getDependencies payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []string{"src/dep.ts"}}}})
		case "/trpc/graph.getSymbolsGraph":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"nodes": []map[string]any{{"id": "sym-1", "name": "demo"}}, "links": []map[string]any{{"source": "src/demo.ts", "target": "sym-1", "type": "defines"}}}}}})
		case "/trpc/tormentnexusContext.list":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []string{"src/app.ts"}}}})
		case "/trpc/tormentnexusContext.add":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": "added src/app.ts"}}})
		case "/trpc/tormentnexusContext.remove":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": "removed src/app.ts"}}})
		case "/trpc/tormentnexusContext.clear":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": "cleared"}}})
		case "/trpc/tormentnexusContext.getPrompt":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": "Prompt context"}}})
		case "/trpc/git.getModules":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []map[string]any{{"name": "tormentnexus", "path": "submodules/tormentnexus"}}}}})
		case "/trpc/git.getLog":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"limit":5`) {
				t.Fatalf("expected git.getLog limit payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []map[string]any{{"hash": "abc123", "message": "ship bridge"}}}}})
		case "/trpc/git.getStatus":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"branch": "main", "clean": false}}}})
		case "/trpc/git.revert":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}}})
		case "/trpc/tests.status":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"isRunning": true, "results": map[string]any{"src/app.ts": map[string]any{"status": "passed"}}}}}})
		case "/trpc/tests.start":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}}})
		case "/trpc/tests.stop":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}}})
		case "/trpc/tests.run":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"filePath":"src/app.ts"`) {
				t.Fatalf("expected tests.run payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true, "testFile": "src/app.test.ts"}}}})
		case "/trpc/tests.results":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []map[string]any{{"file": "src/app.test.ts", "status": "passed"}}}}})
		default:
			t.Fatalf("unexpected upstream path %s", r.URL.Path)
		}
	}))
	defer upstream.Close()

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", upstream.URL+"/trpc")

	cfg := config.Default()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	cases := []struct {
		name      string
		method    string
		path      string
		body      string
		contains  string
		procedure string
	}{
		{name: "graph get", method: http.MethodGet, path: "/api/graph", contains: "\"file-a\"", procedure: "\"procedure\":\"graph.get\""},
		{name: "graph rebuild", method: http.MethodPost, path: "/api/graph/rebuild", contains: "\"success\":true", procedure: "\"procedure\":\"graph.rebuild\""},
		{name: "graph consumers", method: http.MethodGet, path: "/api/graph/consumers?filePath=src/app.ts", contains: "\"src/consumer.ts\"", procedure: "\"procedure\":\"graph.getConsumers\""},
		{name: "graph dependencies", method: http.MethodGet, path: "/api/graph/dependencies?filePath=src/app.ts", contains: "\"src/dep.ts\"", procedure: "\"procedure\":\"graph.getDependencies\""},
		{name: "graph symbols", method: http.MethodGet, path: "/api/graph/symbols", contains: "\"sym-1\"", procedure: "\"procedure\":\"graph.getSymbolsGraph\""},
		{name: "context list", method: http.MethodGet, path: "/api/context/list", contains: "\"src/app.ts\"", procedure: "\"procedure\":\"tormentnexusContext.list\""},
		{name: "context add", method: http.MethodPost, path: "/api/context/add", body: "{\"filePath\":\"src/app.ts\"}", contains: "added src/app.ts", procedure: "\"procedure\":\"tormentnexusContext.add\""},
		{name: "context remove", method: http.MethodPost, path: "/api/context/remove", body: "{\"filePath\":\"src/app.ts\"}", contains: "removed src/app.ts", procedure: "\"procedure\":\"tormentnexusContext.remove\""},
		{name: "context clear", method: http.MethodPost, path: "/api/context/clear", contains: "cleared", procedure: "\"procedure\":\"tormentnexusContext.clear\""},
		{name: "context prompt", method: http.MethodGet, path: "/api/context/prompt", contains: "Prompt context", procedure: "\"procedure\":\"tormentnexusContext.getPrompt\""},
		{name: "git modules", method: http.MethodGet, path: "/api/git/modules", contains: "\"tormentnexus\"", procedure: "\"procedure\":\"git.getModules\""},
		{name: "context list", method: http.MethodGet, path: "/api/context/list", contains: "\"src/app.ts\"", procedure: "\"procedure\":\"tormentnexusContext.list\""},
		{name: "context add", method: http.MethodPost, path: "/api/context/add", body: "{\"filePath\":\"src/app.ts\"}", contains: "added src/app.ts", procedure: "\"procedure\":\"tormentnexusContext.add\""},
		{name: "context remove", method: http.MethodPost, path: "/api/context/remove", body: "{\"filePath\":\"src/app.ts\"}", contains: "removed src/app.ts", procedure: "\"procedure\":\"tormentnexusContext.remove\""},
		{name: "context clear", method: http.MethodPost, path: "/api/context/clear", contains: "cleared", procedure: "\"procedure\":\"tormentnexusContext.clear\""},
		{name: "context prompt", method: http.MethodGet, path: "/api/context/prompt", contains: "Prompt context", procedure: "\"procedure\":\"tormentnexusContext.getPrompt\""},
		{name: "git modules", method: http.MethodGet, path: "/api/git/modules", contains: "\"tormentnexus\"", procedure: "\"procedure\":\"git.getModules\""},
		{name: "git log", method: http.MethodGet, path: "/api/git/log?limit=5", contains: "\"abc123\"", procedure: "\"procedure\":\"git.getLog\""},
		{name: "git status", method: http.MethodGet, path: "/api/git/status", contains: "\"branch\":\"main\"", procedure: "\"procedure\":\"git.getStatus\""},
		{name: "git revert", method: http.MethodPost, path: "/api/git/revert", body: "{\"hash\":\"abc123\"}", contains: "\"success\":true", procedure: "\"procedure\":\"git.revert\""},
		{name: "tests status", method: http.MethodGet, path: "/api/tests/status", contains: "\"isRunning\":true", procedure: "\"procedure\":\"tests.status\""},
		{name: "tests start", method: http.MethodPost, path: "/api/tests/start", contains: "\"success\":true", procedure: "\"procedure\":\"tests.start\""},
		{name: "tests stop", method: http.MethodPost, path: "/api/tests/stop", contains: "\"success\":true", procedure: "\"procedure\":\"tests.stop\""},
		{name: "tests run", method: http.MethodPost, path: "/api/tests/run", body: "{\"filePath\":\"src/app.ts\"}", contains: "\"testFile\":\"src/app.test.ts\"", procedure: "\"procedure\":\"tests.run\""},
		{name: "tests results", method: http.MethodGet, path: "/api/tests/results", contains: "\"src/app.test.ts\"", procedure: "\"procedure\":\"tests.results\""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var body io.Reader
			if tc.body != "" {
				body = strings.NewReader(tc.body)
			}
			request := httptest.NewRequest(tc.method, tc.path, body)
			if tc.body != "" {
				request.Header.Set("content-type", "application/json")
			}
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, request)
			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.contains) {
				t.Fatalf("expected response to contain %s, got %s", tc.contains, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.procedure) {
				t.Fatalf("expected bridge metadata %s, got %s", tc.procedure, recorder.Body.String())
			}
		})
	}
}

func TestAdminBridgeRoutes(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		switch r.URL.Path {
		case "/trpc/metrics.getStats":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"windowMs":60000`) {
				t.Fatalf("expected metrics.getStats payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"counts": map[string]any{"requests": 5}}}}})
		case "/trpc/metrics.track":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}}})
		case "/trpc/metrics.systemSnapshot":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"process": map[string]any{"pid": 1234}, "system": map[string]any{"platform": "win32"}}}}})
		case "/trpc/metrics.getTimeline":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"metricType":"requests"`) {
				t.Fatalf("expected metrics.getTimeline payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"series": []any{map[string]any{"timestamp": 1, "value": 2}}, "buckets": 10}}}})
		case "/trpc/metrics.getProviderBreakdown":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"totalCost": 1.25, "providers": []any{map[string]any{"provider": "openai", "requests": 3}}}}}})
		case "/trpc/metrics.toggleMonitoring":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true, "monitoring": true}}}})
		case "/trpc/metrics.getRoutingHistory":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"limit":7`) {
				t.Fatalf("expected metrics.getRoutingHistory payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []map[string]any{{"provider": "openai", "reason": "quota"}}}}})
		case "/trpc/logs.list":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"sessionId":"sess-1"`) || !strings.Contains(string(body), `"serverName":"core"`) {
				t.Fatalf("expected logs.list payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []map[string]any{{"toolName": "search_tools", "level": "info"}}}}})
		case "/trpc/logs.summary":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"limit":50`) {
				t.Fatalf("expected logs.summary payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"totals": map[string]any{"totalCalls": 3}, "topTools": []map[string]any{{"name": "search_tools", "count": 2}}}}}})
		case "/trpc/logs.clear":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true, "message": "Logs cleared"}}}})
		case "/trpc/serverHealth.check":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"serverUuid":"srv-1"`) {
				t.Fatalf("expected serverHealth.check payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"status": "HEALTHY", "crashCount": 0, "maxAttempts": 3}}}})
		case "/trpc/serverHealth.reset":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}}})
		default:
			t.Fatalf("unexpected upstream path %s", r.URL.Path)
		}
	}))
	defer upstream.Close()

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", upstream.URL+"/trpc")

	cfg := config.Default()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	cases := []struct {
		name      string
		method    string
		path      string
		body      string
		contains  string
		procedure string
	}{
		{name: "metrics stats", method: http.MethodGet, path: "/api/metrics/stats?windowMs=60000", contains: `"requests":5`, procedure: `"procedure":"metrics.getStats"`},
		{name: "metrics track", method: http.MethodPost, path: "/api/metrics/track", body: `{"type":"requests","value":1,"tags":{"provider":"openai"}}`, contains: `"success":true`, procedure: `"procedure":"metrics.track"`},
		{name: "metrics system snapshot", method: http.MethodGet, path: "/api/metrics/system-snapshot", contains: `"platform":"win32"`, procedure: `"procedure":"metrics.systemSnapshot"`},
		{name: "metrics timeline", method: http.MethodGet, path: "/api/metrics/timeline?windowMs=60000&buckets=10&metricType=requests", contains: `"buckets":10`, procedure: `"procedure":"metrics.getTimeline"`},
		{name: "metrics provider breakdown", method: http.MethodGet, path: "/api/metrics/provider-breakdown", contains: `"totalCost":1.25`, procedure: `"procedure":"metrics.getProviderBreakdown"`},
		{name: "metrics monitoring", method: http.MethodPost, path: "/api/metrics/monitoring", body: `{"enabled":true,"intervalMs":5000}`, contains: `"monitoring":true`, procedure: `"procedure":"metrics.toggleMonitoring"`},
		{name: "metrics routing history", method: http.MethodGet, path: "/api/metrics/routing-history?limit=7", contains: `"reason":"quota"`, procedure: `"procedure":"metrics.getRoutingHistory"`},
		{name: "logs list", method: http.MethodGet, path: "/api/logs?limit=10&sessionId=sess-1&serverName=core", contains: `"search_tools"`, procedure: `"procedure":"logs.list"`},
		{name: "logs summary", method: http.MethodGet, path: "/api/logs/summary?limit=50", contains: `"totalCalls":3`, procedure: `"procedure":"logs.summary"`},
		{name: "logs clear", method: http.MethodPost, path: "/api/logs/clear", contains: `"Logs cleared"`, procedure: `"procedure":"logs.clear"`},
		{name: "server health check", method: http.MethodGet, path: "/api/server-health/check?serverUuid=srv-1", contains: `"maxAttempts":3`, procedure: `"procedure":"serverHealth.check"`},
		{name: "server health reset", method: http.MethodPost, path: "/api/server-health/reset", body: `{"serverUuid":"srv-1"}`, contains: `"success":true`, procedure: `"procedure":"serverHealth.reset"`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var body io.Reader
			if tc.body != "" {
				body = strings.NewReader(tc.body)
			}
			request := httptest.NewRequest(tc.method, tc.path, body)
			if tc.body != "" {
				request.Header.Set("content-type", "application/json")
			}
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, request)
			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.contains) {
				t.Fatalf("expected response to contain %s, got %s", tc.contains, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.procedure) {
				t.Fatalf("expected bridge metadata %s, got %s", tc.procedure, recorder.Body.String())
			}
		})
	}
}

func TestControlBridgeRoutes(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		switch r.URL.Path {
		case "/trpc/settings.get":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"theme": "dark"}}}})
		case "/trpc/settings.update":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}}})
		case "/trpc/settings.getProviders":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"id": "openai", "configured": true}}}}})
		case "/trpc/settings.testConnection":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true, "provider": "openai"}}}})
		case "/trpc/settings.getEnvironment":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"platform": "win32"}}}})
		case "/trpc/settings.getMcpServers":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"name": "core"}}}}})
		case "/trpc/settings.updateProviderKey":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true, "updatedKey": "OPENAI_API_KEY"}}}})
		case "/trpc/tools.list":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"name": "search_tools"}}}}})
		case "/trpc/tools.listByServer":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"mcpServerUuid":"srv-1"`) {
				t.Fatalf("expected tools.listByServer payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"server": "core"}}}}})
		case "/trpc/tools.search":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"query":"search"`) {
				t.Fatalf("expected tools.search payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"name": "search_tools"}}}}})
		case "/trpc/tools.detectCliHarnesses":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"id": "tormentnexus"}}}}})
		case "/trpc/tools.detectExecutionEnvironment":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"shell": "pwsh"}}}})
		case "/trpc/tools.detectInstallSurfaces":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"surface": "npm-global"}}}}})
		case "/trpc/tools.get":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"uuid":"search_tools"`) {
				t.Fatalf("expected tools.get payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"name": "search_tools"}}}})
		case "/trpc/tools.create":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}}})
		case "/trpc/tools.upsertBatch":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}}})
		case "/trpc/tools.delete":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}}})
		case "/trpc/tools.setAlwaysOn":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true, "tool": map[string]any{"always_on": true}}}}})
		case "/trpc/toolSets.list":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"uuid": "ts-1"}}}}})
		case "/trpc/toolSets.get":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"uuid":"ts-1"`) {
				t.Fatalf("expected toolSets.get payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"uuid": "ts-1", "name": "Core Tools"}}}})
		case "/trpc/toolSets.create":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"uuid": "ts-2"}}}})
		case "/trpc/toolSets.update":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"uuid": "ts-1", "name": "Updated"}}}})
		case "/trpc/toolSets.delete":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}}})
		case "/trpc/project.getContext":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": "# Project Context"}}})
		case "/trpc/project.updateContext":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}}})
		case "/trpc/project.getHandoffs":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"id": "handoff_1.json"}}}}})
		case "/trpc/shell.logCommand":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true, "id": "cmd-1"}}}})
		case "/trpc/shell.queryHistory":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"query":"pnpm"`) {
				t.Fatalf("expected shell.queryHistory payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"command": "pnpm test"}}}}})
		case "/trpc/shell.getSystemHistory":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"limit":5`) {
				t.Fatalf("expected shell.getSystemHistory payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"command": "git status"}}}}})
		default:
			t.Fatalf("unexpected upstream path %s", r.URL.Path)
		}
	}))
	defer upstream.Close()

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", upstream.URL+"/trpc")

	cfg := config.Default()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	cases := []struct {
		name      string
		method    string
		path      string
		body      string
		contains  string
		procedure string
	}{
		{name: "settings get", method: http.MethodGet, path: "/api/settings", contains: `"theme":"dark"`, procedure: `"procedure":"settings.get"`},
		{name: "settings update", method: http.MethodPost, path: "/api/settings/update", body: `{"config":{"theme":"dark"}}`, contains: `"success":true`, procedure: `"procedure":"settings.update"`},
		{name: "settings providers", method: http.MethodGet, path: "/api/settings/providers", contains: `"configured":true`, procedure: `"procedure":"settings.getProviders"`},
		{name: "settings test connection", method: http.MethodPost, path: "/api/settings/test-connection", body: `{"provider":"openai"}`, contains: `"provider":"openai"`, procedure: `"procedure":"settings.testConnection"`},
		{name: "settings environment", method: http.MethodGet, path: "/api/settings/environment", contains: `"platform":"win32"`, procedure: `"procedure":"settings.getEnvironment"`},
		{name: "settings mcp servers", method: http.MethodGet, path: "/api/settings/mcp-servers", contains: `"name":"core"`, procedure: `"procedure":"settings.getMcpServers"`},
		{name: "settings provider key", method: http.MethodPost, path: "/api/settings/provider-key", body: `{"provider":"openai","key":"sk-test"}`, contains: `"updatedKey":"OPENAI_API_KEY"`, procedure: `"procedure":"settings.updateProviderKey"`},
		{name: "tools list", method: http.MethodGet, path: "/api/tools", contains: `"search_tools"`, procedure: `"procedure":"tools.list"`},
		{name: "tools by server", method: http.MethodGet, path: "/api/tools/by-server?mcpServerUuid=srv-1", contains: `"core"`, procedure: `"procedure":"tools.listByServer"`},
		{name: "tools search", method: http.MethodGet, path: "/api/tools/search?query=search&limit=5", contains: `"search_tools"`, procedure: `"procedure":"tools.search"`},
		{name: "tools detect cli harnesses", method: http.MethodGet, path: "/api/tools/detect-cli-harnesses", contains: `"tormentnexus"`, procedure: `"procedure":"tools.detectCliHarnesses"`},
		{name: "tools detect execution environment", method: http.MethodGet, path: "/api/tools/detect-execution-environment", contains: `"shell":"pwsh"`, procedure: `"procedure":"tools.detectExecutionEnvironment"`},
		{name: "tools detect install surfaces", method: http.MethodGet, path: "/api/tools/detect-install-surfaces", contains: `"npm-global"`, procedure: `"procedure":"tools.detectInstallSurfaces"`},
		{name: "tools get", method: http.MethodGet, path: "/api/tools/get?uuid=search_tools", contains: `"name":"search_tools"`, procedure: `"procedure":"tools.get"`},
		{name: "tools create", method: http.MethodPost, path: "/api/tools/create", body: `{"name":"demo","mcp_server_uuid":"srv-1","description":"Demo","toolSchema":{"type":"object"}}`, contains: `"success":true`, procedure: `"procedure":"tools.create"`},
		{name: "tools upsert batch", method: http.MethodPost, path: "/api/tools/upsert-batch", body: `[{"name":"demo","mcp_server_uuid":"srv-1","description":"Demo","toolSchema":{"type":"object"}}]`, contains: `"success":true`, procedure: `"procedure":"tools.upsertBatch"`},
		{name: "tools delete", method: http.MethodPost, path: "/api/tools/delete", body: `{"uuid":"search_tools"}`, contains: `"success":true`, procedure: `"procedure":"tools.delete"`},
		{name: "tools always on", method: http.MethodPost, path: "/api/tools/always-on", body: `{"uuid":"search_tools","alwaysOn":true}`, contains: `"always_on":true`, procedure: `"procedure":"tools.setAlwaysOn"`},
		{name: "tool sets list", method: http.MethodGet, path: "/api/tool-sets", contains: `"uuid":"ts-1"`, procedure: `"procedure":"toolSets.list"`},
		{name: "tool sets get", method: http.MethodGet, path: "/api/tool-sets/get?uuid=ts-1", contains: `"Core Tools"`, procedure: `"procedure":"toolSets.get"`},
		{name: "tool sets create", method: http.MethodPost, path: "/api/tool-sets/create", body: `{"name":"Core Tools","tools":["search_tools"]}`, contains: `"uuid":"ts-2"`, procedure: `"procedure":"toolSets.create"`},
		{name: "tool sets update", method: http.MethodPost, path: "/api/tool-sets/update", body: `{"uuid":"ts-1","name":"Updated"}`, contains: `"Updated"`, procedure: `"procedure":"toolSets.update"`},
		{name: "tool sets delete", method: http.MethodPost, path: "/api/tool-sets/delete", body: `{"uuid":"ts-1"}`, contains: `"success":true`, procedure: `"procedure":"toolSets.delete"`},
		{name: "project context", method: http.MethodGet, path: "/api/project/context", contains: `# Project Context`, procedure: `"procedure":"project.getContext"`},
		{name: "project context update", method: http.MethodPost, path: "/api/project/context/update", body: `{"content":"# Project Context"}`, contains: `"success":true`, procedure: `"procedure":"project.updateContext"`},
		{name: "project handoffs", method: http.MethodGet, path: "/api/project/handoffs", contains: `"handoff_1.json"`, procedure: `"procedure":"project.getHandoffs"`},
		{name: "shell log", method: http.MethodPost, path: "/api/shell/log", body: `{"command":"pnpm test","cwd":"C:\\repo","session":"sess-1"}`, contains: `"id":"cmd-1"`, procedure: `"procedure":"shell.logCommand"`},
		{name: "shell query history", method: http.MethodGet, path: "/api/shell/history/query?query=pnpm&limit=5", contains: `"pnpm test"`, procedure: `"procedure":"shell.queryHistory"`},
		{name: "shell system history", method: http.MethodGet, path: "/api/shell/history/system?limit=5", contains: `"git status"`, procedure: `"procedure":"shell.getSystemHistory"`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var body io.Reader
			if tc.body != "" {
				body = strings.NewReader(tc.body)
			}
			request := httptest.NewRequest(tc.method, tc.path, body)
			if tc.body != "" {
				request.Header.Set("content-type", "application/json")
			}
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, request)
			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.contains) {
				t.Fatalf("expected response to contain %s, got %s", tc.contains, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.procedure) {
				t.Fatalf("expected bridge metadata %s, got %s", tc.procedure, recorder.Body.String())
			}
		})
	}
}

func TestAgentBridgeRoutes(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		switch r.URL.Path {
		case "/trpc/agent.runTool":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"ok": true, "content": "tool output"}}}})
		case "/trpc/agent.chat":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"response": "hello", "degraded": false}}}})
		case "/trpc/commands.execute":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"input":"/status"`) {
				t.Fatalf("expected commands.execute payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"handled": true, "output": "ok"}}}})
		case "/trpc/commands.list":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"name": "status", "description": "Show status"}}}}})
		case "/trpc/skills.list":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"id": "skill-1", "name": "debug", "description": "Debug help", "content": "Use rg", "path": "skills/debug/SKILL.md"}}}}})
		case "/trpc/skills.read":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"name":"debug"`) {
				t.Fatalf("expected skills.read payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"content": []any{map[string]any{"type": "text", "text": "Skill content"}}}}}})
		case "/trpc/skills.create":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"content": []any{map[string]any{"type": "text", "text": "Created"}}}}}})
		case "/trpc/skills.save":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"content": []any{map[string]any{"type": "text", "text": "Saved"}}}}}})
		case "/trpc/skills.assimilate":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true, "logs": []any{"assimilated"}}}}})
		default:
			t.Fatalf("unexpected upstream path %s", r.URL.Path)
		}
	}))
	defer upstream.Close()

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", upstream.URL+"/trpc")

	cfg := config.Default()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	cases := []struct {
		name      string
		method    string
		path      string
		body      string
		contains  string
		procedure string
	}{
		{name: "agent run tool", method: http.MethodPost, path: "/api/agent/tool", body: `{"toolName":"search_tools","arguments":{"query":"tormentnexus"}}`, contains: `"tool output"`, procedure: `"procedure":"agent.runTool"`},
		{name: "agent chat", method: http.MethodPost, path: "/api/agent/chat", body: `{"message":"hello"}`, contains: `"response":"hello"`, procedure: `"procedure":"agent.chat"`},
		{name: "commands execute", method: http.MethodPost, path: "/api/commands/execute", body: `{"input":"/status"}`, contains: `"handled":true`, procedure: `"procedure":"commands.execute"`},
		{name: "commands list", method: http.MethodGet, path: "/api/commands", contains: `"name":"status"`, procedure: `"procedure":"commands.list"`},
		{name: "skills list", method: http.MethodGet, path: "/api/skills", contains: `"skill-1"`, procedure: `"procedure":"skills.list"`},
		{name: "skills summary", method: http.MethodGet, path: "/api/skills/summary?query=deb", contains: `"folder":"debug"`, procedure: `"procedure":"skills.list"`},
		{name: "skills read", method: http.MethodGet, path: "/api/skills/read?name=debug", contains: `"Skill content"`, procedure: `"procedure":"skills.read"`},
		{name: "skills create", method: http.MethodPost, path: "/api/skills/create", body: `{"id":"skill-2","name":"trace","description":"Trace help"}`, contains: `"Created"`, procedure: `"procedure":"skills.create"`},
		{name: "skills save", method: http.MethodPost, path: "/api/skills/save", body: `{"id":"skill-2","content":"Updated content"}`, contains: `"Saved"`, procedure: `"procedure":"skills.save"`},
		{name: "skills assimilate", method: http.MethodPost, path: "/api/skills/assimilate", body: `{"topic":"debugging","docsUrl":"https://example.com"}`, contains: `"assimilated"`, procedure: `"procedure":"skills.assimilate"`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var body io.Reader
			if tc.body != "" {
				body = strings.NewReader(tc.body)
			}
			request := httptest.NewRequest(tc.method, tc.path, body)
			if tc.body != "" {
				request.Header.Set("content-type", "application/json")
			}
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, request)
			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.contains) {
				t.Fatalf("expected response to contain %s, got %s", tc.contains, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.procedure) {
				t.Fatalf("expected bridge metadata %s, got %s", tc.procedure, recorder.Body.String())
			}
		})
	}
}

func TestWorkflowBridgeRoutes(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		switch r.URL.Path {
		case "/trpc/workflow.list":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"id": "wf-1", "name": "Workflow One"}}}}})
		case "/trpc/workflow.getGraph":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"workflowId":"wf-1"`) {
				t.Fatalf("expected workflowId in getGraph payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"nodes": []any{map[string]any{"id": "n1"}}, "edges": []any{}}}}})
		case "/trpc/workflow.start":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"workflowId":"wf-1"`) {
				t.Fatalf("expected workflow.start payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"executionId": "exec-1", "status": "running"}}}})
		case "/trpc/workflow.listExecutions":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"executionId": "exec-1", "status": "running"}}}}})
		case "/trpc/workflow.getExecution":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"executionId":"exec-1"`) {
				t.Fatalf("expected workflow.getExecution payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"executionId": "exec-1", "status": "running"}}}})
		case "/trpc/workflow.getHistory":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"executionId":"exec-1"`) {
				t.Fatalf("expected workflow.getHistory payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"event": "started"}}}}})
		case "/trpc/workflow.resume":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}}})
		case "/trpc/workflow.pause":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}}})
		case "/trpc/workflow.approve":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}}})
		case "/trpc/workflow.reject":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"reason":"needs changes"`) {
				t.Fatalf("expected workflow.reject payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}}})
		case "/trpc/workflow.listCanvases":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"id": "canvas-1", "name": "Canvas One"}}}}})
		case "/trpc/workflow.loadCanvas":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"id":"canvas-1"`) {
				t.Fatalf("expected workflow.loadCanvas payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"id": "canvas-1", "nodes": []any{}, "edges": []any{}}}}})
		case "/trpc/workflow.saveCanvas":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"name":"Canvas One"`) {
				t.Fatalf("expected workflow.saveCanvas payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"id": "canvas-1"}}}})
		default:
			t.Fatalf("unexpected upstream path %s", r.URL.Path)
		}
	}))
	defer upstream.Close()

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", upstream.URL+"/trpc")

	cfg := config.Default()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	cases := []struct {
		name      string
		method    string
		path      string
		body      string
		contains  string
		procedure string
	}{
		{name: "workflow list", method: http.MethodGet, path: "/api/workflows", contains: `"wf-1"`, procedure: `"procedure":"workflow.list"`},
		{name: "workflow graph", method: http.MethodGet, path: "/api/workflows/graph?workflowId=wf-1", contains: `"nodes"`, procedure: `"procedure":"workflow.getGraph"`},
		{name: "workflow start", method: http.MethodPost, path: "/api/workflows/start", body: `{"workflowId":"wf-1","initialState":{"ticket":"123"}}`, contains: `"executionId":"exec-1"`, procedure: `"procedure":"workflow.start"`},
		{name: "workflow executions", method: http.MethodGet, path: "/api/workflows/executions", contains: `"status":"running"`, procedure: `"procedure":"workflow.listExecutions"`},
		{name: "workflow execution", method: http.MethodGet, path: "/api/workflows/execution?executionId=exec-1", contains: `"executionId":"exec-1"`, procedure: `"procedure":"workflow.getExecution"`},
		{name: "workflow history", method: http.MethodGet, path: "/api/workflows/history?executionId=exec-1", contains: `"started"`, procedure: `"procedure":"workflow.getHistory"`},
		{name: "workflow resume", method: http.MethodPost, path: "/api/workflows/resume", body: `{"executionId":"exec-1"}`, contains: `"success":true`, procedure: `"procedure":"workflow.resume"`},
		{name: "workflow pause", method: http.MethodPost, path: "/api/workflows/pause", body: `{"executionId":"exec-1"}`, contains: `"success":true`, procedure: `"procedure":"workflow.pause"`},
		{name: "workflow approve", method: http.MethodPost, path: "/api/workflows/approve", body: `{"executionId":"exec-1"}`, contains: `"success":true`, procedure: `"procedure":"workflow.approve"`},
		{name: "workflow reject", method: http.MethodPost, path: "/api/workflows/reject", body: `{"executionId":"exec-1","reason":"needs changes"}`, contains: `"success":true`, procedure: `"procedure":"workflow.reject"`},
		{name: "workflow canvases", method: http.MethodGet, path: "/api/workflows/canvases", contains: `"canvas-1"`, procedure: `"procedure":"workflow.listCanvases"`},
		{name: "workflow canvas", method: http.MethodGet, path: "/api/workflows/canvas?id=canvas-1", contains: `"canvas-1"`, procedure: `"procedure":"workflow.loadCanvas"`},
		{name: "workflow canvas save", method: http.MethodPost, path: "/api/workflows/canvas/save", body: `{"name":"Canvas One","description":"desc","nodes":[],"edges":[]}`, contains: `"id":"canvas-1"`, procedure: `"procedure":"workflow.saveCanvas"`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var body io.Reader
			if tc.body != "" {
				body = strings.NewReader(tc.body)
			}
			request := httptest.NewRequest(tc.method, tc.path, body)
			if tc.body != "" {
				request.Header.Set("content-type", "application/json")
			}
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, request)
			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.contains) {
				t.Fatalf("expected response to contain %s, got %s", tc.contains, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.procedure) {
				t.Fatalf("expected bridge metadata %s, got %s", tc.procedure, recorder.Body.String())
			}
		})
	}
}

func TestSymbolsAndLSPBridgeRoutes(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		switch r.URL.Path {
		case "/trpc/symbols.list":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"id": "sym-1", "name": "RunServer"}}}}})
		case "/trpc/symbols.find":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"query":"Run"`) {
				t.Fatalf("expected symbols.find payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"name": "RunServer"}}}}})
		case "/trpc/symbols.pin":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"id": "sym-1"}}}})
		case "/trpc/symbols.unpin":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": true}}})
		case "/trpc/symbols.updatePriority":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": true}}})
		case "/trpc/symbols.addNotes":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": true}}})
		case "/trpc/symbols.clear":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": 2}}})
		case "/trpc/symbols.forFile":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"filePath":"src/app.ts"`) {
				t.Fatalf("expected symbols.forFile payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"id": "sym-1", "file": "src/app.ts"}}}}})
		case "/trpc/lsp.findSymbol":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"symbolName":"RunServer"`) {
				t.Fatalf("expected lsp.findSymbol payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"filePath": "src/app.ts", "line": 10}}}})
		case "/trpc/lsp.findReferences":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"character":3`) {
				t.Fatalf("expected lsp.findReferences payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"filePath": "src/app.ts", "line": 10}}}}})
		case "/trpc/lsp.getSymbols":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"filePath":"src/app.ts"`) {
				t.Fatalf("expected lsp.getSymbols payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"name": "RunServer"}}}}})
		case "/trpc/lsp.searchSymbols":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"query":"Run"`) {
				t.Fatalf("expected lsp.searchSymbols payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"name": "RunServer"}}}}})
		case "/trpc/lsp.indexProject":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}}})
		default:
			t.Fatalf("unexpected upstream path %s", r.URL.Path)
		}
	}))
	defer upstream.Close()

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", upstream.URL+"/trpc")

	cfg := config.Default()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	cases := []struct {
		name      string
		method    string
		path      string
		body      string
		contains  string
		procedure string
	}{
		{name: "symbols list", method: http.MethodGet, path: "/api/symbols", contains: `"sym-1"`, procedure: `"procedure":"symbols.list"`},
		{name: "symbols find", method: http.MethodGet, path: "/api/symbols/find?query=Run&limit=5", contains: `"RunServer"`, procedure: `"procedure":"symbols.find"`},
		{name: "symbols pin", method: http.MethodPost, path: "/api/symbols/pin", body: `{"name":"RunServer","file":"src/app.ts","type":"function"}`, contains: `"id":"sym-1"`, procedure: `"procedure":"symbols.pin"`},
		{name: "symbols unpin", method: http.MethodPost, path: "/api/symbols/unpin", body: `{"id":"sym-1"}`, contains: `"data":true`, procedure: `"procedure":"symbols.unpin"`},
		{name: "symbols priority", method: http.MethodPost, path: "/api/symbols/priority", body: `{"id":"sym-1","priority":7}`, contains: `"data":true`, procedure: `"procedure":"symbols.updatePriority"`},
		{name: "symbols notes", method: http.MethodPost, path: "/api/symbols/notes", body: `{"id":"sym-1","notes":"important"}`, contains: `"data":true`, procedure: `"procedure":"symbols.addNotes"`},
		{name: "symbols clear", method: http.MethodPost, path: "/api/symbols/clear", body: `{}`, contains: `"data":2`, procedure: `"procedure":"symbols.clear"`},
		{name: "symbols for file", method: http.MethodGet, path: "/api/symbols/file?filePath=src%2Fapp.ts", contains: `"src/app.ts"`, procedure: `"procedure":"symbols.forFile"`},
		{name: "lsp find symbol", method: http.MethodGet, path: "/api/lsp/find-symbol?filePath=src%2Fapp.ts&symbolName=RunServer", contains: `"line":10`, procedure: `"procedure":"lsp.findSymbol"`},
		{name: "lsp references", method: http.MethodGet, path: "/api/lsp/find-references?filePath=src%2Fapp.ts&line=10&character=3", contains: `"src/app.ts"`, procedure: `"procedure":"lsp.findReferences"`},
		{name: "lsp symbols", method: http.MethodGet, path: "/api/lsp/symbols?filePath=src%2Fapp.ts", contains: `"RunServer"`, procedure: `"procedure":"lsp.getSymbols"`},
		{name: "lsp search", method: http.MethodGet, path: "/api/lsp/search?query=Run", contains: `"RunServer"`, procedure: `"procedure":"lsp.searchSymbols"`},
		{name: "lsp index", method: http.MethodPost, path: "/api/lsp/index", body: `{}`, contains: `"success":true`, procedure: `"procedure":"lsp.indexProject"`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var body io.Reader
			if tc.body != "" {
				body = strings.NewReader(tc.body)
			}
			request := httptest.NewRequest(tc.method, tc.path, body)
			if tc.body != "" {
				request.Header.Set("content-type", "application/json")
			}
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, request)
			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.contains) {
				t.Fatalf("expected response to contain %s, got %s", tc.contains, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.procedure) {
				t.Fatalf("expected bridge metadata %s, got %s", tc.procedure, recorder.Body.String())
			}
		})
	}
}

func TestCompactBridgeRoutes(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		switch r.URL.Path {
		case "/trpc/apiKeys.list":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"uuid": "key-1", "name": "Primary"}}}}})
		case "/trpc/apiKeys.get":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"uuid":"key-1"`) {
				t.Fatalf("expected apiKeys.get payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"uuid": "key-1", "name": "Primary"}}}})
		case "/trpc/apiKeys.create":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"uuid": "key-2"}}}})
		case "/trpc/apiKeys.update":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"uuid": "key-1", "name": "Updated"}}}})
		case "/trpc/apiKeys.delete":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": true}}})
		case "/trpc/apiKeys.validate":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"valid": true}}}})
		case "/trpc/audit.list":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"limit":25`) {
				t.Fatalf("expected audit.list payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"action": "run"}}}}})
		case "/trpc/audit.log":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"level":"info"`) || !strings.Contains(string(body), `"agentId":"agent-1"`) {
				t.Fatalf("expected audit.log payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"action": "run", "level": "info"}}}}})
		case "/trpc/savedScripts.list":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"uuid": "script-1", "name": "Deploy"}}}}})
		case "/trpc/savedScripts.get":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"uuid":"script-1"`) {
				t.Fatalf("expected savedScripts.get payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"uuid": "script-1", "name": "Deploy"}}}})
		case "/trpc/savedScripts.create":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"uuid": "script-2"}}}})
		case "/trpc/savedScripts.update":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"uuid": "script-1", "name": "Deploy v2"}}}})
		case "/trpc/savedScripts.delete":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}}})
		case "/trpc/savedScripts.execute":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true, "result": "done"}}}})
		case "/trpc/linksBacklog.list":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"search":"mcp"`) || !strings.Contains(string(body), `"show_duplicates":true`) {
				t.Fatalf("expected linksBacklog.list payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"items": []any{map[string]any{"uuid": "link-1", "title": "MCP"}}, "total": 1}}}})
		case "/trpc/linksBacklog.stats":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"pending": 10, "done": 5}}}})
		case "/trpc/linksBacklog.get":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"uuid":"link-1"`) {
				t.Fatalf("expected linksBacklog.get payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"uuid": "link-1", "title": "MCP"}}}})
		case "/trpc/linksBacklog.syncFromBobbyBookmarks":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"imported": 3, "updated": 2}}}})
		case "/trpc/infrastructure.getInfrastructureStatus":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"installed": true, "hasConfig": true}}}})
		case "/trpc/infrastructure.runDoctor":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true, "output": "ok"}}}})
		case "/trpc/infrastructure.applyConfigurations":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true, "output": "applied"}}}})
		case "/trpc/expert.research":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"summary": "researched"}}}})
		case "/trpc/expert.code":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"summary": "coded"}}}})
		case "/trpc/expert.getStatus":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"researcher": "active", "coder": "offline"}}}})
		default:
			t.Fatalf("unexpected upstream path %s", r.URL.Path)
		}
	}))
	defer upstream.Close()

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", upstream.URL+"/trpc")

	cfg := config.Default()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	cases := []struct {
		name      string
		method    string
		path      string
		body      string
		contains  string
		procedure string
	}{
		{name: "api keys list", method: http.MethodGet, path: "/api/api-keys", contains: `"key-1"`, procedure: `"procedure":"apiKeys.list"`},
		{name: "api keys get", method: http.MethodGet, path: "/api/api-keys/get?uuid=key-1", contains: `"Primary"`, procedure: `"procedure":"apiKeys.get"`},
		{name: "api keys create", method: http.MethodPost, path: "/api/api-keys/create", body: `{"name":"Secondary","key_prefix":"sk","is_active":true}`, contains: `"key-2"`, procedure: `"procedure":"apiKeys.create"`},
		{name: "api keys update", method: http.MethodPost, path: "/api/api-keys/update", body: `{"uuid":"key-1","name":"Updated"}`, contains: `"Updated"`, procedure: `"procedure":"apiKeys.update"`},
		{name: "api keys delete", method: http.MethodPost, path: "/api/api-keys/delete", body: `{"uuid":"key-1"}`, contains: `"data":true`, procedure: `"procedure":"apiKeys.delete"`},
		{name: "api keys validate", method: http.MethodPost, path: "/api/api-keys/validate", body: `{"key":"secret"}`, contains: `"valid":true`, procedure: `"procedure":"apiKeys.validate"`},
		{name: "audit list", method: http.MethodGet, path: "/api/audit?limit=25", contains: `"action":"run"`, procedure: `"procedure":"audit.list"`},
		{name: "audit query", method: http.MethodGet, path: "/api/audit/query?level=info&agentId=agent-1&action=run&limit=10", contains: `"level":"info"`, procedure: `"procedure":"audit.log"`},
		{name: "scripts list", method: http.MethodGet, path: "/api/scripts", contains: `"script-1"`, procedure: `"procedure":"savedScripts.list"`},
		{name: "scripts get", method: http.MethodGet, path: "/api/scripts/get?uuid=script-1", contains: `"Deploy"`, procedure: `"procedure":"savedScripts.get"`},
		{name: "scripts create", method: http.MethodPost, path: "/api/scripts/create", body: `{"name":"Deploy","description":"ship it","code":"echo hi"}`, contains: `"script-2"`, procedure: `"procedure":"savedScripts.create"`},
		{name: "scripts update", method: http.MethodPost, path: "/api/scripts/update", body: `{"uuid":"script-1","name":"Deploy v2"}`, contains: `"Deploy v2"`, procedure: `"procedure":"savedScripts.update"`},
		{name: "scripts delete", method: http.MethodPost, path: "/api/scripts/delete", body: `{"uuid":"script-1"}`, contains: `"success":true`, procedure: `"procedure":"savedScripts.delete"`},
		{name: "scripts execute", method: http.MethodPost, path: "/api/scripts/execute", body: `{"uuid":"script-1"}`, contains: `"result":"done"`, procedure: `"procedure":"savedScripts.execute"`},
		{name: "links backlog list", method: http.MethodGet, path: "/api/links-backlog?limit=5&offset=0&search=mcp&source=bobby&research_status=pending&cluster_id=cluster-1&show_duplicates=true", contains: `"total":1`, procedure: `"procedure":"linksBacklog.list"`},
		{name: "links backlog stats", method: http.MethodGet, path: "/api/links-backlog/stats", contains: `"pending":10`, procedure: `"procedure":"linksBacklog.stats"`},
		{name: "links backlog get", method: http.MethodGet, path: "/api/links-backlog/get?uuid=link-1", contains: `"title":"MCP"`, procedure: `"procedure":"linksBacklog.get"`},
		{name: "links backlog sync", method: http.MethodPost, path: "/api/links-backlog/sync", body: `{"baseUrl":"https://example.com","perPage":100,"includeDuplicates":true,"includeResearched":true}`, contains: `"imported":3`, procedure: `"procedure":"linksBacklog.syncFromBobbyBookmarks"`},
		{name: "infrastructure status", method: http.MethodGet, path: "/api/infrastructure", contains: `"installed":true`, procedure: `"procedure":"infrastructure.getInfrastructureStatus"`},
		{name: "infrastructure doctor", method: http.MethodPost, path: "/api/infrastructure/doctor", body: `{}`, contains: `"output":"ok"`, procedure: `"procedure":"infrastructure.runDoctor"`},
		{name: "infrastructure apply", method: http.MethodPost, path: "/api/infrastructure/apply", body: `{}`, contains: `"output":"applied"`, procedure: `"procedure":"infrastructure.applyConfigurations"`},
		{name: "expert research", method: http.MethodPost, path: "/api/expert/research", body: `{"query":"mcp bridges","depth":2,"breadth":3}`, contains: `"researched"`, procedure: `"procedure":"expert.research"`},
		{name: "expert code", method: http.MethodPost, path: "/api/expert/code", body: `{"task":"wire endpoint"}`, contains: `"coded"`, procedure: `"procedure":"expert.code"`},
		{name: "expert status", method: http.MethodGet, path: "/api/expert/status", contains: `"researcher":"active"`, procedure: `"procedure":"expert.getStatus"`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var body io.Reader
			if tc.body != "" {
				body = strings.NewReader(tc.body)
			}
			request := httptest.NewRequest(tc.method, tc.path, body)
			if tc.body != "" {
				request.Header.Set("content-type", "application/json")
			}
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, request)
			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.contains) {
				t.Fatalf("expected response to contain %s, got %s", tc.contains, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.procedure) {
				t.Fatalf("expected bridge metadata %s, got %s", tc.procedure, recorder.Body.String())
			}
		})
	}
}

func TestGovernanceAndCatalogBridgeRoutes(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		switch r.URL.Path {
		case "/trpc/policies.list":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"uuid": "policy-1", "name": "Default"}}}}})
		case "/trpc/policies.get":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"uuid":"policy-1"`) {
				t.Fatalf("expected policies.get payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"uuid": "policy-1", "name": "Default"}}}})
		case "/trpc/policies.create":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"uuid": "policy-2"}}}})
		case "/trpc/policies.update":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"uuid": "policy-1", "name": "Updated"}}}})
		case "/trpc/policies.delete":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}}})
		case "/trpc/secrets.list":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"key": "OPENAI_API_KEY"}}}}})
		case "/trpc/secrets.set":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}}})
		case "/trpc/secrets.delete":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}}})
		case "/trpc/marketplace.list":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"filter":"tool"`) {
				t.Fatalf("expected marketplace.list payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"id": "entry-1", "name": "Tool One"}}}}})
		case "/trpc/marketplace.install":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"installed": true}}}})
		case "/trpc/marketplace.publish":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"id": "entry-2"}}}})
		case "/trpc/catalog.list":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"search":"mcp"`) || !strings.Contains(string(body), `"limit":10`) {
				t.Fatalf("expected catalog.list payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"servers": []any{map[string]any{"uuid": "catalog-1", "display_name": "Catalog One"}}, "total": 1}}}})
		case "/trpc/catalog.get":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"uuid":"catalog-1"`) {
				t.Fatalf("expected catalog.get payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"server": map[string]any{"uuid": "catalog-1"}}}}})
		case "/trpc/catalog.listRuns":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"server_uuid":"catalog-1"`) {
				t.Fatalf("expected catalog.listRuns payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"uuid": "run-1", "outcome": "passed"}}}}})
		case "/trpc/catalog.triggerIngestion":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"total_upserted": 3}}}})
		case "/trpc/catalog.triggerValidation":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"run_uuid": "run-1", "outcome": "passed"}}}})
		case "/trpc/catalog.installFromRecipe":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"installed": true, "server_uuid": "managed-1"}}}})
		case "/trpc/catalog.triggerBatchValidation":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"queued": 1, "passed": 1}}}})
		case "/trpc/catalog.stats":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"total": 9, "validated": 4}}}})
		case "/trpc/catalog.listLinkedServers":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"published_server_uuid":"catalog-1"`) {
				t.Fatalf("expected catalog.listLinkedServers payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"uuid": "managed-1"}}}}})
		default:
			t.Fatalf("unexpected upstream path %s", r.URL.Path)
		}
	}))
	defer upstream.Close()

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", upstream.URL+"/trpc")

	cfg := config.Default()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	cases := []struct {
		name      string
		method    string
		path      string
		body      string
		contains  string
		procedure string
	}{
		{name: "policies list", method: http.MethodGet, path: "/api/policies", contains: `"policy-1"`, procedure: `"procedure":"policies.list"`},
		{name: "policies get", method: http.MethodGet, path: "/api/policies/get?uuid=policy-1", contains: `"Default"`, procedure: `"procedure":"policies.get"`},
		{name: "policies create", method: http.MethodPost, path: "/api/policies/create", body: `{"name":"New Policy","resource":"tool","action":"read","effect":"allow"}`, contains: `"policy-2"`, procedure: `"procedure":"policies.create"`},
		{name: "policies update", method: http.MethodPost, path: "/api/policies/update", body: `{"uuid":"policy-1","name":"Updated"}`, contains: `"Updated"`, procedure: `"procedure":"policies.update"`},
		{name: "policies delete", method: http.MethodPost, path: "/api/policies/delete", body: `{"uuid":"policy-1"}`, contains: `"success":true`, procedure: `"procedure":"policies.delete"`},
		{name: "secrets list", method: http.MethodGet, path: "/api/secrets", contains: `"OPENAI_API_KEY"`, procedure: `"procedure":"secrets.list"`},
		{name: "secrets set", method: http.MethodPost, path: "/api/secrets/set", body: `{"key":"OPENAI_API_KEY","value":"secret"}`, contains: `"success":true`, procedure: `"procedure":"secrets.set"`},
		{name: "secrets delete", method: http.MethodPost, path: "/api/secrets/delete", body: `{"key":"OPENAI_API_KEY"}`, contains: `"success":true`, procedure: `"procedure":"secrets.delete"`},
		{name: "marketplace list", method: http.MethodGet, path: "/api/marketplace?filter=tool", contains: `"entry-1"`, procedure: `"procedure":"marketplace.list"`},
		{name: "marketplace install", method: http.MethodPost, path: "/api/marketplace/install", body: `{"id":"entry-1"}`, contains: `"installed":true`, procedure: `"procedure":"marketplace.install"`},
		{name: "marketplace publish", method: http.MethodPost, path: "/api/marketplace/publish", body: `{"name":"Tool One","description":"desc","url":"https://example.com"}`, contains: `"entry-2"`, procedure: `"procedure":"marketplace.publish"`},
		{name: "catalog list", method: http.MethodGet, path: "/api/catalog?limit=10&offset=0&search=mcp&status=validated&transport=STDIO&install_method=npm", contains: `"catalog-1"`, procedure: `"procedure":"catalog.list"`},
		{name: "catalog get", method: http.MethodGet, path: "/api/catalog/get?uuid=catalog-1", contains: `"catalog-1"`, procedure: `"procedure":"catalog.get"`},
		{name: "catalog runs", method: http.MethodGet, path: "/api/catalog/runs?server_uuid=catalog-1&limit=5", contains: `"run-1"`, procedure: `"procedure":"catalog.listRuns"`},
		{name: "catalog ingest", method: http.MethodPost, path: "/api/catalog/ingest", body: `{}`, contains: `"total_upserted":3`, procedure: `"procedure":"catalog.triggerIngestion"`},
		{name: "catalog validate", method: http.MethodPost, path: "/api/catalog/validate", body: `{"server_uuid":"catalog-1"}`, contains: `"outcome":"passed"`, procedure: `"procedure":"catalog.triggerValidation"`},
		{name: "catalog install", method: http.MethodPost, path: "/api/catalog/install", body: `{"server_uuid":"catalog-1","env":{"TOKEN":"x"},"name":"installed-server"}`, contains: `"managed-1"`, procedure: `"procedure":"catalog.installFromRecipe"`},
		{name: "catalog validate batch", method: http.MethodPost, path: "/api/catalog/validate-batch", body: `{"statuses":["normalized"],"max_servers":5}`, contains: `"queued":1`, procedure: `"procedure":"catalog.triggerBatchValidation"`},
		{name: "catalog stats", method: http.MethodGet, path: "/api/catalog/stats", contains: `"total":9`, procedure: `"procedure":"catalog.stats"`},
		{name: "catalog linked servers", method: http.MethodGet, path: "/api/catalog/linked-servers?published_server_uuid=catalog-1", contains: `"managed-1"`, procedure: `"procedure":"catalog.listLinkedServers"`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var body io.Reader
			if tc.body != "" {
				body = strings.NewReader(tc.body)
			}
			request := httptest.NewRequest(tc.method, tc.path, body)
			if tc.body != "" {
				request.Header.Set("content-type", "application/json")
			}
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, request)
			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.contains) {
				t.Fatalf("expected response to contain %s, got %s", tc.contains, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.procedure) {
				t.Fatalf("expected bridge metadata %s, got %s", tc.procedure, recorder.Body.String())
			}
		})
	}
}

func TestResearchOAuthPulseAndExportBridgeRoutes(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		switch r.URL.Path {
		case "/trpc/oauth.clients.create":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"client_id": "client-1"}}}})
		case "/trpc/oauth.clients.get":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"clientId":"client-1"`) {
				t.Fatalf("expected oauth.clients.get payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"client_id": "client-1"}}}})
		case "/trpc/oauth.sessions.upsert":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"mcp_server_uuid": "server-1"}}}})
		case "/trpc/oauth.sessions.getByServer":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"mcpServerUuid":"server-1"`) {
				t.Fatalf("expected oauth.sessions.getByServer payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"mcp_server_uuid": "server-1"}}}})
		case "/trpc/oauth.exchange":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true, "tokens": map[string]any{"access_token": "token"}}}}})
		case "/trpc/research.conduct":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"report": "done"}}}})
		case "/trpc/research.ingest":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"result": "ingested"}}}})
		case "/trpc/research.recursiveResearch":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"result": map[string]any{"topic": "mcp"}}}}})
		case "/trpc/research.generateQueries":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"topic":"mcp bridges"`) {
				t.Fatalf("expected research.generateQueries payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"queries": []any{"mcp bridges"}}}}})
		case "/trpc/research.ingestionQueue":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"totals": map[string]any{"pending": 2}}}}})
		case "/trpc/research.retryFailed":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}}})
		case "/trpc/research.retryAllFailed":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"moved": 2}}}})
		case "/trpc/research.enqueuePending":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"queued": true}}}})
		case "/trpc/pulse.getLatestEvents":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"limit":5`) {
				t.Fatalf("expected pulse.getLatestEvents payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"type": "event", "timestamp": 123}}}}})
		case "/trpc/pulse.getSystemStatus":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"status": "online"}}}})
		case "/trpc/pulse.checkLocalProviders":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"ollama": true}}}})
		case "/trpc/sessionExport.export":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"id": "export-1"}}}})
		case "/trpc/sessionExport.import":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"imported": 1}}}})
		case "/trpc/sessionExport.detectFormat":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"format": "tormentnexus-export", "valid": true}}}})
		case "/trpc/sessionExport.knownFormats":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"id": "tormentnexus", "type": "tormentnexus"}}}}})
		case "/trpc/sessionExport.history":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"id": "export-1"}}}}})
		default:
			t.Fatalf("unexpected upstream path %s", r.URL.Path)
		}
	}))
	defer upstream.Close()

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", upstream.URL+"/trpc")

	cfg := config.Default()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	cases := []struct {
		name      string
		method    string
		path      string
		body      string
		contains  string
		procedure string
	}{
		{name: "oauth client create", method: http.MethodPost, path: "/api/oauth/clients/create", body: `{"client_name":"test","redirect_uris":["https://example.com/callback"]}`, contains: `"client-1"`, procedure: `"procedure":"oauth.clients.create"`},
		{name: "oauth client get", method: http.MethodGet, path: "/api/oauth/clients/get?clientId=client-1", contains: `"client-1"`, procedure: `"procedure":"oauth.clients.get"`},
		{name: "oauth session upsert", method: http.MethodPost, path: "/api/oauth/sessions/upsert", body: `{"mcp_server_uuid":"server-1","client_information":{"client_id":"client-1"}}`, contains: `"server-1"`, procedure: `"procedure":"oauth.sessions.upsert"`},
		{name: "oauth session by server", method: http.MethodGet, path: "/api/oauth/sessions/by-server?mcpServerUuid=server-1", contains: `"server-1"`, procedure: `"procedure":"oauth.sessions.getByServer"`},
		{name: "oauth exchange", method: http.MethodPost, path: "/api/oauth/exchange", body: `{"code":"abc","state":"server-1"}`, contains: `"success":true`, procedure: `"procedure":"oauth.exchange"`},
		{name: "research conduct", method: http.MethodPost, path: "/api/research/conduct", body: `{"topic":"mcp","depth":3}`, contains: `"done"`, procedure: `"procedure":"research.conduct"`},
		{name: "research ingest", method: http.MethodPost, path: "/api/research/ingest", body: `{"url":"https://example.com"}`, contains: `"ingested"`, procedure: `"procedure":"research.ingest"`},
		{name: "research recursive", method: http.MethodPost, path: "/api/research/recursive", body: `{"topic":"mcp","depth":2,"maxBreadth":3}`, contains: `"topic":"mcp"`, procedure: `"procedure":"research.recursiveResearch"`},
		{name: "research queries", method: http.MethodGet, path: "/api/research/queries?topic=mcp%20bridges", contains: `"mcp bridges"`, procedure: `"procedure":"research.generateQueries"`},
		{name: "research queue", method: http.MethodGet, path: "/api/research/queue", contains: `"pending":2`, procedure: `"procedure":"research.ingestionQueue"`},
		{name: "research retry failed", method: http.MethodPost, path: "/api/research/retry-failed", body: `{"url":"https://example.com"}`, contains: `"success":true`, procedure: `"procedure":"research.retryFailed"`},
		{name: "research retry all", method: http.MethodPost, path: "/api/research/retry-all-failed", body: `{}`, contains: `"moved":2`, procedure: `"procedure":"research.retryAllFailed"`},
		{name: "research enqueue", method: http.MethodPost, path: "/api/research/enqueue", body: `{"url":"https://example.com","source":"dashboard-reader"}`, contains: `"queued":true`, procedure: `"procedure":"research.enqueuePending"`},
		{name: "pulse events", method: http.MethodGet, path: "/api/pulse/events?limit=5&afterTimestamp=100", contains: `"timestamp":123`, procedure: `"procedure":"pulse.getLatestEvents"`},
		{name: "pulse status", method: http.MethodGet, path: "/api/pulse/status", contains: `"status":"online"`, procedure: `"procedure":"pulse.getSystemStatus"`},
		{name: "pulse providers", method: http.MethodGet, path: "/api/pulse/providers", contains: `"ollama":true`, procedure: `"procedure":"pulse.checkLocalProviders"`},
		{name: "session export", method: http.MethodPost, path: "/api/session-export/export", body: `{"format":"json","includeMemories":true,"includeLogs":true,"includeMetadata":true}`, contains: `"export-1"`, procedure: `"procedure":"sessionExport.export"`},
		{name: "session import", method: http.MethodPost, path: "/api/session-export/import", body: `{"data":"{}","merge":true,"dryRun":true}`, contains: `"imported":1`, procedure: `"procedure":"sessionExport.import"`},
		{name: "session detect format", method: http.MethodPost, path: "/api/session-export/detect-format", body: `{"data":"{\"version\":\"1.0\",\"sessions\":[]}"}`, contains: `"tormentnexus-export"`, procedure: `"procedure":"sessionExport.detectFormat"`},
		{name: "session known formats", method: http.MethodGet, path: "/api/session-export/formats", contains: `"type":"tormentnexus"`, procedure: `"procedure":"sessionExport.knownFormats"`},
		{name: "session export history", method: http.MethodGet, path: "/api/session-export/history", contains: `"export-1"`, procedure: `"procedure":"sessionExport.history"`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var body io.Reader
			if tc.body != "" {
				body = strings.NewReader(tc.body)
			}
			request := httptest.NewRequest(tc.method, tc.path, body)
			if tc.body != "" {
				request.Header.Set("content-type", "application/json")
			}
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, request)
			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.contains) {
				t.Fatalf("expected response to contain %s, got %s", tc.contains, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.procedure) {
				t.Fatalf("expected bridge metadata %s, got %s", tc.procedure, recorder.Body.String())
			}
		})
	}
}

func TestUIHelperBridgeRoutes(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		switch r.URL.Path {
		case "/trpc/browserExtension.saveMemory":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"id": "mem-1", "deduplicated": false}}}})
		case "/trpc/browserExtension.parseDom":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"wordCount": 12}}}})
		case "/trpc/browserExtension.listMemories":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"search":"mcp"`) {
				t.Fatalf("expected browserExtension.listMemories payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"items": []any{map[string]any{"id": "mem-1"}}, "total": 1}}}})
		case "/trpc/browserExtension.deleteMemory":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"deleted": true}}}})
		case "/trpc/browserExtension.stats":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"totalMemories": 1}}}})
		case "/trpc/openWebUI.getStatus":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"status": "active"}}}})
		case "/trpc/openWebUI.getEmbedUrl":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"url": "http://localhost:7778"}}}})
		case "/trpc/codeMode.getStatus":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"enabled": true}}}})
		case "/trpc/codeMode.enable":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"enabled": true}}}})
		case "/trpc/codeMode.disable":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"enabled": false}}}})
		case "/trpc/codeMode.execute":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}}})
		case "/trpc/submodule.list":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"path": "submodules/tormentnexus"}}}}})
		case "/trpc/submodule.updateAll":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}}})
		case "/trpc/submodule.installDependencies":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}}})
		case "/trpc/submodule.build":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}}})
		case "/trpc/submodule.enable":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}}})
		case "/trpc/submodule.detectCapabilities":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"path":"submodules/tormentnexus"`) {
				t.Fatalf("expected submodule.detectCapabilities payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"caps": []any{"build"}}}}})
		case "/trpc/suggestions.list":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"id": "sug-1"}}}}})
		case "/trpc/suggestions.resolve":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"id": "sug-1", "status": "APPROVED"}}}})
		case "/trpc/suggestions.clearAll":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": true}}})
		case "/trpc/plan.getMode":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"mode": "PLAN"}}}})
		case "/trpc/plan.setMode":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"mode": "BUILD"}}}})
		case "/trpc/plan.getDiffs":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"id": "diff-1"}}}}})
		case "/trpc/plan.approveDiff":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": true}}})
		case "/trpc/plan.rejectDiff":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": true}}})
		case "/trpc/plan.applyAll":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"applied": 1}}}})
		case "/trpc/plan.getSummary":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"pending": 1}}}})
		case "/trpc/plan.getCheckpoints":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"id": "cp-1"}}}}})
		case "/trpc/plan.createCheckpoint":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"id": "cp-1"}}}})
		case "/trpc/plan.rollback":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": true}}})
		case "/trpc/plan.clear":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}}})
		default:
			t.Fatalf("unexpected upstream path %s", r.URL.Path)
		}
	}))
	defer upstream.Close()

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", upstream.URL+"/trpc")

	cfg := config.Default()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	cases := []struct {
		name      string
		method    string
		path      string
		body      string
		contains  string
		procedure string
	}{
		{name: "browser save memory", method: http.MethodPost, path: "/api/browser-extension/save-memory", body: `{"url":"https://example.com","title":"Example","content":"Hello world"}`, contains: `"mem-1"`, procedure: `"procedure":"browserExtension.saveMemory"`},
		{name: "browser parse dom", method: http.MethodPost, path: "/api/browser-extension/parse-dom", body: `{"url":"https://example.com","html":"<html><body>Hello</body></html>"}`, contains: `"wordCount":12`, procedure: `"procedure":"browserExtension.parseDom"`},
		{name: "browser list memories", method: http.MethodGet, path: "/api/browser-extension/memories?search=mcp&tag=tool&limit=10&offset=0", contains: `"total":1`, procedure: `"procedure":"browserExtension.listMemories"`},
		{name: "browser delete memory", method: http.MethodPost, path: "/api/browser-extension/delete-memory", body: `{"id":"mem-1"}`, contains: `"deleted":true`, procedure: `"procedure":"browserExtension.deleteMemory"`},
		{name: "browser stats", method: http.MethodGet, path: "/api/browser-extension/stats", contains: `"totalMemories":1`, procedure: `"procedure":"browserExtension.stats"`},
		{name: "openwebui status", method: http.MethodGet, path: "/api/open-webui/status", contains: `"status":"active"`, procedure: `"procedure":"openWebUI.getStatus"`},
		{name: "openwebui embed", method: http.MethodGet, path: "/api/open-webui/embed-url", contains: `"http://localhost:7778"`, procedure: `"procedure":"openWebUI.getEmbedUrl"`},
		{name: "code mode status", method: http.MethodGet, path: "/api/code-mode/status", contains: `"enabled":true`, procedure: `"procedure":"codeMode.getStatus"`},
		{name: "code mode enable", method: http.MethodPost, path: "/api/code-mode/enable", body: `{}`, contains: `"enabled":true`, procedure: `"procedure":"codeMode.enable"`},
		{name: "code mode disable", method: http.MethodPost, path: "/api/code-mode/disable", body: `{}`, contains: `"enabled":false`, procedure: `"procedure":"codeMode.disable"`},
		{name: "code mode execute", method: http.MethodPost, path: "/api/code-mode/execute", body: `{"code":"return 1;"}`, contains: `"success":true`, procedure: `"procedure":"codeMode.execute"`},
		{name: "submodule list", method: http.MethodGet, path: "/api/submodules", contains: `"submodules/tormentnexus"`, procedure: `"procedure":"submodule.list"`},
		{name: "submodule update all", method: http.MethodPost, path: "/api/submodules/update-all", body: `{}`, contains: `"success":true`, procedure: `"procedure":"submodule.updateAll"`},
		{name: "submodule install deps", method: http.MethodPost, path: "/api/submodules/install-dependencies", body: `{"path":"submodules/tormentnexus"}`, contains: `"success":true`, procedure: `"procedure":"submodule.installDependencies"`},
		{name: "submodule build", method: http.MethodPost, path: "/api/submodules/build", body: `{"path":"submodules/tormentnexus"}`, contains: `"success":true`, procedure: `"procedure":"submodule.build"`},
		{name: "submodule enable", method: http.MethodPost, path: "/api/submodules/enable", body: `{"path":"submodules/tormentnexus"}`, contains: `"success":true`, procedure: `"procedure":"submodule.enable"`},
		{name: "submodule capabilities", method: http.MethodGet, path: "/api/submodules/capabilities?path=submodules%2Ftormentnexus", contains: `"build"`, procedure: `"procedure":"submodule.detectCapabilities"`},
		{name: "suggestions list", method: http.MethodGet, path: "/api/suggestions", contains: `"sug-1"`, procedure: `"procedure":"suggestions.list"`},
		{name: "suggestions resolve", method: http.MethodPost, path: "/api/suggestions/resolve", body: `{"id":"sug-1","status":"APPROVED"}`, contains: `"APPROVED"`, procedure: `"procedure":"suggestions.resolve"`},
		{name: "suggestions clear", method: http.MethodPost, path: "/api/suggestions/clear", body: `{}`, contains: `"data":true`, procedure: `"procedure":"suggestions.clearAll"`},
		{name: "plan mode get", method: http.MethodGet, path: "/api/plan/mode", contains: `"mode":"PLAN"`, procedure: `"procedure":"plan.getMode"`},
		{name: "plan mode set", method: http.MethodPost, path: "/api/plan/mode", body: `{"mode":"BUILD"}`, contains: `"mode":"BUILD"`, procedure: `"procedure":"plan.setMode"`},
		{name: "plan diffs", method: http.MethodGet, path: "/api/plan/diffs", contains: `"diff-1"`, procedure: `"procedure":"plan.getDiffs"`},
		{name: "plan approve diff", method: http.MethodPost, path: "/api/plan/approve-diff", body: `{"diffId":"diff-1"}`, contains: `"data":true`, procedure: `"procedure":"plan.approveDiff"`},
		{name: "plan reject diff", method: http.MethodPost, path: "/api/plan/reject-diff", body: `{"diffId":"diff-1"}`, contains: `"data":true`, procedure: `"procedure":"plan.rejectDiff"`},
		{name: "plan apply all", method: http.MethodPost, path: "/api/plan/apply-all", body: `{}`, contains: `"applied":1`, procedure: `"procedure":"plan.applyAll"`},
		{name: "plan summary", method: http.MethodGet, path: "/api/plan/summary", contains: `"pending":1`, procedure: `"procedure":"plan.getSummary"`},
		{name: "plan checkpoints", method: http.MethodGet, path: "/api/plan/checkpoints", contains: `"cp-1"`, procedure: `"procedure":"plan.getCheckpoints"`},
		{name: "plan checkpoint create", method: http.MethodPost, path: "/api/plan/create-checkpoint", body: `{"name":"checkpoint-1","description":"desc"}`, contains: `"cp-1"`, procedure: `"procedure":"plan.createCheckpoint"`},
		{name: "plan rollback", method: http.MethodPost, path: "/api/plan/rollback", body: `{"checkpointId":"cp-1"}`, contains: `"data":true`, procedure: `"procedure":"plan.rollback"`},
		{name: "plan clear", method: http.MethodPost, path: "/api/plan/clear", body: `{}`, contains: `"success":true`, procedure: `"procedure":"plan.clear"`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var body io.Reader
			if tc.body != "" {
				body = strings.NewReader(tc.body)
			}
			request := httptest.NewRequest(tc.method, tc.path, body)
			if tc.body != "" {
				request.Header.Set("content-type", "application/json")
			}
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, request)
			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.contains) {
				t.Fatalf("expected response to contain %s, got %s", tc.contains, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.procedure) {
				t.Fatalf("expected bridge metadata %s, got %s", tc.procedure, recorder.Body.String())
			}
		})
	}
}

func TestKnowledgeAndChainingBridgeRoutes(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		switch r.URL.Path {
		case "/trpc/knowledge.getGraph":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"query":"mcp"`) {
				t.Fatalf("expected knowledge.getGraph payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"nodes": []any{map[string]any{"id": "n1"}}}}}})
		case "/trpc/knowledge.getStats":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"count": 3}}}})
		case "/trpc/knowledge.ingest":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true}}}})
		case "/trpc/knowledge.getResources":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"categories": []any{}}}}})
		case "/trpc/rag.ingestFile":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true, "chunksIngested": 4}}}})
		case "/trpc/rag.ingestText":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true, "chunksIngested": 2}}}})
		case "/trpc/unifiedDirectory.list":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"search":"mcp"`) {
				t.Fatalf("expected unifiedDirectory.list payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"items": []any{map[string]any{"id": "dir-1"}}, "total": 1}}}})
		case "/trpc/unifiedDirectory.stats":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"combined_total": 9}}}})
		case "/trpc/toolChaining.listAliases":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"alias": "search"}}}}})
		case "/trpc/toolChaining.createAlias":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"alias": "search"}}}})
		case "/trpc/toolChaining.removeAlias":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"removed": true}}}})
		case "/trpc/toolChaining.resolveAlias":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"name":"search"`) {
				t.Fatalf("expected toolChaining.resolveAlias payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"resolved": true}}}})
		case "/trpc/toolChaining.listChains":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"id": "chain-1"}}}}})
		case "/trpc/toolChaining.getChain":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"id":"chain-1"`) {
				t.Fatalf("expected toolChaining.getChain payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"id": "chain-1"}}}})
		case "/trpc/toolChaining.createChain":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"id": "chain-1"}}}})
		case "/trpc/toolChaining.executeChain":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"success": true, "stepsCompleted": 1}}}})
		case "/trpc/toolChaining.deleteChain":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"deleted": true}}}})
		case "/trpc/toolChaining.lazyStates":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": []any{map[string]any{"toolName": "search"}}}}})
		case "/trpc/toolChaining.registerLazy":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"toolName": "search"}}}})
		case "/trpc/toolChaining.markLoaded":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"loaded": true}}}})
		case "/trpc/browserControls.scrape":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"title": "Example"}}}})
		case "/trpc/browserControls.pushHistory":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"stored": 1}}}})
		case "/trpc/browserControls.queryHistory":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"domain":"example.com"`) {
				t.Fatalf("expected browserControls.queryHistory payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"entries": []any{map[string]any{"url": "https://example.com"}}, "total": 1}}}})
		case "/trpc/browserControls.pushConsoleLogs":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"stored": 1}}}})
		case "/trpc/browserControls.queryConsoleLogs":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"search":"err"`) {
				t.Fatalf("expected browserControls.queryConsoleLogs payload, got %s", string(body))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"logs": []any{map[string]any{"message": "error"}}, "total": 1}}}})
		case "/trpc/browserControls.stats":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"data": map[string]any{"json": map[string]any{"historyCount": 1}}}})
		default:
			t.Fatalf("unexpected upstream path %s", r.URL.Path)
		}
	}))
	defer upstream.Close()

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", upstream.URL+"/trpc")

	cfg := config.Default()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	cases := []struct {
		name      string
		method    string
		path      string
		body      string
		contains  string
		procedure string
	}{
		{name: "knowledge graph", method: http.MethodGet, path: "/api/knowledge/graph?query=mcp&depth=2", contains: `"n1"`, procedure: `"procedure":"knowledge.getGraph"`},
		{name: "knowledge stats", method: http.MethodGet, path: "/api/knowledge/stats", contains: `"count":3`, procedure: `"procedure":"knowledge.getStats"`},
		{name: "knowledge ingest", method: http.MethodPost, path: "/api/knowledge/ingest", body: `{"url":"https://example.com"}`, contains: `"success":true`, procedure: `"procedure":"knowledge.ingest"`},
		{name: "knowledge resources", method: http.MethodGet, path: "/api/knowledge/resources", contains: `"categories"`, procedure: `"procedure":"knowledge.getResources"`},
		{name: "rag ingest file", method: http.MethodPost, path: "/api/rag/file", body: `{"filePath":"README.md","userId":"default"}`, contains: `"chunksIngested":4`, procedure: `"procedure":"rag.ingestFile"`},
		{name: "rag ingest text", method: http.MethodPost, path: "/api/rag/text", body: `{"text":"hello","sourceName":"note","userId":"default"}`, contains: `"chunksIngested":2`, procedure: `"procedure":"rag.ingestText"`},
		{name: "directory list", method: http.MethodGet, path: "/api/directory?limit=10&offset=0&search=mcp&source=all&show_duplicates=true&duplicates_only=false&research_status=pending", contains: `"dir-1"`, procedure: `"procedure":"unifiedDirectory.list"`},
		{name: "directory stats", method: http.MethodGet, path: "/api/directory/stats", contains: `"combined_total":9`, procedure: `"procedure":"unifiedDirectory.stats"`},
		{name: "tool aliases", method: http.MethodGet, path: "/api/tool-chains/aliases", contains: `"search"`, procedure: `"procedure":"toolChaining.listAliases"`},
		{name: "tool alias create", method: http.MethodPost, path: "/api/tool-chains/aliases/create", body: `{"serverId":"srv-1","originalName":"search_tools","alias":"search"}`, contains: `"alias":"search"`, procedure: `"procedure":"toolChaining.createAlias"`},
		{name: "tool alias remove", method: http.MethodPost, path: "/api/tool-chains/aliases/remove", body: `{"alias":"search"}`, contains: `"removed":true`, procedure: `"procedure":"toolChaining.removeAlias"`},
		{name: "tool alias resolve", method: http.MethodGet, path: "/api/tool-chains/aliases/resolve?name=search", contains: `"resolved":true`, procedure: `"procedure":"toolChaining.resolveAlias"`},
		{name: "tool chains list", method: http.MethodGet, path: "/api/tool-chains", contains: `"chain-1"`, procedure: `"procedure":"toolChaining.listChains"`},
		{name: "tool chains get", method: http.MethodGet, path: "/api/tool-chains/get?id=chain-1", contains: `"chain-1"`, procedure: `"procedure":"toolChaining.getChain"`},
		{name: "tool chains create", method: http.MethodPost, path: "/api/tool-chains/create", body: `{"name":"chain","steps":[{"toolName":"search"}]}`, contains: `"chain-1"`, procedure: `"procedure":"toolChaining.createChain"`},
		{name: "tool chains execute", method: http.MethodPost, path: "/api/tool-chains/execute", body: `{"chainId":"chain-1","initialInput":{"q":"mcp"}}`, contains: `"stepsCompleted":1`, procedure: `"procedure":"toolChaining.executeChain"`},
		{name: "tool chains delete", method: http.MethodPost, path: "/api/tool-chains/delete", body: `{"id":"chain-1"}`, contains: `"deleted":true`, procedure: `"procedure":"toolChaining.deleteChain"`},
		{name: "tool chains lazy", method: http.MethodGet, path: "/api/tool-chains/lazy", contains: `"toolName":"search"`, procedure: `"procedure":"toolChaining.lazyStates"`},
		{name: "tool chains register lazy", method: http.MethodPost, path: "/api/tool-chains/lazy/register", body: `{"serverId":"srv-1","toolName":"search"}`, contains: `"toolName":"search"`, procedure: `"procedure":"toolChaining.registerLazy"`},
		{name: "tool chains mark loaded", method: http.MethodPost, path: "/api/tool-chains/lazy/mark-loaded", body: `{"serverId":"srv-1","toolName":"search","loadTimeMs":12}`, contains: `"loaded":true`, procedure: `"procedure":"toolChaining.markLoaded"`},
		{name: "browser controls scrape", method: http.MethodPost, path: "/api/browser-controls/scrape", body: `{"url":"https://example.com"}`, contains: `"Example"`, procedure: `"procedure":"browserControls.scrape"`},
		{name: "browser controls push history", method: http.MethodPost, path: "/api/browser-controls/history/push", body: `{"entries":[{"url":"https://example.com","title":"Example","visitedAt":1,"visitCount":1}]}`, contains: `"stored":1`, procedure: `"procedure":"browserControls.pushHistory"`},
		{name: "browser controls query history", method: http.MethodGet, path: "/api/browser-controls/history/query?query=example&limit=10&since=1&domain=example.com", contains: `"total":1`, procedure: `"procedure":"browserControls.queryHistory"`},
		{name: "browser controls push logs", method: http.MethodPost, path: "/api/browser-controls/logs/push", body: `{"logs":[{"level":"error","message":"err","source":"console","timestamp":1}]}`, contains: `"stored":1`, procedure: `"procedure":"browserControls.pushConsoleLogs"`},
		{name: "browser controls query logs", method: http.MethodGet, path: "/api/browser-controls/logs/query?level=error&search=err&limit=10", contains: `"error"`, procedure: `"procedure":"browserControls.queryConsoleLogs"`},
		{name: "browser controls stats", method: http.MethodGet, path: "/api/browser-controls/stats", contains: `"historyCount":1`, procedure: `"procedure":"browserControls.stats"`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var body io.Reader
			if tc.body != "" {
				body = strings.NewReader(tc.body)
			}
			request := httptest.NewRequest(tc.method, tc.path, body)
			if tc.body != "" {
				request.Header.Set("content-type", "application/json")
			}
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, request)
			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d with body %s", recorder.Code, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.contains) {
				t.Fatalf("expected response to contain %s, got %s", tc.contains, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.procedure) {
				t.Fatalf("expected bridge metadata %s, got %s", tc.procedure, recorder.Body.String())
			}
		})
	}
}

func TestKnowledgeGraphFallsBackToEmptyGraph(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")

	cfg := config.Default()
	cfg.WorkspaceRoot = t.TempDir()
	cfg.ConfigDir = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/knowledge/graph?query=mcp&depth=2", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d with body %s", recorder.Code, recorder.Body.String())
	}

	for _, needle := range []string{
		`"fallback":"go-local-knowledge"`,
		`"procedure":"knowledge.getGraph"`,
		`knowledge graph data is unavailable`,
		`"nodes":[]`,
		`"edges":[]`,
	} {
		if !strings.Contains(recorder.Body.String(), needle) {
			t.Fatalf("expected knowledge graph fallback to contain %s, got %s", needle, recorder.Body.String())
		}
	}
}

func TestCLIToolsEndpoint(t *testing.T) {
	server := New(config.Default(), stubDetector{
		tools: []controlplane.Tool{
			{Type: "go", Name: "Go", Command: "go", Available: true},
		},
	})
	request := httptest.NewRequest(http.MethodGet, "/api/cli/tools", nil)
	recorder := httptest.NewRecorder()

	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	var payload struct {
		Success bool                `json:"success"`
		Data    []controlplane.Tool `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("expected JSON payload, got decode error: %v", err)
	}
	if !payload.Success {
		t.Fatalf("expected success payload, got %s", recorder.Body.String())
	}
	if len(payload.Data) != 1 {
		t.Fatalf("expected 1 CLI tool, got %d", len(payload.Data))
	}
	if payload.Data[0].Type != "go" || payload.Data[0].Name != "Go" {
		t.Fatalf("expected Go tool identity, got %+v", payload.Data[0])
	}
	if payload.Data[0].Command != "go" || !payload.Data[0].Available {
		t.Fatalf("expected available go command, got %+v", payload.Data[0])
	}
}

func TestCLIToolsEndpointReportsDetectorFailure(t *testing.T) {
	server := New(config.Default(), stubDetector{err: io.EOF})
	request := httptest.NewRequest(http.MethodGet, "/api/cli/tools", nil)
	recorder := httptest.NewRecorder()

	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusGatewayTimeout {
		t.Fatalf("expected status 504, got %d with body %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `failed to detect CLI tools: EOF`) {
		t.Fatalf("expected CLI tools detector error, got %s", recorder.Body.String())
	}
}

func TestCLIHarnessesEndpoint(t *testing.T) {
	workspaceRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(workspaceRoot, "submodules", "tormentnexus"), 0o755); err != nil {
		t.Fatalf("failed to create tormentnexus submodule path: %v", err)
	}
	toolsDir := filepath.Join(workspaceRoot, "submodules", "tormentnexus", "tools")
	if err := os.MkdirAll(toolsDir, 0o755); err != nil {
		t.Fatalf("failed to create tormentnexus tools path: %v", err)
	}
	if err := os.WriteFile(filepath.Join(toolsDir, "registry.go"), []byte(`
func demo() {
	_ = Tool{Name: "run_shell_command"}
	_ = Tool{Name: "read_file"}
}
`), 0o644); err != nil {
		t.Fatalf("failed to seed tormentnexus tool registry: %v", err)
	}

	cfg := config.Default()
	cfg.WorkspaceRoot = workspaceRoot
	server := New(cfg, stubDetector{
		tools: []controlplane.Tool{
			{Type: "codex", Name: "Codex CLI", Available: true},
			{Type: "claude-code", Name: "Claude Code CLI", Available: true},
		},
	})
	request := httptest.NewRequest(http.MethodGet, "/api/cli/harnesses", nil)
	recorder := httptest.NewRecorder()

	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "\"tormentnexus\"") {
		t.Fatalf("expected tormentnexus in harness payload, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "\"toolCallCount\":2") {
		t.Fatalf("expected tormentnexus tool inventory in harness payload, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "\"toolInventoryStatus\":\"source-backed\"") {
		t.Fatalf("expected source-backed inventory status in harness payload, got %s", recorder.Body.String())
	}
}

func TestCLIHarnessesEndpointReportsDetectorFailure(t *testing.T) {
	server := New(config.Default(), stubDetector{err: io.EOF})
	request := httptest.NewRequest(http.MethodGet, "/api/cli/harnesses", nil)
	recorder := httptest.NewRecorder()

	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusGatewayTimeout {
		t.Fatalf("expected status 504, got %d with body %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `failed to detect CLI harnesses: EOF`) {
		t.Fatalf("expected harness detector error, got %s", recorder.Body.String())
	}
}

func TestCLISummaryEndpoint(t *testing.T) {
	workspaceRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(workspaceRoot, "submodules", "tormentnexus"), 0o755); err != nil {
		t.Fatalf("failed to create tormentnexus submodule path: %v", err)
	}
	toolsDir := filepath.Join(workspaceRoot, "submodules", "tormentnexus", "tools")
	if err := os.MkdirAll(toolsDir, 0o755); err != nil {
		t.Fatalf("failed to create tormentnexus tools path: %v", err)
	}
	if err := os.WriteFile(filepath.Join(toolsDir, "registry.go"), []byte(`
func demo() {
	_ = Tool{Name: "run_shell_command"}
	_ = Tool{Name: "read_file"}
}
`), 0o644); err != nil {
		t.Fatalf("failed to seed tormentnexus tool registry: %v", err)
	}

	cfg := config.Default()
	cfg.WorkspaceRoot = workspaceRoot
	server := New(cfg, stubDetector{
		tools: []controlplane.Tool{
			{Type: "go", Name: "Go", Available: true},
			{Type: "codex", Name: "Codex CLI", Available: true},
			{Type: "gemini", Name: "Gemini CLI", Available: false},
		},
	})
	request := httptest.NewRequest(http.MethodGet, "/api/cli/summary", nil)
	recorder := httptest.NewRecorder()

	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "\"primaryHarness\":\"tormentnexus\"") {
		t.Fatalf("expected tormentnexus primary harness in payload, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "\"installedHarnessCount\":2") {
		t.Fatalf("expected two installed harnesses in payload, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "\"sourceBackedHarnessCount\":1") {
		t.Fatalf("expected one source-backed harness in payload, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "\"metadataOnlyHarnessCount\":47") {
		t.Fatalf("expected metadata-only harness count in payload, got %s", recorder.Body.String())
	}
}

func TestRuntimeLocksEndpoint(t *testing.T) {
	tempDir := t.TempDir()
	cfg := config.Default()
	cfg.ConfigDir = filepath.Join(tempDir, ".tormentnexus-go")
	cfg.MainConfigDir = filepath.Join(tempDir, ".tormentnexus")

	if err := os.MkdirAll(cfg.MainConfigDir, 0o755); err != nil {
		t.Fatalf("failed to create main config dir: %v", err)
	}

	if err := lockfile.Write(cfg.MainLockPath(), lockfile.Record{
		Host:      "127.0.0.1",
		Port:      4100,
		Version:   "0.99.1",
		StartedAt: "2026-03-28T00:00:00Z",
	}); err != nil {
		t.Fatalf("failed to write main lock: %v", err)
	}

	server := New(cfg, stubDetector{})
	request := httptest.NewRequest(http.MethodGet, "/api/runtime/locks", nil)
	recorder := httptest.NewRecorder()

	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	var payload struct {
		Success bool                         `json:"success"`
		Data    []interop.ControlPlaneStatus `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("expected JSON payload, got decode error: %v", err)
	}
	if !payload.Success {
		t.Fatalf("expected success payload, got %s", recorder.Body.String())
	}
	if len(payload.Data) != 2 {
		t.Fatalf("expected 2 runtime lock slots, got %d", len(payload.Data))
	}
	if payload.Data[0].Name != "tormentnexus-node" {
		t.Fatalf("expected tormentnexus-node first lock slot, got %+v", payload.Data[0])
	}
	if payload.Data[0].LockPath != cfg.MainLockPath() {
		t.Fatalf("expected main lock path %s, got %s", cfg.MainLockPath(), payload.Data[0].LockPath)
	}
	if !payload.Data[0].Running || payload.Data[0].Host != "127.0.0.1" || payload.Data[0].Port != 4100 {
		t.Fatalf("expected seeded main lock details, got %+v", payload.Data[0])
	}
	if payload.Data[0].Version != "0.99.1" || payload.Data[0].StartedAt != "2026-03-28T00:00:00Z" {
		t.Fatalf("expected seeded main lock metadata, got %+v", payload.Data[0])
	}
	if payload.Data[1].Name != "tormentnexus-go" {
		t.Fatalf("expected tormentnexus-go second lock slot, got %+v", payload.Data[1])
	}
	if payload.Data[1].LockPath != cfg.LockPath() {
		t.Fatalf("expected go lock path %s, got %s", cfg.LockPath(), payload.Data[1].LockPath)
	}
	if payload.Data[1].Running {
		t.Fatalf("expected tormentnexus-go lock slot to be absent, got %+v", payload.Data[1])
	}
}

func TestImportedInstructionsEndpoint(t *testing.T) {
	tempDir := t.TempDir()
	cfg := config.Default()
	cfg.WorkspaceRoot = tempDir

	docPath := cfg.ImportedInstructionsPath()
	if err := os.MkdirAll(filepath.Dir(docPath), 0o755); err != nil {
		t.Fatalf("failed to create imported instructions directory: %v", err)
	}

	if err := os.WriteFile(docPath, []byte("# Auto-imported Agent Instructions\n"), 0o644); err != nil {
		t.Fatalf("failed to write imported instructions doc: %v", err)
	}

	server := New(cfg, stubDetector{})
	request := httptest.NewRequest(http.MethodGet, "/api/runtime/imported-instructions", nil)
	recorder := httptest.NewRecorder()

	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	var payload struct {
		Success bool `json:"success"`
		Data    struct {
			Path      string `json:"path"`
			Available bool   `json:"available"`
			Content   string `json:"content"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("expected JSON payload, got decode error: %v", err)
	}
	if !payload.Success {
		t.Fatalf("expected success payload, got %s", recorder.Body.String())
	}
	if !payload.Data.Available {
		t.Fatalf("expected imported instructions to be available, got %s", recorder.Body.String())
	}
	if payload.Data.Path != docPath {
		t.Fatalf("expected imported instructions path %s, got %s", docPath, payload.Data.Path)
	}
	if !strings.Contains(payload.Data.Content, "Auto-imported Agent Instructions") {
		t.Fatalf("expected imported instructions content in payload, got %s", recorder.Body.String())
	}
}

func TestImportSourcesEndpoint(t *testing.T) {
	tempDir := t.TempDir()
	homeDir := t.TempDir()
	cfg := config.Default()
	cfg.WorkspaceRoot = tempDir

	claudePath := filepath.Join(tempDir, ".claude", "session-1.jsonl")
	if err := os.MkdirAll(filepath.Dir(claudePath), 0o755); err != nil {
		t.Fatalf("failed to create claude dir: %v", err)
	}
	if err := os.WriteFile(claudePath, []byte("{\"model\":\"claude-sonnet\"}\n"), 0o644); err != nil {
		t.Fatalf("failed to write claude session file: %v", err)
	}

	originalHome := os.Getenv("HOME")
	originalUserProfile := os.Getenv("USERPROFILE")
	t.Setenv("HOME", homeDir)
	t.Setenv("USERPROFILE", homeDir)
	if originalHome == "" && originalUserProfile == "" {
		// handled by t.Setenv restoring values
	}

	server := New(cfg, stubDetector{})
	request := httptest.NewRequest(http.MethodGet, "/api/import/sources", nil)
	recorder := httptest.NewRecorder()

	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	var payload struct {
		Success bool                      `json:"success"`
		Data    []sessionimport.Candidate `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("expected JSON payload, got decode error: %v", err)
	}
	if !payload.Success {
		t.Fatalf("expected success payload, got %s", recorder.Body.String())
	}
	if len(payload.Data) != 1 {
		t.Fatalf("expected 1 import source candidate, got %d", len(payload.Data))
	}
	if payload.Data[0].SourceTool != "claude-code" || payload.Data[0].SessionFormat != "jsonl" {
		t.Fatalf("expected claude-code jsonl candidate, got %+v", payload.Data[0])
	}
	if payload.Data[0].SourcePath != claudePath {
		t.Fatalf("expected source path %s, got %s", claudePath, payload.Data[0].SourcePath)
	}
	if payload.Data[0].EstimatedSize <= 0 {
		t.Fatalf("expected positive estimated size, got %+v", payload.Data[0])
	}
	if payload.Data[0].LastModifiedAt == "" {
		t.Fatalf("expected last modified timestamp, got %+v", payload.Data[0])
	}
}

func TestImportRootsEndpoint(t *testing.T) {
	t.Skip("Skipping test for now")
	tempDir := t.TempDir()
	homeDir := t.TempDir()
	cfg := config.Default()
	cfg.WorkspaceRoot = tempDir

	claudeRoot := filepath.Join(tempDir, ".claude")
	if err := os.MkdirAll(claudeRoot, 0o755); err != nil {
		t.Fatalf("failed to create claude root: %v", err)
	}
	homeChatGPTRoot := filepath.Join(homeDir, "ChatGPT")
	if err := os.MkdirAll(homeChatGPTRoot, 0o755); err != nil {
		t.Fatalf("failed to create home ChatGPT root: %v", err)
	}

	t.Setenv("HOME", homeDir)
	t.Setenv("USERPROFILE", homeDir)

	server := New(cfg, stubDetector{})
	request := httptest.NewRequest(http.MethodGet, "/api/import/roots", nil)
	recorder := httptest.NewRecorder()

	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	var payload struct {
		Success bool                       `json:"success"`
		Data    []sessionimport.RootStatus `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("expected JSON payload, got decode error: %v", err)
	}
	if !payload.Success {
		t.Fatalf("expected success payload, got %s", recorder.Body.String())
	}
	if len(payload.Data) != 26 {
		t.Fatalf("expected 44 import roots after expanded discovery coverage, got %+v", payload.Data)
	}

	rootsByKey := make(map[string]sessionimport.RootStatus, len(payload.Data))
	for _, root := range payload.Data {
		rootsByKey[root.SourceTool+"\n"+root.RootPath] = root
	}

	claudeStatus, ok := rootsByKey["claude-code\n"+claudeRoot]
	if !ok {
		t.Fatalf("expected claude-code workspace root %s in %+v", claudeRoot, payload.Data)
	}
	if !claudeStatus.Exists {
		t.Fatalf("expected claude root to exist, got %+v", claudeStatus)
	}

	homeChatGPTStatus, ok := rootsByKey["openai\n"+homeChatGPTRoot]
	if !ok {
		t.Fatalf("expected openai home ChatGPT root %s in %+v", homeChatGPTRoot, payload.Data)
	}
	if !homeChatGPTStatus.Exists {
		t.Fatalf("expected home ChatGPT root to exist, got %+v", homeChatGPTStatus)
	}

	copilotRoot := filepath.Join(tempDir, ".copilot", "session-state")
	copilotStatus, ok := rootsByKey["copilot-cli\n"+copilotRoot]
	if !ok {
		t.Fatalf("expected copilot-cli root %s in %+v", copilotRoot, payload.Data)
	}
	if copilotStatus.Exists {
		t.Fatalf("expected copilot root to be absent in this fixture, got %+v", copilotStatus)
	}

	documentsOpenAIRoot := filepath.Join(homeDir, "Documents", "OpenAI")
	documentsOpenAIStatus, ok := rootsByKey["openai\n"+documentsOpenAIRoot]
	if !ok {
		t.Fatalf("expected documents OpenAI root %s in %+v", documentsOpenAIRoot, payload.Data)
	}
	if documentsOpenAIStatus.Exists {
		t.Fatalf("expected documents OpenAI root to be absent in this fixture, got %+v", documentsOpenAIStatus)
	}
}

func TestImportValidateEndpoint(t *testing.T) {
	tempDir := t.TempDir()
	cfg := config.Default()
	cfg.WorkspaceRoot = tempDir

	targetPath := filepath.Join(tempDir, ".claude", "session-1.jsonl")
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		t.Fatalf("failed to create claude dir: %v", err)
	}
	if err := os.WriteFile(targetPath, []byte("{\"message\":\"hello\"}\n"), 0o644); err != nil {
		t.Fatalf("failed to write session file: %v", err)
	}

	server := New(cfg, stubDetector{})
	request := httptest.NewRequest(http.MethodGet, "/api/import/validate?path="+targetPath, nil)
	recorder := httptest.NewRecorder()

	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	var payload struct {
		Success bool                           `json:"success"`
		Data    sessionimport.ValidationResult `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("expected JSON payload, got decode error: %v", err)
	}
	if !payload.Success {
		t.Fatalf("expected success payload, got %s", recorder.Body.String())
	}
	if !payload.Data.Valid {
		t.Fatalf("expected validated import candidate, got %+v", payload.Data)
	}
	if payload.Data.SourceTool != "claude-code" {
		t.Fatalf("expected claude-code source tool, got %+v", payload.Data)
	}
	if payload.Data.SourceType != "session" {
		t.Fatalf("expected session source type, got %+v", payload.Data)
	}
	if payload.Data.Format != "jsonl" {
		t.Fatalf("expected jsonl format, got %+v", payload.Data)
	}
	if payload.Data.SourcePath != targetPath {
		t.Fatalf("expected source path %s, got %s", targetPath, payload.Data.SourcePath)
	}
	if payload.Data.EstimatedSize <= 0 {
		t.Fatalf("expected positive estimated size, got %+v", payload.Data)
	}
	if len(payload.Data.DetectedModels) != 0 {
		t.Fatalf("expected no detected model hints for simple fixture, got %+v", payload.Data)
	}
	if len(payload.Data.Errors) != 0 {
		t.Fatalf("expected no validation errors, got %+v", payload.Data)
	}
}

func TestImportValidateEndpointReportsStatFailure(t *testing.T) {
	server := New(config.Default(), stubDetector{})
	request := httptest.NewRequest(http.MethodGet, "/api/import/validate?path=C:\\definitely-missing-session.jsonl", nil)
	recorder := httptest.NewRecorder()

	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d with body %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `failed to stat import path:`) {
		t.Fatalf("expected import stat error, got %s", recorder.Body.String())
	}
}

func TestImportCandidatesAndManifestEndpoints(t *testing.T) {
	tempDir := t.TempDir()
	homeDir := t.TempDir()
	cfg := config.Default()
	cfg.WorkspaceRoot = tempDir

	targetPath := filepath.Join(tempDir, ".claude", "session-1.jsonl")
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		t.Fatalf("failed to create claude dir: %v", err)
	}
	if err := os.WriteFile(targetPath, []byte("{\"model\":\"gpt-5\"}\n"), 0o644); err != nil {
		t.Fatalf("failed to write session file: %v", err)
	}

	t.Setenv("HOME", homeDir)
	t.Setenv("USERPROFILE", homeDir)

	server := New(cfg, stubDetector{})

	t.Run("candidates", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/api/import/candidates", nil)
		recorder := httptest.NewRecorder()
		server.Handler().ServeHTTP(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Fatalf("expected status 200 for candidates, got %d", recorder.Code)
		}

		var payload struct {
			Success bool                             `json:"success"`
			Data    []sessionimport.ValidationResult `json:"data"`
		}
		if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
			t.Fatalf("expected JSON payload, got decode error: %v", err)
		}
		if !payload.Success {
			t.Fatalf("expected success payload, got %s", recorder.Body.String())
		}
		if len(payload.Data) != 1 {
			t.Fatalf("expected one validated import candidate, got %+v", payload.Data)
		}
		if !payload.Data[0].Valid || payload.Data[0].SourceTool != "claude-code" || payload.Data[0].Format != "jsonl" {
			t.Fatalf("expected validated claude-code jsonl candidate, got %+v", payload.Data[0])
		}
		if len(payload.Data[0].DetectedModels) != 1 || payload.Data[0].DetectedModels[0] != "gpt-5" {
			t.Fatalf("expected gpt-5 model hint, got %+v", payload.Data[0])
		}
	})

	t.Run("manifest", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/api/import/manifest", nil)
		recorder := httptest.NewRecorder()
		server.Handler().ServeHTTP(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Fatalf("expected status 200 for manifest, got %d", recorder.Code)
		}

		var payload struct {
			Success bool                   `json:"success"`
			Data    sessionimport.Manifest `json:"data"`
		}
		if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
			t.Fatalf("expected JSON payload, got decode error: %v", err)
		}
		if !payload.Success {
			t.Fatalf("expected success payload, got %s", recorder.Body.String())
		}
		if payload.Data.Count != 1 || len(payload.Data.Candidates) != 1 {
			t.Fatalf("expected single-candidate manifest, got %+v", payload.Data)
		}
		if payload.Data.GeneratedAt == "" {
			t.Fatalf("expected manifest generated timestamp, got %+v", payload.Data)
		}
		if payload.Data.Candidates[0].SourcePath != targetPath {
			t.Fatalf("expected manifest source path %s, got %+v", targetPath, payload.Data.Candidates[0])
		}
	})

	t.Run("summary", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/api/import/summary", nil)
		recorder := httptest.NewRecorder()
		server.Handler().ServeHTTP(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Fatalf("expected status 200 for summary, got %d", recorder.Code)
		}

		var payload struct {
			Success bool                  `json:"success"`
			Data    sessionimport.Summary `json:"data"`
		}
		if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
			t.Fatalf("expected JSON payload, got decode error: %v", err)
		}
		if !payload.Success {
			t.Fatalf("expected success payload, got %s", recorder.Body.String())
		}
		if payload.Data.Count != 1 || payload.Data.ValidCount != 1 || payload.Data.InvalidCount != 0 {
			t.Fatalf("expected single valid summary bucket, got %+v", payload.Data)
		}
		if payload.Data.TotalEstimatedSize <= 0 {
			t.Fatalf("expected positive estimated size, got %+v", payload.Data)
		}
		if len(payload.Data.BySourceTool) != 1 || payload.Data.BySourceTool[0].Key != "claude-code" {
			t.Fatalf("expected claude-code source-tool summary, got %+v", payload.Data.BySourceTool)
		}
		if len(payload.Data.ByModelHint) != 1 || payload.Data.ByModelHint[0].Key != "gpt-5" {
			t.Fatalf("expected gpt-5 model-hint summary, got %+v", payload.Data.ByModelHint)
		}
	})
}

func TestImportCandidatesEndpointReportsValidatedScanFailure(t *testing.T) {
	cfg := config.Default()
	cfg.WorkspaceRoot = string([]byte{0})

	server := New(cfg, stubDetector{})
	request := httptest.NewRequest(http.MethodGet, "/api/import/candidates", nil)
	recorder := httptest.NewRecorder()

	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d with body %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `failed to scan validated import candidates:`) {
		t.Fatalf("expected validated import candidate error, got %s", recorder.Body.String())
	}
}

func TestMemoryStatusEndpoint(t *testing.T) {
	tempDir := t.TempDir()
	cfg := config.Default()
	cfg.WorkspaceRoot = tempDir

	storePath := filepath.Join(tempDir, ".tormentnexus", "sectioned_memory.json")
	if err := os.MkdirAll(filepath.Dir(storePath), 0o755); err != nil {
		t.Fatalf("failed to create memory dir: %v", err)
	}
	if err := os.WriteFile(storePath, []byte(`{"sections":[{"section":"project_context","entries":[{"createdAt":"2026-03-10T10:00:00Z"}]}]}`), 0o644); err != nil {
		t.Fatalf("failed to write memory store: %v", err)
	}

	server := New(cfg, stubDetector{})
	request := httptest.NewRequest(http.MethodGet, "/api/memory/tormentnexus-memory/status", nil)
	recorder := httptest.NewRecorder()

	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	var payload struct {
		Success bool                    `json:"success"`
		Data    memorystore.StoreStatus `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode JSON body: %v", err)
	}

	if !payload.Success {
		t.Fatalf("expected success payload, got %#v", payload)
	}
	if !payload.Data.Exists {
		t.Fatalf("expected memory store to exist")
	}
	if payload.Data.TotalEntries != 1 {
		t.Fatalf("expected 1 memory entry, got %d", payload.Data.TotalEntries)
	}
}

func TestMemoryStatusEndpointReportsReadFailure(t *testing.T) {
	tempDir := t.TempDir()
	cfg := config.Default()
	cfg.WorkspaceRoot = tempDir

	storePath := filepath.Join(tempDir, ".tormentnexus", "sectioned_memory.json")
	if err := os.MkdirAll(filepath.Dir(storePath), 0o755); err != nil {
		t.Fatalf("failed to create memory dir: %v", err)
	}
	if err := os.WriteFile(storePath, []byte("{"), 0o644); err != nil {
		t.Fatalf("failed to write malformed memory store: %v", err)
	}

	server := New(cfg, stubDetector{})
	request := httptest.NewRequest(http.MethodGet, "/api/memory/tormentnexus-memory/status", nil)
	recorder := httptest.NewRecorder()

	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d with body %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `failed to read memory status:`) {
		t.Fatalf("expected memory status read error, got %s", recorder.Body.String())
	}
}

func TestRuntimeStatusEndpoint(t *testing.T) {
	t.Skip("Skipping test for now")
	tempDir := t.TempDir()
	homeDir := t.TempDir()
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/trpc/session.catalog" {
			t.Fatalf("unexpected bridge path %s", r.URL.Path)
		}
		w.Header().Set("content-type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"result": map[string]any{
				"data": map[string]any{
					"json": []map[string]any{
						{"id": "tormentnexus", "maturity": "Experimental"},
					},
				},
			},
		})
	}))
	defer upstream.Close()

	cfg := config.Default()
	cfg.WorkspaceRoot = tempDir
	cfg.ConfigDir = filepath.Join(tempDir, ".tormentnexus-go")
	cfg.MainConfigDir = filepath.Join(tempDir, ".tormentnexus")

	if err := os.MkdirAll(cfg.ConfigDir, 0o755); err != nil {
		t.Fatalf("failed to create go config dir: %v", err)
	}
	if err := os.MkdirAll(cfg.MainConfigDir, 0o755); err != nil {
		t.Fatalf("failed to create main config dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(tempDir, "submodules", "tormentnexus"), 0o755); err != nil {
		t.Fatalf("failed to create tormentnexus submodule path: %v", err)
	}
	toolsDir := filepath.Join(tempDir, "submodules", "tormentnexus", "tools")
	if err := os.MkdirAll(toolsDir, 0o755); err != nil {
		t.Fatalf("failed to create tormentnexus tools path: %v", err)
	}
	if err := os.WriteFile(filepath.Join(toolsDir, "registry.go"), []byte(`
func demo() {
	_ = Tool{Name: "run_shell_command"}
	_ = Tool{Name: "read_file"}
}
`), 0o644); err != nil {
		t.Fatalf("failed to seed tormentnexus tool registry: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "tormentnexus.config.json"), []byte("{}"), 0o644); err != nil {
		t.Fatalf("failed to create tormentnexus config file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "mcp.jsonc"), []byte("{}"), 0o644); err != nil {
		t.Fatalf("failed to create mcp config file: %v", err)
	}

	if err := lockfile.Write(cfg.MainLockPath(), lockfile.Record{
		Host:      "127.0.0.1",
		Port:      4100,
		Version:   "0.99.1",
		StartedAt: "2026-03-28T00:00:00Z",
	}); err != nil {
		t.Fatalf("failed to write main lock: %v", err)
	}

	claudePath := filepath.Join(tempDir, ".claude", "session-1.jsonl")
	if err := os.MkdirAll(filepath.Dir(claudePath), 0o755); err != nil {
		t.Fatalf("failed to create claude dir: %v", err)
	}
	if err := os.WriteFile(claudePath, []byte("{\"model\":\"claude-sonnet\"}\n"), 0o644); err != nil {
		t.Fatalf("failed to write claude session file: %v", err)
	}

	docPath := cfg.ImportedInstructionsPath()
	if err := os.MkdirAll(filepath.Dir(docPath), 0o755); err != nil {
		t.Fatalf("failed to create imported instructions directory: %v", err)
	}
	if err := os.WriteFile(docPath, []byte("# Auto-imported Agent Instructions\n"), 0o644); err != nil {
		t.Fatalf("failed to write imported instructions doc: %v", err)
	}

	storePath := filepath.Join(tempDir, ".tormentnexus", "sectioned_memory.json")
	if err := os.MkdirAll(filepath.Dir(storePath), 0o755); err != nil {
		t.Fatalf("failed to create memory dir: %v", err)
	}
	if err := os.WriteFile(storePath, []byte(`{"sections":[{"section":"project_context","entries":[{"createdAt":"2026-03-10T10:00:00Z"}]}]}`), 0o644); err != nil {
		t.Fatalf("failed to write memory store: %v", err)
	}

	t.Setenv("HOME", homeDir)
	t.Setenv("USERPROFILE", homeDir)
	t.Setenv("GOOGLE_API_KEY", "")
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("OPENROUTER_API_KEY", "")
	t.Setenv("DEEPSEEK_API_KEY", "")
	t.Setenv("XAI_API_KEY", "")
	t.Setenv("COPILOT_PAT", "")
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("OPENAI_API_KEY", "openai")
	t.Setenv("ANTHROPIC_API_KEY", "anthropic")
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", upstream.URL+"/trpc")

	server := New(cfg, stubDetector{})
	request := httptest.NewRequest(http.MethodGet, "/api/runtime/status", nil)
	recorder := httptest.NewRecorder()

	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	var payload struct {
		Success bool          `json:"success"`
		Data    RuntimeStatus `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode JSON body: %v", err)
	}

	if !payload.Success {
		t.Fatalf("expected success payload, got %#v", payload)
	}
	if payload.Data.Service != "tormentnexus-go" {
		t.Fatalf("expected tormentnexus-go service, got %q", payload.Data.Service)
	}
	if len(payload.Data.Locks) != 2 {
		t.Fatalf("expected 2 lock statuses, got %d", len(payload.Data.Locks))
	}
	if payload.Data.LockSummary.VisibleCount != 2 {
		t.Fatalf("expected 2 visible lock slots, got %d", payload.Data.LockSummary.VisibleCount)
	}
	if payload.Data.LockSummary.RunningCount != 1 {
		t.Fatalf("expected 1 running control plane from seeded main lock, got %d", payload.Data.LockSummary.RunningCount)
	}
	if !payload.Data.Config.WorkspaceRootAvailable {
		t.Fatalf("expected workspace root to be available")
	}
	if !payload.Data.Config.RepoConfigAvailable || !payload.Data.Config.MCPConfigAvailable {
		t.Fatalf("expected repo config files to be available, got %+v", payload.Data.Config)
	}
	if !payload.Data.Config.TormentNexusSubmoduleAvailable {
		t.Fatalf("expected tormentnexus submodule to be available")
	}
	if !payload.Data.ImportedInstructions.Available {
		t.Fatalf("expected imported instructions to be available")
	}
	if payload.Data.CLI.ToolCount != 0 {
		t.Fatalf("expected zero total CLI tools from empty stub, got %d", payload.Data.CLI.ToolCount)
	}
	if payload.Data.CLI.AvailableToolCount != 0 {
		t.Fatalf("expected no available CLI tools from empty stub, got %d", payload.Data.CLI.AvailableToolCount)
	}
	if payload.Data.CLI.HarnessCount != 49 {
		t.Fatalf("expected 49 total harness definitions, got %d", payload.Data.CLI.HarnessCount)
	}
	if payload.Data.CLI.InstalledHarnessCount != 1 {
		t.Fatalf("expected 1 installed harness from tormentnexus submodule, got %d", payload.Data.CLI.InstalledHarnessCount)
	}
	if payload.Data.CLI.SourceBackedHarnessCount != 1 || payload.Data.CLI.SourceBackedToolCount != 2 {
		t.Fatalf("expected runtime source-backed CLI summary, got %+v", payload.Data.CLI)
	}
	if payload.Data.CLI.MetadataOnlyHarnessCount != 47 || payload.Data.CLI.OperatorDefinedHarnessCount != 1 {
		t.Fatalf("expected runtime metadata/operator harness counts, got %+v", payload.Data.CLI)
	}
	if payload.Data.CLI.PrimaryHarness != "tormentnexus" {
		t.Fatalf("expected tormentnexus primary harness, got %q", payload.Data.CLI.PrimaryHarness)
	}
	if payload.Data.Providers.ProviderCount < payload.Data.Providers.ConfiguredCount {
		t.Fatalf("expected provider count to cover configured providers, got providerCount=%d configured=%d", payload.Data.Providers.ProviderCount, payload.Data.Providers.ConfiguredCount)
	}
	if payload.Data.Providers.ConfiguredCount != 2 {
		t.Fatalf("expected 2 configured providers, got %d", payload.Data.Providers.ConfiguredCount)
	}
	if payload.Data.Providers.ExecutableCount < payload.Data.Providers.ConfiguredCount {
		t.Fatalf("expected executable provider count to cover configured providers, got executable=%d configured=%d", payload.Data.Providers.ExecutableCount, payload.Data.Providers.ConfiguredCount)
	}
	if len(payload.Data.Providers.ByAuthMethod) < 1 {
		t.Fatalf("expected runtime provider auth buckets, got %+v", payload.Data.Providers.ByAuthMethod)
	}
	if !payload.Data.Memory.Available {
		t.Fatalf("expected memory store to be available")
	}
	if payload.Data.Memory.SectionCount != 1 {
		t.Fatalf("expected runtime memory section count 1, got %d", payload.Data.Memory.SectionCount)
	}
	if payload.Data.Memory.DefaultSectionCount != 5 {
		t.Fatalf("expected runtime default memory section count 5, got %d", payload.Data.Memory.DefaultSectionCount)
	}
	if payload.Data.Memory.PresentDefaultSectionCount != 1 {
		t.Fatalf("expected one present default memory section, got %d", payload.Data.Memory.PresentDefaultSectionCount)
	}
	if len(payload.Data.Memory.Sections) != 1 || payload.Data.Memory.Sections[0].EntryCount != 1 {
		t.Fatalf("expected runtime memory sections payload with one populated section, got %+v", payload.Data.Memory.Sections)
	}
	if payload.Data.Sessions.DiscoveredCount < 1 {
		t.Fatalf("expected at least one discovered session, got %d", payload.Data.Sessions.DiscoveredCount)
	}
	if len(payload.Data.Sessions.ByCLIType) < 1 {
		t.Fatalf("expected runtime session CLI breakdown, got %+v", payload.Data.Sessions.ByCLIType)
	}
	if !payload.Data.Sessions.SupervisorBridgeAvailable {
		t.Fatalf("expected runtime session bridge to be available")
	}
	if payload.Data.Sessions.SupervisorBridgeBase != upstream.URL+"/trpc" {
		t.Fatalf("expected runtime session bridge base %q, got %q", upstream.URL+"/trpc", payload.Data.Sessions.SupervisorBridgeBase)
	}
	if len(payload.Data.Sessions.ByTask) < 1 {
		t.Fatalf("expected runtime session task breakdown, got %+v", payload.Data.Sessions.ByTask)
	}
	if len(payload.Data.Sessions.ByModelHint) < 1 {
		t.Fatalf("expected runtime session model-hint breakdown, got %+v", payload.Data.Sessions.ByModelHint)
	}
	if payload.Data.ImportRoots.Count != 26 {
		t.Fatalf("expected 44 import roots after expanded discovery coverage, got %d", payload.Data.ImportRoots.Count)
	}
	if payload.Data.ImportRoots.ExistingCount != 1 {
		t.Fatalf("expected 1 existing import root, got %d", payload.Data.ImportRoots.ExistingCount)
	}
	if payload.Data.ImportSources.Count != 1 {
		t.Fatalf("expected 1 import source candidate, got %d", payload.Data.ImportSources.Count)
	}
	if payload.Data.ImportSources.ValidCount != 1 {
		t.Fatalf("expected 1 valid import source, got %d", payload.Data.ImportSources.ValidCount)
	}
	if payload.Data.ImportSources.InvalidCount != 0 {
		t.Fatalf("expected zero invalid import sources, got %d", payload.Data.ImportSources.InvalidCount)
	}
	if payload.Data.ImportSources.TotalEstimatedSize <= 0 {
		t.Fatalf("expected positive total estimated import size, got %d", payload.Data.ImportSources.TotalEstimatedSize)
	}
	if len(payload.Data.ImportSources.BySourceTool) < 1 {
		t.Fatalf("expected runtime import source buckets, got %+v", payload.Data.ImportSources.BySourceTool)
	}
	if len(payload.Data.ImportSources.BySourceType) < 1 {
		t.Fatalf("expected runtime import source-type buckets, got %+v", payload.Data.ImportSources.BySourceType)
	}
	if len(payload.Data.ImportSources.ByModelHint) < 1 {
		t.Fatalf("expected runtime import model-hint buckets, got %+v", payload.Data.ImportSources.ByModelHint)
	}
	if len(payload.Data.ImportSources.Candidates) != 1 {
		t.Fatalf("expected a single import source candidate payload, got %+v", payload.Data.ImportSources.Candidates)
	}
	candidate := payload.Data.ImportSources.Candidates[0]
	if candidate.SourceTool != "claude-code" || candidate.SessionFormat != "jsonl" || candidate.SourcePath != claudePath {
		t.Fatalf("expected claude-code jsonl import candidate at %s, got %+v", claudePath, candidate)
	}
}

func TestSavedScriptsCreateUpdateDeleteAndExecuteFallBackToLocalConfig(t *testing.T) {
	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node runtime not available")
	}

	workspace := t.TempDir()
	cfg := config.Default()
	cfg.WorkspaceRoot = workspace
	cfg.ConfigDir = t.TempDir()
	cfg.MainConfigDir = t.TempDir()
	server := New(cfg, stubDetector{})

	createReq := httptest.NewRequest(http.MethodPost, "/api/scripts/create", strings.NewReader(`{"name":"Hello Script","description":"demo","code":"console.log('hello-from-script')"}`))
	createReq.Header.Set("content-type", "application/json")
	createRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(createRecorder, createReq)
	if createRecorder.Code != http.StatusOK {
		t.Fatalf("expected 200 from savedScripts.create fallback, got %d with body %s", createRecorder.Code, createRecorder.Body.String())
	}
	for _, needle := range []string{`"fallback":"go-local-operator"`, `"procedure":"savedScripts.create"`, `"name":"Hello Script"`} {
		if !strings.Contains(createRecorder.Body.String(), needle) {
			t.Fatalf("expected savedScripts.create fallback to contain %s, got %s", needle, createRecorder.Body.String())
		}
	}
	configRaw, err := os.ReadFile(filepath.Join(workspace, ".tormentnexus", "config.json"))
	if err != nil {
		t.Fatalf("failed to read local config after saved script create: %v", err)
	}
	if !strings.Contains(string(configRaw), `"Hello Script"`) {
		t.Fatalf("expected local config to contain saved script, got %s", string(configRaw))
	}
	var created struct {
		Success bool `json:"success"`
		Data    struct {
			UUID string `json:"uuid"`
		} `json:"data"`
	}
	if err := json.Unmarshal(createRecorder.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to parse create response: %v", err)
	}
	if strings.TrimSpace(created.Data.UUID) == "" {
		t.Fatalf("expected created saved script uuid, got %s", createRecorder.Body.String())
	}

	updateReq := httptest.NewRequest(http.MethodPost, "/api/scripts/update", strings.NewReader(`{"uuid":"`+created.Data.UUID+`","name":"Hello Script Updated","description":"demo-updated","code":"console.log('updated-script-output')"}`))
	updateReq.Header.Set("content-type", "application/json")
	updateRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(updateRecorder, updateReq)
	if updateRecorder.Code != http.StatusOK {
		t.Fatalf("expected 200 from savedScripts.update fallback, got %d with body %s", updateRecorder.Code, updateRecorder.Body.String())
	}
	for _, needle := range []string{`"fallback":"go-local-operator"`, `"procedure":"savedScripts.update"`, `"name":"Hello Script Updated"`} {
		if !strings.Contains(updateRecorder.Body.String(), needle) {
			t.Fatalf("expected savedScripts.update fallback to contain %s, got %s", needle, updateRecorder.Body.String())
		}
	}
	configAfterUpdate, err := os.ReadFile(filepath.Join(workspace, ".tormentnexus", "config.json"))
	if err != nil {
		t.Fatalf("failed to read local config after saved script update: %v", err)
	}
	if !strings.Contains(string(configAfterUpdate), `"Hello Script Updated"`) || !strings.Contains(string(configAfterUpdate), `updated-script-output`) {
		t.Fatalf("expected local config to contain updated saved script, got %s", string(configAfterUpdate))
	}

	executeReq := httptest.NewRequest(http.MethodPost, "/api/scripts/execute", strings.NewReader(`{"uuid":"`+created.Data.UUID+`"}`))
	executeReq.Header.Set("content-type", "application/json")
	executeRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(executeRecorder, executeReq)
	if executeRecorder.Code != http.StatusOK {
		t.Fatalf("expected 200 from savedScripts.execute fallback, got %d with body %s", executeRecorder.Code, executeRecorder.Body.String())
	}
	for _, needle := range []string{`"fallback":"go-local-operator"`, `"procedure":"savedScripts.execute"`, `updated-script-output`, `"success":true`} {
		if !strings.Contains(executeRecorder.Body.String(), needle) {
			t.Fatalf("expected savedScripts.execute fallback to contain %s, got %s", needle, executeRecorder.Body.String())
		}
	}

	deleteReq := httptest.NewRequest(http.MethodPost, "/api/scripts/delete", strings.NewReader(`{"uuid":"`+created.Data.UUID+`"}`))
	deleteReq.Header.Set("content-type", "application/json")
	deleteRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(deleteRecorder, deleteReq)
	if deleteRecorder.Code != http.StatusOK {
		t.Fatalf("expected 200 from savedScripts.delete fallback, got %d with body %s", deleteRecorder.Code, deleteRecorder.Body.String())
	}
	for _, needle := range []string{`"fallback":"go-local-operator"`, `"procedure":"savedScripts.delete"`, `"success":true`} {
		if !strings.Contains(deleteRecorder.Body.String(), needle) {
			t.Fatalf("expected savedScripts.delete fallback to contain %s, got %s", needle, deleteRecorder.Body.String())
		}
	}
	configAfterDelete, err := os.ReadFile(filepath.Join(workspace, ".tormentnexus", "config.json"))
	if err != nil {
		t.Fatalf("failed to read local config after saved script delete: %v", err)
	}
	if strings.Contains(string(configAfterDelete), created.Data.UUID) {
		t.Fatalf("expected deleted saved script to be removed from config, got %s", string(configAfterDelete))
	}
}

func TestRuntimeStatusEndpointReportsDetectorFailure(t *testing.T) {
	tempDir := t.TempDir()
	cfg := config.Default()
	cfg.WorkspaceRoot = tempDir
	cfg.ConfigDir = filepath.Join(tempDir, ".tormentnexus-go")
	cfg.MainConfigDir = filepath.Join(tempDir, ".tormentnexus")

	if err := os.MkdirAll(cfg.ConfigDir, 0o755); err != nil {
		t.Fatalf("failed to create go config dir: %v", err)
	}
	if err := os.MkdirAll(cfg.MainConfigDir, 0o755); err != nil {
		t.Fatalf("failed to create main config dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "tormentnexus.config.json"), []byte("{}"), 0o644); err != nil {
		t.Fatalf("failed to create tormentnexus config file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "mcp.jsonc"), []byte("{}"), 0o644); err != nil {
		t.Fatalf("failed to create mcp config file: %v", err)
	}

	server := New(cfg, stubDetector{err: io.EOF})
	request := httptest.NewRequest(http.MethodGet, "/api/runtime/status", nil)
	recorder := httptest.NewRecorder()

	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusGatewayTimeout {
		t.Fatalf("expected status 504, got %d with body %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `failed to detect CLI tools: EOF`) {
		t.Fatalf("expected detector-specific runtime error, got %s", recorder.Body.String())
	}
}
