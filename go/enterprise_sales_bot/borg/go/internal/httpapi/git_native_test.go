package httpapi

import (
	"bytes"
	"net"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tormentnexushq/tormentnexus-go/internal/config"
)

func TestSubmoduleUpdateAllFallsBackToNativeGitReport(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	// Skip if a real TS server is running — the fallback won't trigger
	if conn, err := net.DialTimeout("tcp", "127.0.0.1:7779", 100*time.Millisecond); err == nil {
		conn.Close()
		t.Skip("TS server running on port 4000 — fallback test requires offline upstream")
	}

	workspace := t.TempDir()
	init := exec.Command("git", "init")
	init.Dir = workspace
	if output, err := init.CombinedOutput(); err != nil {
		t.Fatalf("git init failed: %v (%s)", err, string(output))
	}

	t.Setenv("TORMENTNEXUS_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")
	cfg := config.Default()
	cfg.WorkspaceRoot = workspace
	cfg.ConfigDir = filepath.Join(workspace, ".tormentnexus-go")
	cfg.MainConfigDir = filepath.Join(workspace, ".tormentnexus")
	server := New(cfg, stubDetector{})

	req := httptest.NewRequest(http.MethodPost, "/api/submodules/update-all", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	for _, needle := range []string{
		`"fallback":"go-local-git-orchestration"`,
		`"procedure":"submodule.updateAll"`,
		`"total":0`,
		`"successful":0`,
		`"failed":0`,
	} {
		if !strings.Contains(body, needle) {
			t.Fatalf("expected response to contain %s, got %s", needle, body)
		}
	}
}
